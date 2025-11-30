# Veo3 CLI Quickstart Guide

**Last Updated**: 2025-11-30  
**Target Audience**: First-time users of the Veo3 CLI  
**Time to Complete**: ~5 minutes  
**Prerequisites**: Google Gemini API key

---

## What You'll Learn

By the end of this guide, you'll be able to:
- Install and configure the Veo3 CLI
- Generate your first AI video from a text prompt
- Understand basic CLI options and commands
- Check generation status and manage operations

---

## Step 1: Installation

### Option A: Download Binary (Recommended)

**macOS (Intel)**:
```bash
curl -L https://github.com/yourorg/veo3-cli/releases/latest/download/veo3-darwin-amd64 -o veo3
chmod +x veo3
sudo mv veo3 /usr/local/bin/
```

**macOS (Apple Silicon)**:
```bash
curl -L https://github.com/yourorg/veo3-cli/releases/latest/download/veo3-darwin-arm64 -o veo3
chmod +x veo3
sudo mv veo3 /usr/local/bin/
```

**Linux**:
```bash
curl -L https://github.com/yourorg/veo3-cli/releases/latest/download/veo3-linux-amd64 -o veo3
chmod +x veo3
sudo mv veo3 /usr/local/bin/
```

**Windows** (PowerShell as Administrator):
```powershell
Invoke-WebRequest -Uri https://github.com/yourorg/veo3-cli/releases/latest/download/veo3-windows-amd64.exe -OutFile veo3.exe
Move-Item veo3.exe C:\Windows\System32\
```

### Option B: Install via Homebrew (macOS/Linux)

```bash
brew tap yourorg/veo3
brew install veo3
```

### Option C: Build from Source

Requires Go 1.21+:
```bash
git clone https://github.com/yourorg/veo3-cli.git
cd veo3-cli
go build -o veo3 ./cmd/veo3
sudo mv veo3 /usr/local/bin/
```

### Verify Installation

```bash
veo3 --version
```

Expected output:
```
veo3 version 1.0.0 (commit: abc123, built: 2025-11-30)
```

---

## Step 2: Get Your API Key

