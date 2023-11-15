package main

import (
	"fmt"
	"github.com/ahmetk3436/git-stats-golang/pkg/cli"
	"github.com/spf13/cobra"
	"os"
)

var gitlabHost = ""
var gitlabToken = ""
var githubToken = ""
var projectId int64 = 0

func main() {
	rootCmd.PersistentFlags().StringVar(&gitlabHost, "gitlab-host", "", "Base url for gitlab")
	rootCmd.PersistentFlags().StringVar(&gitlabToken, "gitlab-token", "", "Gitlab Token")
	rootCmd.PersistentFlags().StringVar(&githubToken, "github-token", "", "Github Token")
	rootCmd.PersistentFlags().Int64Var(&projectId, "project-id", 0, "Project ID")
	Execute()
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
