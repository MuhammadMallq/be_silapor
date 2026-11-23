package handler

import (
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"time"

	"be_silapor/config"
	"be_silapor/model"
	"be_silapor/pkg/storage"
	"be_silapor/repository"

	"github.com/gofiber/fiber/v2"
)

// ==========================================
// CONSTANTS, VARIABLES & HELPER FUNCTIONS
// ==========================================

const maxUploadBytes = 4 * 1024 * 1024 // 4 MB

var (
	errFileTooLarge         = errors.New("gagal mengunggah: ukuran file maksimal 4MB")
	errFileTypeNotSupported = errors.New("gagal mengunggah: format file tidak didukung, hanya gunakan jpg, png, atau mp4")
)

var allowedMediaTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"video/mp4":  true,
}

var validStatuses = map[string]bool{
	"ditinjau":   true,
	"diterima":   true,
	"diproses":   true,
	"selesai":    true,
	"ditolak":    true,
	"dilaporkan": true,
	"ditugaskan": true,
	"dikerjakan": true,
}

// parseIDParam extracts and validates uint ID from URL path param :id
func parseIDParam(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// validateAndUploadMedia validates file size & MIME type before uploading to Supabase Storage
func validateAndUploadMedia(file *multipart.FileHeader) (string, error) {
	if file == nil {
		return "", nil
	}
	if file.Size > maxUploadBytes {
		return "", errFileTooLarge
	}
	contentType := file.Header.Get("Content-Type")
	if !allowedMediaTypes[contentType] {
		return "", errFileTypeNotSupported
	}
	return storage.UploadToSupabase(file)
}

// handleUploadError maps upload errors to appropriate HTTP responses
func handleUploadError(c *fiber.Ctx, err error, defaultMsg string) error {
	if errors.Is(err, errFileTooLarge) || errors.Is(err, errFileTypeNotSupported) {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: err.Error(),
		})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
		Message: defaultMsg,
		Error:   err.Error(),
	})
}

// getTargetPetugasID returns assigned Petugas ID, falling back to Category's Petugas ID
func getTargetPetugasID(laporan *model.Laporan) *uint {
	if laporan == nil {
		return nil
	}
	if laporan.PetugasID != nil {
		return laporan.PetugasID
	}
	if laporan.Kategori != nil && laporan.Kategori.PetugasID != nil {
		return laporan.Kategori.PetugasID
	}
	return nil
}

// notifyUser sends a notification record to a specific user
func notifyUser(userID uint, laporanID uint, pesan string) {
	_ = repository.CreateNotifikasi(&model.Notifikasi{
		UserID:    userID,
		LaporanID: &laporanID,
		Pesan:     pesan,
	})
}

// notifyAdmins sends notifications to all admin users with optional exclusion list
func notifyAdmins(laporanID uint, pesan string, excludeUserIDs ...uint) {
	admins, err := repository.FindAllAdmin()
	if err != nil {
		return
	}
	excludeMap := make(map[uint]bool, len(excludeUserIDs))
	for _, id := range excludeUserIDs {
		excludeMap[id] = true
	}
	for _, adm := range admins {
		if !excludeMap[adm.ID] {
			notifyUser(adm.ID, laporanID, pesan)
		}
	}
}

// getStatusTransitionMessages provides standard log & user notification text for status transitions
func getStatusTransitionMessages(status, alasanPenolakan, deskripsi string) (keterangan string, pesanNotif string) {
	switch status {
	case "ditinjau":
		return "Laporan masuk tahap peninjauan oleh Admin", "Laporan Anda sedang ditinjau oleh Admin"
	case "diterima":
		return "Laporan disetujui oleh Admin dan diteruskan ke Petugas", "Kabar baik! Laporan Anda telah disetujui oleh Admin dan segera diproses"
	case "diproses", "dikerjakan":
		return "Laporan sedang diproses dan dikerjakan oleh Petugas", "Laporan Anda sedang diproses oleh petugas teknisi"
	case "selesai":
		return "Laporan telah selesai ditangani", "Laporan Anda telah selesai ditangani! Silakan beri ulasan/rating"
	case "ditolak":
		return "Laporan ditolak oleh Admin. Alasan: " + alasanPenolakan, "Laporan Anda ditolak oleh Admin. Alasan: " + alasanPenolakan
	default:
		return "Status diubah menjadi " + status, "Status laporan Anda tentang " + deskripsi + " diubah menjadi: " + status
	}
}

