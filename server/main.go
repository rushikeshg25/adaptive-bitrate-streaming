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

	// Health endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "ok",
			"message": "Server is running",
		})
	})

	// Upload endpoint
	app.Post("/upload", func(c *fiber.Ctx) error {
		// Get file from form
		file, err := c.FormFile("video")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Could not get uploaded file",
			})
		}

		// Generate unique ID and preserve extension
		id := uuid.New().String()
		ext := filepath.Ext(file.Filename)
		newFilename := fmt.Sprintf("%s%s", id, ext)
		savePath := filepath.Join("./videos", newFilename)

		// Save file to disk
		if err := c.SaveFile(file, savePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not save file",
			})
		}

		//TODO:Process the File here
		//FFMPEG

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message":  "File uploaded successfully",
			"id":       id,
			"filename": newFilename,
			"url":      fmt.Sprintf("/videos/%s", newFilename),
		})
	})

	// Start server
	log.Fatal(app.Listen(":3000"))
}
