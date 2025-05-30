package api

import (
	"encoding/json"
	// "fmt" // Replaced by logrus for structured logging
	"github.com/ahmetk3436/git-stats-golang/pkg/interfaces"
	"github.com/ahmetk3436/git-stats-golang/pkg/prometheus"
	"net/http"
	"os"
	"strconv"
	"time"

	storage "github.com/ahmetk3436/git-stats-golang/internal"
	appMetrics "github.com/ahmetk3436/git-stats-golang/pkg/prometheus" // Alias for clarity
	"github.com/ahmetk3436/git-stats-golang/pkg/repository"
	"github.com/sirupsen/logrus"
)

// Note: The package-level 'log' variable is assumed to be initialized by another file
// in the same 'api' package (e.g., github_api.go) or this init() block can be uncommented
// if gitlab_api.go could be built or tested independently in a way that `log` isn't already set.
/*
func init() {
	if log == nil { // Check if already initialized to avoid re-setting formatter/output/level.
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(logrus.InfoLevel)
	}
}
*/

// GitlabApi handles API requests related to GitLab.
// It uses a GitService for interacting with GitLab and a RedisClient for caching.
type GitlabApi struct {
	Repo  interfaces.GitService // Service for Git operations (GitLab specific implementation).
	Redis *storage.RedisClient  // Client for Redis caching.
}

// NewGitlabApi creates a new instance of GitlabApi.
// It requires a GitService implementation (e.g., *repository.Gitlab) and a RedisClient.
func NewGitlabApi(gitService interfaces.GitService, redisClient *storage.RedisClient) *GitlabApi {
	log.Info("Creating NewGitlabApi with GitService interface.")
	return &GitlabApi{
		Repo:  gitService,
		Redis: redisClient,
	}
}

// GetAllRepos handles requests to get all repositories for a specified owner or the authenticated user on GitLab.
// It checks cache first and falls back to the GitService.
func (glAPI *GitlabApi) GetAllRepos(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	endpointName := "/api/gitlab/repos"
	// Use a context logger for request-specific fields.
	logCtx := log.WithFields(logrus.Fields{"endpoint": endpointName, "method": r.Method, "provider": "gitlab"})
	logCtx.Info("GetAllRepos request received.")
	w.Header().Set("Content-Type", "application/json")

	ownerQueryParam := r.URL.Query().Get("owner") // For GitLab, owner can be a username or group path.
	logCtx = logCtx.WithField("owner_query", ownerQueryParam)

	var reposFromSource []*common_types.Repository
	var err error
	dataSource := "API"

	redisKey := fmt.Sprintf("gitlab_get_all_repos_%s", ownerQueryParam)
	cachedData, redisErr := glAPI.Redis.Get(redisKey)

	if redisErr == nil && cachedData != nil {
		logCtx.WithField("key", redisKey).Info("Cache hit for GetAllRepos.")
		dataSource = "Redis"
		if err = json.Unmarshal(cachedData, &reposFromSource); err != nil {
			logCtx.WithFields(logrus.Fields{"key": redisKey, "error": err}).Error("Error unmarshalling cached data for GetAllRepos.")
			appMetrics.APICallsTotal.WithLabelValues("gitlab", endpointName, "failure").Inc()
			http.Error(w, "Error processing cached data.", http.StatusInternalServerError)
			return
		}
		w.Write(cachedData) // Serve directly from cache.
	} else {
		if redisErr != nil {
			logCtx.WithFields(logrus.Fields{"key": redisKey, "error": redisErr}).Warn("Redis GET error for GetAllRepos; proceeding to fetch from API.")
		} else if cachedData == nil {
			logCtx.WithField("key", redisKey).Info("Cache miss for GetAllRepos; fetching from API.")
		}
		dataSource = "API"
		fetchedRepos, fetchErr := glAPI.Repo.GetAllRepos(ownerQueryParam)
		if fetchErr != nil {
			logCtx.WithField("error", fetchErr).Error("Error fetching repos from GitLab via GitService.")
			appMetrics.APICallsTotal.WithLabelValues("gitlab", endpointName, "failure").Inc()
			appMetrics.RepositoryFetchesTotal.WithLabelValues("gitlab", "all_repos", "failure").Inc()
			http.Error(w, fetchErr.Error(), http.StatusInternalServerError)
			return
		}
		appMetrics.RepositoryFetchesTotal.WithLabelValues("gitlab", "all_repos", "success").Inc()
		reposFromSource = fetchedRepos

		responseBytes, marshalErr := json.Marshal(reposFromSource)
		if marshalErr != nil {
			logCtx.WithField("error", marshalErr).Error("Error marshalling repos response.")
			appMetrics.APICallsTotal.WithLabelValues("gitlab", endpointName, "failure").Inc()
			http.Error(w, marshalErr.Error(), http.StatusInternalServerError)
			return
		}
		if setErr := glAPI.Redis.Set(redisKey, responseBytes, 3600); setErr != nil { // Cache for 1 hour.
			logCtx.WithFields(logrus.Fields{"key": redisKey, "error": setErr}).Error("Redis SET error for GetAllRepos.")
		}
		w.Write(responseBytes)
	}

	duration := time.Since(startTime).Seconds()
	appMetrics.APICallDuration.WithLabelValues("gitlab", endpointName).Observe(duration)
	appMetrics.APICallsTotal.WithLabelValues("gitlab", endpointName, "success").Inc()
	logCtx.WithFields(logrus.Fields{"duration_seconds": duration, "source": dataSource, "items_count": len(reposFromSource)}).Info("GetAllRepos request processed successfully.")
}

