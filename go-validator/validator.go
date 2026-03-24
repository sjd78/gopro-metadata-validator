package main

import (
	"time"
)

type Metadata struct {
	CreationTime *time.Time
	Timecode     string
}

type GPSData struct {
	FirstTimestampMs *int64
	LastTimestampMs  *int64
	SampleCount      int
	HasValidGPS      bool
	FirstGPSTime     *time.Time // Absolute GPS UTC time from GPMF
	LastGPSTime      *time.Time // Last GPS UTC time
}

type ValidationResult struct {
	FilePath string
	Valid    bool
	Issues   []string
	Metadata *Metadata
	GPSData  *GPSData
}

func validateFile(filePath string) (*ValidationResult, error) {
	result := &ValidationResult{
		FilePath: filePath,
		Valid:    true,
		Issues:   make([]string, 0),
		Metadata: &Metadata{},
		GPSData:  &GPSData{},
	}

	// Extract file metadata
	metadata, err := extractFileMetadata(filePath)
	if err != nil {
		result.Valid = false
		result.Issues = append(result.Issues, "Error extracting metadata: "+err.Error())
		return result, nil
	}
	result.Metadata = metadata

	// Extract GPS data from GPMF stream
	gpsData, err := extractGPMF(filePath)
	if err != nil {
		// GPS extraction errors are not fatal - file might not have GPS
		result.GPSData = &GPSData{}
	} else {
		result.GPSData = gpsData
	}

	// Compare and find discrepancies
	issues := compareMetadata(metadata, gpsData)
	result.Issues = issues
	result.Valid = len(issues) == 0

	return result, nil
}
