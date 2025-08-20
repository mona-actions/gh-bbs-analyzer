package analyzer

import (
	"github.com/mona-actions/gh-bbs-analyzer/internal/bitbucket"
)

// BitBucketClient interface defines the contract for BitBucket API operations
// This interface enables dependency injection and easier testing
type BitBucketClient interface {
	GetProject(projectKey string) (*bitbucket.Project, error)
	GetProjects() ([]bitbucket.Project, error)
	GetRepositories(projectKey string) ([]bitbucket.Repository, error)
	GetRepositorySize(projectKey, repoSlug string) (*bitbucket.RepositorySize, error)
	GetPullRequests(projectKey, repoSlug string) ([]bitbucket.PullRequest, error)
}

// Ensure the bitbucket.Client implements our interface
var _ BitBucketClient = (*bitbucket.Client)(nil)
