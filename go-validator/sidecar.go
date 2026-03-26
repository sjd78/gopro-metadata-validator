package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteSidecarForFile creates XMP sidecar for any file path with GPS data
func WriteSidecarForFile(filePath string, gpsData *GPSData, dryRun bool) error {
	if gpsData == nil || !gpsData.HasValidGPS {
		return nil // Skip if no GPS data
	}

	sidecarPath := filePath + ".xmp"

	if dryRun {
		fmt.Printf("📋 Would write sidecar: %s\n", filepath.Base(sidecarPath))
		return nil
	}

	xmpContent := generateXMP(gpsData, filepath.Base(filePath))

	if err := os.WriteFile(sidecarPath, []byte(xmpContent), 0644); err != nil {
		return fmt.Errorf("failed to write sidecar: %w", err)
	}

	fmt.Printf("✓ Wrote sidecar: %s\n", filepath.Base(sidecarPath))
	return nil
}

// WriteSidecarFile creates an XMP sidecar file with GPMF metadata
// Maintained for backward compatibility - calls WriteSidecarForFile
func WriteSidecarFile(result *ValidationResult, dryRun bool) error {
	if result.GPSData == nil || !result.GPSData.HasValidGPS {
		return nil
	}
	return WriteSidecarForFile(result.FilePath, result.GPSData, dryRun)
}

func generateXMP(gpsData *GPSData, _ string) string {
	// Calculate GPS lock delay
	var lockDelay string
	if gpsData.FirstTimestampMs != nil {
		lockDelay = fmt.Sprintf("%.1f", float64(*gpsData.FirstTimestampMs)/1000.0)
	}

	// Get GPS timestamp
	timestamp := ""
	dateTime := ""
	dateTimeWithTZ := ""
	timezone := ""

	if gpsData.FirstGPSTime != nil {
		timestamp = gpsData.FirstGPSTime.Format("2006-01-02T15:04:05Z")
		dateTime = gpsData.FirstGPSTime.Format("2006-01-02T15:04:05Z")
	}

	// Get coordinates if available
	var lat, lon, alt, speed string
	hasCoords := len(gpsData.Coordinates) > 0
	if hasCoords {
		first := gpsData.Coordinates[0]
		lat = fmt.Sprintf("%.7f", first.Latitude) // 7 decimal places for GPS precision
		lon = fmt.Sprintf("%.7f", first.Longitude)
		alt = fmt.Sprintf("%.2f", first.Altitude)      // WGS84 ellipsoid height in meters
		speed = fmt.Sprintf("%.2f", first.Speed2D*3.6) // m/s to km/h

		// Determine timezone from GPS coordinates
		if gpsData.FirstGPSTime != nil {
			tz := getTimezoneFromCoordinates(first.Latitude, first.Longitude)
			timezone = tz.String()

			// Format datetime with timezone offset
			tzTime := gpsData.FirstGPSTime.In(tz)
			dateTimeWithTZ = tzTime.Format("2006-01-02T15:04:05-07:00")
		}
	}

	// Build XMP
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	sb.WriteString(`<x:xmpmeta xmlns:x="adobe:ns:meta/" xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">` + "\n")
	sb.WriteString(`  <rdf:RDF xmlns:exif="http://ns.adobe.com/exif/1.0/"` + "\n")
	sb.WriteString(`           xmlns:xmp="http://ns.adobe.com/xap/1.0/"` + "\n")
	sb.WriteString(`           xmlns:exifEX="http://cipa.jp/exif/1.0/"` + "\n")
	sb.WriteString(`           xmlns:aux="http://ns.adobe.com/exif/1.0/aux/">` + "\n")
	sb.WriteString(`    <rdf:Description rdf:about=""` + "\n")

	// Add GPS coordinates if available
	if hasCoords {
		sb.WriteString(fmt.Sprintf(`      exif:GPSLatitude="%s"`+"\n", lat))
		sb.WriteString(fmt.Sprintf(`      exif:GPSLongitude="%s"`+"\n", lon))
		sb.WriteString(fmt.Sprintf(`      exif:GPSAltitude="%s"`+"\n", alt))
		sb.WriteString(`      exif:GPSAltitudeRef="0"` + "\n") // 0 = above WGS84 ellipsoid
		sb.WriteString(fmt.Sprintf(`      exif:GPSSpeed="%s"`+"\n", speed))
		sb.WriteString(`      exif:GPSSpeedRef="K"` + "\n") // km/h
	}

	// Add timestamps
	if timestamp != "" {
		sb.WriteString(fmt.Sprintf(`      exif:GPSTimeStamp="%s"`+"\n", timestamp))
		sb.WriteString(fmt.Sprintf(`      exif:DateTimeOriginal="%s"`+"\n", dateTime))
		sb.WriteString(fmt.Sprintf(`      xmp:CreateDate="%s"`+"\n", dateTime))
	}

	// Add timezone information if available
	if timezone != "" {
		sb.WriteString(fmt.Sprintf(`      exifEX:TimeZone="%s"`+"\n", timezone))
		if dateTimeWithTZ != "" {
			sb.WriteString(fmt.Sprintf(`      exifEX:DateTimeOriginalTZ="%s"`+"\n", dateTimeWithTZ))
		}
	}

	// Add GPS lock delay (custom field)
	if lockDelay != "" {
		sb.WriteString(fmt.Sprintf(`      exifEX:GPSLockDelay="%s"`+"\n", lockDelay))
	}

	// Add GPS fix type if available
	if gpsData.GPSFix != nil {
		sb.WriteString(fmt.Sprintf(`      aux:GPSFix="%s"`+"\n", *gpsData.GPSFix))
	}

	// Add GPS precision if available
	if gpsData.GPSPrecision != nil {
		sb.WriteString(fmt.Sprintf(`      aux:GPSPrecisionDOP="%.2f"`+"\n", *gpsData.GPSPrecision))
	}

	// Add track point count if we have coordinates
	if hasCoords {
		sb.WriteString(fmt.Sprintf(`      aux:GPSTrackPoints="%d"`+"\n", len(gpsData.Coordinates)))
	}

	// Add processing info
	sb.WriteString(fmt.Sprintf(`      aux:ProcessedBy="GoPro Metadata Validator %s"/>`+"\n", Version))
	sb.WriteString(`  </rdf:RDF>` + "\n")
	sb.WriteString(`</x:xmpmeta>` + "\n")

	return sb.String()
}
