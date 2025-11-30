# Feature Specification: Veo3 CLI

**Feature Branch**: `001-veo3-cli`  
**Created**: 2025-11-30  
**Status**: Draft  
**Input**: User description: "A comprehensive command-line utility for Google's Veo 3.1 video generation API, enabling developers and creators to generate AI videos with native audio, extend existing videos, use reference images, and perform frame interpolation from the terminal."

## Overview

A comprehensive command-line utility for Google's Veo 3.1 video generation API, enabling developers and creators to generate AI videos with native audio, extend existing videos, use reference images, and perform frame interpolation—all from the terminal.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Text-to-Video Generation (Priority: P0)

Users can generate videos from text prompts with optional audio cues, dialogue, and ambient sound descriptions.

**Why this priority**: Core functionality representing the primary use case for video generation. This is the foundation upon which all other features build.

**Independent Test**: Can be fully tested by providing a text prompt and verifying a video file is generated with the expected properties (duration, resolution, aspect ratio). Delivers immediate value as a standalone video creation tool.

**Acceptance Scenarios**:

1. **Given** a user has a valid API key configured, **When** they execute `veo3 generate "A cinematic shot of a majestic lion in the savannah"`, **Then** a video file is generated and saved with default settings (720p, 16:9, 8 seconds)
2. **Given** a user wants custom video properties, **When** they execute with options `--resolution 1080p --duration 8 --aspect-ratio 16:9`, **Then** the generated video matches the specified properties
3. **Given** a user wants to exclude certain elements, **When** they provide `--negative-prompt "cartoon, drawing, low quality"`, **Then** the generated video avoids those elements
4. **Given** a video is being generated, **When** the user waits, **Then** progress updates display elapsed time every 10 seconds
5. **Given** generation completes successfully, **When** the video is ready, **Then** the file is automatically downloaded to the specified output path with success confirmation

---

### User Story 2 - Image-to-Video Generation (Priority: P0)

Users can animate a static image into a video using an input image as the first frame.

**Why this priority**: Essential feature for bringing existing artwork, photos, and AI-generated images to life. Critical for product showcases and creative workflows.

**Independent Test**: Can be tested by providing an image file and verifying the generated video begins with that exact image and animates from there. Delivers value as a standalone image animation tool.

**Acceptance Scenarios**:

1. **Given** a user has an image file (PNG, JPEG, or WebP), **When** they execute `veo3 animate kitten.png --prompt "The kitten wakes up and stretches"`, **Then** a video is generated starting with the provided image
2. **Given** an image exceeds 20MB, **When** the user attempts animation, **Then** an error message displays the size limit and suggests compression
3. **Given** a user provides an unsupported format, **When** they attempt animation, **Then** an error lists supported formats (PNG, JPEG, WebP)
4. **Given** a user wants animation without additional guidance, **When** they execute `veo3 animate image.jpg` without a prompt, **Then** the video animates naturally from the image

---

### User Story 3 - Frame Interpolation (Priority: P1)

Users can generate videos by specifying both the first and last frames, with Veo interpolating the motion between them.

**Why this priority**: Enables precise creative control over video composition and transitions. Important for storyboarding and planned sequences.

**Independent Test**: Can be tested by providing two different images and verifying the generated video smoothly transitions between them over 8 seconds. Delivers value as a transition/morphing tool.

**Acceptance Scenarios**:

1. **Given** a user has first and last frame images, **When** they execute `veo3 interpolate start.png end.png --prompt "Smooth fade transition"`, **Then** an 8-second video is generated interpolating between the frames
2. **Given** the two images have different dimensions, **When** the user attempts interpolation, **Then** an error explains the images must be compatible
3. **Given** a user wants minimal interpolation, **When** they omit the prompt, **Then** the video smoothly transitions between frames using default interpolation

---

### User Story 4 - Reference Image-Guided Generation (Priority: P1)

Users can provide up to three reference images to guide the content, style, and subject appearance in generated videos.

**Why this priority**: Critical for brand consistency, character preservation, and product showcases. Enables professional and commercial use cases.

**Independent Test**: Can be tested by providing reference images and verifying the generated video preserves the visual characteristics (style, subjects, colors) from those references. Delivers value as a brand-consistent content creation tool.

**Acceptance Scenarios**:

