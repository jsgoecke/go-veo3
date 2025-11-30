# Data Model: Veo3 CLI

**Date**: 2025-11-30  
**Purpose**: Define core data structures and their relationships for the Veo3 CLI implementation.

## Core Entities

### GenerationRequest

Represents a video generation job with all necessary parameters.

**Fields**:
- `Prompt` (string, required): Text description for video generation (max 1024 tokens)
- `NegativePrompt` (string, optional): Elements to exclude from generation
- `Model` (string, required): Veo model identifier (e.g., "veo-3.1-generate-preview")
- `AspectRatio` (string, required): Output aspect ratio ("16:9" or "9:16")
- `Resolution` (string, required): Output resolution ("720p" or "1080p")
- `DurationSeconds` (int, required): Video duration (4, 6, or 8 seconds)
- `Seed` (int, optional): Random seed for reproducibility
- `PersonGeneration` (string, optional): Safety filter level ("allow_all", "allow_adult", "dont_allow")

**Validation Rules**:
- Prompt length: 1-1024 tokens
- AspectRatio: Must be "16:9" or "9:16"
- Resolution: "1080p" only valid with 8-second duration
- DurationSeconds: Must be 4, 6, or 8
- Model: Must be a valid Veo model from ModelRegistry

**Go Type**:
```go
type GenerationRequest struct {
    Prompt            string   `json:"prompt" yaml:"prompt"`
    NegativePrompt    string   `json:"negative_prompt,omitempty" yaml:"negative_prompt,omitempty"`
    Model             string   `json:"model" yaml:"model"`
    AspectRatio       string   `json:"aspect_ratio" yaml:"aspect_ratio"`
    Resolution        string   `json:"resolution" yaml:"resolution"`
    DurationSeconds   int      `json:"duration_seconds" yaml:"duration_seconds"`
    Seed              *int     `json:"seed,omitempty" yaml:"seed,omitempty"`
    PersonGeneration  string   `json:"person_generation,omitempty" yaml:"person_generation,omitempty"`
}
```

---

### ImageRequest

Extends GenerationRequest for image-to-video generation.

**Additional Fields**:
- `ImagePath` (string, required): Local path to first frame image
- `ImageData` ([]byte, internal): Image binary data after validation

**Validation Rules**:
- Image file must exist and be readable
- File size: ≤20MB
- Format: PNG, JPEG, or WebP (validated via magic bytes)
- Dimensions: Compatible with chosen aspect ratio

**Go Type**:
```go
type ImageRequest struct {
    GenerationRequest
    ImagePath string `json:"image_path" yaml:"image_path"`
}
```

---

### InterpolationRequest

Extends GenerationRequest for frame interpolation.

**Additional Fields**:
- `FirstFramePath` (string, required): Path to first frame image
- `LastFramePath` (string, required): Path to last frame image
- `FirstFrameData` ([]byte, internal): First frame binary data
- `LastFrameData` ([]byte, internal): Last frame binary data

**Validation Rules**:
- Both images must pass ImageRequest validation
- Images must have compatible dimensions
- Duration fixed at 8 seconds (API requirement)
- Aspect ratio fixed at 16:9 (API requirement)

**Go Type**:
```go
type InterpolationRequest struct {
    GenerationRequest
    FirstFramePath string `json:"first_frame_path" yaml:"first_frame_path"`
    LastFramePath  string `json:"last_frame_path" yaml:"last_frame_path"`
}
```

---

### ReferenceImageRequest

Extends GenerationRequest with reference images for guided generation.

**Additional Fields**:
- `ReferenceImagePaths` ([]string, required): Paths to 1-3 reference images
- `ReferenceImageData` ([][]byte, internal): Reference image binary data

**Validation Rules**:
- 1-3 reference images required
- Each image must pass ImageRequest validation
- Duration fixed at 8 seconds (API requirement)
- Aspect ratio fixed at 16:9 (API requirement)
- Only available with Veo 3.1 models

**Go Type**:
```go
type ReferenceImageRequest struct {
    GenerationRequest
    ReferenceImagePaths []string `json:"reference_image_paths" yaml:"reference_image_paths"`
}
```

---

### ExtensionRequest

Extends GenerationRequest for video extension.

**Additional Fields**:
- `VideoPath` (string, required): Path to Veo-generated video to extend
- `VideoData` ([]byte, internal): Video binary data after validation
- `ExtensionPrompt` (string, optional): Prompt for extension content (overrides base Prompt)

**Validation Rules**:
- Video file must exist and be readable
- Video must be Veo-generated (validated via metadata or server response)
- Video duration: ≤141 seconds
- Video resolution: 720p
- Aspect ratio must match original generation
- Extension adds ≤7 seconds

