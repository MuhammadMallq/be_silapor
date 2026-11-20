package router

import (
	"be_silapor/config/middleware"
	"be_silapor/handler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// SetupRoutes mendaftarkan semua endpoint/route API ke aplikasi Fiber
func SetupRoutes(app *fiber.App) {
	// Health-check
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "API silapor siap"})
	})

	// Endpoint Swagger UI
	app.Get("/docs/*", swagger.HandlerDefault)

	api := app.Group("/api")

	// ── Auth Routes (Publik) ─────────────────────────────────────────────────
	api.Post("/register", handler.Register)                     // Daftar akun baru
	api.Post("/login", handler.Login)                           // Login dan dapat token JWT
	api.Post("/forgot-password", handler.ForgotPassword)       // Minta OTP lupa password
	api.Post("/verify-otp", handler.VerifyOTP)                 // Verifikasi OTP
	api.Post("/reset-password", handler.ResetPassword)         // Reset kata sandi baru

	// ── Protected Routes (Semua role yang sudah login) ────────────────────────
	protected := api.Group("/", middleware.JWTProtected())
	protected.Put("/changepassword", handler.ChangePassword)   // Ganti password sendiri
	protected.Put("/profile/email", handler.UpdateProfileEmail) // Update email pemulihan akun

	// ── Kategori Routes ───────────────────────────────────────────────────────
	api.Get("/kategori", middleware.JWTProtected(), handler.GetAllKategori)

	kategoriAdmin := api.Group("/kategori", middleware.JWTProtected("admin"))
	kategoriAdmin.Post("/", handler.CreateKategori)
	kategoriAdmin.Put("/:id", handler.UpdateKategori)
	kategoriAdmin.Delete("/:id", handler.DeleteKategori)

	// ── Laporan Routes ────────────────────────────────────────────────────────
	api.Get("/laporan", middleware.JWTProtected(), handler.GetAllLaporan)
	api.Get("/laporan/:id", middleware.JWTProtected(), handler.GetLaporanByID)
	api.Get("/laporan/:id/riwayat", middleware.JWTProtected(), handler.GetRiwayatLaporan)

	// POST /api/laporan → mahasiswa, dosen, staff, dan admin bisa membuat laporan baru
	api.Post("/laporan", middleware.JWTProtected("mahasiswa", "dosen", "staff", "admin"), handler.CreateLaporan)

	// PUT /api/laporan/:id/status → hanya admin dan petugas yang bisa ubah status laporan
	api.Put("/laporan/:id/status", middleware.JWTProtected("admin", "petugas"), handler.UpdateStatusLaporan)

	// PUT /api/laporan/:id/rating → pelapor (mahasiswa, dosen, dan staff) yang bisa memberi rating
	api.Put("/laporan/:id/rating", middleware.JWTProtected("mahasiswa", "dosen", "staff"), handler.UpdateRating)

	// PUT /api/laporan/:id/admin-update → hanya admin yang bisa ubah prioritas dan petugas
	api.Put("/laporan/:id/admin-update", middleware.JWTProtected("admin"), handler.AdminUpdateLaporan)

	// DELETE /api/laporan/:id → hanya admin yang bisa menghapus laporan
	api.Delete("/laporan/:id", middleware.JWTProtected("admin"), handler.DeleteLaporan)

	// ── Admin Dashboard Routes ────────────────────────────────────────────────
	api.Get("/dashboard/admin", middleware.JWTProtected("admin"), handler.GetAdminDashboard)

	// ── Users Routes (Admin Only) ─────────────────────────────────────────────
	userAdmin := api.Group("/users", middleware.JWTProtected("admin"))
	userAdmin.Get("/", handler.GetAllUsers)
	userAdmin.Post("/", handler.CreateUserByAdmin)
	userAdmin.Put("/:id/reset-password", handler.ResetPasswordByAdmin) // Reset password manual oleh admin
	userAdmin.Delete("/:id", handler.DeleteUser)

	// ── Notifikasi Routes ─────────────────────────────────────────────────────
	notifikasiRoute := api.Group("/notifikasi", middleware.JWTProtected())
	notifikasiRoute.Get("/", handler.GetNotifikasi)
	notifikasiRoute.Put("/read-all", handler.ReadAllNotifikasi)
	notifikasiRoute.Put("/:id/read", handler.ReadNotifikasi)
}
