# CI/CD Pipeline Setup

## Overview

This document describes the complete CI/CD pipeline setup for the veo3-cli project, including linting, security scanning, testing, and multi-platform builds.

## Components

### 1. Makefile Targets

The Makefile provides several targets for local development and CI/CD automation:

#### Quality Check Targets

- **`make lint`** - Run golangci-lint to check code quality
  - Enforces code standards
  - Checks for common mistakes
  - Configured in `.golangci.yml`

- **`make security`** - Run gosec security scanner
  - Scans for security vulnerabilities
  - Generates `security-report.txt`
  - Automatically installs gosec if not present

- **`make test`** - Run all tests with race detection
  ```bash
  go test -v -race ./...
  ```

- **`make test-coverage`** - Run tests with 80% coverage requirement
  - Generates `coverage.out` and `coverage.html`
  - **BLOCKS** if coverage is below 80%
  - Displays coverage summary

- **`make check`** - Run all quality checks (lint, security, test)
  - Comprehensive pre-commit validation
  - Runs lint → security → test in sequence

#### Build Targets

- **`make build`** - Build binary (runs lint and test first)
  - **Automatically runs `make pre-build`** before building
  - Embeds version and build time
  - Creates `veo3` binary

- **`make build-all`** - Build for all platforms and architectures
  - Builds for: linux, darwin, windows
  - Architectures: amd64, arm64
  - Outputs to `dist/` directory
  - Each binary named: `veo3-{os}-{arch}`

- **`make pre-build`** - Run lint and test before build
  - Called automatically by `make build`
  - Ensures code quality before compilation

#### Other Targets

- **`make fmt`** - Format code with gofmt and goimports
- **`make vet`** - Run go vet
- **`make tidy`** - Tidy go modules
- **`make clean`** - Remove all build artifacts
- **`make help`** - Show all available targets

### 2. GitHub Actions Workflow

Located at: `.github/workflows/ci.yml`

#### Jobs

1. **Lint Job**
   - Runs on: `ubuntu-latest`
   - Uses: `golangci/golangci-lint-action@v3`
   - Timeout: 5 minutes
   - Caches Go modules for speed

2. **Security Job**
   - Runs on: `ubuntu-latest`
   - Uses: `securego/gosec@master`
   - Outputs: SARIF format
   - Uploads results to GitHub Code Scanning

3. **Test Job**
   - Runs on: `ubuntu-latest`
   - Runs tests with race detection
   - **Enforces 80% coverage threshold**
   - Uploads coverage to Codecov

4. **Build Job**
   - **Depends on**: lint, security, test (all must pass)
   - **Matrix build** for:
     - OS: linux, darwin, windows
     - Arch: amd64, arm64
   - Produces 6 artifacts total
   - Uploads artifacts with 30-day retention

5. **Release Job** (only on tags)
   - **Depends on**: build job
   - Triggers on: `refs/tags/v*`
   - Downloads all build artifacts
   - Creates GitHub Release
   - Attaches all binaries to release

#### Workflow Triggers

- **Push** to `main` or `develop` branches
- **Pull Request** to `main` or `develop` branches
- **Manual trigger** via `workflow_dispatch`
- **Tags** matching `v*` (e.g., `v1.0.0`)

### 3. Linter Configuration

File: `.golangci.yml`

#### Enabled Linters

- `govet` - Go's built-in vet tool (all checks enabled)
- `errcheck` - Checks for unchecked errors
- `staticcheck` - Advanced static analysis
- `unused` - Checks for unused code
- `ineffassign` - Detects ineffectual assignments
- `gocyclo` - Cyclomatic complexity (max: 15)
- `misspell` - Spelling mistakes
- `unparam` - Unused function parameters
- `gosec` - Security issues

#### Settings

- Line length: 140 characters
- Complexity threshold: 15
- Timeout: 5 minutes
- Tests included in analysis

### 4. Coverage Requirements

**Minimum Coverage: 80%**

This is enforced at multiple levels:

1. **Makefile**: `make test-coverage` fails if < 80%
2. **GitHub Actions**: Test job fails if < 80%
3. **Pull Requests**: Cannot merge if coverage check fails

Coverage is calculated using:
```bash
go test -coverprofile=coverage.out -covermode=atomic ./...
```

## Local Development Workflow

### Before Committing

```bash
# Run all quality checks
make check

# Or run individually
make lint
make security  
make test
```

### Building Locally

```bash
# Build for current platform (runs lint & test first)
make build

# Build for all platforms
make build-all
```

### Checking Coverage

