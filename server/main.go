package main

import (
	"log"
	"os"
	"server/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Ensure necessary directories exist
	setupDirectories()

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

	// Routes
	app.Get("/api/health", handlers.HealthCheck)
	app.Get("/api/videos", handlers.ListVideos)
	app.Post("/api/upload/chunk", handlers.UploadChunk)

	app.Post("/api/upload/complete", handlers.CompleteUpload)
	app.Post("/api/upload", handlers.LegacyUpload)

	// Start server
	log.Fatal(app.Listen(":3000"))
}

func setupDirectories() {
	dirs := []string{"./videos", "./temp_chunks"}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.Mkdir(dir, 0755); err != nil {
				log.Fatalf("Failed to create directory %s: %v", dir, err)
			}
		}
	}
}
