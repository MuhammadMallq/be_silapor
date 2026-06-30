package repository

import (
	"be_silapor/config"
	"be_silapor/model"
)

// CreateLaporan menyimpan laporan baru ke tabel "laporan"
// Dipanggil saat mahasiswa berhasil mengirim laporan kerusakan
func CreateLaporan(laporan *model.Laporan) error {
	return config.DB.Create(laporan).Error
}

// FindAllLaporan mengambil SEMUA laporan dari database
// Preload("Pelapor") → ikut mengambil data user yang melapor
// Preload("Kategori") → ikut mengambil data kategori fasilitas
// Diurut dari yang terbaru (tanggal_lapor DESC)
// Dipanggil oleh admin yang bisa melihat semua laporan
func FindAllLaporan() ([]model.Laporan, error) {
	var laporans []model.Laporan
	err := config.DB.Preload("Pelapor").Preload("Kategori").
		Order("tanggal_lapor DESC").Find(&laporans).Error
	return laporans, err
}

// FindLaporanByPelaporID mengambil laporan milik satu mahasiswa saja
// Dipakai untuk role "mahasiswa" agar hanya bisa melihat laporannya sendiri
// pelaporID diambil dari token JWT (c.Locals("user_id"))
func FindLaporanByPelaporID(pelaporID uint) ([]model.Laporan, error) {
	var laporans []model.Laporan
	err := config.DB.Preload("Pelapor").Preload("Kategori").
		Where("pelapor_id = ?", pelaporID).
		Order("tanggal_lapor DESC").Find(&laporans).Error
	return laporans, err
}

// FindLaporanByKategoriPetugasID mengambil laporan yang kategorinya ditugaskan ke petugas tertentu
// Menggunakan JOIN ke tabel kategori_fasilitas untuk mencari laporan
// yang kategorinya memiliki petugas_id yang cocok
// Dipakai untuk role "petugas" agar hanya melihat laporan yang menjadi tanggung jawabnya
func FindLaporanByKategoriPetugasID(petugasID uint) ([]model.Laporan, error) {
	var laporans []model.Laporan
	err := config.DB.Preload("Pelapor").Preload("Kategori").Preload("Petugas").
		Joins("JOIN kategori_fasilitas ON kategori_fasilitas.id = laporan.kategori_id").
		Where("laporan.petugas_id = ? OR (laporan.petugas_id IS NULL AND kategori_fasilitas.petugas_id = ?)", petugasID, petugasID).
		Order("laporan.tanggal_lapor DESC").Find(&laporans).Error
	return laporans, err
}

// FindLaporanByID mengambil satu laporan berdasarkan ID-nya
// Preload bertingkat: Kategori → Petugas (untuk tahu siapa petugas yang bertanggung jawab)
// Mengembalikan error jika laporan tidak ditemukan (dipakai untuk validasi)
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
// Dipakai saat status laporan diubah atau prioritas dinaikan oleh scheduler
func UpdateLaporan(laporan *model.Laporan) error {
	return config.DB.Save(laporan).Error
}

// DeleteLaporan menghapus laporan dari database berdasarkan ID
// Hanya admin yang bisa memanggil fungsi ini (diatur di router)
func DeleteLaporan(id uint) error {
	return config.DB.Delete(&model.Laporan{}, id).Error
}

// FindPendingForEscalation mencari laporan yang BELUM selesai dan MASIH prioritas "normal"
// Dipanggil oleh scheduler setiap 30 menit untuk dicek apakah sudah melewati batas SLA
// Status yang dicek: "dilaporkan" dan "ditugaskan" (belum dikerjakan atau selesai)
func FindPendingForEscalation() ([]model.Laporan, error) {
	var laporans []model.Laporan
	err := config.DB.Preload("Kategori").
		Where("status IN ? AND prioritas = ?", []string{"dilaporkan", "ditugaskan"}, "normal").
		Find(&laporans).Error
	return laporans, err
}
