package config

import (
	"log"
	"os"

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

	// AutoMigrate dinonaktifkan karena skema tabel sudah dibuat langsung di Supabase
	// Menggunakan AutoMigrate di sini bisa konflik dengan tabel yang sudah ada
	log.Println("Koneksi database berhasil. (Auto-migrate dinonaktifkan)")
}
