package handler

import (
	"be_silapor/model"
	"be_silapor/repository"
	"strconv"
	"strings"

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

	req.Nama = strings.TrimSpace(req.Nama)
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if req.Nama == "" || req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "Nama, username, dan password wajib diisi",
		})
	}

	// Validasi password kuat
	if err := ValidatePasswordComplexity(req.Password); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: err.Error(),
		})
	}

	// Cek apakah username sudah dipakai
	existingUser, _ := repository.FindUserByUsername(req.Username)
	if existingUser != nil {
		return c.Status(fiber.StatusConflict).JSON(model.Response{
			Message: "username sudah digunakan",
		})
	}

	// Cek apakah email sudah dipakai
	if req.Email != "" {
		existingEmailUser, _ := repository.FindUserByIdentifier(req.Email)
		if existingEmailUser != nil {
			return c.Status(fiber.StatusConflict).JSON(model.Response{
				Message: "email sudah digunakan oleh pengguna lain",
			})
		}
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
		req.Role = "petugas" // default untuk admin create user
	}

	// Validasi role
	validRoles := map[string]bool{
		"admin":     true,
		"petugas":   true,
		"dosen":     true,
		"staff":     true,
		"mahasiswa": true,
	}
	if !validRoles[req.Role] {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "role tidak valid. Gunakan: admin, petugas, dosen, staff, atau mahasiswa",
		})
	}

	user := model.User{
		Nama:     req.Nama,
		Username: req.Username,
		Email:    req.Email,
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

// ResetPasswordByAdmin memungkinkan admin mereset password pengguna manual jika pengguna lupa
func ResetPasswordByAdmin(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "ID tidak valid",
		})
	}

	var req model.AdminResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "payload tidak valid",
			Error:   err.Error(),
		})
	}

	// Validasi password kuat
	if err := ValidatePasswordComplexity(req.NewPassword); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: err.Error(),
		})
	}

	user, err := repository.FindUserByID(uint(id))
	if err != nil || user == nil {
		return c.Status(fiber.StatusNotFound).JSON(model.Response{
			Message: "pengguna tidak ditemukan",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal mengenkripsi password",
		})
	}

	if err := repository.UpdateUserPassword(user.ID, string(hashedPassword)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal mengubah password",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.Response{
		Message: "password pengguna berhasil direset oleh admin",
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
