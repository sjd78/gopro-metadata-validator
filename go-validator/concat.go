package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ChapterSeries represents a group of files that are chapters of the same recording
type ChapterSeries struct {
	BaseNumber string   // e.g., "6978"
	Files      []string // Sorted list of chapter files
	StartTime  string   // Recording start time from GPS
}

// detectChapterSeries groups files into chapter series (GH01xxxx, GH02xxxx, etc.)
func detectChapterSeries(results []*ValidationResult) map[string]*ChapterSeries {
	series := make(map[string]*ChapterSeries)

	// Regex to extract chapter number and base number from GoPro filename
	// GH016978.MP4 -> chapter=01, base=6978
	pattern := regexp.MustCompile(`GH(\d)(\d)(\d{4})\.MP4`)

	for _, result := range results {
		filename := filepath.Base(result.FilePath)
		matches := pattern.FindStringSubmatch(filename)

		if len(matches) != 4 {
			continue
		}

		chapterNum := matches[1] + matches[2] // "01", "02", "03"
		baseNumber := matches[3]              // "6978"

		if _, exists := series[baseNumber]; !exists {
			series[baseNumber] = &ChapterSeries{
				BaseNumber: baseNumber,
				Files:      make([]string, 0),
			}
		}

		series[baseNumber].Files = append(series[baseNumber].Files, result.FilePath)

		// Use the GPS start time from the first chapter (GH01)
		if chapterNum == "01" && result.GPSData.FirstGPSTime != nil {
			actualStart := calculateRecordingStartTime(result.GPSData)
			series[baseNumber].StartTime = actualStart.Format("20060102_150405")
		}
	}

	// Sort files in each series
	for _, s := range series {
		sort.Strings(s.Files)
	}

	// Filter out single-file "series" (not actually chapter recordings)
	filtered := make(map[string]*ChapterSeries)
	for key, s := range series {
		if len(s.Files) > 1 {
			filtered[key] = s
		}
	}

	return filtered
}

func concatenateChapters(results []*ValidationResult, outputDir string, dryRun bool) int {
	series := detectChapterSeries(results)

	if len(series) == 0 {
		fmt.Println("No multi-chapter recordings found.")
		fmt.Println("(Chapter files are named GH01xxxx.MP4, GH02xxxx.MP4, etc. with the same base number)")
		return 0
	}

	if !dryRun {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			return 0
		}
	}

	concatenated := 0
	skipped := 0

	// Track concatenated file paths for sidecar generation
	concatenatedFiles := make(map[string]*GPSData)

	// Sort by base number for consistent output
	var baseNumbers []string
	for baseNum := range series {
		baseNumbers = append(baseNumbers, baseNum)
	}
	sort.Strings(baseNumbers)

	for _, baseNum := range baseNumbers {
		s := series[baseNum]

		// Get GPS data from first chapter
		var gpsData *GPSData
		for _, result := range results {
			if result.FilePath == s.Files[0] {
				gpsData = result.GPSData
				break
			}
		}

		// Determine output filename
		var outputName string
		if s.StartTime != "" {
			outputName = fmt.Sprintf("%s_GH%s_FULL.MP4", s.StartTime, s.BaseNumber)
		} else {
			outputName = fmt.Sprintf("GH%s_FULL.MP4", s.BaseNumber)
		}
		outputPath := filepath.Join(outputDir, outputName)

		// Ensure unique filename to avoid overwriting existing files
		outputPath = GenerateUniqueFilename(outputPath)
		outputName = filepath.Base(outputPath)

		if dryRun {
			fmt.Printf("📋 Would concatenate %d chapters:\n", len(s.Files))
			for i, file := range s.Files {
				fmt.Printf("   [%d] %s\n", i+1, filepath.Base(file))
			}
			fmt.Printf("   → Output: %s\n", outputPath)
			fmt.Printf("   → Sidecar: %s.xmp\n\n", outputName)
			concatenated++

			// Track for sidecar dry-run
			if gpsData != nil && gpsData.HasValidGPS {
				concatenatedFiles[outputPath] = gpsData
			}
		} else {
			fmt.Printf("🔗 Concatenating %d chapters for recording %s...\n", len(s.Files), s.BaseNumber)

			// Create concat file list for ffmpeg
			concatListFile := filepath.Join(os.TempDir(), fmt.Sprintf("concat_%s.txt", s.BaseNumber))
			if err := createConcatList(s.Files, concatListFile); err != nil {
				fmt.Printf("✗ Error creating concat list: %v\n", err)
				skipped++
				continue
			}

			// Use ffmpeg concat demuxer to join files
			cmd := exec.Command("ffmpeg",
				"-f", "concat",
				"-safe", "0",
				"-i", concatListFile,
				"-c", "copy", // Copy all streams without re-encoding
				"-y",
				outputPath,
			)

			// Show ffmpeg progress
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				fmt.Printf("✗ Error concatenating: %v\n", err)
				os.Remove(concatListFile)
				skipped++
				continue
			}

			os.Remove(concatListFile)

			// Verify output file was created
			if _, err := os.Stat(outputPath); err != nil {
				fmt.Printf("✗ Error: Output file not created: %s\n", outputPath)
				skipped++
				continue
			}

			// Get output file size
			info, _ := os.Stat(outputPath)
			sizeGB := float64(info.Size()) / (1024 * 1024 * 1024)

			fmt.Printf("✓ Created: %s (%.1f GB)\n\n", outputName, sizeGB)
			concatenated++

			// Track concatenated path for sidecar creation
			if gpsData != nil && gpsData.HasValidGPS {
				concatenatedFiles[outputPath] = gpsData
			}
		}
	}

	fmt.Println(strings.Repeat("-", 80))
	if dryRun {
		fmt.Printf("Dry run complete: %d chapter series would be concatenated, %d skipped\n", concatenated, skipped)
		fmt.Printf("Run without --dry-run to actually concatenate files\n")
	} else {
		fmt.Printf("Complete: %d recordings concatenated to %s, %d skipped\n", concatenated, outputDir, skipped)
	}

	// Create sidecars for concatenated files
	if len(concatenatedFiles) > 0 {
		fmt.Println("\nCreating XMP sidecars for concatenated files...")
		sidecarCount := 0

		for outputPath, gpsData := range concatenatedFiles {
			if err := WriteSidecarForFile(outputPath, gpsData, dryRun); err != nil {
				fmt.Printf("⚠️  Warning: Could not create sidecar: %v\n", err)
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

	return concatenated
}

func createConcatList(files []string, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, file := range files {
		// Get absolute path
		absPath, err := filepath.Abs(file)
		if err != nil {
			return err
		}
		// Convert to forward slashes for ffmpeg (works on all platforms)
		// Windows: C:\path\file.mp4 -> C:/path/file.mp4
		// Unix: /path/file.mp4 -> /path/file.mp4 (no change)
		absPath = filepath.ToSlash(absPath)
		// Write in ffmpeg concat format
		fmt.Fprintf(f, "file '%s'\n", absPath)
	}

	return nil
}
