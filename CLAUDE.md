# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoPro Metadata Validator & Fixer - validates and corrects GoPro video metadata using GPS timestamps from the GPMF telemetry stream.

**Active development:** All work happens in `/go-validator/` (Go implementation)
**Archived:** `/ts-validator/` is a reference implementation, not maintained

## Quick Commands

### Build
```bash
cd go-validator
make build
```

### Test with Sample Files
```bash
# From project root
make go              # Validate sample files
make concat-dry      # Preview concatenation
make update-dry      # Preview metadata updates
make rename-dry      # Preview file organization
```

### Run Directly
```bash
cd go-validator
./gopro-validator --input /path/to/videos
./gopro-validator --input ../sample-input-files --concat --dry-run
```

### Development
```bash
cd go-validator
make build                              # Build binary
./gopro-validator --input ../sample-input-files

make fmt                                # Format code
make vet                                # Static analysis
make deps                               # Download/tidy dependencies
make test                               # Run tests
```

## Architecture

### Data Flow
```
MP4 File → ffprobe (metadata) → Metadata struct
         → go-mp4 (GPMF track) → KLV parser → GPS timestamps
         → Comparator → ValidationResult
         → Actions (rename/update/concat)
```

### Key Components

**`cmd/main.go`** - CLI argument parsing, orchestration, output formatting
**`internal/validator/gpmf.go`** - GPMF stream extraction and parsing (KLV structure)
**`internal/validator/metadata.go`** - MP4 metadata extraction via ffprobe
**`internal/validator/comparator.go`** - Validation logic, compares GPS vs file metadata
**`internal/validator/actions.go`** - File operations (rename, metadata updates, streaming file copy)
**`internal/validator/concat.go`** - Chapter detection and concatenation
**`internal/validator/sidecar.go`** - XMP sidecar file generation with GPS metadata
**`internal/validator/validator.go`** - Types and main validation orchestration

### GPMF Parsing Strategy

GPMF data is extracted from MP4 files using pure Go (`github.com/abema/go-mp4`):
1. Open file, navigate `moov/trak` boxes to find the "GoPro MET" handler track
2. Read sample table (stco/co64 offsets, stsz sizes, stsc mapping)
3. Read each GPMF sample directly from the file via `io.ReadSeeker` — no temp files, no subprocess
4. Custom KLV parser reads Key-Length-Value structure (DEVC → STRM → data keys)
5. Extract GPSU/GPSUU entries for absolute UTC timestamps
6. Extract STMP (µs, uint64) / TSMP (sample counter) for relative timestamps
7. Extract GPS5 coordinates using per-stream SCAL factors

**No ffmpeg needed** for GPMF extraction. ffmpeg is only used for `--update-metadata` (remux) and `--concat` (chapter joining).

**Critical:** GPS timestamps are relative (milliseconds since recording started), not absolute. Must adjust by subtracting the relative offset from the absolute GPS time to get true recording start time. See `calculateRecordingStartTime()` in `internal/validator/actions.go`.

### Chapter File Detection

GoPro splits long recordings into chapters:
- `GH016978.MP4` - Chapter 1 (first two digits = chapter number)
- `GH026978.MP4` - Chapter 2 (same base number 6978)
- `GH036978.MP4` - Chapter 3

Regex pattern: `(?i)GH(\d)(\d)(\d{4})\.MP4` (case-insensitive, matches both .MP4 and .mp4)

Chapter files have cumulative GPS relative timestamps:
- Chapter 1: ~0s start
- Chapter 2: ~105s start (continues from where ch1 ended)
- Chapter 3: ~210s start (cumulative)

This is NOT an error - it's expected behavior. Use it to verify proper concatenation order.

### Path Handling (Cross-Platform)

**Always use:** `filepath.Join()`, `filepath.Base()`, `filepath.Abs()`
**Never use:** Hardcoded `/` or `\` separators

**Special case for ffmpeg:** Use `filepath.ToSlash()` when creating concat file lists - ffmpeg requires forward slashes even on Windows.

### External Dependencies

**github.com/abema/go-mp4** - MP4 atom parsing for GPMF extraction (compiled in, no runtime install)
**ffprobe** - MP4 container metadata (creation_time, timecode) — required for validation
**ffmpeg** - Video remuxing and concatenation — required only for `--update-metadata` and `--concat`

ffmpeg/ffprobe must be in PATH when needed. Validation-only runs (`--input` without action flags) need only ffprobe.

## Key Concepts

### GPS Ground Truth
GPS timestamps from satellites are the source of truth. File metadata can be wrong due to:
- Camera timezone settings incorrect (e.g., EST marked as UTC)
- Camera date/time not set correctly
- Firmware bugs

Always trust GPS over file metadata when they conflict.

### Dry-Run First
All operations that modify files support `--dry-run`. Always preview before applying:
- `--rename` copies files using streaming I/O (safe, originals untouched, memory-efficient)
- `--concat` creates new files (safe)
- `--update-metadata` modifies files IN PLACE (destructive, recommend dry-run + backup)
- `--write-sidecar` creates XMP sidecar files with GPS metadata

Note: In dry-run mode, unique filename generation tracks paths in memory to correctly show (1), (2) suffixes even when output files don't exist yet.

### Output Directory Behavior
- `--output` and `--concat-output` are relative to CWD by default
- Can be absolute paths
- Directories created if they don't exist

## Common Development Patterns

### Adding a New CLI Flag
1. Add to flag variables in `cmd/main.go`
2. Add logic in `main()` to call appropriate function
3. Update help text in flag definition
4. Update `go-validator/README.md` with new flag
5. Update root `README.md` if it's a major feature

### Adding a New GPMF Field
1. Identify the GPMF key (e.g., GPSU, STMP, GPS5)
2. Add parsing in `parseGPMFData()` in `internal/validator/gpmf.go`
3. Add field to `GPSData` struct in `internal/validator/validator.go`
4. Update comparator logic if needed

### Adding a New Validation Check
1. Add logic to `compareMetadata()` in `internal/validator/comparator.go`
2. Append issues to the `issues` slice
3. Test with sample files that exhibit the issue

## Important File Locations

**Documentation:**
- `README.md` - Main project documentation
- `QUICK-START.md` - User getting started guide
- `USAGE-EXAMPLES.md` - Real-world scenarios
- `go-validator/CONCAT.md` - Chapter concatenation details
- `go-validator/GPS-OFFSET-FIX.md` - GPS lock delay technical details

**Sample Data:**
- `sample-input-files/` - Test videos (NOT in git, too large)
- Files organized by date: `YYYY-MM-DD/HERO7 Black 1/GHxxxxxx.MP4`

## Testing

No formal test suite yet. Test manually using:
1. Sample files in `sample-input-files/` (if available locally)
2. Your own GoPro videos
3. Makefile commands for quick testing

Always test on a small subset before processing large libraries.

## Platform-Specific Notes

**Windows:** All path handling uses `filepath` package (cross-platform). Concat file generation uses `filepath.ToSlash()` for ffmpeg compatibility.

**Memory:** GPMF parsing streams data via ffmpeg, doesn't load entire file. Can handle files of any size.

**Performance:** Bottleneck is ffmpeg I/O, not Go code. Concat is fast (codec copy, no re-encoding).

## When Making Changes

1. **All development in `go-validator/` only** - TypeScript version is archived
2. **Update CHANGELOG.md** for user-facing changes
3. **Test with `make` commands** using sample files
4. **Update documentation** if adding features or changing behavior
5. **Verify cross-platform** - use `filepath` package, avoid OS-specific code
6. **Test dry-run modes** - ensure `--dry-run` flag works for new operations
