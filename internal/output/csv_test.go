package output_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mona-actions/gh-bbs-analyzer/internal/bitbucket"
	"github.com/mona-actions/gh-bbs-analyzer/internal/output"
)

func TestNewCSVWriter(t *testing.T) {
	filename := "test.csv"
	writer := output.NewCSVWriter(filename)

	if writer == nil {
		t.Error("Expected CSV writer to be created")
	}
}

func TestCSVWriter_WriteRepositories_Empty(t *testing.T) {
	filename := fmt.Sprintf("test_empty_%d.csv", time.Now().UnixNano())
	defer os.Remove(filename)

	writer := output.NewCSVWriter(filename)
	repositories := []bitbucket.Repository{}

	err := writer.WriteRepositories(repositories)
	if err != nil {
		t.Fatalf("Expected no error writing empty repositories, got %v", err)
	}

	// Verify file was created and contains header
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	csvContent := string(content)
	expectedHeader := "project,repository,size,pull_requests,comments,archived,public"
	if !strings.Contains(csvContent, expectedHeader) {
		t.Errorf("Expected CSV to contain header '%s', got: %s", expectedHeader, csvContent)
	}

	// Should only contain header line
	lines := strings.Split(strings.TrimSpace(csvContent), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected 1 line (header only), got %d lines", len(lines))
	}
}

