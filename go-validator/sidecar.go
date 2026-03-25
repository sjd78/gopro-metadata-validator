package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteSidecarFile creates an XMP sidecar file with GPMF metadata
func WriteSidecarFile(result *ValidationResult, dryRun bool) error {
	if result.GPSData == nil || !result.GPSData.HasValidGPS {
		return nil // Skip if no GPS data
	}

	// Generate sidecar filename: video.mp4 -> video.mp4.xmp
	sidecarPath := result.FilePath + ".xmp"

	if dryRun {
		fmt.Printf("📋 Would write sidecar: %s\n", sidecarPath)
		return nil
	}

	// Generate XMP content
	xmpContent := generateXMP(result)

	// Write to file
	if err := os.WriteFile(sidecarPath, []byte(xmpContent), 0644); err != nil {
		return fmt.Errorf("failed to write sidecar: %w", err)
	}

	fmt.Printf("✓ Wrote sidecar: %s\n", filepath.Base(sidecarPath))
	return nil
}

func generateXMP(result *ValidationResult) string {
	gps := result.GPSData

	// Calculate GPS lock delay
	var lockDelay string
	if gps.FirstTimestampMs != nil {
		lockDelay = fmt.Sprintf("%.1f", float64(*gps.FirstTimestampMs)/1000.0)
	}

	// Get GPS timestamp
	timestamp := ""
	dateTime := ""
	if gps.FirstGPSTime != nil {
		timestamp = gps.FirstGPSTime.Format("2006-01-02T15:04:05Z")
		dateTime = gps.FirstGPSTime.Format("2006-01-02T15:04:05Z")
	}

	// Get coordinates if available
	var lat, lon, alt, speed string
	hasCoords := len(gps.Coordinates) > 0
	if hasCoords {
		first := gps.Coordinates[0]
		lat = fmt.Sprintf("%.6f", first.Latitude)
		lon = fmt.Sprintf("%.6f", first.Longitude)
		alt = fmt.Sprintf("%.1f", first.Altitude)
		speed = fmt.Sprintf("%.2f", first.Speed2D*3.6) // m/s to km/h
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
		sb.WriteString(fmt.Sprintf(`      exif:GPSLatitude="%s"` + "\n", lat))
		sb.WriteString(fmt.Sprintf(`      exif:GPSLongitude="%s"` + "\n", lon))
		sb.WriteString(fmt.Sprintf(`      exif:GPSAltitude="%s"` + "\n", alt))
		sb.WriteString(fmt.Sprintf(`      exif:GPSSpeed="%s"` + "\n", speed))
		sb.WriteString(`      exif:GPSSpeedRef="K"` + "\n") // km/h
	}

	// Add timestamps
	if timestamp != "" {
		sb.WriteString(fmt.Sprintf(`      exif:GPSTimeStamp="%s"` + "\n", timestamp))
		sb.WriteString(fmt.Sprintf(`      exif:DateTimeOriginal="%s"` + "\n", dateTime))
		sb.WriteString(fmt.Sprintf(`      xmp:CreateDate="%s"` + "\n", dateTime))
	}

	// Add GPS lock delay (custom field)
	if lockDelay != "" {
		sb.WriteString(fmt.Sprintf(`      exifEX:GPSLockDelay="%s"` + "\n", lockDelay))
	}

	// Add GPS fix type if available
	if gps.GPSFix != nil {
		sb.WriteString(fmt.Sprintf(`      aux:GPSFix="%s"` + "\n", *gps.GPSFix))
	}

	// Add GPS precision if available
	if gps.GPSPrecision != nil {
		sb.WriteString(fmt.Sprintf(`      aux:GPSPrecisionDOP="%.2f"` + "\n", *gps.GPSPrecision))
	}

	// Add track point count if we have coordinates
	if hasCoords {
		sb.WriteString(fmt.Sprintf(`      aux:GPSTrackPoints="%d"` + "\n", len(gps.Coordinates)))
	}

	// Add processing info
	sb.WriteString(fmt.Sprintf(`      aux:ProcessedBy="GoPro Metadata Validator %s"/>` + "\n", Version))
	sb.WriteString(`  </rdf:RDF>` + "\n")
	sb.WriteString(`</x:xmpmeta>` + "\n")

	return sb.String()
}
