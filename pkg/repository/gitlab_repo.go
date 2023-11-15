package repository

import "github.com/xanzy/go-gitlab"

type Gitlab struct {
	Client *gitlab.Client
}

func NewGitlabClient(client *gitlab.Client) *Gitlab {
	return &Gitlab{
		Client: client,
	}
}

func ConnectGitlab(token string, host *string) *gitlab.Client {
	var client *gitlab.Client
	if host != nil {
		baseUrl := gitlab.WithBaseURL(*host)
		newClient, err := gitlab.NewClient(token, baseUrl)
		if err != nil {
			panic(err)
		}
		client = newClient
	} else {
		newClient, err := gitlab.NewClient(token)
		client = newClient
		if err != nil {
			panic(err)
		}
	}
	return client
}
func (Gitlab Gitlab) GetAllReposGitlab() ([]*gitlab.Project, error) {
	options := gitlab.ListProjectsOptions{}
	projects, _, err := Gitlab.Client.Projects.ListProjects(&options)
	if err != nil {
		return nil, err
	}
	return projects, nil
}
func (Gitlab Gitlab) GetRepoGitlab(pid interface{}) (*gitlab.Project, error) {
	statistics := true
	options := gitlab.GetProjectOptions{
		Statistics: &statistics,
	}
	projects, _, err := Gitlab.Client.Projects.GetProject(pid, &options)
	if err != nil {
		return nil, err
	}
	return projects, nil
}
func (Gitlab Gitlab) GetAllCommitsGitlab(pid interface{}) ([]*gitlab.Commit, error) {
	withStats := true
	options := gitlab.ListCommitsOptions{
		WithStats: &withStats,
	}
	commits, _, err := Gitlab.Client.Commits.ListCommits(pid, &options)
	if err != nil {
		return nil, err
	}
	return commits, nil
}
