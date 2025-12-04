# Veo3 CLI

[![CI/CD Pipeline](https://github.com/jasongoecke/go-veo3/workflows/CI/CD%20Pipeline/badge.svg)](https://github.com/jasongoecke/go-veo3/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/jasongoecke/go-veo3)](https://goreportcard.com/report/github.com/jasongoecke/go-veo3)
[![codecov](https://codecov.io/gh/jasongoecke/go-veo3/branch/main/graph/badge.svg)](https://codecov.io/gh/jasongoecke/go-veo3)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A command-line interface for Google's Veo 3.1 video generation API, enabling text-to-video, image-to-video, frame interpolation, reference-guided generation, and video extension capabilities.

## Features

- **Text-to-Video Generation**: Generate videos from text prompts with customizable duration, resolution, and aspect ratio
- **Image-to-Video Animation**: Animate static images into videos
- **Frame Interpolation**: Create smooth transitions between two images
- **Reference-Guided Generation**: Guide video generation with up to 3 reference images for style and content consistency
- **Video Extension**: Extend existing Veo-generated videos by up to 7 seconds (chainable)
- **Operation Management**: List, monitor, download, and cancel long-running video generation operations
- **Batch Processing**: Process multiple video generation requests from YAML manifests with concurrent execution
- **Prompt Templates**: Save and reuse prompt templates with variable substitution for consistent generations
- **Configuration Management**: Store API credentials and default settings
- **Documentation Generation**: Generate man pages and markdown documentation for all commands
- **Structured Logging**: Debug output with `--verbose` flag for troubleshooting
- **Multiple Output Formats**: Human-readable and JSON output for automation

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/jasongoecke/go-veo3.git
cd go-veo3

# Build the binary
make build

# Install to $GOPATH/bin
make install
```

### Using Go Install

```bash
go install github.com/jasongoecke/go-veo3/cmd/veo3@latest
```

## Prerequisites

- Go 1.21 or higher
- Google Gemini API key with Veo 3.1 access
- Supported platforms: macOS, Linux, Windows

## Quick Start

### 1. Set Up API Key

```bash
# Set via environment variable
export GEMINI_API_KEY="your-api-key-here"

# Or configure interactively
veo3 config init
```

### 2. Generate Your First Video

```bash
# Text-to-video generation
veo3 generate "A serene lake at sunset with mountains in the background"

# Image-to-video animation
veo3 animate image.jpg --prompt "Add gentle motion to the scene"

# Frame interpolation
veo3 interpolate start.png end.png
```

### 3. Check Operation Status

```bash
# List all operations
veo3 operations list

# Check specific operation
veo3 operations status operations/abc123

# Download completed video
veo3 operations download operations/abc123
```

## Configuration

### Configuration File

The CLI uses the XDG Base Directory specification for configuration:

- **Linux/macOS**: `~/.config/veo3/config.yaml`
- **Windows**: `%APPDATA%\veo3\config.yaml`

### Configuration Options

```yaml
api_key: "your-gemini-api-key"
default_model: "veo-3.1"
default_resolution: "720p"
default_aspect_ratio: "16:9"
default_duration: 6
output_directory: "./videos"
poll_interval_seconds: 10
```

### Configuration Commands

```bash
# Interactive setup
veo3 config init

# Set individual values
veo3 config set api_key "your-key"
veo3 config set default_resolution "1080p"

# Show current configuration
veo3 config show

# Reset to defaults
veo3 config reset
```

## Usage Examples

### Text-to-Video Generation

```bash
# Basic generation
veo3 generate "A cat playing piano"

# With custom settings
veo3 generate "A cat playing piano" \
  --resolution 1080p \
  --duration 8 \
  --aspect-ratio 16:9 \
  --model veo-3.1

# With negative prompt
veo3 generate "A beautiful garden" \
  --negative-prompt "people, cars, buildings"

# Async mode (start and return immediately)
veo3 generate "A sunset" --no-wait

# Save to specific directory
veo3 generate "A sunset" --output ./my-videos/
```

### Image-to-Video Animation

```bash
# Animate an image
veo3 animate photo.jpg

# With custom prompt
veo3 animate photo.jpg --prompt "Add subtle motion and life"

# Specify output settings
veo3 animate photo.jpg \
  --resolution 1080p \
  --duration 8 \
  --output ./animations/
```

### Frame Interpolation

```bash
# Interpolate between two images
veo3 interpolate start.png end.png

# With custom settings (note: interpolation has constraints)
veo3 interpolate start.png end.png \
  --prompt "Smooth transition" \
  --output ./interpolations/
```

### Reference-Guided Generation

```bash
# Use single reference image
veo3 generate "A product shot" \
  --reference brand-style.jpg

# Use multiple references (max 3)
veo3 generate "A new scene in the same style" \
  --reference style1.jpg \
  --reference style2.jpg \
  --reference style3.jpg
```

### Video Extension

```bash
# Extend a video
veo3 extend video.mp4 --prompt "Continue the scene"

# Chain extensions for longer videos
veo3 extend original.mp4 --prompt "Continue" --output part1.mp4
veo3 extend part1.mp4 --prompt "Keep going" --output part2.mp4
```

### Operation Management

```bash
# List all operations
veo3 operations list

# Filter by status
veo3 operations list --status running
veo3 operations list --status done

# Get detailed operation info
veo3 operations status operations/abc123

# Watch operation until completion
veo3 operations status operations/abc123 --watch

# Download completed video
veo3 operations download operations/abc123 --output ./videos/

# Cancel running operation
veo3 operations cancel operations/abc123

# Cancel all running operations
veo3 operations cancel --all
```

### Batch Processing

```bash
# Process multiple generations from manifest
veo3 batch process manifest.yaml

# Control concurrency
veo3 batch process manifest.yaml --concurrency 3

# Generate sample manifest template
veo3 batch template > my-manifest.yaml

# Retry failed jobs
veo3 batch retry results.json
```

### Prompt Templates

```bash
# Save a template with variables
veo3 templates save product-demo "{{product}} rotating on {{surface}} with {{lighting}}"

# List all templates
veo3 templates list

# View template details
veo3 templates get product-demo

# Generate using template
veo3 generate --template product-demo \
  --vars product="smartphone" \
  --vars surface="marble" \
  --vars lighting="soft studio lighting"

# Export templates for sharing
veo3 templates export my-templates.yaml

# Import templates
veo3 templates import shared-templates.yaml

# Delete a template
veo3 templates delete product-demo
```

### Documentation & Shell Completion

```bash
# Generate man pages
veo3 docs man --output ./docs/man

# Generate markdown documentation
veo3 docs markdown --output ./docs/cli

# Generate shell completion
veo3 completion bash > /etc/bash_completion.d/veo3
veo3 completion zsh > ~/.zsh/completion/_veo3

# Use completion in current session
source <(veo3 completion bash)
```

### Model Information

```bash
# List available models
veo3 models list

# Get model details
veo3 models info veo-3.1
```

## Command Reference

### Global Flags

```
--json          Output in JSON format for automation
--verbose       Enable verbose logging
--api-key       Override API key from config/environment
--config        Use custom config file path
```

### Commands

#### `veo3 generate`
Generate video from text prompt

**Flags:**
- `--prompt, -p`: Text prompt describing the video
- `--model, -m`: Model to use (default: veo-3.1)
- `--resolution, -r`: Video resolution (720p, 1080p)
- `--duration, -d`: Duration in seconds (4, 6, 8)
- `--aspect-ratio, -a`: Aspect ratio (16:9, 9:16)
- `--negative-prompt`: Elements to exclude
- `--reference`: Reference image(s) for guidance (repeatable, max 3)
- `--output`: Output directory
- `--filename`: Custom output filename
- `--no-wait`: Return immediately without waiting
- `--no-download`: Skip automatic download

#### `veo3 animate`
Animate a static image into a video

**Arguments:**
- `image-path`: Path to image file (JPEG, PNG, WebP)

**Flags:**
- `--prompt, -p`: Optional animation prompt
- `--resolution, -r`: Video resolution
- `--duration, -d`: Duration in seconds
- `--aspect-ratio, -a`: Aspect ratio
- `--output`: Output directory

#### `veo3 interpolate`
Create video by interpolating between two images

**Arguments:**
- `start-image`: First frame image
- `end-image`: Last frame image

**Flags:**
- `--prompt, -p`: Optional prompt
- `--output`: Output directory

**Constraints:**
- Duration fixed at 8 seconds
- Aspect ratio fixed at 16:9
- Images must have compatible dimensions

#### `veo3 extend`
Extend an existing Veo-generated video

**Arguments:**
- `video-path`: Path to video file

**Flags:**
- `--prompt, -p`: Extension prompt
- `--output`: Output directory
- `--filename`: Custom output filename

**Constraints:**
- Maximum input length: 141 seconds
- Extension length: up to 7 seconds
- Only works with Veo-generated videos

#### `veo3 operations`
Manage video generation operations

**Subcommands:**
- `list`: List all operations
- `status <operation-id>`: Check operation status
- `download <operation-id>`: Download completed video
- `cancel <operation-id>`: Cancel running operation

#### `veo3 models`
View available models and their capabilities

**Subcommands:**
- `list`: List all available models
- `info <model-name>`: Show detailed model information

#### `veo3 config`
Manage CLI configuration

**Subcommands:**
- `init`: Interactive configuration setup
- `set <key> <value>`: Set configuration value
- `show`: Display current configuration
- `reset`: Reset to defaults

#### `veo3 batch`
Batch processing for multiple generations

**Subcommands:**
- `process <manifest.yaml>`: Process batch manifest
- `template`: Generate sample manifest
- `retry <results.json>`: Retry failed jobs

**Flags:**
- `--concurrency, -c`: Number of concurrent jobs (default: 3)

#### `veo3 templates`
Manage prompt templates with variable substitution

**Subcommands:**
- `save <name>`: Save a new template
- `list`: List all saved templates
- `get <name>`: View template details
- `delete <name>`: Remove a template
- `export <file>`: Export templates to YAML
- `import <file>`: Import templates from YAML

**Template Variables:**
Use `{{variable}}` syntax in prompts for substitution

#### `veo3 docs`
Generate documentation

**Subcommands:**
- `man`: Generate man pages
- `markdown`: Generate markdown documentation

#### `veo3 completion`
Generate shell completion scripts

**Arguments:**
- Shell type: `bash`, `zsh`, `fish`, or `powershell`

## Output Formats

### Human-Readable (Default)

```
🎬 Generating video...
⏳ Running... (45%, elapsed: 1:23)
✓ Video generation complete!
📥 Downloaded to: ./video_abc123.mp4
```

### JSON Format

Use `--json` flag for machine-readable output:

```json
{
  "success": true,
  "data": {
    "operation_id": "operations/abc123",
    "status": "DONE",
    "video_uri": "gs://bucket/video.mp4",
    "local_path": "./video_abc123.mp4"
  }
}
```

## Error Handling

The CLI provides clear error messages with actionable guidance:

```bash
# Example: Invalid API key
Error: API authentication failed
Suggestion: Check your API key with 'veo3 config show'

# Example: File too large
Error: Image file exceeds 20MB limit
File: photo.jpg (25.3 MB)
Suggestion: Compress the image or use a smaller file

# Example: Invalid parameters
Error: Resolution "1080p" requires duration of 8 seconds
Current duration: 4 seconds
Suggestion: Use --duration 8 or change resolution to 720p
```

## Development

### Building from Source

```bash
# Clone and build
git clone https://github.com/jasongoecke/go-veo3.git
cd go-veo3
make build

# Run all quality checks (lint, security, test)
make check

# Run tests
make test

# Test GitHub Actions workflows locally with act
make act-install  # Install act (one-time setup)
make act-test     # Run complete CI/CD pipeline locally
make act-lint     # Run just the lint job
make act-unit     # Run just the unit tests job
make act-security # Run just the security scan job
make act-build    # Run just the build job

# Run with coverage (enforces 80% minimum)
make test-coverage

# Lint code
make lint

# Security scan
make security

# Clean build artifacts
make clean
```

### CI/CD Pipeline

This project includes a comprehensive CI/CD pipeline with:
- ✅ Automated linting with golangci-lint
- ✅ Security scanning with gosec
- ✅ Unit tests with 80% coverage requirement (enforced)
- ✅ Multi-platform builds (Linux, macOS, Windows on amd64 and arm64)
- ✅ Automated releases on git tags

**See [docs/CI-CD-SETUP.md](docs/CI-CD-SETUP.md) for complete CI/CD documentation.**

The build pipeline runs on every push and pull request, ensuring code quality before merge.

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
└── specs/              # Design specifications
```

### Running Tests

The project has separate test targets for different purposes:

```bash
# Unit tests only (fast, no API key needed) - Use this for development
make test

# Run all quality checks (lint + security + unit tests)
make check

# With coverage (enforces 80% minimum)
make coverage

# Integration tests (requires RUN_INTEGRATION_TESTS=1 and API key)
make test-integration

# All tests including integration
make test-all
```

**Important**: The CI/CD pipeline and `make build` only run unit tests. Integration tests require a real API key and are opt-in only via `RUN_INTEGRATION_TESTS=1`.

```bash
# Run specific test suites directly with Go
go test ./tests/unit/...           # Unit tests only
go test ./tests/integration/...    # Integration tests (may need API key)
```

## API Rate Limits

Be aware of Google's API rate limits and quotas:

- Default: 2 requests per minute
- Pro tier: Higher limits available
- Use `--no-wait` for async operations to manage multiple generations

## Troubleshooting

### API Key Issues

```bash
# Verify API key is set
veo3 config show

# Test with simple generation
veo3 generate "test" --no-download
```

### File Format Issues

- **Images**: Use JPEG, PNG, or WebP formats (max 20MB)
- **Videos**: Only Veo-generated MP4 files for extension
- Check file with: `file <filename>`

### Network Issues

- Verify internet connectivity
- Check firewall settings
- Use `--verbose` for detailed logs

### Operation Not Found

- Operations may expire after a certain time
- Use `veo3 operations list` to see available operations
- Download videos promptly after completion

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Run linter: `make lint`
6. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with Google's Gemini API and Veo 3.1 model
- Uses [Cobra](https://github.com/spf13/cobra) for CLI framework
- Uses [Viper](https://github.com/spf13/viper) for configuration management

## Documentation

### Available Guides

- **[SECURITY.md](docs/SECURITY.md)** - Security audit report and best practices
- **[PERFORMANCE.md](docs/PERFORMANCE.md)** - Performance profiling and optimization guide
- **[QUICKSTART-VALIDATION.md](docs/QUICKSTART-VALIDATION.md)** - Quickstart guide validation status
- **[CI-CD-SETUP.md](docs/CI-CD-SETUP.md)** - CI/CD pipeline setup and usage
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Contributing guidelines

### Generate Documentation

```bash
# Generate man pages for offline reference
veo3 docs man --output ./man

# Generate markdown command reference
veo3 docs markdown --output ./docs
```

## Support

- **Issues**: [GitHub Issues](https://github.com/jasongoecke/go-veo3/issues)
- **API Docs**: [Google Gemini API Documentation](https://ai.google.dev/api/docs)
- **Security**: Report vulnerabilities privately (do not open public issues)

## Roadmap

See [specs/001-veo3-cli/spec.md](specs/001-veo3-cli/spec.md) for planned features and development roadmap.

