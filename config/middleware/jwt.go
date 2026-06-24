package middleware

import (
	"os"
	"strings"

	"be_silapor/model"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWTProtected middleware checks for valid JWT token and role-based access
func JWTProtected(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.Response{
				Message: "Token tidak ditemukan",
			})
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.Response{
				Message: "Format token tidak valid",
			})
		}

		tokenString := tokenParts[1]
		secret := os.Getenv("JWT_SECRET")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Metode signing tidak valid")
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(model.Response{
				Message: "Token tidak valid atau sudah expired",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(model.Response{
				Message: "Gagal membaca claims token",
			})
		}

		// Store user info in context
		userID := uint(claims["user_id"].(float64))
		userRole := claims["role"].(string)

		c.Locals("user_id", userID)
		c.Locals("role", userRole)

		// Check role-based access if roles are specified
		if len(allowedRoles) > 0 {
			roleAllowed := false
			for _, role := range allowedRoles {
				if userRole == role {
					roleAllowed = true
					break
				}
			}
			if !roleAllowed {
				return c.Status(fiber.StatusForbidden).JSON(model.Response{
					Message: "Anda tidak memiliki akses ke resource ini",
				})
			}
		}

		return c.Next()
	}
}
