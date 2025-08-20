package bitbucket

// Project represents a BitBucket project
type Project struct {
	Key    string `json:"key"`
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Public bool   `json:"public"`
	Type   string `json:"type"`
}

// ProjectResponse represents the API response for projects
type ProjectResponse struct {
	Size          int       `json:"size"`
	Limit         int       `json:"limit"`
	IsLastPage    bool      `json:"isLastPage"`
	Values        []Project `json:"values"`
	Start         int       `json:"start"`
	Filter        string    `json:"filter"`
	NextPageStart int       `json:"nextPageStart"`
}

// Repository represents a BitBucket repository
type Repository struct {
	Slug          string          `json:"slug"`
	ID            int             `json:"id"`
	Name          string          `json:"name"`
	HierarchyId   string          `json:"hierarchyId"`
	ScmId         string          `json:"scmId"`
	State         string          `json:"state"`
	StatusMessage string          `json:"statusMessage"`
	Forkable      bool            `json:"forkable"`
	Public        bool            `json:"public"`
	Archived      bool            `json:"archived"`
	Project       Project         `json:"project"`
	Size          *RepositorySize `json:"-"` // Not part of API response
	PullRequests  []PullRequest   `json:"-"` // Not part of API response
	CommentCount  int             `json:"-"` // Not part of API response
	// Additional migration analysis data
	BranchCount      int              `json:"-"` // Not part of API response
	DefaultBranch    string           `json:"-"` // Not part of API response
	TagCount         int              `json:"-"` // Not part of API response
	ForkCount        int              `json:"-"` // Not part of API response
	WebhookCount     int              `json:"-"` // Not part of API response
	LFSFileCount     int              `json:"-"` // Not part of API response
	LFSSize          int64            `json:"-"` // Not part of API response
	CommitCount      int              `json:"-"` // Not part of API response
	ContributorCount int              `json:"-"` // Not part of API response
	LastActivity     string           `json:"-"` // Not part of API response
	UserPermissions  []UserPermission `json:"-"` // Not part of API response
	Hooks            []Hook           `json:"-"` // Not part of API response
}

// RepositoryResponse represents the API response for repositories
type RepositoryResponse struct {
	Size          int          `json:"size"`
	Limit         int          `json:"limit"`
	IsLastPage    bool         `json:"isLastPage"`
	Values        []Repository `json:"values"`
	Start         int          `json:"start"`
	Filter        string       `json:"filter"`
	NextPageStart int          `json:"nextPageStart"`
}

// RepositorySize represents repository size information
type RepositorySize struct {
	Repository  int `json:"repository"`
	Attachments int `json:"attachments"`
}

// PullRequest represents a BitBucket pull request
type PullRequest struct {
	ID         int                   `json:"id"`
	Properties PullRequestProperties `json:"properties"`
}

// PullRequestProperties represents pull request properties
type PullRequestProperties struct {
	CommentCount int `json:"commentCount"`
}

// PullRequestResponse represents the API response for pull requests
type PullRequestResponse struct {
	Size          int           `json:"size"`
	Limit         int           `json:"limit"`
	IsLastPage    bool          `json:"isLastPage"`
	Values        []PullRequest `json:"values"`
	Start         int           `json:"start"`
	Filter        string        `json:"filter"`
	NextPageStart int           `json:"nextPageStart"`
}

// Branch represents a BitBucket branch
type Branch struct {
	ID              string `json:"id"`
	DisplayID       string `json:"displayId"`
	Type            string `json:"type"`
	LatestCommit    string `json:"latestCommit"`
	LatestChangeset string `json:"latestChangeset"`
	IsDefault       bool   `json:"isDefault"`
}

// BranchResponse represents the API response for branches
type BranchResponse struct {
	Size          int      `json:"size"`
	Limit         int      `json:"limit"`
	IsLastPage    bool     `json:"isLastPage"`
	Values        []Branch `json:"values"`
	Start         int      `json:"start"`
	NextPageStart int      `json:"nextPageStart"`
}

// Tag represents a BitBucket tag
type Tag struct {
	ID              string `json:"id"`
	DisplayID       string `json:"displayId"`
	Type            string `json:"type"`
	LatestCommit    string `json:"latestCommit"`
	LatestChangeset string `json:"latestChangeset"`
}

// TagResponse represents the API response for tags
type TagResponse struct {
	Size          int   `json:"size"`
	Limit         int   `json:"limit"`
	IsLastPage    bool  `json:"isLastPage"`
	Values        []Tag `json:"values"`
	Start         int   `json:"start"`
	NextPageStart int   `json:"nextPageStart"`
}

