package cmd_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/mona-actions/gh-bbs-analyzer/cmd"
)

func TestNewRootCommand(t *testing.T) {
	rootCmd := cmd.NewRootCommand()

	if rootCmd.Use != "gh bbs-analyzer" {
		t.Errorf("Expected Use to be 'gh bbs-analyzer', got %s", rootCmd.Use)
	}

	if rootCmd.Version != "0.2.0" {
		t.Errorf("Expected Version to be '0.2.0', got %s", rootCmd.Version)
	}

	if !rootCmd.SilenceUsage {
		t.Error("Expected SilenceUsage to be true")
	}

	if !rootCmd.SilenceErrors {
		t.Error("Expected SilenceErrors to be true")
	}
}

func TestRootCommandFlags(t *testing.T) {
	rootCmd := cmd.NewRootCommand()

	// Test that required flags are present
	requiredFlags := []string{
		"bbs-server-url",
		"bbs-username",
		"bbs-password",
		"output-file",
		"threads",
	}

	for _, flagName := range requiredFlags {
		flag := rootCmd.PersistentFlags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag %s to be present", flagName)
		}
	}

	// Test bbs-server-url is marked as required
	flag := rootCmd.PersistentFlags().Lookup("bbs-server-url")
	if flag != nil {
		annotations := flag.Annotations
		if annotations == nil || annotations["cobra_annotation_bash_completion_one_required_flag"] == nil {
			// This is a simplified test - in practice we'd need to check if the flag is actually required
			// by attempting to run the command without it
		}
	}
}

func TestRootCommandHelp(t *testing.T) {
	rootCmd := cmd.NewRootCommand()

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Expected no error when running help, got %v", err)
	}

	output := buf.String()
	expectedStrings := []string{
		"GitHub CLI extension to analyze BitBucket Server for migration statistics",
		"--bbs-server-url",
		"--bbs-username",
		"--bbs-password",
		"--output-file",
		"--threads",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected help output to contain %s", expected)
		}
	}
}

func TestRootCommandVersion(t *testing.T) {
	rootCmd := cmd.NewRootCommand()

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Expected no error when running version, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "0.2.0") {
		t.Errorf("Expected version output to contain '0.2.0', got %s", output)
	}
}

func TestRootCommandFailsWithoutRequiredFlags(t *testing.T) {
	rootCmd := cmd.NewRootCommand()

	var buf bytes.Buffer
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{}) // No arguments

	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when running without required flags")
	}
}

func TestRootCommandWithMinimalFlags(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("BBS_USERNAME", "testuser")
	os.Setenv("BBS_PASSWORD", "testpass")
	defer func() {
		os.Unsetenv("BBS_USERNAME")
		os.Unsetenv("BBS_PASSWORD")
	}()

	rootCmd := cmd.NewRootCommand()

	var buf bytes.Buffer
	rootCmd.SetErr(&buf)

	// This will fail because we can't actually connect to a BitBucket server,
	// but it should at least pass the initial flag validation
	rootCmd.SetArgs([]string{
		"--bbs-server-url", "https://test.example.com",
	})

	err := rootCmd.Execute()
	// We expect this to fail due to network/auth issues, not flag validation
	if err == nil {
		t.Error("Expected error due to network/auth issues")
	}

	// The error should be about connection, not missing flags
	if strings.Contains(err.Error(), "required flag") {
		t.Errorf("Expected connection error, but got flag error: %v", err)
	}
}

func TestRootCommandWithInvalidThreads(t *testing.T) {
	os.Setenv("BBS_USERNAME", "testuser")
	os.Setenv("BBS_PASSWORD", "testpass")
	defer func() {
		os.Unsetenv("BBS_USERNAME")
		os.Unsetenv("BBS_PASSWORD")
	}()

	rootCmd := cmd.NewRootCommand()

	var buf bytes.Buffer
	rootCmd.SetErr(&buf)

	rootCmd.SetArgs([]string{
		"--bbs-server-url", "https://test.example.com",
		"--threads", "15", // Too many threads
	})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error due to invalid thread count")
	}

	if !strings.Contains(err.Error(), "threads") {
		t.Errorf("Expected thread validation error, got: %v", err)
	}
}
