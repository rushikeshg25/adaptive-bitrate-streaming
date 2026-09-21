package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func TranscodeToHLS(input, dir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	return TranscodeToHLSContext(ctx, input, dir)
}
func TranscodeToHLSContext(ctx context.Context, input, dir string) error {
	for i := 0; i < 3; i++ {
		if e := os.MkdirAll(filepath.Join(dir, fmt.Sprintf("v%d", i)), 0700); e != nil {
			return e
		}
	}
	probe, e := exec.CommandContext(ctx, "ffprobe", "-v", "error", "-protocol_whitelist", "file,pipe", "-format_whitelist", "mov,matroska,webm,avi,mpegts", "-select_streams", "a:0", "-show_entries", "stream=index", "-of", "csv=p=0", input).Output()
	if e != nil {
		return fmt.Errorf("probe input: %w", e)
	}
	audio := len(strings.TrimSpace(string(probe))) > 0
	args := []string{"-nostdin", "-y", "-protocol_whitelist", "file,pipe", "-format_whitelist", "mov,matroska,webm,avi,mpegts", "-i", input}
	variants := []string{}
	for i := 0; i < 3; i++ {
		args = append(args, "-map", "0:v:0")
		variant := fmt.Sprintf("v:%d", i)
		if audio {
			args = append(args, "-map", "0:a:0")
			variant += fmt.Sprintf(",a:%d", i)
		}
		variants = append(variants, variant)
	}
	args = append(args, "-c:v", "libx264", "-preset", "ultrafast", "-pix_fmt", "yuv420p", "-c:a", "aac", "-ar", "48000", "-b:a", "128k", "-g", "48", "-sc_threshold", "0", "-filter:v:0", "scale=640:360", "-b:v:0", "800k", "-maxrate:v:0", "800k", "-bufsize:v:0", "1200k", "-filter:v:1", "scale=854:480", "-b:v:1", "1400k", "-maxrate:v:1", "1400k", "-bufsize:v:1", "2100k", "-filter:v:2", "scale=1280:720", "-b:v:2", "2800k", "-maxrate:v:2", "2800k", "-bufsize:v:2", "4200k", "-var_stream_map", strings.Join(variants, " "), "-f", "hls", "-hls_time", "2", "-hls_playlist_type", "vod", "-hls_segment_filename", filepath.Join(dir, "v%v/segment%03d.ts"), filepath.Join(dir, "v%v/index.m3u8"))
	if b, e := exec.CommandContext(ctx, "ffmpeg", args...).CombinedOutput(); e != nil {
		return fmt.Errorf("ffmpeg: %w: %.2000s", e, b)
	}
	// Some FFmpeg builds omit master variants for short silent clips even when
	// encoding succeeds. Publish our explicit ladder only after validating media.
	for i := 0; i < 3; i++ {
		p := filepath.Join(dir, fmt.Sprintf("v%d/index.m3u8", i))
		b, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		if !strings.Contains(string(b), "#EXT-X-ENDLIST") {
			return fmt.Errorf("variant %d incomplete", i)
		}
		found := false
		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if filepath.Base(line) != line || filepath.Ext(line) != ".ts" {
				return fmt.Errorf("unexpected segment path")
			}
			stat, e := os.Stat(filepath.Join(filepath.Dir(p), line))
			if e != nil || stat.Size() == 0 {
				return fmt.Errorf("missing segment in variant %d", i)
			}
			found = true
		}
		if !found {
			return fmt.Errorf("variant %d has no media", i)
		}
	}
	master := "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-STREAM-INF:BANDWIDTH=1200000,RESOLUTION=640x360\nv0/index.m3u8\n#EXT-X-STREAM-INF:BANDWIDTH=2000000,RESOLUTION=854x480\nv1/index.m3u8\n#EXT-X-STREAM-INF:BANDWIDTH=4000000,RESOLUTION=1280x720\nv2/index.m3u8\n"
	temp := filepath.Join(dir, "master.tmp")
	if e := os.WriteFile(temp, []byte(master), 0600); e != nil {
		return e
	}
	return os.Rename(temp, filepath.Join(dir, "master.m3u8"))
}
