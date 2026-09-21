package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"server/utils"
	"strconv"
	"sync"
	"time"
)

var uploadMu sync.Mutex
var validUpload = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)
var processing = make(chan struct{}, 1)
var processingCtx, stopProcessing = context.WithCancel(context.Background())
var workers sync.WaitGroup
var transcodeVideo = utils.TranscodeToHLSContext

func StopProcessing()                { uploadMu.Lock(); stopProcessing(); uploadMu.Unlock(); workers.Wait() }
func HealthCheck(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok"}) }
func ListVideos(c *fiber.Ctx) error {
	v, e := utils.GetAllMetadata()
	if e != nil {
		return fiber.NewError(500, "metadata unavailable")
	}
	return c.JSON(v)
}
func UploadChunk(c *fiber.Ctx) error {
	uploadMu.Lock()
	defer uploadMu.Unlock()
	if processingCtx.Err() != nil {
		return fiber.NewError(503, "shutting down")
	}
	id, index := c.FormValue("uploadId"), c.FormValue("index")
	n, e := strconv.Atoi(index)
	if !validUpload.MatchString(id) || e != nil || n < 0 || n >= 200 || strconv.Itoa(n) != index {
		return fiber.NewError(400, "invalid upload ID or chunk index")
	}
	file, e := c.FormFile("chunk")
	if e != nil || file.Size <= 0 || file.Size > 5<<20 {
		return fiber.NewError(400, "chunk must be 1 byte to 5 MiB")
	}
	dir := filepath.Join("temp_chunks", id)
	if _, e = os.Stat(filepath.Join(dir, "receipt.json")); e == nil {
		return fiber.NewError(409, "upload already completed")
	}
	if e = os.MkdirAll(dir, 0700); e != nil {
		return fiber.NewError(500, "storage unavailable")
	}
	tmp, e := os.CreateTemp(dir, "chunk-*")
	if e != nil {
		return fiber.NewError(500, "storage unavailable")
	}
	name := tmp.Name()
	tmp.Close()
	defer os.Remove(name)
	if e = c.SaveFile(file, name); e != nil {
		return fiber.NewError(500, "chunk write failed")
	}
	if e = os.Rename(name, filepath.Join(dir, index)); e != nil {
		return fiber.NewError(500, "chunk publish failed")
	}
	return c.JSON(fiber.Map{"message": "chunk stored"})
}
func CompleteUpload(c *fiber.Ctx) error {
	uploadMu.Lock()
	defer uploadMu.Unlock()
	if processingCtx.Err() != nil {
		return fiber.NewError(503, "shutting down")
	}
	var p struct {
		UploadID string `json:"uploadId"`
		Filename string `json:"filename"`
		Total    int    `json:"total"`
	}
	if c.BodyParser(&p) != nil || !validUpload.MatchString(p.UploadID) || p.Total < 1 || p.Total > 200 || p.Filename == "" || len(p.Filename) > 255 || filepath.Base(p.Filename) != p.Filename {
		return fiber.NewError(400, "invalid completion payload")
	}
	dir := filepath.Join("temp_chunks", p.UploadID)
	if b, e := os.ReadFile(filepath.Join(dir, "receipt.json")); e == nil {
		c.Set("Content-Type", "application/json")
		return c.Status(201).Send(b)
	}
	for i := 0; i < p.Total; i++ {
		s, e := os.Stat(filepath.Join(dir, strconv.Itoa(i)))
		if e != nil || !s.Mode().IsRegular() || s.Size() <= 0 || s.Size() > 5<<20 {
			return fiber.NewError(400, fmt.Sprintf("invalid or missing chunk %d", i))
		}
	}
	entries, e := os.ReadDir(dir)
	if e != nil || len(entries) != p.Total {
		return fiber.NewError(400, "chunk count mismatch")
	}
	select {
	case processing <- struct{}{}:
	default:
		return fiber.NewError(503, "processor busy; retry completion later")
	}
	release := true
	defer func() {
		if release {
			<-processing
		}
	}()
	id := uuid.NewString()
	videoDir := filepath.Join("videos", id)
	if e = os.MkdirAll(videoDir, 0700); e != nil {
		return fiber.NewError(500, "storage unavailable")
	}
	keep := false
	defer func() {
		if !keep {
			os.RemoveAll(videoDir)
		}
	}()
	input := filepath.Join(videoDir, "input")
	out, e := os.OpenFile(input, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return fiber.NewError(500, "storage unavailable")
	}
	for i := 0; i < p.Total; i++ {
		f, e := os.Open(filepath.Join(dir, strconv.Itoa(i)))
		if e != nil {
			out.Close()
			return fiber.NewError(500, "chunk read failed")
		}
		_, e = io.Copy(out, f)
		f.Close()
		if e != nil {
			out.Close()
			return fiber.NewError(500, "assembly failed")
		}
	}
	if e = out.Close(); e != nil {
		return fiber.NewError(500, "assembly failed")
	}
	metadata := utils.VideoMetadata{ID: id, Filename: p.Filename, Status: "Processing", CreatedAt: time.Now().UTC(), URL: "/videos/" + id + "/master.m3u8"}
	if e = utils.SaveMetadata(metadata); e != nil {
		return fiber.NewError(500, "metadata write failed")
	}
	result := fiber.Map{"videoID": id, "filename": p.Filename, "url": metadata.URL, "message": "processing started"}
	b, _ := json.Marshal(result)
	if e = os.WriteFile(filepath.Join(dir, "receipt.json"), b, 0600); e != nil {
		metadata.Status = "Failed"
		_ = utils.SaveMetadata(metadata)
		return fiber.NewError(500, "receipt write failed")
	}
	keep = true
	release = false
	workers.Add(1)
	go func() {
		defer workers.Done()
		defer func() { <-processing }()
		ctx, cancel := context.WithTimeout(processingCtx, 2*time.Minute)
		defer cancel()
		if e := transcodeVideo(ctx, input, videoDir); e != nil {
			metadata.Status = "Failed"
			log.Printf("transcode %s: %v", id, e)
		} else {
			metadata.Status = "Completed"
		}
		os.Remove(input)
		if e := utils.SaveMetadata(metadata); e != nil {
			log.Printf("persist terminal status %s: %v", id, e)
		}
	}()
	for i := 0; i < p.Total; i++ {
		os.Remove(filepath.Join(dir, strconv.Itoa(i)))
	}
	return c.Status(201).JSON(result)
}
func LegacyUpload(c *fiber.Ctx) error { return fiber.NewError(410, "use chunked upload") }

var mediaPath = regexp.MustCompile(`^/videos/([a-f0-9-]{36})/(master\.m3u8|v[0-2]/(index\.m3u8|segment[0-9]+\.ts))$`)

// ServeVideo exposes only the finished HLS output, never input or metadata files.
func ServeVideo(c *fiber.Ctx) error {
	parts := mediaPath.FindStringSubmatch(c.Path())
	if parts == nil {
		return fiber.ErrNotFound
	}
	videos, e := utils.GetAllMetadata()
	if e != nil {
		return fiber.NewError(500, "metadata unavailable")
	}
	ready := false
	for _, v := range videos {
		if v.ID == parts[1] && v.Status == "Completed" {
			ready = true
			break
		}
	}
	if !ready {
		return fiber.ErrNotFound
	}
	if filepath.Ext(parts[2]) == ".m3u8" {
		c.Set("Content-Type", "application/vnd.apple.mpegurl")
	} else {
		c.Set("Content-Type", "video/mp2t")
	}
	return c.SendFile(filepath.Join("videos", parts[1], parts[2]))
}