**Go Type**:
```go
type ExtensionRequest struct {
    VideoPath        string `json:"video_path" yaml:"video_path"`
    ExtensionPrompt  string `json:"extension_prompt,omitempty" yaml:"extension_prompt,omitempty"`
    Model            string `json:"model" yaml:"model"`
}
```

---

### Operation

Represents a long-running async operation tracked by Google API.

**Fields**:
- `ID` (string, required): Operation identifier (e.g., "operations/abc123")
- `Status` (OperationStatus, required): Current operation state
- `Progress` (float64, optional): Completion percentage (0.0-1.0)
- `StartTime` (time.Time, required): When operation started
- `EndTime` (time.Time, optional): When operation completed/failed
- `VideoURI` (string, optional): Download URL for completed video
- `Error` (OperationError, optional): Error details if failed
- `Metadata` (map[string]interface{}, optional): Additional operation metadata

**State Transitions**:
```
PENDING → RUNNING → DONE (with VideoURI)
                  ↘ FAILED (with Error)
                  ↘ CANCELLED
```

**Go Type**:
```go
type Operation struct {
    ID        string                 `json:"id"`
    Status    OperationStatus        `json:"status"`
    Progress  float64                `json:"progress,omitempty"`
    StartTime time.Time              `json:"start_time"`
    EndTime   *time.Time             `json:"end_time,omitempty"`
    VideoURI  string                 `json:"video_uri,omitempty"`
    Error     *OperationError        `json:"error,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type OperationStatus string

const (
    StatusPending   OperationStatus = "PENDING"
    StatusRunning   OperationStatus = "RUNNING"
    StatusDone      OperationStatus = "DONE"
    StatusFailed    OperationStatus = "FAILED"
    StatusCancelled OperationStatus = "CANCELLED"
)
```

---

### OperationError

Error details for failed operations.

**Fields**:
- `Code` (string, required): Error code (e.g., "SAFETY_FILTER", "RATE_LIMIT")
- `Message` (string, required): Human-readable error description
- `Details` (map[string]interface{}, optional): Additional error context
- `Suggestion` (string, optional): Actionable guidance for user

**Go Type**:
```go
type OperationError struct {
    Code       string                 `json:"code"`
    Message    string                 `json:"message"`
    Details    map[string]interface{} `json:"details,omitempty"`
    Suggestion string                 `json:"suggestion,omitempty"`
}
```

---

### Model

Veo model variant with capabilities and constraints.

**Fields**:
- `ID` (string, required): Model identifier
- `Name` (string, required): Display name
- `Capabilities` (ModelCapabilities, required): Supported features
- `Constraints` (ModelConstraints, required): Limitations and requirements
- `Tier` (string, required): Performance/cost tier ("standard", "fast")
- `Version` (string, required): Model version (e.g., "3.1", "3.0", "2.0")

**Go Type**:
```go
type Model struct {
    ID           string             `json:"id"`
    Name         string             `json:"name"`
    Capabilities ModelCapabilities  `json:"capabilities"`
    Constraints  ModelConstraints   `json:"constraints"`
    Tier         string             `json:"tier"`
    Version      string             `json:"version"`
}

type ModelCapabilities struct {
    Audio            bool   `json:"audio"`
    Extension        bool   `json:"extension"`
    ReferenceImages  bool   `json:"reference_images"`
    Resolutions      []string `json:"resolutions"`
    Durations        []int  `json:"durations"`
}

