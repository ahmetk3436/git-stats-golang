package api

import (
	"encoding/json"
	"github.com/ahmetk3436/git-stats-golang/pkg/repository"
	"net/http"
	"strconv"
)

type GitlabApi struct {
	Repo *repository.Gitlab
}

func NewGitlabApi(repo *repository.Gitlab) *GitlabApi {
	return &GitlabApi{
		Repo: repo,
	}
}

func (api *GitlabApi) GetAllRepos(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func (api *GitlabApi) GetRepo(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.URL.Query().Get("projectID")

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

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func (api *GitlabApi) GetAllCommits(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.URL.Query().Get("projectID")

	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid projectID parameter", http.StatusBadRequest)
		return
	}

	commits, err := api.Repo.GetAllCommitsGitlab(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse, err := json.Marshal(commits)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
