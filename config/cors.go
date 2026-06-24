package config

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var allowedOrigins = []string{
	"http://localhost:3000",
	"http://localhost:3001", // Frontend Next.js berjalan di sini
	"http://localhost:5173",
	"http://localhost:5174",
}

// GetAllowedOrigins mengembalikan daftar origin yang diizinkan untuk mengakses API
func GetAllowedOrigins() []string {
	return allowedOrigins
}

// SetupCORS mengatur middleware CORS untuk Fiber app
func SetupCORS(app *fiber.App) {
	app.Use(cors.New(cors.Config{
		AllowOrigins: strings.Join(GetAllowedOrigins(), ","),
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))
}
