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
| **Dependencies** | Only Go compiler + ffmpeg |

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

## Features

- Extracts file metadata (creation time, timecode) from GoPro MP4 files
- Parses GPMF (GoPro Metadata Format) telemetry stream to extract GPS data
- Validates internal consistency between timecode and creation_time
- Reports files without GPS data or with metadata discrepancies

## Installation

```bash
npm install
```

## Usage

Place your GoPro videos in the `input-files` directory (maintains subdirectory structure), then run:

```bash
npm run dev
```

Or build and run:

```bash
npm run build
npm start
```

## How It Works

The tool performs these validation checks:

1. **Timecode vs Creation Time**: Compares the embedded timecode (time-of-day when recording started) against the creation_time metadata. A discrepancy > 2 minutes suggests incorrect metadata date.

2. **GPS Data Availability**: Checks if the GPMF stream contains GPS5 telemetry data.

3. **GPS Timestamp Validation**: Verifies that GPS timestamps start near 0 (recording start) and progress logically.

## Output

```
✓ GH016761.MP4
  Metadata Creation Time: 2016-01-01T00:10:36.000Z
  Timecode: 00:10:35:50
  GPS Samples: 8317
  GPS First Timestamp: 0.000s
  GPS Last Timestamp: 444.4s (duration: 444.4s)
```

## Known Limitations

### Large Files (>2GB)
The `gpmf-extract` library loads the entire MP4 file into memory, which fails for files larger than 2GB. These files will report "No GPS data found" even if GPS data exists.

**Workarounds:**
- Split large files before processing
- Use alternative tools like `exiftool` or `ffprobe` for large file metadata extraction
- Implement custom GPMF binary parser (future enhancement)

### GPS Timestamps
GPS timestamps in GPMF are **relative** (milliseconds since recording start), not absolute dates. The tool validates they start near 0 and progress logically but cannot independently verify the recording date without external reference.

## Libraries Used

- **gpmf-extract** - Extracts GPMF binary data from GoPro MP4 files
- **gopro-telemetry** - Parses GPMF format to structured telemetry data (GPS, accelerometer, etc.)

## Project Structure

```
src/
  index.ts              - Main entry point and file scanner
  metadata-extractor.ts - Extracts MP4 metadata using ffprobe
  gpmf-extractor.ts     - Extracts GPS data from GPMF stream
  comparator.ts         - Validates metadata consistency
```

## Project Structure

```
gopro_renamer/
├── go-validator/          # Go implementation (handles large files)
│   ├── main.go           # Entry point
│   ├── gpmf.go           # Custom GPMF parser
│   ├── metadata.go       # Metadata extraction
│   ├── comparator.go     # Validation logic
│   └── validator.go      # Main validation orchestration
│
├── src/                  # TypeScript implementation
│   ├── index.ts
│   ├── gpmf-extractor.ts
│   ├── metadata-extractor.ts
│   └── comparator.ts
│
├── input-files/          # Place your GoPro videos here
├── Makefile             # Build and run commands
├── RESULTS.md           # Sample validation results
└── README.md            # This file
```

## Results

See [RESULTS.md](RESULTS.md) for detailed validation results on sample files.

**Key Finding:** All sample files have internally consistent metadata. The Go version successfully processes all files including 5 large files (>2GB) that the TypeScript version cannot handle.

## Future Enhancements

- [ ] ✅ ~~Support for large files (>2GB) via streaming GPMF parser~~ **DONE (Go version)**
- [ ] GPS coordinate extraction and mapping
- [ ] Metadata correction/rewriting capabilities
- [ ] Export validation report to JSON/CSV
- [ ] Batch file renaming based on GPS timestamps
- [ ] Web UI for drag-and-drop validation
