# TypeScript Version - Archived

**Status:** Reference implementation, not actively maintained

## Why Archived?

The TypeScript version served as the initial prototype and proof of concept. However, the Go implementation (`../go-validator/`) has become the production version due to:

1. **File size limitations** - TypeScript version can't handle files >2GB
2. **Feature completeness** - Go version has all features (rename, update, concat)
3. **Performance** - Go is significantly faster
4. **Cross-platform** - Better Windows compatibility
5. **Maintenance** - Easier to maintain one codebase

## What This Version Does

- ✅ Validates metadata against GPS timestamps
- ✅ Uses established libraries (`gpmf-extract`, `gopro-telemetry`)
- ⚠️ Limited to files <2GB
- ❌ No rename/update/concat features

## If You Want to Use It

```bash
npm install
npm run dev
```

Scans `../sample-input-files/` and validates all MP4 files under 2GB.

## For Production Use

**Use the Go version:** `../go-validator/`

It has all the features and none of the limitations:
- Handles unlimited file sizes
- Includes rename, update, concat operations
- Better performance
- Active development

## Code Reference

This implementation may be useful if you want to:
- Integrate with existing Node.js workflows
- Study how to use `gpmf-extract` and `gopro-telemetry`
- Port features to another TypeScript project

## Libraries Used

- `gpmf-extract` (v0.3.1) - Extracts GPMF from MP4
- `gopro-telemetry` (v0.6.0) - Parses GPMF data

These are well-maintained libraries, but have inherent limitations (2GB file size due to loading entire file into memory).

## Last Updated

2026-03-24 - Marked as archived, all development moved to Go version
