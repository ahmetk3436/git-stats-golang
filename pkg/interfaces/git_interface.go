package interfaces

import "github.com/google/go-github/v56/github"

type GitInterface interface {
	Connect(token string) *github.Client
	GetAllRepos() ([]*github.Repository, error)
	GetProjectCommits(repoOwner, repoName string) ([]*github.RepositoryCommit, error)
}
