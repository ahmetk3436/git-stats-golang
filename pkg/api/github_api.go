package api

import (
	"encoding/json"
	"fmt"
	storage "github.com/ahmetk3436/git-stats-golang/internal"
	"github.com/ahmetk3436/git-stats-golang/pkg/common_types"
	"github.com/ahmetk3436/git-stats-golang/pkg/interfaces"
	appMetrics "github.com/ahmetk3436/git-stats-golang/pkg/prometheus"
	// "github.com/ahmetk3436/git-stats-golang/pkg/repository" // Interface is used now
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// log is a package-level logrus instance for structured logging within the api package.
// It's initialized in init() to ensure consistent logging setup.
var log = logrus.New()

// init is automatically called when the package is loaded.
// It configures the package-level logger.
func init() {
	// Configure logrus for JSON formatted output for better machine readability.
	log.SetFormatter(&logrus.JSONFormatter{})
	// Output logs to standard output.
	log.SetOutput(os.Stdout)
	// Set the default logging level. Can be overridden by configuration if needed.
	log.SetLevel(logrus.InfoLevel) // Example: logrus.DebugLevel for more verbose output during development.
}

// GithubApi handles API requests related to GitHub.
// It uses a GitService for interacting with the Git provider and a RedisClient for caching.
type GithubApi struct {
	Repo  interfaces.GitService // Service for Git operations (GitHub specific implementation).
	Redis *storage.RedisClient  // Client for Redis caching.
}

// NewGithubApi creates a new instance of GithubApi.
// It requires a GitService implementation (e.g., *repository.GitHubRepo) and a RedisClient.
func NewGithubApi(gitService interfaces.GitService, redisClient *storage.RedisClient) *GithubApi {
	log.Info("Creating NewGithubApi with GitService interface.")
	return &GithubApi{
		Repo:  gitService,
		Redis: redisClient,
	}
}

// GetAllRepos handles requests to get all repositories for the authenticated GitHub user or a specified owner.
// It checks cache first and falls back to the GitService if data is not cached.
func (ghAPI *GithubApi) GetAllRepos(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	endpointName := "/api/github/repos"
	logCtx := log.WithFields(logrus.Fields{"endpoint": endpointName, "method": r.Method})
	logCtx.Info("GetAllRepos request received.")
	w.Header().Set("Content-Type", "application/json")

	// Determine owner from query parameter, if provided.
	// The GitService's GetAllRepos("") is expected to fetch for the authenticated user.
	ownerQueryParam := r.URL.Query().Get("owner") // Example: ?owner=someorg

	var reposFromSource []*common_types.Repository // Standardized repository type.
	var err error
	dataSource := "API" // Indicates data source for logging (API or Redis).

	redisKey := "github_get_all_repos_" + ownerQueryParam // Cache key includes owner.
	cachedData, redisErr := ghAPI.Redis.Get(redisKey)

	if redisErr == nil && cachedData != nil {
		logCtx.WithField("key", redisKey).Info("Cache hit for GetAllRepos.")
		dataSource = "Redis"
		if err = json.Unmarshal(cachedData, &reposFromSource); err != nil {
			logCtx.WithFields(logrus.Fields{"key": redisKey, "error": err}).Error("Error unmarshalling cached data for GetAllRepos.")
			appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
			http.Error(w, "Error processing cached data.", http.StatusInternalServerError)
			return
		}
		w.Write(cachedData) // Serve directly from cache if unmarshalling is not strictly needed here.
	} else {
		if redisErr != nil {
			logCtx.WithFields(logrus.Fields{"key": redisKey, "error": redisErr}).Warn("Redis GET error for GetAllRepos; proceeding to fetch from API.")
		} else if cachedData == nil {
			logCtx.WithField("key", redisKey).Info("Cache miss for GetAllRepos; fetching from API.")
		}
		dataSource = "API"
		fetchedRepos, fetchErr := ghAPI.Repo.GetAllRepos(ownerQueryParam)
		if fetchErr != nil {
			logCtx.WithField("error", fetchErr).Error("Error fetching repos from GitHub via GitService.")
			appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
			appMetrics.RepositoryFetchesTotal.WithLabelValues("github", "all_repos", "failure").Inc()
			http.Error(w, fetchErr.Error(), http.StatusInternalServerError)
			return
		}
		appMetrics.RepositoryFetchesTotal.WithLabelValues("github", "all_repos", "success").Inc()
		reposFromSource = fetchedRepos

		responseBytes, marshalErr := json.Marshal(reposFromSource)
		if marshalErr != nil {
			logCtx.WithField("error", marshalErr).Error("Error marshalling repos response.")
			appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
			http.Error(w, marshalErr.Error(), http.StatusInternalServerError)
			return
		}
		// Cache the newly fetched data. TTL example: 1 hour (3600 seconds).
		if setErr := ghAPI.Redis.Set(redisKey, responseBytes, 3600); setErr != nil {
			logCtx.WithFields(logrus.Fields{"key": redisKey, "error": setErr}).Error("Redis SET error for GetAllRepos.")
		}
		w.Write(responseBytes)
	}

	duration := time.Since(startTime).Seconds()
	appMetrics.APICallDuration.WithLabelValues("github", endpointName).Observe(duration)
	appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "success").Inc()
	logCtx.WithFields(logrus.Fields{"duration_seconds": duration, "source": dataSource, "items_count": len(reposFromSource)}).Info("GetAllRepos request processed successfully.")
}

