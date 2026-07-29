package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	mp4 "github.com/abema/go-mp4"
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

	gpmfData, err := extractGPMFStream(filePath)
	if err != nil {
		return result, err
	}

	if len(gpmfData) == 0 {
		return result, fmt.Errorf("no GPMF data found")
	}

	gpsData, err := parseGPMFData(gpmfData)
	if err != nil {
		return result, err
	}

	return gpsData, nil
}

// extractGPMFStream reads the raw GPMF binary payload from an MP4 file using
// github.com/abema/go-mp4.  It locates the "GoPro MET" handler track, then
// reads each sample chunk directly from the file via io.ReadSeeker — no
// external processes, no temp files, no full-file buffering.
func extractGPMFStream(filePath string) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	// ── 1. Find the GoPro MET metadata track ──────────────────────────────
	trakBoxes, err := mp4.ExtractBox(f, nil,
		mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeTrak()})
	if err != nil || len(trakBoxes) == 0 {
		return nil, fmt.Errorf("no tracks found in %s", filePath)
	}

	var gpmdTrak *mp4.BoxInfo
	for _, trak := range trakBoxes {
		hdlrBoxes, err := mp4.ExtractBoxWithPayload(f, trak,
			mp4.BoxPath{mp4.BoxTypeMdia(), mp4.BoxTypeHdlr()})
		if err != nil || len(hdlrBoxes) == 0 {
			continue
		}
		hdlr := hdlrBoxes[0].Payload.(*mp4.Hdlr)
		if strings.TrimSpace(hdlr.Name) == "GoPro MET" {
			gpmdTrak = trak
			break
		}
	}
	if gpmdTrak == nil {
		return nil, fmt.Errorf("no GPMF stream found (no 'GoPro MET' handler track)")
	}

	// ── 2. Read the sample table to find chunk offsets + sizes ────────────
	// stco (32-bit offsets) or co64 (64-bit offsets for >4 GB files)
	var chunkOffsets []uint64

	stcoBoxes, _ := mp4.ExtractBoxWithPayload(f, gpmdTrak,
		mp4.BoxPath{mp4.BoxTypeMdia(), mp4.BoxTypeMinf(),
			mp4.BoxTypeStbl(), mp4.BoxTypeStco()})
	if len(stcoBoxes) > 0 {
		for _, o := range stcoBoxes[0].Payload.(*mp4.Stco).ChunkOffset {
			chunkOffsets = append(chunkOffsets, uint64(o))
		}
	} else {
		co64Boxes, _ := mp4.ExtractBoxWithPayload(f, gpmdTrak,
			mp4.BoxPath{mp4.BoxTypeMdia(), mp4.BoxTypeMinf(),
				mp4.BoxTypeStbl(), mp4.BoxTypeCo64()})
		if len(co64Boxes) > 0 {
			chunkOffsets = co64Boxes[0].Payload.(*mp4.Co64).ChunkOffset
		}
	}
	if len(chunkOffsets) == 0 {
		return nil, fmt.Errorf("no chunk offsets found in GPMF track")
	}

	// stsz — per-sample sizes
	stszBoxes, err := mp4.ExtractBoxWithPayload(f, gpmdTrak,
		mp4.BoxPath{mp4.BoxTypeMdia(), mp4.BoxTypeMinf(),
			mp4.BoxTypeStbl(), mp4.BoxTypeStsz()})
	if err != nil || len(stszBoxes) == 0 {
		return nil, fmt.Errorf("no stsz box in GPMF track")
	}
	stsz := stszBoxes[0].Payload.(*mp4.Stsz)
	var sampleSizes []uint32
	if stsz.SampleSize != 0 {
		sampleSizes = make([]uint32, stsz.SampleCount)
		for i := range sampleSizes {
			sampleSizes[i] = stsz.SampleSize
		}
	} else {
		sampleSizes = stsz.EntrySize
	}

	// stsc — sample-to-chunk mapping (first_chunk, samples_per_chunk, …)
	stscBoxes, err := mp4.ExtractBoxWithPayload(f, gpmdTrak,
		mp4.BoxPath{mp4.BoxTypeMdia(), mp4.BoxTypeMinf(),
			mp4.BoxTypeStbl(), mp4.BoxTypeStsc()})
	if err != nil || len(stscBoxes) == 0 {
		return nil, fmt.Errorf("no stsc box in GPMF track")
	}
	stscEntries := stscBoxes[0].Payload.(*mp4.Stsc).Entries

	// ── 3. Walk chunk/sample table and read each sample sequentially ──────
	// Build a flat list of (chunkIndex, sampleSize) pairs following the
	// stsc compact encoding, then stream-read from the file.
	type sampleRef struct {
		chunkIdx uint32 // 0-based index into chunkOffsets
		size     uint32
	}
	refs := make([]sampleRef, 0, len(sampleSizes))

	totalChunks := uint32(len(chunkOffsets))
	sampleIdx := uint32(0)

	for i, entry := range stscEntries {
		firstChunk := entry.FirstChunk - 1 // convert to 0-based
		var lastChunk uint32
		if i+1 < len(stscEntries) {
			lastChunk = stscEntries[i+1].FirstChunk - 2 // inclusive, 0-based
		} else {
			lastChunk = totalChunks - 1
		}

		for chunkIdx := firstChunk; chunkIdx <= lastChunk; chunkIdx++ {
			for s := uint32(0); s < entry.SamplesPerChunk; s++ {
				if sampleIdx >= uint32(len(sampleSizes)) {
					break
				}
				refs = append(refs, sampleRef{
					chunkIdx: chunkIdx,
					size:     sampleSizes[sampleIdx],
				})
				sampleIdx++
			}
		}
	}

	// Calculate total size and read all samples
	var totalSize uint64
	for _, r := range refs {
		totalSize += uint64(r.size)
	}

	buf := make([]byte, totalSize)
	var writePos uint64

	// Track current chunk position to avoid redundant seeks
	var currentChunkIdx uint32 = ^uint32(0)
	var currentChunkOffset uint64

	for _, r := range refs {
		if r.chunkIdx != currentChunkIdx {
			currentChunkIdx = r.chunkIdx
			currentChunkOffset = chunkOffsets[r.chunkIdx]
			if _, err := f.Seek(int64(currentChunkOffset), io.SeekStart); err != nil {
				return nil, fmt.Errorf("seek to chunk %d: %w", r.chunkIdx, err)
			}
		}
		if _, err := io.ReadFull(f, buf[writePos:writePos+uint64(r.size)]); err != nil {
			return nil, fmt.Errorf("read sample: %w", err)
		}
		writePos += uint64(r.size)
		currentChunkOffset += uint64(r.size)
	}

	return buf, nil
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
	var currentTimestamp int64 // relative ms paired with the next GPS5 block

	buf := bytes.NewReader(data)

	for buf.Len() > 8 {
		klv, err := readKLV(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			// Skip one byte and try to re-sync
			if buf.Len() > 0 {
				buf.ReadByte()
			}
			continue
		}

		keyStr := string(klv.Key[:])

		// TSMP — Total SaMPle count (running counter, not a timestamp).
		// We record it as a relative-ms placeholder only when we have no
		// STMP data; it is superseded by STMP when present.
		if keyStr == "TSMP" && klv.Type == 'L' {
			if len(klv.Data) >= 4 {
				// TSMP is a sample counter; use it as a last-resort ordering
				// token only — actual timing comes from STMP or GPSU.
				count := int64(binary.BigEndian.Uint32(klv.Data[:4]))
				currentTimestamp = count
				timestamps = append(timestamps, currentTimestamp)
			}
		}

		// STMP — Sample TiMesteP in microseconds (uint64, type 'J').
		// Convert µs → ms for internal use.
		if keyStr == "STMP" && klv.Type == 'J' {
			if len(klv.Data) >= 8 {
				us := binary.BigEndian.Uint64(klv.Data[:8])
				currentTimestamp = int64(us / 1000) // µs → ms
				timestamps = append(timestamps, currentTimestamp)
			}
		}

		// SCAL — scale factor(s), scoped to the current STRM container.
		// Type 'l' = []int32, type 's' = []int16, type 'S' = []uint16.
		if keyStr == "SCAL" {
			scaleFactors = parseScaleFactors(klv)
		}

		// GPS5 — lat, lon, alt, speed2D, speed3D (int16 or int32 per sample)
		if keyStr == "GPS5" && (klv.Type == 's' || klv.Type == 'l') {
			coords := parseGPS5(klv, scaleFactors, currentTimestamp)
			result.Coordinates = append(result.Coordinates, coords...)
		}

		// GPSF — GPS fix type (0=none, 2=2D, 3=3D)
		if keyStr == "GPSF" && klv.Type == 'L' {
			if len(klv.Data) >= 4 {
				fix := binary.BigEndian.Uint32(klv.Data[:4])
				var gpsFix string
				switch fix {
				case 0:
					gpsFix = "NONE"
				case 2:
					gpsFix = "2D"
				case 3:
					gpsFix = "3D"
				default:
					gpsFix = "UNKNOWN"
				}
				result.GPSFix = &gpsFix
			}
		}

		// GPSP — GPS precision / DOP (uint16, scaled ×100)
		if keyStr == "GPSP" && klv.Type == 'H' {
			if len(klv.Data) >= 2 {
				dop := float64(binary.BigEndian.Uint16(klv.Data[:2])) / 100.0
				result.GPSPrecision = &dop
			}
		}

		// GPSU / GPSUU — GPS absolute UTC datetime string "YYMMDDHHMMSS.sss"
		// Type 'U' is the canonical GPMF UTC string type; 'c' also seen in wild.
		if (keyStr == "GPSU" || strings.HasPrefix(keyStr, "GPSU")) &&
			(klv.Type == 'U' || klv.Type == 'c') {
			gpsTimeStr := strings.TrimSpace(string(klv.Data))
			if idx := strings.IndexByte(gpsTimeStr, 0); idx != -1 {
				gpsTimeStr = gpsTimeStr[:idx]
			}
			gpsTimeStr = strings.TrimSpace(gpsTimeStr)
			if len(gpsTimeStr) >= 12 {
				if t, err := parseGPSUTime(gpsTimeStr); err == nil {
					gpsTimes = append(gpsTimes, t)
				}
			}
		}

		// DEVC / STRM — recurse into nested containers.
		// Reset per-stream state (scaleFactors) on each STRM so that a
		// SCAL from one sensor stream cannot bleed into another.
		if keyStr == "DEVC" && len(klv.Data) > 0 {
			nested, _ := parseGPMFData(klv.Data)
			if nested != nil {
				mergeGPSData(result, nested, &timestamps, &gpsTimes)
			}
		}
		if keyStr == "STRM" && len(klv.Data) > 0 {
			// Each STRM is self-contained: parse with fresh state so its
			// own SCAL applies only to its own GPS5/ACCL/GYRO data.
			nested, _ := parseGPMFData(klv.Data)
			if nested != nil {
				mergeGPSData(result, nested, &timestamps, &gpsTimes)
			}
		}
	}

	// Populate summary fields
	result.SampleCount = len(timestamps)
	if len(timestamps) > 0 {
		result.HasValidGPS = true
		first := timestamps[0]
		last := timestamps[len(timestamps)-1]
		result.FirstTimestampMs = &first
		result.LastTimestampMs = &last
	}
	if len(gpsTimes) > 0 {
		result.FirstGPSTime = &gpsTimes[0]
		result.LastGPSTime = &gpsTimes[len(gpsTimes)-1]
	}

	return result, nil
}

