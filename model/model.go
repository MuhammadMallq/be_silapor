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

// KategoriFasilitas merepresentasikan tabel "kategori_fasilitas"
// Menyimpan jenis-jenis fasilitas kampus (misal: Listrik, Air, Meja, dll)
type KategoriFasilitas struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	NamaKategori string `json:"nama_kategori" gorm:"not null"`
	PetugasID    *uint  `json:"petugas_id"`                   // Pointer (*uint) agar bisa bernilai null (belum ada petugas)
	SLAJam       int    `json:"sla_jam" gorm:"not null;default:48"` // Batas waktu penanganan dalam jam (default 48 jam)

	// Relasi: satu kategori punya satu petugas (opsional)
	// omitempty berarti field ini tidak dikirim jika nilainya kosong/null
	Petugas *User `json:"petugas,omitempty" gorm:"foreignKey:PetugasID"`
}

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
// Tanpa ini, GORM akan pakai nama default (users → "users", dst — sudah sama)
func (User) TableName() string              { return "users" }
func (KategoriFasilitas) TableName() string { return "kategori_fasilitas" }
func (Laporan) TableName() string           { return "laporan" }
func (RiwayatStatus) TableName() string     { return "riwayat_status" }
func (Notifikasi) TableName() string        { return "notifikasi" }

// ---- Struct untuk Request & Response API ----
// Struct-struct ini BUKAN tabel database, hanya dipakai untuk membaca/mengirim data JSON

// Response adalah format standar semua response API
// Semua endpoint mengembalikan JSON dengan format ini
type Response struct {
	Message string      `json:"message"`          // Pesan singkat hasil operasi
	Data    interface{} `json:"data,omitempty"`   // Data yang dikembalikan (bisa apapun)
	Error   string      `json:"error,omitempty"`  // Pesan error jika ada (tidak dikirim jika kosong)
}

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

// KategoriRequest adalah body JSON untuk endpoint POST/PUT /kategori
type KategoriRequest struct {
	NamaKategori string `json:"nama_kategori" validate:"required"` // Nama kategori fasilitas
	PetugasID    *uint  `json:"petugas_id"`                        // ID petugas yang bertanggung jawab (opsional)
	SLAJam       int    `json:"sla_jam"`                           // Batas waktu penanganan dalam jam
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
