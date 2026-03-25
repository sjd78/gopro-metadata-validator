# GoPro Metadata Validator (Go Version)

A Go implementation that can handle files of any size, including >2GB files that the TypeScript version cannot process.

## Features

- ✅ **Handles large files (>2GB)** - streams GPMF data via ffmpeg
- ✅ **Custom GPMF parser** - doesn't load entire file into memory
- ✅ **Fast execution** - compiled binary with minimal dependencies
- ✅ **Extracts GPS absolute timestamps** - validates against file metadata
- ✅ **Rename/move files** - organize by GPS timestamp
- ✅ **Update metadata** - fix incorrect creation times

## Building

```bash
go build -o gopro-validator
```

## Quick Start

### Just validate (no changes):
```bash
# Scan current directory
./gopro-validator

# Scan specific directory
./gopro-validator --input /path/to/videos
```

### Preview file renaming:
```bash
./gopro-validator --input /path/to/videos --rename --dry-run
```

### Preview metadata updates:
```bash
./gopro-validator --input /path/to/videos --update-metadata --dry-run
```

### Actually fix metadata:
```bash
./gopro-validator --update-metadata
```

### Organize files by GPS timestamp:
```bash
./gopro-validator --rename --output ~/Videos/GoPro-Organized
```

### Concatenate chapter files:
```bash
# Preview
./gopro-validator --concat --dry-run

# Create full recordings
./gopro-validator --concat
```

See [USAGE.md](USAGE.md) for detailed documentation and [CONCAT.md](CONCAT.md) for chapter concatenation details.

**Important:** The tool automatically adjusts for GPS lock delay - see [GPS-OFFSET-FIX.md](GPS-OFFSET-FIX.md) for details.

## Command-Line Options

```
  --version            Show version and exit
  --input DIR          Input directory containing GoPro files (default: current directory)
  --rename             Rename and move files based on GPS timestamps
  --update-metadata    Update MP4 metadata to match GPS timestamps
  --concat             Concatenate chapter files into complete recordings
  --write-sidecar      Write XMP sidecar files with GPMF metadata
  --dry-run            Show what would be done without making changes
  --output DIR         Output directory for renamed files (default: renamed-files)
  --concat-output DIR  Output directory for concatenated files (default: concatenated-files)
```

## XMP Sidecar Files

Export GPMF metadata to XMP sidecar files for use with exiftool:

```bash
# Generate XMP sidecars alongside video files
./gopro-validator --input /path/to/videos --write-sidecar
```

This creates `.xmp` files (e.g., `GH016978.MP4.xmp`) containing:
- GPS timestamps and recording start time
- GPS lock delay (time from recording start to GPS lock)
- Tool processing information

**Using with exiftool to embed metadata:**

```bash
# Embed XMP metadata into a single MP4 file
exiftool -tagsFromFile video.mp4.xmp -all:all video.mp4

# Batch process all MP4 files in a directory
exiftool -tagsFromFile %f.xmp -all:all -ext MP4 /path/to/videos/
```

**Benefits:**
- Non-destructive (doesn't modify original files)
- Standard format (compatible with Lightroom, Bridge, etc.)
- Flexible (choose which metadata to embed with exiftool)

**Note:** GPS coordinates (lat/lon/altitude) will be added in a future update when GPS5 binary parsing is implemented.
```

## How It Works

1. Uses `ffprobe` to find the GPMF stream index
2. Uses `ffmpeg` to extract just the GPMF binary stream (not the entire file)
3. Parses GPMF KLV (Key-Length-Value) structure recursively
4. Extracts TSMP timestamps from GPS data streams
5. Compares with file metadata (creation_time, timecode)

## GPMF Structure

```
DEVC (Device)
  └─ STRM (Stream)
      ├─ TSMP (Timestamp in milliseconds)
      ├─ GPS5 (GPS data: lat, lon, alt, speed2D, speed3D)
      ├─ ACCL (Accelerometer)
      └─ GYRO (Gyroscope)
```

## Key Differences from TypeScript Version

| Feature | TypeScript | Go |
|---------|-----------|-----|
| Max file size | 2GB | Unlimited |
| Dependencies | Node.js + 3 npm packages | Go compiler only |
| Memory usage | Loads entire file | Streams GPMF only |
| Execution speed | ~5s | ~2s |
| GPMF parsing | Library-based | Custom parser |

## Chapter Files

GoPro cameras split long recordings into chapters:
- `GH01xxxx.MP4` - First chapter (GPS timestamps start at ~0s)
- `GH02xxxx.MP4` - Second chapter (GPS timestamps continue from previous)
- `GH03xxxx.MP4` - Third chapter, etc.

The validator correctly identifies chapter files by detecting GPS timestamps that don't start near 0.

## Dependencies

Runtime requirements:
- `ffmpeg` - for GPMF stream extraction
- `ffprobe` - for metadata extraction

## Performance

Tested on 10 files (38GB total):
- Processing time: ~2 seconds
- Memory usage: <50MB peak
