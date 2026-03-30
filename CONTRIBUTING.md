# Contributing to GoPro Metadata Validator

## Development Setup

### Prerequisites

- Go 1.19 or later
- `ffmpeg` and `ffprobe` in your PATH
- (Optional) `lefthook` for Git hooks

### Setting Up Git Hooks with Lefthook

This project uses [lefthook](https://github.com/evilmartians/lefthook) to automatically format Go code and run checks before commits.

**One-time setup:**

```bash
# Install lefthook
go install github.com/evilmartians/lefthook@latest

# Install hooks in this repository
lefthook install

# Fix PATH if needed (only if hooks show "Can't find lefthook in PATH")
# Add this line after line 1 in .git/hooks/pre-commit:
export PATH="$HOME/go/bin:$PATH"
```

**What the hooks do:**

- **Auto-format**: Runs `gofmt -w` on staged Go files automatically
- **Vet**: Runs `go vet ./...` to catch common mistakes

**Skipping hooks:**

If you need to bypass the hooks (not recommended):
```bash
git commit --no-verify -m "your message"
```

Or temporarily disable:
```bash
LEFTHOOK=0 git commit -m "your message"
```

## Code Formatting

All Go code must pass `gofmt` formatting. The lefthook pre-commit hook will automatically format your code, but you can also run it manually:

```bash
cd go-validator
go fmt ./...
```

## Running Tests

```bash
cd go-validator
go build -o gopro-validator
./gopro-validator --input ../sample-input-files
```

## Code Quality Checks

Before submitting a PR, ensure your code passes:

```bash
# Format check
go fmt ./...

# Vet check
go vet ./...

# Build check
go build -o gopro-validator
```

These checks also run in GitHub Actions CI.

## Project Structure

- `go-validator/` - Main Go implementation
  - `main.go` - CLI entry point
  - `gpmf.go` - GPMF stream parsing
  - `metadata.go` - MP4 metadata extraction
  - `sidecar.go` - XMP sidecar generation
  - `timezone.go` - GPS coordinate to timezone mapping
  - `actions.go` - File operations (rename, metadata update)
  - `concat.go` - Chapter file concatenation
- `.lefthook.yml` - Git hook configuration
- `sample-input-files/` - Test videos (not in git)

## Making Changes

1. Create a feature branch
2. Make your changes
3. Let lefthook auto-format your code on commit
4. Push and create a PR
5. CI will verify formatting and run checks

## Questions?

Open an issue or discussion on GitHub.
