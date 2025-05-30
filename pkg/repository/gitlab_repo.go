package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/ahmetk3436/git-stats-golang/pkg/common_types"
	"github.com/ahmetk3436/git-stats-golang/pkg/interfaces"
	"github.com/xanzy/go-gitlab"
)

// Gitlab implements the interfaces.GitService for GitLab.
// It uses the go-gitlab client to interact with the GitLab API.
type Gitlab struct {
	Client *gitlab.Client // Client is the GitLab API client.
}

// NewGitlabClient creates a new Gitlab service instance.
// It requires a non-nil gitlab.Client.
func NewGitlabClient(gitlabAPIClient *gitlab.Client) (*Gitlab, error) {
	if gitlabAPIClient == nil {
		return nil, fmt.Errorf("gitlab client is nil, cannot create Gitlab service")
	}
	return &Gitlab{
		Client: gitlabAPIClient,
	}, nil
}

// ConnectGitlab creates a new GitLab API client.
// token is the GitLab personal access token.
// host can be a pointer to a string for GitLab self-managed instances, or nil/empty for GitLab.com.
// This is a helper function for initializing the Gitlab service and is not part of the GitService interface.
func ConnectGitlab(token string, hostURL *string) (*gitlab.Client, error) {
	var gitlabClient *gitlab.Client
	var err error

	if hostURL != nil && *hostURL != "" {
		// Use custom host URL for self-managed GitLab.
		gitlabClient, err = gitlab.NewClient(token, gitlab.WithBaseURL(*hostURL))
	} else {
		// Use default GitLab.com URL.
		gitlabClient, err = gitlab.NewClient(token)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create gitlab client: %w", err)
	}
	return gitlabClient, nil
}

// toCommonRepositoryGL converts a GitLab specific project object to the common_types.Repository.
func toCommonRepositoryGL(glProject *gitlab.Project) *common_types.Repository {
	if glProject == nil {
		return nil
	}

	var ownerLogin string
	if glProject.Owner != nil {
		ownerLogin = glProject.Owner.Username // GitLab Project.Owner provides Username.
	} else if glProject.Namespace != nil {
		// Fallback to namespace path or name if owner is not directly available.
		// This is common for group-owned projects.
		ownerLogin = glProject.Namespace.Path
		if ownerLogin == "" {
			ownerLogin = glProject.Namespace.Name
		}
	}

	var createdAt time.Time
	if glProject.CreatedAt != nil {
		createdAt = *glProject.CreatedAt
	}
	// GitLab uses LastActivityAt as a primary indicator of updates.
	var updatedAt time.Time
	if glProject.LastActivityAt != nil {
		updatedAt = *glProject.LastActivityAt
	}

	return &common_types.Repository{
		ID:          int64(glProject.ID), // Ensure ID conversion is safe if GitLab IDs can exceed int64 range (unlikely).
		Name:        glProject.Name,
		Owner:       ownerLogin,
		HTMLURL:     glProject.WebURL,
		CloneURL:    glProject.HTTPURLToRepo, // Or SSHURLToRepo depending on preference.
		Description: glProject.Description,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Stars:       glProject.StarCount,
		Forks:       glProject.ForksCount,
		OpenIssues:  glProject.OpenIssuesCount,
	}
}

// toCommonCommitGL converts a GitLab specific commit object to the common_types.Commit.
func toCommonCommitGL(glCommit *gitlab.Commit) *common_types.Commit {
	if glCommit == nil {
		return nil
	}

	var authoredAt time.Time
	if glCommit.AuthoredDate != nil {
		authoredAt = *glCommit.AuthoredDate
	}

	var commitStats common_types.CommitStats
	if glCommit.Stats != nil {
		commitStats.Additions = glCommit.Stats.Additions
		commitStats.Deletions = glCommit.Stats.Deletions
		commitStats.Total = glCommit.Stats.Total
	}
	// Note: GitLab's commit stats might need `?stats=true` on the API call,
	// which is usually handled by `ListCommitsOptions{WithStats: gitlab.Bool(true)}`.

	return &common_types.Commit{
		SHA: glCommit.ID, // GitLab Commit ID is its SHA.
		Author: common_types.CommitAuthor{
			Name:  glCommit.AuthorName,
			Email: glCommit.AuthorEmail,
			Date:  authoredAt,
		},
		Message: glCommit.Message,
		HTMLURL: glCommit.WebURL, // GitLab Commit has a WebURL.
		Stats:   commitStats,
	}
}

