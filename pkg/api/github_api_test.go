package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ahmetk3436/git-stats-golang/pkg/common_types"
	"github.com/ahmetk3436/git-stats-golang/pkg/interfaces"
	// Import storage if InMemoryDB is directly used, or its interface for mocking
	// "github.com/ahmetk3436/git-stats-golang/internal" // For concrete RedisClient, if not mocking InMemoryDB interface
)

// --- Mocks ---

// MockGitService is a mock implementation of interfaces.GitService
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

// MockRedisClient is a mock implementation of storage.InMemoryDB (or the relevant interface for Redis operations)
// Assuming storage.InMemoryDB has Get and Set methods.
type MockRedisClient struct {
	GetFunc func(key string) ([]byte, error)
	SetFunc func(key string, value interface{}, expirationInSeconds time.Duration) error
	// Add DeleteFunc if needed
}

func (m *MockRedisClient) Get(key string) ([]byte, error) {
	if m.GetFunc != nil {
		return m.GetFunc(key)
	}
	return nil, errors.New("GetFunc not implemented in MockRedisClient")
}

func (m *MockRedisClient) Set(key string, value interface{}, expirationInSeconds time.Duration) error {
	if m.SetFunc != nil {
		// The value in the real Redis client is often marshalled before Set.
		// Here, we might just pass it through or simulate marshalling if necessary for the test.
		return m.SetFunc(key, value, expirationInSeconds)
	}
	return errors.New("SetFunc not implemented in MockRedisClient")
}
func (m *MockRedisClient) Delete(key string) error {
    //TODO implement me
    panic("implement me")
}


// --- Tests for GithubApi Handlers ---

