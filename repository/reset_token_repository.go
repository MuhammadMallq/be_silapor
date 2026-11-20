package repository

import (
	"be_silapor/config"
	"be_silapor/model"
	"time"
)

// CreateResetToken membuat token OTP reset password baru
func CreateResetToken(token *model.PasswordResetToken) error {
	return config.DB.Create(token).Error
}

// FindValidResetToken mencari token berdasarkan hash dan memastikan belum expired / belum used
func FindValidResetToken(tokenHash string) (*model.PasswordResetToken, error) {
	var token model.PasswordResetToken
	now := time.Now()
	err := config.DB.Preload("User").
		Where("token_hash = ? AND used = ? AND expires_at > ?", tokenHash, false, now).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// MarkTokenAsUsed menandai token telah digunakan
func MarkTokenAsUsed(id uint) error {
	return config.DB.Model(&model.PasswordResetToken{}).Where("id = ?", id).Update("used", true).Error
}

// InvalidateOldTokensForUser membatalkan token-token lama milik pengguna saat meminta OTP baru
func InvalidateOldTokensForUser(userID uint) error {
	return config.DB.Model(&model.PasswordResetToken{}).
		Where("user_id = ? AND used = ?", userID, false).
		Update("used", true).Error
}
