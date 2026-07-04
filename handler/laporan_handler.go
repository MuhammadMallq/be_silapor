package handler

import (
	"strconv"
	"time"

	"be_silapor/model"
	"be_silapor/pkg/storage"
	"be_silapor/repository"

	"github.com/gofiber/fiber/v2"
)

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

	var laporans []model.Laporan
	var err error

	switch role {
	case "admin":
		laporans, err = repository.FindAllLaporan()
	case "petugas":
		laporans, err = repository.FindLaporanByKategoriPetugasID(userID)
	case "mahasiswa":
		laporans, err = repository.FindLaporanByPelaporID(userID)
	default:
		return c.Status(fiber.StatusForbidden).JSON(model.Response{
			Message: "role tidak dikenali",
		})
	}

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
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "ID tidak valid",
			Error:   err.Error(),
		})
	}

	laporan, err := repository.FindLaporanByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.Response{
			Message: "laporan tidak ditemukan",
			Error:   err.Error(),
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

	// Handle file upload
	var fotoURL string
	file, err := c.FormFile("bukti")
	if err == nil && file != nil { // File is optional
		// Upload to Supabase
		url, uploadErr := storage.UploadToSupabase(file)
		if uploadErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
				Message: "gagal mengunggah bukti gambar",
				Error:   uploadErr.Error(),
			})
		}
		fotoURL = url
	}

	// Determine initial status based on whether petugas is assigned
	initialStatus := "dilaporkan"
	if kategori.PetugasID != nil {
		initialStatus = "ditugaskan"
	}

	laporan := model.Laporan{
		PelaporID:  userID,
		KategoriID: uint(kategoriID),
		Lokasi:     lokasi,
		Deskripsi:  deskripsi,
		FotoURL:    fotoURL,
		Status:     initialStatus,
		Prioritas:  "normal",
	}

	if err := repository.CreateLaporan(&laporan); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal membuat laporan",
			Error:   err.Error(),
		})
	}

	// Create initial riwayat
	riwayat := model.RiwayatStatus{
		LaporanID:  laporan.ID,
		Status:     initialStatus,
		Keterangan: "Laporan baru dibuat",
	}
	repository.CreateRiwayat(&riwayat)

	// If auto-assigned, create second riwayat entry
	if initialStatus == "ditugaskan" {
		riwayatAssign := model.RiwayatStatus{
			LaporanID:  laporan.ID,
			Status:     "ditugaskan",
			Keterangan: "Otomatis ditugaskan ke petugas kategori",
		}
		repository.CreateRiwayat(&riwayatAssign)
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
	id, err := strconv.Atoi(c.Params("id"))
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

	// Validate status
	validStatuses := map[string]bool{
		"dilaporkan": true,
		"ditugaskan": true,
		"dikerjakan": true,
		"selesai":    true,
	}
	if !validStatuses[req.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "status tidak valid. Gunakan: dilaporkan, ditugaskan, dikerjakan, atau selesai",
		})
	}

	laporan, err := repository.FindLaporanByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.Response{
			Message: "laporan tidak ditemukan",
			Error:   err.Error(),
		})
	}

	laporan.Status = req.Status

	// If status is "selesai", set tanggal_selesai and handle bukti_selesai upload
	if req.Status == "selesai" {
		now := time.Now()
		laporan.TanggalSelesai = &now

		file, err := c.FormFile("bukti_selesai")
		if err != nil { 
			return c.Status(fiber.StatusBadRequest).JSON(model.Response{
				Message: "bukti penyelesaian (foto/video) wajib diunggah",
				Error:   err.Error(),
			})
		}

		url, uploadErr := storage.UploadToSupabase(file)
		if uploadErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
				Message: "gagal mengunggah bukti selesai",
				Error:   uploadErr.Error(),
			})
		}
		laporan.BuktiSelesai = url
	}

	if err := repository.UpdateLaporan(laporan); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal mengubah status laporan",
			Error:   err.Error(),
		})
	}

	// Create riwayat entry
	keterangan := "Status diubah menjadi " + req.Status
	if req.Status == "selesai" {
		keterangan = "Laporan telah selesai ditangani"
	}

	riwayat := model.RiwayatStatus{
		LaporanID:  laporan.ID,
		Status:     req.Status,
		Keterangan: keterangan,
	}
	repository.CreateRiwayat(&riwayat)

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
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "ID tidak valid",
			Error:   err.Error(),
		})
	}

	_, err = repository.FindLaporanByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.Response{
			Message: "laporan tidak ditemukan",
			Error:   err.Error(),
		})
	}

	if err := repository.DeleteLaporan(uint(id)); err != nil {
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
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "ID tidak valid",
			Error:   err.Error(),
		})
	}

	// Verify laporan exists
	_, err = repository.FindLaporanByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.Response{
			Message: "laporan tidak ditemukan",
			Error:   err.Error(),
		})
	}

	riwayats, err := repository.FindRiwayatByLaporanID(uint(id))
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

// UpdateRating memproses pemberian rating dan feedback oleh mahasiswa
func UpdateRating(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{Message: "ID tidak valid"})
	}

	var req model.RatingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{Message: "Data tidak valid"})
	}

	laporan, err := repository.FindLaporanByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.Response{Message: "Laporan tidak ditemukan"})
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

	return c.JSON(model.Response{
		Message: "Rating berhasil disimpan",
	})
}

// AdminUpdateLaporan memproses update prioritas dan penugasan oleh Admin
func AdminUpdateLaporan(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{Message: "ID tidak valid"})
	}

	var req model.AdminUpdateLaporanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{Message: "Data tidak valid"})
	}

	laporan, err := repository.FindLaporanByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.Response{Message: "Laporan tidak ditemukan"})
	}

	// Update Prioritas
	if req.Prioritas != "" {
		if req.Prioritas != laporan.Prioritas {
			repository.CreateRiwayat(&model.RiwayatStatus{
				LaporanID:  laporan.ID,
				Status:     laporan.Status,
				Keterangan: "Prioritas diubah menjadi " + req.Prioritas + " oleh Admin",
			})
		}
		laporan.Prioritas = req.Prioritas
	}

	// Update PetugasID
	if req.PetugasID != nil {
		if *req.PetugasID == 0 {
			laporan.PetugasID = nil
		} else {
			laporan.PetugasID = req.PetugasID
		}
	}

	// Update Tenggat Waktu
	if req.TenggatWaktu != nil {
		laporan.TenggatWaktu = req.TenggatWaktu
	} else {
        // If request sends null (meaning unset the deadline)
        // Check if the JSON field was actually provided or omitted?
        // We'll trust the frontend. If it sends null, we unset it.
        // But with Go, a nil pointer means it could be unset in JSON.
        // We will assume the frontend only sends it if it wants to change it.
        // Wait, the frontend payload sends `tenggat_waktu: null` to unset.
    }
    
    // To properly unset, let's just use what's in req.
    laporan.TenggatWaktu = req.TenggatWaktu

	if err := repository.UpdateLaporan(laporan); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{Message: "Gagal mengupdate laporan"})
	}

	return c.JSON(model.Response{
		Message: "Laporan berhasil diupdate oleh Admin",
	})
}
