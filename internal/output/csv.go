package output

import (
	"fmt"
	"os"
	"strconv"

	"github.com/mona-actions/gh-bbs-analyzer/internal/bitbucket"
)

// CSVWriter handles CSV output operations
type CSVWriter struct {
	filename string
}

// NewCSVWriter creates a new CSV writer
func NewCSVWriter(filename string) *CSVWriter {
	return &CSVWriter{filename: filename}
}

// WriteRepositories writes repository data to CSV file
func (w *CSVWriter) WriteRepositories(repositories []bitbucket.Repository) error {
	file, err := os.Create(w.filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Write header
	header := "project,repository,size,pull_requests,comments,archived,public\n"
	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write data
	for _, repo := range repositories {
		size := 0
		if repo.Size != nil {
			size = repo.Size.Repository
		}

		line := fmt.Sprintf("%s,%s,%d,%d,%d,%s,%s\n",
			repo.Project.Key,
			repo.Slug,
			size,
			len(repo.PullRequests),
			repo.CommentCount,
			strconv.FormatBool(repo.Archived),
			strconv.FormatBool(repo.Public),
		)

		if _, err := file.WriteString(line); err != nil {
			return fmt.Errorf("failed to write repository data: %w", err)
		}
	}

	return nil
}
