package main

// Note on certificates:
// The files cert.pem and key.pem in the cmd/ directory are intended for local development HTTPS.
// They are likely self-signed. Do not use these specific files in a production environment.
// For production, use properly provisioned SSL/TLS certificates.

import (
	"fmt"
	storage "github.com/ahmetk3436/git-stats-golang/internal"
	"github.com/ahmetk3436/git-stats-golang/pkg/api"
	"github.com/ahmetk3436/git-stats-golang/pkg/cli"
	appMetrics "github.com/ahmetk3436/git-stats-golang/pkg/prometheus" // Alias for clarity
	"github.com/ahmetk3436/git-stats-golang/pkg/repository"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	// "log" // Standard log package replaced by logrus for structured logging.
	"net/http"
	"os"
)

// Global variables to hold flag values. These are populated by Cobra.
var (
	gitlabHostVar  string // Stores the GitLab host URL provided via flag or env.
	gitlabTokenVar string // Stores the GitLab token provided via flag or env.
	githubTokenVar string // Stores the GitHub token provided via flag or env.
	projectIDVar   int64  // Stores the Project ID (if any) provided via flag.
)

// log is a global logrus instance used for structured logging throughout the application.
var log = logrus.New()

// init function is called when the package is initialized.
// It sets up the global logger configuration and initializes Prometheus metrics.
func init() {
	// Configure logrus for JSON formatted output.
	log.SetFormatter(&logrus.JSONFormatter{})
	// Output logs to standard output.
	log.SetOutput(os.Stdout)
	// Set the default logging level to Info. Debug messages will be suppressed unless level is changed.
	log.SetLevel(logrus.InfoLevel) // TODO: Consider making log level configurable via flag/env.

	// Initialize and register custom Prometheus metrics.
	appMetrics.InitMetrics()
	log.Info("Logger and Prometheus metrics initialized.")
}

// getEnv retrieves an environment variable by key.
// If the variable is not set or empty, it returns the provided fallback string.
// It also logs whether the environment variable was found or if the fallback is being used.
func getEnv(key, fallback string) string {
	if value, isSet := os.LookupEnv(key); isSet && value != "" {
		log.WithFields(logrus.Fields{"key": key, "source": "environment"}).Debug("Environment variable used.")
		return value
	}
	log.WithFields(logrus.Fields{"key": key, "source": "fallback"}).Debug("Fallback value used for environment variable.")
	return fallback
}

