# GitHub CLI Extension: gh-bbs-analyzer

## Project Overview

This is a GitHub CLI extension written in Go that analyzes BitBucket Server instances to gather migration statistics and insights. The tool helps organizations assess their BitBucket repositories before migrating to GitHub, providing detailed metrics about repository sizes, pull requests, comments, and other migration-relevant data.

## Application Architecture

### Core Design Principles

- **Clean Architecture**: Separation of concerns with clear package boundaries
- **Dependency Injection**: Configurable components with minimal coupling
- **Configuration Management**: Command-line flags and environment variables
- **Concurrent Processing**: Worker pool pattern for efficient API processing
- **Error Handling**: Comprehensive error propagation with context
- **Observability**: Structured logging with progress indication

### Package Structure

```
├── cmd/                    # Command-line interface layer
│   └── root.go            # Cobra command definitions and flag binding
├── internal/              # Private application logic
│   ├── bitbucket/        # BitBucket API client implementation
│   │   ├── client.go     # HTTP client with retry logic
│   │   └── types.go      # API response data structures
│   ├── config/           # Configuration management
│   │   ├── config.go     # Viper-based config loading/validation
│   │   └── config_test.go # Configuration tests
│   └── output/           # Output formatting and logging
│       ├── logger.go     # Structured logging with color output
│       └── csv.go        # CSV export functionality
├── pkg/                  # Public/reusable components
│   └── analyzer/         # Core analysis engine
│       └── analyzer.go   # Main business logic orchestration
└── main.go               # Application entry point
```

## Core Components

### 1. Command Layer (`cmd/`)

**Purpose**: Handle CLI interactions using Cobra framework
**Key Features**:

- Command definition and flag management
- Viper integration for configuration binding
- Input validation and error handling
- Version and help information

**Best Practices**:

- Use Cobra's persistent flags for global options
- Implement proper flag validation and required flag checking
- Provide clear help text and usage examples
- Support both short and long flag formats

### 2. Configuration Management (`internal/config/`)

**Purpose**: Centralized configuration handling with command-line flags and environment variables
**Key Features**:

- Environment variable binding (`BBS_USERNAME`, `BBS_PASSWORD`)
- Command-line flag support with Viper integration
- Configuration validation with meaningful error messages
- No configuration file dependency

**Configuration Sources (Priority Order)**:

1. Command-line flags (highest priority)
2. Environment variables
3. Default values (lowest priority)

**Security Considerations**:

- Never log passwords or sensitive credentials
- Support environment variables for sensitive data
- Validate SSL/TLS settings appropriately
- No sensitive data storage in configuration files

### 3. BitBucket API Client (`internal/bitbucket/`)

**Purpose**: Handle all interactions with BitBucket Server REST API
**Key Features**:

- HTTP client with configurable timeouts and retry logic
- SSL verification bypass for self-signed certificates
- Rate limiting and error handling
- Structured API response types

**API Endpoints Used**:

- `/rest/api/1.0/projects` - List all projects
- `/rest/api/1.0/projects/{key}/repos` - List repositories in project
- `/projects/{key}/repos/{slug}/sizes` - Get repository disk usage
- `/rest/api/1.0/projects/{key}/repos/{slug}/pull-requests` - List pull requests

**Error Handling**:

- Implement exponential backoff for rate limiting
- Handle authentication failures gracefully
- Provide meaningful error messages for API failures
- Support partial failures during bulk operations

### 4. Analysis Engine (`pkg/analyzer/`)

**Purpose**: Orchestrate the analysis workflow and coordinate data collection
**Key Features**:

- Concurrent processing with configurable worker pools
- Progress tracking with visual feedback (spinner)
- Context-aware cancellation support
- Result aggregation and statistics calculation

**Processing Flow**:

1. Authenticate with BitBucket Server
2. Discover projects (all or specific project)
3. Enumerate repositories within projects
4. Collect repository metrics concurrently
5. Aggregate results and generate statistics
6. Export data to specified format

### 5. Output and Logging (`internal/output/`)

**Purpose**: Handle result formatting and user feedback
**Key Features**:

- CSV export with configurable output paths
- Colored console output for better UX
- Structured logging with different verbosity levels
- Progress indication during long-running operations

## Technical Requirements

### Dependencies

**Core Libraries**:

- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `github.com/fatih/color` - Colored terminal output
- `github.com/briandowns/spinner` - Progress indication

**Development Tools**:

- `golangci-lint` - Comprehensive linting
- Go modules for dependency management
- Make for build automation

### Go Version

- **Minimum**: Go 1.21+
- **Recommended**: Latest stable Go version

### Build System

Use standard Go tooling for development tasks:

```bash
go build -o gh-bbs-analyzer    # Build the application
go test ./...                  # Run tests with coverage
golangci-lint run             # Run linter checks
go fmt ./...                  # Format code
gh extension install .        # Install as GitHub CLI extension
go clean                     # Clean build artifacts
go mod tidy                   # Setup development dependencies
```

## GitHub CLI Extension Standards

### Extension Naming

