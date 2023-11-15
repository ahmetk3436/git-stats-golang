package cli

import (
	"encoding/json"
	"fmt"
	"github.com/ahmetk3436/git-stats-golang/pkg/repository"
	"io/ioutil"
	"net/http"
)

func TakeAllCommitsGithub(token string) {
	//"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"
	commitStats := make(map[string]struct {
		Add    int
		Delete int
		Total  int
	})
	gitClient := repository.ConnectGithub(token)
	github, err := repository.NewGithubRepo(gitClient)
	if err != nil {
		panic(err)
	}

	repos, err := github.GetAllReposGithub()
	if err != nil {
		panic(err)
	}

	for _, repo := range repos {
		if repo.Owner == nil || repo.Name == nil {
			fmt.Println("Error: Owner or Name is nil for a repository.")
			continue
		}

		commits, err := github.GetProjectCommitsGithub(*repo.Owner.Login, *repo.Name)
		if err != nil {
			fmt.Printf("Error getting commits for %s/%s: %s\n", *repo.Owner.Login, *repo.Name, err)
			continue
		}

		for _, commit := range commits {

			req, err := http.NewRequest("GET", commit.GetURL(), nil)
			if err != nil {
				panic(err)
			}

			// Set the Authorization header
			req.Header.Set("Authorization", "Bearer "+token)

			// Make the request
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			var commitInfo map[string]interface{}
			if err := json.Unmarshal(body, &commitInfo); err != nil {
				panic(err)
			}

			if stats, ok := commitInfo["stats"].(map[string]interface{}); ok {
				additions := int(stats["additions"].(float64))

				deletions := int(stats["deletions"].(float64))

				total := int(stats["total"].(float64))

				author := commitInfo["author"].(map[string]interface{})

				login := author["login"].(string)

				stats := commitStats[login]
				stats.Add += additions
				stats.Total += total
				stats.Delete += deletions
				commitStats[login] = stats
			}
		}
		for user, stats := range commitStats {
			println("User : " + user + fmt.Sprintf(" Add : %d Delete : %d Total : %d", stats.Add, stats.Delete, stats.Total))
		}
	}
}

func TakeCommitsGithub(token string, projectID int64) {
	//"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"
	commitStats := make(map[string]struct {
		Add    int
		Delete int
		Total  int
	})
	gitClient := repository.ConnectGithub(token)
	github, err := repository.NewGithubRepo(gitClient)
	if err != nil {
		panic(err)
	}

	repo, err := github.GetRepoGithub(projectID)
	if err != nil {
		panic(err)
	}
	if repo.Owner == nil || repo.Name == nil {
		fmt.Println("Error: Owner or Name is nil for a repository.")
	}

	commits, err := github.GetProjectCommitsGithub(*repo.Owner.Login, *repo.Name)
	if err != nil {
		fmt.Printf("Error getting commits for %s/%s: %s\n", *repo.Owner.Login, *repo.Name, err)
	}

	for _, commit := range commits {

		req, err := http.NewRequest("GET", commit.GetURL(), nil)
		if err != nil {
			panic(err)
		}

		// Set the Authorization header
		req.Header.Set("Authorization", "Bearer "+token)

		// Make the request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var commitInfo map[string]interface{}
		if err := json.Unmarshal(body, &commitInfo); err != nil {
			panic(err)
		}

		if stats, ok := commitInfo["stats"].(map[string]interface{}); ok {
			additions := int(stats["additions"].(float64))

			deletions := int(stats["deletions"].(float64))

			total := int(stats["total"].(float64))

			author := commitInfo["author"].(map[string]interface{})

			login := author["login"].(string)

			stats := commitStats[login]
			stats.Add += additions
			stats.Total += total
			stats.Delete += deletions
			commitStats[login] = stats
		}
	}
	for user, stats := range commitStats {
		println("User : " + user + fmt.Sprintf(" Add : %d Delete : %d Total : %d", stats.Add, stats.Delete, stats.Total))
	}
}
