package repository

import (
	"testing"
	"time"

	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/ahmetk3436/git-stats-golang/pkg/common_types"
	"github.com/xanzy/go-gitlab"
)

// Helper function to create a pointer to a string (if not already available globally for tests)
// func String(s string) *string { return &s }

// Helper function to create a pointer to an int (if not already available globally for tests)
// func Int(i int) *int { return &i }

// Helper function to create a pointer to time.Time
func Time(t time.Time) *time.Time {
	return &t
}


func TestToCommonRepositoryGL(t *testing.T) {
	testTime := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		glProject *gitlab.Project
		expected  *common_types.Repository
	}{
		{
			name: "nil input",
			glProject: nil,
			expected:  nil,
		},
		{
			name: "basic project",
			glProject: &gitlab.Project{
				ID:          10,
				Name:        "gitlab-test-repo",
				Description: "A GitLab test repository",
				WebURL:      "https://gitlab.com/owner/gitlab-test-repo",
				HTTPURLToRepo: "https://gitlab.com/owner/gitlab-test-repo.git",
				Owner: &gitlab.User{
					Username: "gitlabowner",
				},
				Namespace: &gitlab.Namespace{
					Path: "gitlabowner",
					Name: "GitLab Owner Group",
				},
				StarCount:       20,
				ForksCount:      15,
				OpenIssuesCount: 5,
				CreatedAt:       &testTime,
				LastActivityAt:  &testTime,
			},
			expected: &common_types.Repository{
				ID:          10,
				Name:        "gitlab-test-repo",
				Owner:       "gitlabowner", // Prefers Owner.Username
				HTMLURL:     "https://gitlab.com/owner/gitlab-test-repo",
				CloneURL:    "https://gitlab.com/owner/gitlab-test-repo.git",
				Description: "A GitLab test repository",
				CreatedAt:   testTime,
				UpdatedAt:   testTime,
				Stars:       20,
				Forks:       15,
				OpenIssues:  5,
			},
		},
		{
			name: "project with nil owner, uses namespace",
			glProject: &gitlab.Project{
				ID:   11,
				Name: "repo-ns-owner",
				Owner: nil,
				Namespace: &gitlab.Namespace{Path: "groupname"},
				WebURL:"https://gitlab.com/groupname/repo-ns-owner",
			},
			expected: &common_types.Repository{
				ID: 11,
				Name: "repo-ns-owner",
				Owner: "groupname",
				HTMLURL: "https://gitlab.com/groupname/repo-ns-owner",
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
			},
		},
		{
			name: "project with some nil fields",
			glProject: &gitlab.Project{
				ID:   12,
				Name: "partial-gl-repo",
				Owner: &gitlab.User{Username: "user1"},
				// Description, HTTPURLToRepo, etc., are nil
			},
			expected: &common_types.Repository{
				ID:    12,
				Name:  "partial-gl-repo",
				Owner: "user1",
				HTMLURL: "",
				CloneURL: "",
				Description: "",
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toCommonRepositoryGL(tt.glProject)
			if result == nil && tt.expected == nil {
				return
			}
			if result == nil || tt.expected == nil {
				t.Errorf("toCommonRepositoryGL() returned %v, expected %v", result, tt.expected)
				return
			}

			if result.ID != tt.expected.ID ||
				result.Name != tt.expected.Name ||
				result.Owner != tt.expected.Owner ||
				result.HTMLURL != tt.expected.HTMLURL ||
				result.CloneURL != tt.expected.CloneURL ||
				result.Description != tt.expected.Description ||
				!result.CreatedAt.Equal(tt.expected.CreatedAt) ||
				!result.UpdatedAt.Equal(tt.expected.UpdatedAt) ||
				result.Stars != tt.expected.Stars ||
				result.Forks != tt.expected.Forks ||
				result.OpenIssues != tt.expected.OpenIssues {
				t.Errorf("toCommonRepositoryGL() mismatch:\nGot:    %+v\nWanted: %+v", result, tt.expected)
			}
		})
	}
}


