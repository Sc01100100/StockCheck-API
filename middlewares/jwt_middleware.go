package middlewares

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/Sc01100100/SaveCash-API/utils"
)

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			log.Println("Authorization token is missing")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Authorization token is missing",
			})
		}

		token = strings.Replace(token, "Bearer ", "", 1)

		userID, userRole, err := utils.ValidateJWT(token)
		if err != nil {
			log.Printf("Token validation error: %v\n", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid token",
			})
		}

		log.Printf("Middleware extracted UserID: %d, Role: %s\n", userID, userRole)

		if userID == 0 {
			log.Println("Extracted UserID is 0, invalid token")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid UserID in token",
			})
		}

		c.Locals("user_id", userID)
		c.Locals("role", userRole)

		return c.Next()
	}
}

func AdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role") 

		if role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "error",
				"message": "Access denied: admin role required",
			})
		}

		return c.Next() 
	}
}