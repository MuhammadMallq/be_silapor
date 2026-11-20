package repository

import (
	"be_silapor/config"
	"be_silapor/model"
)

// LaporanFilter menampung parameter filter laporan
type LaporanFilter struct {
	Status      string
	Riwayat     *bool // nil = tanpa filter; true = hanya selesai & ditolak; false = hanya laporan aktif
	RolePelapor string
	KategoriID  uint
	Prioritas   string
	Terlambat   bool
	PelaporID   uint // > 0 jika khusus pelapor tertentu
	PetugasID   uint // > 0 jika khusus petugas tertentu
}

// CreateLaporan menyimpan laporan baru ke tabel "laporan"
func CreateLaporan(laporan *model.Laporan) error {
	return config.DB.Create(laporan).Error
}

// FindAllLaporan mengambil SEMUA laporan dari database
func FindAllLaporan() ([]model.Laporan, error) {
	var laporans []model.Laporan
	err := config.DB.Preload("Pelapor").Preload("Kategori").
		Order("tanggal_lapor DESC").Find(&laporans).Error
	return laporans, err
}

// FindLaporanWithFilter mengambil laporan dengan filter dinamis di backend
func FindLaporanWithFilter(filter LaporanFilter) ([]model.Laporan, error) {
	var laporans []model.Laporan
	query := config.DB.Model(&model.Laporan{}).
		Select("laporan.*").
		Preload("Pelapor").
		Preload("Kategori").
		Preload("Petugas")

	if filter.PelaporID > 0 {
		query = query.Where("laporan.pelapor_id = ?", filter.PelaporID)
	}

	if filter.PetugasID > 0 {
		query = query.Joins("LEFT JOIN kategori_fasilitas AS k_pet ON k_pet.id = laporan.kategori_id").
			Where("laporan.petugas_id = ? OR (laporan.petugas_id IS NULL AND k_pet.petugas_id = ?)", filter.PetugasID, filter.PetugasID)
	}

	if filter.RolePelapor != "" && filter.RolePelapor != "semua" {
		query = query.Joins("JOIN users AS u_pelapor ON u_pelapor.id = laporan.pelapor_id").
			Where("u_pelapor.role = ?", filter.RolePelapor)
	}

	if filter.KategoriID > 0 {
		query = query.Where("laporan.kategori_id = ?", filter.KategoriID)
	}

	if filter.Prioritas != "" && filter.Prioritas != "semua" {
		query = query.Where("laporan.prioritas = ?", filter.Prioritas)
	}

	if filter.Status != "" && filter.Status != "semua" {
		if filter.Status == "aktif" {
			query = query.Where("laporan.status NOT IN (?)", []string{"selesai", "ditolak"})
		} else {
			query = query.Where("laporan.status = ?", filter.Status)
		}
	}

	if filter.Riwayat != nil {
		if *filter.Riwayat {
			query = query.Where("laporan.status IN (?)", []string{"selesai", "ditolak"})
		} else {
			query = query.Where("laporan.status NOT IN (?)", []string{"selesai", "ditolak"})
		}
	}

	if filter.Terlambat {
		query = query.Joins("JOIN kategori_fasilitas AS k_sla ON k_sla.id = laporan.kategori_id").
			Where("laporan.status NOT IN (?)", []string{"selesai", "ditolak"}).
			Where("(laporan.tenggat_waktu IS NOT NULL AND laporan.tenggat_waktu < NOW()) OR (laporan.tenggat_waktu IS NULL AND (laporan.tanggal_lapor + (COALESCE(k_sla.sla_jam, 48) * INTERVAL '1 hour')) < NOW())")
	}

	err := query.Order("laporan.tanggal_lapor DESC").Find(&laporans).Error
	return laporans, err
}

// FindLaporanByPelaporID mengambil laporan milik satu mahasiswa/dosen/staff
func FindLaporanByPelaporID(pelaporID uint) ([]model.Laporan, error) {
	var laporans []model.Laporan
	err := config.DB.Preload("Pelapor").Preload("Kategori").
		Where("pelapor_id = ?", pelaporID).
		Order("tanggal_lapor DESC").Find(&laporans).Error
	return laporans, err
}

// FindLaporanByKategoriPetugasID mengambil laporan yang kategorinya ditugaskan ke petugas tertentu
func FindLaporanByKategoriPetugasID(petugasID uint) ([]model.Laporan, error) {
	var laporans []model.Laporan
	err := config.DB.Preload("Pelapor").Preload("Kategori").Preload("Petugas").
		Joins("JOIN kategori_fasilitas ON kategori_fasilitas.id = laporan.kategori_id").
		Where("laporan.petugas_id = ? OR (laporan.petugas_id IS NULL AND kategori_fasilitas.petugas_id = ?)", petugasID, petugasID).
		Order("laporan.tanggal_lapor DESC").Find(&laporans).Error
	return laporans, err
}

// FindLaporanByID mengambil satu laporan berdasarkan ID-nya
func FindLaporanByID(id uint) (*model.Laporan, error) {
	var laporan model.Laporan
	err := config.DB.Preload("Pelapor").Preload("Kategori").Preload("Kategori.Petugas").Preload("Petugas").
		First(&laporan, id).Error
	if err != nil {
		return nil, err
	}
	return &laporan, nil
}

// UpdateLaporan menyimpan perubahan data laporan ke database
func UpdateLaporan(laporan *model.Laporan) error {
	return config.DB.Save(laporan).Error
}

// DeleteLaporan menghapus laporan dari database berdasarkan ID
func DeleteLaporan(id uint) error {
	return config.DB.Delete(&model.Laporan{}, id).Error
}

// FindPendingForEscalation mencari laporan yang BELUM selesai dan MASIH prioritas "normal"
func FindPendingForEscalation() ([]model.Laporan, error) {
	var laporans []model.Laporan
	err := config.DB.Preload("Kategori").
		Where("status IN ? AND prioritas = ?", []string{"ditinjau", "diterima", "dilaporkan", "ditugaskan"}, "normal").
		Find(&laporans).Error
	return laporans, err
}
