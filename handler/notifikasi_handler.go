package handler

import (
	"strconv"

	"be_silapor/model"
	"be_silapor/repository"

	"github.com/gofiber/fiber/v2"
)

// GetNotifikasi mengambil notifikasi untuk pengguna yang sedang login
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

// ReadNotifikasi menandai satu notifikasi sebagai sudah dibaca
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

// ReadAllNotifikasi menandai semua notifikasi pengguna sebagai sudah dibaca
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
