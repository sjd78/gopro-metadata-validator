# Development Notes

## Active Development

**All development happens in `/go-validator/`**

The TypeScript version (`/ts-validator/`) is archived and not maintained.

## Contributing

When adding features or fixes:
- Work in `/go-validator/` only
- Update documentation in the root and `/go-validator/` directories
- Test on sample files: `make go`, `make concat-dry`, etc.
- Update CHANGELOG.md

## Building

```bash
cd go-validator
go build -o gopro-validator
```

## Testing

```bash
# Use sample files
make go
make concat-dry
make update-dry

# Or test directly
cd go-validator
./gopro-validator --input ../sample-input-files
```

## Release Checklist

- [ ] All tests pass
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Cross-platform tested (or design verified)
- [ ] Sample files validated successfully

## Architecture

See `/go-validator/` for current implementation:
- `main.go` - CLI and orchestration
- `gpmf.go` - GPMF parsing and GPS extraction
- `metadata.go` - MP4 metadata extraction
- `comparator.go` - Validation logic
- `actions.go` - Rename/update operations
- `concat.go` - Chapter concatenation