// main is the entry point of the application.
// It parses command-line arguments to determine if the application should run in CLI or API mode.
func main() {
	log.Info("Application starting...")

	if len(os.Args) < 2 {
		log.Error("Insufficient arguments. Usage: go run main.go [cli|api] [flags]")
		fmt.Println("Usage: go run main.go [cli|api] [flags]") // Also print to console for user.
		os.Exit(1)
	}

	// Determine application mode (CLI or API) from the first argument.
	appMode := os.Args[1]
	log.WithField("mode", appMode).Info("Application mode selected.")

	switch appMode {
	case "cli":
		// Setup Cobra command flags for CLI mode.
		// Values are bound to the global flag variables (gitlabHostVar, etc.).
		// getEnv is used to provide default values from environment variables if flags are not explicitly set.
		rootCmd.PersistentFlags().StringVar(&gitlabHostVar, "gitlab-host", getEnv("GITLAB_HOST", ""), "Base URL for GitLab (e.g., https://gitlab.example.com). Can also be set via GITLAB_HOST env var.")
		rootCmd.PersistentFlags().StringVar(&gitlabTokenVar, "gitlab-token", getEnv("GITLAB_TOKEN", ""), "GitLab Personal Access Token. Can also be set via GITLAB_TOKEN env var.")
		rootCmd.PersistentFlags().StringVar(&githubTokenVar, "github-token", getEnv("GITHUB_TOKEN", ""), "GitHub Personal Access Token. Can also be set via GITHUB_TOKEN env var.")
		rootCmd.PersistentFlags().Int64Var(&projectIDVar, "project-id", 0, "Optional Project ID for specific actions (applies to both GitHub and GitLab where appropriate).")

		log.Info("Executing CLI mode.")
		Execute() // Calls Cobra's command execution.
	case "api":
		log.Info("Starting API mode.")
		// API mode setup: Read configurations, initialize services, and start the HTTP server.

		// Configuration for services, preferring environment variables with sensible defaults.
		redisHost := getEnv("REDIS_HOST", "redis:6379")
		redisPassword := getEnv("REDIS_PASSWORD", "toor") // TODO: Ensure 'toor' is a dev-only default.
		// Default tokens below are for example/development.
		// In production, these must be securely managed and not have hardcoded fallbacks if they are sensitive.
		// Ensure GITHUB_TOKEN and GITLAB_TOKEN are set in the environment for production.
		githubToken := getEnv("GITHUB_TOKEN", "") // No hardcoded fallback for actual tokens.
		gitlabToken := getEnv("GITLAB_TOKEN", "") // No hardcoded fallback.
		gitlabAPIHost := getEnv("GITLAB_HOST", "https://gitlab.com") // Default to GitLab.com if not specified.
		// frontendGitHubToken is no longer used as token is not sent to frontend.

		// Create a new Gorilla Mux router.
		router := mux.NewRouter()

		// API endpoint for frontend configuration (non-sensitive). Token is no longer exposed.
		router.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
			log.WithFields(logrus.Fields{"path": r.URL.Path}).Info("Handling /api/config request.")
			w.Header().Set("Content-Type", "application/json")
			if _, err := fmt.Fprintf(w, `{}`); err != nil { // Send empty JSON or other non-sensitive config.
				log.WithFields(logrus.Fields{"path": r.URL.Path, "error": err}).Error("Error writing response for /api/config.")
			}
		})

		// Middleware for CORS and common security headers.
		headersMiddleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.WithFields(logrus.Fields{"method": r.Method, "uri": r.RequestURI}).Debug("Received API request.")

				// CORS Headers.
				// For production, Access-Control-Allow-Origin should be restricted to specific frontend domain(s).
				w.Header().Set("Access-Control-Allow-Origin", getEnv("CORS_ALLOWED_ORIGIN", "*")) // Default to * if not set.
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // Authorization for potential future token-based auth from frontend to backend.

				// Basic Security Headers.
				w.Header().Set("X-Content-Type-Options", "nosniff") // Prevents MIME sniffing.
				w.Header().Set("X-Frame-Options", "DENY")           // Prevents clickjacking.
				// Consider adding Content-Security-Policy and Strict-Transport-Security for enhanced security.

				// Handle OPTIONS preflight requests for CORS.
				if r.Method == http.MethodOptions {
					log.Debug("Handling OPTIONS preflight request.")
					w.WriteHeader(http.StatusOK)
					return
				}
				next.ServeHTTP(w, r)
			})
		}
		router.Use(headersMiddleware)

		// Initialize Redis client.
		redisClient, err := storage.NewRedisClient(redisHost, redisPassword)
		if err != nil {
			log.WithFields(logrus.Fields{"redis_host": redisHost, "error": err}).Fatal("Failed to connect to Redis.")
		}
		log.Info("Successfully connected to Redis.")

		// Setup GitHub API service if token is provided.
		if githubToken != "" {
			log.Info("Initializing GitHub service.")
			ghSdkClient := repository.ConnectGithub(githubToken)    // Creates underlying GitHub SDK client.
			ghRepoService, err := repository.NewGithubRepo(ghSdkClient) // Wraps SDK client with our GitService implementation.
			if err != nil {
				log.WithField("error", err).Fatal("Failed to create GitHubRepo service.")
			}
			githubAPIHandler := api.NewGithubApi(ghRepoService, redisClient) // Injects GitService.

			// Register GitHub API routes.
			ghRouter := router.PathPrefix("/api/github").Subrouter()
			ghRouter.HandleFunc("/commits", githubAPIHandler.GetAllCommits).Methods(http.MethodGet, http.MethodOptions)
			ghRouter.HandleFunc("/repo", githubAPIHandler.GetRepo).Methods(http.MethodGet, http.MethodOptions)
			ghRouter.HandleFunc("/repos", githubAPIHandler.GetAllRepos).Methods(http.MethodGet, http.MethodOptions)
			ghRouter.HandleFunc("/loc", githubAPIHandler.GetRepoTotalLinesOfCode).Methods(http.MethodGet, http.MethodOptions)
			ghRouter.HandleFunc("/contributors", githubAPIHandler.GetContributors).Methods(http.MethodGet, http.MethodOptions)
			log.Info("GitHub API routes registered.")
		} else {
			log.Warn("GITHUB_TOKEN not provided. GitHub API routes will not be available.")
		}

		// Setup GitLab API service if token is provided.
		if gitlabToken != "" {
			log.Info("Initializing GitLab service.")
			glSdkClient, err := repository.ConnectGitlab(gitlabToken, &gitlabAPIHost) // Creates underlying GitLab SDK client.
			if err != nil {
				log.WithFields(logrus.Fields{"gitlab_host": gitlabAPIHost, "error": err}).Fatal("Failed to connect to GitLab client/service.")
			}
			glRepoService, err := repository.NewGitlabClient(glSdkClient) // Wraps SDK client with our GitService implementation.
			if err != nil {
				log.WithField("error", err).Fatal("Failed to create GitLabClient service.")
			}
			gitlabAPIHandler := api.NewGitlabApi(glRepoService, redisClient) // Injects GitService.

			// Register GitLab API routes.
			glRouter := router.PathPrefix("/api/gitlab").Subrouter()
			glRouter.HandleFunc("/commits", gitlabAPIHandler.GetAllCommits).Methods(http.MethodGet, http.MethodOptions)
			glRouter.HandleFunc("/repo", gitlabAPIHandler.GetRepo).Methods(http.MethodGet, http.MethodOptions)
			glRouter.HandleFunc("/repos", gitlabAPIHandler.GetAllRepos).Methods(http.MethodGet, http.MethodOptions)
			// TODO: Implement GetRepoTotalLinesOfCode and GetContributors for GitLab if needed.
			// glRouter.HandleFunc("/loc", gitlabAPIHandler.GetRepoTotalLinesOfCode).Methods(http.MethodGet, http.MethodOptions)
			// glRouter.HandleFunc("/contributors", gitlabAPIHandler.GetContributors).Methods(http.MethodGet, http.MethodOptions)
			log.Info("GitLab API routes registered.")
		} else {
			log.Warn("GITLAB_TOKEN not provided. GitLab API routes will not be available.")
		}

		// Prometheus metrics endpoint.
		router.Handle("/metrics", promhttp.Handler())
		log.Info("Metrics endpoint /metrics registered.")

		// Start HTTP server.
		serverAddress := getEnv("SERVER_ADDRESS", ":1323") // Allow server address to be configured.
		log.Infof("Starting server on %s", serverAddress)
		// For production, ListenAndServeTLS with valid certificates loaded securely is recommended.
		// Example for HTTPS: err = http.ListenAndServeTLS(serverAddress, "path/to/cert.pem", "path/to/key.pem", router)
		if err := http.ListenAndServe(serverAddress, router); err != nil { // Using HTTP for simplicity here.
			log.WithField("error", err).Fatalf("Failed to start server on %s.", serverAddress)
		}

	default:
		log.WithField("command", appMode).Error("Invalid command. Use 'cli' or 'api'.")
		fmt.Println("Invalid command. Use 'cli' or 'api'.") // Also print to console.
		os.Exit(1)
	}
	log.Info("Application ended.")
}

