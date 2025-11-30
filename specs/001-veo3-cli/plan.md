# Implementation Plan: Veo3 CLI

**Branch**: `001-veo3-cli` | **Date**: 2025-11-30 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `/specs/001-veo3-cli/spec.md`

## Summary

Build a comprehensive command-line utility in Go for Google's Veo 3.1 video generation API. The CLI will enable text-to-video generation, image-to-video animation, frame interpolation, video extension, reference image-guided generation, and complete operation management. Implementation follows Library-First principle with all core logic in testable packages and Test-First development with 80% minimum coverage. Uses official Google Go client libraries for Gemini API integration.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: 
- `google.golang.org/api/aiplatform/v1beta` - Google AI Platform API client
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `gopkg.in/yaml.v3` - YAML parsing for manifests/templates
- `github.com/schollz/progressbar/v3` - Progress display

**Storage**: File system for configuration (~/.config/veo3/config.yaml) and local video cache  
**Testing**: Go's built-in `testing` package, `github.com/stretchr/testify` for assertions  
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows)  
**Project Type**: Single project - CLI executable with library packages  
**Performance Goals**: 
- Configuration commands: <100ms response time
- API calls: <500ms overhead (excluding network)
- Memory footprint: <100MB during active generation
- Concurrent batch processing: 20+ jobs without degradation

**Constraints**:
- API rate limits: Subject to Google Gemini API quotas
- Generation latency: 11 seconds to 6 minutes (API-dependent)
- Memory: <100MB CLI footprint (SC-009)
- File size: 20MB max for input images (API constraint)
- Video retention: 2 days server-side (must download promptly)

**Scale/Scope**: 
- Single-user CLI tool (not multi-tenant)
- 10 main commands (generate, animate, interpolate, extend, operations, models, templates, batch, config, version)
- Support for concurrent batch operations (20+)
- Local config file, no database required

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

This feature must comply with all constitutional principles. Verify:

- [x] **Test-First (Principle II)**: Feature spec includes explicit test requirements and success criteria (NON-NEGOTIABLE)
  - ✅ All 10 user stories include acceptance scenarios with Given/When/Then
  - ✅ Spec includes 10 measurable success criteria
  - ✅ Testing framework identified (Go testing + testify)
  - ✅ 80% minimum coverage target will be enforced

- [x] **Quality Gates (Principle VI)**: 80% minimum test coverage target defined
  - ✅ Will use `go test -cover` with minimum 80% threshold
  - ✅ CI/CD will block merges below coverage threshold
  - ✅ `golangci-lint` for code quality and static analysis

- [x] **Library-First (Principle I)**: Complex logic extracted to standalone libraries with CLI interfaces (if applicable)
  - ✅ Core generation logic in `pkg/veo3` package
  - ✅ Configuration management in `pkg/config` package
  - ✅ Operation polling in `pkg/operations` package
  - ✅ CLI in `cmd/veo3` package (thin wrapper over libraries)
  - ✅ Each package independently testable

- [x] **Integration Testing (Principle III)**: Critical paths identified for integration tests (if applicable)
  - ✅ Google API integration tests (with mocked API responses)
  - ✅ End-to-end CLI command tests
  - ✅ File I/O integration tests (image upload, video download)
  - ✅ Configuration file persistence tests

- [x] **Observability (Principle IV)**: Structured logging and metrics plan included (if applicable)
  - ✅ JSON output mode for automation (`--json` flag)
  - ✅ Verbose logging mode (`--verbose` flag)
  - ✅ Progress indicators for long-running operations
  - ✅ Error messages with actionable guidance (SC-007)

- [x] **Documentation First (Principle V)**: Documentation created before implementation begins
  - ✅ Feature spec completed (spec.md)
  - ✅ Implementation plan (this file)
  - ✅ Will create quickstart.md in Phase 1
  - ✅ CLI help text embedded in code

- [x] **APIs as First-Class (Principle VII)**: API design meets public-facing standards (if applicable)
  - ✅ Package APIs designed for potential library reuse
  - ✅ Clean separation between CLI and library logic
  - ✅ Well-documented exported functions and types