1. Go to [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Sign in with your Google account
3. Click **"Get API Key"** or **"Create API Key"**
4. Copy your API key (keep it secure!)

**Note**: The API key looks like: `AIzaSyABC123_your_actual_key_here_DEF456`

---

## Step 3: Configure the CLI

### Interactive Setup (Recommended)

```bash
veo3 config init
```

You'll be prompted for:
- **API Key**: Paste your Gemini API key
- **Default Model**: Press Enter for `veo-3.1-generate-preview` (recommended)
- **Default Resolution**: Press Enter for `720p`
- **Default Duration**: Press Enter for `8` seconds
- **Output Directory**: Press Enter for current directory or specify a path

Example session:
```
Welcome to Veo3 CLI configuration!

Enter your Gemini API key: AIzaSyABC123_your_key_here
✓ API key validated successfully

Default model [veo-3.1-generate-preview]: 
✓ Using veo-3.1-generate-preview

Default resolution [720p]: 
✓ Using 720p

Default duration (4, 6, or 8 seconds) [8]: 
✓ Using 8 seconds

Output directory [.]: ~/Videos/veo3
✓ Using ~/Videos/veo3

Configuration saved to ~/.config/veo3/config.yaml
You're ready to generate videos!
```

### Manual Configuration

Create `~/.config/veo3/config.yaml`:
```yaml
api_key: AIzaSyABC123_your_key_here
default_model: veo-3.1-generate-preview
default_resolution: 720p
default_aspect_ratio: "16:9"
default_duration: 8
output_directory: ~/Videos/veo3
poll_interval_seconds: 10
version: "1.0"
```

### Alternative: Environment Variable

For temporary use or CI/CD:
```bash
export GEMINI_API_KEY="AIzaSyABC123_your_key_here"
```

---

## Step 4: Generate Your First Video

### Basic Text-to-Video

```bash
veo3 generate "A majestic lion walking through the African savannah at sunset"
```

**What happens**:
1. CLI validates your prompt and configuration
2. Request submitted to Veo API
3. Progress updates display every 10 seconds
4. Video automatically downloads when complete
5. Success message shows file location

Expected output:
```
Starting video generation...
Model: veo-3.1-generate-preview
Prompt: "A majestic lion walking through the African savannah at sunset"

⏳ Generating video... (elapsed: 0:00:10)
⏳ Generating video... (elapsed: 0:00:20)
⏳ Generating video... (elapsed: 0:00:35)
⏳ Generating video... (elapsed: 0:00:50)

✅ Video generated successfully!
📁 Saved to: ~/Videos/veo3/lion_savannah_20251130_143025.mp4
📊 Duration: 8 seconds | Resolution: 720p | Size: 12.4 MB
⏱️  Generation time: 52 seconds

Tip: View your video with: open ~/Videos/veo3/lion_savannah_20251130_143025.mp4
```

### Generate with Custom Options

```bash
veo3 generate "A cinematic close-up of rain drops on a window" \
  --resolution 1080p \
  --duration 8 \
  --aspect-ratio 16:9 \
  --negative-prompt "cartoon, animated, drawing" \
  --output rainy_window.mp4
```

---

## Step 5: Explore More Features

### Animate a Static Image

```bash
veo3 animate path/to/photo.jpg \
  --prompt "The person smiles and waves at the camera" \
  --output animated_photo.mp4
```

### Check Available Models

```bash
veo3 models list
```

Output:
```
Available Veo Models:

veo-3.1-generate-preview (recommended)
  Audio:      ✓
  Resolution: 720p, 1080p
  Duration:   4s, 6s, 8s
  Extension:  ✓
  References: ✓ (up to 3 images)
  Tier:       Standard

veo-3.1-fast-generate-preview
  Audio:      ✓
  Resolution: 720p, 1080p
  Duration:   4s, 6s, 8s
  Extension:  ✓
  References: ✓ (up to 3 images)
  Tier:       Fast (optimized for speed)

[... other models ...]
```

### Manage Operations

If your terminal disconnects during generation:

```bash
# List recent operations
veo3 operations list

# Check specific operation status
veo3 operations status operations/abc123def456

# Download completed video
veo3 operations download operations/abc123def456 --output my_video.mp4

# Cancel pending operation
veo3 operations cancel operations/abc123def456
```

---

## Common Use Cases

### Creative Content

```bash
# Cinematic establishing shot
veo3 generate "Wide aerial shot of a futuristic city at night with neon lights" \
  --resolution 1080p --duration 8

# Product showcase
veo3 animate product_photo.png \
  --prompt "Product rotates slowly on a reflective surface with soft lighting"

# Nature scene
veo3 generate "Time-lapse of clouds moving over mountain peaks" \
  --duration 6
```

### With Audio Cues

```bash
# Dialogue
veo3 generate "Two people whispering 'This is the secret' in a library" \
  --duration 8

# Ambient sound
veo3 generate "Ocean waves crashing on a beach with seagulls calling" \
  --duration 8

# Music reference
veo3 generate "A drummer playing energetic rock music" \
  --duration 6
```

---

## Troubleshooting

### "API key not valid"

**Problem**: Authentication failed  
**Solution**:
```bash
# Verify config
veo3 config show

# Reset and reconfigure
veo3 config init
```

### "Safety filter triggered"

**Problem**: Content blocked by safety filters  
**Solution**: Revise your prompt to avoid:
- Violence or weapons
- Explicit adult content
- Copyrighted characters/brands
- Dangerous activities

Try adding positive elements instead:
```bash
# Instead of: "person with weapon"
veo3 generate "person exploring ancient ruins"

# Instead of: "car crash"
veo3 generate "car driving safely on scenic road"
```

### "Image file too large"

**Problem**: Input image exceeds 20MB  
**Solution**: Compress the image:
```bash
# macOS/Linux
convert input.jpg -quality 85 -resize 2048x2048\> output.jpg

# Or use online tools
# Then try again
veo3 animate output.jpg --prompt "your prompt"
```

### Generation Takes Too Long

**Problem**: Video generation exceeding 5 minutes  
**Solution**:
- This is normal during high API load
- Use `--async` flag to return immediately:
  ```bash
  veo3 generate "your prompt" --async
  # Returns operation ID instantly
  
  # Check later
  veo3 operations status operations/your-operation-id
  ```

---

## Tips for Better Videos

### Writing Effective Prompts

✅ **Good Prompts**:
- "A cinematic close-up of a woman's face as she smiles warmly"
- "Wide-angle shot of a skateboarder performing a trick in slow motion"
- "Panning shot across a field of blooming sunflowers"

❌ **Avoid**:
- Too vague: "A video of a person"
- Too complex: "A 3-minute epic film with multiple scenes..."
- Impossible: "A person teleporting"

### Camera Movements

Include camera directions for better results:
- "Slow zoom in on..."
- "Panning shot across..."
- "Aerial drone shot flying over..."
- "Close-up of... pulling back to reveal..."
- "Tracking shot following..."

### Lighting and Mood

Specify lighting for atmosphere:
- "Golden hour lighting"
- "Soft studio lighting"
- "Dramatic side lighting"
- "Neon lights in the background"
- "Natural daylight"

---

## Next Steps

### Learn More Commands

```bash
# View all commands
veo3 --help

# Get help for specific command
veo3 generate --help
veo3 operations --help
```

### Advanced Features

1. **Batch Processing**: Generate multiple videos from a manifest
   ```bash
   veo3 batch process manifest.yaml
   ```

2. **Prompt Templates**: Save reusable prompts with variables
   ```bash
   veo3 templates save product-showcase --prompt "{{product}} rotates on {{background}}"
   veo3 generate --template product-showcase --var product="watch" --var background="marble"
   ```

3. **Video Extension**: Create longer videos by chaining extensions
   ```bash
   veo3 extend video1.mp4 --prompt "Action continues" --output video2.mp4
   veo3 extend video2.mp4 --prompt "Climax scene" --output video3.mp4
   ```

### Explore Documentation

- Full command reference: `veo3 --help`
- Model specifications: `veo3 models info <model-name>`
- Configuration options: `veo3 config --help`

---

## Getting Help

### Command-Line Help

```bash
# General help
veo3 --help

# Command-specific help
veo3 generate --help
veo3 config --help

# Verbose output for debugging
veo3 generate "prompt" --verbose
```

### Community & Support

- GitHub Issues: Report bugs and request features
- Documentation: Full guide and API reference
- Examples: Sample prompts and use cases

---

## Quick Reference Card

```bash
# Generate video from text
veo3 generate "your prompt"

# Animate an image
veo3 animate image.png --prompt "animation description"

# Check operations
veo3 operations list

# View configuration
veo3 config show

# Get help
veo3 --help
```

---

**Congratulations!** 🎉 You've successfully generated your first AI video with the Veo3 CLI.

Happy creating! 🎬