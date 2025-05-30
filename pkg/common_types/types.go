package common_types

import "time"

// Repository holds common, provider-agnostic repository information.
// This struct is used to standardize repository data from different Git providers.
type Repository struct {
	ID          int64     // Unique identifier for the repository.
	Name        string    // Name of the repository.
	Owner       string    // Login name of the repository owner (user or organization).
	HTMLURL     string    // URL to the repository's main page.
	CloneURL    string    // URL used for cloning the repository (typically HTTPS).
	Description string    // Short description of the repository.
	CreatedAt   time.Time // Timestamp when the repository was created.
	UpdatedAt   time.Time // Timestamp when the repository was last updated.
	Stars       int       // Number of stars or likes.
	Forks       int       // Number of forks.
	OpenIssues  int       // Number of open issues.
}

// CommitAuthor holds common, provider-agnostic commit author information.
// This includes details about the person who authored the commit.
type CommitAuthor struct {
	Name  string    // Name of the commit author.
	Email string    // Email address of the commit author.
	Date  time.Time // Timestamp when the commit was authored.
}

// Commit holds common, provider-agnostic commit information.
// This struct standardizes commit data from different Git providers.
type Commit struct {
	SHA     string       // SHA hash of the commit.
	Author  CommitAuthor // Information about the commit author.
	Message string       // Commit message.
	HTMLURL string       // URL to the commit's page.
	Stats   CommitStats  // Statistics related to the commit (additions, deletions).
}

// CommitStats holds common, provider-agnostic commit statistics.
// This includes the number of additions, deletions, and total changes.
type CommitStats struct {
	Additions int // Number of lines added.
	Deletions int // Number of lines deleted.
	Total     int // Total number of lines changed (additions + deletions).
}

// User holds common, provider-agnostic user information.
// This struct is simplified and standardizes user data from different Git providers.
type User struct {
	Login     string // Username or login identifier.
	ID        int64  // Unique identifier for the user.
	AvatarURL string // URL to the user's avatar image.
	HTMLURL   string // URL to the user's profile page.
	Name      string // Real name of the user, if available (might be empty).
}
