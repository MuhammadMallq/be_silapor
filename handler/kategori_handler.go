package handler

import (
	"be_silapor/model"
	"be_silapor/repository"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// GetAllKategori godoc
// @Summary Daftar kategori fasilitas
// @Description Mengambil semua kategori fasilitas
// @Tags Kategori
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.Response
// @Router /api/kategori [get]
func GetAllKategori(c *fiber.Ctx) error {
	kategoris, err := repository.FindAllKategori()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal mengambil data kategori",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.Response{
		Message: "berhasil mengambil data kategori",
		Data:    kategoris,
	})
}

// CreateKategori godoc
// @Summary Tambah kategori baru
// @Description Admin menambah kategori fasilitas baru
// @Tags Kategori
// @Accept json
// @Produce json
// @Param body body model.KategoriRequest true "Data kategori"
// @Security BearerAuth
// @Success 201 {object} model.Response
// @Failure 400 {object} model.Response
// @Router /api/kategori [post]
func CreateKategori(c *fiber.Ctx) error {
	var req model.KategoriRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "payload tidak valid",
			Error:   err.Error(),
		})
	}

	if req.NamaKategori == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "nama kategori wajib diisi",
		})
	}

	if req.SLAJam == 0 {
		req.SLAJam = 48 // Default SLA
	}

	kategori := model.KategoriFasilitas{
		NamaKategori: req.NamaKategori,
		PetugasID:    req.PetugasID,
		SLAJam:       req.SLAJam,
	}

	if err := repository.CreateKategori(&kategori); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal membuat kategori",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(model.Response{
		Message: "kategori berhasil dibuat",
		Data:    kategori,
	})
}

// UpdateKategori godoc
// @Summary Ubah kategori
// @Description Admin mengubah data kategori fasilitas
// @Tags Kategori
// @Accept json
// @Produce json
// @Param id path int true "Kategori ID"
// @Param body body model.KategoriRequest true "Data kategori"
// @Security BearerAuth
// @Success 200 {object} model.Response
// @Failure 400 {object} model.Response
// @Failure 404 {object} model.Response
// @Router /api/kategori/{id} [put]
func UpdateKategori(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "ID tidak valid",
			Error:   err.Error(),
		})
	}

	kategori, err := repository.FindKategoriByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.Response{
			Message: "kategori tidak ditemukan",
			Error:   err.Error(),
		})
	}

	var req model.KategoriRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "payload tidak valid",
			Error:   err.Error(),
		})
	}

	if req.NamaKategori != "" {
		kategori.NamaKategori = req.NamaKategori
	}
	if req.PetugasID != nil {
		kategori.PetugasID = req.PetugasID
	}
	if req.SLAJam > 0 {
		kategori.SLAJam = req.SLAJam
	}

	if err := repository.UpdateKategori(kategori); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal mengubah kategori",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.Response{
		Message: "kategori berhasil diubah",
		Data:    kategori,
	})
}

// DeleteKategori godoc
// @Summary Hapus kategori
// @Description Admin menghapus kategori fasilitas
// @Tags Kategori
// @Produce json
// @Param id path int true "Kategori ID"
// @Security BearerAuth
// @Success 200 {object} model.Response
// @Failure 404 {object} model.Response
// @Router /api/kategori/{id} [delete]
func DeleteKategori(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "ID tidak valid",
			Error:   err.Error(),
		})
	}

	_, err = repository.FindKategoriByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.Response{
			Message: "kategori tidak ditemukan",
			Error:   err.Error(),
		})
	}

	if err := repository.DeleteKategori(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal menghapus kategori",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.Response{
		Message: "kategori berhasil dihapus",
	})
}
