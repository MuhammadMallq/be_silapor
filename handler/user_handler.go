package handler

import (
	"be_silapor/model"
	"be_silapor/repository"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// GetAllUsers godoc
func GetAllUsers(c *fiber.Ctx) error {
	users, err := repository.FindAllUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal mengambil data user",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.Response{
		Message: "berhasil mengambil data user",
		Data:    users,
	})
}

// CreateUserByAdmin godoc
func CreateUserByAdmin(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "payload tidak valid",
			Error:   err.Error(),
		})
	}

	// Cek apakah username sudah dipakai
	existingUser, _ := repository.FindUserByUsername(req.Username)
	if existingUser != nil {
		return c.Status(fiber.StatusConflict).JSON(model.Response{
			Message: "username sudah digunakan",
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal memproses password",
			Error:   err.Error(),
		})
	}

	if req.Role == "" {
		req.Role = "mahasiswa" // default
	}

	user := model.User{
		Nama:     req.Nama,
		Username: req.Username,
		Password: string(hashedPassword),
		Role:     req.Role,
	}

	if err := repository.CreateUser(&user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal menyimpan user",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(model.Response{
		Message: "user berhasil dibuat",
		Data:    user,
	})
}

// DeleteUser godoc
func DeleteUser(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "ID tidak valid",
		})
	}

	if err := repository.DeleteUser(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal menghapus user",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.Response{
		Message: "user berhasil dihapus",
	})
}