// ==========================================
// LAPORAN HTTP HANDLERS
// ==========================================

// GetAllLaporan godoc
// @Summary Daftar laporan
// @Description Mengambil daftar laporan sesuai role user
// @Tags Laporan
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.Response
// @Router /api/laporan [get]
func GetAllLaporan(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	role := c.Locals("role").(string)

	filter := repository.LaporanFilter{
		Status:      c.Query("status"),
		RolePelapor: c.Query("role_pelapor"),
		Prioritas:   c.Query("prioritas"),
	}

	if katIDStr := c.Query("kategori_id"); katIDStr != "" {
		if kid, err := strconv.ParseUint(katIDStr, 10, 32); err == nil {
			filter.KategoriID = uint(kid)
		}
	}

	if c.Query("terlambat") == "true" {
		filter.Terlambat = true
	}

	if riwayatStr := c.Query("riwayat"); riwayatStr == "true" {
		rTrue := true
		filter.Riwayat = &rTrue
	} else if riwayatStr == "false" {
		rFalse := false
		filter.Riwayat = &rFalse
	}

	switch role {
	case "admin":
		// Admin dapat melihat seluruh laporan berdasarkan filter query
	case "petugas":
		filter.PetugasID = userID
	case "mahasiswa", "dosen", "staff":
		filter.PelaporID = userID
	default:
		return c.Status(fiber.StatusForbidden).JSON(model.Response{
			Message: "role tidak dikenali",
		})
	}

	laporans, err := repository.FindLaporanWithFilter(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal mengambil data laporan",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.Response{
		Message: "berhasil mengambil data laporan",
		Data:    laporans,
	})
}

// GetLaporanByID godoc
// @Summary Detail laporan
// @Description Mengambil detail satu laporan
// @Tags Laporan
// @Produce json
// @Param id path int true "Laporan ID"
// @Security BearerAuth
// @Success 200 {object} model.Response
// @Failure 404 {object} model.Response
// @Router /api/laporan/{id} [get]
func GetLaporanByID(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "ID tidak valid",
			Error:   err.Error(),
		})
	}

	laporan, err := repository.FindLaporanByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.Response{
			Message: "laporan tidak ditemukan",
			Error:   err.Error(),
		})
	}

	// Authorization check
	userID := c.Locals("user_id").(uint)
	role := c.Locals("role").(string)

	if (role == "mahasiswa" || role == "dosen" || role == "staff") && laporan.PelaporID != userID {
		return c.Status(fiber.StatusForbidden).JSON(model.Response{
			Message: "Akses ditolak: Anda tidak memiliki akses ke laporan ini",
		})
	}

	return c.JSON(model.Response{
		Message: "berhasil mengambil detail laporan",
		Data:    laporan,
	})
}

