package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GPMF KLV (Key-Length-Value) structure
type gpmfKLV struct {
	Key        [4]byte
	Type       byte
	StructSize byte
	Repeat     uint16
	Data       []byte
}

func extractGPMF(filePath string) (*GPSData, error) {
	result := &GPSData{
		SampleCount: 0,
		HasValidGPS: false,
	}

	// Extract GPMF stream using ffmpeg
	gpmfData, err := extractGPMFStream(filePath)
	if err != nil {
		return result, err
	}

	if len(gpmfData) == 0 {
		return result, fmt.Errorf("no GPMF data found")
	}

	// Parse GPMF data - returns complete GPSData
	gpsData, err := parseGPMFData(gpmfData)
	if err != nil {
		return result, err
	}

	return gpsData, nil
}

func extractGPMFStream(filePath string) ([]byte, error) {
	// First, find the GPMF stream index
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "stream=index,codec_tag_string",
		"-of", "csv=p=0",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	// Find stream with codec_tag 'gpmd'
	streamIndex := -1
	lines := bytes.Split(output, []byte("\n"))
	for _, line := range lines {
		parts := bytes.Split(line, []byte(","))
		if len(parts) == 2 && string(parts[1]) == "gpmd" {
			fmt.Sscanf(string(parts[0]), "%d", &streamIndex)
			break
		}
	}

	if streamIndex == -1 {
		return nil, fmt.Errorf("no GPMF stream found")
	}

	// Create temporary file for GPMF data
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("gpmf-%d.bin", os.Getpid()))
	defer os.Remove(tmpFile)

	// Extract the GPMF stream
	cmd = exec.Command("ffmpeg",
		"-y",
		"-i", filePath,
		"-map", fmt.Sprintf("0:%d", streamIndex),
		"-codec", "copy",
		"-f", "rawvideo",
		tmpFile,
	)

	// Suppress ffmpeg output
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg extraction failed: %w", err)
	}

	// Read the extracted data
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read extracted GPMF: %w", err)
	}

	return data, nil
}

