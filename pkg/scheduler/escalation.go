package scheduler

import (
	"log"
	"time"

	"be_silapor/model"
	"be_silapor/repository"
)

// StartEscalationScheduler menjalankan scheduler SLA di background menggunakan goroutine
// Goroutine = thread ringan di Go yang berjalan paralel tanpa memblokir aplikasi utama
// Scheduler ini berjalan terus selama aplikasi hidup, mengecek setiap 30 menit sekali
func StartEscalationScheduler() {
	// Ticker adalah "jam alarm" yang berulang setiap interval yang ditentukan
	ticker := time.NewTicker(30 * time.Minute)

	// Jalankan di goroutine (background) agar tidak menghentikan proses utama (main.go)
	go func() {
		log.Println("🔄 Scheduler eskalasi otomatis dimulai (interval: 30 menit)")
		for {
			// Tunggu sampai ticker berdetak (setiap 30 menit)
			<-ticker.C
			processEscalations()
		}
	}()
}

// processEscalations adalah fungsi inti yang dijalankan setiap 30 menit
// Tugasnya: cari laporan yang belum selesai dan sudah melewati batas waktu SLA
// Jika ditemukan, prioritasnya dinaikkan dari "normal" ke "tinggi"
func processEscalations() {
	// Ambil laporan yang status-nya masih "dilaporkan" atau "ditugaskan"
	// dan prioritasnya masih "normal" (belum pernah dieskalasi)
	laporans, err := repository.FindPendingForEscalation()
	if err != nil {
		log.Printf("❌ Error mengambil laporan untuk eskalasi: %v\n", err)
		return
	}

	now := time.Now()
	for _, lap := range laporans {
		// Hitung deadline berdasarkan SLA kategori (dalam jam)
		// Contoh: laporan dibuat jam 10.00, SLA 48 jam → deadline jam 10.00 besok lusanya
		slaHours := lap.Kategori.SLAJam
		if slaHours == 0 {
			slaHours = 48 // Gunakan default 48 jam jika SLA kategori belum diset
		}
		deadline := lap.TanggalLapor.Add(time.Duration(slaHours) * time.Hour)

		// Jika waktu sekarang sudah melewati deadline, eskalasi prioritas
		if now.After(deadline) {
			log.Printf("⚠️ Laporan ID %d melewati SLA (%d jam). Eskalasi ke prioritas 'tinggi'!\n", lap.ID, slaHours)

			// Ubah prioritas laporan menjadi "tinggi"
			lap.Prioritas = "tinggi"
			if err := repository.UpdateLaporan(&lap); err != nil {
				log.Printf("❌ Gagal update prioritas laporan ID %d: %v\n", lap.ID, err)
				continue // Lewati laporan ini, cek laporan berikutnya
			}

			// Catat perubahan prioritas ke tabel riwayat_status sebagai log otomatis sistem
			riwayat := model.RiwayatStatus{
				LaporanID:  lap.ID,
				Status:     lap.Status,
				Keterangan: "Sistem: Prioritas ditingkatkan karena melewati batas waktu penanganan (SLA)",
			}
			repository.CreateRiwayat(&riwayat)
		}
	}
}
