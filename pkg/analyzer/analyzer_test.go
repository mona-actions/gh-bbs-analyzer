package analyzer_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/mona-actions/gh-bbs-analyzer/internal/bitbucket"
	"github.com/mona-actions/gh-bbs-analyzer/internal/config"
	"github.com/mona-actions/gh-bbs-analyzer/internal/output"
	"github.com/mona-actions/gh-bbs-analyzer/pkg/analyzer"
)

// MockBitBucketClient implements a mock BitBucket client for testing
type MockBitBucketClient struct {
	projects     []bitbucket.Project
	repositories map[string][]bitbucket.Repository
	sizes        map[string]*bitbucket.RepositorySize
	pullRequests map[string][]bitbucket.PullRequest
	projectError error
	repoError    error
	sizeError    error
	prError      error
}

func NewMockBitBucketClient() *MockBitBucketClient {
	return &MockBitBucketClient{
		repositories: make(map[string][]bitbucket.Repository),
		sizes:        make(map[string]*bitbucket.RepositorySize),
		pullRequests: make(map[string][]bitbucket.PullRequest),
	}
}

func (m *MockBitBucketClient) GetProject(projectKey string) (*bitbucket.Project, error) {
	if m.projectError != nil {
		return nil, m.projectError
	}

	for _, project := range m.projects {
		if project.Key == projectKey {
			return &project, nil
		}
	}

	return nil, fmt.Errorf("project not found: %s", projectKey)
}

func (m *MockBitBucketClient) GetProjects() ([]bitbucket.Project, error) {
	if m.projectError != nil {
		return nil, m.projectError
	}
	return m.projects, nil
}

func (m *MockBitBucketClient) GetRepositories(projectKey string) ([]bitbucket.Repository, error) {
	if m.repoError != nil {
		return nil, m.repoError
	}
	return m.repositories[projectKey], nil
}

func (m *MockBitBucketClient) GetRepositorySize(projectKey, repoSlug string) (*bitbucket.RepositorySize, error) {
	if m.sizeError != nil {
		return nil, m.sizeError
	}
	key := fmt.Sprintf("%s/%s", projectKey, repoSlug)
	return m.sizes[key], nil
}

func (m *MockBitBucketClient) GetPullRequests(projectKey, repoSlug string) ([]bitbucket.PullRequest, error) {
	if m.prError != nil {
		return nil, m.prError
	}
	key := fmt.Sprintf("%s/%s", projectKey, repoSlug)
	return m.pullRequests[key], nil
}

// New methods for enhanced migration analysis - return empty data for tests
func (m *MockBitBucketClient) GetBranches(projectKey, repoSlug string) ([]bitbucket.Branch, error) {
	return []bitbucket.Branch{
		{ID: "refs/heads/main", DisplayID: "main", IsDefault: true},
		{ID: "refs/heads/develop", DisplayID: "develop", IsDefault: false},
	}, nil
}

func (m *MockBitBucketClient) GetTags(projectKey, repoSlug string) ([]bitbucket.Tag, error) {
	return []bitbucket.Tag{
		{ID: "refs/tags/v1.0.0", DisplayID: "v1.0.0"},
	}, nil
}

func (m *MockBitBucketClient) GetUserPermissions(projectKey, repoSlug string) ([]bitbucket.UserPermission, error) {
	return []bitbucket.UserPermission{
		{
			User: bitbucket.User{
				Name:        "testuser",
				DisplayName: "Test User",
				Active:      true,
			},
			Permission: "REPO_ADMIN",
		},
	}, nil
}

func (m *MockBitBucketClient) GetHooks(projectKey, repoSlug string) ([]bitbucket.Hook, error) {
	return []bitbucket.Hook{}, nil
}

func (m *MockBitBucketClient) GetWebhooks(projectKey, repoSlug string) ([]bitbucket.Webhook, error) {
	return []bitbucket.Webhook{}, nil
}

func (m *MockBitBucketClient) GetLFSFiles(projectKey, repoSlug string) ([]bitbucket.LFSFile, error) {
	return []bitbucket.LFSFile{}, nil
}

func (m *MockBitBucketClient) GetCommits(projectKey, repoSlug string, limit int) ([]bitbucket.Commit, error) {
	return []bitbucket.Commit{
		{
			ID:         "abc123",
			DisplayID:  "abc123",
			AuthorDate: 1693497600000, // August 31, 2023 in milliseconds
			Author: bitbucket.CommitUser{
				Name:         "testuser",
				EmailAddress: "test@example.com",
			},
			Message: "Test commit",
		},
	}, nil
}