// GetRepo handles requests to get a specific repository by ID or "owner/name" string.
// It checks cache first and falls back to the GitService.
func (ghAPI *GithubApi) GetRepo(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	endpointName := "/api/github/repo"
	repoIdentifierQuery := r.URL.Query().Get("projectID") // "projectID" is the legacy query param name.
	logCtx := log.WithFields(logrus.Fields{"endpoint": endpointName, "method": r.Method, "identifier_query": repoIdentifierQuery})
	logCtx.Info("GetRepo request received.")
	w.Header().Set("Content-Type", "application/json")

	if repoIdentifierQuery == "" {
		logCtx.Error("Missing projectID query parameter.")
		appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
		http.Error(w, "projectID query parameter is required.", http.StatusBadRequest)
		return
	}

	var repoData *common_types.Repository
	var err error
	dataSource := "API"

	redisKey := "github_get_repo_" + repoIdentifierQuery
	cachedData, redisErr := ghAPI.Redis.Get(redisKey)

	if redisErr == nil && cachedData != nil {
		logCtx.WithField("key", redisKey).Info("Cache hit for GetRepo.")
		dataSource = "Redis"
		if err = json.Unmarshal(cachedData, &repoData); err != nil {
			logCtx.WithFields(logrus.Fields{"key": redisKey, "error": err}).Error("Error unmarshalling cached data for GetRepo.")
			appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
			http.Error(w, "Error processing cached data.", http.StatusInternalServerError)
			return
		}
		w.Write(cachedData)
	} else {
		if redisErr != nil {
			logCtx.WithFields(logrus.Fields{"key": redisKey, "error": redisErr}).Warn("Redis GET error for GetRepo; proceeding to fetch from API.")
		} else if cachedData == nil {
			logCtx.WithField("key", redisKey).Info("Cache miss for GetRepo; fetching from API.")
		}
		dataSource = "API"
		var identifierToFetch interface{}
		parsedID, parseErr := strconv.ParseInt(repoIdentifierQuery, 10, 64)
		if parseErr == nil {
			identifierToFetch = parsedID // Use int64 ID if parsing succeeds.
		} else {
			identifierToFetch = repoIdentifierQuery // Otherwise, assume it's an "owner/repo" string.
			logCtx.WithField("input", repoIdentifierQuery).Debug("Interpreting projectID as owner/repo string due to ParseInt failure.")
		}

		fetchedRepo, fetchErr := ghAPI.Repo.GetRepo(identifierToFetch)
		if fetchErr != nil {
			logCtx.WithFields(logrus.Fields{"identifier": identifierToFetch, "error": fetchErr}).Error("Error fetching repo from GitHub via GitService.")
			appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
			appMetrics.RepositoryFetchesTotal.WithLabelValues("github", "single_repo", "failure").Inc()
			http.Error(w, "Cannot get project: "+fetchErr.Error(), http.StatusBadRequest) // Provide more specific error.
			return
		}
		appMetrics.RepositoryFetchesTotal.WithLabelValues("github", "single_repo", "success").Inc()
		repoData = fetchedRepo

		responseBytes, marshalErr := json.Marshal(repoData)
		if marshalErr != nil {
			logCtx.WithField("error", marshalErr).Error("Error marshalling repo response.")
			appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
			http.Error(w, "Error marshalling project data.", http.StatusInternalServerError)
			return
		}
		if setErr := ghAPI.Redis.Set(redisKey, responseBytes, 3600); setErr != nil { // Cache for 1 hour.
			logCtx.WithFields(logrus.Fields{"key": redisKey, "error": setErr}).Error("Redis SET error for GetRepo.")
		}
		w.Write(responseBytes)
	}

	duration := time.Since(startTime).Seconds()
	appMetrics.APICallDuration.WithLabelValues("github", endpointName).Observe(duration)
	appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "success").Inc()
	logCtx.WithFields(logrus.Fields{"duration_seconds": duration, "source": dataSource, "repo_name": repoData.Name}).Info("GetRepo request processed successfully.")
}

