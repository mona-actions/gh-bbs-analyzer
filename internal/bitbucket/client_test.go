package bitbucket_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mona-actions/gh-bbs-analyzer/internal/bitbucket"
)

func TestNewClient(t *testing.T) {
	client := bitbucket.NewClient("https://test.com", "user", "pass", false)
	if client == nil {
		t.Error("Expected client to be created")
	}
}

func TestNewClientWithSSLDisabled(t *testing.T) {
	client := bitbucket.NewClient("https://test.com", "user", "pass", true)
	if client == nil {
		t.Error("Expected client to be created with SSL verification disabled")
	}
}

func TestClient_GetProject(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/1.0/projects/TEST" {
			t.Errorf("Expected path '/rest/api/1.0/projects/TEST', got %s", r.URL.Path)
		}

		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		// Check basic auth
		username, password, ok := r.BasicAuth()
		if !ok || username != "testuser" || password != "testpass" {
			t.Error("Expected basic auth with testuser:testpass")
		}

		response := `{
			"key": "TEST",
			"id": 1,
			"name": "Test Project",
			"public": false,
			"type": "NORMAL"
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := bitbucket.NewClient(server.URL, "testuser", "testpass", false)
	project, err := client.GetProject("TEST")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if project.Key != "TEST" {
		t.Errorf("Expected project key 'TEST', got %s", project.Key)
	}

	if project.Name != "Test Project" {
		t.Errorf("Expected project name 'Test Project', got %s", project.Name)
	}

	if project.ID != 1 {
		t.Errorf("Expected project ID 1, got %d", project.ID)
	}
}

func TestClient_GetProjectNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := bitbucket.NewClient(server.URL, "testuser", "testpass", false)
	_, err := client.GetProject("NOTFOUND")

	if err == nil {
		t.Error("Expected error for 404 response")
	}

	if err.Error() != "API request failed with status 404" {
		t.Errorf("Expected 404 error message, got %v", err)
	}
}

func TestClient_GetProjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/1.0/projects" {
			t.Errorf("Expected path '/rest/api/1.0/projects', got %s", r.URL.Path)
		}

		response := `{
			"size": 2,
			"limit": 100,
			"isLastPage": true,
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
					"type": "NORMAL"
				}
			],
			"start": 0
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := bitbucket.NewClient(server.URL, "testuser", "testpass", false)
	projects, err := client.GetProjects()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(projects) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(projects))
	}

	if projects[0].Key != "PROJ1" {
		t.Errorf("Expected first project key 'PROJ1', got %s", projects[0].Key)
	}

	if projects[1].Key != "PROJ2" {
		t.Errorf("Expected second project key 'PROJ2', got %s", projects[1].Key)
	}
}

func TestClient_GetProjectsPaginated(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		var response string
		if requestCount == 1 {
			// First page
			response = `{
				"size": 1,
				"limit": 1,
				"isLastPage": false,
				"nextPageStart": 1,
				"values": [
					{
						"key": "PROJ1",
						"id": 1,
						"name": "Project 1",
						"public": false,
						"type": "NORMAL"
					}
				],
				"start": 0
			}`
		} else {
			// Second page
			response = `{
				"size": 1,
				"limit": 1,
				"isLastPage": true,
				"values": [
					{
						"key": "PROJ2",
						"id": 2,
						"name": "Project 2",
						"public": true,
						"type": "NORMAL"
					}
				],
				"start": 1
			}`
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := bitbucket.NewClient(server.URL, "testuser", "testpass", false)
	projects, err := client.GetProjects()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(projects) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(projects))
	}

	if requestCount != 2 {
		t.Errorf("Expected 2 API requests for pagination, got %d", requestCount)
	}
}

func TestClient_GetRepositories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/rest/api/1.0/projects/TEST/repos"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got %s", expectedPath, r.URL.Path)
		}

		response := `{
			"size": 1,
			"limit": 100,
			"isLastPage": true,
			"values": [
				{
					"slug": "test-repo",
					"id": 1,
					"name": "Test Repository",
					"hierarchyId": "1",
					"scmId": "git",
					"state": "AVAILABLE",
					"statusMessage": "Available",
					"forkable": true,
					"public": false,
					"archived": false,
					"project": {
						"key": "TEST",
						"id": 1,
						"name": "Test Project",
						"public": false,
						"type": "NORMAL"
					}
				}
			],
			"start": 0
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := bitbucket.NewClient(server.URL, "testuser", "testpass", false)
	repos, err := client.GetRepositories("TEST")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(repos) != 1 {
		t.Errorf("Expected 1 repository, got %d", len(repos))
	}

	repo := repos[0]
	if repo.Slug != "test-repo" {
		t.Errorf("Expected repository slug 'test-repo', got %s", repo.Slug)
	}

	if repo.Name != "Test Repository" {
		t.Errorf("Expected repository name 'Test Repository', got %s", repo.Name)
	}

	if repo.Project.Key != "TEST" {
		t.Errorf("Expected project key 'TEST', got %s", repo.Project.Key)
	}
}

func TestClient_GetRepositorySize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/projects/TEST/repos/test-repo/sizes"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got %s", expectedPath, r.URL.Path)
		}

		response := `{
			"repository": 1024000,
			"attachments": 512000
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := bitbucket.NewClient(server.URL, "testuser", "testpass", false)
	size, err := client.GetRepositorySize("TEST", "test-repo")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if size.Repository != 1024000 {
		t.Errorf("Expected repository size 1024000, got %d", size.Repository)
	}

	if size.Attachments != 512000 {
		t.Errorf("Expected attachments size 512000, got %d", size.Attachments)
	}
}

func TestClient_GetPullRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/rest/api/1.0/projects/TEST/repos/test-repo/pull-requests"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got %s", expectedPath, r.URL.Path)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("state") != "all" {
			t.Errorf("Expected state=all query param, got %s", query.Get("state"))
		}

		response := `{
			"size": 2,
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
						"commentCount": 3
					}
				}
			],
			"start": 0
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := bitbucket.NewClient(server.URL, "testuser", "testpass", false)
	pullRequests, err := client.GetPullRequests("TEST", "test-repo")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(pullRequests) != 2 {
		t.Errorf("Expected 2 pull requests, got %d", len(pullRequests))
	}

	if pullRequests[0].ID != 1 {
		t.Errorf("Expected first PR ID 1, got %d", pullRequests[0].ID)
	}

	if pullRequests[0].Properties.CommentCount != 5 {
		t.Errorf("Expected first PR comment count 5, got %d", pullRequests[0].Properties.CommentCount)
	}

	if pullRequests[1].Properties.CommentCount != 3 {
		t.Errorf("Expected second PR comment count 3, got %d", pullRequests[1].Properties.CommentCount)
	}
}

func TestClient_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := bitbucket.NewClient(server.URL, "testuser", "testpass", false)
	_, err := client.GetProject("TEST")

	if err == nil {
		t.Error("Expected error for invalid JSON response")
	}

	if !strings.Contains(fmt.Sprintf("%v", err), "unmarshal") {
		t.Errorf("Expected unmarshal error, got %v", err)
	}
}

func TestClient_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := bitbucket.NewClient(server.URL, "baduser", "badpass", false)
	_, err := client.GetProject("TEST")

	if err == nil {
		t.Error("Expected error for unauthorized access")
	}

	if err.Error() != "API request failed with status 401" {
		t.Errorf("Expected 401 error message, got %v", err)
	}
}
