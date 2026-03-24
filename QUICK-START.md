# Quick Start Guide

## For Your Own Videos

### 1. Build the Tool

```bash
cd go-validator
go build -o gopro-validator
```

### 2. Run on Your Videos

```bash
# Option A: Navigate to your video folder
cd /path/to/your/gopro/videos
/path/to/gopro-validator

# Option B: Specify the path
./gopro-validator --input /path/to/your/gopro/videos
```

### 3. Common Operations

**Check for issues:**
```bash
./gopro-validator --input /path/to/videos
```

**Fix metadata:**
```bash
./gopro-validator --input /path/to/videos --update-metadata --dry-run  # Preview
./gopro-validator --input /path/to/videos --update-metadata            # Apply
```

**Organize files:**
```bash
./gopro-validator --input /path/to/videos --rename --dry-run           # Preview
./gopro-validator --input /path/to/videos --rename                     # Apply
```

**Concatenate chapters:**
```bash
./gopro-validator --input /path/to/videos --concat --dry-run           # Preview
./gopro-validator --input /path/to/videos --concat                     # Apply
```

## Testing with Sample Files

The project includes sample GoPro files for testing:

```bash
# From project root
make go              # Validate sample files
make rename-dry      # Preview renaming
make concat-dry      # Preview concatenation
make update-dry      # Preview metadata updates
```

## Safety Tips

1. **Always use --dry-run first**
   ```bash
   ./gopro-validator --input /path --rename --dry-run
   ```

2. **Backup before metadata updates**
   ```bash
   cp -r /path/to/videos /path/to/videos-backup
   ./gopro-validator --input /path/to/videos --update-metadata
   ```

3. **Rename copies files (safe)**
   - Original files are NOT modified
   - Creates organized copies in output directory

4. **Update-metadata modifies originals**
   - Changes files in place
   - Make backups first!

## What Each Operation Does

| Operation | What it does | Modifies originals? |
|-----------|--------------|---------------------|
| Validate only | Checks metadata vs GPS | ❌ No |
| --rename | Creates organized copies | ❌ No |
| --update-metadata | Fixes creation_time | ⚠️ **YES** |
| --concat | Creates combined files | ❌ No |

## Full Workflow Example

```bash
# 1. Check what's wrong
./gopro-validator --input ~/Videos/GoPro

# 2. Backup
cp -r ~/Videos/GoPro ~/Videos/GoPro-Backup

# 3. Fix metadata
./gopro-validator --input ~/Videos/GoPro --update-metadata

# 4. Create organized library
./gopro-validator --input ~/Videos/GoPro --rename --output ~/Videos/GoPro-Organized

# 5. Create full recordings
./gopro-validator --input ~/Videos/GoPro --concat --concat-output ~/Videos/Full-Recordings

# 6. Verify
./gopro-validator --input ~/Videos/GoPro
```

## Need Help?

- **Usage examples:** See [USAGE-EXAMPLES.md](USAGE-EXAMPLES.md)
- **All features:** See [FEATURES.md](FEATURES.md)
- **Windows users:** See [WINDOWS.md](WINDOWS.md)
- **Detailed docs:** See [go-validator/USAGE.md](go-validator/USAGE.md)

## Common Questions

**Q: Where does it scan by default?**
A: Current directory and all subdirectories

**Q: Can I specify which files to process?**
A: Point --input to the directory containing the files you want

**Q: Will it modify my original files?**
A: Only --update-metadata modifies originals. --rename and --concat create new files.

**Q: How do I undo changes?**
A: For --update-metadata, restore from backup. For --rename/--concat, just delete the output directory.

**Q: Does it work on Windows?**
A: Yes! See [WINDOWS.md](WINDOWS.md) for setup instructions.