type ModelConstraints struct {
    MaxReferenceImages  int    `json:"max_reference_images"`
    RequiredAspectRatio string `json:"required_aspect_ratio,omitempty"`
    RequiredDuration    int    `json:"required_duration,omitempty"`
}
```

---

### Configuration

User settings and preferences.

**Fields**:
- `APIKey` (string, sensitive): Google Gemini API key
- `APIKeyEnv` (string, optional): Environment variable name for API key
- `DefaultModel` (string, required): Default model for generations
- `DefaultResolution` (string, required): Default resolution ("720p" or "1080p")
- `DefaultAspectRatio` (string, required): Default aspect ratio ("16:9" or "9:16")
- `DefaultDuration` (int, required): Default duration (4, 6, or 8)
- `OutputDirectory` (string, required): Default output directory for videos
- `PollIntervalSeconds` (int, required): Status polling interval (default: 10)
- `ConfigVersion` (string, required): Config file format version

**Persistence**:
- File path: `~/.config/veo3/config.yaml` (Linux/macOS) or `%APPDATA%\veo3\config.yaml` (Windows)
- File permissions: 0600 (user read/write only)

**Go Type**:
```go
type Configuration struct {
    APIKey              string `yaml:"api_key,omitempty" json:"-"`
    APIKeyEnv           string `yaml:"api_key_env,omitempty" json:"api_key_env,omitempty"`
    DefaultModel        string `yaml:"default_model" json:"default_model"`
    DefaultResolution   string `yaml:"default_resolution" json:"default_resolution"`
    DefaultAspectRatio  string `yaml:"default_aspect_ratio" json:"default_aspect_ratio"`
    DefaultDuration     int    `yaml:"default_duration" json:"default_duration"`
    OutputDirectory     string `yaml:"output_directory" json:"output_directory"`
    PollIntervalSeconds int    `yaml:"poll_interval_seconds" json:"poll_interval_seconds"`
    ConfigVersion       string `yaml:"version" json:"version"`
}
```

---

### PromptTemplate

Reusable prompt pattern with variable substitution.

**Fields**:
- `Name` (string, required): Unique template identifier
- `Prompt` (string, required): Template text with {{variable}} placeholders
- `Variables` ([]string, derived): List of variable names extracted from prompt
- `Description` (string, optional): Template purpose and usage notes
- `Tags` ([]string, optional): Categorization tags for organization
- `CreatedAt` (time.Time, required): Template creation timestamp
- `UpdatedAt` (time.Time, required): Last modification timestamp

**Variable Syntax**: Mustache-style `{{variable_name}}`

**Go Type**:
```go
type PromptTemplate struct {
    Name        string    `yaml:"name" json:"name"`
    Prompt      string    `yaml:"prompt" json:"prompt"`
    Description string    `yaml:"description,omitempty" json:"description,omitempty"`
    Tags        []string  `yaml:"tags,omitempty" json:"tags,omitempty"`
    CreatedAt   time.Time `yaml:"created_at" json:"created_at"`
    UpdatedAt   time.Time `yaml:"updated_at" json:"updated_at"`
}

// Derived at runtime
func (t *PromptTemplate) Variables() []string {
    // Extract {{variable}} patterns from Prompt
}
```

---

### BatchManifest

Collection of generation requests for bulk processing.

**Fields**:
- `Jobs` ([]BatchJob, required): List of generation jobs
- `Concurrency` (int, optional): Max parallel jobs (default: 3)
- `ContinueOnError` (bool, optional): Don't stop batch on individual failures (default: true)
- `OutputDirectory` (string, optional): Override output directory for all jobs

**Go Type**:
```go
type BatchManifest struct {
    Jobs            []BatchJob `yaml:"jobs" json:"jobs"`
    Concurrency     int        `yaml:"concurrency,omitempty" json:"concurrency,omitempty"`
    ContinueOnError bool       `yaml:"continue_on_error,omitempty" json:"continue_on_error,omitempty"`
    OutputDirectory string     `yaml:"output_directory,omitempty" json:"output_directory,omitempty"`
}

type BatchJob struct {
    ID      string                 `yaml:"id" json:"id"`
    Type    string                 `yaml:"type" json:"type"` // "generate", "animate", "interpolate", "extend"
    Options map[string]interface{} `yaml:"options" json:"options"`
    Output  string                 `yaml:"output" json:"output"`
}
```

---

### GeneratedVideo

Output video file with metadata.

**Fields**:
- `FilePath` (string, required): Local path to downloaded video
- `OperationID` (string, required): Operation that generated this video
- `Model` (string, required): Model used for generation
- `Prompt` (string, optional): Original generation prompt
- `DurationSeconds` (int, required): Actual video duration
- `Resolution` (string, required): Video resolution
- `AspectRatio` (string, required): Video aspect ratio
- `FileSizeBytes` (int64, required): File size in bytes
- `GenerationTimeSeconds` (int, required): Time taken to generate
- `CreatedAt` (time.Time, required): Download/creation timestamp

**Go Type**:
```go
type GeneratedVideo struct {
    FilePath               string    `json:"file_path"`
    OperationID            string    `json:"operation_id"`
    Model                  string    `json:"model"`
    Prompt                 string    `json:"prompt,omitempty"`
    DurationSeconds        int       `json:"duration_seconds"`
    Resolution             string    `json:"resolution"`
    AspectRatio            string    `json:"aspect_ratio"`
    FileSizeBytes          int64     `json:"file_size_bytes"`
    GenerationTimeSeconds  int       `json:"generation_time_seconds"`
    CreatedAt              time.Time `json:"created_at"`
}
```

---

## Relationships

```
Configuration
    ├─ provides defaults → GenerationRequest
    └─ stores credentials → APIClient

