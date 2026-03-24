# Complete Feature List

## 1. Validation ✅

**What it does:** Compares file metadata against GPS ground truth

```bash
make go
```

**Checks:**
- GPS absolute time vs file creation time
- Timecode vs creation time consistency
- GPS data availability and validity
- Relative timestamp progression

**Output:**
- ✓ Valid files (metadata matches GPS)
- ✗ Issues found (with detailed explanations)
- Summary of all problems

---

## 2. Metadata Update ✅

**What it does:** Fixes file `creation_time` to match GPS timestamps

```bash
# Preview
make update-dry

# Apply
make update
```

**Features:**
- Adjusts for GPS lock delay
- Updates metadata in-place
- Creates temp file for safety
- Skips already-correct files
- Preserves all video/audio data

**Use case:** Fix timezone issues and incorrect dates

---

## 3. File Renaming ✅

**What it does:** Organizes files by GPS timestamp into dated folders

```bash
# Preview
make rename-dry

# Apply
make rename
```

**Output structure:**
```
renamed-files/
├── 2020-04-01/
│   ├── 20200401_192625_GH016252.MP4
│   └── 20200401_204007_GH016254.MP4
├── 2021-04-04/
│   └── 20210404_122305_GH016761.MP4
└── 2024-02-22/
    ├── 20240222_165742_GH016978.MP4
    └── ...
```

**Features:**
- GPS-adjusted timestamps (accounts for lock delay)
- Organized by date folders
- Descriptive filenames with timestamp
- Copies files (originals untouched)
- Custom output directory support

**Use case:** Create organized video library with accurate timestamps

---

## 4. Chapter Concatenation ✅ NEW!

**What it does:** Joins multi-chapter recordings into single complete files

```bash
# Preview
make concat-dry

# Apply
make concat
```

**How it works:**
- Automatically detects chapter series (GH01/02/03xxxx)
- Groups files by base number
- Concatenates using ffmpeg codec copy
- Preserves ALL tracks (video, audio, GPMF, timecode)
- No re-encoding (fast, lossless)

**Example:**
```
Input:
  GH016980.MP4 (3.7 GB)
  GH026980.MP4 (3.7 GB)
  GH036980.MP4 (1.1 GB)

Output:
  20240222_183345_GH6980_FULL.MP4 (8.5 GB)
```

**Features:**
- Maintains GPMF telemetry data
- GPS-adjusted start time in filename
- All streams preserved exactly
- Original quality maintained
- Fast (no re-encoding)

**Use case:**
- Watch complete recordings in one file
- Easier editing workflows
- Simpler sharing
- Maintain GPS data timeline

See [go-validator/CONCAT.md](go-validator/CONCAT.md) for details.

---

## 5. GPS Lock Delay Adjustment ✅

**What it does:** Automatically adjusts timestamps for GPS lock delay

**The problem:**
```
Recording starts:     12:00:00
GPS locks 5s later:   12:00:05
Without adjustment:   File timestamped 12:00:05 ❌
With adjustment:      File timestamped 12:00:00 ✅
```

**How it works:**
```
Actual Start = GPS Sample Time - GPS Relative Offset
```

**Applied to:**
- File renaming
- Metadata updates
- Validation comparisons

**Impact:**
- Chapter files would be MINUTES off without this
- Ensures accurate timestamps
- Maintains chronological sorting

See [go-validator/GPS-OFFSET-FIX.md](go-validator/GPS-OFFSET-FIX.md) for technical details.

---

## Combined Operations

You can combine multiple operations:

### Complete Fix
```bash
make fix-all-dry    # Preview rename + update
make fix-all        # Apply both
```

### Everything
```bash
./gopro-validator --rename --update-metadata --concat --dry-run
./gopro-validator --rename --update-metadata --concat
```

Result:
- ✅ Metadata corrected
- ✅ Files organized by date
- ✅ Chapters concatenated
- ✅ Complete organized library

---

## Safety Features

All operations include:

- **--dry-run** - Preview before applying
- **Non-destructive** - Originals never modified (except --update-metadata)
- **Error handling** - Graceful failure with clear messages
- **Validation** - Can verify results afterwards
- **Atomic updates** - Temp files used for metadata changes

---

## Command Reference

| Command | What it does |
|---------|--------------|
| `make go` | Validate all files |
| `make rename-dry` | Preview file organization |
| `make rename` | Organize files by GPS time |
| `make update-dry` | Preview metadata fixes |
| `make update` | Fix file metadata |
| `make concat-dry` | Preview chapter concatenation |
| `make concat` | Concatenate chapters |
| `make fix-all-dry` | Preview rename + update |
| `make fix-all` | Apply rename + update |

Or use directly:
```bash
cd go-validator
./gopro-validator [options]

Options:
  --rename              Rename files by GPS time
  --update-metadata     Fix metadata
  --concat              Concatenate chapters
  --dry-run            Preview only
  --output DIR         Output for renamed files
  --concat-output DIR  Output for concatenated files
```

---

## Use Cases

### 1. Fix One File with Wrong Date
```bash
./gopro-validator                # See the error
make update                      # Fix metadata
make rename                      # Move to correct date folder
```

### 2. Organize Entire Library
```bash
make rename --output ~/Videos/GoPro-Library
```

### 3. Create Full Recordings
```bash
make concat --concat-output ~/Videos/Full-Recordings
```

### 4. Complete Workflow
```bash
# 1. Backup
cp -r input-files input-files-backup

# 2. Preview
make fix-all-dry
make concat-dry

# 3. Apply
make update          # Fix metadata in place
make rename          # Organize by date
make concat          # Create full recordings

# 4. Verify
make go             # Should show all valid now
```

---

## Performance

| Operation | 10 Files (38GB) | Notes |
|-----------|----------------|-------|
| Validation | ~2 seconds | Fast metadata scanning |
| Metadata Update | ~15 seconds | ffmpeg remux |
| Rename | ~30 seconds | Copy operations |
| Concat (3 chapters) | ~15 seconds | Codec copy, no encode |

All operations handle files of any size (tested up to 4GB per file).

---

## What Makes This Tool Unique

1. **GPS Truth** - Uses actual GPS satellite time as ground truth
2. **Lock Delay Adjustment** - Accounts for GPS acquisition time
3. **Chapter Handling** - Properly processes multi-part recordings
4. **All Streams Preserved** - Maintains GPMF telemetry data
5. **Large File Support** - Streams GPMF, handles files >2GB
6. **Zero Re-encoding** - Preserves original quality
7. **Automatic Detection** - Finds chapter series automatically

---

## Documentation

- **[README.md](README.md)** - Project overview
- **[SUMMARY.md](SUMMARY.md)** - Quick start guide
- **[DISCOVERIES.md](DISCOVERIES.md)** - Errors found in sample files
- **[go-validator/USAGE.md](go-validator/USAGE.md)** - Detailed usage
- **[go-validator/CONCAT.md](go-validator/CONCAT.md)** - Chapter concatenation
- **[go-validator/GPS-OFFSET-FIX.md](go-validator/GPS-OFFSET-FIX.md)** - Lock delay technical details
