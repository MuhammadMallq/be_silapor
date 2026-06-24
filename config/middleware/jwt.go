package middleware

import (
	"os"
	"strings"

	"be_silapor/model"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWTProtected adalah middleware untuk memproteksi endpoint
// Cara pakai:
//   - middleware.JWTProtected()           → hanya cek token valid, semua role boleh akses
//   - middleware.JWTProtected("admin")    → hanya role admin yang boleh akses
//   - middleware.JWTProtected("admin", "petugas") → admin atau petugas yang boleh akses
func JWTProtected(allowedRoles ...string) fiber.Handler {
	// Middleware Fiber selalu mengembalikan func(c *fiber.Ctx) error
	return func(c *fiber.Ctx) error {
		// Ambil header Authorization dari request
		// Format yang benar: "Bearer eyJhbGciOi..."
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.Response{
				Message: "Token tidak ditemukan",
			})
		}

		// Pisahkan "Bearer" dari token-nya
		// tokenParts[0] = "Bearer", tokenParts[1] = token JWT
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.Response{
				Message: "Format token tidak valid",
			})
		}

		tokenString := tokenParts[1]
		// Ambil secret key dari .env untuk memverifikasi tanda tangan token
		secret := os.Getenv("JWT_SECRET")

		// Parse dan verifikasi token JWT
		// Fungsi di dalamnya memastikan token ditandatangani dengan algoritma HMAC (HS256)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Metode signing tidak valid")
			}
			return []byte(secret), nil
		})

		// Jika token tidak valid atau sudah expired, tolak request
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(model.Response{
				Message: "Token tidak valid atau sudah expired",
			})
		}

		// Ambil data (claims) yang tersimpan di dalam token JWT
		// Claims berisi: user_id, username, role, exp (waktu kedaluwarsa)
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(model.Response{
				Message: "Gagal membaca claims token",
			})
		}

		// Simpan user_id dan role ke dalam context Fiber
		// Data ini bisa diambil di handler dengan: c.Locals("user_id") dan c.Locals("role")
		userID := uint(claims["user_id"].(float64)) // JWT menyimpan angka sebagai float64
		userRole := claims["role"].(string)

		c.Locals("user_id", userID)
		c.Locals("role", userRole)

		// Jika middleware dipanggil dengan daftar role (misal: "admin", "petugas"),
		// cek apakah role user termasuk dalam daftar yang diizinkan
		if len(allowedRoles) > 0 {
			roleAllowed := false
			for _, role := range allowedRoles {
				if userRole == role {
					roleAllowed = true
					break
				}
			}
			// Jika role tidak cocok, kembalikan 403 Forbidden
			if !roleAllowed {
				return c.Status(fiber.StatusForbidden).JSON(model.Response{
					Message: "Anda tidak memiliki akses ke resource ini",
				})
			}
		}

		// Jika semua pengecekan lolos, lanjutkan ke handler berikutnya
		return c.Next()
	}
}
