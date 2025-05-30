package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ahmetk3436/git-stats-golang/pkg/common_types"
	"github.com/ahmetk3436/git-stats-golang/pkg/interfaces"
	"github.com/google/go-github/v56/github"
)

// GitHubRepo implements the interfaces.GitService for GitHub.
// It uses the go-github client to interact with the GitHub API.
type GitHubRepo struct {
	Client *github.Client // Client is the GitHub API client.
}

// NewGithubRepo creates a new GitHubRepo instance.
// It requires a non-nil github.Client.
func NewGithubRepo(gitHubClient *github.Client) (*GitHubRepo, error) {
	if gitHubClient == nil {
		return nil, fmt.Errorf("github client is nil, cannot create GitHubRepo")
	}
	return &GitHubRepo{
		Client: gitHubClient,
	}, nil
}

// ConnectGithub creates a new GitHub API client authenticated with the provided token.
// This is a helper function for initializing the GitHubRepo and is not part of the GitService interface.
func ConnectGithub(token string) *github.Client {
	// TODO: Consider adding a context for cancellation or timeout if needed for client creation.
	client := github.NewClient(nil).WithAuthToken(token)
	return client
}

// toCommonRepository converts a GitHub specific repository object to the common_types.Repository.
func toCommonRepository(ghRepo *github.Repository) *common_types.Repository {
	if ghRepo == nil {
		return nil // Or return an empty common_types.Repository{}, depending on desired behavior for nil inputs.
	}
	ownerLogin := ""
	if ghRepo.Owner != nil && ghRepo.Owner.Login != nil {
		ownerLogin = *ghRepo.Owner.Login
	}
	// Ensure timestamps are correctly handled if they can be nil from the API.
	// GetCreatedAt() and GetUpdatedAt() return github.Timestamp, which has a .Time field.
	return &common_types.Repository{
		ID:          ghRepo.GetID(), // GetID returns int64, matching common_types
		Name:        ghRepo.GetName(),
		Owner:       ownerLogin,
		HTMLURL:     ghRepo.GetHTMLURL(),
		CloneURL:    ghRepo.GetCloneURL(),
		Description: ghRepo.GetDescription(),
		CreatedAt:   ghRepo.GetCreatedAt().Time, // .Time is already a time.Time
		UpdatedAt:   ghRepo.GetUpdatedAt().Time, // .Time is already a time.Time
		Stars:       ghRepo.GetStargazersCount(),
		Forks:       ghRepo.GetForksCount(),
		OpenIssues:  ghRepo.GetOpenIssuesCount(),
	}
}

