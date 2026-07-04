package router

import (
	"be_silapor/config/middleware"
	"be_silapor/handler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// SetupRoutes mendaftarkan semua endpoint/route API ke aplikasi Fiber
// Dipanggil sekali dari main.go setelah koneksi database berhasil
func SetupRoutes(app *fiber.App) {
	// Endpoint Swagger UI: dokumentasi API yang bisa diakses lewat browser
	// Buka: http://localhost:3000/swagger/index.html
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Semua route API dikelompokkan di bawah prefix /api
	// Contoh: /api/login, /api/laporan, dst.
	api := app.Group("/api")

	// ── Auth Routes (Publik - tidak perlu login) ──────────────────────────────
	api.Post("/register", handler.Register) // Daftar akun baru
	api.Post("/login", handler.Login)       // Login dan dapat token JWT

	// ── Protected Routes (Semua role yang sudah login) ────────────────────────
	// middleware.JWTProtected() tanpa argumen = cukup punya token yang valid
	protected := api.Group("/", middleware.JWTProtected())
	protected.Put("/changepassword", handler.ChangePassword) // Ganti password sendiri

	// ── Kategori Routes ───────────────────────────────────────────────────────
	// GET /api/kategori → semua user yang login bisa melihat daftar kategori
	api.Get("/kategori", middleware.JWTProtected(), handler.GetAllKategori)

	// POST, PUT, DELETE /api/kategori → hanya admin yang bisa mengelola kategori
	kategoriAdmin := api.Group("/kategori", middleware.JWTProtected("admin"))
	kategoriAdmin.Post("/", handler.CreateKategori)    // Tambah kategori baru
	kategoriAdmin.Put("/:id", handler.UpdateKategori)  // Edit kategori (berdasarkan ID)
	kategoriAdmin.Delete("/:id", handler.DeleteKategori) // Hapus kategori (berdasarkan ID)

	// ── Laporan Routes ────────────────────────────────────────────────────────

	// GET /api/laporan → semua role yang login bisa lihat laporan
	// CATATAN: data yang ditampilkan otomatis difilter berdasarkan role:
	//   - mahasiswa → hanya laporan miliknya sendiri
	//   - petugas   → hanya laporan kategori yang ditugaskan kepadanya
	//   - admin     → semua laporan
	api.Get("/laporan", middleware.JWTProtected(), handler.GetAllLaporan)

	// GET /api/laporan/:id → lihat detail satu laporan (semua role)
	api.Get("/laporan/:id", middleware.JWTProtected(), handler.GetLaporanByID)

	// GET /api/laporan/:id/riwayat → lihat timeline perubahan status laporan (semua role)
	api.Get("/laporan/:id/riwayat", middleware.JWTProtected(), handler.GetRiwayatLaporan)

	// POST /api/laporan → hanya mahasiswa yang bisa membuat laporan baru
	api.Post("/laporan", middleware.JWTProtected("mahasiswa"), handler.CreateLaporan)

	// PUT /api/laporan/:id/status → hanya admin dan petugas yang bisa ubah status laporan
	api.Put("/laporan/:id/status", middleware.JWTProtected("admin", "petugas"), handler.UpdateStatusLaporan)

	// PUT /api/laporan/:id/rating → hanya mahasiswa yang bisa memberi rating
	api.Put("/laporan/:id/rating", middleware.JWTProtected("mahasiswa"), handler.UpdateRating)

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
	userAdmin.Delete("/:id", handler.DeleteUser)

	// ── Notifikasi Routes ─────────────────────────────────────────────────────
	notifikasiRoute := api.Group("/notifikasi", middleware.JWTProtected())
	notifikasiRoute.Get("/", handler.GetNotifikasi)
	notifikasiRoute.Put("/read-all", handler.ReadAllNotifikasi)
	notifikasiRoute.Put("/:id/read", handler.ReadNotifikasi)
}