1. **Given** a user has a reference image, **When** they execute `veo3 generate "A woman walks through a garden" --reference dress.png`, **Then** the generated video features the dress design from the reference
2. **Given** a user provides three reference images, **When** they execute with `--reference char.png --reference costume.png --reference prop.png`, **Then** all three visual elements appear in the generated video
3. **Given** a user tries to provide four reference images, **When** they execute, **Then** an error explains the maximum is three references
4. **Given** reference images are used, **When** generation completes, **Then** the output is 8 seconds at 16:9 aspect ratio (API requirement)

---

### User Story 5 - Video Extension (Priority: P1)

Users can extend previously generated Veo videos by up to 7 seconds, chainable up to 20 times for videos up to 148 seconds total.

**Why this priority**: Enables creation of longer narrative content from initial generations. Essential for storytelling and extended content creation.

**Independent Test**: Can be tested by generating a short video, then extending it with a new prompt, and verifying the extension seamlessly continues from the original. Delivers value as a long-form content creation tool.

**Acceptance Scenarios**:

1. **Given** a user has a Veo-generated video, **When** they execute `veo3 extend original.mp4 --prompt "The scene continues with new action"`, **Then** a new video is generated combining the original and extension
2. **Given** a video exceeds 141 seconds, **When** the user attempts extension, **Then** an error explains the maximum input length
3. **Given** a video is not Veo-generated, **When** the user attempts extension, **Then** an error explains only Veo videos can be extended
4. **Given** a user wants simple continuation, **When** they execute `veo3 extend video.mp4` without a prompt, **Then** the video extends naturally

---

### User Story 6 - Operation Management (Priority: P2)

Users can manage long-running video generation operations, check status, and retrieve completed videos.

**Why this priority**: Essential for handling asynchronous operations and recovering from interruptions. Important for production workflows.

**Independent Test**: Can be tested by starting a generation, listing operations, checking status, and downloading the result. Delivers value as an operation tracking and recovery tool.

**Acceptance Scenarios**:

1. **Given** a user has active generations, **When** they execute `veo3 operations list`, **Then** all pending and recent operations display with status
2. **Given** a user has an operation ID, **When** they execute `veo3 operations status operations/abc123`, **Then** detailed status information displays
3. **Given** an operation completes, **When** the user executes `veo3 operations download operations/abc123`, **Then** the video downloads to the specified path
4. **Given** a user wants to cancel, **When** they execute `veo3 operations cancel operations/abc123`, **Then** the operation is cancelled and confirmed
5. **Given** a user wants async generation, **When** they add `--async` flag, **Then** the operation ID returns immediately without blocking

---

### User Story 7 - Model Selection and Information (Priority: P2)

Users can list available models and their capabilities, and select specific model versions.

**Why this priority**: Enables informed model selection based on speed, quality, and feature requirements. Important for optimizing cost and quality.

**Independent Test**: Can be tested by listing models, viewing details, and setting a default model in configuration. Delivers value as a model discovery and selection tool.

**Acceptance Scenarios**:

1. **Given** a user wants to see available models, **When** they execute `veo3 models list`, **Then** all Veo models display with key capabilities
2. **Given** a user wants detailed information, **When** they execute `veo3 models info veo-3.1-generate-preview`, **Then** full model specifications display including limitations
3. **Given** a user wants to set a default, **When** they execute `veo3 config set default-model veo-3.1-fast-generate-preview`, **Then** future commands use that model by default

---

### User Story 8 - Configuration Management (Priority: P2)

Users can configure API credentials, default settings, and preferences.

**Why this priority**: Reduces repetitive flag usage and enables secure credential storage. Important for usability and security.

**Independent Test**: Can be tested by setting configuration values, viewing them, and verifying commands use those defaults. Delivers value as a configuration and credential management tool.

**Acceptance Scenarios**:

1. **Given** a new user, **When** they execute `veo3 config init`, **Then** an interactive setup guides them through configuration
2. **Given** a user has an API key, **When** they execute `veo3 config set api-key $GEMINI_API_KEY`, **Then** the key is securely stored
3. **Given** a user sets defaults, **When** they execute commands without options, **Then** the configured defaults are used
4. **Given** a user wants to see settings, **When** they execute `veo3 config show`, **Then** current configuration displays (with sensitive data masked)

