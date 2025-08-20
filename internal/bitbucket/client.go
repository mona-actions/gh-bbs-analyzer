package bitbucket

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client represents a BitBucket Server client
type Client struct {
	serverURL   string
	username    string
	password    string
	noSSLVerify bool
	httpClient  *http.Client
}

// NewClient creates a new BitBucket Server client
func NewClient(serverURL, username, password string, noSSLVerify bool) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: noSSLVerify},
	}

	return &Client{
		serverURL:   serverURL,
		username:    username,
		password:    password,
		noSSLVerify: noSSLVerify,
		httpClient:  &http.Client{Transport: tr},
	}
}

// GetProject retrieves a specific project by key
func (c *Client) GetProject(projectKey string) (*Project, error) {
	endpoint := fmt.Sprintf("/rest/api/1.0/projects/%s", projectKey)
	data, err := c.makeRequest(endpoint, "GET")
	if err != nil {
		return nil, err
	}

	var project Project
	if err := json.Unmarshal([]byte(data), &project); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project response: %w", err)
	}

	return &project, nil
}

// GetProjects retrieves all projects with pagination
func (c *Client) GetProjects() ([]Project, error) {
	var allProjects []Project
	start := 0
	limit := 100

	for {
		endpoint := fmt.Sprintf("/rest/api/1.0/projects?limit=%d&start=%d", limit, start)
		data, err := c.makeRequest(endpoint, "GET")
		if err != nil {
			return nil, err
		}

		var response ProjectResponse
		if err := json.Unmarshal([]byte(data), &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal projects response: %w", err)
		}

		allProjects = append(allProjects, response.Values...)

		if response.IsLastPage {
			break
		}
		start = response.NextPageStart
	}

	return allProjects, nil
}

// GetRepositories retrieves all repositories for a project with pagination
func (c *Client) GetRepositories(projectKey string) ([]Repository, error) {
	var allRepos []Repository
	start := 0
	limit := 100

	for {
		endpoint := fmt.Sprintf("/rest/api/1.0/projects/%s/repos?limit=%d&start=%d", projectKey, limit, start)
		data, err := c.makeRequest(endpoint, "GET")
		if err != nil {
			return nil, err
		}

		var response RepositoryResponse
		if err := json.Unmarshal([]byte(data), &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal repositories response: %w", err)
		}

		allRepos = append(allRepos, response.Values...)

		if response.IsLastPage {
			break
		}
		start = response.NextPageStart
	}

	return allRepos, nil
}

// GetRepositorySize retrieves the size information for a repository
func (c *Client) GetRepositorySize(projectKey, repoSlug string) (*RepositorySize, error) {
	endpoint := fmt.Sprintf("/projects/%s/repos/%s/sizes", projectKey, repoSlug)
	data, err := c.makeRequest(endpoint, "GET")
	if err != nil {
		return nil, err
	}

	var size RepositorySize
	if err := json.Unmarshal([]byte(data), &size); err != nil {
		return nil, fmt.Errorf("failed to unmarshal repository size response: %w", err)
	}

	return &size, nil
}

// GetPullRequests retrieves all pull requests for a repository with pagination
func (c *Client) GetPullRequests(projectKey, repoSlug string) ([]PullRequest, error) {
	var allPRs []PullRequest
	start := 0
	limit := 100

	for {
		endpoint := fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/pull-requests?state=all&limit=%d&start=%d",
			projectKey, repoSlug, limit, start)
		data, err := c.makeRequest(endpoint, "GET")
		if err != nil {
			return nil, err
		}

		var response PullRequestResponse
		if err := json.Unmarshal([]byte(data), &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal pull requests response: %w", err)
		}

		allPRs = append(allPRs, response.Values...)

		if response.IsLastPage {
			break
		}
		start = response.NextPageStart
	}

	return allPRs, nil
}

// GetBranches retrieves all branches for a repository
func (c *Client) GetBranches(projectKey, repoSlug string) ([]Branch, error) {
	var allBranches []Branch
	start := 0
	limit := 100

	for {
		endpoint := fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/branches?limit=%d&start=%d",
			projectKey, repoSlug, limit, start)
		data, err := c.makeRequest(endpoint, "GET")
		if err != nil {
			return nil, err
		}

		var response BranchResponse
		if err := json.Unmarshal([]byte(data), &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal branches response: %w", err)
		}

		allBranches = append(allBranches, response.Values...)

		if response.IsLastPage {
			break
		}
		start = response.NextPageStart
	}

	return allBranches, nil
}

