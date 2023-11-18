package api

import (
	"encoding/json"
	"fmt"
	storage "github.com/ahmetk3436/git-stats-golang/internal"
	"github.com/ahmetk3436/git-stats-golang/pkg/repository"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

type GithubApi struct {
	Repo  *repository.GitHubRepo
	Redis *storage.RedisClient
}

func NewGithubApi(repo *repository.GitHubRepo, redis *storage.RedisClient) *GithubApi {
	return &GithubApi{
		Repo:  repo,
		Redis: redis,
	}
}

func (api *GithubApi) GetAllRepos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	redisRepos, err := api.Redis.Get("get_all_repos_github")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if redisRepos != nil {
		println("redis datası")
		w.Write(redisRepos)
		return
	}
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
	err = api.Redis.Set("get_all_repos_github", response, 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	println("kod burada")
	w.Write(response)
}

func (api *GithubApi) GetRepo(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.URL.Query().Get("projectID")
	w.Header().Set("Content-Type", "application/json")
	redisRepos, err := api.Redis.Get("get_repo_" + projectIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if redisRepos != nil {
		println("redis datası")
		w.Write(redisRepos)
		return
	}
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
	err = api.Redis.Set("get_repo_"+projectIDStr, json, 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(json)
}
func (api *GithubApi) GetAllCommits(w http.ResponseWriter, r *http.Request) {
	projectOwner := r.URL.Query().Get("projectOwner")
	repoName := r.URL.Query().Get("repoName")
	w.Header().Set("Content-Type", "application/json")
	redisRepos, err := api.Redis.Get("get_commits_" + projectOwner + " " + repoName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if redisRepos != nil {
		println("redis datası")
		w.Write(redisRepos)
		return
	}

	commits, err := api.Repo.GetProjectCommitsGithub(projectOwner, repoName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	json, err := json.Marshal(commits)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = api.Redis.Set("get_commits"+projectOwner+" "+repoName, json, 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(json)
}
func (api *GithubApi) GetRepoTotalLinesOfCode(w http.ResponseWriter, r *http.Request) {
	repoURL := r.URL.Query().Get("repoUrl")
	w.Header().Set("Content-Type", "application/json")
	redisRepos, err := api.Redis.Get("get_loc_" + repoURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if redisRepos != nil {
		println("redis datası")
		w.Write(redisRepos)
		return
	}
	tempDir, err := os.MkdirTemp("", "temp-repo")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating temp dir: %v", err), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	if err := cloneRepository(repoURL, tempDir); err != nil {
		http.Error(w, fmt.Sprintf("Error cloning repo: %v", err), http.StatusInternalServerError)
		return
	}
	output, err := runCommand("git ls-files | xargs wc -l | tail -n 1", tempDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error running command: %v", err), http.StatusInternalServerError)
		return
	}

	totalLines, err := extractTotalLines(output)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error extracting total lines: %v", err), http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{"totalLines": totalLines}
	jsonResult, err := json.Marshal(result)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
		return
	}
	err = api.Redis.Set("get_loc_"+repoURL, jsonResult, 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(jsonResult)
}

func cloneRepository(repoURL, targetDirectory string) error {
	cmd := exec.Command("git", "clone", repoURL, targetDirectory)
	return cmd.Run()
}

func runCommand(command, cwd string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = cwd
	output, err := cmd.Output()
	return string(output), err
}

func extractTotalLines(output string) (int, error) {
	var totalLines int
	_, err := fmt.Sscanf(output, "%d total", &totalLines)
	if err != nil {
		return 0, err
	}
	return totalLines, nil
}