// rootCmd represents the base command when called without any subcommands.
// It's the entry point for the CLI mode of the application.
var rootCmd = &cobra.Command{
	Use:   "gitstats",
	Short: "A CLI tool to fetch Git statistics from providers like GitHub and GitLab.",
	Long: `gitstats is a command-line tool that interacts with Git provider APIs
to retrieve repository information, commit history, and other statistics.
It can be configured using flags or environment variables for tokens and host URLs.`,
	Run: dispatchCliCommands, // dispatchCliCommands contains the logic for handling CLI actions.
}

// dispatchCliCommands is the core function executed when the CLI mode is run.
// It uses the flag values (populated by Cobra from command-line flags or environment variables)
// to determine which Git provider actions to perform (e.g., fetching commits).
func dispatchCliCommands(cmd *cobra.Command, args []string) {
	log.Info("Dispatching CLI command based on provided flags...")

	// Retrieve flag values. Cobra ensures these are populated.
	gitlabHost := gitlabHostVar  // Host for GitLab (if self-managed).
	gitlabToken := gitlabTokenVar // Token for GitLab.
	githubToken := githubTokenVar // Token for GitHub.
	projectID := projectIDVar    // Optional project ID for specific actions.

	// GitLab actions processing block.
	if gitlabToken != "" {
		log.WithFields(logrus.Fields{"provider": "gitlab"}).Info("GitLab token provided. Processing GitLab actions...")
		var effectiveGitlabHost *string // Pointer to allow nil for default GitLab.com.
		if gitlabHost != "" {
			effectiveGitlabHost = &gitlabHost
			log.WithFields(logrus.Fields{"provider": "gitlab", "customHost": gitlabHost}).Info("Using custom GitLab host.")
		} else {
			log.WithFields(logrus.Fields{"provider": "gitlab"}).Info("Using default GitLab host (GitLab.com).")
		}

		if projectID == 0 {
			// Fetch all commits for all accessible projects on GitLab.
			log.Info("Action: Fetch all commits for all GitLab projects.")
			cli.TakeAllCommitsGitlab(gitlabToken, effectiveGitlabHost)
		} else {
			// Fetch commits for a specific project ID on GitLab.
			log.WithField("projectID", projectID).Info("Action: Fetch commits for specific GitLab project.")
			cli.TakeCommitsGitlab(gitlabToken, effectiveGitlabHost, int(projectID)) // cli functions expect int for projectID.
		}
	}

	// GitHub actions processing block.
	if githubToken != "" {
		log.WithFields(logrus.Fields{"provider": "github"}).Info("GitHub token provided. Processing GitHub actions...")
		// Note: Current cli.TakeAllCommitsGithub and cli.TakeCommitsGithub do not support custom GitHub Enterprise hosts.
		// They would need refactoring to accept a host parameter for that functionality.
		if projectID == 0 {
			// Fetch all commits for all accessible projects on GitHub.
			log.Info("Action: Fetch all commits for all GitHub projects.")
			cli.TakeAllCommitsGithub(githubToken)
		} else {
			// Fetch commits for a specific project ID on GitHub.
			log.WithField("projectID", projectID).Info("Action: Fetch commits for specific GitHub project.")
			cli.TakeCommitsGithub(githubToken, projectID) // Assumes projectID is int64 as per flag type.
		}
	}

	// Inform user if no tokens were provided, hence no action taken.
	if gitlabToken == "" && githubToken == "" {
		log.Warn("No GitLab or GitHub token provided. No CLI actions will be performed.")
		fmt.Println("Please provide a GitLab or GitHub token using flags (e.g., --github-token YOUR_TOKEN) or environment variables.")
	}
}

// Execute is called by main.main() to run the Cobra root command.
// It handles command parsing and execution. Errors are logged and result in process exit.
func Execute() {
	log.Info("Executing root command via Cobra.")
	if err := rootCmd.Execute(); err != nil {
		log.WithField("error", err).Error("Error executing command via Cobra.")
		// Cobra typically prints the error to stderr itself.
		os.Exit(1) // Exit with error status.
	}
}