// GetRepo handles requests to get a specific GitLab repository by ID or "namespace/path" string.
// It checks cache first and falls back to the GitService.
func (glAPI *GitlabApi) GetRepo(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	endpointName := "/api/gitlab/repo"
	// "projectID" is the query parameter, but it can be an ID or a path for GitLab.
	repoIdentifierQuery := r.URL.Query().Get("projectID")
	logCtx := log.WithFields(logrus.Fields{"endpoint": endpointName, "method": r.Method, "identifier_query": repoIdentifierQuery, "provider": "gitlab"})
	logCtx.Info("GetRepo request received.")
	w.Header().Set("Content-Type", "application/json")

	if repoIdentifierQuery == "" {
		logCtx.Error("Missing projectID query parameter.")
		appMetrics.APICallsTotal.WithLabelValues("gitlab", endpointName, "failure").Inc()
		http.Error(w, "projectID query parameter is required.", http.StatusBadRequest)
		return
	}

	var repoData *common_types.Repository
	var err error
	dataSource := "API"

	redisKey := "gitlab_get_repo_" + repoIdentifierQuery
	cachedData, redisErr := glAPI.Redis.Get(redisKey)

	if redisErr == nil && cachedData != nil {
		logCtx.WithField("key", redisKey).Info("Cache hit for GetRepo.")
		dataSource = "Redis"
		if err = json.Unmarshal(cachedData, &repoData); err != nil {
			logCtx.WithFields(logrus.Fields{"key": redisKey, "error": err}).Error("Error unmarshalling cached data for GetRepo.")
			appMetrics.APICallsTotal.WithLabelValues("gitlab", endpointName, "failure").Inc()
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
		// GitLab project IDs are integers. Paths are strings.
		parsedID, parseErr := strconv.Atoi(repoIdentifierQuery)
		if parseErr == nil {
			identifierToFetch = parsedID // Use int ID if parsing succeeds.
		} else {
			identifierToFetch = repoIdentifierQuery // Otherwise, assume it's a "namespace/path" string.
			logCtx.WithField("input", repoIdentifierQuery).Debug("Interpreting projectID as namespace/path string due to Atoi failure.")
		}

		fetchedRepo, fetchErr := glAPI.Repo.GetRepo(identifierToFetch)
		if fetchErr != nil {
			logCtx.WithFields(logrus.Fields{"identifier": identifierToFetch, "error": fetchErr}).Error("Error fetching repo from GitLab via GitService.")
			appMetrics.APICallsTotal.WithLabelValues("gitlab", endpointName, "failure").Inc()
			appMetrics.RepositoryFetchesTotal.WithLabelValues("gitlab", "single_repo", "failure").Inc()
			http.Error(w, "Cannot get project: "+fetchErr.Error(), http.StatusBadRequest)
			return
		}
		appMetrics.RepositoryFetchesTotal.WithLabelValues("gitlab", "single_repo", "success").Inc()
		repoData = fetchedRepo

		responseBytes, marshalErr := json.Marshal(repoData)
		if marshalErr != nil {
			logCtx.WithField("error", marshalErr).Error("Error marshalling repo response.")
			appMetrics.APICallsTotal.WithLabelValues("gitlab", endpointName, "failure").Inc()
			http.Error(w, "Error marshalling project data.", http.StatusInternalServerError)
			return
		}
		if setErr := glAPI.Redis.Set(redisKey, responseBytes, 3600); setErr != nil { // Cache for 1 hour.
			logCtx.WithFields(logrus.Fields{"key": redisKey, "error": setErr}).Error("Redis SET error for GetRepo.")
		}
		w.Write(responseBytes)
	}

	duration := time.Since(startTime).Seconds()
	appMetrics.APICallDuration.WithLabelValues("gitlab", endpointName).Observe(duration)
	appMetrics.APICallsTotal.WithLabelValues("gitlab", endpointName, "success").Inc()
	logCtx.WithFields(logrus.Fields{"duration_seconds": duration, "source": dataSource, "repo_name": repoData.Name}).Info("GetRepo request processed successfully.")
}

// GetAllCommits handles requests to get all commits for a specific GitLab repository.
// Repository is identified by 'projectID' query parameter (ID or namespace/path).
// It checks cache first and falls back to the GitService.
func (glAPI *GitlabApi) GetAllCommits(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	endpointName := "/api/gitlab/commits"
	repoIdentifierQuery := r.URL.Query().Get("projectID") // Can be ID or namespace/path.
	logCtx := log.WithFields(logrus.Fields{"endpoint": endpointName, "method": r.Method, "identifier_query": repoIdentifierQuery, "provider": "gitlab"})
	logCtx.Info("GetAllCommits request received.")
	w.Header().Set("Content-Type", "application/json")

	if repoIdentifierQuery == "" {
		logCtx.Error("Missing projectID query parameter.")
		appMetrics.APICallsTotal.WithLabelValues("gitlab", endpointName, "failure").Inc()
		http.Error(w, "projectID query parameter is required.", http.StatusBadRequest)
		return
	}

	var commitsFromSource []*common_types.Commit
	var err error
	dataSource := "API"

	redisKey := "gitlab_get_commits_" + repoIdentifierQuery
	cachedData, redisErr := glAPI.Redis.Get(redisKey)

	if redisErr == nil && cachedData != nil {
		logCtx.WithField("key", redisKey).Info("Cache hit for GetAllCommits.")
		dataSource = "Redis"
		if err = json.Unmarshal(cachedData, &commitsFromSource); err != nil {
			logCtx.WithFields(logrus.Fields{"key": redisKey, "error": err}).Error("Error unmarshalling cached data for GetAllCommits.")
			appMetrics.APICallsTotal.WithLabelValues("gitlab", endpointName, "failure").Inc()
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

		var identifierToFetch interface{}
		parsedID, parseErr := strconv.Atoi(repoIdentifierQuery)
		if parseErr == nil {
			identifierToFetch = parsedID
		} else {
			identifierToFetch = repoIdentifierQuery
			logCtx.WithField("input", repoIdentifierQuery).Debug("Interpreting projectID as namespace/path string due to Atoi failure.")
		}

		// TODO: Populate CommitListOptions from query params if needed (e.g., branch, page).
		commitOpts := &interfaces.CommitListOptions{PerPage: 100} // Default: 100 commits.

		fetchedCommits, fetchErr := glAPI.Repo.GetProjectCommits(identifierToFetch, commitOpts)
		if fetchErr != nil {
			logCtx.WithFields(logrus.Fields{"identifier": identifierToFetch, "error": fetchErr}).Error("Error fetching commits from GitLab via GitService.")
			appMetrics.APICallsTotal.WithLabelValues("gitlab", endpointName, "failure").Inc()
			appMetrics.RepositoryFetchesTotal.WithLabelValues("gitlab", "commits", "failure").Inc()
			http.Error(w, fetchErr.Error(), http.StatusInternalServerError)
			return
		}
		appMetrics.RepositoryFetchesTotal.WithLabelValues("gitlab", "commits", "success").Inc()
		commitsFromSource = fetchedCommits

		responseBytes, marshalErr := json.Marshal(commitsFromSource)
		if marshalErr != nil {
			logCtx.WithField("error", marshalErr).Error("Error marshalling commits response.")
			appMetrics.APICallsTotal.WithLabelValues("gitlab", endpointName, "failure").Inc()
			http.Error(w, marshalErr.Error(), http.StatusInternalServerError)
			return
		}
		if setErr := glAPI.Redis.Set(redisKey, responseBytes, 3600); setErr != nil { // Cache for 1 hour.
			logCtx.WithFields(logrus.Fields{"key": redisKey, "error": setErr}).Error("Redis SET error for GetAllCommits.")
		}
		w.Write(responseBytes)
	}

	duration := time.Since(startTime).Seconds()
	appMetrics.APICallDuration.WithLabelValues("gitlab", endpointName).Observe(duration)
	appMetrics.APICallsTotal.WithLabelValues("gitlab", endpointName, "success").Inc()
	logCtx.WithFields(logrus.Fields{"duration_seconds": duration, "source": dataSource, "items_count": len(commitsFromSource)}).Info("GetAllCommits request processed successfully.")
}

// Note: GetRepoTotalLinesOfCode is currently only implemented in github_api.go.
// If it were to be implemented for GitLab, similar logging and metrics would be added.
// A GetContributors endpoint for GitLab would also need to be added if required by the frontend.
