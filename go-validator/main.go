package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	// Version is set via ldflags at build time
	Version = "dev"

	inputDir        = flag.String("input", "", "Input directory containing GoPro files (default: current directory)")
	renameFiles     = flag.Bool("rename", false, "Rename and move files based on GPS timestamps")
	updateMetadata  = flag.Bool("update-metadata", false, "Update MP4 metadata to match GPS timestamps")
	concatChapters  = flag.Bool("concat", false, "Concatenate chapter files into complete recordings")
	writeSidecar    = flag.Bool("write-sidecar", false, "Write XMP sidecar files with GPMF metadata")
	dryRun          = flag.Bool("dry-run", false, "Show what would be done without making changes")
	versionFlag     = flag.Bool("version", false, "Show version and exit")
	outputDir       = flag.String("output", "renamed-files", "Output directory for renamed files")
	concatOutputDir = flag.String("concat-output", "concatenated-files", "Output directory for concatenated files")
)

func main() {
	flag.Parse()

	// Handle version flag
	if *versionFlag {
		fmt.Printf("gopro-validator %s\n", Version)
		os.Exit(0)
	}

	// Determine input directory
	scanDir := *inputDir
	if scanDir == "" {
		// Default to current working directory
		var err error
		scanDir, err = os.Getwd()
		if err != nil {
			log.Fatalf("Error getting current directory: %v", err)
		}
	}

	// Convert to absolute path
	absInputDir, err := filepath.Abs(scanDir)
	if err != nil {
		log.Fatalf("Error resolving input path: %v", err)
	}

	fmt.Printf("🔍 Scanning for GoPro video files in: %s\n\n", absInputDir)

	files, err := findMP4Files(absInputDir)
	if err != nil {
		log.Fatalf("Error scanning directory: %v", err)
	}

	fmt.Printf("Found %d MP4 files\n\n", len(files))

	results := make([]*ValidationResult, 0, len(files))

	for _, file := range files {
		fmt.Printf("Processing: %s\n", file)
		result, err := validateFile(file)
		if err != nil {
			log.Printf("Error processing %s: %v", file, err)
			continue
		}
		results = append(results, result)
	}

	printResults(results)

	// Track operation counts for exiftool instructions
	renamedCount := 0
	concatCount := 0

	// Write sidecar files if requested
	if *writeSidecar {
		fmt.Println("\n" + "================================================================================")
		fmt.Println("WRITING SIDECAR FILES")
		fmt.Println("================================================================================")
		writeSidecarFiles(results, *dryRun)
	}

	// Perform actions based on flags
	if *renameFiles {
		fmt.Println("\n" + "================================================================================")
		fmt.Println("FILE RENAMING PLAN")
		fmt.Println("================================================================================")
		renamedCount = renameFilesBasedOnGPS(results, *outputDir, *dryRun)
	}

	if *updateMetadata {
		fmt.Println("\n" + "================================================================================")
		fmt.Println("METADATA UPDATE PLAN")
		fmt.Println("================================================================================")
		updateFileMetadata(results, *dryRun)
	}

	if *concatChapters {
		fmt.Println("\n" + "================================================================================")
		fmt.Println("CHAPTER CONCATENATION PLAN")
		fmt.Println("================================================================================")
		concatCount = concatenateChapters(results, *concatOutputDir, *dryRun)
	}

	// Show exiftool instructions if sidecars were created
	if !*dryRun {
		showExiftoolInstructions(renamedCount, concatCount, *outputDir, *concatOutputDir)
	}
}

