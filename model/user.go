package model

import "time"

// User merepresentasikan tabel "users" di database
// Menyimpan data akun pengguna: mahasiswa, dosen, staff, petugas, dan admin
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Nama      string    `json:"nama" gorm:"not null"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null"`   // Harus unik, tidak boleh sama
	Email     string    `json:"email" gorm:"index"`                     // Opsional / nullable untuk pemulihan password
	Password  string    `json:"-" gorm:"not null"`                      // json:"-" berarti password TIDAK dikirim ke response
	Role      string    `json:"role" gorm:"not null;default:mahasiswa"` // nilai: mahasiswa / dosen / staff / petugas / admin
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`       // Diisi otomatis saat data dibuat

	// Relasi: User (Petugas) bisa memegang beberapa kategori (opsional, untuk preload di GetAllUsers)
	KategoriFasilitas []KategoriFasilitas `json:"kategori_fasilitas,omitempty" gorm:"foreignKey:PetugasID"`
}

// TableName memberi tahu GORM nama tabel yang digunakan di database
func (User) TableName() string { return "users" }

// PasswordResetToken merepresentasikan tabel "password_reset_tokens"
type PasswordResetToken struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	TokenHash string    `json:"token_hash" gorm:"not null;index"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	Used      bool      `json:"used" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (PasswordResetToken) TableName() string { return "password_reset_tokens" }

// RegisterRequest adalah body JSON yang diterima saat endpoint POST /register atau POST /users dipanggil
type RegisterRequest struct {
	Nama     string `json:"nama" validate:"required"`
	Username string `json:"username" validate:"required"`
	Email    string `json:"email"` // Opsional
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role"` // Opsional: mahasiswa (default) / dosen / staff / petugas / admin
}

// LoginRequest adalah body JSON yang diterima saat endpoint POST /login dipanggil
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse adalah data yang dikembalikan setelah login berhasil
type LoginResponse struct {
	Token string `json:"token"` // JWT token untuk autentikasi request berikutnya
	User  User   `json:"user"`  // Data profil user yang baru login
}

// ChangePasswordRequest adalah body JSON untuk endpoint PUT /changepassword
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ForgotPasswordRequest meminta pengiriman OTP
type ForgotPasswordRequest struct {
	Identifier string `json:"identifier" validate:"required"` // Username atau Email
}

// VerifyOTPRequest memverifikasi kode OTP
type VerifyOTPRequest struct {
	Identifier string `json:"identifier" validate:"required"`
	OTP        string `json:"otp" validate:"required"`
}

// ResetPasswordRequest mereset password baru dengan token verifikasi OTP
type ResetPasswordRequest struct {
	ResetToken  string `json:"reset_token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// UpdateProfileRequest untuk update email pengguna di Profil
type UpdateProfileRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// AdminResetPasswordRequest untuk admin mereset password pengguna secara manual
type AdminResetPasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required,min=8"`
}