// toCommonCommit converts a GitHub specific commit object to the common_types.Commit.
// It requires the GitHub client to fetch detailed commit stats if not available in the initial commit object.
// owner and repoName are necessary for the GetCommit API call.
func toCommonCommit(ghCommit *github.RepositoryCommit, ghClient *github.Client, ownerLogin string, repoName string, commitSHA string) (*common_types.Commit, error) {
	if ghCommit == nil {
		return nil, fmt.Errorf("github repository commit is nil")
	}
	if ghClient == nil {
		return nil, fmt.Errorf("github client is nil, cannot fetch detailed commit stats")
	}

	// The ListCommits endpoint (which often provides ghCommit) may not populate stats.
	// A separate call to GetCommit is usually needed for detailed stats.
	// This makes an additional API call per commit, which can be a performance consideration.
	// For now, we prioritize getting complete data.
	// TODO: Add logging for this operation, e.g., log.Debugf("Fetching detailed commit for SHA %s", commitSHA)
	detailedCommit, _, err := ghClient.Repositories.GetCommit(context.Background(), ownerLogin, repoName, commitSHA, nil)
	if err != nil {
		// If fetching detailed commit fails, log the error and proceed with basic info. Stats will be zero.
		// Consider how critical stats are; if absolutely required, this could return an error.
		// log.Warnf("Could not fetch detailed commit for %s/%s SHA %s: %v. Proceeding with basic info.", owner, repoName, commitSHA, err)
		// For now, we'll allow creation of common_types.Commit even if detailed stats are missing.
	}

	var stats common_types.CommitStats
	if detailedCommit != nil && detailedCommit.Stats != nil {
		stats = common_types.CommitStats{
			Additions: detailedCommit.Stats.GetAdditions(),
			Deletions: detailedCommit.Stats.GetDeletions(),
			Total:     detailedCommit.Stats.GetTotal(),
		}
	} else if ghCommit.GetStats() != nil { // Fallback to stats from the list view if available (rarely populated)
		stats = common_types.CommitStats{
			Additions: ghCommit.GetStats().GetAdditions(),
			Deletions: ghCommit.GetStats().GetDeletions(),
			Total:     ghCommit.GetStats().GetTotal(),
		}
	}

	author := common_types.CommitAuthor{}
	// GitHub's RepositoryCommit has Author (github.User) and Commit.Author (github.CommitAuthor)
	// Prefer Commit.Author for actual authorship time, Author is the committer if different.
	commitAuthor := ghCommit.GetCommit().GetAuthor()
	if commitAuthor != nil {
		author.Name = commitAuthor.GetName()
		author.Email = commitAuthor.GetEmail()
		if commitAuthor.Date != nil {
			author.Date = commitAuthor.GetDate().Time
		}
	} else if ghCommit.GetAuthor() != nil { // Fallback to the top-level User if Commit.Author is nil
		// This case might indicate a commit made by a GitHub user who isn't the original author.
		// The 'Author' field on RepositoryCommit usually represents the GitHub user who made the commit (if available).
		// The 'Commit.Author' field represents the author information from the commit data itself.
		// For common_types.CommitAuthor, we typically want the original author info.
		// This fallback might be less accurate for authorship date.
		author.Name = ghCommit.GetAuthor().GetLogin() // Or GetName if available and preferred
		// Email might not be available on the User object from ListCommits.
	}

	return &common_types.Commit{
		SHA:     ghCommit.GetSHA(),
		Author:  author,
		Message: ghCommit.GetCommit().GetMessage(),
		HTMLURL: ghCommit.GetHTMLURL(),
		Stats:   stats,
	}, nil
}

// toCommonUser converts a GitHub specific user object to the common_types.User.
func toCommonUser(ghUser *github.User) *common_types.User {
	if ghUser == nil {
		return nil
	}
	return &common_types.User{
		Login:     ghUser.GetLogin(),
		ID:        ghUser.GetID(),
		AvatarURL: ghUser.GetAvatarURL(),
		HTMLURL:   ghUser.GetHTMLURL(),
		Name:      ghUser.GetName(), // GitHub User objects usually have Name.
	}
}

// GetAllRepos implements interfaces.GitService.
// It retrieves repositories based on the provided owner string.
// If owner is empty, it lists repositories for the authenticated user (owner, collaborator, org member).
// If owner is provided, it lists repositories for that specific organization.
func (ghRepo *GitHubRepo) GetAllRepos(ownerLogin string) ([]*common_types.Repository, error) {
	ctx := context.Background() // TODO: Consider passing context from callers.
	var githubRepositories []*github.Repository
	var err error

	// Default ListOptions for pagination.
	// TODO: These could be exposed or configured if more flexibility is needed.
	listOptions := github.ListOptions{PerPage: 100}

	if ownerLogin == "" {
		// List repositories for the authenticated user.
		repoListOptions := github.RepositoryListOptions{
			ListOptions: listOptions,
			Affiliation: "owner,collaborator,organization_member", // Includes owned, collaborated, and org member repos.
			Sort:        "updated",                                // Sort by last updated.
			Direction:   "desc",                                   // Descending order.
		}
		githubRepositories, _, err = ghRepo.Client.Repositories.List(ctx, "", &repoListOptions)
	} else {
		// List repositories for a specific organization.
		repoListByOrgOptions := github.RepositoryListByOrgOptions{
			ListOptions: listOptions,
			Type:        "all", // Includes all types: public, private, forks, sources, member.
		}
		githubRepositories, _, err = ghRepo.Client.Repositories.ListByOrg(ctx, ownerLogin, &repoListByOrgOptions)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list github repositories (owner: '%s'): %w", ownerLogin, err)
	}

	commonRepos := make([]*common_types.Repository, 0, len(githubRepositories))
	for _, githubRepository := range githubRepositories {
		commonRepos = append(commonRepos, toCommonRepository(githubRepository))
	}
	return commonRepos, nil
}