func parseGPMFData(data []byte) (*GPSData, error) {
	result := &GPSData{
		SampleCount: 0,
		HasValidGPS: false,
		Coordinates: make([]GPSCoordinate, 0),
	}

	timestamps := make([]int64, 0)
	gpsTimes := make([]time.Time, 0)
	var scaleFactors []int32
	var currentTimestamp int64 // Track current TSMP for pairing with GPS5

	buf := bytes.NewReader(data)

	for buf.Len() > 8 { // Need at least 8 bytes for header
		klv, err := readKLV(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			// Try to skip ahead
			if buf.Len() > 0 {
				buf.ReadByte()
			}
			continue
		}

		keyStr := string(klv.Key[:])

		// Look for TSMP (GPS timestamp in milliseconds) - must parse before GPS5
		if keyStr == "TSMP" && klv.Type == 'L' { // 'L' is unsigned long (uint32)
			if len(klv.Data) >= 4 {
				ts := binary.BigEndian.Uint32(klv.Data[:4])
				currentTimestamp = int64(ts)
				timestamps = append(timestamps, currentTimestamp)
			}
		}

		// Look for STMP (sample timestamp)
		if keyStr == "STMP" && klv.Type == 'J' {
			if len(klv.Data) >= 4 {
				ts := binary.BigEndian.Uint32(klv.Data[:4])
				currentTimestamp = int64(ts)
				timestamps = append(timestamps, currentTimestamp)
			}
		}

		// Look for SCAL (scale factors) - must parse before GPS5
		if keyStr == "SCAL" && klv.Type == 'l' {
			scaleFactors = make([]int32, 0)
			for i := 0; i < int(klv.Repeat); i++ {
				offset := i * 4
				if offset+4 <= len(klv.Data) {
					scale := int32(binary.BigEndian.Uint32(klv.Data[offset : offset+4]))
					scaleFactors = append(scaleFactors, scale)
				}
			}
		}

		// Look for GPS5 (latitude, longitude, altitude, speed2D, speed3D)
		// Note: Type can be 's' (int16) or 'l' (int32) depending on camera model
		if keyStr == "GPS5" && (klv.Type == 's' || klv.Type == 'l') {
			var bytesPerValue int
			if klv.Type == 's' {
				bytesPerValue = 2 // int16
			} else {
				bytesPerValue = 4 // int32
			}

			samplesPerEntry := 5
			bytesPerSample := bytesPerValue * samplesPerEntry

			for i := 0; i < int(klv.Repeat); i++ {
				offset := i * bytesPerSample

				// Bounds check before reading
				if offset+bytesPerSample > len(klv.Data) {
					continue // Skip incomplete sample
				}

				var lat, lon, alt, speed2D, speed3D int32

				if klv.Type == 's' {
					// Read 5 int16 values (cast Uint16 to int16 for signed values)
					lat = int32(int16(binary.BigEndian.Uint16(klv.Data[offset : offset+2])))
					lon = int32(int16(binary.BigEndian.Uint16(klv.Data[offset+2 : offset+4])))
					alt = int32(int16(binary.BigEndian.Uint16(klv.Data[offset+4 : offset+6])))
					speed2D = int32(int16(binary.BigEndian.Uint16(klv.Data[offset+6 : offset+8])))
					speed3D = int32(int16(binary.BigEndian.Uint16(klv.Data[offset+8 : offset+10])))
				} else {
					// Read 5 int32 values
					lat = int32(binary.BigEndian.Uint32(klv.Data[offset : offset+4]))
					lon = int32(binary.BigEndian.Uint32(klv.Data[offset+4 : offset+8]))
					alt = int32(binary.BigEndian.Uint32(klv.Data[offset+8 : offset+12]))
					speed2D = int32(binary.BigEndian.Uint32(klv.Data[offset+12 : offset+16]))
					speed3D = int32(binary.BigEndian.Uint32(klv.Data[offset+16 : offset+20]))
				}

				// Apply scale factors (use confirmed standard defaults if SCAL missing)
				latScale := int32(10000000)
				lonScale := int32(10000000)
				altScale := int32(100) // Altitude uses 100, not 1000
				speed2DScale := int32(1000)
				speed3DScale := int32(100) // 3D speed uses 100

				if len(scaleFactors) >= 5 {
					latScale = scaleFactors[0]
					lonScale = scaleFactors[1]
					altScale = scaleFactors[2]
					speed2DScale = scaleFactors[3]
					speed3DScale = scaleFactors[4]
				}

				// Create coordinate entry
				coord := GPSCoordinate{
					Timestamp: currentTimestamp, // From paired TSMP (relative ms)
					Latitude:  float64(lat) / float64(latScale),
					Longitude: float64(lon) / float64(lonScale),
					Altitude:  float64(alt) / float64(altScale), // WGS84 ellipsoid height
					Speed2D:   float64(speed2D) / float64(speed2DScale),
					Speed3D:   float64(speed3D) / float64(speed3DScale),
				}

				result.Coordinates = append(result.Coordinates, coord)
			}
		}

		// Look for GPSF (GPS fix type) - quality indicator
		if keyStr == "GPSF" && klv.Type == 'L' {
			if len(klv.Data) >= 4 {
				fix := binary.BigEndian.Uint32(klv.Data[:4])
				var gpsFix string
				switch fix {
				case 0:
					gpsFix = "NONE" // Invalid GPS data
				case 2:
					gpsFix = "2D" // Lat/lon valid, altitude unreliable
				case 3:
					gpsFix = "3D" // All values valid
				default:
					gpsFix = "UNKNOWN"
				}
				result.GPSFix = &gpsFix
			}
		}

		// Look for GPSP (GPS precision/DOP) - accuracy indicator
		if keyStr == "GPSP" && klv.Type == 'H' {
			if len(klv.Data) >= 2 {
				dop := float64(binary.BigEndian.Uint16(klv.Data[:2])) / 100.0
				result.GPSPrecision = &dop
			}
		}

		// Look for GPSUU (GPS UTC time) - contains datetime strings like "240222170635.690"
		// Can also be GPSU (older format)
		if (keyStr == "GPSU" || strings.HasPrefix(keyStr, "GPSU")) && (klv.Type == 'c' || klv.Type == 'U') {
			gpsTimeStr := strings.TrimSpace(string(klv.Data))
			// Remove any null terminators or trailing data
			if idx := strings.IndexByte(gpsTimeStr, 0); idx != -1 {
				gpsTimeStr = gpsTimeStr[:idx]
			}
			gpsTimeStr = strings.TrimSpace(gpsTimeStr)

			if len(gpsTimeStr) >= 12 {
				// Format: YYMMDDHHMMSS.sss
				gpsTime, err := parseGPSUTime(gpsTimeStr)
				if err == nil {
					gpsTimes = append(gpsTimes, gpsTime)
				}
			}
		}

		// Recursively parse nested structures (DEVC, STRM)
		if (keyStr == "DEVC" || keyStr == "STRM") && len(klv.Data) > 0 {
			nestedData, _ := parseGPMFData(klv.Data)
			if nestedData != nil {
				timestamps = append(timestamps, extractTimestampsFromGPSData(nestedData)...)
				if nestedData.FirstGPSTime != nil {
					gpsTimes = append(gpsTimes, *nestedData.FirstGPSTime)
				}
				if nestedData.LastGPSTime != nil && nestedData.LastGPSTime != nestedData.FirstGPSTime {
					gpsTimes = append(gpsTimes, *nestedData.LastGPSTime)
				}
				result.Coordinates = append(result.Coordinates, nestedData.Coordinates...)
				if nestedData.GPSFix != nil && result.GPSFix == nil {
					result.GPSFix = nestedData.GPSFix
				}
				if nestedData.GPSPrecision != nil && result.GPSPrecision == nil {
					result.GPSPrecision = nestedData.GPSPrecision
				}
			}
		}
	}

	// Populate result fields
	result.SampleCount = len(timestamps)
	if len(timestamps) > 0 {
		result.HasValidGPS = true
		first := timestamps[0]
		last := timestamps[len(timestamps)-1]
		result.FirstTimestampMs = &first
		result.LastTimestampMs = &last
	}

	// Add absolute GPS times if available
	if len(gpsTimes) > 0 {
		result.FirstGPSTime = &gpsTimes[0]
		result.LastGPSTime = &gpsTimes[len(gpsTimes)-1]
	}

	return result, nil
}

