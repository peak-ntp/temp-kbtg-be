package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Create a new Fiber instance
	app := fiber.New(fiber.Config{
		AppName: "KBTG Backend API v1.0.0",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Routes
	setupRoutes(app)

	// Start server on port 3000
	log.Fatal(app.Listen(":3000"))
}

func setupRoutes(app *fiber.App) {
	// API v1 group
	api := app.Group("/api/v1")

	// Hello World endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello World!",
			"status":  "success",
			"data":    "Welcome to KBTG Backend API",
		})
	})

	// Health check endpoint
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Service is running",
		})
	})

	// Additional Hello endpoint under API group
	api.Get("/hello", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello from API v1!",
			"status":  "success",
		})
	})
}