func (m *MockBitBucketClient) GetForks(projectKey, repoSlug string) ([]bitbucket.Fork, error) {
	return []bitbucket.Fork{}, nil
}

// Test helper to create a test logger
func createTestLogger(t *testing.T) *output.Logger {
	return output.NewLogger()
}

func TestNewAnalyzer(t *testing.T) {
	cfg := &config.Config{
		BitBucket: config.BitBucketConfig{
			ServerURL: "https://test.com",
			Username:  "testuser",
			Password:  "testpass",
		},
		Threads: 3,
	}

	logger := createTestLogger(t)
	defer logger.Close()

	mockClient := NewMockBitBucketClient()

	analyzer := analyzer.NewAnalyzer(mockClient, logger, cfg)
	if analyzer == nil {
		t.Error("Expected analyzer to be created")
	}
}

func TestNewAnalyzerFromConfig(t *testing.T) {
	cfg := &config.Config{
		BitBucket: config.BitBucketConfig{
			ServerURL: "https://test.com",
			Username:  "testuser",
			Password:  "testpass",
		},
		Threads: 3,
	}

	logger := createTestLogger(t)
	defer logger.Close()

	analyzer := analyzer.NewAnalyzerFromConfig(cfg, logger)
	if analyzer == nil {
		t.Error("Expected analyzer to be created from config")
	}
}

func TestAnalyzer_RunWithSingleProject(t *testing.T) {
	cfg := &config.Config{
		BitBucket: config.BitBucketConfig{
			ServerURL: "https://test.com",
			Username:  "testuser",
			Password:  "testpass",
			Project:   "TEST",
		},
		Threads: 2,
	}

	logger := createTestLogger(t)
	defer logger.Close()

	// Create mock data
	project := bitbucket.Project{
		Key:  "TEST",
		ID:   1,
		Name: "Test Project",
		Type: "NORMAL",
	}

	repo1 := bitbucket.Repository{
		Slug:    "repo1",
		ID:      1,
		Name:    "Repository 1",
		Project: project,
	}

	repo2 := bitbucket.Repository{
		Slug:    "repo2",
		ID:      2,
		Name:    "Repository 2",
		Project: project,
	}

	size1 := &bitbucket.RepositorySize{
		Repository:  1024000,
		Attachments: 512000,
	}

	size2 := &bitbucket.RepositorySize{
		Repository:  2048000,
		Attachments: 256000,
	}

	pr1 := []bitbucket.PullRequest{
		{ID: 1, Properties: bitbucket.PullRequestProperties{CommentCount: 5}},
		{ID: 2, Properties: bitbucket.PullRequestProperties{CommentCount: 3}},
	}

	pr2 := []bitbucket.PullRequest{
		{ID: 3, Properties: bitbucket.PullRequestProperties{CommentCount: 10}},
	}

	// Set up mock client
	mockClient := NewMockBitBucketClient()
	mockClient.projects = []bitbucket.Project{project}
	mockClient.repositories["TEST"] = []bitbucket.Repository{repo1, repo2}
	mockClient.sizes["TEST/repo1"] = size1
	mockClient.sizes["TEST/repo2"] = size2
	mockClient.pullRequests["TEST/repo1"] = pr1
	mockClient.pullRequests["TEST/repo2"] = pr2

	analyzer := analyzer.NewAnalyzer(mockClient, logger, cfg)

	ctx := context.Background()
	results, err := analyzer.Run(ctx)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if results == nil {
		t.Fatal("Expected results to be returned")
	}

	// Verify results
	if len(results.Projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(results.Projects))
	}

	if results.Projects[0].Key != "TEST" {
		t.Errorf("Expected project key 'TEST', got %s", results.Projects[0].Key)
	}

	if len(results.Repositories) != 2 {
		t.Errorf("Expected 2 repositories, got %d", len(results.Repositories))
	}

	// Verify statistics
	if results.Statistics.TotalProjects != 1 {
		t.Errorf("Expected 1 total project, got %d", results.Statistics.TotalProjects)
	}

	if results.Statistics.TotalRepos != 2 {
		t.Errorf("Expected 2 total repos, got %d", results.Statistics.TotalRepos)
	}

	if results.Statistics.TotalPRs != 3 {
		t.Errorf("Expected 3 total PRs, got %d", results.Statistics.TotalPRs)
	}

	if results.Statistics.TotalComments != 18 { // 5+3+10
		t.Errorf("Expected 18 total comments, got %d", results.Statistics.TotalComments)
	}

	expectedSize := int64(3072000) // 1024000 + 2048000
	if results.Statistics.TotalSize != expectedSize {
		t.Errorf("Expected total size %d, got %d", expectedSize, results.Statistics.TotalSize)
	}

	// Verify new migration-relevant statistics
	if results.Statistics.TotalBranches != 4 { // 2 branches per repo * 2 repos
		t.Errorf("Expected 4 total branches, got %d", results.Statistics.TotalBranches)
	}

	if results.Statistics.TotalTags != 2 { // 1 tag per repo * 2 repos
		t.Errorf("Expected 2 total tags, got %d", results.Statistics.TotalTags)
	}

	if results.Statistics.TotalPermissions != 2 { // 1 permission per repo * 2 repos
		t.Errorf("Expected 2 total permissions, got %d", results.Statistics.TotalPermissions)
	}

	if results.Statistics.ActiveRepos != 2 { // Both repos have commits
		t.Errorf("Expected 2 active repos, got %d", results.Statistics.ActiveRepos)
	}

	// Verify repository details
	for _, repo := range results.Repositories {
		if repo.Size == nil {
			t.Errorf("Expected repository %s to have size information", repo.Name)
		}

		if len(repo.PullRequests) == 0 && repo.Slug == "repo1" {
			t.Errorf("Expected repository %s to have pull requests", repo.Name)
		}

		if repo.CommentCount == 0 && repo.Slug == "repo1" {
			t.Errorf("Expected repository %s to have comments", repo.Name)
		}

		// Verify new migration-relevant fields
		if repo.BranchCount == 0 {
			t.Errorf("Expected repository %s to have branch count", repo.Name)
		}

		if repo.DefaultBranch == "" {
			t.Errorf("Expected repository %s to have default branch", repo.Name)
		}

		if repo.TagCount == 0 {
			t.Errorf("Expected repository %s to have tag count", repo.Name)
		}

		if repo.CommitCount == 0 {
			t.Errorf("Expected repository %s to have commit count", repo.Name)
		}

		if repo.ContributorCount == 0 {
			t.Errorf("Expected repository %s to have contributor count", repo.Name)
		}
	}
}