func TestToCommonCommitGL(t *testing.T) {
	testTime := time.Date(2024, 1, 2, 11, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		glCommit *gitlab.Commit
		expected *common_types.Commit
	}{
		{
			name: "nil input",
			glCommit: nil,
			expected: nil,
		},
		{
			name: "basic commit",
			glCommit: &gitlab.Commit{
				ID:           "xyz789sha",
				AuthorName:   "GitLab User",
				AuthorEmail:  "gluser@example.com",
				AuthoredDate: &testTime,
				Message:      "Initial GitLab commit",
				WebURL:       "https://gitlab.com/commit/xyz789sha",
				Stats: &gitlab.CommitStats{
					Additions: 100,
					Deletions: 20,
					Total:     120,
				},
			},
			expected: &common_types.Commit{
				SHA: "xyz789sha",
				Author: common_types.CommitAuthor{
					Name:  "GitLab User",
					Email: "gluser@example.com",
					Date:  testTime,
				},
				Message: "Initial GitLab commit",
				HTMLURL: "https://gitlab.com/commit/xyz789sha",
				Stats:   common_types.CommitStats{Additions: 100, Deletions: 20, Total: 120},
			},
		},
		{
			name: "commit with nil stats",
			glCommit: &gitlab.Commit{
				ID:           "abc123sha",
				AuthorName:   "Another User",
				AuthoredDate: &testTime,
				Message:      "Commit without stats",
				Stats:        nil,
			},
			expected: &common_types.Commit{
				SHA: "abc123sha",
				Author: common_types.CommitAuthor{
					Name: "Another User",
					Date: testTime,
				},
				Message: "Commit without stats",
				Stats:   common_types.CommitStats{Additions: 0, Deletions: 0, Total: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toCommonCommitGL(tt.glCommit)
			if result == nil && tt.expected == nil {
				return
			}
			if result == nil || tt.expected == nil {
				t.Errorf("toCommonCommitGL() result is %v, expected %v", result, tt.expected)
				return
			}

			if result.SHA != tt.expected.SHA ||
				result.Author.Name != tt.expected.Author.Name ||
				result.Author.Email != tt.expected.Author.Email ||
				!result.Author.Date.Equal(tt.expected.Author.Date) ||
				result.Message != tt.expected.Message ||
				result.HTMLURL != tt.expected.HTMLURL ||
				result.Stats != tt.expected.Stats {
				t.Errorf("toCommonCommitGL() mismatch:\nGot:    %+v\nWanted: %+v", result, tt.expected)
			}
		})
	}
}

func TestToCommonUserGL(t *testing.T) {
	tests := []struct {
		name     string
		glUser   *gitlab.User
		expected *common_types.User
	}{
		{
			name: "nil input",
			glUser:   nil,
			expected: nil,
		},
		{
			name: "basic user",
			glUser: &gitlab.User{
				ID:        200,
				Username:  "gitlab_user_login",
				Name:      "GitLab Test User",
				AvatarURL: "https://gitlab.com/avatar/gitlab_user_login",
				WebURL:    "https://gitlab.com/gitlab_user_login",
			},
			expected: &common_types.User{
				Login:     "gitlab_user_login",
				ID:        200,
				AvatarURL: "https://gitlab.com/avatar/gitlab_user_login",
				HTMLURL:   "https://gitlab.com/gitlab_user_login",
				Name:      "GitLab Test User",
			},
		},
		{
			name: "user with some nil/empty fields",
			glUser: &gitlab.User{
				ID:       201,
				Username: "partial_gl_user",
				// Name, AvatarURL, WebURL are empty
			},
			expected: &common_types.User{
				Login:     "partial_gl_user",
				ID:        201,
				AvatarURL: "",
				HTMLURL:   "",
				Name:      "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toCommonUserGL(tt.glUser)
			if result == nil && tt.expected == nil {
				return
			}
			if result == nil || tt.expected == nil {
				t.Errorf("toCommonUserGL() result is %v, expected %v", result, tt.expected)
				return
			}

			if result.Login != tt.expected.Login ||
				result.ID != tt.expected.ID ||
				result.AvatarURL != tt.expected.AvatarURL ||
				result.HTMLURL != tt.expected.HTMLURL ||
				result.Name != tt.expected.Name {
				t.Errorf("toCommonUserGL() mismatch:\nGot:    %+v\nWanted: %+v", result, tt.expected)
			}
		})
	}
}