// GetRepo implements interfaces.GitService.
// identifier can be an int64 (GitHub Repository ID) or a string "ownerLogin/repoName".
func (ghRepo *GitHubRepo) GetRepo(identifier interface{}) (*common_types.Repository, error) {
	ctx := context.Background() // TODO: Pass context from callers.
	var githubRepository *github.Repository
	var err error

	switch id := identifier.(type) {
	case int64:
		githubRepository, _, err = ghRepo.Client.Repositories.GetByID(ctx, id)
	case string:
		parts := strings.Split(id, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid repo identifier string format: expected 'ownerLogin/repoName', got '%s'", id)
		}
		owner, name := parts[0], parts[1]
		githubRepository, _, err = ghRepo.Client.Repositories.Get(ctx, owner, name)
	default:
		return nil, fmt.Errorf("unsupported identifier type for GetRepo: %T (expected int64 or string)", identifier)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get github repository (identifier: '%v'): %w", identifier, err)
	}
	return toCommonRepository(githubRepository), nil
}

// GetProjectCommits implements interfaces.GitService.
// repoIdentifier can be an int64 (GitHub Repository ID) or a string "ownerLogin/repoName".
// options allows for filtering by SHA (branch/tag/commit), Path, Author, and pagination.
func (ghRepo *GitHubRepo) GetProjectCommits(repoIdentifier interface{}, options *interfaces.CommitListOptions) ([]*common_types.Commit, error) {
	ctx := context.Background() // TODO: Pass context.
	var ownerLogin, repositoryName string

	// Determine ownerLogin and repositoryName from repoIdentifier.
	targetRepo, err := ghRepo.GetRepo(repoIdentifier) // Leverage existing GetRepo to resolve identifier.
	if err != nil {
		return nil, fmt.Errorf("failed to get repository details for listing commits (identifier: '%v'): %w", repoIdentifier, err)
	}
	if targetRepo == nil { // Should be caught by GetRepo's error handling.
		return nil, fmt.Errorf("repository not found for identifier: %v", repoIdentifier)
	}
	ownerLogin = targetRepo.Owner
	repositoryName = targetRepo.Name

	if ownerLogin == "" || repositoryName == "" { // Should not happen if GetRepo succeeded with a valid repo.
		return nil, fmt.Errorf("could not determine owner and repository name for identifier: %v", repoIdentifier)
	}

	commitListOpts := github.CommitsListOptions{
		ListOptions: github.ListOptions{PerPage: 100}, // Default PerPage.
	}
	if options != nil {
		if options.SHA != "" {
			commitListOpts.SHA = options.SHA
		}
		if options.Path != "" {
			commitListOpts.Path = options.Path
		}
		if options.Author != "" {
			// GitHub API uses author's login or email for filtering.
			// Assuming options.Author is one of these.
			commitListOpts.Author = options.Author
		}
		if options.Page > 0 {
			commitListOpts.ListOptions.Page = options.Page
		}
		if options.PerPage > 0 {
			commitListOpts.ListOptions.PerPage = options.PerPage
		}
	}

	githubCommits, _, err := ghRepo.Client.Repositories.ListCommits(ctx, ownerLogin, repositoryName, &commitListOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list github commits for %s/%s: %w", ownerLogin, repositoryName, err)
	}

	commonCommits := make([]*common_types.Commit, 0, len(githubCommits))
	for _, githubCommit := range githubCommits {
		commonCommit, conversionErr := toCommonCommit(githubCommit, ghRepo.Client, ownerLogin, repositoryName, githubCommit.GetSHA())
		if conversionErr != nil {
			// Log or handle error converting individual commit.
			// For now, skip bad ones and log a warning.
			// log.Warnf("Could not convert commit %s for %s/%s: %v", githubCommit.GetSHA(), ownerLogin, repositoryName, conversionErr)
			fmt.Printf("Warning: Could not convert commit %s for %s/%s: %v\n", githubCommit.GetSHA(), ownerLogin, repositoryName, conversionErr) // Temporary logging
			continue
		}
		commonCommits = append(commonCommits, commonCommit)
	}
	return commonCommits, nil
}

