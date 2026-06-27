package config

import (
	"log"
	"os"

	"be_silapor/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB adalah variabel global yang menyimpan koneksi database
// Bisa diakses dari package lain dengan config.DB
var DB *gorm.DB

// ConnectDatabase membuka koneksi ke database Supabase PostgreSQL
// Dipanggil sekali saat aplikasi pertama kali dijalankan di main.go
func ConnectDatabase() {
	// Membaca connection string dari variabel SUPABASE_DSN di file .env
	// Format DSN: postgresql://user:password@host:port/dbname
	dsn := os.Getenv("SUPABASE_DSN")

	var err error
	// Membuka koneksi menggunakan GORM dengan driver PostgreSQL
	// PreferSimpleProtocol: true → wajib untuk Supabase pooler di port 6543
	// Tanpa ini, koneksi ke Supabase bisa gagal karena protokol yang tidak kompatibel
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		// LogMode Info: menampilkan query SQL di terminal untuk keperluan debugging
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		// Jika koneksi gagal, aplikasi langsung berhenti (Fatal)
		// Karena tanpa database, aplikasi tidak bisa berjalan sama sekali
		log.Fatal("Gagal koneksi ke database: ", err)
	}

	log.Println("Berhasil terhubung ke database Supabase")

	// Tambahkan kolom bukti_selesai secara manual jika belum ada, 
	// agar tidak memicu error AutoMigrate pada relasi tabel users.
	if !DB.Migrator().HasColumn(&model.Laporan{}, "BuktiSelesai") {
		err = DB.Migrator().AddColumn(&model.Laporan{}, "BuktiSelesai")
		if err != nil {
			log.Println("Peringatan: Gagal menambahkan kolom bukti_selesai:", err)
		} else {
			log.Println("Berhasil menambahkan kolom bukti_selesai ke tabel laporan.")
		}
	} else {
		log.Println("Kolom bukti_selesai sudah ada di tabel laporan.")
	}

	log.Println("Koneksi database berhasil.")
}
