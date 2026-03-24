package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type ffprobeOutput struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		Tags      struct {
			CreationTime string `json:"creation_time"`
			Timecode     string `json:"timecode"`
		} `json:"tags"`
	} `json:"streams"`
	Format struct {
		Tags struct {
			CreationTime string `json:"creation_time"`
		} `json:"tags"`
	} `json:"format"`
}

func extractFileMetadata(filePath string) (*Metadata, error) {
	metadata := &Metadata{}

	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	var data ffprobeOutput
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	// Extract creation_time from video stream
	for _, stream := range data.Streams {
		if stream.CodecType == "video" {
			if stream.Tags.CreationTime != "" {
				t, err := time.Parse(time.RFC3339, stream.Tags.CreationTime)
				if err == nil {
					metadata.CreationTime = &t
				}
			}
			if stream.Tags.Timecode != "" {
				metadata.Timecode = stream.Tags.Timecode
			}
			break
		}
	}

	// Fallback to format-level creation_time
	if metadata.CreationTime == nil && data.Format.Tags.CreationTime != "" {
		t, err := time.Parse(time.RFC3339, data.Format.Tags.CreationTime)
		if err == nil {
			metadata.CreationTime = &t
		}
	}

	return metadata, nil
}
