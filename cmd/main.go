package main

import (
	"fmt"
	"github.com/ahmetk3436/git-stats-golang/pkg/api"
	"github.com/ahmetk3436/git-stats-golang/pkg/cli"
	"github.com/ahmetk3436/git-stats-golang/pkg/repository"
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
		mux := http.NewServeMux()
		githubClient := repository.ConnectGithub("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")
		githubRepo, _ := repository.NewGithubRepo(githubClient)
		githubApi := api.NewGithubApi(githubRepo)
		gitlabHost := "https://gitlab.youandus.net"
		gitlabClient := repository.ConnectGitlab("glpat-FiBYym_JyJPkhsmxVydv", &gitlabHost)
		gitlabRepo := repository.NewGitlabClient(gitlabClient)
		gitlabApi := api.NewGitlabApi(gitlabRepo)
		// GITHUB
		mux.HandleFunc("/api/github/commits", githubApi.GetAllCommits)
		mux.HandleFunc("/api/github/repos", githubApi.GetRepo)
		mux.HandleFunc("/api/github/repos", githubApi.GetAllRepos)
		// GITLAB
		mux.HandleFunc("/api/gitlab/commits", gitlabApi.GetAllCommits)
		mux.HandleFunc("/api/gitlab/repo", githubApi.GetRepo)
		mux.HandleFunc("/api/gitlab/repos", githubApi.GetAllRepos)
		err := http.ListenAndServe(":1323", mux)
		if err != nil {
			panic(err)
		}
	default:
		fmt.Println("Invalid command. Use 'cli' or 'api'.")
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "git-stats",
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
