package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

func compareMetadata(metadata *Metadata, gpsData *GPSData) []string {
	issues := make([]string, 0)

	// Check if GPS data exists
	if !gpsData.HasValidGPS {
		if gpsData.SampleCount == 0 {
			issues = append(issues, "No GPS data found in GPMF stream")
		} else {
			issues = append(issues, "GPS data present but no valid timestamps found")
		}
	}

	// PRIMARY CHECK: Compare GPS absolute time with file metadata creation time
	// Adjust GPS time for lock delay to get actual recording start time
	if gpsData.FirstGPSTime != nil && metadata.CreationTime != nil {
		// Calculate actual recording start time (GPS time - GPS lock delay)
		actualStartTime := *gpsData.FirstGPSTime
		if gpsData.FirstTimestampMs != nil {
			offset := time.Duration(*gpsData.FirstTimestampMs) * time.Millisecond
			actualStartTime = actualStartTime.Add(-offset)
		}

		// Calculate difference between actual recording start and file creation time
		diff := actualStartTime.Sub(*metadata.CreationTime)
		diffSeconds := math.Abs(diff.Seconds())

		// Allow for small differences (up to 2 seconds) due to processing delays
		if diffSeconds > 2.0 {
			diffMinutes := int(diffSeconds / 60)
			diffHours := int(diffMinutes / 60)

			var diffStr string
			if diffHours > 0 {
				diffStr = fmt.Sprintf("%d hours %d minutes", diffHours, diffMinutes%60)
			} else if diffMinutes > 0 {
				diffStr = fmt.Sprintf("%d minutes", diffMinutes)
			} else {
				diffStr = fmt.Sprintf("%.1f seconds", diffSeconds)
			}

			issues = append(issues,
				fmt.Sprintf("⚠️  GPS recording start (%s) differs from file creation time (%s) by %s",
					actualStartTime.Format("2006-01-02T15:04:05Z"),
					metadata.CreationTime.Format("2006-01-02T15:04:05Z"),
					diffStr))

			if diff.Seconds() > 0 {
				issues = append(issues, "GPS time is LATER than file creation time")
			} else {
				issues = append(issues, "GPS time is EARLIER than file creation time")
			}
		}
	}

	// Check if timecode and creation time are consistent
	if metadata.CreationTime != nil && metadata.Timecode != "" {
		tc := parseTimecode(metadata.Timecode)
		if tc != nil {
			// Extract time-of-day from creation time (UTC)
			creationHours := metadata.CreationTime.UTC().Hour()
			creationMinutes := metadata.CreationTime.UTC().Minute()
			creationSeconds := metadata.CreationTime.UTC().Second()

			// Calculate total seconds for comparison
			timecodeSeconds := tc.hours*3600 + tc.minutes*60 + tc.seconds
			creationTimeSeconds := creationHours*3600 + creationMinutes*60 + creationSeconds

			// Allow for small differences (up to 2 minutes) due to processing/encoding delays
			diffSeconds := int(math.Abs(float64(timecodeSeconds - creationTimeSeconds)))

			if diffSeconds > 120 {
				diffMinutes := diffSeconds / 60
				issues = append(issues,
					fmt.Sprintf("Timecode (%s) and creation time (%s) differ by %d minutes",
						metadata.Timecode,
						metadata.CreationTime.Format("2006-01-02T15:04:05Z"),
						diffMinutes))
				issues = append(issues, "This suggests the metadata creation date may be incorrect")
			}
		}
	}

	// If we have GPS data with relative timestamps, verify it's reasonable
	if gpsData.HasValidGPS && gpsData.FirstTimestampMs != nil && metadata.CreationTime != nil {
		// GPS timestamps should start near 0 (within a few seconds of recording start)
		gpsStartSeconds := float64(*gpsData.FirstTimestampMs) / 1000.0

		if gpsStartSeconds > 60 {
			issues = append(issues,
				fmt.Sprintf("GPS first timestamp is %.1fs - expected to start near 0s (likely a chapter continuation)", gpsStartSeconds))
		}
	}

	if metadata.CreationTime == nil {
		issues = append(issues, "No creation time found in metadata")
	}

	if metadata.Timecode == "" {
		issues = append(issues, "No timecode found in metadata")
	}

	return issues
}

type timecode struct {
	hours   int
	minutes int
	seconds int
	frames  int
}

func parseTimecode(tc string) *timecode {
	// Timecode format: HH:MM:SS:FF
	parts := strings.Split(tc, ":")
	if len(parts) != 4 {
		return nil
	}

	hours, err1 := strconv.Atoi(parts[0])
	minutes, err2 := strconv.Atoi(parts[1])
	seconds, err3 := strconv.Atoi(parts[2])
	frames, err4 := strconv.Atoi(parts[3])

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return nil
	}

	return &timecode{
		hours:   hours,
		minutes: minutes,
		seconds: seconds,
		frames:  frames,
	}
}