// toCommonUserGL converts a GitLab specific user object to the common_types.User.
// Note: This is for full GitLab User objects. Contributor lists might provide different structures.
func toCommonUserGL(glUser *gitlab.User) *common_types.User {
	if glUser == nil {
		return nil
	}
	return &common_types.User{
		Login:     glUser.Username,
		ID:        int64(glUser.ID),
		AvatarURL: glUser.AvatarURL,
		HTMLURL:   glUser.WebURL, // GitLab User has WebURL.
		Name:      glUser.Name,
	}
}

// GetAllRepos implements interfaces.GitService.
// For GitLab, 'ownerLogin' can be a username or a group's path/name.
// If ownerLogin is empty, it lists projects accessible by the authenticated user (considering membership).
func (g *Gitlab) GetAllRepos(ownerLogin string) ([]*common_types.Repository, error) {
	// Options for listing projects. Includes pagination and sorting.
	listProjectsOptions := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{PerPage: 100}, // Fetch 100 items per page.
		OrderBy:     gitlab.String("last_activity_at"), // Sort by last activity.
		Sort:        gitlab.String("desc"),             // Descending order.
		Membership:  gitlab.Bool(true),                 // Include projects where the user is a member.
	}

	var gitlabProjects []*gitlab.Project
	var err error

	if ownerLogin != "" {
		// Listing projects by a specific owner (user or group) in GitLab can be complex.
		// The API has separate endpoints for user projects (`/users/:user_id/projects`)
		// and group projects (`/groups/:group_id/projects`).
		// A simpler approach for this generic interface, if less efficient, is to list all accessible
		// projects and then filter them by the namespace if an ownerLogin is provided.
		// This assumes ownerLogin matches a namespace path.
		// TODO: For performance, especially with many projects, directly using user/group project endpoints
		// would be better but requires knowing if 'ownerLogin' is a user or group and potentially their ID.
		allProjects, _, listErr := g.Client.Projects.ListProjects(listProjectsOptions)
		if listErr != nil {
			return nil, fmt.Errorf("failed to list all gitlab projects for filtering (owner: '%s'): %w", ownerLogin, listErr)
		}
		for _, project := range allProjects {
			if project.Namespace != nil && (project.Namespace.Path == ownerLogin || project.Namespace.FullPath == ownerLogin) {
				gitlabProjects = append(gitlabProjects, project)
			}
		}
	} else {
		// List projects for the authenticated user (based on token permissions and membership).
		gitlabProjects, _, err = g.Client.Projects.ListProjects(listProjectsOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to list gitlab projects for authenticated user: %w", err)
		}
	}

	commonRepos := make([]*common_types.Repository, 0, len(gitlabProjects))
	for _, glProject := range gitlabProjects {
		commonRepos = append(commonRepos, toCommonRepositoryGL(glProject))
	}
	return commonRepos, nil
}

// GetRepo implements interfaces.GitService.
// identifier can be an int (GitLab Project ID) or a string "namespace/project_path".
func (g *Gitlab) GetRepo(identifier interface{}) (*common_types.Repository, error) {
	// Options for getting a single project. Can include statistics.
	getProjectOptions := &gitlab.GetProjectOptions{
		Statistics: gitlab.Bool(true), // Request statistics to be included.
	}

	// The GetProject call handles both int (ID) and string (path) identifiers.
	gitlabProject, _, err := g.Client.Projects.GetProject(identifier, getProjectOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get gitlab repository (identifier: '%v'): %w", identifier, err)
	}
	return toCommonRepositoryGL(gitlabProject), nil
}

