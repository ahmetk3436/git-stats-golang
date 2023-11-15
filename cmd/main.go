package main

import (
	"fmt"
	"github.com/ahmetk3436/git-stats-golang/pkg/repository"
)

type Stats struct {
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
	Total     int `json:"total"`
}

type Commit struct {
	AuthorName string `json:"author_name"`
	Stats      *Stats `json:"stats"`
}

type CommitStats struct {
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
	Total     int `json:"total"`
}

func main() {
	gitlabClient := repository.ConnectGitlab("glpat-FiBYym_JyJPkhsmxVydv")
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
			println("Name : " + commit.AuthorName)
			
			formattedStats := fmt.Sprintf("Add : %d Delete : %d Total : %d", commit.Stats.Additions, commit.Stats.Deletions, commit.Stats.Total)
			fmt.Println(formattedStats)
		}

	}
}