// GetAllCommits handles requests to get all commits for a specific repository.
// Repository is identified by 'projectOwner' and 'repoName' query parameters.
// It checks cache first and falls back to the GitService.
func (ghAPI *GithubApi) GetAllCommits(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	endpointName := "/api/github/commits"
	projectOwner := r.URL.Query().Get("projectOwner")
	repoName := r.URL.Query().Get("repoName")
	logCtx := log.WithFields(logrus.Fields{"endpoint": endpointName, "method": r.Method, "owner": projectOwner, "repo": repoName})
	logCtx.Info("GetAllCommits request received.")
	w.Header().Set("Content-Type", "application/json")

	if projectOwner == "" || repoName == "" {
		logCtx.Error("Missing projectOwner or repoName query parameters.")
		appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
		http.Error(w, "projectOwner and repoName query parameters are required.", http.StatusBadRequest)
		return
	}

	var commitsFromSource []*common_types.Commit
	var err error
	dataSource := "API"

	redisKey := fmt.Sprintf("github_get_commits_%s_%s", projectOwner, repoName)
	cachedData, redisErr := ghAPI.Redis.Get(redisKey)

	if redisErr == nil && cachedData != nil {
		logCtx.WithField("key", redisKey).Info("Cache hit for GetAllCommits.")
		dataSource = "Redis"
		if err = json.Unmarshal(cachedData, &commitsFromSource); err != nil {
			logCtx.WithFields(logrus.Fields{"key": redisKey, "error": err}).Error("Error unmarshalling cached data for GetAllCommits.")
			appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
			http.Error(w, "Error processing cached data.", http.StatusInternalServerError)
			return
		}
		w.Write(cachedData)
	} else {
		if redisErr != nil {
			logCtx.WithFields(logrus.Fields{"key": redisKey, "error": redisErr}).Warn("Redis GET error for GetAllCommits; proceeding to fetch from API.")
		} else if cachedData == nil {
			logCtx.WithField("key", redisKey).Info("Cache miss for GetAllCommits; fetching from API.")
		}
		dataSource = "API"
		repoIdentifier := fmt.Sprintf("%s/%s", projectOwner, repoName)
		// TODO: Populate CommitListOptions from query params if needed (e.g., branch, page).
		commitOpts := &interfaces.CommitListOptions{PerPage: 100} // Default: 100 commits per page.

		fetchedCommits, fetchErr := ghAPI.Repo.GetProjectCommits(repoIdentifier, commitOpts)
		if fetchErr != nil {
			logCtx.WithFields(logrus.Fields{"repo": repoIdentifier, "error": fetchErr}).Error("Error fetching commits from GitHub via GitService.")
			appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
			appMetrics.RepositoryFetchesTotal.WithLabelValues("github", "commits", "failure").Inc()
			http.Error(w, fetchErr.Error(), http.StatusInternalServerError)
			return
		}
		appMetrics.RepositoryFetchesTotal.WithLabelValues("github", "commits", "success").Inc()
		commitsFromSource = fetchedCommits

		responseBytes, marshalErr := json.Marshal(commitsFromSource)
		if marshalErr != nil {
			logCtx.WithField("error", marshalErr).Error("Error marshalling commits response.")
			appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
			http.Error(w, marshalErr.Error(), http.StatusInternalServerError)
			return
		}
		if setErr := ghAPI.Redis.Set(redisKey, responseBytes, 3600); setErr != nil { // Cache for 1 hour.
			logCtx.WithFields(logrus.Fields{"key": redisKey, "error": setErr}).Error("Redis SET error for GetAllCommits.")
		}
		w.Write(responseBytes)
	}

	duration := time.Since(startTime).Seconds()
	appMetrics.APICallDuration.WithLabelValues("github", endpointName).Observe(duration)
	appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "success").Inc()
	logCtx.WithFields(logrus.Fields{"duration_seconds": duration, "source": dataSource, "items_count": len(commitsFromSource)}).Info("GetAllCommits request processed successfully.")
}

