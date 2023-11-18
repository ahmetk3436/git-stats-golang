package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	storage "github.com/ahmetk3436/git-stats-golang/internal"
	"github.com/ahmetk3436/git-stats-golang/pkg/repository"
)

type GitlabApi struct {
	Repo  *repository.Gitlab
	Redis *storage.RedisClient
}

func NewGitlabApi(repo *repository.Gitlab, redis *storage.RedisClient) *GitlabApi {
	return &GitlabApi{
		Repo:  repo,
		Redis: redis,
	}
}

func (api *GitlabApi) GetAllRepos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	redisRepos, err := api.Redis.Get("gitlab_get_all_repos")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if redisRepos != nil {
		fmt.Println("Redis data")
		w.Write(redisRepos)
		return
	}

	repos, err := api.Repo.GetAllReposGitlab()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(repos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = api.Redis.Set("gitlab_get_all_repos", response, 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Code is here")
	w.Write(response)
}

func (api *GitlabApi) GetRepo(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.URL.Query().Get("projectID")
	w.Header().Set("Content-Type", "application/json")

	redisRepos, err := api.Redis.Get("gitlab_get_repo_" + projectIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if redisRepos != nil {
		fmt.Println("Redis data")
		w.Write(redisRepos)
		return
	}

	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid projectID parameter", http.StatusBadRequest)
		return
	}
	repo, err := api.Repo.GetRepoGitlab(projectID)
	if err != nil {
		http.Error(w, "Can't get project", http.StatusBadRequest)
		return
	}
	jsonResponse, err := json.Marshal(repo)
	if err != nil {
		http.Error(w, "Project can't convert to JSON", http.StatusBadRequest)
		return
	}
	err = api.Redis.Set("gitlab_get_repo_"+projectIDStr, jsonResponse, 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonResponse)
}

func (api *GitlabApi) GetAllCommits(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.URL.Query().Get("projectID")
	w.Header().Set("Content-Type", "application/json")

	redisRepos, err := api.Redis.Get("gitlab_get_commits_" + projectIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if redisRepos != nil {
		fmt.Println("Redis data")
		w.Write(redisRepos)
		return
	}

	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid projectID parameter", http.StatusBadRequest)
		return
	}
	commits, err := api.Repo.GetAllCommitsGitlab(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	jsonResponse, err := json.Marshal(commits)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = api.Redis.Set("gitlab_get_commits_"+projectIDStr, jsonResponse, 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(jsonResponse)
}