---

### User Story 9 - Batch Processing (Priority: P3)

Users can process multiple video generation requests from a manifest file.

**Why this priority**: Enables automated workflows and bulk content creation. Valuable for production use but not essential for initial adoption.

**Independent Test**: Can be tested by creating a manifest with multiple jobs and verifying all are processed with appropriate progress tracking. Delivers value as a bulk generation automation tool.

**Acceptance Scenarios**:

1. **Given** a user has a YAML manifest, **When** they execute `veo3 batch process manifest.yaml`, **Then** all jobs process sequentially with progress updates
2. **Given** a user wants parallel processing, **When** they add `--concurrency 3`, **Then** up to 3 jobs run simultaneously
3. **Given** some jobs fail, **When** batch completes, **Then** a summary report shows successes and failures
4. **Given** a previous batch had failures, **When** the user executes `veo3 batch retry results.json`, **Then** only failed jobs are retried

---

### User Story 10 - Prompt Templates (Priority: P3)

Users can save and reuse prompt templates with variable substitution.

**Why this priority**: Accelerates workflow for users generating similar content types. Nice-to-have but not essential for core functionality.

**Independent Test**: Can be tested by saving a template with variables, using it with different values, and verifying output matches expectations. Delivers value as a prompt management and reuse tool.

**Acceptance Scenarios**:

1. **Given** a user has a reusable prompt pattern, **When** they execute `veo3 templates save product-showcase --prompt "A {{product}} rotates on {{background}}"`, **Then** the template is saved with placeholders
2. **Given** a saved template exists, **When** they execute `veo3 generate --template product-showcase --var product="watch" --var background="marble"`, **Then** a video generates with substituted values
3. **Given** a user wants to share templates, **When** they execute `veo3 templates export`, **Then** templates export to a shareable YAML file

---

### Edge Cases

- **What happens when the safety filter blocks content?** Display the specific reason (violence, adult content, etc.) and suggest prompt revisions to pass the filter
- **What happens when rate limits are exceeded?** Display the retry-after time and offer to queue the request for automatic retry
- **What happens when an operation times out?** Offer to retry with the same parameters or check operation status
- **What happens when the API key is invalid?** Display clear authentication error and guide user to configuration
- **What happens when output file already exists?** Prompt for overwrite confirmation or suggest automatic filename with suffix
- **What happens when insufficient disk space?** Detect before download and display required vs available space
- **What happens when network connection is lost during generation?** Preserve operation ID and allow resuming via operations command
- **What happens when input files are corrupt or invalid?** Validate before uploading and display specific format/corruption issues

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST authenticate requests using API key from configuration, environment variable, or command flag
- **FR-002**: System MUST support text-to-video generation with prompts up to 1024 tokens
- **FR-003**: System MUST support image-to-video generation using PNG, JPEG, or WebP images up to 20MB
- **FR-004**: System MUST support frame interpolation between two images to generate 8-second videos
- **FR-005**: System MUST support reference image-guided generation with up to 3 reference images
- **FR-006**: System MUST support extending Veo-generated videos by up to 7 seconds
- **FR-007**: System MUST validate all input files (images, videos) for format, size, and compatibility before upload
- **FR-008**: System MUST poll operation status at configurable intervals (default 10 seconds)
- **FR-009**: System MUST display progress updates during generation showing elapsed time
- **FR-010**: System MUST download completed videos automatically to specified output path
- **FR-011**: System MUST support model selection from available Veo models (3.1, 3.0, 2.0 variants)
- **FR-012**: System MUST allow configuration of aspect ratio (16:9 or 9:16), resolution (720p or 1080p), and duration (4, 6, or 8 seconds)
- **FR-013**: System MUST persist configuration in standard config file location (~/.config/veo3/config.yaml)
- **FR-014**: System MUST list, check status, download, and cancel operations via operations management commands
- **FR-015**: System MUST support batch processing from YAML manifest files with parallel execution
- **FR-016**: System MUST support prompt templates with variable substitution
- **FR-017**: System MUST handle API errors gracefully with clear error messages and suggested actions
- **FR-018**: System MUST validate model-specific constraints (e.g., reference images only with Veo 3.1, 1080p only for 8-second videos)
- **FR-019**: System MUST support both blocking (wait for completion) and non-blocking (async) generation modes
- **FR-020**: System MUST output results in human-readable format by default, with optional JSON output via --json flag