// CreateLaporan godoc
// @Summary Buat laporan baru
// @Description Mahasiswa membuat laporan kerusakan baru
// @Tags Laporan
// @Accept multipart/form-data
// @Produce json
// @Param deskripsi formData string true "Penjelasan kerusakan"
// @Param lokasi formData string true "Lokasi kerusakan"
// @Param kategori_id formData int true "Kategori ID"
// @Param bukti formData file false "Foto bukti kerusakan (opsional)"
// @Security BearerAuth
// @Success 201 {object} model.Response
// @Failure 400 {object} model.Response
// @Router /api/laporan [post]
func CreateLaporan(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	role := c.Locals("role").(string)

	kategoriIDStr := c.FormValue("kategori_id")
	lokasi := c.FormValue("lokasi")
	deskripsi := c.FormValue("deskripsi")

	if kategoriIDStr == "" || lokasi == "" || deskripsi == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "kategori, lokasi, dan deskripsi wajib diisi",
		})
	}

	kategoriID, err := strconv.ParseUint(kategoriIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "kategori_id tidak valid",
			Error:   err.Error(),
		})
	}

	// Verify kategori exists
	kategori, err := repository.FindKategoriByID(uint(kategoriID))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "kategori tidak ditemukan",
			Error:   err.Error(),
		})
	}

	// Tentukan siapa Pelapor:
	// - Jika role non-admin, pakai ID dirinya sendiri
	// - Jika role "admin", bisa isi form "pelapor_id" untuk buat laporan atas nama pengguna tertentu.
	pelaporID := userID
	if role == "admin" {
		if pelaporIDStr := c.FormValue("pelapor_id"); pelaporIDStr != "" {
			parsedID, parseErr := strconv.ParseUint(pelaporIDStr, 10, 32)
			if parseErr != nil {
				return c.Status(fiber.StatusBadRequest).JSON(model.Response{
					Message: "pelapor_id tidak valid",
					Error:   parseErr.Error(),
				})
			}
			pelaporID = uint(parsedID)
		}
	}

	// Handle file upload
	var fotoURL string
	if file, err := c.FormFile("bukti"); err == nil && file != nil {
		url, uploadErr := validateAndUploadMedia(file)
		if uploadErr != nil {
			return handleUploadError(c, uploadErr, "gagal mengunggah bukti")
		}
		fotoURL = url
	}

	// Status awal: Setiap laporan masuk ke status "ditinjau" oleh admin
	const initialStatus = "ditinjau"
	laporan := model.Laporan{
		PelaporID:  pelaporID,
		KategoriID: uint(kategoriID),
		Lokasi:     lokasi,
		Deskripsi:  deskripsi,
		FotoURL:    fotoURL,
		Status:     initialStatus,
		Prioritas:  "normal",
		PetugasID:  kategori.PetugasID,
	}

	if err := repository.CreateLaporan(&laporan); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal membuat laporan",
			Error:   err.Error(),
		})
	}

	// Buat riwayat awal
	repository.CreateRiwayat(&model.RiwayatStatus{
		LaporanID:  laporan.ID,
		Status:     initialStatus,
		Keterangan: "Laporan baru dibuat dan sedang ditinjau oleh Admin",
	})

	// Broadcast notifikasi laporan baru ke semua Admin
	pelaporName := "Pengguna"
	if pUser, _ := repository.FindUserByID(pelaporID); pUser != nil {
		pelaporName = pUser.Nama
	}
	notifyAdmins(laporan.ID, fmt.Sprintf("📋 Laporan baru dari %s: %s (%s)", pelaporName, laporan.Deskripsi, laporan.Lokasi))

	// Notifikasi konfirmasi ke Pelapor
	notifyUser(pelaporID, laporan.ID, fmt.Sprintf("✅ Laporan Anda (%s) berhasil dibuat dan sedang menunggu peninjauan Admin", laporan.Deskripsi))

	// Notifikasi ke petugas kategori (jika sudah ada penanggung jawabnya)
	if kategori.PetugasID != nil {
		notifyUser(*kategori.PetugasID, laporan.ID, fmt.Sprintf("📋 Laporan baru di kategori Anda: %s (%s)", laporan.Deskripsi, laporan.Lokasi))
	}

	// Reload with relations
	result, _ := repository.FindLaporanByID(laporan.ID)

	return c.Status(fiber.StatusCreated).JSON(model.Response{
		Message: "laporan berhasil dibuat",
		Data:    result,
	})
}

