package main

import (
	"fmt"
	storage "github.com/ahmetk3436/git-stats-golang/internal"
	"github.com/ahmetk3436/git-stats-golang/pkg/api"
	"github.com/ahmetk3436/git-stats-golang/pkg/cli"
	"github.com/ahmetk3436/git-stats-golang/pkg/repository"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

var gitlabHost = ""
var gitlabToken = ""
var githubToken = ""
var projectId int64 = 0

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go [cli|api] [flags]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "cli":
		rootCmd.PersistentFlags().StringVar(&gitlabHost, "gitlab-host", "", "Base url for gitlab")
		rootCmd.PersistentFlags().StringVar(&gitlabToken, "gitlab-token", "", "Gitlab Token")
		rootCmd.PersistentFlags().StringVar(&githubToken, "github-token", "", "Github Token")
		rootCmd.PersistentFlags().Int64Var(&projectId, "project-id", 0, "Project ID")
		Execute()
	case "api":
		r := mux.NewRouter()

		headersMiddleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusOK)
					return
				}

				next.ServeHTTP(w, r)
			})
		}

		r.Use(headersMiddleware)
		redis, err := storage.NewRedisClient("redis:6379", "toor")
		if err != nil {
			panic(err)
		}
		// GitHub API
		githubClient := repository.ConnectGithub("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")
		githubRepo, _ := repository.NewGithubRepo(githubClient)
		githubApi := api.NewGithubApi(githubRepo, redis)
		r.HandleFunc("/api/github/commits", githubApi.GetAllCommits)
		r.HandleFunc("/api/github/repo", githubApi.GetRepo)
		r.HandleFunc("/api/github/repos", githubApi.GetAllRepos)
		r.HandleFunc("/api/github/loc", githubApi.GetRepoTotalLinesOfCode)
		// GitLab API
		gitlabHost := "https://gitlab.youandus.net"
		gitlabClient := repository.ConnectGitlab("glpat-FiBYym_JyJPkhsmxVydv", &gitlabHost)
		gitlabRepo := repository.NewGitlabClient(gitlabClient)
		gitlabApi := api.NewGitlabApi(gitlabRepo, redis)
		r.HandleFunc("/api/gitlab/commits", gitlabApi.GetAllCommits)
		r.HandleFunc("/api/gitlab/repo", githubApi.GetRepo)
		r.HandleFunc("/api/gitlab/repos", githubApi.GetAllRepos)
		r.HandleFunc("/api/gitlab/loc", githubApi.GetRepoTotalLinesOfCode)
		r.Handle("/metrics", promhttp.Handler())
		err = http.ListenAndServeTLS(":1323", "cert.pem", "key.pem", r)
		if err != nil {
			panic(err)
		}
	default:
		fmt.Println("Invalid command. Use 'cli' or 'api'.")
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "gitstats",
	Short: "Get Git Stats",
	Run: func(cmd *cobra.Command, args []string) {
		if gitlabToken != "" && gitlabHost != "" && projectId == 0 {
			cli.TakeAllCommitsGitlab(gitlabToken, &gitlabHost)
		}
		if gitlabToken != "" && gitlabHost == "" && projectId == 0 {
			cli.TakeAllCommitsGitlab(gitlabToken, nil)
		}
		if gitlabToken != "" && gitlabHost != "" && projectId != 0 {
			cli.TakeCommitsGitlab(gitlabToken, &gitlabHost, int(projectId))
		}
		if gitlabToken != "" && gitlabHost == "" && projectId != 0 {
			cli.TakeCommitsGitlab(gitlabToken, nil, int(projectId))
		}
		if githubToken != "" && projectId == 0 {
			cli.TakeAllCommitsGithub(githubToken)
		}
		if githubToken != "" && projectId != 0 {
			cli.TakeCommitsGithub(githubToken, projectId)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
