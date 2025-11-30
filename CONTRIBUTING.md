# Contributing to Veo3 CLI

Thank you for your interest in contributing to Veo3 CLI! This document provides guidelines and workflows for contributing to the project.

## Table of Contents

- [Development Setup](#development-setup)
- [Git Workflow](#git-workflow)
- [GitHub Issue Integration](#github-issue-integration)
- [Code Standards](#code-standards)
- [Testing Requirements](#testing-requirements)
- [Pull Request Process](#pull-request-process)

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git
- Make
- golangci-lint (for linting)

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/jasongoecke/go-veo3.git
cd go-veo3

# Install dependencies
go mod download

# Build the project
make build

# Run tests
make test

# Run linter
make lint
```

### Project Structure

```
go-veo3/
├── cmd/veo3/           # Main entry point
├── pkg/                # Reusable library packages
│   ├── veo3/          # API client and generation logic
│   ├── config/        # Configuration management
│   ├── operations/    # Operation lifecycle management
│   └── cli/           # CLI command implementations
├── internal/           # Private application code
│   ├── validation/    # Input validation
│   └── format/        # Output formatting
├── tests/              # Tests and fixtures
│   ├── unit/          # Unit tests
│   └── integration/   # Integration tests
├── specs/              # Design specifications
│   └── 001-veo3-cli/  # Current feature specs
└── .specify/           # SpecKit workflow files
```

## Git Workflow

### Branch Naming

- Feature branches: `feature/<issue-number>-<brief-description>`
- Bug fixes: `fix/<issue-number>-<brief-description>`
- Documentation: `docs/<brief-description>`

Example: `feature/42-add-batch-processing`

### Commit Message Format

Follow the Conventional Commits specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Build process or auxiliary tool changes

**Example:**
```
feat(cli): add batch processing command

Implement batch processing functionality that allows users to process
multiple video generation requests from a YAML manifest file.

Implements #42
Closes #42
```

## GitHub Issue Integration

### **REQUIRED: Git Commit → GitHub Issue Updates**

**Every git commit MUST reference and update corresponding GitHub Issues using the following rules:**

### 1. Reference Issues in Commits

Every commit message MUST include issue references in the footer:

```
feat(operations): implement status polling

Add exponential backoff for operation status polling to handle
long-running video generation tasks efficiently.

Implements #15
Updates #12
```

### 2. Automatic Issue Updates

Use these keywords in commit messages to automatically update issues:

**Progress Updates (keeps issue open):**
- `Updates #<issue>` - General progress update
- `Implements #<issue>` - Implementation in progress
- `Fixes #<issue>` - Fix in progress but not complete
- `Addresses #<issue>` - Addresses the issue partially

**Issue Closure (closes issue automatically):**
- `Closes #<issue>` - Completes the issue
- `Resolves #<issue>` - Resolves the issue
- `Completes #<issue>` - Marks as complete

### 3. Task Tracking in Commits

When completing tasks from specs/001-veo3-cli/tasks.md:

```
feat(config): implement configuration manager

Complete configuration loading and saving functionality with
XDG Base Directory support.

Tasks completed:
- [x] T010 Create pkg/config/manager.go
- [x] T098 Implement XDG Base Directory specification

Closes #35
```

### 4. Multiple Issues

If a commit affects multiple issues:

```
feat(cli): add generate command with validation

Implements text-to-video generation with full parameter validation
and error handling.

Implements #21
Updates #12
Closes #24
```

### 5. Cross-References

Link related issues or pull requests:

```
fix(validation): correct image size validation

Fix max image size check to use 20MB instead of 10MB per API spec.

Fixes #56
Related to #42
See also #38
```

### 6. Breaking Changes

Prefix with `BREAKING CHANGE:` in commit body:

```
feat(api)!: change client initialization signature

BREAKING CHANGE: NewClient now requires context.Context as first parameter

Before: veo3.NewClient(apiKey)
After: veo3.NewClient(ctx, apiKey)

Closes #67
```

### 7. Commit Hook (Recommended)

Create `.git/hooks/prepare-commit-msg` to enforce issue references:

```bash
#!/bin/bash
# Ensure commit messages reference issues

COMMIT_MSG_FILE=$1
COMMIT_SOURCE=$2

# Skip for merge commits
if [ "$COMMIT_SOURCE" = "merge" ]; then
  exit 0
fi

# Check if commit message has issue reference
if ! grep -qE '#[0-9]+' "$COMMIT_MSG_FILE"; then
  echo "ERROR: Commit message must reference at least one GitHub issue"
  echo "Use: Implements #<issue>, Updates #<issue>, or Closes #<issue>"
  exit 1
fi
```

Make it executable:
```bash
chmod +x .git/hooks/prepare-commit-msg
```

### 8. Workflow Example

**When starting work on an issue:**

1. Create feature branch from issue:
   ```bash
   git checkout -b feature/42-batch-processing
   ```

2. Make changes and commit with issue references:
   ```bash
   git commit -m "feat(batch): add manifest parser

   Implement YAML manifest parsing for batch processing.

   Implements #42"
   ```

3. Continue with multiple commits, each referencing the issue:
   ```bash
   git commit -m "feat(batch): add job processor

   Implement concurrent job execution with worker pool.

   Updates #42"
   ```

4. Final commit closes the issue:
   ```bash
   git commit -m "feat(batch): complete batch processing

   Add CLI integration and documentation for batch processing.

   Closes #42"
   ```

5. Push and create pull request:
   ```bash
   git push origin feature/42-batch-processing
   ```

## Code Standards

### Go Style Guide

Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines and:

- Use `gofmt` for formatting
- Use `golangci-lint` for linting
- Follow standard Go naming conventions
- Add godoc comments for all exported functions

### Code Organization

- **Library-first design**: Implement functionality in `pkg/` packages
- **Thin CLI layer**: CLI commands in `pkg/cli/` should be thin wrappers
- **Private code**: Use `internal/` for implementation details
- **Clear separation**: Keep API client, business logic, and CLI separate

### Example Structure

```go
// pkg/veo3/generate.go - Core logic
func GenerateVideo(ctx context.Context, req *GenerationRequest) (*Operation, error) {
    // Implementation
}

// pkg/cli/generate.go - CLI wrapper
func runGenerate(cmd *cobra.Command, args []string) error {
    req := buildRequest(cmd, args)
    op, err := client.GenerateVideo(ctx, req)
    return handleResult(op, err)
}
```

## Testing Requirements

### Test-First Development (TDD)

**MANDATORY**: Tests must be written BEFORE implementation per Constitution Principle II.

1. Write failing tests first
2. Implement minimum code to pass tests
3. Refactor while keeping tests green

### Coverage Requirements

- **Overall coverage**: Minimum 80% (measured by `go test -coverprofile`)
- **Critical packages**: 100% coverage required
  - `pkg/veo3/` - API client and generation logic
  - `pkg/config/` - Configuration management
  - `pkg/operations/` - Operation lifecycle

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make coverage

# Run specific package
go test ./pkg/veo3/...

# Run specific test
go test -run TestGenerateVideo ./pkg/veo3/

# Run with race detector
go test -race ./...
```

### Test Organization

```
tests/
├── unit/              # Unit tests (fast, isolated)
│   ├── veo3/
│   ├── config/
│   └── operations/
├── integration/       # Integration tests (slower, end-to-end)
│   └── cli_test.go
└── fixtures/          # Test data
    ├── images/
    └── videos/
```

### Test Naming

- Test functions: `TestFunctionName`
- Table-driven tests: `TestFunctionName_WithScenario`
- Subtests: `t.Run("scenario", func(t *testing.T) {...})`

### Example Test

```go
func TestGenerateVideo_Success(t *testing.T) {
    tests := []struct {
        name    string
        request *GenerationRequest
        want    *Operation
        wantErr bool
    }{
        {
            name: "valid request",
            request: &GenerationRequest{
                Prompt: "test video",
                Model:  "veo-3.1",
            },
            want: &Operation{
                Status: StatusPending,
            },
            wantErr: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := GenerateVideo(context.Background(), tt.request)
            if (err != nil) != tt.wantErr {
                t.Errorf("GenerateVideo() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("GenerateVideo() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Pull Request Process

### Before Submitting

1. **Run tests**: `make test`
2. **Run linter**: `make lint`
3. **Check coverage**: `make coverage`
4. **Update documentation** if needed
5. **Update CHANGELOG.md** (if applicable)

### PR Description Template

```markdown
## Description
Brief description of changes

## Related Issues
Closes #<issue>
Updates #<other-issue>

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Coverage remains above 80%
- [ ] All tests pass

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] No new warnings from linter
- [ ] Git commits reference GitHub issues
```

### Review Process

1. At least one approving review required
2. All CI checks must pass
3. Coverage must not decrease
4. No merge conflicts

### Merging

- Use "Squash and merge" for feature branches
- Use "Rebase and merge" for hotfixes
- Delete branch after merging

## Code Review Guidelines

### For Authors

- Keep PRs focused and reasonably sized
- Provide context in PR description
- Respond to feedback promptly
- Update code based on review comments

### For Reviewers

- Review within 24-48 hours if possible
- Be constructive and specific
- Focus on code quality, not style preferences
- Approve when standards are met

## Documentation

### Code Documentation

- Add godoc comments for all exported types and functions
- Include examples in godoc when helpful
- Keep README.md up to date
- Update CLI help text when adding commands

### Commit Documentation

Every commit should explain:
- **What** changed
- **Why** it changed
- **How** it addresses the issue
- **Which issue(s)** it relates to

## Questions or Issues?

- **General questions**: Open a GitHub Discussion
- **Bug reports**: Create a GitHub Issue with bug template
- **Feature requests**: Create a GitHub Issue with feature template
- **Security issues**: Email maintainers directly (do not create public issue)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Attribution

Contributors will be acknowledged in the project README and release notes.

Thank you for contributing to Veo3 CLI! 🎬