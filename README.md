# GoPro Metadata Validator & Fixer

Validate, detect errors, and fix GoPro video metadata using GPS timestamps embedded in the GPMF stream.

## 🎯 What This Does

1. **Validates** - Compares file metadata against GPS ground truth
2. **Detects Errors** - Finds wrong dates, timezone issues, metadata corruption
3. **Fixes Metadata** - Updates creation_time to match GPS timestamps
4. **Organizes Files** - Renames and sorts by actual recording time
5. **Concatenates Chapters** - Joins multi-part recordings into single files (preserves GPMF!)

## 📊 What We Found

Testing on sample files revealed:
- ❌ **1 file with completely wrong date** (5 years off!)
- ⚠️ **9 files with timezone issues** (local time marked as UTC)
- ✅ **GPS timestamps are 100% accurate** (satellite ground truth)

See [DISCOVERIES.md](DISCOVERIES.md) for details.

## Go Implementation (Production)

This project uses **Go** for the production implementation. A TypeScript version exists in `ts-validator/` for reference but is **not actively maintained**.

### Why Go?

| Feature | Capability |
|---------|-----------|
| **File size** | ♾️ Unlimited (handles >2GB files) |
| **Speed** | ~2 seconds for 10 files (38GB) |
| **GPS absolute time** | ✅ Extracts from GPMF stream |
| **Rename files** | ✅ Organize by GPS timestamp |
| **Update metadata** | ✅ Fix creation_time |
| **Concat chapters** | ✅ Join multi-part recordings |
| **Cross-platform** | ✅ Windows, Linux, macOS |
| **Dependencies** | Go compiler + ffmpeg (for fix/concat only) |

**All development happens in `go-validator/`**

---

## Installation

### Option 1: Download Pre-built Binary

Download the latest release for your platform from the [Releases page](https://github.com/sjd78/gopro-metadata-validator/releases).

**Stable Releases:**
- [Latest stable release](https://github.com/sjd78/gopro-metadata-validator/releases/latest) - Recommended for most users

**Development Builds:**
- [Latest development build](https://github.com/sjd78/gopro-metadata-validator/releases/tag/latest) - Cutting-edge features from main branch

**Platform Downloads:**

```bash
# Linux (x86_64)
wget https://github.com/sjd78/gopro-metadata-validator/releases/latest/download/gopro-validator-linux-amd64
chmod +x gopro-validator-linux-amd64
./gopro-validator-linux-amd64 --version

# Linux (ARM64 - Raspberry Pi, etc.)
wget https://github.com/sjd78/gopro-metadata-validator/releases/latest/download/gopro-validator-linux-arm64
chmod +x gopro-validator-linux-arm64

# macOS (Apple Silicon)
wget https://github.com/sjd78/gopro-metadata-validator/releases/latest/download/gopro-validator-darwin-arm64
chmod +x gopro-validator-darwin-arm64

# macOS (Intel)
wget https://github.com/sjd78/gopro-metadata-validator/releases/latest/download/gopro-validator-darwin-amd64
chmod +x gopro-validator-darwin-amd64

# Windows (PowerShell) - x86_64
Invoke-WebRequest -Uri "https://github.com/sjd78/gopro-metadata-validator/releases/latest/download/gopro-validator-windows-amd64.exe" -OutFile "gopro-validator.exe"
```

**Verify Downloads:**
Download `checksums.txt` and verify:
```bash
sha256sum -c checksums.txt
```

### Option 2: Build from Source

Requires Go 1.21 or later and ffmpeg/ffprobe.

```bash
git clone https://github.com/sjd78/gopro-metadata-validator.git
cd gopro-metadata-validator/go-validator
go build -o gopro-validator
./gopro-validator --version
```

---

## Quick Start

### Production Version (Go)

**Just validate (no changes):**
```bash
cd go-validator
go build -o gopro-validator
./gopro-validator --input /path/to/your/videos
```

**Fix metadata based on GPS timestamps:**
```bash
# Preview changes first
./gopro-validator --input /path/to/videos --update-metadata --dry-run

# Apply fixes
./gopro-validator --input /path/to/videos --update-metadata
```

**Rename/organize files by GPS timestamp:**
```bash
# Preview
./gopro-validator --input /path/to/videos --rename --dry-run

# Copy to organized folders
./gopro-validator --input /path/to/videos --rename --output ~/Videos/GoPro-Organized
```

See [go-validator/USAGE.md](go-validator/USAGE.md) for detailed documentation.

### TypeScript Version (Reference Only)
```bash
cd ts-validator
npm install
npm run dev
```
**Note:** TypeScript version is **not actively maintained**. Limited to validation only, files <2GB. Use Go version for production.

---

## 📚 Documentation

- **[SUMMARY.md](SUMMARY.md)** - Project overview and quick start
- **[USAGE-EXAMPLES.md](USAGE-EXAMPLES.md)** - Real-world usage scenarios
- **[FEATURES.md](FEATURES.md)** - Complete feature reference
- **[DISCOVERIES.md](DISCOVERIES.md)** - Errors found in sample files
- **[go-validator/USAGE.md](go-validator/USAGE.md)** - Detailed usage guide
- **[WINDOWS.md](WINDOWS.md)** - Windows compatibility guide
- **[CHANGELOG.md](CHANGELOG.md)** - Version history and changes

---

## TypeScript Version (Archived)

A Node.js/TypeScript reference implementation in `ts-validator/` directory. **Not actively maintained.** Use the Go version for all production work.

## How It Works

1. Opens MP4 files and reads the GPMF telemetry track using pure Go (`github.com/abema/go-mp4`)
2. Parses GPMF KLV structure to extract GPS timestamps, coordinates, and quality data
3. Reads MP4 container metadata (creation_time, timecode) via `ffprobe`
4. Compares GPS ground truth against file metadata to detect errors
5. Applies corrections (rename, metadata update, concatenation) as requested

## Project Structure

```
gopro_renamer/
├── go-validator/          ⭐ Active development
│   ├── main.go           # CLI parsing, orchestration
│   ├── gpmf.go           # GPMF extraction (go-mp4) and KLV parsing
│   ├── metadata.go       # MP4 metadata via ffprobe
│   ├── comparator.go     # Validation logic (GPS vs file metadata)
│   ├── actions.go        # File operations (rename, metadata update)
│   ├── concat.go         # Chapter detection and concatenation
│   ├── sidecar.go        # XMP sidecar generation
│   ├── timezone.go       # GPS coordinate → timezone lookup
│   └── validator.go      # Main validation orchestration
│
├── ts-validator/         📦 Archived (reference only)
├── sample-input-files/   Test data (not in git)
├── Makefile             Quick commands
└── Documentation files
```
