package repository

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func TakeAllCommitsGithub(token string) {
	//"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"
	commitStats := make(map[string]int)
	gitClient := ConnectGithub(token)
	github, err := NewGithubRepo(gitClient)
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

				avatarUrl := author["avatar_url"].(string)

				println("Login User : " + login + " Avatar url : " + avatarUrl)
				fmt.Printf("Commit Stats for %s/%s: Additions: %d, Deletions: %d, Total: %d\n",
					*repo.Owner.Login, *repo.Name, additions, deletions, total)
				commitStats[login] += additions
			}
		}
		for key, value := range commitStats {
			println("Key : " + key + " Value : " + fmt.Sprintf("%d", value))
		}

		break
	}
}

func TakeCommitsGithub(token string, projectID int64) {
	//"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"
	commitStats := make(map[string]int)
	gitClient := ConnectGithub(token)
	github, err := NewGithubRepo(gitClient)
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

			avatarUrl := author["avatar_url"].(string)

			println("Login User : " + login + " Avatar url : " + avatarUrl)
			fmt.Printf("Commit Stats for %s/%s: Additions: %d, Deletions: %d, Total: %d\n",
				*repo.Owner.Login, *repo.Name, additions, deletions, total)
			commitStats[login] += additions
		}
	}
	for key, value := range commitStats {
		println("Key : " + key + " Value : " + fmt.Sprintf("%d", value))
	}
}