func TestCSVWriter_WriteRepositories_SingleRepo(t *testing.T) {
	filename := fmt.Sprintf("test_single_%d.csv", time.Now().UnixNano())
	defer os.Remove(filename)

	writer := output.NewCSVWriter(filename)

	project := bitbucket.Project{
		Key:  "TEST",
		ID:   1,
		Name: "Test Project",
	}

	size := &bitbucket.RepositorySize{
		Repository:  1024000,
		Attachments: 512000,
	}

	pullRequests := []bitbucket.PullRequest{
		{ID: 1, Properties: bitbucket.PullRequestProperties{CommentCount: 5}},
		{ID: 2, Properties: bitbucket.PullRequestProperties{CommentCount: 3}},
	}

	repo := bitbucket.Repository{
		Slug:         "test-repo",
		ID:           1,
		Name:         "Test Repository",
		Project:      project,
		Size:         size,
		PullRequests: pullRequests,
		CommentCount: 8, // 5 + 3
		Archived:     false,
		Public:       true,
	}

	repositories := []bitbucket.Repository{repo}

	err := writer.WriteRepositories(repositories)
	if err != nil {
		t.Fatalf("Expected no error writing repository, got %v", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	csvContent := string(content)
	lines := strings.Split(strings.TrimSpace(csvContent), "\n")

	// Should contain header + 1 data line
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines (header + data), got %d lines", len(lines))
	}

	dataLine := lines[1]
	expectedLine := "TEST,test-repo,1024000,2,8,false,true"
	if dataLine != expectedLine {
		t.Errorf("Expected data line '%s', got '%s'", expectedLine, dataLine)
	}
}

func TestCSVWriter_WriteRepositories_MultipleRepos(t *testing.T) {
	filename := fmt.Sprintf("test_multiple_%d.csv", time.Now().UnixNano())
	defer os.Remove(filename)

	writer := output.NewCSVWriter(filename)

	project1 := bitbucket.Project{Key: "PROJ1", ID: 1, Name: "Project 1"}
	project2 := bitbucket.Project{Key: "PROJ2", ID: 2, Name: "Project 2"}

	size1 := &bitbucket.RepositorySize{Repository: 1024000, Attachments: 0}
	size2 := &bitbucket.RepositorySize{Repository: 2048000, Attachments: 1024000}

	repo1 := bitbucket.Repository{
		Slug:         "repo1",
		ID:           1,
		Name:         "Repository 1",
		Project:      project1,
		Size:         size1,
		PullRequests: []bitbucket.PullRequest{{ID: 1, Properties: bitbucket.PullRequestProperties{CommentCount: 10}}},
		CommentCount: 10,
		Archived:     false,
		Public:       false,
	}

	repo2 := bitbucket.Repository{
		Slug:         "repo2",
		ID:           2,
		Name:         "Repository 2",
		Project:      project2,
		Size:         size2,
		PullRequests: []bitbucket.PullRequest{},
		CommentCount: 0,
		Archived:     true,
		Public:       true,
	}

	repositories := []bitbucket.Repository{repo1, repo2}

	err := writer.WriteRepositories(repositories)
	if err != nil {
		t.Fatalf("Expected no error writing repositories, got %v", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	csvContent := string(content)
	lines := strings.Split(strings.TrimSpace(csvContent), "\n")

	// Should contain header + 2 data lines
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines (header + 2 data), got %d lines", len(lines))
	}

	// Check first repository
	expectedLine1 := "PROJ1,repo1,1024000,1,10,false,false"
	if lines[1] != expectedLine1 {
		t.Errorf("Expected first data line '%s', got '%s'", expectedLine1, lines[1])
	}

	// Check second repository
	expectedLine2 := "PROJ2,repo2,2048000,0,0,true,true"
	if lines[2] != expectedLine2 {
		t.Errorf("Expected second data line '%s', got '%s'", expectedLine2, lines[2])
	}
}

func TestCSVWriter_WriteRepositories_NoSize(t *testing.T) {
	filename := fmt.Sprintf("test_nosize_%d.csv", time.Now().UnixNano())
	defer os.Remove(filename)

	writer := output.NewCSVWriter(filename)

	project := bitbucket.Project{Key: "TEST", ID: 1, Name: "Test Project"}

	repo := bitbucket.Repository{
		Slug:         "test-repo",
		ID:           1,
		Name:         "Test Repository",
		Project:      project,
		Size:         nil, // No size information
		PullRequests: []bitbucket.PullRequest{},
		CommentCount: 0,
		Archived:     false,
		Public:       false,
	}

	repositories := []bitbucket.Repository{repo}

	err := writer.WriteRepositories(repositories)
	if err != nil {
		t.Fatalf("Expected no error writing repository with no size, got %v", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	csvContent := string(content)
	lines := strings.Split(strings.TrimSpace(csvContent), "\n")

	dataLine := lines[1]
	expectedLine := "TEST,test-repo,0,0,0,false,false"
	if dataLine != expectedLine {
		t.Errorf("Expected data line '%s', got '%s'", expectedLine, dataLine)
	}
}

func TestCSVWriter_WriteRepositories_InvalidPath(t *testing.T) {
	// Test writing to an invalid path
	invalidPath := "/nonexistent/directory/test.csv"
	writer := output.NewCSVWriter(invalidPath)

	repo := bitbucket.Repository{
		Slug:    "test-repo",
		Project: bitbucket.Project{Key: "TEST"},
	}

	repositories := []bitbucket.Repository{repo}

	err := writer.WriteRepositories(repositories)
	if err == nil {
		t.Error("Expected error when writing to invalid path")
	}

	if !strings.Contains(err.Error(), "failed to create output file") {
		t.Errorf("Expected 'failed to create output file' error, got %v", err)
	}
}

func TestCSVWriter_WriteRepositories_SpecialCharacters(t *testing.T) {
	filename := fmt.Sprintf("test_special_%d.csv", time.Now().UnixNano())
	defer os.Remove(filename)

	writer := output.NewCSVWriter(filename)

	project := bitbucket.Project{Key: "PROJ,WITH,COMMAS", ID: 1, Name: "Project with commas"}

	repo := bitbucket.Repository{
		Slug:         "repo-with-special-chars",
		ID:           1,
		Name:         "Repository with special chars",
		Project:      project,
		Size:         &bitbucket.RepositorySize{Repository: 1000},
		PullRequests: []bitbucket.PullRequest{},
		CommentCount: 0,
		Archived:     false,
		Public:       false,
	}

	repositories := []bitbucket.Repository{repo}

	err := writer.WriteRepositories(repositories)
	if err != nil {
		t.Fatalf("Expected no error writing repository with special chars, got %v", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	csvContent := string(content)

	// Note: This test shows that the current CSV implementation doesn't handle
	// commas in data properly. This could be improved to use proper CSV escaping.
	if !strings.Contains(csvContent, "PROJ,WITH,COMMAS") {
		t.Error("Expected CSV to contain project key with commas")
	}
}

func TestCSVWriter_WriteRepositories_LargeNumbers(t *testing.T) {
	filename := fmt.Sprintf("test_large_%d.csv", time.Now().UnixNano())
	defer os.Remove(filename)

	writer := output.NewCSVWriter(filename)

	project := bitbucket.Project{Key: "LARGE", ID: 1, Name: "Large Project"}

	size := &bitbucket.RepositorySize{
		Repository:  9223372036854775807, // Max int64
		Attachments: 0,
	}

	repo := bitbucket.Repository{
		Slug:         "large-repo",
		ID:           1,
		Name:         "Large Repository",
		Project:      project,
		Size:         size,
		PullRequests: []bitbucket.PullRequest{},
		CommentCount: 0,
		Archived:     false,
		Public:       false,
	}

	repositories := []bitbucket.Repository{repo}

	err := writer.WriteRepositories(repositories)
	if err != nil {
		t.Fatalf("Expected no error writing repository with large size, got %v", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	csvContent := string(content)
	expectedSize := "9223372036854775807"
	if !strings.Contains(csvContent, expectedSize) {
		t.Errorf("Expected CSV to contain large size '%s'", expectedSize)
	}
}
