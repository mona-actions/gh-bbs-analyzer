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
	// New methods for enhanced migration analysis
	GetBranches(projectKey, repoSlug string) ([]bitbucket.Branch, error)
	GetTags(projectKey, repoSlug string) ([]bitbucket.Tag, error)
	GetUserPermissions(projectKey, repoSlug string) ([]bitbucket.UserPermission, error)
	GetHooks(projectKey, repoSlug string) ([]bitbucket.Hook, error)
	GetWebhooks(projectKey, repoSlug string) ([]bitbucket.Webhook, error)
	GetLFSFiles(projectKey, repoSlug string) ([]bitbucket.LFSFile, error)
	GetCommits(projectKey, repoSlug string, limit int) ([]bitbucket.Commit, error)
	GetForks(projectKey, repoSlug string) ([]bitbucket.Fork, error)
}

// Ensure the bitbucket.Client implements our interface
var _ BitBucketClient = (*bitbucket.Client)(nil)