- [x] **Scope-Based Auth (Principle VIII)**: Permission scopes defined for all capabilities (if applicable)
  - ⚠️ N/A - Single-user CLI tool, no multi-tenant authorization
  - ✅ API key authentication handled securely

- [x] **Feature Flags (Principle IX)**: Feature flag strategy defined for controlled rollout
  - ⚠️ Modified approach: Use semantic versioning for feature rollout
  - ✅ Beta features behind explicit opt-in flags
  - ✅ Experimental commands in separate subcommands

- [x] **Backend/Frontend Isolation (Principle X)**: Clear API boundaries maintained (if applicable)
  - ✅ CLI commands consume library packages (not direct API calls)
  - ✅ Google API interactions isolated in `pkg/veo3/client.go`
  - ✅ Clear separation: cmd/ (UI) → pkg/ (logic) → API

- [x] **Built for Compliance (Principle XI)**: Compliance implications documented (SOC2, ISO27001, HiTrust, HIPAA)
  - ✅ API key stored in config file with appropriate file permissions (0600)
  - ✅ No PII collected or transmitted beyond Google API requirements
  - ✅ Audit logging available via `--verbose` flag
  - ✅ Downloaded videos stored locally with user control

- [x] **Semantic Versioning (Principle XII)**: Version impact assessed (MAJOR/MINOR/PATCH)
  - ✅ Initial release: v0.1.0 (development)
  - ✅ First stable: v1.0.0 (all P0-P2 user stories)
  - ✅ Conventional commits enforced in commit messages

**Complexity Justification Required**: None - all principles satisfied or reasonably adapted for CLI context.

## Project Structure

### Documentation (this feature)

```text
specs/001-veo3-cli/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── api-spec.yaml    # Google Veo API contract
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
# Single project structure - Go CLI application

cmd/
└── veo3/
    └── main.go          # Entry point, cobra root command setup

pkg/
├── veo3/
│   ├── client.go        # Google API client wrapper
│   ├── generate.go      # Text-to-video generation
│   ├── animate.go       # Image-to-video generation
│   ├── interpolate.go   # Frame interpolation
│   ├── extend.go        # Video extension
│   ├── reference.go     # Reference image handling
│   └── models.go        # Model information and validation
├── operations/
│   ├── manager.go       # Operation lifecycle management
│   ├── poller.go        # Status polling logic
│   └── downloader.go    # Video download logic
├── config/
│   ├── config.go        # Configuration struct and loading
│   ├── manager.go       # Config file management
│   └── defaults.go      # Default values
├── templates/
│   ├── manager.go       # Template storage and retrieval
│   └── parser.go        # Variable substitution
├── batch/
│   ├── processor.go     # Batch job execution
│   └── manifest.go      # YAML manifest parsing
└── cli/
    ├── generate.go      # 'generate' command
    ├── animate.go       # 'animate' command
    ├── interpolate.go   # 'interpolate' command
    ├── extend.go        # 'extend' command
    ├── operations.go    # 'operations' command group
    ├── models.go        # 'models' command group
    ├── templates.go     # 'templates' command group
    ├── batch.go         # 'batch' command group
    └── config.go        # 'config' command group

internal/
├── validation/
│   ├── files.go         # Image/video validation
│   └── params.go        # Parameter validation
└── format/
    ├── output.go        # Human-readable output formatting
    └── json.go          # JSON output formatting

tests/
├── integration/
│   ├── api_test.go      # Google API integration tests (mocked)
│   ├── cli_test.go      # End-to-end CLI tests
│   └── config_test.go   # Config file integration tests
└── fixtures/
    ├── images/          # Sample images for testing
    ├── videos/          # Sample videos for testing
    └── manifests/       # Sample batch manifests

go.mod                   # Go module definition
go.sum                   # Dependency checksums
.golangci.yml            # Linter configuration
Makefile                 # Build and test targets
README.md                # User-facing documentation
```

**Structure Decision**: Single project structure is appropriate for a CLI tool. All logic is organized into independently testable packages under `pkg/`, with CLI command implementations in `pkg/cli/` and the main entry point in `cmd/veo3/`. This follows Go community best practices and enables easy library extraction if needed in the future.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations to justify. All constitutional principles are satisfied or appropriately adapted for CLI tool context.
