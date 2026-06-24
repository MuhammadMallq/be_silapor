package router

import (
	"be_silapor/config/middleware"
	"be_silapor/handler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func SetupRoutes(app *fiber.App) {
	// Swagger
	app.Get("/swagger/*", swagger.HandlerDefault)

	api := app.Group("/api")

	// Auth routes
	api.Post("/register", handler.Register)
	api.Post("/login", handler.Login)

	// Protected routes
	protected := api.Group("/", middleware.JWTProtected())

	// Profile / Auth
	protected.Put("/changepassword", handler.ChangePassword)

	// Kategori (Admin Only for CUD, All authenticated for Read)
	api.Get("/kategori", middleware.JWTProtected(), handler.GetAllKategori)
	kategoriAdmin := api.Group("/kategori", middleware.JWTProtected("admin"))
	kategoriAdmin.Post("/", handler.CreateKategori)
	kategoriAdmin.Put("/:id", handler.UpdateKategori)
	kategoriAdmin.Delete("/:id", handler.DeleteKategori)

	// Laporan
	api.Get("/laporan", middleware.JWTProtected(), handler.GetAllLaporan)
	api.Get("/laporan/:id", middleware.JWTProtected(), handler.GetLaporanByID)
	api.Get("/laporan/:id/riwayat", middleware.JWTProtected(), handler.GetRiwayatLaporan)

	// Laporan - Mahasiswa (Create)
	api.Post("/laporan", middleware.JWTProtected("mahasiswa"), handler.CreateLaporan)

	// Laporan - Petugas & Admin (Update Status)
	api.Put("/laporan/:id/status", middleware.JWTProtected("admin", "petugas"), handler.UpdateStatusLaporan)

	// Laporan - Admin (Delete)
	api.Delete("/laporan/:id", middleware.JWTProtected("admin"), handler.DeleteLaporan)
}
