package repository

import (
	"be_silapor/config"
	"be_silapor/model"
)

// CreateRiwayat menyimpan satu entri riwayat perubahan status ke database
// Dipanggil setiap kali status laporan berubah, termasuk saat laporan pertama dibuat
// Hasilnya membentuk timeline/log yang bisa dilihat di endpoint GET /laporan/:id/riwayat
func CreateRiwayat(riwayat *model.RiwayatStatus) error {
	return config.DB.Create(riwayat).Error
}

// FindRiwayatByLaporanID mengambil semua riwayat status dari satu laporan tertentu
// Diurutkan dari yang paling lama (waktu ASC) agar tampil seperti timeline kronologis
// Contoh urutan: dilaporkan → ditugaskan → dikerjakan → selesai
func FindRiwayatByLaporanID(laporanID uint) ([]model.RiwayatStatus, error) {
	var riwayats []model.RiwayatStatus
	err := config.DB.Where("laporan_id = ?", laporanID).
		Order("waktu ASC").Find(&riwayats).Error
	return riwayats, err
}