// GetTags retrieves all tags for a repository
func (c *Client) GetTags(projectKey, repoSlug string) ([]Tag, error) {
	var allTags []Tag
	start := 0
	limit := 100

	for {
		endpoint := fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/tags?limit=%d&start=%d",
			projectKey, repoSlug, limit, start)
		data, err := c.makeRequest(endpoint, "GET")
		if err != nil {
			return nil, err
		}

		var response TagResponse
		if err := json.Unmarshal([]byte(data), &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags response: %w", err)
		}

		allTags = append(allTags, response.Values...)

		if response.IsLastPage {
			break
		}
		start = response.NextPageStart
	}

	return allTags, nil
}

// GetUserPermissions retrieves user permissions for a repository
func (c *Client) GetUserPermissions(projectKey, repoSlug string) ([]UserPermission, error) {
	var allPermissions []UserPermission
	start := 0
	limit := 100

	for {
		endpoint := fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/permissions/users?limit=%d&start=%d",
			projectKey, repoSlug, limit, start)
		data, err := c.makeRequest(endpoint, "GET")
		if err != nil {
			// Permissions might not be accessible, return empty slice
			return allPermissions, nil
		}

		var response UserPermissionResponse
		if err := json.Unmarshal([]byte(data), &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user permissions response: %w", err)
		}

		allPermissions = append(allPermissions, response.Values...)

		if response.IsLastPage {
			break
		}
		start = response.NextPageStart
	}

	return allPermissions, nil
}

// GetHooks retrieves repository hooks
func (c *Client) GetHooks(projectKey, repoSlug string) ([]Hook, error) {
	endpoint := fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/settings/hooks", projectKey, repoSlug)
	data, err := c.makeRequest(endpoint, "GET")
	if err != nil {
		// Hooks might not be accessible, return empty slice
		return []Hook{}, nil
	}

	var response HookResponse
	if err := json.Unmarshal([]byte(data), &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hooks response: %w", err)
	}

	return response.Values, nil
}

// GetWebhooks retrieves repository webhooks
func (c *Client) GetWebhooks(projectKey, repoSlug string) ([]Webhook, error) {
	var allWebhooks []Webhook
	start := 0
	limit := 100

	for {
		endpoint := fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/webhooks?limit=%d&start=%d",
			projectKey, repoSlug, limit, start)
		data, err := c.makeRequest(endpoint, "GET")
		if err != nil {
			// Webhooks might not be accessible, return empty slice
			return allWebhooks, nil
		}

		var response WebhookResponse
		if err := json.Unmarshal([]byte(data), &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal webhooks response: %w", err)
		}

		allWebhooks = append(allWebhooks, response.Values...)

		if response.IsLastPage {
			break
		}
		start = response.NextPageStart
	}

	return allWebhooks, nil
}

// GetLFSFiles retrieves Git LFS files for a repository
func (c *Client) GetLFSFiles(projectKey, repoSlug string) ([]LFSFile, error) {
	var allLFSFiles []LFSFile
	start := 0
	limit := 100

	for {
		endpoint := fmt.Sprintf("/rest/git-lfs/1.0/projects/%s/repos/%s/files?limit=%d&start=%d",
			projectKey, repoSlug, limit, start)
		data, err := c.makeRequest(endpoint, "GET")
		if err != nil {
			// LFS might not be enabled, return empty slice
			return allLFSFiles, nil
		}

		var response LFSResponse
		if err := json.Unmarshal([]byte(data), &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal LFS files response: %w", err)
		}

		allLFSFiles = append(allLFSFiles, response.Values...)

		if response.IsLastPage {
			break
		}
		start = response.NextPageStart
	}

	return allLFSFiles, nil
}

// GetCommits retrieves recent commits for a repository to determine activity
func (c *Client) GetCommits(projectKey, repoSlug string, limit int) ([]Commit, error) {
	endpoint := fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/commits?limit=%d",
		projectKey, repoSlug, limit)
	data, err := c.makeRequest(endpoint, "GET")
	if err != nil {
		return nil, err
	}

	var response CommitResponse
	if err := json.Unmarshal([]byte(data), &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal commits response: %w", err)
	}

	return response.Values, nil
}

// GetForks retrieves forks of a repository
func (c *Client) GetForks(projectKey, repoSlug string) ([]Fork, error) {
	var allForks []Fork
	start := 0
	limit := 100

	for {
		endpoint := fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/forks?limit=%d&start=%d",
			projectKey, repoSlug, limit, start)
		data, err := c.makeRequest(endpoint, "GET")
		if err != nil {
			// Forks might not be accessible, return empty slice
			return allForks, nil
		}

		var response ForkResponse
		if err := json.Unmarshal([]byte(data), &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal forks response: %w", err)
		}

		allForks = append(allForks, response.Values...)

		if response.IsLastPage {
			break
		}
		start = response.NextPageStart
	}

	return allForks, nil
}

// makeRequest makes an HTTP request to the BitBucket Server API
func (c *Client) makeRequest(endpoint, method string) (string, error) {
	url := fmt.Sprintf("%s%s", c.serverURL, endpoint)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(bodyBytes), nil
}
