package config

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := os.Getenv("SUPABASE_DSN")

	var err error
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // required for Supabase pooler (port 6543)
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Gagal koneksi ke database: ", err)
	}

	log.Println("✅ Berhasil terhubung ke database Supabase")

	// Disable AutoMigrate since we use database.sql for schema creation
	// This prevents conflicts between GORM and Supabase's existing tables
	log.Println("✅ Koneksi database berhasil. (Auto-migrate dinonaktifkan)")
}
