package analyzer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mona-actions/gh-bbs-analyzer/internal/bitbucket"
	"github.com/mona-actions/gh-bbs-analyzer/internal/config"
	"github.com/mona-actions/gh-bbs-analyzer/internal/output"
)

// Results holds the analysis results
type Results struct {
	Projects     []bitbucket.Project
	Repositories []bitbucket.Repository
	Statistics   Statistics
}

// Statistics holds summary statistics
type Statistics struct {
	TotalProjects int
	TotalRepos    int
	TotalPRs      int
	TotalComments int
	TotalSize     int64
	FormattedSize string
	// New migration-relevant metrics
	TotalBranches    int
	TotalTags        int
	TotalForks       int
	TotalWebhooks    int
	TotalLFSFiles    int
	TotalLFSSize     int64
	FormattedLFSSize string
	TotalPermissions int
	TotalHooks       int
	ActiveRepos      int // Repos with commits in last 90 days
	ArchivedRepos    int // Archived repositories
}

// Analyzer handles the analysis of BitBucket repositories using dependency injection
// This version is testable as it doesn't create its own dependencies
type Analyzer struct {
	client BitBucketClient
	logger *output.Logger
	config *config.Config
}

// NewAnalyzer creates a new analyzer with dependency injection
func NewAnalyzer(client BitBucketClient, logger *output.Logger, cfg *config.Config) *Analyzer {
	return &Analyzer{
		client: client,
		logger: logger,
		config: cfg,
	}
}

// NewAnalyzerFromConfig creates a new analyzer with a real BitBucket client
// This maintains the existing API while enabling dependency injection
func NewAnalyzerFromConfig(cfg *config.Config, logger *output.Logger) *Analyzer {
	client := bitbucket.NewClient(
		cfg.BitBucket.ServerURL,
		cfg.BitBucket.Username,
		cfg.BitBucket.Password,
		cfg.NoSSLVerify,
	)

	return NewAnalyzer(client, logger, cfg)
}

// Run performs the complete analysis using the injected client
func (a *Analyzer) Run(ctx context.Context) (*Results, error) {
	a.logger.Info("Starting repository analysis")

	// Get projects
	projects, err := a.getProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	if len(projects) == 0 {
		return nil, fmt.Errorf("no projects found")
	}

	// Get all repositories
	var allRepos []bitbucket.Repository
	for _, project := range projects {
		repos, err := a.client.GetRepositories(project.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to get repositories for project %s: %w", project.Key, err)
		}
		allRepos = append(allRepos, repos...)
	}

	// Analyze repositories concurrently
	analyzedRepos, err := a.analyzeRepositories(ctx, allRepos)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze repositories: %w", err)
	}

	// Calculate statistics
	stats := a.calculateStatistics(projects, analyzedRepos)

	return &Results{
		Projects:     projects,
		Repositories: analyzedRepos,
		Statistics:   stats,
	}, nil
}

// getProjects retrieves projects based on configuration
func (a *Analyzer) getProjects() ([]bitbucket.Project, error) {
	if a.config.BitBucket.Project != "" {
		// Get specific project
		project, err := a.client.GetProject(a.config.BitBucket.Project)
		if err != nil {
			return nil, err
		}
		return []bitbucket.Project{*project}, nil
	}

	// Get all projects
	return a.client.GetProjects()
}

// analyzeRepositories analyzes repositories concurrently
func (a *Analyzer) analyzeRepositories(ctx context.Context, repos []bitbucket.Repository) ([]bitbucket.Repository, error) {
	maxWorkers := a.config.Threads
	if maxWorkers > len(repos) {
		maxWorkers = len(repos)
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results = make([]bitbucket.Repository, len(repos))
		errCh   = make(chan error, len(repos))
	)

	// Create worker pool
	semaphore := make(chan struct{}, maxWorkers)

	for i, repo := range repos {
		wg.Add(1)
		go func(idx int, repository bitbucket.Repository) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Check context cancellation
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
			}

			// Analyze repository
			analyzedRepo, err := a.analyzeRepository(repository)
			if err != nil {
				errCh <- fmt.Errorf("failed to analyze repository %s: %w", repository.Name, err)
				return
			}

			// Store result
			mu.Lock()
			results[idx] = analyzedRepo
			mu.Unlock()

			a.logger.Info(fmt.Sprintf("Completed analysis of repository: %s", repository.Name))
		}(i, repo)
	}

	// Wait for all workers to complete
	wg.Wait()
	close(errCh)

	// Check for errors
	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

