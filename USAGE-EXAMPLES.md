# Usage Examples

## Basic Usage

### Validate Files in Current Directory

```bash
cd /path/to/your/gopro/videos
/path/to/gopro-validator
```

The tool will scan the current directory and all subdirectories for MP4 files.

### Validate Files in Specific Directory

```bash
gopro-validator --input /path/to/videos
```

### Validate Sample Files (Testing)

```bash
# From project root
make go

# Or directly
cd go-validator
./gopro-validator --input ../sample-input-files
```

## Real-World Workflows

### Scenario 1: Organize Your Entire GoPro Library

You have a messy folder with years of GoPro videos.

```bash
# 1. Navigate to your GoPro folder
cd ~/Videos/GoPro-Archive

# 2. Preview what would be fixed
gopro-validator --rename --dry-run

# 3. Create organized library
gopro-validator --rename --output ~/Videos/GoPro-Organized
```

**Result:**
```
~/Videos/GoPro-Organized/
├── 2020-04-01/
│   ├── 20200401_192625_GH016252.MP4
│   └── 20200401_204007_GH016254.MP4
├── 2021-04-04/
│   └── 20210404_122305_GH016761.MP4
└── 2024-02-22/
    └── ...
```

### Scenario 2: Fix Files with Wrong Metadata

You notice files have incorrect dates.

```bash
# 1. Check what's wrong
gopro-validator --input ~/Videos/BadDates

# 2. Preview fixes
gopro-validator --input ~/Videos/BadDates --update-metadata --dry-run

# 3. Backup first!
cp -r ~/Videos/BadDates ~/Videos/BadDates-Backup

# 4. Apply metadata fixes
gopro-validator --input ~/Videos/BadDates --update-metadata

# 5. Verify
gopro-validator --input ~/Videos/BadDates
```

### Scenario 3: Create Full Recordings from Chapters

You have multi-part recordings (GH01/02/03 files).

```bash
# 1. See what chapters exist
gopro-validator --input ~/Videos/Chapters --concat --dry-run

# 2. Create full recordings
gopro-validator --input ~/Videos/Chapters --concat --concat-output ~/Videos/Full-Recordings

# 3. Optionally organize the originals too
gopro-validator --input ~/Videos/Chapters --rename --output ~/Videos/Chapters-Organized
```

**Result:**
```
~/Videos/Full-Recordings/
├── 20240222_165742_GH6978_FULL.MP4  (5.7 GB - all chapters combined)
├── 20240222_171708_GH6979_FULL.MP4  (7.0 GB)
└── 20240222_183345_GH6980_FULL.MP4  (8.5 GB)
```

### Scenario 4: Complete Camera Card Workflow

You just finished a day of shooting.

```bash
# 1. Copy files from SD card
cp -r /media/GOPRO/DCIM/100GOPRO ~/Videos/Today

# 2. Validate and check for issues
gopro-validator --input ~/Videos/Today

# 3. Fix metadata if needed
gopro-validator --input ~/Videos/Today --update-metadata

# 4. Create full recordings
gopro-validator --input ~/Videos/Today --concat --concat-output ~/Videos/Full

# 5. Organize by date
gopro-validator --input ~/Videos/Today --rename --output ~/Videos/Organized

# Now you have:
# - Original files (unchanged): ~/Videos/Today
# - Full recordings: ~/Videos/Full
# - Organized copies: ~/Videos/Organized
```

### Scenario 5: Quick Check Before Editing

Before importing to video editor:

```bash
# Quick validation
gopro-validator --input ~/Videos/ProjectX

# If issues found, fix them
gopro-validator --input ~/Videos/ProjectX --update-metadata

# Concat chapters if needed
gopro-validator --input ~/Videos/ProjectX --concat
```

## Working with Different Directory Structures

### Flat Directory

```
Videos/
├── GH016978.MP4
├── GH026978.MP4
├── GH016979.MP4
└── ...
```

```bash
gopro-validator --input Videos --concat
```

### Date-Based Organization

```
Videos/
├── 2024-02-22/
│   ├── GH016978.MP4
│   └── GH026978.MP4
└── 2024-02-23/
    └── GH016980.MP4
```

```bash
gopro-validator --input Videos
# Automatically scans all subdirectories
```

### Camera-Based Organization

