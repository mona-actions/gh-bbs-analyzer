package config_test

import (
	"os"
	"testing"

	"github.com/mona-actions/gh-bbs-analyzer/internal/config"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &config.Config{
				BitBucket: config.BitBucketConfig{
					ServerURL: "https://bitbucket.example.com",
					Username:  "testuser",
					Password:  "testpass",
				},
				Threads: 3,
			},
			wantErr: false,
		},
		{
			name: "missing server URL",
			config: &config.Config{
				BitBucket: config.BitBucketConfig{
					Username: "testuser",
					Password: "testpass",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid server URL",
			config: &config.Config{
				BitBucket: config.BitBucketConfig{
					ServerURL: "bitbucket.example.com",
					Username:  "testuser",
					Password:  "testpass",
				},
			},
			wantErr: true,
		},
		{
			name: "missing username",
			config: &config.Config{
				BitBucket: config.BitBucketConfig{
					ServerURL: "https://bitbucket.example.com",
					Password:  "testpass",
				},
			},
			wantErr: true,
		},
		{
			name: "missing password",
			config: &config.Config{
				BitBucket: config.BitBucketConfig{
					ServerURL: "https://bitbucket.example.com",
					Username:  "testuser",
				},
			},
			wantErr: true,
		},
		{
			name: "too many threads",
			config: &config.Config{
				BitBucket: config.BitBucketConfig{
					ServerURL: "https://bitbucket.example.com",
					Username:  "testuser",
					Password:  "testpass",
				},
				Threads: 11,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Load_EnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("BBS_USERNAME", "envuser")
	os.Setenv("BBS_PASSWORD", "envpass")
	defer func() {
		os.Unsetenv("BBS_USERNAME")
		os.Unsetenv("BBS_PASSWORD")
	}()

	cfg := config.NewConfig()
	err := cfg.Load()
	if err != nil {
		t.Fatalf("Config.Load() error = %v", err)
	}

	if cfg.BitBucket.Username != "envuser" {
		t.Errorf("Expected username from env var, got %s", cfg.BitBucket.Username)
	}

	if cfg.BitBucket.Password != "envpass" {
		t.Errorf("Expected password from env var, got %s", cfg.BitBucket.Password)
	}
}
