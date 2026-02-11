# Development Guide

This guide is for developers who want to contribute to helm-oss or build it from source.

## Prerequisites

- **Go 1.25+** - Required for building the project
- **golangci-lint** - For code quality checks (optional but recommended)
- **Docker** - For building Docker images (optional)

## Building from Source

Clone the repository and build the binary:

```bash
# Clone the repository
git clone https://github.com/Timozer/helm-oss.git
cd helm-oss

# Build the binary
go build -o helm-oss ./cmd/helm-oss

# Or build with optimizations (smaller binary)
CGO_ENABLED=0 go build \
  -ldflags="-s -w" \
  -trimpath \
  -o helm-oss \
  ./cmd/helm-oss
```

## Running Tests

Run the test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...
```

## Code Quality

This project uses [golangci-lint](https://golangci-lint.run/) for code quality checks.

### Install golangci-lint

**macOS:**
```bash
brew install golangci-lint
```

**Linux:**
```bash
# Binary installation
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

**Using Go:**
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Run Linter

```bash
# Run all configured linters
golangci-lint run

# Run with auto-fix (fixes issues automatically when possible)
golangci-lint run --fix

# Run on specific files or directories
golangci-lint run ./cmd/...
```

### Linter Configuration

The project uses `.golangci.yml` for linter configuration. Key linters enabled:

- **errcheck** - Checks for unchecked errors
- **govet** - Examines Go source code and reports suspicious constructs
- **staticcheck** - Advanced static analysis
- **unused** - Checks for unused code
- **misspell** - Finds commonly misspelled words
- **gocyclo** - Checks cyclomatic complexity
- **goconst** - Finds repeated strings that could be constants
- **revive** - Fast, configurable, extensible linter

## Project Structure

```
helm-oss/
├── cmd/
│   └── helm-oss/          # Main application entry point
├── internal/
│   ├── helmutil/          # Helm utilities
│   └── oss/               # OSS storage implementation
├── docs/                  # Documentation
│   ├── en/                # English documentation
│   └── zh/                # Chinese documentation
├── .github/
│   └── workflows/         # GitHub Actions CI/CD
├── .golangci.yml          # Linter configuration
├── .goreleaser.yml        # Release configuration
├── go.mod                 # Go module definition
├── Dockerfile             # Docker image definition
└── plugin.yaml            # Helm plugin metadata
```

## Development Workflow

1. **Make changes** to the code
2. **Run tests** to ensure nothing breaks:
   ```bash
   go test ./...
   ```
3. **Run linter** to check code quality:
   ```bash
   golangci-lint run
   ```
4. **Build** the binary to verify compilation:
   ```bash
   go build ./cmd/helm-oss
   ```
5. **Test manually** with the built binary:
   ```bash
   ./helm-oss --help
   ```

## Building Docker Images

Build the Docker image locally:

```bash
# Build for current platform
docker build -t helm-oss:dev .

# Build for specific platform
docker buildx build --platform linux/amd64 -t helm-oss:dev .

# Build multi-platform
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t helm-oss:dev \
  .
```

## CI/CD Pipeline

The project uses GitHub Actions for CI/CD:

### CI Workflow (on every push/PR)

Runs 4 parallel jobs:
- **Test** - Run all tests with coverage
- **Lint** - Run golangci-lint
- **Build** - Verify compilation
- **Docker** - Verify Docker image builds

### Release Workflow (on tag push)

Automatically:
1. Runs tests
2. Builds binaries for multiple platforms
3. Builds and pushes Docker images (multi-arch)
4. Signs artifacts with GPG
5. Creates GitHub Release with all assets

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and linter
5. Commit your changes (`git commit -m 'feat: add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Commit Message Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `test:` - Test changes
- `chore:` - Maintenance tasks
- `refactor:` - Code refactoring
- `perf:` - Performance improvements

## Release Process

1. Ensure all tests pass
2. Update version in relevant files
3. Create and push a tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
4. GitHub Actions will automatically:
   - Build binaries for all platforms
   - Build and push Docker images
   - Create GitHub Release

## Troubleshooting

### Build Issues

If you encounter build issues:

```bash
# Clean Go cache
go clean -cache -modcache -i -r

# Download dependencies
go mod download

# Verify dependencies
go mod verify

# Rebuild
go build ./cmd/helm-oss
```

### Test Issues

If tests fail:

```bash
# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v -run TestName ./path/to/package
```

## Getting Help

- Open an [issue](https://github.com/Timozer/helm-oss/issues) for bugs
- Start a [discussion](https://github.com/Timozer/helm-oss/discussions) for questions
- Check existing issues and discussions first
