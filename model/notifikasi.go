package model

import "time"

// Notifikasi merepresentasikan tabel "notifikasi"
// Menyimpan pesan notifikasi in-app untuk pengguna
type Notifikasi struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"` // Penerima notifikasi
	LaporanID *uint     `json:"laporan_id"`              // Opsional, referensi ke laporan
	Pesan     string    `json:"pesan" gorm:"not null"`   // Isi pesan
	IsRead    bool      `json:"is_read" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	User    *User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Laporan *Laporan `json:"laporan,omitempty" gorm:"foreignKey:LaporanID"`
}

// TableName memberi tahu GORM nama tabel yang digunakan di database
func (Notifikasi) TableName() string { return "notifikasi" }
