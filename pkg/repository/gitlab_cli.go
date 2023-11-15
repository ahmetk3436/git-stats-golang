package repository

import "fmt"

func TakeAllCommitsGitlab(token string, host *string) {
	//"glpat-FiBYym_JyJPkhsmxVydv"
	gitlabClient := ConnectGitlab(token, host)
	client := NewGitlabClient(gitlabClient)
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
			println("Name : " + commit.AuthorName)

			formattedStats := fmt.Sprintf("Add : %d Delete : %d Total : %d", commit.Stats.Additions, commit.Stats.Deletions, commit.Stats.Total)
			fmt.Println(formattedStats)
		}

	}
}
func TakeCommitsGitlab(token string, host *string, projectID int) {
	allCommits := make(map[string]struct {
		Add    int
		Delete int
		Total  int
	})

	gitlabClient := ConnectGitlab(token, host)
	client := NewGitlabClient(gitlabClient)
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