- **Binary Name**: `gh-bbs-analyzer`
- **Command**: `gh bbs-analyzer`
- **Repository**: `mona-actions/gh-bbs-analyzer`

### Extension Manifest

Ensure the repository contains:

- Proper `gh-extension` topic
- Clear description and documentation
- Release binaries for multiple platforms
- Installation instructions

### User Experience

**Command Structure**:

```bash
gh bbs-analyzer [global-flags] [command-specific-flags]
```

**Help and Documentation**:

- Comprehensive `--help` output
- Clear error messages with actionable guidance
- Examples in help text
- Version information (`--version`)

**Output Standards**:

- Human-readable progress indication
- Structured data export (CSV)
- Colorized console output
- Proper exit codes

## Security Best Practices

### Credential Management

- **Environment Variables**: Prefer `BBS_USERNAME`/`BBS_PASSWORD` env vars
- **Command Line**: Accept but don't log password flags
- **Memory**: Clear sensitive data when possible

### SSL/TLS Handling

- **Default**: Enable SSL verification
- **Override**: Provide `--no-ssl-verify` flag for self-signed certificates
- **Warnings**: Alert users when SSL verification is disabled

### API Rate Limiting

- Implement exponential backoff
- Respect server rate limits
- Allow configuration of concurrent threads
- Provide graceful degradation under load

## Testing Strategy

### Unit Tests

- **Config Package**: Test configuration loading and validation
- **BitBucket Client**: Mock API responses and error conditions
- **Analyzer**: Test business logic with dependency injection
- **Output**: Test CSV generation and logging

### Integration Tests

- Test full CLI command execution
- Test error handling scenarios

### Test Coverage

- Target minimum 80% test coverage
- Use table-driven tests for multiple scenarios
- Test both success and failure paths

## Performance Considerations

### Concurrency

- **Worker Pools**: Configurable concurrent processing (default: 3 threads)
- **Context Cancellation**: Support for graceful shutdown
- **Resource Management**: Proper cleanup of HTTP connections
- **Memory Usage**: Stream processing for large datasets

### API Efficiency

- **Pagination**: Handle paginated API responses properly
- **Bulk Operations**: Batch API calls where possible
- **Caching**: Cache project/repository lists during analysis
- **Request Optimization**: Minimize redundant API calls

## Error Handling

### Error Categories

1. **Configuration Errors**: Invalid config values, missing required fields
2. **Authentication Errors**: Invalid credentials, permission issues
3. **Network Errors**: Connection timeouts, DNS resolution failures
4. **API Errors**: Rate limiting, server errors, malformed responses
5. **File System Errors**: Unable to write output files

### Error Response

- Provide clear, actionable error messages
- Include relevant context (which repository, API endpoint, etc.)
- Suggest resolution steps where possible
- Use appropriate exit codes for different error types

## Extensibility Points

### Output Formats

- **Current**: CSV export
- **Future**: JSON, HTML reports, custom templates
- **Design**: Abstract output interface for new formats

### Analysis Metrics

- **Current**: Repository sizes, PR counts, comment volumes
- **Future**: File analysis, LFS detection, attachment metrics
- **Design**: Plugin-style metric collectors

### API Clients

- **Current**: BitBucket Server
- **Future**: BitBucket Cloud, other source control systems
- **Design**: Interface-based client abstraction

## Documentation Requirements

### Code Documentation

- Package-level documentation for all packages
- Function documentation for public APIs
- Inline comments for complex business logic
- Examples in documentation where helpful

### User Documentation

- **README**: Installation, usage, examples
- **Configuration**: Complete configuration reference
- **Troubleshooting**: Common issues and solutions
- **Migration Guide**: How to use results for planning

## Release Management

### Versioning

- Use semantic versioning (MAJOR.MINOR.PATCH)
- Tag releases in Git
- Maintain CHANGELOG.md

### Distribution

- **GitHub Releases**: Automated builds for multiple platforms
- **GitHub CLI**: Installable via `gh extension install`
- **Binary Assets**: Provide compiled binaries for major platforms

### Compatibility

- Maintain backward compatibility for command-line flags
- Document breaking changes clearly
- Provide migration guides for major version updates

## Monitoring and Observability

### Logging

- Structured logging with appropriate levels
- Configuration validation logging
- API call logging (without sensitive data)
- Performance metrics logging

### Metrics Collection

- Track analysis duration
- Monitor API response times
- Count processed repositories and projects
- Record error rates by category

## Future Enhancements

### Planned Features

1. **Enhanced Analysis**: Large file detection, LFS usage analysis
2. **Advanced Reporting**: HTML reports, trend analysis
3. **Incremental Analysis**: Delta processing capabilities
4. **Multi-Source Support**: Support for other SCM platforms
5. **CI/CD Integration**: GitHub Actions for automated analysis

### Technical Debt Management

- Regular dependency updates
- Performance optimization
- Code quality improvements
- Test coverage expansion

---

This specification provides a comprehensive guide for building and maintaining the `gh-bbs-analyzer` GitHub CLI extension following Go and GitHub CLI best practices.
