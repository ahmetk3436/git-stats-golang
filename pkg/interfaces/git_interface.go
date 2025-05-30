package interfaces

import (
	"github.com/ahmetk3436/git-stats-golang/pkg/common_types"
)

// GitService defines a generic interface for performing common Git operations
// across different providers like GitHub, GitLab, etc.
// Implementations of this interface will handle provider-specific API interactions
// and map the results to the provider-agnostic common_types.
type GitService interface {
	// GetAllRepos retrieves all repositories accessible by the authenticated user or for a specific owner.
	// The 'owner' parameter specifies the user or organization whose repositories are to be listed.
	// If 'owner' is an empty string, implementations may list repositories accessible by the authenticated user
	// (e.g., owned, member, collaborator repos), behavior might vary by provider.
	GetAllRepos(owner string) ([]*common_types.Repository, error)

	// GetRepo retrieves a specific repository.
	// The 'identifier' can be a provider-specific ID (e.g., int64 for GitHub, int for GitLab)
	// or a string in the format "owner/repository_name".
	// Implementations are responsible for parsing and handling the identifier appropriately.
	GetRepo(identifier interface{}) (*common_types.Repository, error)

	// GetProjectCommits retrieves commits for a specific repository.
	// The 'repoIdentifier' is similar to GetRepo's 'identifier'.
	// 'options' allows specifying parameters like branch/SHA, path, author, and pagination.
	// If 'options' is nil, default values will be used by the implementation.
	GetProjectCommits(repoIdentifier interface{}, options *CommitListOptions) ([]*common_types.Commit, error)

	// GetRepoContributors retrieves contributors for a specific repository.
	// The 'repoIdentifier' is similar to GetRepo's 'identifier'.
	// Note: The level of detail in common_types.User for contributors may vary
	// depending on what the provider's API returns for contributors.
	GetRepoContributors(repoIdentifier interface{}) ([]*common_types.User, error)
}

// CommitListOptions provides optional parameters for listing commits.
// Zero-values for fields usually mean the provider's default will be used.
type CommitListOptions struct {
	SHA     string // Branch name, tag name, or commit SHA to list commits from. Empty for default branch.
	Path    string // File path to filter commits by. Empty if not filtering by path.
	Author  string // Commit author (e.g., email or username) to filter by. Empty if not filtering by author.
	Page    int    // Page number for pagination. Typically 1-based. 0 or 1 means first page.
	PerPage int    // Number of items per page for pagination. 0 means provider's default.
}