// Helper to extract timestamps from GPSData for recursive parsing
func extractTimestampsFromGPSData(data *GPSData) []int64 {
	timestamps := make([]int64, 0)
	if data.FirstTimestampMs != nil {
		timestamps = append(timestamps, *data.FirstTimestampMs)
	}
	if data.LastTimestampMs != nil && data.LastTimestampMs != data.FirstTimestampMs {
		timestamps = append(timestamps, *data.LastTimestampMs)
	}
	return timestamps
}

func parseGPSUTime(gpsTimeStr string) (time.Time, error) {
	// Format: YYMMDDHHMMSS.sss (e.g., "240222170635.690")
	// Year is 2-digit, need to convert to 4-digit
	if len(gpsTimeStr) < 12 {
		return time.Time{}, fmt.Errorf("GPS time string too short: %s", gpsTimeStr)
	}

	// Extract components
	year := "20" + gpsTimeStr[0:2]
	month := gpsTimeStr[2:4]
	day := gpsTimeStr[4:6]
	hour := gpsTimeStr[6:8]
	minute := gpsTimeStr[8:10]
	second := gpsTimeStr[10:12]

	// Milliseconds if present
	millis := "000"
	if len(gpsTimeStr) > 13 && gpsTimeStr[12] == '.' {
		millis = gpsTimeStr[13:16]
	}

	// Parse as UTC time
	timeStr := fmt.Sprintf("%s-%s-%sT%s:%s:%s.%sZ", year, month, day, hour, minute, second, millis)
	return time.Parse(time.RFC3339, timeStr)
}

func readKLV(r *bytes.Reader) (*gpmfKLV, error) {
	klv := &gpmfKLV{}

	// Read key (4 bytes)
	if _, err := r.Read(klv.Key[:]); err != nil {
		return nil, err
	}

	// Read type (1 byte)
	if err := binary.Read(r, binary.BigEndian, &klv.Type); err != nil {
		return nil, err
	}

	// Read struct size (1 byte)
	if err := binary.Read(r, binary.BigEndian, &klv.StructSize); err != nil {
		return nil, err
	}

	// Read repeat count (2 bytes, big endian)
	if err := binary.Read(r, binary.BigEndian, &klv.Repeat); err != nil {
		return nil, err
	}

	// Calculate data size
	dataSize := int(klv.StructSize) * int(klv.Repeat)

	// Align to 4-byte boundary
	alignedSize := (dataSize + 3) & ^3

	// Read data
	klv.Data = make([]byte, alignedSize)
	if _, err := r.Read(klv.Data); err != nil {
		return nil, err
	}

	// Trim to actual data size (remove padding)
	klv.Data = klv.Data[:dataSize]

	return klv, nil
}
