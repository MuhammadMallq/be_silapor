package handler

import (
	"be_silapor/config"
	"be_silapor/model"
	"time"

	"github.com/gofiber/fiber/v2"
)

type CategoryStat struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
	Fill  string `json:"fill"`
}

type DashboardResponse struct {
	TotalLaporan    int64               `json:"total_laporan"`
	BelumSelesai    int64               `json:"belum_selesai"`
	Dieskalasi      int64               `json:"dieskalasi"`
	Terlambat       int64               `json:"terlambat"`
	TotalPengguna   int64               `json:"total_pengguna"`
	PriorityReports []model.Laporan     `json:"priority_reports"`
	LateReports     []model.Laporan     `json:"late_reports"`
	CategoryStats   []CategoryStat      `json:"category_stats"`
}

// GetAdminDashboard godoc
// @Summary Statistik Dashboard Admin
// @Description Mengambil rekapitulasi data statistik laporan (total, belum selesai, terlambat, eskalasi prioritas tinggi) dan statistik per kategori
// @Tags Dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.Response{data=DashboardResponse}
// @Router /api/dashboard/admin [get]
func GetAdminDashboard(c *fiber.Ctx) error {
	var resp DashboardResponse

	// Total Laporan
	config.DB.Model(&model.Laporan{}).Count(&resp.TotalLaporan)

	// Belum Selesai
	config.DB.Model(&model.Laporan{}).Where("status != ?", "selesai").Count(&resp.BelumSelesai)

	// Dieskalasi (Prioritas Tinggi) - hanya yang belum selesai
	config.DB.Model(&model.Laporan{}).Where("prioritas = ?", "tinggi").Where("status != ?", "selesai").Count(&resp.Dieskalasi)

	// Total Pengguna
	config.DB.Model(&model.User{}).Count(&resp.TotalPengguna)

	// Priority Reports (Preload Kategori & Pelapor) - hanya yang belum selesai
	config.DB.Preload("Kategori").Preload("Pelapor").
		Where("prioritas = ?", "tinggi").
		Where("status != ?", "selesai").
		Order("tanggal_lapor desc").
		Limit(5).
		Find(&resp.PriorityReports)

	// Filter Laporan Terlambat (Belum Selesai & Lewat SLA / Tenggat Waktu)
	var activeLaporans []model.Laporan
	config.DB.Preload("Kategori").Preload("Pelapor").
		Where("status != ?", "selesai").
		Order("tanggal_lapor asc"). // yang paling lama dibuat / paling lama telat
		Find(&activeLaporans)

	now := time.Now()
	for _, lap := range activeLaporans {
		var deadline time.Time
		if lap.TenggatWaktu != nil {
			deadline = *lap.TenggatWaktu
		} else {
			slaHours := lap.Kategori.SLAJam
			if slaHours == 0 {
				slaHours = 48
			}
			deadline = lap.TanggalLapor.Add(time.Duration(slaHours) * time.Hour)
		}

		if now.After(deadline) {
			resp.Terlambat++
			if len(resp.LateReports) < 5 {
				resp.LateReports = append(resp.LateReports, lap)
			}
		}
	}

	// Category Stats
	var kats []model.KategoriFasilitas
	config.DB.Find(&kats)

	colors := []string{"#3b82f6", "#10b981", "#f59e0b", "#8b5cf6", "#ef4444", "#06b6d4"}

	for i, k := range kats {
		var count int64
		config.DB.Model(&model.Laporan{}).Where("kategori_id = ?", k.ID).Count(&count)
		color := colors[i%len(colors)]
		resp.CategoryStats = append(resp.CategoryStats, CategoryStat{
			Name:  k.NamaKategori,
			Value: count,
			Fill:  color,
		})
	}

	return c.JSON(model.Response{
		Message: "Berhasil mengambil statistik dashboard",
		Data:    resp,
	})
}
