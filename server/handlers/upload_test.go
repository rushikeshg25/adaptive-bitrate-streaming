package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
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