// mergeGPSData folds nested parse results back into the parent, appending
// timestamps and GPS times and taking the first non-nil quality fields.
func mergeGPSData(parent, child *GPSData, timestamps *[]int64, gpsTimes *[]time.Time) {
	if child.FirstTimestampMs != nil {
		*timestamps = append(*timestamps, *child.FirstTimestampMs)
	}
	if child.LastTimestampMs != nil &&
		(child.FirstTimestampMs == nil || *child.LastTimestampMs != *child.FirstTimestampMs) {
		*timestamps = append(*timestamps, *child.LastTimestampMs)
	}
	if child.FirstGPSTime != nil {
		*gpsTimes = append(*gpsTimes, *child.FirstGPSTime)
	}
	if child.LastGPSTime != nil &&
		(child.FirstGPSTime == nil || !child.LastGPSTime.Equal(*child.FirstGPSTime)) {
		*gpsTimes = append(*gpsTimes, *child.LastGPSTime)
	}
	parent.Coordinates = append(parent.Coordinates, child.Coordinates...)
	if child.GPSFix != nil && parent.GPSFix == nil {
		parent.GPSFix = child.GPSFix
	}
	if child.GPSPrecision != nil && parent.GPSPrecision == nil {
		parent.GPSPrecision = child.GPSPrecision
	}
}

