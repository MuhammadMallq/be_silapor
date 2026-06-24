package repository

import (
	"be_silapor/config"
	"be_silapor/model"
)

// CreateKategori menyimpan kategori fasilitas baru ke database
// Dipanggil saat admin menambah kategori baru (misal: "Kerusakan Listrik", "Kebersihan")
func CreateKategori(kategori *model.KategoriFasilitas) error {
	return config.DB.Create(kategori).Error
}

// FindAllKategori mengambil semua kategori fasilitas dari database
// Preload("Petugas") → ikut mengambil data petugas yang bertanggung jawab di tiap kategori
// Dipakai saat mahasiswa ingin memilih kategori saat membuat laporan
func FindAllKategori() ([]model.KategoriFasilitas, error) {
	var kategoris []model.KategoriFasilitas
	err := config.DB.Preload("Petugas").Find(&kategoris).Error
	return kategoris, err
}

// FindKategoriByID mengambil satu kategori berdasarkan ID-nya
// Dipakai untuk validasi sebelum update/delete dan saat membuat laporan baru
func FindKategoriByID(id uint) (*model.KategoriFasilitas, error) {
	var kategori model.KategoriFasilitas
	err := config.DB.Preload("Petugas").First(&kategori, id).Error
	if err != nil {
		return nil, err
	}
	return &kategori, nil
}

// UpdateKategori menyimpan perubahan data kategori ke database
// Dipakai saat admin mengedit nama kategori, mengganti petugas, atau mengubah SLA
func UpdateKategori(kategori *model.KategoriFasilitas) error {
	return config.DB.Save(kategori).Error
}

// DeleteKategori menghapus kategori berdasarkan ID
// Hanya admin yang bisa melakukan ini (diatur di router)
func DeleteKategori(id uint) error {
	return config.DB.Delete(&model.KategoriFasilitas{}, id).Error
}
