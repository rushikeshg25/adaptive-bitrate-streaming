package handlers

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"server/utils"

	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// HealthCheck handles the health check endpoint
func HealthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"message": "Server is running",
	})
}

// ListVideos returns all uploaded videos metadata
func ListVideos(c *fiber.Ctx) error {
	videos, err := utils.GetAllMetadata()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not fetch videos",
		})
	}
	return c.JSON(videos)
}

// UploadChunk handles uploading individual video chunks
func UploadChunk(c *fiber.Ctx) error {
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
}

// CompleteUpload handles merging chunks and triggering transcoding
func CompleteUpload(c *fiber.Ctx) error {
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

	// Create metadata record
	metadata := utils.VideoMetadata{
		ID:        videoID,
		Filename:  payload.Filename,
		Status:    "Processing",
		CreatedAt: time.Now(),
		URL:       fmt.Sprintf("/videos/%s/master.m3u8", videoID),
	}
	utils.SaveMetadata(metadata)

	// Transcode to HLS in background
	go func() {
		log.Printf("Starting background transcoding for video %s", videoID)
		if err := utils.TranscodeToHLS(tempMP4Path, videoDir); err != nil {
			log.Printf("Transcoding failed for video %s: %v", videoID, err)
			metadata.Status = "Failed"
			utils.SaveMetadata(metadata)
			return
		}
		log.Printf("Successfully transcoded video %s", videoID)
		metadata.Status = "Completed"
		utils.SaveMetadata(metadata)
	}()

	// Cleanup chunks
	os.RemoveAll(uploadDir)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Video upload success, transcoding started",
		"videoID":  videoID,
		"filename": payload.Filename,
		"url":      metadata.URL,
	})
}

// LegacyUpload handles the deprecated single-request upload
func LegacyUpload(c *fiber.Ctx) error {
	return c.Status(fiber.StatusGone).JSON(fiber.Map{
		"error": "Use chunked upload instead",
	})
}
