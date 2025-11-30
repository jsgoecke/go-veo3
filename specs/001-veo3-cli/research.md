# Research: Veo3 CLI

**Date**: 2025-11-30  
**Purpose**: Document technology decisions, best practices, and architectural patterns for the Veo3 CLI implementation.

## Technology Stack Research

### Go Version Selection

**Decision**: Go 1.21+

**Rationale**:
- Latest stable version with generics support (introduced in 1.18)
- Improved error handling patterns
- Better performance and reduced memory footprint
- Strong cross-platform compilation support
- Excellent standard library for CLI applications

**Alternatives Considered**:
- **Go 1.20**: Would work but lacks some optimization improvements in 1.21
- **Go 1.22/1.23**: Too new, may have compatibility issues with some libraries

### CLI Framework

**Decision**: Cobra (`github.com/spf13/cobra`)

**Rationale**:
- Industry standard for Go CLIs (used by kubectl, Docker, GitHub CLI)
- Built-in support for subcommands, flags, and help generation
- Excellent integration with Viper for configuration
- Automatic shell completion generation
- Active maintenance and large community

**Alternatives Considered**:
- **urfave/cli**: Simpler but less feature-rich, harder to maintain large command hierarchies
- **Standard flag package**: Too low-level, would require significant boilerplate
- **Kong**: Good for smaller CLIs but less established ecosystem

### Configuration Management

**Decision**: Viper (`github.com/spf13/viper`)

**Rationale**:
- Seamless integration with Cobra
- Supports multiple configuration sources (file, env, flags) with precedence
- Built-in file watching for hot reloading (future enhancement)
- YAML, JSON, TOML support
- Environment variable binding with automatic uppercase conversion

**Alternatives Considered**:
- **godotenv**: Only handles .env files, not comprehensive enough
- **Custom implementation**: Reinventing the wheel, prone to bugs

### Google API Client

**Decision**: Official Google Go client libraries (`google.golang.org/api`)

**Rationale**:
- Official Google support with guaranteed compatibility
- Auto-generated from API discovery documents
- Built-in retry logic and error handling
- OAuth2 and API key authentication support
- Type-safe request/response structs

**Alternatives Considered**:
- **Raw HTTP requests**: Too low-level, error-prone, no type safety
- **Third-party wrappers**: Not officially maintained, may lag behind API updates

### Progress Indication

**Decision**: `github.com/schollz/progressbar/v3`

**Rationale**:
- Clean API with customizable display
- Thread-safe for concurrent operations
- Support for estimated time remaining
- Works well in CI/CD environments (can detect non-TTY)
- Minimal dependencies

**Alternatives Considered**:
- **cheggaaa/pb**: Less actively maintained
- **Custom spinner**: Would need significant development for all features

### Testing Framework

**Decision**: Built-in `testing` package + `github.com/stretchr/testify`

**Rationale**:
- Standard Go testing package is robust and well-integrated with tooling
- Testify adds convenient assertions and mocking capabilities
- `testify/mock` for interface mocking (API client mocking)
- `testify/assert` for readable test assertions
- `testify/suite` for test setup/teardown

**Alternatives Considered**:
- **Ginkgo/Gomega**: More verbose, BDD-style testing not needed for CLI
- **Testing package only**: Would require more verbose assertions

### YAML Processing

**Decision**: `gopkg.in/yaml.v3`

**Rationale**:
- Most mature and stable YAML library for Go
- Supports complex YAML features (anchors, references)
- Better error messages than v2
- Used by Kubernetes and other major projects

**Alternatives Considered**:
- **yaml.v2**: Older version, less features
- **JSON only**: Less human-readable for manifests and templates

## Architecture Patterns

### Package Organization

**Decision**: Standard Go project layout with domain-driven packages

**Pattern**:
```
cmd/       - Application entry points
pkg/       - Public library code (reusable)
internal/  - Private application code (not reusable)
tests/     - Integration tests and fixtures
```

**Rationale**:
- Follows Go community standards
- Clear separation of concerns
- Enables future library extraction
- Internal packages prevent external use of unstable code

### Error Handling

**Decision**: Wrapped errors with context using `fmt.Errorf` with `%w` verb

**Pattern**:
```go
if err != nil {
    return fmt.Errorf("failed to generate video: %w", err)
}
```

