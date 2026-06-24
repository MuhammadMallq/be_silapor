package main

import (
	"be_silapor/config"
	_ "be_silapor/docs"
	"be_silapor/pkg/scheduler"
	"be_silapor/router"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

// @title SiLapor API
// @version 1.0
// @description API untuk Sistem Pengaduan & Tracking Fasilitas Kampus
// @host localhost:3000
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load env variables
	godotenv.Load()

	app := fiber.New()
	app.Use(logger.New())

	// CORS configuration
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Connect database
	config.ConnectDatabase()

	// Setup routes
	router.SetupRoutes(app)

	// Start SLA Escalation Scheduler
	scheduler.StartEscalationScheduler()

	app.Listen(":3000")
}
