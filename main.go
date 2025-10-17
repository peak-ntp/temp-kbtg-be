package main

import (
	"log"
	"time"

	"kbtg-backend/internal/database"
	"kbtg-backend/internal/handlers"
	"kbtg-backend/internal/repositories"
	"kbtg-backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Initialize database
	db, err := database.NewConnection("./kbtg.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.DB.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Seed database with sample data
	seedDatabase(db)

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db.DB)
	transferRepo := repositories.NewTransferRepository(db.DB)

	// Initialize services
	userService := services.NewUserService(userRepo)
	transferService := services.NewTransferService(transferRepo)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	transferHandler := handlers.NewTransferHandler(transferService)

	// Create a new Fiber instance
	app := fiber.New(fiber.Config{
		AppName: "KBTG Backend API v1.0.0",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Serve static files (for Swagger UI)
	app.Static("/", "./public")
	app.Static("/swagger.yml", "./swagger.yml")

	// Routes
	setupRoutes(app, userHandler, transferHandler)

	// Start server on port 3000
	log.Fatal(app.Listen(":3000"))
}

func setupRoutes(app *fiber.App, userHandler *handlers.UserHandler, transferHandler *handlers.TransferHandler) {
	// API v1 group
	api := app.Group("/api/v1")

	// API Root endpoint
	api.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to KBTG Backend API",
			"status":  "success",
			"version": "1.0.0",
			"endpoints": fiber.Map{
				"swagger":   "/",
				"health":    "/api/v1/health",
				"users":     "/api/v1/users",
				"transfers": "/api/v1/transfers",
			},
		})
	})

	// Health check endpoint
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Service is running",
		})
	})

	// User CRUD endpoints
	users := api.Group("/users")
	users.Get("/", userHandler.GetUsers)         // GET /api/v1/users
	users.Get("/:id", userHandler.GetUser)       // GET /api/v1/users/:id
	users.Post("/", userHandler.CreateUser)      // POST /api/v1/users
	users.Put("/:id", userHandler.UpdateUser)    // PUT /api/v1/users/:id
	users.Delete("/:id", userHandler.DeleteUser) // DELETE /api/v1/users/:id

	// Transfer endpoints
	transfers := api.Group("/transfers")
	transfers.Post("/", transferHandler.CreateTransfer) // POST /api/v1/transfers
	transfers.Get("/", transferHandler.GetTransfers)    // GET /api/v1/transfers?userId=X
	transfers.Get("/:id", transferHandler.GetTransfer)  // GET /api/v1/transfers/:id
}

func seedDatabase(db *database.DB) {
	// Check if users already exist
	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Printf("Failed to check existing users: %v", err)
		return
	}

	if count > 0 {
		log.Println("Database already has users, skipping seed")
		return
	}

	// Insert sample user based on the image provided (สมชาย ใจดี)
	membershipDate, _ := time.Parse("2006-01-02", "2023-06-15") // 15/6/2566
	now := time.Now()

	query := `
		INSERT INTO users (member_id, first_name, last_name, phone, email, 
		                  membership_date, membership_level, points, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = db.DB.Exec(query,
		"LBK001234",
		"สมชาย",
		"ใจดี",
		"081-234-5678",
		"somchai@example.com",
		membershipDate,
		"Gold",
		15420,
		now,
		now,
	)
	if err != nil {
		log.Printf("Failed to seed user: %v", err)
		return
	}

	// Add more sample users
	users := []map[string]interface{}{
		{
			"member_id":        "LBK001235",
			"first_name":       "สมหญิง",
			"last_name":        "รักดี",
			"phone":            "082-345-6789",
			"email":            "somying@example.com",
			"membership_level": "Silver",
			"points":           8750,
		},
		{
			"member_id":        "LBK001236",
			"first_name":       "วิชัย",
			"last_name":        "ยิ้มแย้ม",
			"phone":            "083-456-7890",
			"email":            "wichai@example.com",
			"membership_level": "Bronze",
			"points":           3200,
		},
	}

	for _, user := range users {
		_, err = db.DB.Exec(query,
			user["member_id"],
			user["first_name"],
			user["last_name"],
			user["phone"],
			user["email"],
			membershipDate,
			user["membership_level"],
			user["points"],
			now,
			now,
		)
		if err != nil {
			log.Printf("Failed to seed user %s: %v", user["member_id"], err)
		}
	}

	log.Println("✅ Database seeded successfully with sample users")
}
