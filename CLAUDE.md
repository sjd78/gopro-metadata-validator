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
go build -o gopro-validator
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
# Build and test
cd go-validator
go build -o gopro-validator
./gopro-validator --input ../sample-input-files

# Format code
go fmt ./...

# Check for issues
go vet ./...
```

## Architecture

### Data Flow
```
MP4 File → ffprobe (metadata) → Metadata struct
         → ffmpeg (GPMF extract) → GPMF parser → GPS timestamps
         → Comparator → ValidationResult
         → Actions (rename/update/concat)
```

### Key Components

**`main.go`** - CLI argument parsing, orchestration, output formatting
**`gpmf.go`** - GPMF stream extraction and parsing (KLV structure)
**`metadata.go`** - MP4 metadata extraction via ffprobe
**`comparator.go`** - Validation logic, compares GPS vs file metadata
**`actions.go`** - File operations (rename, metadata updates)
**`concat.go`** - Chapter detection and concatenation
**`validator.go`** - Main validation orchestration

### GPMF Parsing Strategy

GPMF data is extracted from MP4 files using ffmpeg (doesn't load entire file):
1. `ffprobe` finds the GPMF stream index (usually stream 3, codec tag 'gpmd')
2. `ffmpeg` extracts just that stream to temp file
3. Custom KLV parser reads Key-Length-Value structure
4. Extract GPSU/GPSUU entries for absolute UTC timestamps
5. Extract TSMP/STMP entries for relative timestamps (ms since recording start)

**Critical:** GPS timestamps are relative (milliseconds since recording started), not absolute. Must adjust by subtracting the relative offset from the absolute GPS time to get true recording start time. See `calculateRecordingStartTime()` in `actions.go`.

### Chapter File Detection

GoPro splits long recordings into chapters:
- `GH016978.MP4` - Chapter 1 (first two digits = chapter number)
- `GH026978.MP4` - Chapter 2 (same base number 6978)
- `GH036978.MP4` - Chapter 3

Regex pattern: `GH(\d)(\d)(\d{4})\.MP4`

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

**ffmpeg** - Video processing, GPMF stream extraction
**ffprobe** - Metadata reading, stream detection

Both must be in PATH. Check with `exec.Command("ffmpeg", ...)` - will fail gracefully if not found.

## Key Concepts

### GPS Ground Truth
GPS timestamps from satellites are the source of truth. File metadata can be wrong due to:
- Camera timezone settings incorrect (e.g., EST marked as UTC)
- Camera date/time not set correctly
- Firmware bugs

Always trust GPS over file metadata when they conflict.

### Dry-Run First
All operations that modify files support `--dry-run`. Always preview before applying:
- `--rename` copies files (safe, originals untouched)
- `--concat` creates new files (safe)
- `--update-metadata` modifies files IN PLACE (destructive, recommend dry-run + backup)

### Output Directory Behavior
- `--output` and `--concat-output` are relative to CWD by default
- Can be absolute paths
- Directories created if they don't exist

## Common Development Patterns

### Adding a New CLI Flag
1. Add to flag variables in `main.go`
2. Add logic in `main()` to call appropriate function
3. Update help text in flag definition
4. Update `go-validator/README.md` with new flag
5. Update root `README.md` if it's a major feature

### Adding a New GPMF Field
1. Identify the GPMF key (e.g., GPSU, STMP, GPS5)
2. Add parsing in `parseGPMFData()` in `gpmf.go`
3. Add field to `GPSData` struct in `validator.go`
4. Update comparator logic if needed

### Adding a New Validation Check
1. Add logic to `compareMetadata()` in `comparator.go`
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
