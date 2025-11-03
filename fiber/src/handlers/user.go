package handlers

import (
	"api/src/config"
	lib "api/src/lib/general"

	"github.com/gofiber/fiber/v2"
)

// - /users/me
func GetMe(c *fiber.Ctx) error {
	user, err := lib.GetReqUser(c)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

// - /users/me
func PatchMe(c *fiber.Ctx) error {
	user, err := lib.GetReqUser(c)
	if err != nil {
		return err
	}

	type UserPatchSchema struct {
		Username *string `json:"username"`
	}

	var data UserPatchSchema
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body, could not parse JSON",
		})
	}

	if data.Username != nil {
		user.Username = *data.Username
	}

	// Patch updated user
	if err := config.DB.Save(user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user",
		})
	}

	return c.Status(fiber.StatusOK).JSON(user)

}

// - /users/me
func DeleteMe(c *fiber.Ctx) error {
	user, err := lib.GetReqUser(c)
	if err != nil {
		return err
	}

	if err := config.DB.Delete(user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not delete user",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}