```bash
# Run with coverage report
make test-coverage

# View HTML coverage report
make coverage
open coverage.html
```

## CI/CD Workflow

### On Pull Request

1. **Lint** job runs
   - If lint fails → PR blocked
   
2. **Security** job runs
   - Scans for vulnerabilities
   - Uploads to GitHub Security
   
3. **Test** job runs
   - Runs all tests
   - Checks 80% coverage
   - If coverage < 80% → PR blocked
   
4. **Build** job runs (only if all above pass)
   - Builds for all platforms
   - Uploads artifacts

### On Merge to Main

Same as pull request, plus:
- All artifacts are uploaded
- Available for download for 30 days

### On Tag (Release)

1. All quality checks run
2. Build job creates all platform binaries
3. **Release** job:
   - Creates GitHub Release
   - Attaches all binaries
   - Generates release notes

## Creating a Release

1. Ensure all tests pass
2. Create and push a tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
3. GitHub Actions automatically:
   - Runs all checks
   - Builds all platforms
   - Creates release
   - Attaches binaries

## Platform Support

### Supported Platforms

- **Linux** (amd64, arm64)
- **macOS** (amd64, arm64)
- **Windows** (amd64, arm64)

### Binary Naming

- Linux: `veo3-linux-amd64`, `veo3-linux-arm64`
- macOS: `veo3-darwin-amd64`, `veo3-darwin-arm64`
- Windows: `veo3-windows-amd64.exe`, `veo3-windows-arm64.exe`

## Quality Gates Summary

| Gate | Tool | Threshold | Blocking |
|------|------|-----------|----------|
| Linting | golangci-lint | 0 issues | Yes |
| Security | gosec | Report only | No* |
| Test Coverage | go test | ≥ 80% | Yes |
| Race Detection | go test -race | 0 races | Yes |
| Build | go build | Must compile | Yes |

*Security findings are reported but don't block (review required)

## Troubleshooting

### Coverage Below 80%

```bash
# Generate detailed coverage report
make test-coverage

# View in browser
make coverage
open coverage.html

# Find uncovered lines and add tests
```

### Linter Errors

```bash
# Run linter locally
make lint

# Auto-format code
make fmt

# Check specific linter issues
golangci-lint run --enable-only=errcheck
```

### Build Failures

```bash
# Clean and rebuild
make clean
make build

# Check for dependency issues
go mod tidy
go mod verify
```

## Best Practices

1. **Always run `make check` before committing**
2. **Write tests for new features** (maintain 80% coverage)
3. **Fix linter issues immediately** (don't accumulate debt)
4. **Review security scan results** (even if not blocking)
5. **Tag releases semantically** (v1.0.0, v1.1.0, v2.0.0)

## Integration with Development Tools

### Pre-commit Hook (Recommended)

Create `.git/hooks/pre-commit`:
```bash
#!/bin/bash
make check
```

Make it executable:
```bash
chmod +x .git/hooks/pre-commit
```

### VS Code Integration

Install the Go extension and add to `.vscode/settings.json`:
```json
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.coverOnSave": true,
  "go.coverageDecorator": {
    "type": "gutter"
  }
}
```

### GitHub Branch Protection

Recommended settings for `main` branch:
- ✅ Require pull request reviews (1+ reviewers)
- ✅ Require status checks to pass
  - lint
  - security
  - test (with coverage check)
  - build
- ✅ Require branches to be up to date
- ✅ Require linear history

## Maintenance

### Updating Dependencies

```bash
# Update all dependencies
go get -u ./...
go mod tidy

# Run tests to verify
make test
```

### Updating Linter

```bash
# Update golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Update configuration if needed
# Edit .golangci.yml
```

### Updating CI/CD

GitHub Actions versions are pinned (e.g., `@v3`, `@v4`). Check for updates:
- `actions/checkout`
- `actions/setup-go`
- `golangci/golangci-lint-action`
- `securego/gosec`

## Metrics & Monitoring

### Coverage Trends

View coverage trends on Codecov dashboard:
- codecov.io/gh/{username}/go-veo3

### Build Times

Monitor GitHub Actions for build performance:
- Typical lint time: ~30s
- Typical test time: ~1-2min
- Typical build time per platform: ~30s
- Total pipeline time: ~5-7min

## References

- [golangci-lint documentation](https://golangci-lint.run/)
- [gosec documentation](https://github.com/securego/gosec)
- [GitHub Actions documentation](https://docs.github.com/en/actions)
- [Go testing documentation](https://golang.org/pkg/testing/)