func TestAnalyzer_RunWithAllProjects(t *testing.T) {
	cfg := &config.Config{
		BitBucket: config.BitBucketConfig{
			ServerURL: "https://test.com",
			Username:  "testuser",
			Password:  "testpass",
			// No specific project - should get all projects
		},
		Threads: 1,
	}

	logger := createTestLogger(t)
	defer logger.Close()

	project1 := bitbucket.Project{Key: "PROJ1", ID: 1, Name: "Project 1"}
	project2 := bitbucket.Project{Key: "PROJ2", ID: 2, Name: "Project 2"}

	repo1 := bitbucket.Repository{Slug: "repo1", ID: 1, Name: "Repository 1", Project: project1}
	repo2 := bitbucket.Repository{Slug: "repo2", ID: 2, Name: "Repository 2", Project: project2}

	mockClient := NewMockBitBucketClient()
	mockClient.projects = []bitbucket.Project{project1, project2}
	mockClient.repositories["PROJ1"] = []bitbucket.Repository{repo1}
	mockClient.repositories["PROJ2"] = []bitbucket.Repository{repo2}
	mockClient.sizes["PROJ1/repo1"] = &bitbucket.RepositorySize{Repository: 1000}
	mockClient.sizes["PROJ2/repo2"] = &bitbucket.RepositorySize{Repository: 2000}
	mockClient.pullRequests["PROJ1/repo1"] = []bitbucket.PullRequest{}
	mockClient.pullRequests["PROJ2/repo2"] = []bitbucket.PullRequest{}

	analyzer := analyzer.NewAnalyzer(mockClient, logger, cfg)

	ctx := context.Background()
	results, err := analyzer.Run(ctx)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results.Projects) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(results.Projects))
	}

	if len(results.Repositories) != 2 {
		t.Errorf("Expected 2 repositories, got %d", len(results.Repositories))
	}
}

