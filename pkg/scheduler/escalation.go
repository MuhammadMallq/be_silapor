package scheduler

import (
	"log"
	"time"

	"be_silapor/model"
	"be_silapor/repository"
)

// StartEscalationScheduler runs periodically to check for reports that missed SLA
func StartEscalationScheduler() {
	ticker := time.NewTicker(30 * time.Minute)

	go func() {
		log.Println("🔄 Scheduler eskalasi otomatis dimulai (interval: 30 menit)")
		for {
			<-ticker.C
			processEscalations()
		}
	}()
}

func processEscalations() {
	laporans, err := repository.FindPendingForEscalation()
	if err != nil {
		log.Printf("❌ Error mengambil laporan untuk eskalasi: %v\n", err)
		return
	}

	now := time.Now()
	for _, lap := range laporans {
		// Calculate SLA deadline
		slaHours := lap.Kategori.SLAJam
		if slaHours == 0 {
			slaHours = 48 // fallback default
		}
		deadline := lap.TanggalLapor.Add(time.Duration(slaHours) * time.Hour)

		if now.After(deadline) {
			log.Printf("⚠️ Laporan ID %d melewati SLA (%d jam). Eskalasi ke prioritas 'tinggi'!\n", lap.ID, slaHours)

			// Update prioritas
			lap.Prioritas = "tinggi"
			if err := repository.UpdateLaporan(&lap); err != nil {
				log.Printf("❌ Gagal update prioritas laporan ID %d: %v\n", lap.ID, err)
				continue
			}

			// Catat ke riwayat
			riwayat := model.RiwayatStatus{
				LaporanID:  lap.ID,
				Status:     lap.Status,
				Keterangan: "Sistem: Prioritas ditingkatkan karena melewati batas waktu penanganan (SLA)",
			}
			repository.CreateRiwayat(&riwayat)
		}
	}
}
