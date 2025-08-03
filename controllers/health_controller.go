package controllers

import (
	"github.com/gofiber/fiber/v2"
)

// HealthCheck responds with a simple status message to indicate the service is running

func HealthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "up",
	})
}