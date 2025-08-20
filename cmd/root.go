package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mona-actions/gh-bbs-analyzer/internal/config"
	"github.com/mona-actions/gh-bbs-analyzer/internal/output"
	"github.com/mona-actions/gh-bbs-analyzer/pkg/analyzer"
)

const (
	version     = "0.2.0"
	description = "GitHub CLI extension to analyze BitBucket Server for migration statistics"
)

// NewRootCommand creates the root cobra command
func NewRootCommand() *cobra.Command {
	var cfg config.Config

	rootCmd := &cobra.Command{
		Use:           "gh bbs-analyzer",
		Short:         description,
		Long:          description,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnalysis(&cfg)
		},
	}

	// Add flags
	addFlags(rootCmd, &cfg)

	// Set up Viper to bind flags
	bindFlags(rootCmd)

	return rootCmd
}

// addFlags adds all command-line flags
func addFlags(cmd *cobra.Command, cfg *config.Config) {
	flags := cmd.PersistentFlags()

	// BitBucket Server configuration
	flags.StringVarP(&cfg.BitBucket.ServerURL, "bbs-server-url", "s", "",
		"The full URL of the Bitbucket Server/Data Center to migrate from. E.g. http://bitbucket.contoso.com:7990")
	flags.StringVarP(&cfg.BitBucket.Username, "bbs-username", "u", "",
		"The Bitbucket username of a user with site admin privileges. If not set will be read from BBS_USERNAME environment variable.")
	flags.StringVarP(&cfg.BitBucket.Password, "bbs-password", "p", "",
		"The Bitbucket password of the user specified by --bbs-username. If not set will be read from BBS_PASSWORD environment variable.")
	flags.StringVar(&cfg.BitBucket.Project, "bbs-project", "",
		"A specific Bitbucket project instead of analyzing all projects.")

	// SSL configuration
	flags.BoolVar(&cfg.NoSSLVerify, "no-ssl-verify", false,
		"Disables SSL verification when communicating with your Bitbucket Server/Data Center instance.")

	// Output configuration
	flags.StringVarP(&cfg.Output.File, "output-file", "o", "results.csv",
		"The file to output the results to.")

	// Performance configuration
	flags.IntVarP(&cfg.Threads, "threads", "t", 3,
		"Number of threads to process concurrently.")

	// Mark required flags
	cmd.MarkPersistentFlagRequired("bbs-server-url")
}

// bindFlags binds flags to viper for configuration management
func bindFlags(cmd *cobra.Command) {
	viper.BindPFlag("bitbucket.server_url", cmd.PersistentFlags().Lookup("bbs-server-url"))
	viper.BindPFlag("bitbucket.username", cmd.PersistentFlags().Lookup("bbs-username"))
	viper.BindPFlag("bitbucket.password", cmd.PersistentFlags().Lookup("bbs-password"))
	viper.BindPFlag("bitbucket.project", cmd.PersistentFlags().Lookup("bbs-project"))
	viper.BindPFlag("no_ssl_verify", cmd.PersistentFlags().Lookup("no-ssl-verify"))
	viper.BindPFlag("output.file", cmd.PersistentFlags().Lookup("output-file"))
	viper.BindPFlag("threads", cmd.PersistentFlags().Lookup("threads"))
}

// runAnalysis executes the main analysis logic
func runAnalysis(cfg *config.Config) error {
	// Load configuration from flags and environment variables
	if err := cfg.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Create logger
	logger := output.NewLogger()

	// Display configuration
	displayConfiguration(cfg, logger)

	// Validate threads configuration
	if cfg.Threads > 3 {
		logger.Warning("Number of concurrent threads is higher than 3. This could result in extreme load on your server.")
	}

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Info("Interrupt received, cancelling operations...")
		cancel()
	}()

	// Create analyzer
	analyzer := analyzer.NewAnalyzerFromConfig(cfg, logger)

	// Run analysis
	results, err := analyzer.Run(ctx)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Display results
	displayResults(results, logger)

	// Write CSV output
	csvWriter := output.NewCSVWriter(cfg.Output.File)
	if err := csvWriter.WriteRepositories(results.Repositories); err != nil {
		return fmt.Errorf("failed to write CSV output: %w", err)
	}

	logger.Info(fmt.Sprintf("Results written to: %s", cfg.Output.File))

	return nil
}

// displayConfiguration shows the current configuration
func displayConfiguration(cfg *config.Config, logger *output.Logger) {
	logger.Info("---- CONFIGURATION ----")
	logger.Flag("BitBucket Server URL", cfg.BitBucket.ServerURL)
	logger.Flag("BitBucket Username", cfg.BitBucket.Username)
	logger.Flag("BitBucket Password", "**********")
	if cfg.BitBucket.Project != "" {
		logger.Flag("BitBucket Project", cfg.BitBucket.Project)
	}
	logger.Flag("SSL Verification Disabled", fmt.Sprintf("%t", cfg.NoSSLVerify))
	logger.Flag("Threads", fmt.Sprintf("%d", cfg.Threads))
	logger.Flag("Output File", cfg.Output.File)
	logger.Info("")
}

// displayResults shows the analysis results
func displayResults(results *analyzer.Results, logger *output.Logger) {
	logger.Info("---- ANALYSIS RESULTS ----")
	logger.Flag("Projects", fmt.Sprintf("%d", results.Statistics.TotalProjects))
	logger.Flag("Repositories", fmt.Sprintf("%d", results.Statistics.TotalRepos))
	logger.Flag("Pull Requests", fmt.Sprintf("%d", results.Statistics.TotalPRs))
	logger.Flag("Comments", fmt.Sprintf("%d", results.Statistics.TotalComments))
	logger.Flag("Total Disk Size", results.Statistics.FormattedSize)
	logger.Info("")
}
