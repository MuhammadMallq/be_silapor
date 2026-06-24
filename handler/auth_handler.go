package handler

import (
	"os"
	"time"

	"be_silapor/model"
	"be_silapor/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Register godoc
// @Summary Register user baru
// @Description Mendaftarkan user baru ke sistem
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body model.RegisterRequest true "Data registrasi"
// @Success 201 {object} model.Response
// @Failure 400 {object} model.Response
// @Failure 409 {object} model.Response
// @Router /register [post]
func Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "payload tidak valid",
			Error:   err.Error(),
		})
	}

	// Validate required fields
	if req.Nama == "" || req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "Nama, username, dan password wajib diisi",
		})
	}

	if len(req.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "Password minimal 6 karakter",
		})
	}

	// Check if username already exists
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
			Message: "gagal membuat hash password",
			Error:   err.Error(),
		})
	}

	// Set default role
	role := "mahasiswa"
	if req.Role == "petugas" || req.Role == "admin" {
		role = req.Role
	}

	user := model.User{
		Nama:     req.Nama,
		Username: req.Username,
		Password: string(hashedPassword),
		Role:     role,
	}

	if err := repository.CreateUser(&user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal mendaftarkan user",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(model.Response{
		Message: "register berhasil",
		Data:    user,
	})
}

// Login godoc
// @Summary Login user
// @Description Login dan mendapatkan JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body model.LoginRequest true "Data login"
// @Success 200 {object} model.Response
// @Failure 400 {object} model.Response
// @Failure 401 {object} model.Response
// @Router /login [post]
func Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "payload tidak valid",
			Error:   err.Error(),
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "username dan password wajib diisi",
		})
	}

	// Find user
	user, err := repository.FindUserByUsername(req.Username)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.Response{
			Message: "username atau password salah",
		})
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.Response{
			Message: "username atau password salah",
		})
	}

	// Generate JWT
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal membuat token",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.Response{
		Message: "login berhasil",
		Data: model.LoginResponse{
			Token: tokenString,
			User:  *user,
		},
	})
}

// ChangePassword godoc
// @Summary Ubah password
// @Description Mengubah password user yang sedang login
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body model.ChangePasswordRequest true "Data password"
// @Security BearerAuth
// @Success 200 {object} model.Response
// @Failure 400 {object} model.Response
// @Failure 401 {object} model.Response
// @Router /changepassword [put]
func ChangePassword(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var req model.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "payload tidak valid",
			Error:   err.Error(),
		})
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "password lama dan password baru wajib diisi",
		})
	}

	if len(req.NewPassword) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "password baru minimal 6 karakter",
		})
	}

	user, err := repository.FindUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.Response{
			Message: "gagal mencari user",
			Error:   err.Error(),
		})
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.Response{
			Message: "password lama salah",
		})
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal membuat hash password",
			Error:   err.Error(),
		})
	}

	user.Password = string(hashedPassword)
	if err := repository.UpdateUser(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal mengubah password",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.Response{
		Message: "password berhasil diubah",
	})
}