GenerationRequest (base)
    ├─ extended by → ImageRequest
    ├─ extended by → InterpolationRequest
    ├─ extended by → ReferenceImageRequest
    └─ used by → ExtensionRequest

GenerationRequest
    └─ creates → Operation

Operation
    ├─ tracked by → OperationManager
    ├─ may have → OperationError
    └─ produces → GeneratedVideo

Model
    ├─ validates → GenerationRequest
    └─ selected in → Configuration

PromptTemplate
    ├─ stored in → TemplateManager
    └─ expands to → GenerationRequest.Prompt

BatchManifest
    └─ contains many → BatchJob
        └─ each creates → Operation

GeneratedVideo
    └─ may be extended by → ExtensionRequest
```

## Validation State Machine

```
Request Creation
    ↓
Parameter Validation (local)
    ├─ invalid → Error (with suggestions)
    └─ valid → File Validation
        ├─ invalid → Error (format/size/corruption)
        └─ valid → Model Constraint Validation
            ├─ invalid → Error (incompatible features)
            └─ valid → API Submission
                └─ creates → Operation (PENDING)
```

## Storage Patterns

### Configuration
- **Format**: YAML
- **Location**: XDG config directory
- **Permissions**: 0600 (user-only)
- **Backup**: Atomic write via temp file + rename

### Templates
- **Format**: YAML (multiple templates per file)
- **Location**: `~/.config/veo3/templates.yaml`
- **Structure**: Array of PromptTemplate objects

### Batch Results
- **Format**: JSON
- **Location**: User-specified or default output directory
- **Naming**: `batch_results_{timestamp}.json`
- **Content**: Array of operation results with metadata

### Operation Cache
- **Format**: JSON
- **Location**: `~/.cache/veo3/operations/` (optional persistence)
- **Purpose**: Resume interrupted operations
- **Expiry**: Clean up completed operations >24h old

## Type Hierarchies

### Request Hierarchy
```
GenerationRequest (base)
    ├─ ImageRequest (+ image file)
    ├─ InterpolationRequest (+ first + last frames)
    ├─ ReferenceImageRequest (+ reference images)
    └─ ExtensionRequest (+ video file)
```

### Result Types
```
CommandResult (interface)
    ├─ OperationResult (async operation created)
    ├─ VideoResult (video downloaded)
    ├─ ConfigResult (config changed)
    ├─ ListResult (items listed)
    └─ ErrorResult (command failed)
```

## Data Flow

### Generation Flow
```
User Input (CLI)
    → Parse Flags/Args
    → Load Configuration
    → Create Request Object
    → Validate Request
    → Load/Validate Files
    → Submit to API
    → Create Operation
    → Poll Status (with progress)
    → Download Video
    → Save GeneratedVideo metadata
    → Display Success
```

### Configuration Flow
```
Load Order (highest priority first):
    1. CLI Flags (--api-key, --model, etc.)
    2. Environment Variables (GEMINI_API_KEY, etc.)
    3. Config File (~/.config/veo3/config.yaml)
    4. Defaults (hardcoded in code)
```

### Template Flow
```
Template Usage:
    Load Template by Name
    → Extract Variables
    → Prompt User for Values (or use --var flags)
    → Substitute Variables
    → Create GenerationRequest with expanded prompt
    → Continue normal generation flow
```

## Constants

### Default Values
```go
const (
    DefaultModel          = "veo-3.1-generate-preview"
    DefaultResolution     = "720p"
    DefaultAspectRatio    = "16:9"
    DefaultDuration       = 8
    DefaultPollInterval   = 10  // seconds
    DefaultConcurrency    = 3   // batch jobs
    
    MaxImageSize         = 20 * 1024 * 1024  // 20MB
    MaxVideoLength       = 141               // seconds
    MaxPromptLength      = 1024              // tokens (approximate chars)
    MaxReferenceImages   = 3
)
```

### Error Codes
```go
const (
    ErrCodeSafetyFilter     = "SAFETY_FILTER"
    ErrCodeRateLimit        = "RATE_LIMIT"
    ErrCodeInvalidImage     = "INVALID_IMAGE"
    ErrCodeInvalidVideo     = "INVALID_VIDEO"
    ErrCodeInvalidParams    = "INVALID_PARAMS"
    ErrCodeTimeout          = "TIMEOUT"
    ErrCodeUnauthenticated  = "UNAUTHENTICATED"
)
```

---

## Summary

This data model provides:
- Clear type hierarchy for different generation modes
- Comprehensive validation rules at multiple layers
- Proper state management for async operations
- Flexible configuration with sensible defaults
- Extensibility for future features
- Type safety through Go structs
- JSON/YAML serialization for all data types