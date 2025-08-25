package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)


var unprotectedRoutes = []string{
	"/api/health",
	"/api/register",
	"/api/login",
}

// skip unprotected routes

func IsUnprotectedRoute(c *fiber.Ctx) bool {
	for _, route := range unprotectedRoutes {
		if c.Path() == route {
			return true
		}
	}
	return false
}

func CheckAuth() fiber.Handler {
	return func (c *fiber.Ctx) error {

		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}

		if IsUnprotectedRoute(c) {
			return c.Next()
		}

		authHeader := c.Get("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header is missing or invalid",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ") // Extract the token string from the header

		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token is missing",
			})
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token signing method")
			}

			return []byte(os.Getenv("JWT_SECRET_KEY")), nil // Use the secret key from environment variables
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID, ok := claims["userId"].(string)
			if !ok {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid token claims",
				})
			}

			c.Locals("userId", userID) // Store user ID in context for later use
			return c.Next()
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}
}