// GetContributors handles requests to get contributors for a specific repository.
// Repository is identified by 'owner' and 'repoName' query parameters.
// It uses the GitService to fetch and return contributor data.
func (ghAPI *GithubApi) GetContributors(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	endpointName := "/api/github/contributors"
	ownerName := r.URL.Query().Get("owner")
	repoName := r.URL.Query().Get("repoName")
	logCtx := log.WithFields(logrus.Fields{"endpoint": endpointName, "method": r.Method, "owner": ownerName, "repo": repoName})
	logCtx.Info("GetContributors request received.")
	w.Header().Set("Content-Type", "application/json")

	if ownerName == "" || repoName == "" {
		logCtx.Error("Owner and repoName query parameters are required for GetContributors.")
		appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
		http.Error(w, "Owner and repoName query parameters are required.", http.StatusBadRequest)
		return
	}

	repoIdentifier := fmt.Sprintf("%s/%s", ownerName, repoName)
	// Caching for contributors can be added here if desired, similar to other handlers.
	// For simplicity in this example, direct API call via GitService is shown.
	dataSource := "API"

	contributors, fetchErr := ghAPI.Repo.GetRepoContributors(repoIdentifier)
	if fetchErr != nil {
		logCtx.WithFields(logrus.Fields{"repo": repoIdentifier, "error": fetchErr}).Error("Error fetching contributors from GitHub via GitService.")
		appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
		appMetrics.RepositoryFetchesTotal.WithLabelValues("github", "contributors", "failure").Inc()
		http.Error(w, fetchErr.Error(), http.StatusInternalServerError)
		return
	}
	appMetrics.RepositoryFetchesTotal.WithLabelValues("github", "contributors", "success").Inc()

	responseBytes, marshalErr := json.Marshal(contributors)
	if marshalErr != nil {
		logCtx.WithField("error", marshalErr).Error("Error marshalling contributors response.")
		appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
		http.Error(w, marshalErr.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(responseBytes)

	duration := time.Since(startTime).Seconds()
	appMetrics.APICallDuration.WithLabelValues("github", endpointName).Observe(duration)
	appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "success").Inc()
	logCtx.WithFields(logrus.Fields{"duration_seconds": duration, "source": dataSource, "items_count": len(contributors)}).Info("GetContributors request processed successfully.")
}

// GetRepoTotalLinesOfCode handles requests to calculate the total lines of code for a repository.
// The repository is identified by 'repoUrl' query parameter (URL to clone).
// This method involves cloning the repository locally to perform line counting.
func (ghAPI *GithubApi) GetRepoTotalLinesOfCode(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	endpointName := "/api/github/loc"
	repoCloneURL := r.URL.Query().Get("repoUrl") // Expects the full clone URL.
	logCtx := log.WithFields(logrus.Fields{"endpoint": endpointName, "method": r.Method, "repo_url": repoCloneURL})
	logCtx.Info("GetRepoTotalLinesOfCode request received.")
	w.Header().Set("Content-Type", "application/json")

	if repoCloneURL == "" {
		logCtx.Error("Missing repoUrl query parameter.")
		appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
		http.Error(w, "repoUrl query parameter is required.", http.StatusBadRequest)
		return
	}

	var linesOfCodeResult []byte
	// var err error // Removed as 'err' is shadowed in blocks below.
	dataSource := "API" // Or "Calculation" as it's not a direct Git provider API call for data.

	redisKey := "github_get_loc_" + repoCloneURL // Cache key based on repo URL.
	cachedData, redisErr := ghAPI.Redis.Get(redisKey)

	if redisErr == nil && cachedData != nil {
		logCtx.WithField("key", redisKey).Info("Cache hit for GetRepoTotalLinesOfCode.")
		dataSource = "Redis"
		linesOfCodeResult = cachedData
		w.Write(linesOfCodeResult)
	} else {
		if redisErr != nil {
			logCtx.WithFields(logrus.Fields{"key": redisKey, "error": redisErr}).Warn("Redis GET error for LOC; proceeding to calculate.")
		} else if cachedData == nil {
			logCtx.WithField("key", redisKey).Info("Cache miss for LOC; calculating now.")
		}
		dataSource = "Calculation" // More accurate term for this path.

		// Create a temporary directory for cloning the repository.
		tempDir, mkDirErr := os.MkdirTemp("", "temp-repo-loc-*") // Pattern for identifiable temp dirs.
		if mkDirErr != nil {
			logCtx.WithField("error", mkDirErr).Error("Error creating temporary directory for LOC calculation.")
			appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
			http.Error(w, fmt.Sprintf("Error creating temp dir: %v", mkDirErr), http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := os.RemoveAll(tempDir); err != nil {
				logCtx.WithFields(logrus.Fields{"tempDir": tempDir, "error": err}).Error("Failed to remove temporary directory.")
			}
		}()
		logCtx.WithField("tempDir", tempDir).Debug("Temporary directory created for cloning.")

		if cloneErr := cloneRepository(repoCloneURL, tempDir); cloneErr != nil {
			logCtx.WithFields(logrus.Fields{"repo_url": repoCloneURL, "tempDir": tempDir, "error": cloneErr}).Error("Error cloning repository for LOC calculation.")
			appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
			appMetrics.RepositoryFetchesTotal.WithLabelValues("github", "loc_clone", "failure").Inc() // Metric for clone attempt.
			http.Error(w, fmt.Sprintf("Error cloning repo: %v", cloneErr), http.StatusInternalServerError)
			return
		}
		appMetrics.RepositoryFetchesTotal.WithLabelValues("github", "loc_clone", "success").Inc()

		// Command to count lines: list files, then count lines for each, sum them up.
		// `git ls-files` lists all tracked files. `xargs wc -l` counts lines for these files. `tail -n 1` gets the total.
		locCommand := "git ls-files | xargs wc -l | tail -n 1"
		output, cmdErr := runCommand(locCommand, tempDir)
		if cmdErr != nil {
			logCtx.WithFields(logrus.Fields{"repo_url": repoCloneURL, "command": locCommand, "error": cmdErr}).Error("Error running command for LOC calculation.")
			appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
			http.Error(w, fmt.Sprintf("Error running command: %v", cmdErr), http.StatusInternalServerError)
			return
		}

		totalLines, extractErr := extractTotalLines(output)
		if extractErr != nil {
			logCtx.WithFields(logrus.Fields{"raw_output": output, "error": extractErr}).Error("Error extracting total lines from command output.")
			appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
			http.Error(w, fmt.Sprintf("Error extracting total lines: %v", extractErr), http.StatusInternalServerError)
			return
		}

		resultMap := map[string]interface{}{"totalLines": totalLines}
		jsonResult, marshalErr := json.Marshal(resultMap)
		if marshalErr != nil {
			logCtx.WithField("error", marshalErr).Error("Error marshalling LOC result.")
			appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "failure").Inc()
			http.Error(w, fmt.Sprintf("Error encoding LOC JSON: %v", marshalErr), http.StatusInternalServerError)
			return
		}
		linesOfCodeResult = jsonResult
		// Cache LOC result for 24 hours (86400 seconds).
		if setErr := ghAPI.Redis.Set(redisKey, linesOfCodeResult, 86400); setErr != nil {
			logCtx.WithFields(logrus.Fields{"key": redisKey, "error": setErr}).Error("Redis SET error for LOC.")
		}
		w.Write(linesOfCodeResult)
	}

	duration := time.Since(startTime).Seconds()
	appMetrics.APICallDuration.WithLabelValues("github", endpointName).Observe(duration)
	appMetrics.APICallsTotal.WithLabelValues("github", endpointName, "success").Inc()
	logCtx.WithFields(logrus.Fields{"duration_seconds": duration, "source": dataSource}).Info("GetRepoTotalLinesOfCode request processed successfully.")
}

// cloneRepository is a helper function to clone a Git repository into a specified directory.
func cloneRepository(repoURL, targetDirectory string) error {
	log.WithFields(logrus.Fields{"repo_url": repoURL, "target_dir": targetDirectory}).Info("Cloning repository...")
	// Basic git clone command. For production, consider more robust error handling,
	// authentication for private repos (if needed, though tokens are server-side), and depth control.
	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, targetDirectory) // Shallow clone for LOC.
	err := cmd.Run()
	if err != nil {
		log.WithFields(logrus.Fields{"repo_url": repoURL, "error": err}).Error("Failed to clone repository.")
		return fmt.Errorf("cloning repository failed: %w", err) // Wrap error for context.
	}
	log.WithFields(logrus.Fields{"repo_url": repoURL}).Info("Repository cloned successfully.")
	return nil
}

