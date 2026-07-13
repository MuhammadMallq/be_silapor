package model

import "time"

// User merepresentasikan tabel "users" di database
// Menyimpan data akun pengguna: mahasiswa, petugas, dan admin
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Nama      string    `json:"nama" gorm:"not null"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null"` // Harus unik, tidak boleh sama
	Password  string    `json:"-" gorm:"not null"`                    // json:"-" berarti password TIDAK dikirim ke response
	Role      string    `json:"role" gorm:"not null;default:mahasiswa"` // nilai: mahasiswa / petugas / admin
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`     // Diisi otomatis saat data dibuat

	// Relasi: User (Petugas) bisa memegang beberapa kategori (opsional, untuk preload di GetAllUsers)
	KategoriFasilitas []KategoriFasilitas `json:"kategori_fasilitas,omitempty" gorm:"foreignKey:PetugasID"`
}

// TableName memberi tahu GORM nama tabel yang digunakan di database
func (User) TableName() string { return "users" }

// RegisterRequest adalah body JSON yang diterima saat endpoint POST /register dipanggil
type RegisterRequest struct {
	Nama     string `json:"nama" validate:"required"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role"` // Opsional: mahasiswa (default) / petugas / admin
}

// LoginRequest adalah body JSON yang diterima saat endpoint POST /login dipanggil
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse adalah data yang dikembalikan setelah login berhasil
// Berisi token JWT dan data user yang bisa dipakai oleh frontend
type LoginResponse struct {
	Token string `json:"token"` // JWT token untuk autentikasi request berikutnya
	User  User   `json:"user"`  // Data profil user yang baru login
}

// ChangePasswordRequest adalah body JSON untuk endpoint PUT /changepassword
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}
