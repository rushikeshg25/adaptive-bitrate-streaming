package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func TranscodeToHLS(inputPath string, outputDir string) error {
	// Ensure variant directories exist for FFmpeg
	for _, v := range []string{"v0", "v1", "v2"} {
		os.MkdirAll(filepath.Join(outputDir, v), 0755)
	}

	// Create original FFmpeg command for ABR HLS

	// We generate 3 qualities: 360p, 480p, 720p
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		// Map streams for each variant
		"-map", "0:v:0", "-map", "0:a:0",
		"-map", "0:v:0", "-map", "0:a:0",
		"-map", "0:v:0", "-map", "0:a:0",
		// Video and Audio codec
		"-c:v", "libx264", "-crf", "22", "-c:a", "aac", "-ar", "48000",
		// 360p variant
		"-filter:v:0", "scale=w=640:h=360", "-maxrate:v:0", "800k", "-bufsize:v:0", "1200k",
		// 480p variant
		"-filter:v:1", "scale=w=854:h=480", "-maxrate:v:1", "1400k", "-bufsize:v:1", "2100k",
		// 720p variant
		"-filter:v:2", "scale=w=1280:h=720", "-maxrate:v:2", "2800k", "-bufsize:v:2", "4200k",
		// Map descriptors to variant streams
		"-var_stream_map", "v:0,a:0 v:1,a:1 v:2,a:2",
		// HLS settings
		"-master_pl_name", "master.m3u8",
		"-f", "hls",
		"-hls_time", "10",
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", filepath.Join(outputDir, "v%v/segment%03d.ts"),
		filepath.Join(outputDir, "v%v/index.m3u8"),
	)

	// Run command and capture output for debugging
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %v, output: %s", err, string(output))
	}

	return nil
}
