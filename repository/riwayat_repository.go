package repository

import (
	"be_silapor/config"
	"be_silapor/model"
)

func CreateRiwayat(riwayat *model.RiwayatStatus) error {
	return config.DB.Create(riwayat).Error
}

func FindRiwayatByLaporanID(laporanID uint) ([]model.RiwayatStatus, error) {
	var riwayats []model.RiwayatStatus
	err := config.DB.Where("laporan_id = ?", laporanID).
		Order("waktu ASC").Find(&riwayats).Error
	return riwayats, err
}
