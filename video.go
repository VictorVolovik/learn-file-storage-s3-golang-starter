package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
)

type VideoInfo struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
}

const sixteenNineRatio = 16.0 / 9.0 // ≈1.777777...
const ninesixteenRatio = 9.0 / 16.0 // ≈0.5625

func getVideoAspectRatio(filepath string) (string, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-print_format", "json",
		"-show_streams",
		filepath,
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("unable to get video information: %w", err)
	}

	var videoInfo VideoInfo
	err = json.Unmarshal(out.Bytes(), &videoInfo)
	if err != nil {
		return "", fmt.Errorf("unable to unmarshal video information: %w", err)
	}

	if len(videoInfo.Streams) == 0 {
		return "", fmt.Errorf("no streams found in the video file")
	}

	for _, stream := range videoInfo.Streams {
		if stream.CodecType == "video" {
			ratio := float64(stream.Width) / float64(stream.Height)

			if math.Abs(ratio-sixteenNineRatio) < 0.1 {
				return "16:9", nil
			} else if math.Abs(ratio-ninesixteenRatio) < 0.1 {
				return "9:16", nil
			} else {
				return "", nil
			}
		}
	}

	return "", fmt.Errorf("no video stream with a valid aspect ratio found")
}
