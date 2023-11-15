package cli

import (
	"fmt"
	"github.com/ahmetk3436/git-stats-golang/pkg/repository"
)

func TakeAllCommitsGitlab(token string, host *string) {
	//"glpat-FiBYym_JyJPkhsmxVydv"
	allCommits := make(map[string]struct {
		Add    int
		Delete int
		Total  int
	})
	gitlabClient := repository.ConnectGitlab(token, host)
	client := repository.NewGitlabClient(gitlabClient)
	projects, err := client.GetAllReposGitlab()
	if err != nil {
		panic(err)
	}
	for _, project := range projects {
		commits, err := client.GetAllCommitsGitlab(project.ID)
		if err != nil {
			panic(err)
		}
		for _, commit := range commits {
			myStats := allCommits[commit.AuthorName]

			myStats.Add += commit.Stats.Additions
			myStats.Delete += commit.Stats.Deletions
			myStats.Total += commit.Stats.Total

			allCommits[commit.AuthorName] = myStats
		}
	}
	for user, stats := range allCommits {
		fmt.Printf("User: %s, Add: %d, Delete: %d, Total: %d\n", user, stats.Add, stats.Delete, stats.Total)
	}
}
func TakeCommitsGitlab(token string, host *string, projectID int) {
	allCommits := make(map[string]struct {
		Add    int
		Delete int
		Total  int
	})

	gitlabClient := repository.ConnectGitlab(token, host)
	client := repository.NewGitlabClient(gitlabClient)
	commits, err := client.GetAllCommitsGitlab(projectID)
	if err != nil {
		panic(err)
	}

	for _, commit := range commits {
		myStats := allCommits[commit.AuthorName]
		myStats.Add += commit.Stats.Additions
		myStats.Delete += commit.Stats.Deletions
		myStats.Total += commit.Stats.Total
		allCommits[commit.AuthorName] = myStats
	}

	for user, stats := range allCommits {
		fmt.Printf("User: %s, Add: %d, Delete: %d, Total: %d\n", user, stats.Add, stats.Delete, stats.Total)
	}
}
