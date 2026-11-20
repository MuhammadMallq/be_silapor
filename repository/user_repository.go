package repository

import (
	"be_silapor/config"
	"be_silapor/model"
)

// CreateUser menyimpan data user baru ke tabel "users"
func CreateUser(user *model.User) error {
	return config.DB.Create(user).Error
}

// FindUserByUsername mencari satu user berdasarkan username-nya
func FindUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := config.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByIdentifier mencari user berdasarkan username atau email
func FindUserByIdentifier(identifier string) (*model.User, error) {
	var user model.User
	err := config.DB.Where("username = ? OR (email = ? AND email != '')", identifier, identifier).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByID mencari satu user berdasarkan ID-nya
func FindUserByID(id uint) (*model.User, error) {
	var user model.User
	err := config.DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser menyimpan perubahan data user ke database
func UpdateUser(user *model.User) error {
	return config.DB.Save(user).Error
}

// UpdateUserEmail mengupdate alamat email pengguna
func UpdateUserEmail(userID uint, email string) error {
	return config.DB.Model(&model.User{}).Where("id = ?", userID).Update("email", email).Error
}

// UpdateUserPassword mengupdate password pengguna
func UpdateUserPassword(userID uint, hashedPassword string) error {
	return config.DB.Model(&model.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

// FindAllUsers mengambil semua user dari database
func FindAllUsers() ([]model.User, error) {
	var users []model.User
	err := config.DB.Preload("KategoriFasilitas").Find(&users).Error
	return users, err
}

// FindUsersByRole mengambil semua user yang memiliki role tertentu
func FindUsersByRole(role string) ([]model.User, error) {
	var users []model.User
	err := config.DB.Where("role = ?", role).Find(&users).Error
	return users, err
}

// FindAllAdmin mengambil seluruh akun bertipe admin untuk menerima notifikasi
func FindAllAdmin() ([]model.User, error) {
	var users []model.User
	err := config.DB.Where("role = ?", "admin").Find(&users).Error
	return users, err
}

// DeleteUser menghapus user berdasarkan ID
func DeleteUser(id uint) error {
	return config.DB.Delete(&model.User{}, id).Error
}