```
Videos/
├── HERO7/
│   └── 2024-02-22/
├── HERO9/
│   └── 2024-02-23/
└── HERO11/
```

```bash
# Process all cameras
gopro-validator --input Videos --rename --output Videos-Organized

# Process one camera
gopro-validator --input Videos/HERO7
```

## Custom Output Locations

### Use Existing Folder Structure

```bash
# Organize within same parent folder
gopro-validator --input ~/Videos/Raw --rename --output ~/Videos/Organized
```

### Network/External Drive

```bash
# Copy to NAS
gopro-validator --input ~/Videos/GoPro --rename --output /mnt/nas/Videos/GoPro

# Copy to external drive
gopro-validator --input ~/Videos/GoPro --rename --output /media/external/GoPro-Backup
```

### Temporary Working Directory

```bash
# Work in temp, then move
gopro-validator --input ~/Videos/Raw --concat --concat-output /tmp/concat
# Review the files, then move
mv /tmp/concat/* ~/Videos/Final/
```

## Combining Operations

### Fix Everything at Once

```bash
gopro-validator --input ~/Videos/Raw \
  --update-metadata \
  --rename --output ~/Videos/Organized \
  --concat --concat-output ~/Videos/Full \
  --dry-run  # Remove this to actually do it
```

This will:
1. Fix metadata in place
2. Create organized copies
3. Create concatenated full recordings

### Selective Operations

```bash
# Only concat, don't rename
gopro-validator --input ~/Videos --concat

# Only rename, don't concat
gopro-validator --input ~/Videos --rename

# Fix metadata, then organize
gopro-validator --input ~/Videos --update-metadata
gopro-validator --input ~/Videos --rename --output ~/Videos/Fixed
```

## Testing and Safety

### Always Dry-Run First

```bash
# Preview all changes
gopro-validator --input ~/Videos \
  --rename --update-metadata --concat \
  --dry-run
```

### Test on Small Sample

```bash
# Create test folder with a few files
mkdir ~/Videos/test
cp ~/Videos/BigLibrary/GH01*.MP4 ~/Videos/test/

# Test workflow
gopro-validator --input ~/Videos/test --rename --dry-run
gopro-validator --input ~/Videos/test --rename --output ~/Videos/test-organized
```

### Backup Before Metadata Changes

```bash
# Metadata updates modify files in place!
rsync -av ~/Videos/Original/ ~/Videos/Original-Backup/

gopro-validator --input ~/Videos/Original --update-metadata
```

## Integration with Other Tools

### With rsync (Backup After Processing)

```bash
gopro-validator --input ~/Videos/Raw --rename --output ~/Videos/Organized
rsync -av ~/Videos/Organized/ user@backup:/Videos/
```

### With find (Process Specific Files)

```bash
# Only process files from February
find ~/Videos -name "GH*.MP4" -newermt 2024-02-01 ! -newermt 2024-03-01 \
  -exec dirname {} \; | sort -u | \
  xargs -I {} gopro-validator --input {}
```

### With File Manager

```bash
# Right-click "Open in Terminal" on a folder, then:
gopro-validator --rename --output ../Organized
```

## Makefile Shortcuts (Project Development)

If you're in the project directory:

```bash
# Validate sample files
make go

# Rename sample files
make rename-dry    # Preview
make rename        # Actually do it

# Update metadata
make update-dry    # Preview
make update        # Actually do it

# Concatenate
make concat-dry    # Preview
make concat        # Actually do it

# Everything
make fix-all-dry   # Preview
make fix-all       # Do it
```

## Tips

### Check Disk Space First

```bash
# Before concatenating large files
df -h ~/Videos
```

### Process in Batches

```bash
# For very large libraries, process by year
for year in 2020 2021 2022 2023 2024; do
  gopro-validator --input ~/Videos/$year --rename --output ~/Videos/Organized
done
```

### Keep Originals Safe

```bash
# Copy, don't move
gopro-validator --input ~/Videos/SD-Card --rename --output ~/Videos/Library
# SD-Card files remain untouched

# Only metadata updates modify originals
```

### Use Absolute Paths

```bash
# Safer than relative paths
gopro-validator --input /home/user/Videos/GoPro --output /home/user/Videos/Organized

# Rather than
gopro-validator --input ../../Videos --output ../Organized  # Can be confusing
```