// runCommand is a helper function to execute a shell command in a given working directory.
// It returns the command's standard output as a string.
func runCommand(commandString, workingDir string) (string, error) {
	log.WithFields(logrus.Fields{"command": commandString, "cwd": workingDir}).Info("Running command...")
	cmd := exec.Command("bash", "-c", commandString) // Use bash to interpret pipes, etc.
	cmd.Dir = workingDir
	outputBytes, err := cmd.Output() // Runs command and gets stdout. Use CombinedOutput for stdout & stderr.
	if err != nil {
		// If there's an error, log it along with any output captured so far.
		// exitError, ok := err.(*exec.ExitError)
		// if ok { // If it's an ExitError, Stderr might contain useful info.
		//  log.WithFields(logrus.Fields{"command": commandString, "error": err, "stderr": string(exitError.Stderr)}).Error("Command execution failed.")
		// } else {
		log.WithFields(logrus.Fields{"command": commandString, "error": err, "output": string(outputBytes)}).Error("Command execution failed or produced error output.")
		// }
		return string(outputBytes), fmt.Errorf("running command '%s' failed: %w", commandString, err)
	}
	log.WithFields(logrus.Fields{"command": commandString, "output_length": len(outputBytes)}).Info("Command executed successfully.")
	return string(outputBytes), nil
}

