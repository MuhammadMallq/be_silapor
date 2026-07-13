package main

import (
	"be_silapor/config"
	_ "be_silapor/docs"
	"be_silapor/router"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

// @title SiLapor API
// @version 1.0
// @description ## Sistem Pengaduan & Tracking Fasilitas Kampus
// @description Aplikasi ini mengelola seluruh siklus pengaduan kerusakan fasilitas, mulai dari pelaporan oleh mahasiswa, penugasan dan pengerjaan oleh teknisi (petugas), hingga evaluasi dan pemantauan oleh Admin.
// @description 
// @description **Perhatian Keamanan**: API ini diperuntukkan bagi penggunaan internal aplikasi SiLapor. Beberapa *endpoint* manajemen administratif bersifat tertutup dan sengaja tidak diekspos di dalam dokumentasi publik ini demi keamanan.
// @contact.name Tim Pengembang SiLapor
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Masukkan token JWT Anda dengan format: **Bearer &lt;token&gt;**
func main() {
	// Load env variables (Ignore error jika file .env tidak ada, misalnya saat di Railway)
	godotenv.Load()

	app := fiber.New()

	// Setup Middleware (CORS & Logger)
	config.SetupCORS(app)
	app.Use(logger.New())

	// Connect database
	config.ConnectDatabase()

	// Setup routes
	router.SetupRoutes(app)

	// Start SLA Escalation Scheduler (Dinonaktifkan atas permintaan user - eskalasi hanya manual oleh Admin)
	// scheduler.StartEscalationScheduler()

	// Gunakan PORT dari environment variable, default ke 3000 jika tidak ada
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	app.Listen(":" + port)
}
