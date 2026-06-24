package repository

import (
	"be_silapor/config"
	"be_silapor/model"
)

func CreateKategori(kategori *model.KategoriFasilitas) error {
	return config.DB.Create(kategori).Error
}

func FindAllKategori() ([]model.KategoriFasilitas, error) {
	var kategoris []model.KategoriFasilitas
	err := config.DB.Preload("Petugas").Find(&kategoris).Error
	return kategoris, err
}

func FindKategoriByID(id uint) (*model.KategoriFasilitas, error) {
	var kategori model.KategoriFasilitas
	err := config.DB.Preload("Petugas").First(&kategori, id).Error
	if err != nil {
		return nil, err
	}
	return &kategori, nil
}

func UpdateKategori(kategori *model.KategoriFasilitas) error {
	return config.DB.Save(kategori).Error
}

func DeleteKategori(id uint) error {
	return config.DB.Delete(&model.KategoriFasilitas{}, id).Error
}