func TestAnalyzer_RunWithCancellation(t *testing.T) {
	cfg := &config.Config{
		BitBucket: config.BitBucketConfig{
			ServerURL: "https://test.com",
			Username:  "testuser",
			Password:  "testpass",
		},
		Threads: 1,
	}

	logger := createTestLogger(t)
	defer logger.Close()

	// Create context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	project := bitbucket.Project{Key: "TEST", ID: 1, Name: "Test Project"}
	repo := bitbucket.Repository{Slug: "repo1", ID: 1, Name: "Repository 1", Project: project}

	mockClient := NewMockBitBucketClient()
	mockClient.projects = []bitbucket.Project{project}
	mockClient.repositories["TEST"] = []bitbucket.Repository{repo}

	analyzer := analyzer.NewAnalyzer(mockClient, logger, cfg)

	_, err := analyzer.Run(ctx)

	if err == nil {
		t.Error("Expected error due to cancelled context")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestAnalyzer_RunWithErrors(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*MockBitBucketClient)
		expectedError string
	}{
		{
			name: "project_error",
			setupMock: func(m *MockBitBucketClient) {
				m.projectError = errors.New("failed to get projects")
			},
			expectedError: "failed to get projects",
		},
		{
			name: "no_projects",
			setupMock: func(m *MockBitBucketClient) {
				m.projects = []bitbucket.Project{}
			},
			expectedError: "no projects found",
		},
		{
			name: "repository_error",
			setupMock: func(m *MockBitBucketClient) {
				m.projects = []bitbucket.Project{{Key: "TEST", ID: 1, Name: "Test"}}
				m.repoError = errors.New("failed to get repositories")
			},
			expectedError: "failed to get repositories for project TEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				BitBucket: config.BitBucketConfig{
					ServerURL: "https://test.com",
					Username:  "testuser",
					Password:  "testpass",
				},
				Threads: 1,
			}

			logger := createTestLogger(t)
			defer logger.Close()

			mockClient := NewMockBitBucketClient()
			tt.setupMock(mockClient)

			analyzer := analyzer.NewAnalyzer(mockClient, logger, cfg)

			ctx := context.Background()
			_, err := analyzer.Run(ctx)

			if err == nil {
				t.Error("Expected error")
			}

			if !strings.Contains(fmt.Sprintf("%v", err), tt.expectedError) {
				t.Errorf("Expected error containing '%s', got %v", tt.expectedError, err)
			}
		})
	}
}

func TestAnalyzer_SizeAndPRErrors(t *testing.T) {
	// Test that the analyzer continues working even when size and PR requests fail
	cfg := &config.Config{
		BitBucket: config.BitBucketConfig{
			ServerURL: "https://test.com",
			Username:  "testuser",
			Password:  "testpass",
		},
		Threads: 1,
	}

	logger := createTestLogger(t)
	defer logger.Close()

	project := bitbucket.Project{Key: "TEST", ID: 1, Name: "Test Project"}
	repo := bitbucket.Repository{Slug: "repo1", ID: 1, Name: "Repository 1", Project: project}

	mockClient := NewMockBitBucketClient()
	mockClient.projects = []bitbucket.Project{project}
	mockClient.repositories["TEST"] = []bitbucket.Repository{repo}
	mockClient.sizeError = errors.New("size API error")
	mockClient.prError = errors.New("PR API error")

	analyzer := analyzer.NewAnalyzer(mockClient, logger, cfg)

	ctx := context.Background()
	results, err := analyzer.Run(ctx)

	// Should not fail even though size and PR requests failed
	if err != nil {
		t.Fatalf("Expected no error despite API failures, got %v", err)
	}

	if len(results.Repositories) != 1 {
		t.Errorf("Expected 1 repository, got %d", len(results.Repositories))
	}

	repo = results.Repositories[0]
	if repo.Size != nil {
		t.Error("Expected repository size to be nil due to API error")
	}

	if len(repo.PullRequests) != 0 {
		t.Error("Expected no pull requests due to API error")
	}

	if repo.CommentCount != 0 {
		t.Error("Expected comment count to be 0 due to no pull requests")
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"bytes", 512, "512 B"},
		{"kilobytes", 1536, "1 KB"},       // 1.5 KB rounds down
		{"megabytes", 1572864, "1 MB"},    // 1.5 MB rounds down
		{"gigabytes", 1610612736, "1 GB"}, // 1.5 GB rounds down
		{"large_gigabytes", 5368709120, "5 GB"},
		{"zero", 0, "0 B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.FormatSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
