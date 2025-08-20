package output

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mona-actions/gh-bbs-analyzer/internal/bitbucket"
)

// Results represents analysis results (defined locally to avoid import cycle)
type Results struct {
	Projects     []bitbucket.Project
	Repositories []bitbucket.Repository
	Statistics   Statistics
}

// Statistics represents analysis statistics (defined locally to avoid import cycle)
type Statistics struct {
	TotalProjects    int
	TotalRepos       int
	TotalPRs         int
	TotalComments    int
	TotalSize        int64
	FormattedSize    string
	TotalBranches    int
	TotalTags        int
	TotalForks       int
	TotalWebhooks    int
	TotalLFSFiles    int
	TotalLFSSize     int64
	FormattedLFSSize string
	TotalPermissions int
	TotalHooks       int
	ActiveRepos      int
	ArchivedRepos    int
}

// formatSize formats bytes into human-readable format (local copy to avoid import cycle)
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%d GB", bytes/GB)
	case bytes >= MB:
		return fmt.Sprintf("%d MB", bytes/MB)
	case bytes >= KB:
		return fmt.Sprintf("%d KB", bytes/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// MigrationReportWriter handles migration report generation
type MigrationReportWriter struct {
	filename string
}

// NewMigrationReportWriter creates a new migration report writer
func NewMigrationReportWriter(filename string) *MigrationReportWriter {
	return &MigrationReportWriter{filename: filename}
}

// WriteMigrationReport writes a comprehensive migration analysis report
func (w *MigrationReportWriter) WriteMigrationReport(results *Results) error {
	file, err := os.Create(w.filename)
	if err != nil {
		return fmt.Errorf("failed to create migration report file: %w", err)
	}
	defer file.Close()

	report := w.generateMigrationReport(results)

	if _, err := file.WriteString(report); err != nil {
		return fmt.Errorf("failed to write migration report: %w", err)
	}

	return nil
}

func (w *MigrationReportWriter) generateMigrationReport(results *Results) string {
	var report strings.Builder

	// Header
	report.WriteString("# BitBucket to GitHub Migration Analysis Report\n\n")
	report.WriteString(fmt.Sprintf("Generated on: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Executive Summary
	report.WriteString("## Executive Summary\n\n")
	report.WriteString(fmt.Sprintf("- **%d** projects analyzed\n", results.Statistics.TotalProjects))
	report.WriteString(fmt.Sprintf("- **%d** repositories found (%d active, %d archived)\n",
		results.Statistics.TotalRepos, results.Statistics.ActiveRepos, results.Statistics.ArchivedRepos))
	report.WriteString(fmt.Sprintf("- **%s** total repository data (%s in Git LFS)\n",
		results.Statistics.FormattedSize, results.Statistics.FormattedLFSSize))
	report.WriteString(fmt.Sprintf("- **%d** pull requests with **%d** comments to migrate\n",
		results.Statistics.TotalPRs, results.Statistics.TotalComments))
	report.WriteString("\n")

	// Migration Complexity Assessment
	report.WriteString("## Migration Complexity Assessment\n\n")

	// Code Structure Complexity
	report.WriteString("### Code Structure\n")
	report.WriteString(fmt.Sprintf("- **%d** branches across all repositories\n", results.Statistics.TotalBranches))
	report.WriteString(fmt.Sprintf("- **%d** tags/releases to preserve\n", results.Statistics.TotalTags))
	report.WriteString(fmt.Sprintf("- **%d** repository forks requiring network migration\n", results.Statistics.TotalForks))

	if results.Statistics.TotalLFSFiles > 0 {
		report.WriteString(fmt.Sprintf("- **%d** Git LFS files (%s) - requires GitHub LFS planning\n",
			results.Statistics.TotalLFSFiles, results.Statistics.FormattedLFSSize))
	}
	report.WriteString("\n")

	// Integration Complexity
	report.WriteString("### Integration & Automation\n")
	report.WriteString(fmt.Sprintf("- **%d** webhooks requiring reconfiguration\n", results.Statistics.TotalWebhooks))
	report.WriteString(fmt.Sprintf("- **%d** repository hooks needing GitHub equivalent\n", results.Statistics.TotalHooks))
	report.WriteString("\n")

	// Access Control Complexity
	report.WriteString("### Access Control\n")
	report.WriteString(fmt.Sprintf("- **%d** user permissions to migrate\n", results.Statistics.TotalPermissions))
	report.WriteString("- Review required for GitHub's different permission model\n")
	report.WriteString("\n")

	// Repository Prioritization
	report.WriteString("## Repository Migration Prioritization\n\n")

	activeRepos := []bitbucket.Repository{}
	archivedRepos := []bitbucket.Repository{}
	largeRepos := []bitbucket.Repository{}
	lfsRepos := []bitbucket.Repository{}
	complexRepos := []bitbucket.Repository{}

	for _, repo := range results.Repositories {
		if repo.Archived {
			archivedRepos = append(archivedRepos, repo)
		} else if repo.CommitCount > 0 {
			activeRepos = append(activeRepos, repo)
		}

		if repo.Size != nil && repo.Size.Repository > 100*1024*1024 { // >100MB
			largeRepos = append(largeRepos, repo)
		}

		if repo.LFSFileCount > 0 {
			lfsRepos = append(lfsRepos, repo)
		}

		if repo.WebhookCount > 0 || len(repo.Hooks) > 0 || repo.ForkCount > 0 {
			complexRepos = append(complexRepos, repo)
		}
	}

	// High Priority - Active with complexity
	report.WriteString("### High Priority (Active + Complex)\n")
	written := false
	for _, repo := range activeRepos {
		if repo.WebhookCount > 0 || len(repo.Hooks) > 0 || repo.ForkCount > 0 {
			report.WriteString(fmt.Sprintf("- **%s/%s** - ", repo.Project.Key, repo.Slug))
			reasons := []string{}
			if repo.WebhookCount > 0 {
				reasons = append(reasons, fmt.Sprintf("%d webhooks", repo.WebhookCount))
			}
			if len(repo.Hooks) > 0 {
				reasons = append(reasons, fmt.Sprintf("%d hooks", len(repo.Hooks)))
			}
			if repo.ForkCount > 0 {
				reasons = append(reasons, fmt.Sprintf("%d forks", repo.ForkCount))
			}
			report.WriteString(strings.Join(reasons, ", "))
			report.WriteString("\n")
			written = true
		}
	}
	if !written {
		report.WriteString("- No high-priority repositories identified\n")
	}
	report.WriteString("\n")

	// Medium Priority - Large or LFS repositories
	report.WriteString("### Medium Priority (Large Files)\n")
	written = false
	for _, repo := range lfsRepos {
		report.WriteString(fmt.Sprintf("- **%s/%s** - %d LFS files (%s)\n",
			repo.Project.Key, repo.Slug, repo.LFSFileCount,
			formatSize(repo.LFSSize)))
		written = true
	}
	for _, repo := range largeRepos {
		if repo.LFSFileCount == 0 { // Don't duplicate LFS repos
			size := "unknown"
			if repo.Size != nil {
				size = formatSize(int64(repo.Size.Repository))
			}
			report.WriteString(fmt.Sprintf("- **%s/%s** - Large repository (%s)\n",
				repo.Project.Key, repo.Slug, size))
			written = true
		}
	}
	if !written {
		report.WriteString("- No large file repositories identified\n")
	}
	report.WriteString("\n")

	// Low Priority - Archived
	report.WriteString("### Low Priority (Archived)\n")
	if len(archivedRepos) > 0 {
		report.WriteString(fmt.Sprintf("- **%d** archived repositories\n", len(archivedRepos)))
		report.WriteString("- Consider whether migration is necessary\n")
		report.WriteString("- Can be migrated last or potentially excluded\n")
	} else {
		report.WriteString("- No archived repositories found\n")
	}
	report.WriteString("\n")

	// Action Items
	report.WriteString("## Recommended Action Items\n\n")

	report.WriteString("### Pre-Migration Planning\n")
	report.WriteString("1. **Access Control Mapping**\n")
	report.WriteString("   - Map BitBucket permissions to GitHub teams/roles\n")
	report.WriteString("   - Plan GitHub organization structure\n\n")

	if results.Statistics.TotalWebhooks > 0 || results.Statistics.TotalHooks > 0 {
		report.WriteString("2. **Integration Migration**\n")
		if results.Statistics.TotalWebhooks > 0 {
			report.WriteString(fmt.Sprintf("   - Inventory %d webhooks and their purposes\n", results.Statistics.TotalWebhooks))
		}
		if results.Statistics.TotalHooks > 0 {
			report.WriteString(fmt.Sprintf("   - Plan GitHub equivalents for %d repository hooks\n", results.Statistics.TotalHooks))
		}
		report.WriteString("   - Consider GitHub Actions for CI/CD workflows\n\n")
	}

	if results.Statistics.TotalLFSFiles > 0 {
		report.WriteString("3. **Git LFS Planning**\n")
		report.WriteString(fmt.Sprintf("   - Budget for %s of LFS storage\n", results.Statistics.FormattedLFSSize))
		report.WriteString("   - Review GitHub LFS pricing and limits\n\n")
	}

	if results.Statistics.TotalForks > 0 {
		report.WriteString("4. **Fork Network Migration**\n")
		report.WriteString(fmt.Sprintf("   - Plan migration strategy for %d repository forks\n", results.Statistics.TotalForks))
		report.WriteString("   - Consider fork relationships and dependencies\n\n")
	}

	report.WriteString("### Migration Execution\n")
	report.WriteString("1. **Pilot Migration**\n")
	report.WriteString("   - Start with simple, low-risk repositories\n")
	report.WriteString("   - Test migration tools and processes\n\n")

	report.WriteString("2. **Batch Migration**\n")
	report.WriteString("   - Migrate active repositories first\n")
	report.WriteString("   - Handle complex repositories with extra planning\n")
	report.WriteString("   - Archive or exclude truly dormant repositories\n\n")

	report.WriteString("3. **Post-Migration**\n")
	report.WriteString("   - Verify all integrations work correctly\n")
	report.WriteString("   - Update team workflows and documentation\n")
	report.WriteString("   - Monitor GitHub usage and costs\n\n")

	// Footer
	report.WriteString("---\n")
	report.WriteString("*This report was generated by gh-bbs-analyzer*\n")

	return report.String()
}
