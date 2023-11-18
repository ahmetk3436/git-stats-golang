package api

import (
	"encoding/json"
	"github.com/ahmetk3436/git-stats-golang/pkg/repository"
	"net/http"
	"strconv"
)

type GithubApi struct {
	Repo *repository.GitHubRepo
}

func NewGithubApi(repo *repository.GitHubRepo) *GithubApi {
	return &GithubApi{
		Repo: repo,
	}
}

func (api *GithubApi) GetAllRepos(w http.ResponseWriter, r *http.Request) {
	repos, err := api.Repo.GetAllReposGithub()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(repos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
func (api *GithubApi) GetRepo(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.URL.Query().Get("projectID")

	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid projectID parameter", http.StatusBadRequest)
		return
	}
	repo, err := api.Repo.GetRepoGithub(projectID)
	if err != nil {
		http.Error(w, "Cant get project ", http.StatusBadRequest)
		return
	}
	json, err := json.Marshal(repo)
	if err != nil {
		http.Error(w, "Project can't convert json", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
func (api *GithubApi) GetAllCommits(w http.ResponseWriter, r *http.Request) {
	projectOwner := r.URL.Query().Get("projectOwner")
	repoName := r.URL.Query().Get("repoName")

	commits, err := api.Repo.GetProjectCommitsGithub(projectOwner, repoName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	json, err := json.Marshal(commits)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
