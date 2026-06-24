package repository

import (
	"be_silapor/config"
	"be_silapor/model"
)

func CreateUser(user *model.User) error {
	return config.DB.Create(user).Error
}

func FindUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := config.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func FindUserByID(id uint) (*model.User, error) {
	var user model.User
	err := config.DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUser(user *model.User) error {
	return config.DB.Save(user).Error
}

func FindAllUsers() ([]model.User, error) {
	var users []model.User
	err := config.DB.Find(&users).Error
	return users, err
}

func FindUsersByRole(role string) ([]model.User, error) {
	var users []model.User
	err := config.DB.Where("role = ?", role).Find(&users).Error
	return users, err
}
