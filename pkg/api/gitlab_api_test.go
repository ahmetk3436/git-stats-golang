package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ahmetk3436/git-stats-golang/pkg/common_types"
	"github.com/ahmetk3436/git-stats-golang/pkg/interfaces"
	// storage "github.com/ahmetk3436/git-stats-golang/internal" // For concrete type if not using interface for Redis
)

// --- Mocks (can be shared from github_api_test.go if in the same package, or defined in a common test utility file) ---

// MockGitService is a mock implementation of interfaces.GitService (redefined here for clarity, ideally shared)
type MockGitService struct {
	GetAllReposFunc         func(owner string) ([]*common_types.Repository, error)
	GetRepoFunc             func(identifier interface{}) (*common_types.Repository, error)
	GetProjectCommitsFunc   func(repoIdentifier interface{}, options *interfaces.CommitListOptions) ([]*common_types.Commit, error)
	GetRepoContributorsFunc func(repoIdentifier interface{}) ([]*common_types.User, error)
}

func (m *MockGitService) GetAllRepos(owner string) ([]*common_types.Repository, error) {
	if m.GetAllReposFunc != nil {
		return m.GetAllReposFunc(owner)
	}
	return nil, errors.New("GetAllReposFunc not implemented")
}

func (m *MockGitService) GetRepo(identifier interface{}) (*common_types.Repository, error) {
	if m.GetRepoFunc != nil {
		return m.GetRepoFunc(identifier)
	}
	return nil, errors.New("GetRepoFunc not implemented")
}

func (m *MockGitService) GetProjectCommits(repoIdentifier interface{}, options *interfaces.CommitListOptions) ([]*common_types.Commit, error) {
	if m.GetProjectCommitsFunc != nil {
		return m.GetProjectCommitsFunc(repoIdentifier, options)
	}
	return nil, errors.New("GetProjectCommitsFunc not implemented")
}

func (m *MockGitService) GetRepoContributors(repoIdentifier interface{}) ([]*common_types.User, error) {
	if m.GetRepoContributorsFunc != nil {
		return m.GetRepoContributorsFunc(repoIdentifier)
	}
	return nil, errors.New("GetRepoContributorsFunc not implemented")
}

// MockRedisClient (redefined here for clarity, ideally shared)
type MockRedisClient struct {
	GetFunc func(key string) ([]byte, error)
	SetFunc func(key string, value interface{}, expirationInSeconds time.Duration) error
	DeleteFunc func(key string) error
}

func (m *MockRedisClient) Get(key string) ([]byte, error) {
	if m.GetFunc != nil {
		return m.GetFunc(key)
	}
	return nil, errors.New("GetFunc not implemented")
}

func (m *MockRedisClient) Set(key string, value interface{}, expirationInSeconds time.Duration) error {
	if m.SetFunc != nil {
		return m.SetFunc(key, value, expirationInSeconds)
	}
	return errors.New("SetFunc not implemented")
}

func (m *MockRedisClient) Delete(key string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(key)
	}
	return errors.New("DeleteFunc not implemented")
}


// --- Tests for GitlabApi Handlers ---

func TestGitlabApi_GetAllRepos_Success_NoCache(t *testing.T) {
	mockGitService := &MockGitService{
		GetAllReposFunc: func(owner string) ([]*common_types.Repository, error) {
			// Simulate fetching GitLab repos
			if owner == "mygroup" {
				return []*common_types.Repository{
					{ID: 101, Name: "gl-repo1", Owner: "mygroup"},
				}, nil
			}
			return []*common_types.Repository{
				{ID: 102, Name: "user-repo", Owner: "user"},
			}, nil
		},
	}
	mockRedisClient := &MockRedisClient{
		GetFunc: func(key string) ([]byte, error) { return nil, nil }, // Cache miss
		SetFunc: func(key string, value interface{}, expirationInSeconds time.Duration) error { return nil },
	}
	gitlabAPI := NewGitlabApi(mockGitService, mockRedisClient)

	// Test with owner param
	reqWithOwner, _ := http.NewRequest("GET", "/api/gitlab/repos?owner=mygroup", nil)
	rrWithOwner := httptest.NewRecorder()
	gitlabAPI.GetAllRepos(rrWithOwner, reqWithOwner)

	if status := rrWithOwner.Code; status != http.StatusOK {
		t.Errorf("GetAllRepos (owner) returned wrong status: got %v want %v", status, http.StatusOK)
	}
	var reposWithOwner []*common_types.Repository
	if err := json.Unmarshal(rrWithOwner.Body.Bytes(), &reposWithOwner); err != nil {
		t.Fatalf("GetAllRepos (owner) could not unmarshal: %v", err)
	}
	if len(reposWithOwner) != 1 || reposWithOwner[0].Name != "gl-repo1" {
		t.Errorf("GetAllRepos (owner) unexpected body: got %s", rrWithOwner.Body.String())
	}

	// Test without owner param (for authenticated user's repos)
	reqNoOwner, _ := http.NewRequest("GET", "/api/gitlab/repos", nil)
	rrNoOwner := httptest.NewRecorder()
	gitlabAPI.GetAllRepos(rrNoOwner, reqNoOwner)

	if status := rrNoOwner.Code; status != http.StatusOK {
		t.Errorf("GetAllRepos (no owner) returned wrong status: got %v want %v", status, http.StatusOK)
	}
	var reposNoOwner []*common_types.Repository
	if err := json.Unmarshal(rrNoOwner.Body.Bytes(), &reposNoOwner); err != nil {
		t.Fatalf("GetAllRepos (no owner) could not unmarshal: %v", err)
	}
	if len(reposNoOwner) != 1 || reposNoOwner[0].Name != "user-repo" {
		t.Errorf("GetAllRepos (no owner) unexpected body: got %s", rrNoOwner.Body.String())
	}
}

