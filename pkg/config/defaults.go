package config

const (
	DefaultModel         = "veo-3.1-generate-preview"
	DefaultResolution    = "720p"
	DefaultAspectRatio   = "16:9"
	DefaultDuration      = 8
	DefaultPollInterval  = 10 // seconds
	DefaultConcurrency   = 3  // batch jobs
	DefaultConfigVersion = "1.0"

	MaxImageSize       = 20 * 1024 * 1024 // 20MB
	MaxVideoLength     = 141              // seconds
	MaxPromptLength    = 1024             // tokens (approximate chars)
	MaxReferenceImages = 3
)
