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
