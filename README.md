# gh-bbs-analyzer

[![build](https://github.com/mona-actions/gh-bbs-analyzer/actions/workflows/build.yaml/badge.svg)](https://github.com/mona-actions/gh-bbs-analyzer/actions/workflows/build.yaml)
[![release](https://github.com/mona-actions/gh-bbs-analyzer/actions/workflows/release.yaml/badge.svg)](https://github.com/mona-actions/gh-bbs-analyzer/actions/workflows/release.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mona-actions/gh-bbs-analyzer)](https://goreportcard.com/report/github.com/mona-actions/gh-bbs-analyzer)

A GitHub CLI extension for analyzing BitBucket Server instances to gather migration statistics and insights.

## ✨ Features

- **Repository Analysis**: Gather detailed statistics about repositories, including disk sizes, pull request counts, and comment volumes
- **Concurrent Processing**: Configurable multi-threaded analysis for faster processing
- **Flexible Configuration**: Support for environment variables and command-line flags
- **CSV Export**: Export results to CSV format for further analysis
- **Secure**: SSL verification support with optional bypass for self-signed certificates
- **Project Filtering**: Analyze specific projects or all projects

## 📊 Analyzed Metrics

- Repository disk sizes (with human-readable formatting)
- Pull request counts per repository
- Comment counts across all pull requests
- Repository archival status
- Repository visibility (public/private)
- Project and repository counts

## 🧪 Tested Versions

- BitBucket Server 8.6.1
- Additional versions welcome via community testing

## 🚀 Installation

### Prerequisites

- [GitHub CLI](https://cli.github.com/manual/installation) v2.0.0 or later
- Go 1.21+ (for building from source)

### Install via GitHub CLI

```bash
gh extension install mona-actions/gh-bbs-analyzer
```

### Upgrade

```bash
gh extension upgrade bbs-analyzer
```

### Build from Source

```bash
git clone https://github.com/mona-actions/gh-bbs-analyzer.git
cd gh-bbs-analyzer
go build -o gh-bbs-analyzer
gh extension install .
```

## 🔐 Permissions Required

To ensure comprehensive analysis, use a BitBucket Server system administrator account. This provides access to:

- All projects and repositories
- Complete repository metadata
- Full pull request and comment data

**API Endpoints Used:**

- `<server>/rest/api/1.0/projects`
- `<server>/rest/api/1.0/projects/<project-key>/repos`
- `<server>/projects/<project-key>/repos/<repo-slug>/sizes` (non-API endpoint)
- `<server>/rest/api/1.0/projects/<project-key>/repos/<repo-slug>/pull-requests`

## 📖 Usage

### Basic Usage

```bash
gh bbs-analyzer --bbs-server-url https://bitbucket.example.com:7990 \
                --bbs-username admin \
                --bbs-password your-password
```

### Environment Variables

Set credentials via environment variables for security:

```bash
export BBS_USERNAME=admin
export BBS_PASSWORD=your-password
gh bbs-analyzer --bbs-server-url https://bitbucket.example.com:7990
```

### Command Line Options

```txt
Usage:
  gh bbs-analyzer [flags]

Flags:
  -s, --bbs-server-url string   The full URL of the Bitbucket Server/Data Center
  -u, --bbs-username string     Bitbucket username with admin privileges
  -p, --bbs-password string     Bitbucket password (prefer BBS_PASSWORD env var)
      --bbs-project string      Analyze specific project only
      --no-ssl-verify           Disable SSL verification for self-signed certificates
  -o, --output-file string      Output CSV file (default "results.csv")
  -t, --threads int             Concurrent processing threads (default 3, max 10)
  -h, --help                    Show help information
  -v, --version                 Show version information
```

### Examples

Analyze all projects with custom output:

```bash
gh bbs-analyzer -s https://bitbucket.company.com:7990 \
                -u admin \
                -o company-migration-report.csv \
                -t 5
```

Analyze specific project only:

```bash
gh bbs-analyzer -s https://bitbucket.company.com:7990 \
                -u admin \
                --bbs-project MYPROJECT
```

Self-signed SSL certificate:

```bash
gh bbs-analyzer -s https://bitbucket.company.com:7990 \
                -u admin \
                --no-ssl-verify
```

## 📈 Output

### Console Output

- Real-time progress with spinner
- Configuration summary
- Analysis results summary
- Color-coded warnings and errors

### CSV Export

The tool generates a CSV file with the following columns:

- `project`: Project key
- `repository`: Repository name
- `size`: Repository size in bytes
- `pull_requests`: Number of pull requests
- `comments`: Total comment count
- `archived`: Whether repository is archived (true/false)
- `public`: Whether repository is public (true/false)

## 🛠️ Development

### Setup Development Environment

```bash
git clone https://github.com/mona-actions/gh-bbs-analyzer.git
cd gh-bbs-analyzer
go mod tidy
```

### Building and Testing

```bash
# Build the application
go build -o gh-bbs-analyzer

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Format code
go fmt ./...

# Run linter (if golangci-lint is installed)
golangci-lint run

# Install extension locally
gh extension install .

# Clean build artifacts
go clean
```

### Project Structure

```
├── cmd/                    # Command definitions
├── internal/              # Private application code
│   ├── bitbucket/        # BitBucket API client
│   ├── config/           # Configuration management
│   └── output/           # Output and logging
├── pkg/                  # Public libraries
│   └── analyzer/         # Core analysis logic
└── main.go              # Application entry point
```

## 🔮 Roadmap

- **File Analysis**: Large file detection (>100MB)
- **LFS Detection**: Git LFS usage analysis
- **Attachment Metrics**: Repository attachment analysis
- **Advanced Reporting**: Additional export formats (JSON, HTML)
- **Performance Metrics**: Analysis timing and performance data
- **Incremental Analysis**: Delta analysis capabilities

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## ⚠️ Disclaimer

This tool is designed for migration planning and analysis. Always:

- Test in non-production environments first
- Ensure you have proper permissions before running analysis
- Be mindful of server load when using high thread counts
- Review the generated data before making migration decisions
