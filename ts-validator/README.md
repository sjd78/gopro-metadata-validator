# GoPro Metadata Validator (TypeScript Version)

**⚠️ ARCHIVED - Not Actively Maintained**

See [ARCHIVED.md](ARCHIVED.md) for details. **Use `../go-validator/` for production.**

---

A TypeScript/Node.js reference implementation using established GoPro libraries.

## Features

- ✅ Validates metadata against GPS timestamps
- ✅ Uses `gpmf-extract` and `gopro-telemetry` libraries
- ⚠️ Limited to files <2GB (library limitation)
- ℹ️ Validation only (no rename/update features)

## Installation

```bash
npm install
```

## Usage

```bash
npm run dev
```

The tool scans `../input-files` directory and validates all MP4 files.

## Limitations

### File Size
Cannot process files >2GB due to `gpmf-extract` loading entire file into memory.

**Files that will fail:**
- Large GoPro recordings (typically split at 4GB)
- Any file over 2GB

**Workaround:** Use the Go version in `../go-validator/`

### Features
This version only validates. For additional features:
- **Rename files** - Use Go version
- **Update metadata** - Use Go version
- **Concatenate chapters** - Use Go version

## Output

Example:
```
✓ GH016761.MP4
  Metadata Creation Time: 2016-01-01T00:10:36.000Z
  Timecode: 00:10:35:50
  GPS Samples: 8317
  GPS First Timestamp: 0.000s
  GPS Last Timestamp: 444.4s (duration: 444.4s)
```

## Libraries Used

- **gpmf-extract** (v0.3.1) - Extracts GPMF binary data
- **gopro-telemetry** (v0.6.0) - Parses GPMF to structured data

## Recommendation

**For production use, use the Go version:**
- Handles unlimited file sizes
- Includes rename/update/concat features
- Faster execution
- No memory limitations

This TypeScript version is best for:
- Development/prototyping
- Small file sets (<2GB each)
- Integration with Node.js workflows
- Learning the GPMF structure

## Project Structure

```
ts-validator/
├── src/
│   ├── index.ts              - Main entry point
│   ├── gpmf-extractor.ts     - GPS data extraction
│   ├── metadata-extractor.ts - MP4 metadata extraction
│   └── comparator.ts         - Validation logic
├── package.json
├── tsconfig.json
└── README.md
```

## Development

```bash
# Install dependencies
npm install

# Run in development mode
npm run dev

# Build
npm run build

# Run built version
npm start
```

## See Also

- [../go-validator/](../go-validator/) - Production-ready Go implementation
- [../README.md](../README.md) - Project overview
- [../FEATURES.md](../FEATURES.md) - Complete feature list
