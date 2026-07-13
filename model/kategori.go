package model

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

// TableName memberi tahu GORM nama tabel yang digunakan di database
func (KategoriFasilitas) TableName() string { return "kategori_fasilitas" }

// KategoriRequest adalah body JSON untuk endpoint POST/PUT /kategori
type KategoriRequest struct {
	NamaKategori string `json:"nama_kategori" validate:"required"` // Nama kategori fasilitas
	PetugasID    *uint  `json:"petugas_id"`                        // ID petugas yang bertanggung jawab (opsional)
	SLAJam       int    `json:"sla_jam"`                           // Batas waktu penanganan dalam jam
}
