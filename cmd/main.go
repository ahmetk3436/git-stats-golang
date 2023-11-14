package main

import (
	"encoding/json"
	"fmt"
	"github.com/ahmetk3436/git-stats-golang/pkg/repository"
	"io/ioutil"
	"net/http"
)

// CommitStats represents the structure of the stats field in a GitHub commit.
type CommitStats struct {
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
	Total     int `json:"total"`
}

func main() {
	gitClient := repository.Connect("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")
	github, err := repository.NewGithubRepo(gitClient)
	if err != nil {
		panic(err)
	}

	repos, err := github.GetAllRepos()
	if err != nil {
		panic(err)
	}

	for _, repo := range repos {
		if repo.Owner == nil || repo.Name == nil {
			// Handle the case where Owner or Name is nil
			fmt.Println("Error: Owner or Name is nil for a repository.")
			continue
		}

		commits, err := github.GetProjectCommits(*repo.Owner.Login, *repo.Name)
		if err != nil {
			fmt.Printf("Error getting commits for %s/%s: %s\n", *repo.Owner.Login, *repo.Name, err)
			continue
		}

		for _, commit := range commits {
			// Make an HTTP request to the commit URL
			resp, err := http.Get(commit.GetURL())
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			// Read the response body
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			// Parse the commit body as JSON
			var commitInfo map[string]interface{}
			if err := json.Unmarshal(body, &commitInfo); err != nil {
				panic(err)
			}

			// Extract and print commit stats
			if stats, ok := commitInfo["stats"].(map[string]interface{}); ok {
				additions := int(stats["additions"].(float64))
				deletions := int(stats["deletions"].(float64))
				total := int(stats["total"].(float64))

				fmt.Printf("Commit Stats for %s/%s: Additions: %d, Deletions: %d, Total: %d\n",
					*repo.Owner.Login, *repo.Name, additions, deletions, total)
			}

		}
	}
}