func TestGitlabApi_GetRepo_Success_NoCache(t *testing.T) {
    mockGitService := &MockGitService{
        GetRepoFunc: func(identifier interface{}) (*common_types.Repository, error) {
            if id, ok := identifier.(int); ok && id == 123 {
                return &common_types.Repository{ID: 123, Name: "gitlab-repo-id", Owner: "group"}, nil
            }
            if idStr, ok := identifier.(string); ok && idStr == "group/repo-path" {
                return &common_types.Repository{ID: 124, Name: "gitlab-repo-path", Owner: "group"}, nil
            }
            return nil, fmt.Errorf("unexpected identifier: %v", identifier)
        },
    }
    mockRedis := &MockRedisClient{
        GetFunc: func(key string) ([]byte, error) { return nil, nil }, // Cache miss
        SetFunc: func(key string, value interface{}, expirationInSeconds time.Duration) error { return nil },
    }
    gitlabApi := NewGitlabApi(mockGitService, mockRedis)

    // Test with int identifier
    reqInt, _ := http.NewRequest("GET", "/api/gitlab/repo?projectID=123", nil)
    rrInt := httptest.NewRecorder()
    gitlabApi.GetRepo(rrInt, reqInt)

    if status := rrInt.Code; status != http.StatusOK {
        t.Errorf("GetRepo with int ID status code: got %v want %v", status, http.StatusOK)
    }
    var repoInt common_types.Repository
    if err := json.Unmarshal(rrInt.Body.Bytes(), &repoInt); err != nil {
        t.Fatalf("GetRepo with int ID could not unmarshal response: %v", err)
    }
    if repoInt.ID != 123 || repoInt.Name != "gitlab-repo-id" {
        t.Errorf("GetRepo with int ID unexpected body: got %+v", repoInt)
    }

    // Test with string identifier
    reqStr, _ := http.NewRequest("GET", "/api/gitlab/repo?projectID=group/repo-path", nil)
    rrStr := httptest.NewRecorder()
    gitlabApi.GetRepo(rrStr, reqStr)

    if status := rrStr.Code; status != http.StatusOK {
        t.Errorf("GetRepo with string ID status code: got %v want %v", status, http.StatusOK)
    }
    var repoStr common_types.Repository
    if err := json.Unmarshal(rrStr.Body.Bytes(), &repoStr); err != nil {
        t.Fatalf("GetRepo with string ID could not unmarshal response: %v", err)
    }
    if repoStr.ID != 124 || repoStr.Name != "gitlab-repo-path" {
        t.Errorf("GetRepo with string ID unexpected body: got %+v", repoStr)
    }
}