// UserPermission represents user permissions on a repository
type UserPermission struct {
	User       User   `json:"user"`
	Permission string `json:"permission"`
}

// User represents a BitBucket user
type User struct {
	Name         string `json:"name"`
	EmailAddress string `json:"emailAddress"`
	ID           int    `json:"id"`
	DisplayName  string `json:"displayName"`
	Active       bool   `json:"active"`
	Slug         string `json:"slug"`
	Type         string `json:"type"`
}

// UserPermissionResponse represents the API response for user permissions
type UserPermissionResponse struct {
	Size          int              `json:"size"`
	Limit         int              `json:"limit"`
	IsLastPage    bool             `json:"isLastPage"`
	Values        []UserPermission `json:"values"`
	Start         int              `json:"start"`
	NextPageStart int              `json:"nextPageStart"`
}

// Hook represents a repository hook
type Hook struct {
	Details    HookDetails `json:"details"`
	Enabled    bool        `json:"enabled"`
	Configured bool        `json:"configured"`
	Scope      HookScope   `json:"scope"`
}

// HookDetails represents hook details
type HookDetails struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

// HookScope represents hook scope
type HookScope struct {
	Type string `json:"type"`
}

// HookResponse represents the API response for hooks
type HookResponse struct {
	Size          int    `json:"size"`
	Limit         int    `json:"limit"`
	IsLastPage    bool   `json:"isLastPage"`
	Values        []Hook `json:"values"`
	Start         int    `json:"start"`
	NextPageStart int    `json:"nextPageStart"`
}

// Webhook represents a repository webhook
type Webhook struct {
	ID     int           `json:"id"`
	Name   string        `json:"name"`
	URL    string        `json:"url"`
	Events []string      `json:"events"`
	Active bool          `json:"active"`
	Config WebhookConfig `json:"configuration"`
}

// WebhookConfig represents webhook configuration
type WebhookConfig struct {
	Secret string `json:"secret"`
}

// WebhookResponse represents the API response for webhooks
type WebhookResponse struct {
	Size          int       `json:"size"`
	Limit         int       `json:"limit"`
	IsLastPage    bool      `json:"isLastPage"`
	Values        []Webhook `json:"values"`
	Start         int       `json:"start"`
	NextPageStart int       `json:"nextPageStart"`
}

// LFSFile represents a Git LFS file
type LFSFile struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
	OID  string `json:"oid"`
}

// LFSResponse represents the API response for LFS files
type LFSResponse struct {
	Size          int       `json:"size"`
	Limit         int       `json:"limit"`
	IsLastPage    bool      `json:"isLastPage"`
	Values        []LFSFile `json:"values"`
	Start         int       `json:"start"`
	NextPageStart int       `json:"nextPageStart"`
}

// Commit represents a repository commit
type Commit struct {
	ID            string      `json:"id"`
	DisplayID     string      `json:"displayId"`
	Author        CommitUser  `json:"author"`
	AuthorDate    int64       `json:"authorTimestamp"`
	Committer     CommitUser  `json:"committer"`
	CommitterDate int64       `json:"committerTimestamp"`
	Message       string      `json:"message"`
	Parents       []CommitRef `json:"parents"`
}

// CommitUser represents commit author/committer
type CommitUser struct {
	Name         string `json:"name"`
	EmailAddress string `json:"emailAddress"`
}

// CommitRef represents a commit reference
type CommitRef struct {
	ID        string `json:"id"`
	DisplayID string `json:"displayId"`
}

// CommitResponse represents the API response for commits
type CommitResponse struct {
	Size          int      `json:"size"`
	Limit         int      `json:"limit"`
	IsLastPage    bool     `json:"isLastPage"`
	Values        []Commit `json:"values"`
	Start         int      `json:"start"`
	NextPageStart int      `json:"nextPageStart"`
}

// Fork represents a repository fork
type Fork struct {
	Slug    string  `json:"slug"`
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Project Project `json:"project"`
	Public  bool    `json:"public"`
}

// ForkResponse represents the API response for forks
type ForkResponse struct {
	Size          int    `json:"size"`
	Limit         int    `json:"limit"`
	IsLastPage    bool   `json:"isLastPage"`
	Values        []Fork `json:"values"`
	Start         int    `json:"start"`
	NextPageStart int    `json:"nextPageStart"`
}
