package handler

import (
	"strconv"

	"be_silapor/model"
	"be_silapor/repository"

	"github.com/gofiber/fiber/v2"
)

// GetNotifikasi godoc
// @Summary Daftar Notifikasi
// @Description Mengambil daftar notifikasi untuk pengguna yang sedang login
// @Tags Notifikasi
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.Response
// @Router /api/notifikasi [get]
func GetNotifikasi(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	notifikasis, err := repository.GetNotifikasiByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "Gagal mengambil notifikasi",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.Response{
		Message: "Berhasil mengambil notifikasi",
		Data:    notifikasis,
	})
}

// ReadNotifikasi godoc
// @Summary Tandai notifikasi dibaca
// @Description Menandai satu notifikasi spesifik sebagai sudah dibaca
// @Tags Notifikasi
// @Produce json
// @Param id path int true "Notifikasi ID"
// @Security BearerAuth
// @Success 200 {object} model.Response
// @Router /api/notifikasi/{id}/read [put]
func ReadNotifikasi(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	notifIDStr := c.Params("id")
	notifID, err := strconv.Atoi(notifIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "ID notifikasi tidak valid",
		})
	}

	err = repository.MarkAsRead(uint(notifID), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "Gagal menandai notifikasi",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.Response{
		Message: "Notifikasi ditandai sudah dibaca",
	})
}

// ReadAllNotifikasi godoc
// @Summary Tandai semua notifikasi dibaca
// @Description Menandai seluruh notifikasi milik pengguna sebagai sudah dibaca
// @Tags Notifikasi
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.Response
// @Router /api/notifikasi/read-all [put]
func ReadAllNotifikasi(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	err := repository.MarkAllAsRead(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "Gagal menandai semua notifikasi",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.Response{
		Message: "Semua notifikasi ditandai sudah dibaca",
	})
}