**Rationale**:
- Native Go 1.13+ error wrapping
- Preserves error chain for debugging
- Compatible with errors.Is() and errors.As()
- No external dependencies needed

### API Client Design

**Decision**: Single client struct with method-based operations

**Pattern**:
```go
type Client struct {
    service   *aiplatform.Service
    projectID string
    location  string
}

func (c *Client) GenerateVideo(ctx context.Context, req *GenerateRequest) (*Operation, error)
```

**Rationale**:
- Encapsulates authentication and configuration
- Easy to mock for testing
- Stateful connection management
- Context-aware for cancellation

### Operation Polling

**Decision**: Exponential backoff with configurable intervals

**Pattern**:
```go
func (p *Poller) Wait(ctx context.Context, opID string) (*Result, error) {
    ticker := time.NewTicker(p.interval)
    backoff := p.interval
    for {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-ticker.C:
            status, err := p.checkStatus(opID)
            if status.Done {
                return status.Result, nil
            }
            backoff = min(backoff*2, maxBackoff)
        }
    }
}
```

**Rationale**:
- Reduces API call frequency over time
- Respects context cancellation
- Configurable initial interval
- Prevents excessive API usage

## Best Practices Research

### CLI Design Principles

**Findings**:
1. **Progressive Disclosure**: Essential flags upfront, advanced options discoverable
2. **Fail Fast**: Validate inputs before expensive API calls
3. **Human-Friendly Defaults**: Sensible defaults reduce flag usage
4. **Machine-Readable Output**: `--json` flag for automation
5. **Consistency**: Similar flags across commands (`--output`, `--model`)

**Application**:
- Required arguments as positional parameters
- Optional flags with sensible defaults
- Global flags available to all commands
- Consistent flag names across command tree

### Configuration Hierarchy

**Decision**: Flags > Environment > Config File > Defaults

**Rationale**:
- Flags: Highest priority for one-off overrides
- Environment: Good for CI/CD and containerization
- Config File: User preferences
- Defaults: Fallback values

**Implementation**:
```go
apiKey := cobra.GetString("api-key")  // Flag
if apiKey == "" {
    apiKey = os.Getenv("GEMINI_API_KEY")  // Environment
}
if apiKey == "" {
    apiKey = config.Get("api_key")  // Config file
}
```

### File Operations

**Findings**:
- Use `os.CreateTemp()` for atomic writes (write temp, then rename)
- Check available disk space before downloads
- Use file locks for concurrent access to config
- Set proper permissions (0600 for config, 0644 for videos)

**Application**:
```go
// Atomic config write
tmp, err := os.CreateTemp(filepath.Dir(configPath), ".config.*.yaml")
yaml.NewEncoder(tmp).Encode(config)
os.Chmod(tmp.Name(), 0600)
os.Rename(tmp.Name(), configPath)
```

### Concurrent Batch Processing

**Decision**: Worker pool pattern with semaphore

**Pattern**:
```go
sem := make(chan struct{}, concurrency)
for _, job := range jobs {
    sem <- struct{}{}  // Acquire
    go func(j Job) {
        defer func() { <-sem }()  // Release
        process(j)
    }(job)
}
// Wait for all workers
for i := 0; i < cap(sem); i++ {
    sem <- struct{}{}
}
```

**Rationale**:
- Limits concurrent API calls to avoid rate limits
- Simple implementation without external dependencies
- Easy to test and reason about

## Security Considerations

### API Key Storage

**Decision**: Store in config file with 0600 permissions

**Security Measures**:
1. Never log API keys (even in verbose mode)
2. Mask in `config show` output (show last 4 chars only)
3. Warn if config file has incorrect permissions
4. Support environment variable for CI/CD (ephemeral)

**Implementation**:
```go
if stat.Mode().Perm() != 0600 {
    fmt.Fprintf(os.Stderr, "Warning: Config file has insecure permissions. Run: chmod 600 %s\n", path)
}
```

### Input Validation

**Decision**: Multi-layer validation (client-side and API contract)

**Validations**:
1. File existence and readability
2. File size limits (20MB for images)
3. File format detection (magic bytes, not just extension)
4. Prompt length (1024 tokens)
5. Parameter combinations (e.g., reference images require 8s + 16:9)

### Error Messages

**Decision**: User-friendly errors with actionable guidance

