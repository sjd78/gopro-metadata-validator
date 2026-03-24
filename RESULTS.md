# Validation Results

Results from running both TypeScript and Go validators on the sample GoPro files.

## Summary

| Implementation | Files Processed | Valid | Issues | Large Files Handled |
|----------------|----------------|-------|--------|---------------------|
| TypeScript | 10 | 5 | 5 | ❌ No (0/5) |
| Go | 10 | 6 | 4 | ✅ Yes (5/5) |

## Key Findings

### File Organization

Files are organized by date in folders:
- `2015-12-31/` - 1 file
- `2020-04-01/` - 2 files
- `2024-02-22/` - 7 files

### Metadata Consistency

All files show **internally consistent metadata**:
- ✅ Timecode matches creation_time (within ~1 minute tolerance)
- ✅ GPS timestamps start near 0s for first chapters
- ✅ No evidence of incorrect dates in metadata

### Chapter Files Detected

GoPro splits long recordings into chapters. The Go version correctly identified:

**First Chapters** (GH01xxxx.MP4):
- GH016761.MP4 - GPS starts at 0.189s ✓
- GH016252.MP4 - GPS starts at 0.192s ✓
- GH016254.MP4 - GPS starts at 0.192s ✓
- GH016978.MP4 - GPS starts at 0.198s ✓
- GH016979.MP4 - GPS starts at 0.195s ✓
- GH016980.MP4 - GPS starts at 0.192s ✓

**Continuation Chapters** (GH02/03xxxx.MP4):
- GH026978.MP4 - GPS starts at 105.578s (2nd chapter of 6978)
- GH026979.MP4 - GPS starts at 105.384s (2nd chapter of 6979)
- GH026980.MP4 - GPS starts at 105.580s (2nd chapter of 6980)
- GH036980.MP4 - GPS starts at ~106s (3rd chapter of 6980)

This is **expected behavior** - continuation chapters have GPS timestamps continuing from where the previous chapter ended.

## File Size Distribution

| File | Size | Processed by TS | Processed by Go |
|------|------|----------------|-----------------|
| GH016761.MP4 | 1.7GB | ✅ | ✅ |
| GH016252.MP4 | 1.4GB | ✅ | ✅ |
| GH016254.MP4 | 2.1GB | ✅ | ✅ |
| GH016978.MP4 | 3.7GB | ❌ | ✅ |
| GH016979.MP4 | 3.7GB | ❌ | ✅ |
| GH016980.MP4 | 3.7GB | ❌ | ✅ |
| GH026978.MP4 | 2.0GB | ✅ | ✅ |
| GH026979.MP4 | 3.3GB | ❌ | ✅ |
| GH026980.MP4 | 3.7GB | ❌ | ✅ |
| GH036980.MP4 | 1.1GB | ✅ | ✅ |

## Conclusions

1. **No metadata errors detected** - All files have consistent timecode and creation_time
2. **Folder dates are accurate** - Files in 2024-02-22 folder were indeed recorded on that date
3. **GPS data is present** - All files contain valid GPMF GPS telemetry
4. **Go version superior for production** - Handles all file sizes, faster execution

## Recommendations

### For This Dataset
- No corrections needed - metadata is accurate
- Chapter files are properly linked (GH01→GH02→GH03)
- Folder organization reflects actual recording dates

### For Future Use
- Use **Go version** for:
  - Large file processing (>2GB)
  - Production workflows
  - Batch processing

- Use **TypeScript version** for:
  - Development and prototyping
  - Integration with Node.js workflows
  - Files <2GB
