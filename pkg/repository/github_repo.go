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
func ConnectGithub(token string) *github.Client {
	client := github.NewClient(nil).WithAuthToken(token)
	return client
}
func (repo GitHubRepo) GetRepoGithub(projectId int64) (*github.Repository, error) {
	ctx := context.Background()
	githubRepo, _, err := repo.Client.Repositories.GetByID(ctx, projectId)
	if err != nil {
		return nil, err
	}
	return githubRepo, nil
}
func (repo GitHubRepo) GetAllReposGithub() ([]*github.Repository, error) {
	ctx := context.Background()
	repoList := github.RepositoryListOptions{Type: "owner", Sort: "updated", Direction: "desc"}
	repos, _, err := repo.Client.Repositories.List(ctx, "", &repoList)
	if err != nil {
		return nil, err
	}

	return repos, nil
}
func (repo GitHubRepo) GetProjectCommitsGithub(repoOwner, repoName string) ([]*github.RepositoryCommit, error) {
	ctx := context.Background()
	options := github.CommitsListOptions{}
	commits, _, err := repo.Client.Repositories.ListCommits(ctx, repoOwner, repoName, &options)
	if err != nil {
		panic(err)
	}
	return commits, nil
}