### Key Entities

- **Generation Request**: Represents a video generation job with parameters (prompt, model, options, input files)
- **Operation**: Long-running async operation tracked by Google API with ID, status, progress, and result
- **Model**: Veo model variant with capabilities, constraints, and compatibility rules
- **Configuration**: User settings including API key, default values, output directory, and preferences
- **Prompt Template**: Reusable prompt pattern with variable placeholders for substitution
- **Batch Manifest**: Collection of generation requests with metadata for bulk processing
- **Generated Video**: Output video file with metadata (model used, generation parameters, file size, duration)

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can generate their first video within 2 minutes of installing the CLI (including configuration)
- **SC-002**: 95% of text-to-video generations complete successfully within 6 minutes under normal API load
- **SC-003**: Users can extend videos with seamless transitions (no visible cut) between original and extended segments
- **SC-004**: Reference image-guided generation preserves subject appearance in 90% of cases as judged by users
- **SC-005**: Batch processing handles 20+ concurrent jobs without CLI crashes or hung operations
- **SC-006**: Configuration commands complete in under 1 second for immediate user feedback
- **SC-007**: Error messages provide actionable next steps in 100% of common error scenarios (invalid auth, rate limits, file issues)
- **SC-008**: Users can successfully resume interrupted generations using operation IDs in 100% of cases
- **SC-009**: CLI memory footprint stays under 100MB during active generation for resource efficiency
- **SC-010**: Command help text and examples enable 80% of users to complete tasks without external documentation

---

## Technical Specifications

### API Integration

**Base URL**: `https://generativelanguage.googleapis.com/v1beta`

**Authentication**: API key via `x-goog-api-key` header

**Key Endpoints**:
- `POST /models/{model}:predictLongRunning` - Start generation
- `GET /{operation_name}` - Poll operation status
- `GET {video_uri}` - Download generated video

### Parameters Reference

| Parameter | Type | Values | Default | Notes |
|-----------|------|--------|---------|-------|
| `prompt` | string | max 1024 tokens | required | Text description |
| `negativePrompt` | string | - | none | Exclusions |
| `image` | Image | PNG/JPEG/WebP ≤20MB | none | First frame |
| `lastFrame` | Image | PNG/JPEG/WebP ≤20MB | none | For interpolation |
| `referenceImages` | Image[] | max 3 images | none | Content guidance |
| `video` | Video | Veo-generated ≤141s | none | For extension |
| `aspectRatio` | string | "16:9", "9:16" | "16:9" | Output ratio |
| `resolution` | string | "720p", "1080p" | "720p" | Output quality |
| `durationSeconds` | string | "4", "6", "8" | "8" | Video length |
| `personGeneration` | string | "allow_all", "allow_adult", "dont_allow" | varies | Safety filter |
| `seed` | integer | - | random | Reproducibility hint |

### Error Handling

| Error Code | Description | CLI Behavior |
|------------|-------------|--------------|
| Safety filter triggered | Content blocked | Display reason, suggest prompt revision |
| Invalid image format | Unsupported input | List supported formats |
| Rate limit exceeded | Too many requests | Display retry-after, offer queue |
| Operation timeout | Generation failed | Offer retry with same parameters |
| Invalid parameters | Bad request | Show valid options |

### Output Format

**Default Output**:
```
Starting video generation...
Model: veo-3.1-generate-preview
Prompt: "A cinematic shot of a majestic lion..."

⏳ Generating video... (elapsed: 0:00:10)
⏳ Generating video... (elapsed: 0:00:20)
⏳ Generating video... (elapsed: 0:00:35)

✅ Video generated successfully!
📁 Saved to: lion_savannah.mp4
📊 Duration: 8 seconds | Resolution: 1080p | Size: 12.4 MB
```

**JSON Output** (with `--json` flag):
```json
{
  "success": true,
  "operation_id": "operations/abc123",
  "output_file": "lion_savannah.mp4",
  "metadata": {
    "model": "veo-3.1-generate-preview",
    "duration_seconds": 8,
    "resolution": "1080p",
    "aspect_ratio": "16:9",
    "file_size_bytes": 13002547,
    "generation_time_seconds": 35
  }
}
```

