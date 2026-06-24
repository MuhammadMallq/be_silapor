package model

import "time"

// User represents the users table
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Nama      string    `json:"nama" gorm:"not null"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"`
	Role      string    `json:"role" gorm:"not null;default:mahasiswa"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// KategoriFasilitas represents the kategori_fasilitas table
type KategoriFasilitas struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	NamaKategori string `json:"nama_kategori" gorm:"not null"`
	PetugasID    *uint  `json:"petugas_id"`
	SLAJam       int    `json:"sla_jam" gorm:"not null;default:48"`

	// Relations
	Petugas *User `json:"petugas,omitempty" gorm:"foreignKey:PetugasID"`
}

// Laporan represents the laporan table
type Laporan struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	PelaporID      uint       `json:"pelapor_id" gorm:"not null"`
	KategoriID     uint       `json:"kategori_id" gorm:"not null"`
	Lokasi         string     `json:"lokasi" gorm:"not null"`
	Deskripsi      string     `json:"deskripsi" gorm:"not null"`
	FotoURL        string     `json:"foto_url"`
	Status         string     `json:"status" gorm:"not null;default:dilaporkan"`
	Prioritas      string     `json:"prioritas" gorm:"not null;default:normal"`
	TanggalLapor   time.Time  `json:"tanggal_lapor" gorm:"autoCreateTime"`
	TanggalSelesai *time.Time `json:"tanggal_selesai"`

	// Relations
	Pelapor  *User              `json:"pelapor,omitempty" gorm:"foreignKey:PelaporID"`
	Kategori *KategoriFasilitas `json:"kategori,omitempty" gorm:"foreignKey:KategoriID"`
}

// RiwayatStatus represents the riwayat_status table
type RiwayatStatus struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	LaporanID  uint      `json:"laporan_id" gorm:"not null"`
	Status     string    `json:"status" gorm:"not null"`
	Keterangan string    `json:"keterangan"`
	Waktu      time.Time `json:"waktu" gorm:"autoCreateTime"`

	// Relations
	Laporan *Laporan `json:"laporan,omitempty" gorm:"foreignKey:LaporanID"`
}

// TableName overrides
func (User) TableName() string              { return "users" }
func (KategoriFasilitas) TableName() string { return "kategori_fasilitas" }
func (Laporan) TableName() string           { return "laporan" }
func (RiwayatStatus) TableName() string     { return "riwayat_status" }

// ---- Request / Response structs ----

type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type RegisterRequest struct {
	Nama     string `json:"nama" validate:"required"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

type CreateLaporanRequest struct {
	KategoriID uint   `json:"kategori_id" validate:"required"`
	Lokasi     string `json:"lokasi" validate:"required"`
	Deskripsi  string `json:"deskripsi" validate:"required"`
	FotoURL    string `json:"foto_url"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

type KategoriRequest struct {
	NamaKategori string `json:"nama_kategori" validate:"required"`
	PetugasID    *uint  `json:"petugas_id"`
	SLAJam       int    `json:"sla_jam"`
}