// parseScaleFactors extracts SCAL values for the current STRM context.
// SCAL can be int16 ('s'), int32 ('l'), or uint16 ('S').
func parseScaleFactors(klv *gpmfKLV) []int32 {
	out := make([]int32, 0, int(klv.Repeat))
	switch klv.Type {
	case 'l': // int32
		for i := 0; i < int(klv.Repeat); i++ {
			off := i * 4
			if off+4 > len(klv.Data) {
				break
			}
			out = append(out, int32(binary.BigEndian.Uint32(klv.Data[off:off+4])))
		}
	case 's': // int16
		for i := 0; i < int(klv.Repeat); i++ {
			off := i * 2
			if off+2 > len(klv.Data) {
				break
			}
			out = append(out, int32(int16(binary.BigEndian.Uint16(klv.Data[off : off+2]))))
		}
	case 'S': // uint16
		for i := 0; i < int(klv.Repeat); i++ {
			off := i * 2
			if off+2 > len(klv.Data) {
				break
			}
			out = append(out, int32(binary.BigEndian.Uint16(klv.Data[off:off+2])))
		}
	}
	return out
}

// parseGPS5 decodes a GPS5 KLV payload into GPSCoordinate slices.
// Applies per-stream SCAL values; falls back to GoPro defaults when absent.
func parseGPS5(klv *gpmfKLV, scaleFactors []int32, timestamp int64) []GPSCoordinate {
	var bytesPerValue int
	if klv.Type == 's' {
		bytesPerValue = 2 // int16
	} else {
		bytesPerValue = 4 // int32
	}

	const fieldsPerSample = 5 // lat, lon, alt, speed2D, speed3D
	bytesPerSample := bytesPerValue * fieldsPerSample

	// GoPro-documented defaults when SCAL is missing or incomplete
	latScale := int32(10000000)
	lonScale := int32(10000000)
	altScale := int32(100)
	speed2DScale := int32(1000)
	speed3DScale := int32(100)

	if len(scaleFactors) >= 5 {
		latScale = scaleFactors[0]
		lonScale = scaleFactors[1]
		altScale = scaleFactors[2]
		speed2DScale = scaleFactors[3]
		speed3DScale = scaleFactors[4]
	} else if len(scaleFactors) == 1 {
		// Single scalar — apply to all fields
		latScale = scaleFactors[0]
		lonScale = scaleFactors[0]
		altScale = scaleFactors[0]
		speed2DScale = scaleFactors[0]
		speed3DScale = scaleFactors[0]
	}

	coords := make([]GPSCoordinate, 0, int(klv.Repeat))
	for i := 0; i < int(klv.Repeat); i++ {
		off := i * bytesPerSample
		if off+bytesPerSample > len(klv.Data) {
			break
		}

		var lat, lon, alt, speed2D, speed3D int32
		if klv.Type == 's' {
			lat = int32(int16(binary.BigEndian.Uint16(klv.Data[off:])))
			lon = int32(int16(binary.BigEndian.Uint16(klv.Data[off+2:])))
			alt = int32(int16(binary.BigEndian.Uint16(klv.Data[off+4:])))
			speed2D = int32(int16(binary.BigEndian.Uint16(klv.Data[off+6:])))
			speed3D = int32(int16(binary.BigEndian.Uint16(klv.Data[off+8:])))
		} else {
			lat = int32(binary.BigEndian.Uint32(klv.Data[off:]))
			lon = int32(binary.BigEndian.Uint32(klv.Data[off+4:]))
			alt = int32(binary.BigEndian.Uint32(klv.Data[off+8:]))
			speed2D = int32(binary.BigEndian.Uint32(klv.Data[off+12:]))
			speed3D = int32(binary.BigEndian.Uint32(klv.Data[off+16:]))
		}

		coords = append(coords, GPSCoordinate{
			Timestamp: timestamp,
			Latitude:  float64(lat) / float64(latScale),
			Longitude: float64(lon) / float64(lonScale),
			Altitude:  float64(alt) / float64(altScale),
			Speed2D:   float64(speed2D) / float64(speed2DScale),
			Speed3D:   float64(speed3D) / float64(speed3DScale),
		})
	}
	return coords
}