func TestGitlabApi_GetAllCommits_Success_NoCache(t *testing.T) {
    mockGitService := &MockGitService{
        GetProjectCommitsFunc: func(repoIdentifier interface{}, options *interfaces.CommitListOptions) ([]*common_types.Commit, error) {
            if id, ok := repoIdentifier.(int); ok && id == 456 {
                 return []*common_types.Commit{{SHA: "glcommit1", Message: "GitLab commit by ID"}}, nil
            }
            if idStr, ok := repoIdentifier.(string); ok && idStr == "group/project" {
                return []*common_types.Commit{{SHA: "glcommit2", Message: "GitLab commit by Path"}}, nil
            }
            return nil, fmt.Errorf("GetAllCommits mock: unexpected identifier %v", repoIdentifier)
        },
    }
    mockRedis := &MockRedisClient{
        GetFunc: func(key string) ([]byte, error) { return nil, nil }, // Cache miss
        SetFunc: func(key string, value interface{}, expirationInSeconds time.Duration) error { return nil },
    }
    gitlabApi := NewGitlabApi(mockGitService, mockRedis)

    // Test with int identifier
    reqInt, _ := http.NewRequest("GET", "/api/gitlab/commits?projectID=456", nil)
    rrInt := httptest.NewRecorder()
    gitlabApi.GetAllCommits(rrInt, reqInt)

    if status := rrInt.Code; status != http.StatusOK {
        t.Errorf("GetAllCommits with int ID status: got %v want %v", status, http.StatusOK)
    }
    var commitsInt []*common_types.Commit
    if err := json.Unmarshal(rrInt.Body.Bytes(), &commitsInt); err != nil {
        t.Fatalf("GetAllCommits with int ID could not unmarshal: %v", err)
    }
    if len(commitsInt) != 1 || commitsInt[0].SHA != "glcommit1" {
        t.Errorf("GetAllCommits with int ID unexpected body: got %+v", commitsInt)
    }
}

// TODO: Add more tests for GitLab API handlers:
// - Cache hit scenarios for each endpoint.
// - Error handling from GitService for each endpoint.
// - Invalid/missing parameters for GetRepo and GetAllCommits.
// - If GetRepoTotalLinesOfCode and GetContributors are implemented for GitLab, test them.

func TestGitlabApi_GetAllRepos_Success_WithCache(t *testing.T) {
	cachedRepoData := []*common_types.Repository{
		{ID: 201, Name: "cached-gl-repo", Owner: "gl-owner"},
	}
	cachedBytes, _ := json.Marshal(cachedRepoData)
	redisKey := "gitlab_get_all_repos_" // Assuming empty owner for this cache test

	mockRedisClient := &MockRedisClient{
		GetFunc: func(key string) ([]byte, error) {
			if key == redisKey {
				return cachedBytes, nil
			}
			return nil, fmt.Errorf("unexpected key in GetAllRepos cache test: got %s, want %s", key, redisKey)
		},
	}
	mockGitService := &MockGitService{
		GetAllReposFunc: func(owner string) ([]*common_types.Repository, error) {
			t.Error("GitService.GetAllRepos called unexpectedly in cache hit scenario")
			return nil, errors.New("GitService.GetAllRepos should not be called")
		},
	}
	gitlabAPI := NewGitlabApi(mockGitService, mockRedisClient)

	req, _ := http.NewRequest("GET", "/api/gitlab/repos", nil) // No owner query param
	rr := httptest.NewRecorder()
	gitlabAPI.GetAllRepos(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetAllRepos (cache hit) status: got %v want %v", status, http.StatusOK)
	}
	var repos []*common_types.Repository
	if err := json.Unmarshal(rr.Body.Bytes(), &repos); err != nil {
		t.Fatalf("GetAllRepos (cache hit) could not unmarshal: %v", err)
	}
	if len(repos) != 1 || repos[0].Name != "cached-gl-repo" {
		t.Errorf("GetAllRepos (cache hit) unexpected body: got %s", rr.Body.String())
	}
}

func TestGitlabApi_GetRepo_Success_WithCache(t *testing.T) {
    cachedRepo := &common_types.Repository{ID: 789, Name: "cached-gitlab-repo", Owner: "cache-owner-gl"}
    cachedBytes, _ := json.Marshal(cachedRepo)
    redisKey := "gitlab_get_repo_789"

    mockRedisClient := &MockRedisClient{
        GetFunc: func(key string) ([]byte, error) {
            if key == redisKey {
                return cachedBytes, nil
            }
            return nil, fmt.Errorf("unexpected key in GetRepo cache test: got %s, want %s", key, redisKey)
        },
    }
    mockGitService := &MockGitService{ // Should not be called
        GetRepoFunc: func(identifier interface{}) (*common_types.Repository, error) {
            t.Error("GitService.GetRepo was called unexpectedly in GetRepo cache hit scenario")
            return nil, errors.New("should not be called")
        },
    }
    gitlabAPI := NewGitlabApi(mockGitService, mockRedisClient)

    req, _ := http.NewRequest("GET", "/api/gitlab/repo?projectID=789", nil)
    rr := httptest.NewRecorder()
    gitlabAPI.GetRepo(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("GetRepo (cache hit) returned wrong status: got %v want %v", status, http.StatusOK)
    }
    var repo common_types.Repository
    if err := json.Unmarshal(rr.Body.Bytes(), &repo); err != nil {
        t.Fatalf("GetRepo (cache hit) could not unmarshal: %v", err)
    }
    if repo.ID != 789 || repo.Name != "cached-gitlab-repo" {
        t.Errorf("GetRepo (cache hit) unexpected body: got %+v", repo)
    }
}
