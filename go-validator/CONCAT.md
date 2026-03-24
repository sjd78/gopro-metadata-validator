# Chapter Concatenation

## What Are Chapters?

GoPro cameras split long recordings into multiple files (chapters) to avoid file system limitations:
- **FAT32 limit**: 4GB max file size
- **Time-based**: Some models split at specific time intervals

Chapter files are named sequentially:
```
GH016978.MP4  ← Chapter 1
GH026978.MP4  ← Chapter 2
GH036978.MP4  ← Chapter 3 (if recording continued)
```

All chapters have the same base number (`6978`) and increment the first two digits.

## Why Concatenate?

**Problems with chapters:**
- Hard to watch continuously (have to open multiple files)
- Editing software may not recognize them as related
- Metadata only in first chapter
- Difficult to share complete recordings

**Benefits of concatenation:**
- Single file for entire recording
- Easier to edit and share
- All tracks preserved (video, audio, GPMF, timecode)
- Proper duration and metadata

## How It Works

### Automatic Detection

The tool automatically:
1. Scans all files for GoPro chapter naming pattern
2. Groups files by base number
3. Sorts chapters in order
4. Detects only multi-chapter recordings (ignores single files)

### Concatenation Process

Uses `ffmpeg` concat demuxer with codec copy:
```bash
ffmpeg -f concat -safe 0 -i filelist.txt -c copy output.mp4
```

**Why codec copy?**
- No re-encoding (fast, lossless)
- Preserves all streams exactly
- Maintains GPMF telemetry data
- Keeps original quality

### All Tracks Preserved

The concatenated file includes:
- ✅ Video stream (H.264)
- ✅ Audio stream (AAC)
- ✅ Timecode track
- ✅ GPMF telemetry (GPS, accelerometer, gyro)
- ✅ Metadata

## Usage

### Preview What Would Be Concatenated

```bash
make concat-dry
# or
./gopro-validator --concat --dry-run
```

Example output:
```
📋 Would concatenate 3 chapters:
   [1] GH016980.MP4
   [2] GH026980.MP4
   [3] GH036980.MP4
   → Output: 20240222_183345_GH6980_FULL.MP4
```

### Actually Concatenate

```bash
make concat
# or
./gopro-validator --concat
```

Output:
```
🔗 Concatenating 3 chapters for recording 6980...
✓ Created: 20240222_183345_GH6980_FULL.MP4 (11.2 GB)
```

### Custom Output Directory

```bash
./gopro-validator --concat --concat-output ~/Videos/Full-Recordings
```

## Output Naming

Files are named with GPS-adjusted start time:
```
YYYYMMDD_HHMMSS_GHxxxx_FULL.MP4
```

Examples:
```
20240222_165742_GH6978_FULL.MP4  ← Recording 6978 started at 16:57:42 UTC
20240222_171708_GH6979_FULL.MP4  ← Recording 6979 started at 17:17:08 UTC
20240222_183345_GH6980_FULL.MP4  ← Recording 6980 started at 18:33:45 UTC
```

The timestamp is adjusted for GPS lock delay (see [GPS-OFFSET-FIX.md](GPS-OFFSET-FIX.md)).

## Chapter Detection Example

From your sample files:

### Recording 6978 (2 chapters)
```
GH016978.MP4  ← 3.7 GB (first 8.9 minutes)
GH026978.MP4  ← 2.0 GB (next 3.5 minutes)
───────────────────────────────────────
TOTAL: 5.7 GB, 12.4 minutes
→ 20240222_165742_GH6978_FULL.MP4
```

### Recording 6979 (2 chapters)
```
GH016979.MP4  ← 3.7 GB
GH026979.MP4  ← 3.3 GB
───────────────────────────────────────
TOTAL: 7.0 GB
→ 20240222_171708_GH6979_FULL.MP4
```

### Recording 6980 (3 chapters)
```
GH016980.MP4  ← 3.7 GB
GH026980.MP4  ← 3.7 GB
GH036980.MP4  ← 1.1 GB
───────────────────────────────────────
TOTAL: 8.5 GB
→ 20240222_183345_GH6980_FULL.MP4
```

