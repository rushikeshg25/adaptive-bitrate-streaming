package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"server/utils"
	"strings"
	"testing"
)

func TestRejectUnsafeAndMissingChunks(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(old)
	app := fiber.New()
	app.Post("/chunk", UploadChunk)
	app.Post("/complete", CompleteUpload)
	var body bytes.Buffer
	m := multipart.NewWriter(&body)
	m.WriteField("uploadId", "../escape")
	m.WriteField("index", "0")
	f, _ := m.CreateFormFile("chunk", "test")
	f.Write([]byte("x"))
	m.Close()
	r := httptest.NewRequest("POST", "/chunk", &body)
	r.Header.Set("Content-Type", m.FormDataContentType())
	resp, e := app.Test(r)
	if e != nil {
		t.Fatal(e)
	}
	resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Fatal(resp.StatusCode)
	}
	for _, payload := range []map[string]any{{"uploadId": "good", "filename": "x.mp4", "total": 0}, {"uploadId": "good", "filename": "x.mp4", "total": 1}} {
		b, _ := json.Marshal(payload)
		r = httptest.NewRequest("POST", "/complete", bytes.NewReader(b))
		r.Header.Set("Content-Type", "application/json")
		resp, e = app.Test(r)
		if e != nil {
			t.Fatal(e)
		}
		b, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 400 {
			t.Fatal(resp.StatusCode, string(b))
		}
	}
}

func TestCompleteAssemblyAndReceipt(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(old)
	prior := transcodeVideo
	transcodeVideo = func(ctx context.Context, input, output string) error {
		b, e := os.ReadFile(input)
		if e != nil {
			return e
		}
		if string(b) != "AB" {
			return fmt.Errorf("assembled %q", b)
		}
		return os.WriteFile(filepath.Join(output, "master.m3u8"), []byte("#EXTM3U"), 0600)
	}
	defer func() { workers.Wait(); transcodeVideo = prior }()
	os.MkdirAll("temp_chunks/test", 0700)
	os.WriteFile("temp_chunks/test/0", []byte("A"), 0600)
	os.WriteFile("temp_chunks/test/1", []byte("B"), 0600)
	app := fiber.New()
	app.Post("/complete", CompleteUpload)
	var id string
	for i := 0; i < 2; i++ {
		r := httptest.NewRequest("POST", "/complete", strings.NewReader(`{"uploadId":"test","filename":"clip.mp4","total":2}`))
		r.Header.Set("Content-Type", "application/json")
		resp, e := app.Test(r)
		if e != nil {
			t.Fatal(e)
		}
		var result map[string]string
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		if resp.StatusCode != 201 {
			t.Fatal(resp.StatusCode, result)
		}
		if i == 0 {
			id = result["videoID"]
		} else if id != result["videoID"] {
			t.Fatal("completion not idempotent")
		}
	}
	workers.Wait()
	videos, e := utils.GetAllMetadata()
	if e != nil || len(videos) != 1 || videos[0].Status != "Completed" {
		t.Fatal(videos, e)
	}
}
