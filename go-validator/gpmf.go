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

	// Parse GPMF data for both relative timestamps and absolute GPS times
	timestamps, gpsTimes, err := parseGPMFData(gpmfData)
	if err != nil {
		return result, err
	}

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

func parseGPMFData(data []byte) ([]int64, []time.Time, error) {
	timestamps := make([]int64, 0)
	gpsTimes := make([]time.Time, 0)
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

		// Look for TSMP (GPS timestamp in milliseconds)
		if keyStr == "TSMP" && klv.Type == 'L' { // 'L' is unsigned long (uint32)
			if len(klv.Data) >= 4 {
				ts := binary.BigEndian.Uint32(klv.Data[:4])
				timestamps = append(timestamps, int64(ts))
			}
		}

		// Look for STMP (sample timestamp)
		if keyStr == "STMP" && klv.Type == 'J' {
			if len(klv.Data) >= 4 {
				ts := binary.BigEndian.Uint32(klv.Data[:4])
				timestamps = append(timestamps, int64(ts))
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
			nestedTimestamps, nestedGPSTimes, _ := parseGPMFData(klv.Data)
			timestamps = append(timestamps, nestedTimestamps...)
			gpsTimes = append(gpsTimes, nestedGPSTimes...)
		}
	}

	return timestamps, gpsTimes, nil
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
