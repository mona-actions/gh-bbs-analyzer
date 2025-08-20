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