// UpdateStatusLaporan godoc
// @Summary Ubah status laporan
// @Description Petugas mengubah status laporan
// @Tags Laporan
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Laporan ID"
// @Param status formData string true "Status baru (dilaporkan/ditugaskan/dikerjakan/selesai)"
// @Param bukti_selesai formData file false "Bukti penyelesaian (wajib jika status selesai)"
// @Security BearerAuth
// @Success 200 {object} model.Response
// @Failure 400 {object} model.Response
// @Failure 404 {object} model.Response
// @Router /api/laporan/{id}/status [put]
func UpdateStatusLaporan(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "ID tidak valid",
			Error:   err.Error(),
		})
	}

	var req model.UpdateStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "payload tidak valid",
			Error:   err.Error(),
		})
	}

	if !validStatuses[req.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "status tidak valid. Gunakan: ditinjau, diterima, diproses, selesai, atau ditolak",
		})
	}

	laporan, err := repository.FindLaporanByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.Response{
			Message: "laporan tidak ditemukan",
			Error:   err.Error(),
		})
	}

	// Authorization check
	userID := c.Locals("user_id").(uint)
	role := c.Locals("role").(string)

	// Validasi peninjauan oleh Admin (hanya admin yang berhak menerima atau menolak)
	if (req.Status == "diterima" || req.Status == "ditolak") && role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(model.Response{
			Message: "Akses ditolak: Hanya Admin yang berwenang menerima atau menolak laporan",
		})
	}

	if role == "petugas" {
		targetPetugas := getTargetPetugasID(laporan)
		if targetPetugas == nil || *targetPetugas != userID {
			return c.Status(fiber.StatusForbidden).JSON(model.Response{
				Message: "Akses ditolak: Anda tidak memiliki wewenang mengubah status laporan ini",
			})
		}
	}

	// Penanganan status Ditolak
	if req.Status == "ditolak" {
		alasan := req.AlasanPenolakan
		if alasan == "" {
			alasan = c.FormValue("alasan_penolakan")
		}
		if alasan == "" {
			return c.Status(fiber.StatusBadRequest).JSON(model.Response{
				Message: "alasan penolakan wajib diisi",
			})
		}
		laporan.AlasanPenolakan = alasan

		// Upload bukti penolakan jika ada (opsional)
		if filePenolakan, errFile := c.FormFile("bukti_penolakan"); errFile == nil && filePenolakan != nil {
			urlBukti, uploadErr := validateAndUploadMedia(filePenolakan)
			if uploadErr != nil {
				return handleUploadError(c, uploadErr, "gagal mengunggah bukti penolakan")
			}
			laporan.BuktiPenolakan = urlBukti
		}
	}

	// Penanganan status Selesai
	if req.Status == "selesai" {
		now := time.Now()
		laporan.TanggalSelesai = &now

		file, err := c.FormFile("bukti_selesai")
		if err != nil || file == nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.Response{
				Message: "bukti penyelesaian (foto/video) wajib diunggah",
			})
		}

		url, uploadErr := validateAndUploadMedia(file)
		if uploadErr != nil {
			return handleUploadError(c, uploadErr, "gagal mengunggah bukti selesai")
		}
		laporan.BuktiSelesai = url
	}

	laporan.Status = req.Status
	keterangan, pesanNotif := getStatusTransitionMessages(req.Status, laporan.AlasanPenolakan, laporan.Deskripsi)

	if err := repository.UpdateLaporan(laporan); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal mengubah status laporan",
			Error:   err.Error(),
		})
	}

	repository.CreateRiwayat(&model.RiwayatStatus{
		LaporanID:  laporan.ID,
		Status:     req.Status,
		Keterangan: keterangan,
	})

	// Buat notifikasi ke Pelapor
	notifyUser(laporan.PelaporID, laporan.ID, pesanNotif)

	// Notifikasi ke Petugas jika laporan diterima atau ditugaskan
	if req.Status == "diterima" || req.Status == "ditugaskan" {
		if petugasTargetID := getTargetPetugasID(laporan); petugasTargetID != nil {
			notifyUser(*petugasTargetID, laporan.ID, fmt.Sprintf("📋 Laporan #%d telah disetujui Admin dan siap Anda kerjakan: %s", laporan.ID, laporan.Deskripsi))
		}
	}

	// Notifikasi ke Admin jika status laporan diproses atau selesai oleh petugas
	if req.Status == "diproses" || req.Status == "dikerjakan" || req.Status == "selesai" {
		adminMsg := fmt.Sprintf("🔧 Laporan #%d mulai dikerjakan oleh petugas", laporan.ID)
		if req.Status == "selesai" {
			adminMsg = fmt.Sprintf("✅ Laporan #%d telah diselesaikan oleh petugas", laporan.ID)
		}
		notifyAdmins(laporan.ID, adminMsg, userID)
	}

	return c.JSON(model.Response{
		Message: "status laporan berhasil diubah",
		Data:    laporan,
	})
}

