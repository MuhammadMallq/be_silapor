package model

import "time"

// Laporan merepresentasikan tabel "laporan"
// Menyimpan setiap pengaduan kerusakan fasilitas yang dikirim mahasiswa
type Laporan struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	PelaporID      uint       `json:"pelapor_id" gorm:"not null"`   // ID mahasiswa yang membuat laporan
	KategoriID     uint       `json:"kategori_id" gorm:"not null"`  // ID kategori fasilitas yang dilaporkan
	Lokasi         string     `json:"lokasi" gorm:"not null"`       // Lokasi kerusakan (misal: "Gedung A Lantai 2")
	Deskripsi      string     `json:"deskripsi" gorm:"not null"`    // Penjelasan detail kerusakan
	FotoURL        string     `json:"foto_url"`                     // Link foto kerusakan (opsional)
	Status         string     `json:"status" gorm:"not null;default:dilaporkan"` // dilaporkan/ditugaskan/dikerjakan/selesai
	Prioritas      string     `json:"prioritas" gorm:"not null;default:normal"`  // normal / tinggi (naik otomatis jika lewat SLA)
	TanggalLapor   time.Time  `json:"tanggal_lapor" gorm:"autoCreateTime"` // Diisi otomatis saat laporan dibuat
	TanggalSelesai *time.Time `json:"tanggal_selesai"` // Diisi saat status berubah menjadi "selesai"
	BuktiSelesai   string     `json:"bukti_selesai"`   // Bukti foto setelah dikerjakan petugas

	PetugasID    *uint      `json:"petugas_id"`   // ID Petugas spesifik yang ditugaskan admin
	TenggatWaktu *time.Time `json:"tenggat_waktu"` // Tenggat waktu spesifik dari admin
	Rating       int        `json:"rating" gorm:"default:0"`
	Feedback     string     `json:"feedback"`

	// Relasi: laporan dimiliki oleh satu pelapor dan satu kategori
	Pelapor  *User              `json:"pelapor,omitempty" gorm:"foreignKey:PelaporID"`
	Kategori *KategoriFasilitas `json:"kategori,omitempty" gorm:"foreignKey:KategoriID"`
	Petugas  *User              `json:"petugas,omitempty" gorm:"foreignKey:PetugasID"`
}

// RiwayatStatus merepresentasikan tabel "riwayat_status"
// Mencatat setiap perubahan status pada sebuah laporan (seperti timeline/log)
type RiwayatStatus struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	LaporanID  uint      `json:"laporan_id" gorm:"not null"` // Laporan mana yang berubah statusnya
	Status     string    `json:"status" gorm:"not null"`     // Status baru saat perubahan terjadi
	Keterangan string    `json:"keterangan"`                 // Penjelasan singkat perubahan status
	Waktu      time.Time `json:"waktu" gorm:"autoCreateTime"` // Waktu perubahan status

	// Relasi ke tabel laporan
	Laporan *Laporan `json:"laporan,omitempty" gorm:"foreignKey:LaporanID"`
}

// TableName memberi tahu GORM nama tabel yang digunakan di database
func (Laporan) TableName() string           { return "laporan" }
func (RiwayatStatus) TableName() string     { return "riwayat_status" }

// CreateLaporanRequest adalah body JSON untuk endpoint POST /laporan
type CreateLaporanRequest struct {
	KategoriID uint   `json:"kategori_id" validate:"required"` // Wajib: pilih kategori fasilitas
	Lokasi     string `json:"lokasi" validate:"required"`      // Wajib: lokasi kerusakan
	Deskripsi  string `json:"deskripsi" validate:"required"`   // Wajib: penjelasan kerusakan
	FotoURL    string `json:"foto_url"`                        // Opsional: link foto
}

// UpdateStatusRequest adalah body JSON untuk endpoint PUT /laporan/:id/status
type UpdateStatusRequest struct {
	Status string `json:"status" form:"status" validate:"required"` // Nilai: dilaporkan/ditugaskan/dikerjakan/selesai
}

// RatingRequest adalah body JSON untuk endpoint PUT /laporan/:id/rating
type RatingRequest struct {
	Rating   int    `json:"rating" validate:"required,min=1,max=5"`
	Feedback string `json:"feedback"`
}

// AdminUpdateLaporanRequest adalah body JSON untuk endpoint PUT /laporan/:id/admin-update
type AdminUpdateLaporanRequest struct {
	Prioritas    string     `json:"prioritas" validate:"required"`
	PetugasID    *uint      `json:"petugas_id"`
	TenggatWaktu *time.Time `json:"tenggat_waktu"`
}
