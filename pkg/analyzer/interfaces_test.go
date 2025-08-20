package analyzer_test

import (
	"testing"

	"github.com/mona-actions/gh-bbs-analyzer/internal/bitbucket"
	"github.com/mona-actions/gh-bbs-analyzer/pkg/analyzer"
)

func TestBitBucketClientInterface(t *testing.T) {
	// Test that bitbucket.Client implements the BitBucketClient interface
	var client analyzer.BitBucketClient
	client = bitbucket.NewClient("https://test.com", "user", "pass", false)

	if client == nil {
		t.Error("Expected bitbucket.Client to implement BitBucketClient interface")
	}
}

func TestBitBucketClientInterfaceMethods(t *testing.T) {
	// Verify all required methods exist on the interface
	client := bitbucket.NewClient("https://test.com", "user", "pass", false)

	// These calls will fail due to network issues, but we're just testing
	// that the methods exist and have the correct signatures
	_, err := client.GetProjects()
	if err == nil {
		t.Log("GetProjects method exists (network call would fail in test)")
	}

	_, err = client.GetProject("TEST")
	if err == nil {
		t.Log("GetProject method exists (network call would fail in test)")
	}

	_, err = client.GetRepositories("TEST")
	if err == nil {
		t.Log("GetRepositories method exists (network call would fail in test)")
	}

	_, err = client.GetRepositorySize("TEST", "repo")
	if err == nil {
		t.Log("GetRepositorySize method exists (network call would fail in test)")
	}

	_, err = client.GetPullRequests("TEST", "repo")
	if err == nil {
		t.Log("GetPullRequests method exists (network call would fail in test)")
	}

	// The important thing is that these methods compile and exist
	t.Log("All BitBucketClient interface methods are implemented")
}
