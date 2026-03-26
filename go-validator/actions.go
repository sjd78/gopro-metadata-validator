package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func renameFilesBasedOnGPS(results []*ValidationResult, outputDir string, dryRun bool) int {
	if !dryRun {
		// Create output directory
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			return 0
		}
	}

	renamed := 0
	skipped := 0

	// Track renamed file paths for sidecar generation
	renamedFiles := make(map[string]*GPSData)

	for _, result := range results {
		if result.GPSData.FirstGPSTime == nil {
			fmt.Printf("⊘ %s - No GPS timestamp, skipping\n", filepath.Base(result.FilePath))
			skipped++
			continue
		}

		// Calculate actual recording start time by adjusting for GPS lock delay
		// FirstGPSTime is when GPS got a lock, FirstTimestampMs is how long after recording started
		actualStartTime := calculateRecordingStartTime(result.GPSData)

		// Generate new filename: YYYYMMDD_HHMMSS_original.MP4
		newFilename := fmt.Sprintf("%s_%s.MP4",
			actualStartTime.Format("20060102_150405"),
			strings.TrimSuffix(filepath.Base(result.FilePath), filepath.Ext(result.FilePath)),
		)

		// Create subdirectory based on date
		dateDir := filepath.Join(outputDir, actualStartTime.Format("2006-01-02"))
		newPath := filepath.Join(dateDir, newFilename)

		// Ensure unique filename to avoid overwriting existing files
		newPath = GenerateUniqueFilename(newPath)

		if dryRun {
			fmt.Printf("📋 Would rename:\n")
			fmt.Printf("   From: %s\n", result.FilePath)
			fmt.Printf("   To:   %s\n\n", newPath)
			renamed++

			// Track for sidecar dry-run
			if result.GPSData != nil && result.GPSData.HasValidGPS {
				renamedFiles[newPath] = result.GPSData
			}
		} else {
			// Create date subdirectory
			if err := os.MkdirAll(dateDir, 0755); err != nil {
				fmt.Printf("✗ Error creating directory %s: %v\n", dateDir, err)
				skipped++
				continue
			}

			// Copy file to new location
			if err := copyFile(result.FilePath, newPath); err != nil {
				fmt.Printf("✗ Error copying %s: %v\n", filepath.Base(result.FilePath), err)
				skipped++
				continue
			}

			fmt.Printf("✓ Renamed: %s -> %s\n", filepath.Base(result.FilePath), newFilename)
			renamed++

			// Track renamed path for sidecar creation
			if result.GPSData != nil && result.GPSData.HasValidGPS {
				renamedFiles[newPath] = result.GPSData
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("-", 80))
	if dryRun {
		fmt.Printf("Dry run complete: %d files would be renamed, %d skipped\n", renamed, skipped)
		fmt.Printf("Run without --dry-run to actually rename files\n")
	} else {
		fmt.Printf("Complete: %d files renamed to %s, %d skipped\n", renamed, outputDir, skipped)
	}

	// Create sidecars for renamed files
	if len(renamedFiles) > 0 {
		fmt.Println("\nCreating XMP sidecars for renamed files...")
		sidecarCount := 0

		for newPath, gpsData := range renamedFiles {
			if err := WriteSidecarForFile(newPath, gpsData, dryRun); err != nil {
				fmt.Printf("⚠️  Error creating sidecar for %s: %v\n", filepath.Base(newPath), err)
			} else {
				sidecarCount++
			}
		}

		if dryRun {
			fmt.Printf("%d sidecar files would be created\n", sidecarCount)
		} else {
			fmt.Printf("%d sidecar files created\n", sidecarCount)
		}
	}

	return renamed
}

func updateFileMetadata(results []*ValidationResult, dryRun bool) {
	updated := 0
	skipped := 0

	for _, result := range results {
		if result.GPSData.FirstGPSTime == nil {
			fmt.Printf("⊘ %s - No GPS timestamp, skipping\n", filepath.Base(result.FilePath))
			skipped++
			continue
		}

		// Calculate actual recording start time by adjusting for GPS lock delay
		actualStartTime := calculateRecordingStartTime(result.GPSData)

		// Check if update is needed
		if result.Metadata.CreationTime != nil {
			diff := actualStartTime.Sub(*result.Metadata.CreationTime)
			if diff < 2*time.Second && diff > -2*time.Second {
				fmt.Printf("⊘ %s - Already correct, skipping\n", filepath.Base(result.FilePath))
				skipped++
				continue
			}
		}

		if dryRun {
			fmt.Printf("📋 Would update metadata:\n")
			fmt.Printf("   File: %s\n", filepath.Base(result.FilePath))
			if result.Metadata.CreationTime != nil {
				fmt.Printf("   Current: %s\n", result.Metadata.CreationTime.Format("2006-01-02T15:04:05Z"))
			}
			fmt.Printf("   New:     %s\n", actualStartTime.Format("2006-01-02T15:04:05Z"))
			if result.GPSData.FirstTimestampMs != nil && *result.GPSData.FirstTimestampMs > 1000 {
				fmt.Printf("   Note: GPS lock took %.1fs after recording started\n", float64(*result.GPSData.FirstTimestampMs)/1000.0)
			}
			fmt.Println()
			updated++
		} else {
			// Create temporary file path
			tmpFile := result.FilePath + ".tmp"

			// Use ffmpeg to update metadata
			creationTime := actualStartTime.Format("2006-01-02T15:04:05")
			cmd := exec.Command("ffmpeg",
				"-i", result.FilePath,
				"-c", "copy",
				"-metadata", fmt.Sprintf("creation_time=%s", creationTime),
				"-y",
				tmpFile,
			)

			// Suppress ffmpeg output
			cmd.Stderr = nil
			cmd.Stdout = nil

			if err := cmd.Run(); err != nil {
				fmt.Printf("✗ Error updating %s: %v\n", filepath.Base(result.FilePath), err)
				os.Remove(tmpFile)
				skipped++
				continue
			}

			// Replace original with updated file
			if err := os.Rename(tmpFile, result.FilePath); err != nil {
				fmt.Printf("✗ Error replacing %s: %v\n", filepath.Base(result.FilePath), err)
				os.Remove(tmpFile)
				skipped++
				continue
			}

			fmt.Printf("✓ Updated: %s (%s)\n", filepath.Base(result.FilePath), actualStartTime.Format("2006-01-02T15:04:05Z"))
			updated++
		}
	}

	fmt.Println("\n" + strings.Repeat("-", 80))
	if dryRun {
		fmt.Printf("Dry run complete: %d files would be updated, %d skipped\n", updated, skipped)
		fmt.Printf("Run without --dry-run to actually update metadata\n")
	} else {
		fmt.Printf("Complete: %d files updated, %d skipped\n", updated, skipped)
	}
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}

// calculateRecordingStartTime adjusts the GPS time for GPS lock delay
// FirstGPSTime is when GPS got a lock, FirstTimestampMs is milliseconds after recording started
// Actual start time = FirstGPSTime - FirstTimestampMs
func calculateRecordingStartTime(gpsData *GPSData) time.Time {
	if gpsData.FirstGPSTime == nil {
		return time.Time{}
	}

	actualStart := *gpsData.FirstGPSTime

	// Adjust for GPS lock delay if we have relative timestamp info
	if gpsData.FirstTimestampMs != nil {
		// Subtract the offset to get when recording actually started
		offset := time.Duration(*gpsData.FirstTimestampMs) * time.Millisecond
		actualStart = actualStart.Add(-offset)
	}

	return actualStart
}

// GenerateUniqueFilename ensures the output path doesn't exist by appending (1), (2), etc.
// Pattern: "file.mp4" → "file (1).mp4" → "file (2).mp4"
func GenerateUniqueFilename(path string) string {
	// If file doesn't exist, return original path
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	// File exists, need to add suffix
	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	base := filepath.Base(path)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	// Try (1), (2), (3), etc. until we find a free name
	for i := 1; ; i++ {
		newName := fmt.Sprintf("%s (%d)%s", nameWithoutExt, i, ext)
		newPath := filepath.Join(dir, newName)

		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}
