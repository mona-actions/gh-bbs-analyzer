package config

import (
	"fmt"
	"regexp"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	BitBucket   BitBucketConfig `mapstructure:"bitbucket"`
	Output      OutputConfig    `mapstructure:"output"`
	Threads     int             `mapstructure:"threads"`
	NoSSLVerify bool            `mapstructure:"no_ssl_verify"`
}

// BitBucketConfig holds BitBucket server configuration
type BitBucketConfig struct {
	ServerURL string `mapstructure:"server_url"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	Project   string `mapstructure:"project"`
}

// OutputConfig holds output configuration
type OutputConfig struct {
	File string `mapstructure:"file"`
}

// NewConfig creates a new configuration with defaults
func NewConfig() *Config {
	return &Config{
		BitBucket: BitBucketConfig{},
		Output: OutputConfig{
			File: "results.csv",
		},
		Threads:     3,
		NoSSLVerify: false,
	}
}

// Load loads configuration from command line flags and environment variables only
func (c *Config) Load() error {
	// Set defaults
	viper.SetDefault("output.file", "results.csv")
	viper.SetDefault("threads", 3)
	viper.SetDefault("no_ssl_verify", false)

	// Set environment variable prefix and enable automatic env binding
	viper.SetEnvPrefix("BBS")
	viper.AutomaticEnv()

	// Bind specific environment variables
	viper.BindEnv("bitbucket.username", "BBS_USERNAME")
	viper.BindEnv("bitbucket.password", "BBS_PASSWORD")

	// Unmarshal configuration from Viper (flags + env vars)
	if err := viper.Unmarshal(c); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate BitBucket server URL
	if c.BitBucket.ServerURL == "" {
		return fmt.Errorf("BitBucket server URL must be provided")
	}

	r, _ := regexp.Compile("^http(s|):(//|)")
	if !r.MatchString(c.BitBucket.ServerURL) {
		return fmt.Errorf("BitBucket server URL should contain http(s) prefix")
	}

	// Validate credentials
	if c.BitBucket.Username == "" {
		return fmt.Errorf("BitBucket username must be provided via --bbs-username flag or BBS_USERNAME environment variable")
	}

	if c.BitBucket.Password == "" {
		return fmt.Errorf("BitBucket password must be provided via --bbs-password flag or BBS_PASSWORD environment variable")
	}

	// Validate threads
	if c.Threads > 10 {
		return fmt.Errorf("number of concurrent threads cannot be higher than 10")
	}

	return nil
}
