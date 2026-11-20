package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strings"
	"time"

	"be_silapor/model"
	"be_silapor/pkg/email"
	"be_silapor/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ValidatePasswordComplexity memastikan password memenuhi kriteria kuat:
// minimal 8 karakter, huruf besar, huruf kecil, angka, dan karakter khusus
func ValidatePasswordComplexity(password string) error {
	if len(password) < 8 {
		return errors.New("password minimal 8 karakter")
	}
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return errors.New("password harus mengandung minimal 1 huruf besar (A-Z)")
	}
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return errors.New("password harus mengandung minimal 1 huruf kecil (a-z)")
	}
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return errors.New("password harus mengandung minimal 1 angka (0-9)")
	}
	if !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]`).MatchString(password) {
		return errors.New("password harus mengandung minimal 1 karakter khusus / simbol")
	}
	return nil
}

// maskEmail menyamarkan email untuk privasi (contoh: r***i@gmail.com)
func maskEmail(em string) string {
	parts := strings.Split(em, "@")
	if len(parts) != 2 {
		return em
	}
	name := parts[0]
	domain := parts[1]
	if len(name) <= 2 {
		return name[:1] + "***@" + domain
	}
	return name[:1] + "***" + name[len(name)-1:] + "@" + domain
}

// Register godoc
func Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "payload tidak valid",
			Error:   err.Error(),
		})
	}

	// Validate required fields
	if strings.TrimSpace(req.Nama) == "" || strings.TrimSpace(req.Username) == "" || req.Password == "" {
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

	// Check if username already exists
	existingUser, _ := repository.FindUserByUsername(req.Username)
	if existingUser != nil {
		return c.Status(fiber.StatusConflict).JSON(model.Response{
			Message: "username sudah digunakan",
		})
	}

	// Check email jika diisi
	req.Email = strings.TrimSpace(req.Email)
	if req.Email != "" {
		existingEmailUser, _ := repository.FindUserByIdentifier(req.Email)
		if existingEmailUser != nil {
			return c.Status(fiber.StatusConflict).JSON(model.Response{
				Message: "email sudah digunakan oleh akun lain",
			})
		}
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

	user := model.User{
		Nama:     req.Nama,
		Username: req.Username,
		Email:    req.Email,
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

	// Validasi password kuat
	if err := ValidatePasswordComplexity(req.NewPassword); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: err.Error(),
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

// UpdateProfileEmail untuk menambahkan atau mengubah email pemulihan
func UpdateProfileEmail(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var req model.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "payload tidak valid",
			Error:   err.Error(),
		})
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "email tidak boleh kosong",
		})
	}

	// Pastikan email belum dipakai akun lain
	existingUser, _ := repository.FindUserByIdentifier(req.Email)
	if existingUser != nil && existingUser.ID != userID {
		return c.Status(fiber.StatusConflict).JSON(model.Response{
			Message: "email sudah digunakan oleh akun lain",
		})
	}

	if err := repository.UpdateUserEmail(userID, req.Email); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal mengupdate email pemulihan",
			Error:   err.Error(),
		})
	}

	user, _ := repository.FindUserByID(userID)

	return c.JSON(model.Response{
		Message: "email pemulihan berhasil diperbarui",
		Data:    user,
	})
}

// ForgotPassword mengawali proses reset password dengan verifikasi dan pengiriman OTP
func ForgotPassword(c *fiber.Ctx) error {
	var req model.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "payload tidak valid",
			Error:   err.Error(),
		})
	}

	req.Identifier = strings.TrimSpace(req.Identifier)
	if req.Identifier == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "masukkan username atau email Anda",
		})
	}

	user, err := repository.FindUserByIdentifier(req.Identifier)
	if err != nil || user == nil {
		return c.Status(fiber.StatusNotFound).JSON(model.Response{
			Message: "pengguna dengan username atau email tersebut tidak ditemukan",
		})
	}

	// Admin dikecualikan dari fitur lupa password via email
	if user.Role == "admin" {
		return c.Status(fiber.StatusForbidden).JSON(model.Response{
			Message: "Akun Admin tidak dapat menggunakan fitur lupa password via email. Silakan hubungi superadmin.",
		})
	}

	// Cek apakah akun sudah mendaftarkan email
	if strings.TrimSpace(user.Email) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "Akun Anda belum memiliki email pemulihan terdaftar. Silakan hubungi Admin Kampus untuk melakukan reset password.",
		})
	}

	// Batalkan token-token lama
	_ = repository.InvalidateOldTokensForUser(user.ID)

	// Buat 6-digit numeric OTP menggunakan crypto/rand
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal membuat kode OTP",
		})
	}
	otp := fmt.Sprintf("%06d", n.Int64()+100000)

	// Hash token OTP menggunakan SHA256
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s:%s", user.Username, otp)))
	tokenHash := hex.EncodeToString(h.Sum(nil))

	resetToken := model.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Used:      false,
	}

	if err := repository.CreateResetToken(&resetToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal menyimpan token reset password",
			Error:   err.Error(),
		})
	}

	// Kirim email
	go func() {
		_ = email.SendOTP(user.Email, user.Nama, otp)
	}()

	return c.JSON(model.Response{
		Message: fmt.Sprintf("Kode OTP telah dikirimkan ke email %s. Silakan periksa kotak masuk atau spam.", maskEmail(user.Email)),
		Data: fiber.Map{
			"email": maskEmail(user.Email),
		},
	})
}

// VerifyOTP memverifikasi kecocokan OTP dan mengembalikan token sementara untuk reset password
func VerifyOTP(c *fiber.Ctx) error {
	var req model.VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "payload tidak valid",
			Error:   err.Error(),
		})
	}

	req.Identifier = strings.TrimSpace(req.Identifier)
	req.OTP = strings.TrimSpace(req.OTP)

	if req.Identifier == "" || req.OTP == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "identifier dan kode OTP wajib diisi",
		})
	}

	user, err := repository.FindUserByIdentifier(req.Identifier)
	if err != nil || user == nil {
		return c.Status(fiber.StatusNotFound).JSON(model.Response{
			Message: "pengguna tidak ditemukan",
		})
	}

	// Hitung token hash
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s:%s", user.Username, req.OTP)))
	tokenHash := hex.EncodeToString(h.Sum(nil))

	tokenRecord, err := repository.FindValidResetToken(tokenHash)
	if err != nil || tokenRecord == nil || tokenRecord.UserID != user.ID {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "kode OTP salah atau sudah kedaluwarsa",
		})
	}

	// Tandai token OTP ini sudah terpakai
	_ = repository.MarkTokenAsUsed(tokenRecord.ID)

	// Buat token sementara (JWT 15 menit) untuk melakukan reset password
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"type":    "reset_password_authorized",
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	tempToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal membuat sesi reset password",
		})
	}

	return c.JSON(model.Response{
		Message: "verifikasi OTP berhasil",
		Data: fiber.Map{
			"reset_token": tempToken,
		},
	})
}

// ResetPassword mereset password baru menggunakan reset token yang sudah terverifikasi
func ResetPassword(c *fiber.Ctx) error {
	var req model.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "payload tidak valid",
			Error:   err.Error(),
		})
	}

	if req.ResetToken == "" || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: "token dan password baru wajib diisi",
		})
	}

	// Validasi password kuat
	if err := ValidatePasswordComplexity(req.NewPassword); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Message: err.Error(),
		})
	}

	// Validasi JWT reset token
	secret := os.Getenv("JWT_SECRET")
	parsedToken, err := jwt.Parse(req.ResetToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("metode signing tidak valid")
		}
		return []byte(secret), nil
	})

	if err != nil || !parsedToken.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(model.Response{
			Message: "sesi reset password tidak valid atau sudah kedaluwarsa, silakan ulangi proses lupa password",
		})
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || claims["type"] != "reset_password_authorized" {
		return c.Status(fiber.StatusUnauthorized).JSON(model.Response{
			Message: "token tidak valid untuk reset password",
		})
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(model.Response{
			Message: "identitas user tidak valid dalam token",
		})
	}
	userID := uint(userIDFloat)

	// Hash password baru
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal mengenkripsi password",
		})
	}

	if err := repository.UpdateUserPassword(userID, string(hashedPassword)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Message: "gagal menyimpan password baru",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.Response{
		Message: "password berhasil direset! Silakan login dengan kata sandi baru Anda.",
	})
}