// GetProjectCommits implements interfaces.GitService.
// repoIdentifier can be an int (GitLab Project ID) or a string "namespace/project_path".
// options allows for filtering by RefName (branch/tag/SHA), Path, and pagination.
func (g *Gitlab) GetProjectCommits(repoIdentifier interface{}, options *interfaces.CommitListOptions) ([]*common_types.Commit, error) {
	listCommitsOptions := &gitlab.ListCommitsOptions{
		ListOptions: gitlab.ListOptions{PerPage: 100}, // Default PerPage.
		WithStats:   gitlab.Bool(true),                 // Request commit stats.
	}

	if options != nil {
		if options.SHA != "" { // SHA in CommitListOptions maps to RefName in GitLab.
			listCommitsOptions.RefName = gitlab.String(options.SHA)
		}
		if options.Path != "" {
			listCommitsOptions.Path = gitlab.String(options.Path)
		}
		// Note: GitLab's ListCommitsOptions does not directly support filtering by author string (name/email).
		// This would require client-side filtering or a different approach if strictly needed.
		if options.Page > 0 {
			listCommitsOptions.ListOptions.Page = options.Page
		}
		if options.PerPage > 0 {
			listCommitsOptions.ListOptions.PerPage = options.PerPage
		}
	}

	// Ensure repoIdentifier is suitable for ListCommits (int or string).
	var projectIDForCommits interface{}
	switch id := repoIdentifier.(type) {
	case int:
		projectIDForCommits = id
	case int64: // go-gitlab typically expects int for project ID.
		projectIDForCommits = int(id)
	case string:
		projectIDForCommits = id
	default:
		return nil, fmt.Errorf("unsupported repoIdentifier type for GetProjectCommits: %T (expected int or string)", repoIdentifier)
	}

	gitlabCommits, _, err := g.Client.Commits.ListCommits(projectIDForCommits, listCommitsOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list gitlab commits for repo '%v': %w", repoIdentifier, err)
	}

	commonCommits := make([]*common_types.Commit, 0, len(gitlabCommits))
	for _, glCommit := range gitlabCommits {
		commonCommits = append(commonCommits, toCommonCommitGL(glCommit))
	}
	return commonCommits, nil
}

// GetRepoContributors implements interfaces.GitService.
// repoIdentifier can be an int (GitLab Project ID) or a string "namespace/project_path".
// GitLab's contributor concept differs from GitHub's. It returns a list of users
// with commit counts, not full user profiles directly.
func (g *Gitlab) GetRepoContributors(repoIdentifier interface{}) ([]*common_types.User, error) {
	// Ensure repoIdentifier is suitable for Contributors call (int or string).
	var projectIDForContributors interface{}
	switch id := repoIdentifier.(type) {
	case int:
		projectIDForContributors = id
	case int64:
		projectIDForContributors = int(id)
	case string:
		projectIDForContributors = id
	default:
		return nil, fmt.Errorf("unsupported repoIdentifier type for GetRepoContributors: %T", repoIdentifier)
	}

	// Options for listing contributors, e.g., sort by commits.
	listContributorsOptions := &gitlab.ListContributorsOptions{
		Sort: gitlab.String("commits"), // Sort by number of commits.
	}
	gitlabContributors, _, err := g.Client.Repositories.Contributors(projectIDForContributors, listContributorsOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list gitlab contributors for repo '%v': %w", repoIdentifier, err)
	}

	commonUsers := make([]*common_types.User, 0, len(gitlabContributors))
	for _, glContributor := range gitlabContributors {
		// GitLab's Contributor type has Name, Email, Commits, Additions, Deletions.
		// It does not directly provide User ID, Username (Login), AvatarURL, or HTMLURL.
		// Mapping to common_types.User requires either accepting these limitations
		// or making additional API calls to find a matching user by email or name (which can be unreliable).
		// For this implementation, we create a simplified common_types.User.
		// TODO: Consider enhancing this by attempting to look up users by email if more details are needed.
		user := &common_types.User{
			Name: glContributor.Name, // Directly available.
			// Login: // Not available from Contributor struct. Could use email or part of email if unique.
			// ID: // Not available.
			// AvatarURL: // Not available.
			// HTMLURL: // Not available.
		}
		// Example: If you wanted to use email as a Login placeholder (not ideal but possible):
		// user.Login = glContributor.Email

		commonUsers = append(commonUsers, user)
	}
	return commonUsers, nil
}

// Ensure Gitlab implements GitService.
// This line provides a compile-time check that the Gitlab struct
// correctly implements all methods of the interfaces.GitService interface.
var _ interfaces.GitService = (*Gitlab)(nil)
