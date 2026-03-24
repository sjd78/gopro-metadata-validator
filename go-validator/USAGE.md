# GoPro Validator - Usage Guide

## Basic Validation

Validate files without making any changes:

```bash
./gopro-validator
```

This will:
- Scan all MP4 files in `../input-files`
- Extract GPS timestamps from GPMF stream
- Compare with file metadata
- Report any discrepancies

## Renaming Files Based on GPS Time

### Dry Run (Recommended First)

See what would be renamed without making changes:

```bash
./gopro-validator --rename --dry-run
```

Example output:
```
📋 Would rename:
   From: ../input-files/2015-12-31/HERO7 Black 1/GH016761.MP4
   To:   ../renamed-files/2021-04-04/20210404_122305_GH016761.MP4
```

### Actually Rename Files

Once you're happy with the plan:

```bash
./gopro-validator --rename
```

This will:
- Copy files to `../renamed-files/YYYY-MM-DD/` folders
- Rename to `YYYYMMDD_HHMMSS_original.MP4` format
- Use GPS UTC time for naming
- Keep original files intact (copies, doesn't move)

### Custom Output Directory

```bash
./gopro-validator --rename --output /path/to/output
```

## Updating File Metadata

### Dry Run (Recommended First)

See what metadata would be updated:

```bash
./gopro-validator --update-metadata --dry-run
```

Example output:
```
📋 Would update metadata:
   File: GH016761.MP4
   Current: 2016-01-01T00:10:36Z
   New:     2021-04-04T12:23:05Z
```

### Actually Update Metadata

⚠️ **WARNING:** This modifies the original MP4 files!

```bash
./gopro-validator --update-metadata
```

This will:
- Update the `creation_time` metadata in each MP4
- Use GPS UTC time from GPMF stream
- Create temporary file during update
- Replace original only if successful
- Skip files where metadata already matches

**Recommendation:** Make backups before running without `--dry-run`!

## Combined Operations

You can combine operations:

```bash
# Dry run both rename and metadata update
./gopro-validator --rename --update-metadata --dry-run

# Actually perform both
./gopro-validator --rename --update-metadata
```

## Workflow Recommendations

### For Production Use:

1. **First, validate:**
   ```bash
   ./gopro-validator
   ```

2. **See what would change:**
   ```bash
   ./gopro-validator --rename --update-metadata --dry-run
   ```

3. **Backup originals:**
   ```bash
   cp -r ../input-files ../input-files-backup
   ```

4. **Update metadata in place:**
   ```bash
   ./gopro-validator --update-metadata
   ```

5. **Create renamed copies:**
   ```bash
   ./gopro-validator --rename --output ~/Videos/GoPro-Organized
   ```

### For Testing:

Always use `--dry-run` first!

```bash
./gopro-validator --rename --dry-run
```

## Understanding the Output

### Validation Results

```
✓ GH016978.MP4                    # ✓ = Valid, ✗ = Issues found
  Metadata Creation Time: 2024-02-22T11:57:42Z
  Timecode: 11:56:59:52
  GPS Samples: 5851
  GPS Absolute Time (first): 2024-02-22T16:57:43Z    # Ground truth from GPS
  GPS Absolute Time (last):  2024-02-22T17:06:34Z
  GPS Relative Start: 0.198s                         # Milliseconds since recording
  GPS Relative End: 16.0s
  Issues:
    - ⚠️  GPS time differs from file creation time by 5 hours
```

### Common Issues

**"GPS time differs by 5 hours"**
- Likely timezone issue
- Camera saved local time but marked as UTC
- Use `--update-metadata` to fix

**"GPS first timestamp is 105.6s - likely a chapter continuation"**
- This is normal for GH02/GH03 files (multi-part recordings)
- Not an error, just informational

**"GPS time differs by 5 years"**
- Camera date was set incorrectly when recording
- Use `--update-metadata` AND `--rename` to fix

## Safety Features

- `--dry-run` shows changes without modifying anything
- `--rename` COPIES files (originals untouched)
- `--update-metadata` creates temp file first, only replaces on success
- Skips files that already have correct metadata
- Clear summary of what was changed

## Examples

### Fix timezone issues in 2024 videos:

```bash
# See what would change
./gopro-validator --update-metadata --dry-run

# Apply fixes
./gopro-validator --update-metadata
```

### Organize all videos by correct date:

```bash
# Create properly organized library
./gopro-validator --rename --output ~/Videos/GoPro-Library

# Result:
# ~/Videos/GoPro-Library/
#   2020-04-01/
#     20200401_192625_GH016252.MP4
#     20200401_204007_GH016254.MP4
#   2021-04-04/
#     20210404_122305_GH016761.MP4
#   2024-02-22/
#     20240222_165743_GH016978.MP4
#     ...
```

### Complete workflow:

```bash
# 1. Backup
cp -r input-files input-files-backup

# 2. Preview changes
./gopro-validator --rename --update-metadata --dry-run

# 3. Fix metadata
./gopro-validator --update-metadata

# 4. Create organized library
./gopro-validator --rename --output ~/Videos/GoPro-Organized

# 5. Verify
./gopro-validator
```
