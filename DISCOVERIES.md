# Key Discoveries

## GPS Timestamps in GPMF Stream

The GPMF stream contains **absolute GPS UTC timestamps** in the format:
```
240222170635.690  →  2024-02-22 17:06:35.690 UTC
```

These are stored in the `GPSU` or `GPSUU` KLV entries and represent the **ground truth** recording time.

## Metadata Errors Found

### Critical Error: Completely Wrong Date

**File:** `GH016761.MP4`
- **Folder:** `2015-12-31/`
- **File metadata:** 2016-01-01 00:10:36 UTC
- **GPS says:** 2021-04-04 12:23:05 UTC
- **Error:** File is **5 years off!** Should be in a 2021-04-04 folder

### Timezone Issues: Wrong Timezone Tag

**All 2024-02-22 files** have a systematic 5-hour offset:
- **File metadata:** Times like 11:57, 12:17, 13:33 (marked as UTC)
- **GPS UTC time:** Times like 16:57, 17:17, 18:33 (true UTC)
- **Root cause:** Camera recorded in **Eastern Time (UTC-5)** but tagged times as UTC

Example:
```
GH016978.MP4
  File says:  2024-02-22T11:57:42Z  ← incorrectly marked as UTC
  GPS says:   2024-02-22T16:57:43Z  ← actual UTC
  Reality:    2024-02-22 11:57 EST = 16:57 UTC
```

**All 2020-04-01 files** have a 4-hour offset (EDT = UTC-4, daylight saving time)

### The Issue

GoPro cameras sometimes save the **local time** in metadata but incorrectly mark it as UTC. The GPS stream always has the correct UTC time because it comes from GPS satellites.

## Solution

The tool can now:

1. **Detect** discrepancies between file metadata and GPS timestamps
2. **Rename/move** files based on GPS truth (correct dates and folders)
3. **Update** metadata in MP4 files to match GPS timestamps

## Data Integrity

| File | Original Metadata | GPS Truth | Status |
|------|------------------|-----------|---------|
| GH016761.MP4 | 2016-01-01 | **2021-04-04** | ❌ Wrong date |
| 2020 files | Correct date | Correct date | ⚠️ Wrong timezone |
| 2024 files | Correct date | Correct date | ⚠️ Wrong timezone |

## Why This Matters

**Before discovery:**
- Files organized by incorrect dates
- Metadata doesn't match reality
- Can't trust file timestamps for sorting/organizing

**After using this tool:**
- Files organized by true GPS recording time
- Metadata matches GPS ground truth
- Reliable timestamps for all operations

## How the Fix Works

### Renaming
```
Before: input-files/2015-12-31/GH016761.MP4
After:  renamed-files/2021-04-04/20210404_122305_GH016761.MP4
```

### Metadata Update
```
Before: creation_time=2016-01-01T00:10:36Z (wrong)
After:  creation_time=2021-04-04T12:23:05Z (GPS truth)
```

## Chapter Files

GoPro splits long recordings into chapters (GH01, GH02, GH03, etc.).

The tool correctly identifies:
- **GH01xxxx** - First chapter (GPS starts at ~0s)
- **GH02xxxx** - Second chapter (GPS continues from previous, ~105s)
- **GH03xxxx** - Third chapter (GPS continues, ~210s)

This is expected behavior and not an error.

## Command Examples

### See what's wrong:
```bash
./gopro-validator
```

### Preview fixes:
```bash
./gopro-validator --rename --update-metadata --dry-run
```

### Fix the file with wrong date:
```bash
# Update metadata
./gopro-validator --update-metadata

# Organize into correct folder
./gopro-validator --rename
```

Result:
- Metadata corrected from 2016 → 2021
- File moved to `2021-04-04/` folder
- Renamed to `20210404_122305_GH016761.MP4`

### Fix timezone issues on 2024 files:
```bash
# Just update the metadata times
./gopro-validator --update-metadata

# Or create properly organized copies
./gopro-validator --rename --output ~/Videos/GoPro-Organized
```

Result:
- Times corrected (11:57 → 16:57 UTC)
- Files still in 2024-02-22 folder (date was already correct)

## Technical Details

### GPMF Structure
```
DEVC (Device container)
  └─ STRM (Stream)
      ├─ GPS5 (GPS coordinates)
      ├─ GPSU/GPSUU (GPS UTC datetime string)  ← Ground truth!
      ├─ TSMP (Relative timestamp in ms)
      └─ ACCL/GYRO (Sensor data)
```

### Parsing Method
1. Extract GPMF stream via `ffmpeg` (doesn't load entire file)
2. Parse KLV (Key-Length-Value) structure recursively
3. Extract `GPSU`/`GPSUU` entries containing datetime strings
4. Parse format: `YYMMDDHHMMSS.sss`
5. Compare with file metadata `creation_time`

### Why ffmpeg + Custom Parser
- **gpmf-extract** library can't handle files >2GB (loads into memory)
- **ffmpeg** can extract just the GPMF stream (few MB) from any size file
- Custom parser processes the extracted stream efficiently