func TestGithubApi_GetAllRepos_Success_NoCache(t *testing.T) {
	mockGitService := &MockGitService{
		GetAllReposFunc: func(owner string) ([]*common_types.Repository, error) {
			return []*common_types.Repository{
				{ID: 1, Name: "repo1", Owner: "owner1"},
			}, nil
		},
	}
	mockRedisClient := &MockRedisClient{
		GetFunc: func(key string) ([]byte, error) {
			return nil, nil // Cache miss
		},
		SetFunc: func(key string, value interface{}, expirationInSeconds time.Duration) error {
			return nil // Assume Set is successful
		},
	}

	githubAPI := NewGithubApi(mockGitService, mockRedisClient)

	req, err := http.NewRequest("GET", "/api/github/repos", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(githubAPI.GetAllRepos)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedBody := `[{"ID":1,"Name":"repo1","Owner":"owner1","HTMLURL":"","CloneURL":"","Description":"","CreatedAt":"0001-01-01T00:00:00Z","UpdatedAt":"0001-01-01T00:00:00Z","Stars":0,"Forks":0,"OpenIssues":0}]`
	// Note: Time fields will be zero value. Marshal common_types.Repository to be sure of exact expected output.
	var respRepos []*common_types.Repository
	if err := json.Unmarshal(rr.Body.Bytes(), &respRepos); err != nil {
		t.Fatalf("Could not unmarshal response body: %v", err)
	}

	if len(respRepos) != 1 || respRepos[0].Name != "repo1" {
		t.Errorf("handler returned unexpected body: got %s want containing repo1", rr.Body.String())
	}
	// A more thorough check would compare all fields or use a library for deep equality.
}

func TestGithubApi_GetAllRepos_Success_WithCache(t *testing.T) {
	cachedRepoData := []*common_types.Repository{
		{ID: 2, Name: "cached-repo", Owner: "owner2"},
	}
	cachedBytes, _ := json.Marshal(cachedRepoData)

	mockRedisClient := &MockRedisClient{
		GetFunc: func(key string) ([]byte, error) {
			return cachedBytes, nil // Cache hit
		},
		// SetFunc is not expected to be called in cache hit scenario for GetAllRepos
	}
	// GitService mock is not strictly needed here if cache is hit, but API setup requires it.
	mockGitService := &MockGitService{
		GetAllReposFunc: func(owner string) ([]*common_types.Repository, error) {
			// This should not be called if cache is hit
			t.Error("GitService.GetAllRepos was called unexpectedly in cache hit scenario")
			return nil, errors.New("should not be called")
		},
	}


	githubAPI := NewGithubApi(mockGitService, mockRedisClient)

	req, err := http.NewRequest("GET", "/api/github/repos", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(githubAPI.GetAllRepos)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var respRepos []*common_types.Repository
	if err := json.Unmarshal(rr.Body.Bytes(), &respRepos); err != nil {
		t.Fatalf("Could not unmarshal response body: %v", err)
	}
	if len(respRepos) != 1 || respRepos[0].Name != "cached-repo" {
		t.Errorf("handler returned unexpected body from cache: got %s", rr.Body.String())
	}
}

func TestGithubApi_GetAllRepos_ErrorFromService(t *testing.T) {
	mockGitService := &MockGitService{
		GetAllReposFunc: func(owner string) ([]*common_types.Repository, error) {
			return nil, errors.New("git service error")
		},
	}
	mockRedisClient := &MockRedisClient{
		GetFunc: func(key string) ([]byte, error) {
			return nil, nil // Cache miss
		},
		// SetFunc might be called or not depending on error handling point.
		// If error occurs before Set, it won't be called.
	}

	githubAPI := NewGithubApi(mockGitService, mockRedisClient)

	req, err := http.NewRequest("GET", "/api/github/repos", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(githubAPI.GetAllRepos)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	// Optionally, check error message in body if your API writes one.
	// e.g., if http.Error(w, fetchErr.Error(), http.StatusInternalServerError) is used.
	// expectedErrorMsg := "git service error"
	// if !strings.Contains(rr.Body.String(), expectedErrorMsg) {
	//  t.Errorf("handler body does not contain expected error: got %s", rr.Body.String())
	// }
}

// TODO: Add tests for GetRepo, GetAllCommits, GetContributors, GetRepoTotalLinesOfCode
// - Test cases for cache hits and misses.
// - Test cases for errors returned by the GitService.
// - Test cases for invalid query parameters (e.g., missing projectID for GetRepo).
// - For GetRepoTotalLinesOfCode, mocking os.MkdirTemp, os.RemoveAll, exec.Command might be too complex
//   without a proper filesystem mocking library or refactoring to make it more testable.
//   Focus on testing its behavior with Redis cache and error handling from those parts.
//   The actual git cloning and command execution part is harder to unit test purely.

func TestGithubApi_GetRepo_Success_NoCache(t *testing.T) {
	mockGitService := &MockGitService{
		GetRepoFunc: func(identifier interface{}) (*common_types.Repository, error) {
			// Check if identifier is what's expected, e.g., int64(123) or "owner/repo"
			if id, ok := identifier.(int64); ok && id == 123 {
				return &common_types.Repository{ID: 123, Name: "test-repo", Owner: "test-owner"}, nil
			}
			if idStr, ok := identifier.(string); ok && idStr == "test-owner/test-repo" {
				return &common_types.Repository{ID: 124, Name: "test-repo-str", Owner: "test-owner"}, nil
			}
			return nil, errors.New("unexpected identifier in mock GetRepoFunc")
		},
	}
	mockRedisClient := &MockRedisClient{
		GetFunc: func(key string) ([]byte, error) { return nil, nil }, // Cache miss
		SetFunc: func(key string, value interface{}, expirationInSeconds time.Duration) error { return nil },
	}
	githubAPI := NewGithubApi(mockGitService, mockRedisClient)

	// Test with int64 identifier
	reqInt, _ := http.NewRequest("GET", "/api/github/repo?projectID=123", nil)
	rrInt := httptest.NewRecorder()
	githubAPI.GetRepo(rrInt, reqInt)

	if status := rrInt.Code; status != http.StatusOK {
		t.Errorf("GetRepo (int ID) returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	var repoInt common_types.Repository
	if err := json.Unmarshal(rrInt.Body.Bytes(), &repoInt); err != nil {
		t.Fatalf("GetRepo (int ID) could not unmarshal response: %v", err)
	}
	if repoInt.ID != 123 || repoInt.Name != "test-repo" {
		t.Errorf("GetRepo (int ID) returned unexpected body: got %+v", repoInt)
	}

	// Test with string identifier
	reqStr, _ := http.NewRequest("GET", "/api/github/repo?projectID=test-owner/test-repo", nil)
	rrStr := httptest.NewRecorder()
	githubAPI.GetRepo(rrStr, reqStr)

	if status := rrStr.Code; status != http.StatusOK {
		t.Errorf("GetRepo (str ID) returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	var repoStr common_types.Repository
	if err := json.Unmarshal(rrStr.Body.Bytes(), &repoStr); err != nil {
		t.Fatalf("GetRepo (str ID) could not unmarshal response: %v", err)
	}
	if repoStr.ID != 124 || repoStr.Name != "test-repo-str" {
		t.Errorf("GetRepo (str ID) returned unexpected body: got %+v", repoStr)
	}
}

func TestGithubApi_GetRepo_MissingProjectID(t *testing.T) {
	githubAPI := NewGithubApi(&MockGitService{}, &MockRedisClient{}) // Mocks don't need specific behavior for this test

	req, _ := http.NewRequest("GET", "/api/github/repo", nil) // No projectID query param
	rr := httptest.NewRecorder()
	githubAPI.GetRepo(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("GetRepo (missing projectID) returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
	// Optionally check body for error message "projectID query parameter is required."
}

func TestGithubApi_GetAllCommits_Success_NoCache(t *testing.T) {
	mockGitService := &MockGitService{
		GetProjectCommitsFunc: func(repoIdentifier interface{}, options *interfaces.CommitListOptions) ([]*common_types.Commit, error) {
			if repoIdentifier == "test-owner/test-repo" {
				return []*common_types.Commit{
					{SHA: "sha1", Message: "commit1"},
				}, nil
			}
			return nil, errors.New("unexpected repoIdentifier in mock GetProjectCommitsFunc")
		},
	}
	mockRedisClient := &MockRedisClient{
		GetFunc: func(key string) ([]byte, error) { return nil, nil }, // Cache miss
		SetFunc: func(key string, value interface{}, expirationInSeconds time.Duration) error { return nil },
	}
	githubAPI := NewGithubApi(mockGitService, mockRedisClient)

	req, _ := http.NewRequest("GET", "/api/github/commits?projectOwner=test-owner&repoName=test-repo", nil)
	rr := httptest.NewRecorder()
	githubAPI.GetAllCommits(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetAllCommits returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	var commits []*common_types.Commit
	if err := json.Unmarshal(rr.Body.Bytes(), &commits); err != nil {
		t.Fatalf("GetAllCommits could not unmarshal response: %v", err)
	}
	if len(commits) != 1 || commits[0].SHA != "sha1" {
		t.Errorf("GetAllCommits returned unexpected body: got %+v", commits)
	}
}

func TestGithubApi_GetAllCommits_MissingParams(t *testing.T) {
	githubAPI := NewGithubApi(&MockGitService{}, &MockRedisClient{})

	tests := []struct {
		name        string
		queryString string
	}{
		{"missing projectOwner", "?repoName=test-repo"},
		{"missing repoName", "?projectOwner=test-owner"},
		{"both missing", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/github/commits"+tt.queryString, nil)
			rr := httptest.NewRecorder()
			githubAPI.GetAllCommits(rr, req)

			if status := rr.Code; status != http.StatusBadRequest {
				t.Errorf("GetAllCommits with %s returned wrong status code: got %v want %v", tt.name, status, http.StatusBadRequest)
			}
		})
	}
}

func TestGithubApi_GetRepoTotalLinesOfCode_Success_WithCache(t *testing.T) {
	expectedLOC := map[string]interface{}{"totalLines": 12345}
	cachedBytes, _ := json.Marshal(expectedLOC)

	mockRedisClient := &MockRedisClient{
		GetFunc: func(key string) ([]byte, error) {
			// Check if the key matches what GetRepoTotalLinesOfCode would use
			if key == "github_get_loc_https://example.com/repo.git" {
				return cachedBytes, nil // Cache hit
			}
			return nil, errors.New("unexpected key for redis Get in GetRepoTotalLinesOfCode test")
		},
	}
	// GitService is not directly used by GetRepoTotalLinesOfCode for the primary logic, only for Redis.
	// The actual LOC calculation involves os/exec, not GitService methods.
	githubAPI := NewGithubApi(&MockGitService{}, mockRedisClient)

	req, _ := http.NewRequest("GET", "/api/github/loc?repoUrl=https://example.com/repo.git", nil)
	rr := httptest.NewRecorder()
	githubAPI.GetRepoTotalLinesOfCode(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetRepoTotalLinesOfCode (cache hit) returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var respLOC map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &respLOC); err != nil {
		t.Fatalf("GetRepoTotalLinesOfCode (cache hit) could not unmarshal response: %v", err)
	}
	if respLOC["totalLines"] == nil { // Could be float64 or int after unmarshal
		t.Errorf("GetRepoTotalLinesOfCode (cache hit) 'totalLines' is nil")
		return
	}
	if int(respLOC["totalLines"].(float64)) != 12345 {
		t.Errorf("GetRepoTotalLinesOfCode (cache hit) returned unexpected body: got %v want %v", respLOC["totalLines"], expectedLOC["totalLines"])
	}
}


func TestGithubApi_GetRepoTotalLinesOfCode_MissingRepoUrl(t *testing.T) {
	githubAPI := NewGithubApi(&MockGitService{}, &MockRedisClient{})

	req, _ := http.NewRequest("GET", "/api/github/loc", nil) // No repoUrl query param
	rr := httptest.NewRecorder()
	githubAPI.GetRepoTotalLinesOfCode(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("GetRepoTotalLinesOfCode (missing repoUrl) returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
	// Expected body: "repoUrl query parameter is required."
	// Check if the error message is present (optional, but good for completeness)
	// if !strings.Contains(rr.Body.String(), "repoUrl query parameter is required") {
	// 	t.Errorf("Expected error message not found in response body: got '%s'", rr.Body.String())
	// }
}

// Note: Testing the actual LOC calculation (no cache) for GetRepoTotalLinesOfCode is an integration test
// as it involves cloning a repo and running shell commands. It's excluded from these unit tests.
// Unit tests for helper functions like `extractTotalLines` can be done separately if they contain complex logic.
// `cloneRepository` and `runCommand` are thin wrappers around `os/exec` and are hard to unit test without
// a proper way to mock `exec.Command`.

func TestGithubApi_GetRepo_Success_WithCache(t *testing.T) {
	cachedRepo := &common_types.Repository{ID: 789, Name: "cached-repo", Owner: "cache-owner"}
	cachedBytes, _ := json.Marshal(cachedRepo)

	mockRedisClient := &MockRedisClient{
		GetFunc: func(key string) ([]byte, error) {
			if key == "github_get_repo_789" { // Assuming key format
				return cachedBytes, nil
			}
			return nil, errors.New("unexpected key in GetRepo cache test")
		},
	}
	mockGitService := &MockGitService{ // Should not be called
		GetRepoFunc: func(identifier interface{}) (*common_types.Repository, error) {
			t.Error("GitService.GetRepo was called unexpectedly in GetRepo cache hit scenario")
			return nil, errors.New("should not be called")
		},
	}
	githubAPI := NewGithubApi(mockGitService, mockRedisClient)

	req, _ := http.NewRequest("GET", "/api/github/repo?projectID=789", nil)
	rr := httptest.NewRecorder()
	githubAPI.GetRepo(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetRepo (cache hit) returned wrong status: got %v want %v", status, http.StatusOK)
	}
	var repo common_types.Repository
	if err := json.Unmarshal(rr.Body.Bytes(), &repo); err != nil {
		t.Fatalf("GetRepo (cache hit) could not unmarshal: %v", err)
	}
	if repo.ID != 789 || repo.Name != "cached-repo" {
		t.Errorf("GetRepo (cache hit) unexpected body: got %+v", repo)
	}
}

func TestGithubApi_GetAllCommits_Success_WithCache(t *testing.T) {
	cachedCommits := []*common_types.Commit{
		{SHA: "cachedsha1", Message: "cached commit1"},
	}
	cachedBytes, _ := json.Marshal(cachedCommits)
	redisKey := "github_get_commits_test-owner_test-repo"

	mockRedisClient := &MockRedisClient{
		GetFunc: func(key string) ([]byte, error) {
			if key == redisKey {
				return cachedBytes, nil
			}
			return nil, errors.New("unexpected key in GetAllCommits cache test")
		},
	}
	mockGitService := &MockGitService{ // Should not be called
		GetProjectCommitsFunc: func(repoIdentifier interface{}, options *interfaces.CommitListOptions) ([]*common_types.Commit, error) {
			t.Error("GitService.GetProjectCommits was called unexpectedly in GetAllCommits cache hit scenario")
			return nil, errors.New("should not be called")
		},
	}
	githubAPI := NewGithubApi(mockGitService, mockRedisClient)

	req, _ := http.NewRequest("GET", "/api/github/commits?projectOwner=test-owner&repoName=test-repo", nil)
	rr := httptest.NewRecorder()
	githubAPI.GetAllCommits(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetAllCommits (cache hit) returned wrong status: got %v want %v", status, http.StatusOK)
	}
	var commits []*common_types.Commit
	if err := json.Unmarshal(rr.Body.Bytes(), &commits); err != nil {
		t.Fatalf("GetAllCommits (cache hit) could not unmarshal: %v", err)
	}
	if len(commits) != 1 || commits[0].SHA != "cachedsha1" {
		t.Errorf("GetAllCommits (cache hit) unexpected body: got %+v", commits)
	}
}

func TestGithubApi_GetContributors_Success_NoCache(t *testing.T) {
	mockGitService := &MockGitService{
		GetRepoContributorsFunc: func(repoIdentifier interface{}) ([]*common_types.User, error) {
			if repoIdentifier == "test-owner/test-repo" {
				return []*common_types.User{
					{Login: "user1", ID: 1},
					{Login: "user2", ID: 2},
				}, nil
			}
			return nil, errors.New("unexpected repoIdentifier in mock GetRepoContributorsFunc")
		},
	}
	// Note: GetContributors in the current implementation does not use Redis, so MockRedisClient is not strictly needed
	// but the NewGithubApi constructor requires it.
	mockRedisClient := &MockRedisClient{
		GetFunc: func(key string) ([]byte, error) { return nil, errors.New("redis Get should not be called by GetContributors") },
		SetFunc: func(key string, value interface{}, expirationInSeconds time.Duration) error { return errors.New("redis Set should not be called by GetContributors") },
	}
	githubAPI := NewGithubApi(mockGitService, mockRedisClient)

	req, _ := http.NewRequest("GET", "/api/github/contributors?owner=test-owner&repoName=test-repo", nil)
	rr := httptest.NewRecorder()
	githubAPI.GetContributors(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetContributors returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	var contributors []*common_types.User
	if err := json.Unmarshal(rr.Body.Bytes(), &contributors); err != nil {
		t.Fatalf("GetContributors could not unmarshal response: %v", err)
	}
	if len(contributors) != 2 || contributors[0].Login != "user1" {
		t.Errorf("GetContributors returned unexpected body: got %+v", contributors)
	}
}

func TestGithubApi_GetContributors_MissingParams(t *testing.T) {
	githubAPI := NewGithubApi(&MockGitService{}, &MockRedisClient{})

	tests := []struct {
		name        string
		queryString string
	}{
		{"missing owner", "?repoName=test-repo"},
		{"missing repoName", "?owner=test-owner"},
		{"both missing", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/github/contributors"+tt.queryString, nil)
			rr := httptest.NewRecorder()
			githubAPI.GetContributors(rr, req)

			if status := rr.Code; status != http.StatusBadRequest {
				t.Errorf("GetContributors with %s returned wrong status code: got %v want %v", tt.name, status, http.StatusBadRequest)
			}
		})
	}
}
