package general

import (
	"api/src/models"

	"github.com/gofiber/fiber/v2"
)

func GetReqUser(c *fiber.Ctx) (*models.Users, error) {
	if data, ok := c.Locals("user").(models.Users); ok {
		return &data, nil
	}
	return nil, c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": "Could not parse user obj attached to request",
	})
}

func GetReqSession(c *fiber.Ctx) (*models.Sessions, error) {
	if data, ok := c.Locals("session").(models.Sessions); ok {
		return &data, nil
	}
	return nil, c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": "Could not parse session obj attached to request",
	})
}
