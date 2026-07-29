# GoPro Metadata Validator (Go Version)

A Go implementation that can handle files of any size, including >2GB files that the TypeScript version cannot process.

## Features

- ✅ **Handles large files (>2GB)** - reads only GPMF sample chunks via `io.ReadSeeker`
- ✅ **Pure Go GPMF extraction** - no ffmpeg needed for validation (uses `github.com/abema/go-mp4`)
- ✅ **Fast execution** - compiled binary, no subprocess spawning for validation
- ✅ **Extracts GPS absolute timestamps** - validates against file metadata
- ✅ **Rename/move files** - organize by GPS timestamp
- ✅ **Update metadata** - fix incorrect creation times
- ✅ **XMP sidecar generation** - GPS coordinates, timestamps, timezone, and quality data

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

GPS coordinates (lat/lon/altitude), speed, GPS fix type, precision/DOP, timezone, and GPS lock delay are all included in the sidecar output.

## How It Works

1. Opens the MP4 file and navigates to the "GoPro MET" handler track using `github.com/abema/go-mp4`
2. Reads the sample table (stco/co64, stsz, stsc) to locate GPMF data chunks
3. Reads each GPMF sample directly from the file via `io.ReadSeeker` — no temp files, no subprocess
4. Parses GPMF KLV (Key-Length-Value) structure recursively (DEVC → STRM → data keys)
5. Extracts GPS timestamps (GPSU, STMP), coordinates (GPS5), and quality data (GPSF, GPSP)
6. Uses `ffprobe` separately for MP4 container metadata (creation_time, timecode)
7. Compares GPS ground truth against file metadata to detect errors

## GPMF Keys Extracted

```
DEVC (Device container)
  └─ STRM (Stream container — each has its own SCAL scope)
      ├─ GPSU/GPSUU  Absolute UTC datetime (GPS ground truth)
      ├─ STMP        Relative µs timestamp (GPS lock delay calculation)
      ├─ TSMP        Sample counter (fallback ordering)
      ├─ GPS5        Lat, lon, alt, speed2D, speed3D
      ├─ SCAL        Scale factors (required to decode GPS5)
      ├─ GPSF        GPS fix type (none/2D/3D)
      └─ GPSP        Precision / DOP
```

Additional GPMF keys (ACCL, GYRO, CORI, ISOE, SHUT, WBAL, etc.) are available
in the stream but not currently extracted. See `gpmf.go` for the full reference list.

## Chapter Files

GoPro cameras split long recordings into chapters:
- `GH01xxxx.MP4` - First chapter (GPS timestamps start at ~0s)
- `GH02xxxx.MP4` - Second chapter (GPS timestamps continue from previous)
- `GH03xxxx.MP4` - Third chapter, etc.

The validator correctly identifies chapter files by detecting GPS timestamps that don't start near 0.

## Dependencies

**Validation only (no external tools needed):**
- GPMF extraction uses pure Go (`github.com/abema/go-mp4`) — compiled into the binary
- `ffprobe` — for MP4 container metadata (creation_time, timecode)

**For file operations:**
- `ffmpeg` — required only for `--update-metadata` (remux) and `--concat` (chapter joining)

## Performance

Tested on 10 files (38GB total):
- Processing time: ~2 seconds
- Memory usage: <50MB peak
