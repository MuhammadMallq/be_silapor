package repository

import (
	"be_silapor/config"
	"be_silapor/model"
)

// CreateUser menyimpan data user baru ke tabel "users"
// Dipanggil saat mahasiswa/admin/petugas mendaftar akun baru
func CreateUser(user *model.User) error {
	return config.DB.Create(user).Error
}

// FindUserByUsername mencari satu user berdasarkan username-nya
// Dipanggil saat proses login untuk memverifikasi identitas pengguna
func FindUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := config.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByID mencari satu user berdasarkan ID-nya
// Dipanggil saat user ingin ganti password (butuh data user saat ini)
func FindUserByID(id uint) (*model.User, error) {
	var user model.User
	err := config.DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser menyimpan perubahan data user ke database
// Dipanggil saat password user diubah (menyimpan hash password baru)
func UpdateUser(user *model.User) error {
	return config.DB.Save(user).Error
}

// FindAllUsers mengambil semua user dari database
// Bisa dipakai untuk fitur manajemen user oleh admin
func FindAllUsers() ([]model.User, error) {
	var users []model.User
	err := config.DB.Find(&users).Error
	return users, err
}

// FindUsersByRole mengambil semua user yang memiliki role tertentu
// Contoh: FindUsersByRole("petugas") → ambil semua petugas
// Berguna saat admin ingin menugaskan petugas ke suatu kategori
func FindUsersByRole(role string) ([]model.User, error) {
	var users []model.User
	err := config.DB.Where("role = ?", role).Find(&users).Error
	return users, err
}