## GPMF Telemetry Preservation

The concatenated file preserves all GPMF data from all chapters:

**Before (3 separate files):**
```
GH016980.MP4 - GPMF: 0s to 531s
GH026980.MP4 - GPMF: 531s to 1062s
GH036980.MP4 - GPMF: 1062s to 1170s
```

**After (1 concatenated file):**
```
20240222_183345_GH6980_FULL.MP4 - GPMF: 0s to 1170s (complete timeline)
```

You can still extract GPS data, accelerometer data, etc. from the concatenated file.

## Verification

After concatenation, you can verify:

### Check streams are preserved:
```bash
ffprobe -show_streams 20240222_183345_GH6980_FULL.MP4
```

Should show:
- Stream 0: Video (h264)
- Stream 1: Audio (aac)
- Stream 2: Data (tmcd) - Timecode
- Stream 3: Data (gpmd) - GoPro Metadata
- Stream 4: Data (fdsc)

### Check duration:
```bash
ffprobe -show_format 20240222_183345_GH6980_FULL.MP4 | grep duration
```

Should be sum of all chapter durations.

### Validate GPS data:
```bash
./gopro-validator
```

The concatenated file will have:
- GPS samples from all chapters
- Continuous GPMF timeline
- Correct total duration

## Combining with Other Operations

You can combine concatenation with other operations:

### Concat and organize:
```bash
./gopro-validator --concat --rename
```

Creates:
- `concatenated-files/` - Full recordings
- `renamed-files/` - Individual chapters organized by date

### Complete workflow:
```bash
# 1. Preview everything
./gopro-validator --concat --rename --update-metadata --dry-run

# 2. Apply all fixes
./gopro-validator --concat --rename --update-metadata
```

Result:
- Chapter files have corrected metadata
- Chapter files organized by date
- Full concatenated recordings available

## Troubleshooting

### "No multi-chapter recordings found"

The tool only concatenates files with multiple chapters. Single files are skipped.

Check your filenames match the pattern:
- ✅ `GH016978.MP4` and `GH026978.MP4` → Will concatenate
- ❌ `GH016978.MP4` alone → Skipped (single file)

### Concatenation fails

Common issues:
- **Different codecs**: Chapters must have same video/audio codecs
- **Corrupted file**: One chapter is damaged
- **Disk space**: Need space for output file (sum of all chapters)

Check ffmpeg output for specific errors.

### GPMF data missing after concat

If you used `-c:v copy -c:a copy` instead of `-c copy`:
- This copies only video and audio
- **Use `-c copy`** to copy ALL streams including GPMF

The tool uses `-c copy` by default.

## Performance

Concatenation is fast because it uses codec copy (no re-encoding):

| Total Size | Time |
|-----------|------|
| 5 GB (2 chapters) | ~10 seconds |
| 8 GB (3 chapters) | ~15 seconds |
| 15 GB (4 chapters) | ~25 seconds |

Actual time depends on disk speed.

## Safety

- ✅ Original files are **never modified**
- ✅ Creates new concatenated files
- ✅ Dry-run mode shows what would be created
- ✅ Can delete concatenated files and re-run if needed

## Example Workflow

```bash
# 1. See what chapters exist
make concat-dry

# 2. Create concatenated files
make concat

# 3. Verify they're correct
./gopro-validator

# 4. Optional: Delete original chapters if you only want full recordings
# (Make backups first!)
```

## Technical Details

### ffmpeg concat demuxer

The tool creates a temporary file list:
```
file '/path/to/GH016980.MP4'
file '/path/to/GH026980.MP4'
file '/path/to/GH036980.MP4'
```

Then runs:
```bash
ffmpeg -f concat -safe 0 -i filelist.txt -c copy output.mp4
```

**Flags:**
- `-f concat` - Use concat demuxer (better than filter for identical streams)
- `-safe 0` - Allow absolute paths
- `-c copy` - Copy all streams without re-encoding

### Why concat demuxer?

Better than concat filter because:
- ✅ Preserves all metadata
- ✅ Maintains exact timestamps
- ✅ Copies GPMF data intact
- ✅ Much faster (no decode/encode)
- ✅ Lossless
