package repository

import (
	"be_silapor/config"
	"be_silapor/model"
)

func CreateLaporan(laporan *model.Laporan) error {
	return config.DB.Create(laporan).Error
}

func FindAllLaporan() ([]model.Laporan, error) {
	var laporans []model.Laporan
	err := config.DB.Preload("Pelapor").Preload("Kategori").
		Order("tanggal_lapor DESC").Find(&laporans).Error
	return laporans, err
}

func FindLaporanByPelaporID(pelaporID uint) ([]model.Laporan, error) {
	var laporans []model.Laporan
	err := config.DB.Preload("Pelapor").Preload("Kategori").
		Where("pelapor_id = ?", pelaporID).
		Order("tanggal_lapor DESC").Find(&laporans).Error
	return laporans, err
}

func FindLaporanByKategoriPetugasID(petugasID uint) ([]model.Laporan, error) {
	var laporans []model.Laporan
	err := config.DB.Preload("Pelapor").Preload("Kategori").
		Joins("JOIN kategori_fasilitas ON kategori_fasilitas.id = laporan.kategori_id").
		Where("kategori_fasilitas.petugas_id = ?", petugasID).
		Order("laporan.tanggal_lapor DESC").Find(&laporans).Error
	return laporans, err
}

func FindLaporanByID(id uint) (*model.Laporan, error) {
	var laporan model.Laporan
	err := config.DB.Preload("Pelapor").Preload("Kategori").Preload("Kategori.Petugas").
		First(&laporan, id).Error
	if err != nil {
		return nil, err
	}
	return &laporan, nil
}

func UpdateLaporan(laporan *model.Laporan) error {
	return config.DB.Save(laporan).Error
}

func DeleteLaporan(id uint) error {
	return config.DB.Delete(&model.Laporan{}, id).Error
}

// FindPendingForEscalation finds reports that might need SLA escalation
func FindPendingForEscalation() ([]model.Laporan, error) {
	var laporans []model.Laporan
	err := config.DB.Preload("Kategori").
		Where("status IN ? AND prioritas = ?", []string{"dilaporkan", "ditugaskan"}, "normal").
		Find(&laporans).Error
	return laporans, err
}
