package bitbucket_test

import (
	"encoding/json"
	"testing"

	"github.com/mona-actions/gh-bbs-analyzer/internal/bitbucket"
)

func TestProject_JSONSerialization(t *testing.T) {
	jsonData := `{
		"key": "TEST",
		"id": 123,
		"name": "Test Project",
		"public": true,
		"type": "NORMAL"
	}`

	var project bitbucket.Project
	err := json.Unmarshal([]byte(jsonData), &project)
	if err != nil {
		t.Fatalf("Failed to unmarshal project JSON: %v", err)
	}

	if project.Key != "TEST" {
		t.Errorf("Expected key 'TEST', got %s", project.Key)
	}

	if project.ID != 123 {
		t.Errorf("Expected ID 123, got %d", project.ID)
	}

	if project.Name != "Test Project" {
		t.Errorf("Expected name 'Test Project', got %s", project.Name)
	}

	if !project.Public {
		t.Error("Expected public to be true")
	}

	if project.Type != "NORMAL" {
		t.Errorf("Expected type 'NORMAL', got %s", project.Type)
	}
}

func TestProjectResponse_JSONSerialization(t *testing.T) {
	jsonData := `{
		"size": 2,
		"limit": 25,
		"isLastPage": false,
		"values": [
			{
				"key": "PROJ1",
				"id": 1,
				"name": "Project 1",
				"public": false,
				"type": "NORMAL"
			},
			{
				"key": "PROJ2",
				"id": 2,
				"name": "Project 2",
				"public": true,
				"type": "PERSONAL"
			}
		],
		"start": 0,
		"filter": "",
		"nextPageStart": 25
	}`

	var response bitbucket.ProjectResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal project response JSON: %v", err)
	}

	if response.Size != 2 {
		t.Errorf("Expected size 2, got %d", response.Size)
	}

	if response.Limit != 25 {
		t.Errorf("Expected limit 25, got %d", response.Limit)
	}

	if response.IsLastPage {
		t.Error("Expected isLastPage to be false")
	}

	if len(response.Values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(response.Values))
	}

	if response.Values[0].Key != "PROJ1" {
		t.Errorf("Expected first project key 'PROJ1', got %s", response.Values[0].Key)
	}

	if response.NextPageStart != 25 {
		t.Errorf("Expected nextPageStart 25, got %d", response.NextPageStart)
	}
}

func TestRepository_JSONSerialization(t *testing.T) {
	jsonData := `{
		"slug": "test-repo",
		"id": 456,
		"name": "Test Repository",
		"hierarchyId": "abc123",
		"scmId": "git",
		"state": "AVAILABLE",
		"statusMessage": "Available",
		"forkable": true,
		"public": false,
		"archived": true,
		"project": {
			"key": "TEST",
			"id": 123,
			"name": "Test Project",
			"public": false,
			"type": "NORMAL"
		}
	}`

	var repo bitbucket.Repository
	err := json.Unmarshal([]byte(jsonData), &repo)
	if err != nil {
		t.Fatalf("Failed to unmarshal repository JSON: %v", err)
	}

	if repo.Slug != "test-repo" {
		t.Errorf("Expected slug 'test-repo', got %s", repo.Slug)
	}

	if repo.ID != 456 {
		t.Errorf("Expected ID 456, got %d", repo.ID)
	}

	if repo.Name != "Test Repository" {
		t.Errorf("Expected name 'Test Repository', got %s", repo.Name)
	}

	if repo.HierarchyId != "abc123" {
		t.Errorf("Expected hierarchyId 'abc123', got %s", repo.HierarchyId)
	}

	if repo.ScmId != "git" {
		t.Errorf("Expected scmId 'git', got %s", repo.ScmId)
	}

	if repo.State != "AVAILABLE" {
		t.Errorf("Expected state 'AVAILABLE', got %s", repo.State)
	}

	if !repo.Forkable {
		t.Error("Expected forkable to be true")
	}

	if repo.Public {
		t.Error("Expected public to be false")
	}

	if !repo.Archived {
		t.Error("Expected archived to be true")
	}

	if repo.Project.Key != "TEST" {
		t.Errorf("Expected project key 'TEST', got %s", repo.Project.Key)
	}
}

func TestRepositorySize_JSONSerialization(t *testing.T) {
	jsonData := `{
		"repository": 1024000,
		"attachments": 512000
	}`

	var size bitbucket.RepositorySize
	err := json.Unmarshal([]byte(jsonData), &size)
	if err != nil {
		t.Fatalf("Failed to unmarshal repository size JSON: %v", err)
	}

	if size.Repository != 1024000 {
		t.Errorf("Expected repository size 1024000, got %d", size.Repository)
	}

	if size.Attachments != 512000 {
		t.Errorf("Expected attachments size 512000, got %d", size.Attachments)
	}
}

func TestPullRequest_JSONSerialization(t *testing.T) {
	jsonData := `{
		"id": 789,
		"properties": {
			"commentCount": 15
		}
	}`

	var pr bitbucket.PullRequest
	err := json.Unmarshal([]byte(jsonData), &pr)
	if err != nil {
		t.Fatalf("Failed to unmarshal pull request JSON: %v", err)
	}

	if pr.ID != 789 {
		t.Errorf("Expected ID 789, got %d", pr.ID)
	}

	if pr.Properties.CommentCount != 15 {
		t.Errorf("Expected comment count 15, got %d", pr.Properties.CommentCount)
	}
}

func TestPullRequestResponse_JSONSerialization(t *testing.T) {
	jsonData := `{
		"size": 3,
		"limit": 100,
		"isLastPage": true,
		"values": [
			{
				"id": 1,
				"properties": {
					"commentCount": 5
				}
			},
			{
				"id": 2,
				"properties": {
					"commentCount": 8
				}
			},
			{
				"id": 3,
				"properties": {
					"commentCount": 0
				}
			}
		],
		"start": 0,
		"filter": "state=OPEN",
		"nextPageStart": 0
	}`

	var response bitbucket.PullRequestResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal pull request response JSON: %v", err)
	}

	if response.Size != 3 {
		t.Errorf("Expected size 3, got %d", response.Size)
	}

	if !response.IsLastPage {
		t.Error("Expected isLastPage to be true")
	}

	if len(response.Values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(response.Values))
	}

	if response.Values[1].Properties.CommentCount != 8 {
		t.Errorf("Expected second PR comment count 8, got %d", response.Values[1].Properties.CommentCount)
	}

	if response.Filter != "state=OPEN" {
		t.Errorf("Expected filter 'state=OPEN', got %s", response.Filter)
	}
}

func TestRepository_DefaultValues(t *testing.T) {
	// Test that non-API fields have proper zero values
	repo := bitbucket.Repository{}

	if repo.Size != nil {
		t.Error("Expected Size to be nil by default")
	}

	if repo.PullRequests != nil {
		t.Error("Expected PullRequests to be nil by default")
	}

	if repo.CommentCount != 0 {
		t.Errorf("Expected CommentCount to be 0 by default, got %d", repo.CommentCount)
	}
}