// analyzeRepository analyzes a single repository
func (a *Analyzer) analyzeRepository(repo bitbucket.Repository) (bitbucket.Repository, error) {
	// Get repository size
	size, err := a.client.GetRepositorySize(repo.Project.Key, repo.Slug)
	if err != nil {
		a.logger.Warning(fmt.Sprintf("Failed to get size for repository %s: %v", repo.Name, err))
		// Continue without size information
	}
	repo.Size = size

	// Get pull requests
	pullRequests, err := a.client.GetPullRequests(repo.Project.Key, repo.Slug)
	if err != nil {
		a.logger.Warning(fmt.Sprintf("Failed to get pull requests for repository %s: %v", repo.Name, err))
		// Continue without PR information
	}
	repo.PullRequests = pullRequests

	// Calculate comment count
	commentCount := 0
	for _, pr := range pullRequests {
		commentCount += pr.Properties.CommentCount
	}
	repo.CommentCount = commentCount

	// Get branches
	branches, err := a.client.GetBranches(repo.Project.Key, repo.Slug)
	if err != nil {
		a.logger.Warning(fmt.Sprintf("Failed to get branches for repository %s: %v", repo.Name, err))
	} else {
		repo.BranchCount = len(branches)
		// Find default branch
		for _, branch := range branches {
			if branch.IsDefault {
				repo.DefaultBranch = branch.DisplayID
				break
			}
		}
	}

	// Get tags
	tags, err := a.client.GetTags(repo.Project.Key, repo.Slug)
	if err != nil {
		a.logger.Warning(fmt.Sprintf("Failed to get tags for repository %s: %v", repo.Name, err))
	} else {
		repo.TagCount = len(tags)
	}

	// Get user permissions
	permissions, err := a.client.GetUserPermissions(repo.Project.Key, repo.Slug)
	if err != nil {
		a.logger.Warning(fmt.Sprintf("Failed to get permissions for repository %s: %v", repo.Name, err))
	} else {
		repo.UserPermissions = permissions
	}

	// Get hooks
	hooks, err := a.client.GetHooks(repo.Project.Key, repo.Slug)
	if err != nil {
		a.logger.Warning(fmt.Sprintf("Failed to get hooks for repository %s: %v", repo.Name, err))
	} else {
		repo.Hooks = hooks
	}

	// Get webhooks
	webhooks, err := a.client.GetWebhooks(repo.Project.Key, repo.Slug)
	if err != nil {
		a.logger.Warning(fmt.Sprintf("Failed to get webhooks for repository %s: %v", repo.Name, err))
	} else {
		repo.WebhookCount = len(webhooks)
	}

	// Get LFS files
	lfsFiles, err := a.client.GetLFSFiles(repo.Project.Key, repo.Slug)
	if err != nil {
		a.logger.Warning(fmt.Sprintf("Failed to get LFS files for repository %s: %v", repo.Name, err))
	} else {
		repo.LFSFileCount = len(lfsFiles)
		var lfsSize int64
		for _, file := range lfsFiles {
			lfsSize += file.Size
		}
		repo.LFSSize = lfsSize
	}

	// Get recent commits to determine activity
	commits, err := a.client.GetCommits(repo.Project.Key, repo.Slug, 10)
	if err != nil {
		a.logger.Warning(fmt.Sprintf("Failed to get commits for repository %s: %v", repo.Name, err))
	} else {
		repo.CommitCount = len(commits)
		// Extract unique contributors
		contributors := make(map[string]bool)
		for _, commit := range commits {
			contributors[commit.Author.EmailAddress] = true
		}
		repo.ContributorCount = len(contributors)

		// Set last activity if we have commits
		if len(commits) > 0 {
			// Use the author timestamp of the most recent commit
			if commits[0].AuthorDate > 0 {
				lastActivityTime := time.Unix(commits[0].AuthorDate/1000, 0) // BitBucket uses milliseconds
				repo.LastActivity = lastActivityTime.Format("2006-01-02")
			} else {
				repo.LastActivity = "unknown"
			}
		}
	}

	// Get forks
	forks, err := a.client.GetForks(repo.Project.Key, repo.Slug)
	if err != nil {
		a.logger.Warning(fmt.Sprintf("Failed to get forks for repository %s: %v", repo.Name, err))
	} else {
		repo.ForkCount = len(forks)
	}

	return repo, nil
}

// calculateStatistics calculates summary statistics
func (a *Analyzer) calculateStatistics(projects []bitbucket.Project, repos []bitbucket.Repository) Statistics {
	var (
		totalPRs         int
		totalComments    int
		totalSize        int64
		totalBranches    int
		totalTags        int
		totalForks       int
		totalWebhooks    int
		totalLFSFiles    int
		totalLFSSize     int64
		totalPermissions int
		totalHooks       int
		activeRepos      int
		archivedRepos    int
	)

	for _, repo := range repos {
		totalPRs += len(repo.PullRequests)
		totalComments += repo.CommentCount
		totalBranches += repo.BranchCount
		totalTags += repo.TagCount
		totalForks += repo.ForkCount
		totalWebhooks += repo.WebhookCount
		totalLFSFiles += repo.LFSFileCount
		totalLFSSize += repo.LFSSize
		totalPermissions += len(repo.UserPermissions)
		totalHooks += len(repo.Hooks)

		if repo.Size != nil {
			totalSize += int64(repo.Size.Repository)
		}

		// Count archived repositories
		if repo.Archived {
			archivedRepos++
		}

		// Count active repositories (those with recent commits)
		if repo.CommitCount > 0 {
			activeRepos++
		}
	}

	formattedSize := FormatSize(totalSize)
	formattedLFSSize := FormatSize(totalLFSSize)

	return Statistics{
		TotalProjects:    len(projects),
		TotalRepos:       len(repos),
		TotalPRs:         totalPRs,
		TotalComments:    totalComments,
		TotalSize:        totalSize,
		FormattedSize:    formattedSize,
		TotalBranches:    totalBranches,
		TotalTags:        totalTags,
		TotalForks:       totalForks,
		TotalWebhooks:    totalWebhooks,
		TotalLFSFiles:    totalLFSFiles,
		TotalLFSSize:     totalLFSSize,
		FormattedLFSSize: formattedLFSSize,
		TotalPermissions: totalPermissions,
		TotalHooks:       totalHooks,
		ActiveRepos:      activeRepos,
		ArchivedRepos:    archivedRepos,
	}
}

// FormatSize formats bytes into human-readable format (exported for testing)
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%d GB", bytes/GB)
	case bytes >= MB:
		return fmt.Sprintf("%d MB", bytes/MB)
	case bytes >= KB:
		return fmt.Sprintf("%d KB", bytes/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
