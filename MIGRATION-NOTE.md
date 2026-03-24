# Migration Note: TypeScript Code Moved

**Date:** 2026-03-24

## What Changed

The TypeScript validator has been moved to its own directory for better project organization.

### Old Structure
```
gopro_renamer/
├── src/              ← TypeScript source
├── package.json      ← TypeScript config
├── tsconfig.json     ← TypeScript config
└── ...
```

### New Structure
```
gopro_renamer/
├── ts-validator/     ← All TypeScript code here
│   ├── src/
│   ├── package.json
│   ├── tsconfig.json
│   └── README.md
├── go-validator/     ← All Go code here
└── ...
```

## How to Update

### If You Have a Local Clone

**Option 1: Pull fresh**
```bash
cd /path/to/gopro_renamer
git pull
cd ts-validator
npm install
```

**Option 2: Manual migration**
```bash
cd /path/to/gopro_renamer
mkdir -p ts-validator
mv src package.json package-lock.json tsconfig.json ts-validator/
mv node_modules ts-validator/  # if exists
```

### If You're Using the Makefile

No changes needed! The Makefile has been updated:

```bash
make ts        # Still works, now runs from ts-validator/
make go        # Still works
make help      # See all commands
```

### If You're Running Commands Directly

**Old:**
```bash
npm install
npm run dev
```

**New:**
```bash
cd ts-validator
npm install
npm run dev
```

## What Wasn't Moved

These remain in the root directory:
- `.gitignore`
- `input-files/`
- All documentation (README.md, FEATURES.md, etc.)
- `Makefile`

## Benefits of This Change

1. **Better organization** - Clear separation of Go and TypeScript code
2. **Easier to understand** - Each implementation in its own folder
3. **Parallel development** - Develop both versions independently
4. **Cleaner root** - Less clutter in project root
5. **Consistent structure** - Mirrors `go-validator/` organization

## No Breaking Changes

- All functionality remains the same
- Makefile commands work as before
- Documentation updated automatically
- Both validators still use `../input-files/`

## Questions?

See:
- [README.md](README.md) - Updated project overview
- [ts-validator/README.md](ts-validator/README.md) - TypeScript-specific docs
- [go-validator/README.md](go-validator/README.md) - Go-specific docs