func parseGPSUTime(gpsTimeStr string) (time.Time, error) {
	// Format: YYMMDDHHMMSS.sss  e.g. "240222170635.690"
	if len(gpsTimeStr) < 12 {
		return time.Time{}, fmt.Errorf("GPS time string too short: %s", gpsTimeStr)
	}

	year := "20" + gpsTimeStr[0:2]
	month := gpsTimeStr[2:4]
	day := gpsTimeStr[4:6]
	hour := gpsTimeStr[6:8]
	minute := gpsTimeStr[8:10]
	second := gpsTimeStr[10:12]

	millis := "000"
	if len(gpsTimeStr) >= 16 && gpsTimeStr[12] == '.' {
		millis = gpsTimeStr[13:16]
	} else if len(gpsTimeStr) > 13 && gpsTimeStr[12] == '.' {
		millis = (gpsTimeStr[13:] + "000")[:3]
	}

	timeStr := fmt.Sprintf("%s-%s-%sT%s:%s:%s.%sZ",
		year, month, day, hour, minute, second, millis)
	return time.Parse(time.RFC3339, timeStr)
}

func readKLV(r *bytes.Reader) (*gpmfKLV, error) {
	klv := &gpmfKLV{}

	if _, err := r.Read(klv.Key[:]); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &klv.Type); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &klv.StructSize); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &klv.Repeat); err != nil {
		return nil, err
	}

	dataSize := int(klv.StructSize) * int(klv.Repeat)
	alignedSize := (dataSize + 3) &^ 3

	klv.Data = make([]byte, alignedSize)
	if _, err := r.Read(klv.Data); err != nil {
		return nil, err
	}
	klv.Data = klv.Data[:dataSize]

	return klv, nil
}