// GetRepoContributors implements interfaces.GitService.
// repoIdentifier can be an int64 (GitHub Repository ID) or a string "ownerLogin/repoName".
// Fetches contributors for a repository and maps them to common_types.User.
// Note: GitHub's "contributor" might be an anonymous user or a full GitHub user.
// The common_types.User.Name might be empty if not available directly from the contributor stats.
func (ghRepo *GitHubRepo) GetRepoContributors(repoIdentifier interface{}) ([]*common_types.User, error) {
	ctx := context.Background() // TODO: Pass context.
	var ownerLogin, repositoryName string

	targetRepo, err := ghRepo.GetRepo(repoIdentifier)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository details for listing contributors (identifier: '%v'): %w", repoIdentifier, err)
	}
	if targetRepo == nil {
		return nil, fmt.Errorf("repository not found for identifier: %v", repoIdentifier)
	}
	ownerLogin = targetRepo.Owner
	repositoryName = targetRepo.Name

	if ownerLogin == "" || repositoryName == "" {
		return nil, fmt.Errorf("could not determine owner and repository name for identifier: %v", repoIdentifier)
	}

	// Options for listing contributors, e.g., include anonymous.
	// For now, using default options (which usually exclude anonymous unless explicitly requested).
	contributorOpts := &github.ListContributorsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	// The GitHub API returns an array of `github.Contributor` objects.
	// A `github.Contributor` contains fields like Login, ID, AvatarURL, HTMLURL, Contributions, etc.
	// It's essentially a `github.User` with an additional `Contributions` field.
	githubContributors, _, err := ghRepo.Client.Repositories.ListContributors(ctx, ownerLogin, repositoryName, contributorOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list github contributors for %s/%s: %w", ownerLogin, repositoryName, err)
	}

	commonContributors := make([]*common_types.User, 0, len(githubContributors))
	for _, ghContributor := range githubContributors {
		if ghContributor.GetLogin() == "" { // Skip if login is empty (e.g. some types of anonymous)
			continue
		}
		// Map github.Contributor to common_types.User.
		// github.Contributor has most fields of github.User.
		// We can effectively treat ghContributor as a ghUser for toCommonUser.
		// However, toCommonUser expects *github.User.
		// Create a temporary github.User from github.Contributor for mapping.
		tempGhUser := &github.User{
			Login:     ghContributor.Login,
			ID:        ghContributor.ID,
			AvatarURL: ghContributor.AvatarURL,
			HTMLURL:   ghContributor.HTMLURL,
			Name:      ghContributor.Name, // Name might or might not be populated on Contributor object.
			// Other fields of github.User are not present in github.Contributor,
			// but toCommonUser handles nils for those.
		}
		commonUser := toCommonUser(tempGhUser)
		if commonUser != nil {
			commonContributors = append(commonContributors, commonUser)
		}
	}
	return commonContributors, nil
}

// Ensure GitHubRepo implements GitService.
// This line will cause a compile-time error if the interface is not properly implemented.
var _ interfaces.GitService = (*GitHubRepo)(nil)