---

## Command Reference

```
veo3 - Google Veo 3.1 Video Generation CLI

USAGE:
    veo3 <COMMAND> [OPTIONS]

COMMANDS:
    generate      Generate video from text prompt
    animate       Generate video from image (image-to-video)
    interpolate   Generate video between first and last frames
    extend        Extend an existing Veo-generated video
    operations    Manage generation operations
    models        List and inspect available models
    templates     Manage prompt templates
    batch         Process multiple generations from manifest
    config        Manage CLI configuration

GLOBAL OPTIONS:
    --api-key <KEY>     Override API key
    --model <MODEL>     Override default model
    --json              Output in JSON format
    --quiet             Suppress progress output
    --verbose           Enable debug logging
    --version           Show version information
    --help              Show help

ENVIRONMENT VARIABLES:
    GEMINI_API_KEY      API key for authentication
    VEO3_CONFIG_PATH    Custom config file location
    VEO3_OUTPUT_DIR     Default output directory
```

---

## Constraints & Limitations

1. **Video Retention**: Generated videos stored server-side for 2 days only; must download promptly
2. **Extension Input**: Only Veo-generated videos can be extended (not arbitrary uploads)
3. **Reference Images**: Only available with Veo 3.1 models, requires 8s duration and 16:9 aspect
4. **1080p Constraints**: Only available for 8-second duration videos
5. **Regional Restrictions**: EU/UK/CH/MENA have limited `personGeneration` options
6. **Rate Limits**: Subject to Google API rate limits (specific limits not documented in spec)
7. **Latency**: Generation takes 11 seconds to 6 minutes depending on load
8. **Watermarking**: All outputs include SynthID watermark for AI content identification
9. **File Size**: Input images limited to 20MB maximum
10. **Token Limit**: Prompts limited to 1024 tokens maximum

---

## Model Reference

| Model | Audio | Resolution | Duration | Extension | Reference Images |
|-------|-------|------------|----------|-----------|------------------|
| `veo-3.1-generate-preview` | ✓ | 720p, 1080p | 4, 6, 8s | ✓ | ✓ (up to 3) |
| `veo-3.1-fast-generate-preview` | ✓ | 720p, 1080p | 4, 6, 8s | ✓ | ✓ (up to 3) |
| `veo-3-generate-preview` | ✓ | 720p, 1080p* | 4, 6, 8s | ✗ | ✗ |
| `veo-3-fast-generate-preview` | ✓ | 720p, 1080p* | 4, 6, 8s | ✗ | ✗ |
| `veo-2.0-generate-001` | ✗ | 720p | 5, 6, 8s | ✗ | ✗ |

*1080p only for 16:9 aspect ratio

---

## Assumptions

- Users have basic command-line familiarity and can follow installation instructions
- Users have obtained a valid Google Gemini API key before using the CLI
- Users understand video generation is asynchronous and may take several minutes
- Users have sufficient disk space for downloaded videos (typically 10-30MB per video)
- Users have stable internet connection for uploading inputs and downloading outputs
- Configuration file location follows XDG Base Directory specification on Linux/macOS
- Batch manifest format uses YAML for human readability and ease of editing
- Prompt templates use Mustache-style variable syntax ({{variable}}) for familiarity
- Default polling interval (10 seconds) balances responsiveness with API efficiency
- Users prefer immediate feedback and clear progress indicators during generation

---

## Future Considerations

- **Watch Mode**: Monitor directory for new images and auto-generate videos
- **Pipeline Integration**: stdin/stdout support for Unix pipelines
- **Plugin System**: Custom post-processing hooks (ffmpeg integration for format conversion, trimming)
- **Cloud Storage**: Direct upload to GCS/S3/Azure Blob without local download
- **Webhook Support**: Callback URL for completion notifications in long-running batches
- **GUI Companion**: Optional Electron/Tauri GUI that wraps the CLI for visual workflow
- **Smart Retry**: Automatic retry with exponential backoff for transient failures
- **Cost Tracking**: Track API usage and estimated costs per generation
- **Quality Presets**: Named presets (draft/standard/premium) for common option combinations
- **Video Metadata**: Embed generation parameters in video metadata for reproducibility