**Pattern**:
```go
if size > maxSize {
    return fmt.Errorf(
        "image file too large: %d MB (maximum: 20 MB)\n" +
        "Suggestion: Compress the image or reduce resolution",
        size/1024/1024,
    )
}
```

## Performance Optimization

### Memory Management

**Strategies**:
1. Stream large files instead of loading into memory
2. Use bufio for efficient I/O
3. Close resources promptly (defer close())
4. Reuse HTTP clients (connection pooling)

### Binary Size

**Optimization**:
```bash
go build -ldflags="-s -w" -o veo3 ./cmd/veo3
```
- `-s`: Omit symbol table
- `-w`: Omit DWARF debug info
- Reduces binary size by ~30%

### Cross-Compilation

**Build Targets**:
```bash
GOOS=linux GOARCH=amd64 go build
GOOS=darwin GOARCH=amd64 go build  # Intel Mac
GOOS=darwin GOARCH=arm64 go build  # Apple Silicon
GOOS=windows GOARCH=amd64 go build
```

## Testing Strategy

### Unit Tests

**Coverage**: Each package independently testable

**Mock Strategy**:
```go
type APIClient interface {
    GenerateVideo(context.Context, *Request) (*Operation, error)
}

type mockClient struct {
    mock.Mock
}
```

### Integration Tests

**Approach**:
1. Mock Google API responses using httptest
2. Test full command execution paths
3. Validate file I/O operations
4. Test configuration persistence

**Example**:
```go
func TestGenerateCommand(t *testing.T) {
    server := httptest.NewServer(mockAPIHandler())
    defer server.Close()
    
    cmd := generateCmd()
    cmd.SetArgs([]string{"test prompt", "--api-url", server.URL})
    err := cmd.Execute()
    assert.NoError(t, err)
}
```

### Test Coverage

**Target**: 80% minimum per constitution

**Measurement**:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**CI/CD Integration**:
```yaml
- name: Test with coverage
  run: |
    go test -coverprofile=coverage.out ./...
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if [ $(echo "$coverage < 80" | bc) -eq 1 ]; then
      echo "Coverage $coverage% is below 80%"
      exit 1
    fi
```

## Documentation Strategy

### Code Documentation

**Standards**:
- Package-level doc comment in doc.go files
- Exported functions/types documented with examples
- Complex logic explained with inline comments

**Example**:
```go
// Package veo3 provides a client for Google's Veo 3.1 video generation API.
//
// Usage:
//
//     client, err := veo3.NewClient(apiKey)
//     op, err := client.GenerateVideo(ctx, &veo3.GenerateRequest{
//         Prompt: "A cinematic shot...",
//     })
```

### User Documentation

**Components**:
1. README.md - Installation, quick start, examples
2. quickstart.md - Step-by-step first video generation
3. CLI help text - Embedded in commands
4. Man pages - Generated from Cobra documentation

## Dependency Management

**Strategy**:
- Pin major versions in go.mod
- Regular `go mod tidy` to prune unused deps
- Use `go mod verify` in CI/CD
- Review licenses (Apache 2.0, MIT, BSD preferred)

**Security**:
```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## Release Strategy

**Versioning**: Semantic Versioning (v1.2.3)

**Release Process**:
1. Tag release: `git tag v1.0.0`
2. Build binaries for all platforms
3. Generate changelog from conventional commits
4. Create GitHub release with binaries
5. Update package managers (Homebrew, Scoop)

**Automation**:
- GitHub Actions for CI/CD
- GoReleaser for multi-platform builds
- Automatic changelog generation

## Decisions Summary

| Category | Decision | Key Benefit |
|----------|----------|-------------|
| Language | Go 1.21+ | Cross-platform, performance, stdlib |
| CLI Framework | Cobra | Industry standard, rich features |
| Config | Viper | Multi-source, Cobra integration |
| API Client | google.golang.org/api | Official, type-safe, maintained |
| Progress | progressbar/v3 | Thread-safe, feature-rich |
| Testing | testing + testify | Standard + convenient assertions |
| YAML | gopkg.in/yaml.v3 | Mature, feature-complete |
| Error Handling | Wrapped errors (%w) | Native, chain-preserving |
| Architecture | Domain packages | Reusable, testable, standard |
| Security | File permissions + validation | Defense in depth |