func findMP4Files(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (filepath.Ext(path) == ".MP4" || filepath.Ext(path) == ".mp4") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func printResults(results []*ValidationResult) {
	fmt.Println("\n" + "================================================================================")
	fmt.Println("VALIDATION RESULTS")
	fmt.Println("================================================================================")

	validCount := 0
	invalidCount := 0

	for _, result := range results {
		status := "✓"
		if !result.Valid {
			status = "✗"
		}

		fmt.Printf("%s %s\n", status, filepath.Base(result.FilePath))

		if result.Metadata.CreationTime != nil {
			fmt.Printf("  Metadata Creation Time: %s\n", result.Metadata.CreationTime.Format("2006-01-02T15:04:05Z"))
		}
		if result.Metadata.Timecode != "" {
			fmt.Printf("  Timecode: %s\n", result.Metadata.Timecode)
		}

		if result.GPSData.HasValidGPS && result.GPSData.SampleCount > 0 {
			fmt.Printf("  GPS Samples: %d\n", result.GPSData.SampleCount)

			// Show GPS coordinates if available
			if len(result.GPSData.Coordinates) > 0 {
				first := result.GPSData.Coordinates[0]
				fmt.Printf("  GPS Coordinates (first): %.7f°, %.7f° (alt: %.2fm WGS84)\n",
					first.Latitude, first.Longitude, first.Altitude)

				// Show GPS fix type and precision
				fixInfo := ""
				if result.GPSData.GPSFix != nil {
					fixInfo = fmt.Sprintf("%s fix", *result.GPSData.GPSFix)
				}
				if result.GPSData.GPSPrecision != nil {
					if fixInfo != "" {
						fixInfo += fmt.Sprintf(", DOP: %.2f", *result.GPSData.GPSPrecision)
					} else {
						fixInfo = fmt.Sprintf("DOP: %.2f", *result.GPSData.GPSPrecision)
					}
				}
				if fixInfo != "" {
					fmt.Printf("  GPS Quality: %s\n", fixInfo)
				}

				fmt.Printf("  GPS Track Points: %d\n", len(result.GPSData.Coordinates))
			}

			// Show absolute GPS time if available
			if result.GPSData.FirstGPSTime != nil {
				fmt.Printf("  GPS Absolute Time (first): %s\n", result.GPSData.FirstGPSTime.Format("2006-01-02T15:04:05Z"))
			}
			if result.GPSData.LastGPSTime != nil {
				fmt.Printf("  GPS Absolute Time (last):  %s\n", result.GPSData.LastGPSTime.Format("2006-01-02T15:04:05Z"))
			}

			// Show relative timestamps
			if result.GPSData.FirstTimestampMs != nil {
				fmt.Printf("  GPS Relative Start: %.3fs\n", float64(*result.GPSData.FirstTimestampMs)/1000.0)
			}
			if result.GPSData.LastTimestampMs != nil {
				duration := float64(*result.GPSData.LastTimestampMs) / 1000.0
				fmt.Printf("  GPS Relative End: %.1fs\n", duration)
			}
		} else if result.GPSData.SampleCount > 0 {
			fmt.Printf("  GPS Samples: %d (no valid timestamps)\n", result.GPSData.SampleCount)
		} else {
			fmt.Println("  GPS Data: Not available")
		}

		if len(result.Issues) > 0 {
			invalidCount++
			fmt.Println("  Issues:")
			for _, issue := range result.Issues {
				fmt.Printf("    - %s\n", issue)
			}
		} else {
			validCount++
		}
		fmt.Println()
	}

	fmt.Println("================================================================================")
	fmt.Printf("Summary: %d valid, %d with issues\n", validCount, invalidCount)
	fmt.Println("================================================================================")
}

func writeSidecarFiles(results []*ValidationResult, dryRun bool) {
	written := 0
	skipped := 0

	for _, result := range results {
		if err := WriteSidecarFile(result, dryRun); err != nil {
			fmt.Printf("✗ Error writing sidecar for %s: %v\n",
				filepath.Base(result.FilePath), err)
			skipped++
			continue
		}

		if result.GPSData != nil && result.GPSData.HasValidGPS {
			written++
		} else {
			skipped++
		}
	}

	fmt.Println("\n" + "--------------------------------------------------------------------------------")
	if dryRun {
		fmt.Printf("Dry run: %d sidecar files would be written, %d skipped\n", written, skipped)
		fmt.Printf("Run without --dry-run to actually write sidecar files\n")
	} else {
		fmt.Printf("Complete: %d sidecar files written, %d skipped (no GPS data)\n", written, skipped)
	}
}

func showExiftoolInstructions(renamedCount, concatCount int, outputDir, concatOutputDir string) {
	if renamedCount == 0 && concatCount == 0 {
		return // No sidecars were created
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("EMBEDDING SIDECAR METADATA WITH EXIFTOOL")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()
	fmt.Println("XMP sidecar files have been created with GPS coordinates and timezone information.")
	fmt.Println("To embed this metadata into your MP4 files, use exiftool:")
	fmt.Println()

	if renamedCount > 0 {
		fmt.Println("For renamed files:")
		fmt.Printf("  exiftool -tagsFromFile %%f.xmp -all:all -ext MP4 %s/\n", outputDir)
		fmt.Println()
	}

	if concatCount > 0 {
		fmt.Println("For concatenated files:")
		fmt.Printf("  exiftool -tagsFromFile %%f.xmp -all:all -ext MP4 %s/\n", concatOutputDir)
		fmt.Println()
	}

	fmt.Println("Or for a single file:")
	fmt.Println("  exiftool -tagsFromFile video.mp4.xmp -all:all video.mp4")
	fmt.Println()
	fmt.Println("To verify embedded GPS data:")
	fmt.Println("  exiftool -GPS* -TimeZone -DateTimeOriginal -CreateDate video.mp4")
	fmt.Println()
	fmt.Println("Note: exiftool must be installed (https://exiftool.org/)")
	fmt.Println(strings.Repeat("=", 80))
}
