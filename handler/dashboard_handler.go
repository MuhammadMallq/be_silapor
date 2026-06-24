package handler

import (
	"be_silapor/config"
	"be_silapor/model"

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
	TotalPengguna   int64               `json:"total_pengguna"`
	PriorityReports []model.Laporan     `json:"priority_reports"`
	CategoryStats   []CategoryStat      `json:"category_stats"`
}

func GetAdminDashboard(c *fiber.Ctx) error {
	var resp DashboardResponse

	// Total Laporan
	config.DB.Model(&model.Laporan{}).Count(&resp.TotalLaporan)

	// Belum Selesai
	config.DB.Model(&model.Laporan{}).Where("status != ?", "selesai").Count(&resp.BelumSelesai)

	// Dieskalasi (Prioritas Tinggi)
	config.DB.Model(&model.Laporan{}).Where("prioritas = ?", "tinggi").Count(&resp.Dieskalasi)

	// Total Pengguna
	config.DB.Model(&model.User{}).Count(&resp.TotalPengguna)

	// Priority Reports (Preload Kategori & Pelapor)
	config.DB.Preload("Kategori").Preload("Pelapor").
		Where("prioritas = ?", "tinggi").
		Order("tanggal_lapor desc").
		Limit(5).
		Find(&resp.PriorityReports)

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
