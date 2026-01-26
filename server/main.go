package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
)

func main() {
	// Ensure videos directory exists
	if _, err := os.Stat("./videos"); os.IsNotExist(err) {
		os.Mkdir("./videos", 0755)
	}

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:   "Adaptive Bitrate Streaming API",
		BodyLimit: 100 * 1024 * 1024, // 100MB
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Static files - serve uploaded videos
	app.Static("/videos", "./videos")

	// Ensure temp directory exists
	if _, err := os.Stat("./temp_chunks"); os.IsNotExist(err) {
		os.Mkdir("./temp_chunks", 0755)
	}

	// Health endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "ok",
			"message": "Server is running",
		})
	})

	// Upload Chunk endpoint
	app.Post("/upload/chunk", func(c *fiber.Ctx) error {
		uploadID := c.FormValue("uploadId")
		chunkIndex := c.FormValue("index")

		if uploadID == "" || chunkIndex == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Missing uploadId or index",
			})
		}

		file, err := c.FormFile("chunk")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Could not get chunk",
			})
		}

		// Create directory for this upload
		uploadDir := filepath.Join("./temp_chunks", uploadID)
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.Mkdir(uploadDir, 0755)
		}

		// Save chunk
		chunkPath := filepath.Join(uploadDir, chunkIndex)
		if err := c.SaveFile(file, chunkPath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not save chunk",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Chunk uploaded successfully",
		})
	})

	// Complete Upload endpoint
	app.Post("/upload/complete", func(c *fiber.Ctx) error {
		var payload struct {
			UploadID string `json:"uploadId"`
			Filename string `json:"filename"`
			Total    int    `json:"total"`
		}

		if err := c.BodyParser(&payload); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid payload",
			})
		}

		uploadDir := filepath.Join("./temp_chunks", payload.UploadID)
		videoID := uuid.New().String()
		videoDir := filepath.Join("./videos", videoID)
		os.MkdirAll(videoDir, 0755)

		ext := filepath.Ext(payload.Filename)
		tempMP4Path := filepath.Join(videoDir, "input"+ext)

		// Create final file
		finalFile, err := os.Create(tempMP4Path)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not create temp mp4 file",
			})
		}

		// Merge chunks in order
		for i := 0; i < payload.Total; i++ {
			chunkPath := filepath.Join(uploadDir, fmt.Sprintf("%d", i))
			chunkBytes, err := os.ReadFile(chunkPath)
			if err != nil {
				finalFile.Close()
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fmt.Sprintf("Missing chunk %d", i),
				})
			}

			if _, err := finalFile.Write(chunkBytes); err != nil {
				finalFile.Close()
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to write chunk to final file",
				})
			}
		}
		finalFile.Close()

		// Transcode to HLS
		if err := transcodeToHLS(tempMP4Path, videoDir); err != nil {
			log.Printf("Transcoding failed: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Transcoding failed",
			})
		}

		// Optional: Cleanup the original MP4 if you only want to serve HLS
		// os.Remove(tempMP4Path)

		// Cleanup chunks
		os.RemoveAll(uploadDir)

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message":  "Video processed successfully",
			"videoID":  videoID,
			"filename": payload.Filename,
			"url":      fmt.Sprintf("/videos/%s/master.m3u8", videoID),
		})
	})

	// Legacy Upload endpoint (for backward compatibility if needed, or remove)
	app.Post("/upload", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusGone).JSON(fiber.Map{
			"error": "Use chunked upload instead",
		})
	})

	// Start server
	log.Fatal(app.Listen(":3000"))
}

func transcodeToHLS(inputPath string, outputDir string) error {
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
