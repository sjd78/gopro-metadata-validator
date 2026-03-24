package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	inputDir        = flag.String("input", "", "Input directory containing GoPro files (default: current directory)")
	renameFiles     = flag.Bool("rename", false, "Rename and move files based on GPS timestamps")
	updateMetadata  = flag.Bool("update-metadata", false, "Update MP4 metadata to match GPS timestamps")
	concatChapters  = flag.Bool("concat", false, "Concatenate chapter files into complete recordings")
	dryRun          = flag.Bool("dry-run", false, "Show what would be done without making changes")
	outputDir       = flag.String("output", "renamed-files", "Output directory for renamed files")
	concatOutputDir = flag.String("concat-output", "concatenated-files", "Output directory for concatenated files")
)

func main() {
	flag.Parse()

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

	// Perform actions based on flags
	if *renameFiles {
		fmt.Println("\n" + "================================================================================")
		fmt.Println("FILE RENAMING PLAN")
		fmt.Println("================================================================================\n")
		renameFilesBasedOnGPS(results, *outputDir, *dryRun)
	}

	if *updateMetadata {
		fmt.Println("\n" + "================================================================================")
		fmt.Println("METADATA UPDATE PLAN")
		fmt.Println("================================================================================\n")
		updateFileMetadata(results, *dryRun)
	}

	if *concatChapters {
		fmt.Println("\n" + "================================================================================")
		fmt.Println("CHAPTER CONCATENATION PLAN")
		fmt.Println("================================================================================\n")
		concatenateChapters(results, *concatOutputDir, *dryRun)
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
	fmt.Println("================================================================================\n")

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