// DeleteLaporan godoc
// @Summary Hapus laporan
// @Description Admin menghapus laporan
// @Tags Laporan
// @Produce json
// @Param id path int true "Laporan ID"
// @Security BearerAuth
// @Success 200 {object} model.Response
// @Failure 404 {object} model.Response
// @Router /api/laporan/{id} [delete]
func DeleteLaporan(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "ID tidak valid",
			Error:   err.Error(),
		})
	}

	if _, err := repository.FindLaporanByID(id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.Response{
			Message: "laporan tidak ditemukan",
			Error:   err.Error(),
		})
	}

	if err := repository.DeleteLaporan(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal menghapus laporan",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.Response{
		Message: "laporan berhasil dihapus",
	})
}

// GetRiwayatLaporan godoc
// @Summary Riwayat status laporan
// @Description Melihat timeline riwayat perubahan status
// @Tags Laporan
// @Produce json
// @Param id path int true "Laporan ID"
// @Security BearerAuth
// @Success 200 {object} model.Response
// @Router /api/laporan/{id}/riwayat [get]
func GetRiwayatLaporan(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "ID tidak valid",
			Error:   err.Error(),
		})
	}

	// Verify laporan exists
	if _, err := repository.FindLaporanByID(id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.Response{
			Message: "laporan tidak ditemukan",
			Error:   err.Error(),
		})
	}

	riwayats, err := repository.FindRiwayatByLaporanID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal mengambil riwayat status",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.Response{
		Message: "berhasil mengambil riwayat status laporan",
		Data:    riwayats,
	})
}

// UpdateRating godoc
// @Summary Beri rating laporan
// @Description Mahasiswa memberikan rating (1-5) dan ulasan (feedback) pada laporan yang sudah selesai
// @Tags Laporan
// @Accept json
// @Produce json
// @Param id path int true "Laporan ID"
// @Param body body model.RatingRequest true "Data rating & feedback"
// @Security BearerAuth
// @Success 200 {object} model.Response
// @Failure 400 {object} model.Response
// @Failure 403 {object} model.Response
// @Failure 404 {object} model.Response
// @Router /api/laporan/{id}/rating [put]
func UpdateRating(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{Message: "ID tidak valid"})
	}

	var req model.RatingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{Message: "Data tidak valid"})
	}

	laporan, err := repository.FindLaporanByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.Response{Message: "Laporan tidak ditemukan"})
	}

	// Authorization check
	userID := c.Locals("user_id").(uint)
	if laporan.PelaporID != userID {
		return c.Status(fiber.StatusForbidden).JSON(model.Response{Message: "Akses ditolak: Anda hanya dapat memberi rating pada laporan Anda sendiri"})
	}

	// Pastikan status sudah selesai
	if laporan.Status != "selesai" {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{Message: "Laporan belum selesai, tidak bisa memberi rating"})
	}

	// Update laporan
	laporan.Rating = req.Rating
	laporan.Feedback = req.Feedback

	if err := repository.UpdateLaporan(laporan); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{Message: "Gagal menyimpan rating"})
	}

	// Buat notifikasi ke Petugas
	if petugasRatingTarget := getTargetPetugasID(laporan); petugasRatingTarget != nil {
		notifyUser(*petugasRatingTarget, laporan.ID, fmt.Sprintf("⭐ Anda mendapat rating %d/5 untuk laporan: %s", req.Rating, laporan.Deskripsi))
	}

	// Buat notifikasi ke seluruh Admin
	notifyAdmins(laporan.ID, fmt.Sprintf("⭐ Laporan #%d telah diberi rating %d/5 oleh pelapor", id, req.Rating))

	return c.JSON(model.Response{
		Message: "Rating berhasil disimpan",
	})
}