// extractTotalLines parses the output of 'wc -l' (specifically the total line) to get the total number of lines.
func extractTotalLines(wcOutput string) (int, error) {
	log.WithField("wc_output", wcOutput).Debug("Extracting total lines from wc output.")
	// Expected format is typically "  <lines> total" (with leading spaces from wc).
	// Some versions might just output the number if it's a single file's count piped differently.
	// We need to robustly find the number associated with "total" or the last number if "total" is not present.

	lines := strings.Split(strings.TrimSpace(wcOutput), "\n")
	if len(lines) == 0 {
		return 0, fmt.Errorf("empty output from wc -l")
	}

	// The last line should contain the total.
	lastLine := strings.TrimSpace(lines[len(lines)-1])
	
	var totalLines int
	// Try to parse " <number> total"
	if _, err := fmt.Sscanf(lastLine, "%d total", &totalLines); err == nil {
		log.WithField("total_lines", totalLines).Debug("Total lines extracted using 'sscanf %d total'.")
		return totalLines, nil
	}
	
	// Fallback: if "total" is not found or Sscanf fails, try to parse the first number on the last line.
	// This handles cases like just " <number>" if only one file was processed by wc -l directly.
	fields := strings.Fields(lastLine)
	if len(fields) > 0 {
		// Take the first field, assuming it's the line count.
		parsedNum, parseErr := strconv.Atoi(fields[0])
		if parseErr == nil {
			log.WithFields(logrus.Fields{"total_lines_fallback": parsedNum, "line_content": lastLine}).Debug("Total lines extracted using fallback parsing.")
			return parsedNum, nil
		}
		// If the first field is not a number, but there's a "total" field, try the field before "total".
		for i, field := range fields {
			if field == "total" && i > 0 {
				parsedNum, parseErr = strconv.Atoi(fields[i-1])
				if parseErr == nil {
					log.WithFields(logrus.Fields{"total_lines_fallback_total_keyword": parsedNum}).Debug("Total lines extracted using fallback with 'total' keyword.")
					return parsedNum, nil
				}
			}
		}
	}
	
	err := fmt.Errorf("could not parse total lines from wc output: '%s'", wcOutput)
	log.WithField("wc_output", wcOutput).Error(err.Error())
	return 0, err
}
