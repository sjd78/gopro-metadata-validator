# Changelog

## 2026-03-24 (Update 3)

### Project Focus
- **Archived** TypeScript version - marked as reference implementation
- **All development** now focused exclusively on Go version
- **Updated** documentation to reflect Go as the production implementation
- **Added** ARCHIVED.md in ts-validator explaining the decision

### Rationale
- Go version is feature-complete and production-ready
- TypeScript version has fundamental limitations (2GB file size)
- Maintaining one implementation ensures quality and focus
- TypeScript code remains as reference for library usage

## 2026-03-24 (Update 2)

### CLI Improvements
- **Added** `--input` flag to specify input directory (defaults to current directory)
- **Changed** Default output directories to be relative to current directory
  - `--output` now defaults to `renamed-files/` (was `../renamed-files/`)
  - `--concat-output` now defaults to `concatenated-files/` (was `../concatenated-files/`)
- **Renamed** `input-files/` to `sample-input-files/` for clarity
- **Updated** Makefile to use `--input ../sample-input-files` for backward compatibility
- **Updated** All documentation with new usage patterns

### Benefits
- **More flexible:** Can now run the tool from any directory on your videos
- **Standard behavior:** Defaults to current directory like most CLI tools
- **Backward compatible:** Makefile commands still work with sample files

### Migration Guide
```bash
# Old way (hardcoded path)
cd go-validator
./gopro-validator

# New way (specify your videos)
./gopro-validator --input /path/to/your/videos

# Or use current directory
cd /path/to/your/videos
/path/to/gopro-validator
```

## 2026-03-24 (Update 1)

### Project Reorganization
- **Moved** TypeScript code to `ts-validator/` directory for better organization
- **Updated** Makefile to reference new structure
- **Updated** `.gitignore` for new directory structure
- **Added** README.md in `ts-validator/` directory

### Windows Compatibility
- **Fixed** Windows path handling in concat feature
  - Issue: Windows backslashes weren't working with ffmpeg
  - Solution: Added `filepath.ToSlash()` conversion
- **Added** comprehensive Windows documentation ([WINDOWS.md](WINDOWS.md))
- **Added** platform compatibility matrix ([PLATFORM-COMPATIBILITY.md](PLATFORM-COMPATIBILITY.md))

### New Features
- **Added** Chapter concatenation feature
  - Automatically detects multi-chapter recordings
  - Joins chapters into single complete files
  - Preserves all tracks including GPMF telemetry
  - Uses ffmpeg codec copy (fast, lossless)
  - See [go-validator/CONCAT.md](go-validator/CONCAT.md)

### GPS Improvements
- **Fixed** GPS lock delay adjustment
  - Accounts for time between recording start and GPS lock
  - Critical for chapter files (would be minutes off without this)
  - See [go-validator/GPS-OFFSET-FIX.md](go-validator/GPS-OFFSET-FIX.md)
- **Added** GPS absolute timestamp extraction from GPMF stream
  - Reads GPSU/GPSUU entries containing actual UTC time
  - Compares against file metadata for validation
  - Enables accurate metadata corrections

### Documentation Improvements
- **Added** [FEATURES.md](FEATURES.md) - Complete feature reference
- **Added** [WINDOWS.md](WINDOWS.md) - Windows setup and usage guide
- **Added** [PLATFORM-COMPATIBILITY.md](PLATFORM-COMPATIBILITY.md) - Cross-platform compatibility info
- **Added** [go-validator/CONCAT.md](go-validator/CONCAT.md) - Chapter concatenation details
- **Added** [go-validator/GPS-OFFSET-FIX.md](go-validator/GPS-OFFSET-FIX.md) - GPS offset technical details
- **Added** [go-validator/USAGE.md](go-validator/USAGE.md) - Detailed usage guide
- **Updated** All documentation to reflect new directory structure

### Makefile
- **Added** `make concat-dry` - Preview chapter concatenation
- **Added** `make concat` - Actually concatenate chapters
- **Updated** `make clean` to handle new directory structure
- **Updated** `make ts` to run from `ts-validator/` directory

## Initial Release (2026-03-22 to 2026-03-23)

### Core Features
- GPS-based metadata validation
- Metadata update capability
- File renaming and organization
- Support for files of any size (Go version)
- TypeScript version for smaller files

### Validation
- Compares file metadata against GPS ground truth
- Detects timezone issues
- Identifies incorrect dates
- Reports chapter continuations

### Operations
- **Rename** - Organize files by GPS timestamp
- **Update Metadata** - Fix creation_time using GPS data
- **Dry-run mode** - Preview all changes before applying

### Technical
- Custom GPMF parser for large files
- Streams via ffmpeg (doesn't load entire file)
- Cross-platform Go implementation
- TypeScript version using established libraries

### Documentation
- Comprehensive README
- Detailed discoveries from sample files
- Results documentation
- Summary and quick start guide