// AdminUpdateLaporan godoc
// @Summary Update pengaturan laporan khusus Admin
// @Description Admin dapat merubah tingkat prioritas laporan, menunjuk ulang petugas, atau mengatur tenggat waktu khusus (SLA kustom)
// @Tags Laporan
// @Accept json
// @Produce json
// @Param id path int true "Laporan ID"
// @Param body body model.AdminUpdateLaporanRequest true "Data prioritas, penugasan, dan tenggat"
// @Security BearerAuth
// @Success 200 {object} model.Response
// @Failure 400 {object} model.Response
// @Failure 404 {object} model.Response
// @Router /api/laporan/{id}/admin-update [put]
func AdminUpdateLaporan(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{Message: "ID tidak valid"})
	}

	var req model.AdminUpdateLaporanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{Message: "Data tidak valid"})
	}

	laporan, err := repository.FindLaporanByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.Response{Message: "Laporan tidak ditemukan"})
	}

	// Update Prioritas
	if req.Prioritas != "" && req.Prioritas != laporan.Prioritas {
		repository.CreateRiwayat(&model.RiwayatStatus{
			LaporanID:  laporan.ID,
			Status:     laporan.Status,
			Keterangan: "Prioritas diubah menjadi " + req.Prioritas + " oleh Admin",
		})
		if req.Prioritas == "tinggi" {
			if pTarget := getTargetPetugasID(laporan); pTarget != nil {
				notifyUser(*pTarget, laporan.ID, fmt.Sprintf("⚠️ Laporan #%d (%s) dinaikkan ke prioritas TINGGI oleh Admin", laporan.ID, laporan.Deskripsi))
			}
		}
	}

	// Siapkan map kolom yang akan diupdate secara eksplisit
	// Menggunakan map agar GORM tidak menimpa field lain via asosiasi
	updates := make(map[string]interface{})

	if req.Prioritas != "" {
		updates["prioritas"] = req.Prioritas
	}

	// Update PetugasID
	if req.PetugasID != nil {
		if *req.PetugasID == 0 {
			// Kosongkan petugas (set NULL)
			updates["petugas_id"] = nil
			// Kembalikan status ke ditinjau jika petugas dikosongkan
			if laporan.Status == "diterima" || laporan.Status == "ditugaskan" {
				updates["status"] = "ditinjau"
			}
		} else {
			updates["petugas_id"] = *req.PetugasID
			// Otomatis ubah status menjadi "diterima" saat ada petugas
			if laporan.Status == "ditinjau" || laporan.Status == "dilaporkan" {
				updates["status"] = "diterima"
				repository.CreateRiwayat(&model.RiwayatStatus{
					LaporanID:  laporan.ID,
					Status:     "diterima",
					Keterangan: "Laporan diterima dan ditugaskan ke petugas oleh Admin",
				})
			}
			// Buat notifikasi ke petugas baru (hanya jika petugas berubah)
			if laporan.PetugasID == nil || *laporan.PetugasID != *req.PetugasID {
				notifyUser(*req.PetugasID, laporan.ID, "Anda ditugaskan pada laporan: "+laporan.Deskripsi)
			}
		}
	}

	// Update Tenggat Waktu (nil = hapus tenggat)
	updates["tenggat_waktu"] = req.TenggatWaktu

	if len(updates) == 0 {
		return c.JSON(model.Response{Message: "Tidak ada perubahan yang dilakukan"})
	}

	// Gunakan Updates dengan map agar hanya kolom yang dituju yang diupdate
	// Ini menghindari bug GORM Save() yang bisa menimpa relasi
	if err := config.DB.Model(&model.Laporan{}).Where("id = ?", laporan.ID).Updates(updates).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{Message: "Gagal mengupdate laporan"})
	}

	return c.JSON(model.Response{
		Message: "Laporan berhasil diupdate oleh Admin",
	})
}
