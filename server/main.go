package main

import (
	"fmt"
	"log"
	"os"
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
		AppName: "Adaptive Bitrate Streaming API",
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
		ext := filepath.Ext(payload.Filename)
		newFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
		finalPath := filepath.Join("./videos", newFilename)

		// Create final file
		finalFile, err := os.Create(finalPath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not create final file",
			})
		}
		defer finalFile.Close()

		// Merge chunks in order
		for i := 0; i < payload.Total; i++ {
			chunkPath := filepath.Join(uploadDir, fmt.Sprintf("%d", i))
			chunkBytes, err := os.ReadFile(chunkPath)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fmt.Sprintf("Missing chunk %d", i),
				})
			}

			if _, err := finalFile.Write(chunkBytes); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to write chunk to final file",
				})
			}
		}

		// Cleanup
		os.RemoveAll(uploadDir)

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message":  "File reassembled successfully",
			"filename": newFilename,
			"url":      fmt.Sprintf("/videos/%s", newFilename),
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
