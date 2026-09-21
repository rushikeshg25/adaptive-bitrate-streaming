package utils

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTranscodeSilentVideo(t *testing.T) {
	if _, e := exec.LookPath("ffmpeg"); e != nil {
		t.Skip("ffmpeg unavailable")
	}
	dir := t.TempDir()
	input := filepath.Join(dir, "input.mp4")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	b, e := exec.CommandContext(ctx, "ffmpeg", "-nostdin", "-y", "-f", "lavfi", "-i", "color=c=blue:s=64x64:d=0.3", "-c:v", "libx264", input).CombinedOutput()
	if e != nil {
		t.Fatal(e, string(b))
	}
	if e = TranscodeToHLSContext(ctx, input, dir); e != nil {
		t.Fatal(e)
	}
	b, e = os.ReadFile(filepath.Join(dir, "master.m3u8"))
	if e != nil || strings.Count(string(b), "#EXT-X-STREAM-INF") != 3 {
		t.Fatal(string(b), e)
	}
}
func TestMetadataRejectsCorruption(t *testing.T) {
	old := metadataFile
	metadataFile = filepath.Join(t.TempDir(), "metadata.json")
	defer func() { metadataFile = old }()
	if e := SaveMetadata(VideoMetadata{ID: "one", Status: "Processing"}); e != nil {
		t.Fatal(e)
	}
	if e := RecoverProcessing(); e != nil {
		t.Fatal(e)
	}
	v, _ := GetAllMetadata()
	if v[0].Status != "Failed" {
		t.Fatal(v)
	}
	os.WriteFile(metadataFile, []byte("broken"), 0600)
	if SaveMetadata(VideoMetadata{ID: "two"}) == nil {
		t.Fatal("corrupt metadata overwritten")
	}
	b, _ := os.ReadFile(metadataFile)
	if string(b) != "broken" {
		t.Fatal(string(b))
	}
}
