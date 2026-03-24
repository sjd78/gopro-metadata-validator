.PHONY: all go ts clean help build rename rename-dry update update-dry concat concat-dry fix-all fix-all-dry

help:
	@echo "GoPro Metadata Validator"
	@echo ""
	@echo "Validation:"
	@echo "  make go            - Build and run Go validator (production)"
	@echo "  make ts            - Run TypeScript validator (archived/reference only)"
	@echo ""
	@echo "Actions (Go only):"
	@echo "  make rename-dry    - Preview file renaming based on GPS time"
	@echo "  make rename        - Actually rename/organize files"
	@echo "  make update-dry    - Preview metadata updates"
	@echo "  make update        - Actually update file metadata"
	@echo "  make concat-dry    - Preview chapter concatenation"
	@echo "  make concat        - Actually concatenate chapter files"
	@echo "  make fix-all-dry   - Preview both rename and update"
	@echo "  make fix-all       - Apply both rename and update"
	@echo ""
	@echo "Build:"
	@echo "  make build         - Build Go binary"
	@echo "  make clean         - Remove build artifacts"

all: go

go: build
	@cd go-validator && ./gopro-validator --input ../sample-input-files

ts:
	@cd ts-validator && npm run dev

build:
	@echo "Building Go validator..."
	@cd go-validator && go build -o gopro-validator

rename-dry: build
	@cd go-validator && ./gopro-validator --input ../sample-input-files --rename --dry-run

rename: build
	@cd go-validator && ./gopro-validator --input ../sample-input-files --rename

update-dry: build
	@cd go-validator && ./gopro-validator --input ../sample-input-files --update-metadata --dry-run

update: build
	@cd go-validator && ./gopro-validator --input ../sample-input-files --update-metadata

fix-all-dry: build
	@cd go-validator && ./gopro-validator --input ../sample-input-files --rename --update-metadata --dry-run

fix-all: build
	@cd go-validator && ./gopro-validator --input ../sample-input-files --rename --update-metadata

concat-dry: build
	@cd go-validator && ./gopro-validator --input ../sample-input-files --concat --dry-run

concat: build
	@cd go-validator && ./gopro-validator --input ../sample-input-files --concat

clean:
	@rm -f go-validator/gopro-validator
	@rm -rf ts-validator/dist ts-validator/node_modules renamed-files concatenated-files
