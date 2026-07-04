package repository

import (
	"be_silapor/config"
	"be_silapor/model"
)

// CreateNotifikasi menyimpan notifikasi baru ke database
func CreateNotifikasi(notifikasi *model.Notifikasi) error {
	return config.DB.Create(notifikasi).Error
}

// GetNotifikasiByUserID mengambil daftar notifikasi milik pengguna tertentu
func GetNotifikasiByUserID(userID uint) ([]model.Notifikasi, error) {
	var notifikasis []model.Notifikasi
	// Ambil notifikasi, urutkan dari yang terbaru, misal batasi 50 terbaru
	err := config.DB.Where("user_id = ?", userID).Order("created_at desc").Limit(50).Find(&notifikasis).Error
	return notifikasis, err
}

// MarkAsRead menandai satu notifikasi sebagai sudah dibaca
func MarkAsRead(notifID, userID uint) error {
	return config.DB.Model(&model.Notifikasi{}).
		Where("id = ? AND user_id = ?", notifID, userID).
		Update("is_read", true).Error
}

// MarkAllAsRead menandai semua notifikasi milik pengguna sebagai sudah dibaca
func MarkAllAsRead(userID uint) error {
	return config.DB.Model(&model.Notifikasi{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}
