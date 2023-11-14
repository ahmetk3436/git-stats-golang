package repository

import (
	"context"
	"github.com/google/go-github/v56/github"
)

type GitHubRepo struct {
	Client *github.Client
}

func NewGithubRepo(gitClient *github.Client) (*GitHubRepo, error) {
	return &GitHubRepo{
		Client: gitClient,
	}, nil
}
func Connect(token string) *github.Client {
	client := github.NewClient(nil).WithAuthToken(token)
	return client
}

func (repo GitHubRepo) GetAllRepos() ([]*github.Repository, error) {
	ctx := context.Background()
	repoList := github.RepositoryListOptions{Type: "owner", Sort: "updated", Direction: "desc"}
	repos, _, err := repo.Client.Repositories.List(ctx, "", &repoList)
	if err != nil {
		return nil, err
	}

	return repos, nil
}
func (repo GitHubRepo) GetProjectCommits(repoOwner, repoName string) ([]*github.RepositoryCommit, error) {
	ctx := context.Background()
	options := github.CommitsListOptions{}
	commits, _, err := repo.Client.Repositories.ListCommits(ctx, repoOwner, repoName, &options)
	if err != nil {
		panic(err)
	}
	return commits, nil
}
