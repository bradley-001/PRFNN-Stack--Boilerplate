package handlers

import (
	"errors"
	"fmt"
	"time"

	"api/src/config"
	"api/src/constants"
	"api/src/lib/general"
	"api/src/lib/security"
	"api/src/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var httpsOn = general.GetEnv("NODE_ENV", "") == "production"

// - /auth/register
// No user attached to this request, this is a non authenticated route.
func PostRegister(c *fiber.Ctx) error {
	type RegistrationSchema struct {
		Username    string `json:"username" validate:"required,min=3,max=50"`
		RawPassword string `json:"raw_password" validate:"required,min=8"`
	}

	var data RegistrationSchema
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if username is bewteen 3 & 50 characters
	if len(data.Username) < 3 || len(data.Username) > 50 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username must be between 3 and 50 characters",
		})
	}

	// Check that password is mor than 8 characters
	if len(data.RawPassword) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must be at least 8 characters",
		})
	}

	// Start concurrent password hashing process
	type hashRes struct {
		hash string
		err  error
	}

	hashConc := make(chan hashRes, 1)

	go func() {
		hash, err := security.HashBcrypt(data.RawPassword)
		hashConc <- hashRes{hash: hash, err: err}
	}()

	// Check if user already exists in database -----------------
	var existingUser models.Users

	if err := config.DB.First(&existingUser, "username = ?", data.Username).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Username already taken",
		})
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		config.Log("Database error", 2, false, false)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	// Await Password Hashing
	hashResult := <-hashConc
	if hashResult.err != nil {
		config.Log(fmt.Sprintf("Hashing Process Failure: %s", hashResult.err), 3, false, false)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Hashing process failure",
		})
	}

	// Database transaction for user & session
	var user models.Users
	var session models.Sessions
	var token string

	expirtyDateTime := time.Now().Add(constants.SESSION_DURATION)

	if err := config.DB.Transaction(func(tx *gorm.DB) error {
		// Create user record
		user = models.Users{
			Username: data.Username,
			Password: hashResult.hash,
		}

		if err := tx.Create(&user).Error; err != nil {
			return err // Transaction rollback
		}

		// Create session (long-lived) record
		session = models.Sessions{
			UserId:    user.Id,
			ExpiresAt: expirtyDateTime,
		}

		if err := tx.Create(&session).Error; err != nil {
			return err // Transaction rollback
		}

		// Generate JWT token
		if newToken, err := security.GenerateJWT(user.Id, session.Id); err != nil {
			return err
		} else {
			token = newToken
		}

		return nil // Commit transaction

	}); err != nil {
		config.Log("Could not create user and associated session during registration - database transaction failed.", 1, false, true)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Account registration failure",
		})
	}

	// Append JWT cookie to response header
	c.Cookie(&fiber.Cookie{
		Name:     "jwt_token",
		Value:    token,
		Expires:  expirtyDateTime,
		HTTPOnly: true,
		Secure:   httpsOn,
		SameSite: "Strict",
		Path:     "/",
	})

	// Unauthorised Route Specific Security Headers
	c.Set("Cache-Control", "no-store, no-cache, must-revalidate")
	c.Set("Pragma", "no-cache")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":              user.Id,
		"username":        user.Username,
		"is_verified":     user.IsVerified,
		"created_at":      user.CreatedAt,
		"last_updated_at": user.LastUpdatedAt,
	})
}

// - /auth/login
// No user attached to this request, this is a non authenticated route.
func PostLogin(c *fiber.Ctx) error {
	type LoginSchema struct {
		Username    string `json:"username"`
		RawPassword string `json:"raw_password"`
	}

	var data LoginSchema
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Input validation for username
	if len(data.Username) < 3 || len(data.Username) > 50 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "A standard username is between 3 and 50 characters",
		})
	}

	// Input validation for password
	if len(data.RawPassword) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "A standard password is at least 8 characters",
		})
	}

	// Get user from database
	var user models.Users
	if err := config.DB.First(&user, "username = ?", data.Username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid username or password",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	// Verify password (bcrypt handles timing-safe comparison)
	if valid, err := security.CheckHashBcrypt(data.RawPassword, user.Password); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Authentication failed",
		})
	} else if !valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid username or password",
		})
	}

	// Create session and generate JWT
	var session models.Sessions
	var token string

	if err := config.DB.Transaction(func(tx *gorm.DB) error {
		// Create session record
		session = models.Sessions{
			UserId:    user.Id,
			ExpiresAt: time.Now().Add(constants.SESSION_DURATION),
		}

		if err := tx.Create(&session).Error; err != nil {
			return err
		}

		// Generate JWT token
		if newToken, err := security.GenerateJWT(user.Id, session.Id); err != nil {
			return err
		} else {
			token = newToken
		}

		return nil

	}); err != nil {
		fmt.Printf("[ERROR] Login transaction failed for user %s: %v\n", data.Username, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Login failed",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "jwt_token",
		Value:    token,
		Expires:  time.Now().Add(constants.SESSION_DURATION),
		HTTPOnly: true,
		Secure:   httpsOn,
		SameSite: "Strict",
		Path:     "/",
	})

	// Unauthorized Route Specific Security Headers
	c.Set("Cache-Control", "no-store, no-cache, must-revalidate")
	c.Set("Pragma", "no-cache")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id":              user.Id,
		"username":        user.Username,
		"is_verified":     user.IsVerified,
		"created_at":      user.CreatedAt,
		"last_updated_at": user.LastUpdatedAt,
	})
}

// - /auth/logout
func DeleteLogout(c *fiber.Ctx) error {
	session, err := general.GetReqSession(c)
	if err != nil {
		return err
	}

	if err := config.DB.Delete(&session).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "jwt_token",
		Value:    "",
		Expires:  time.Now().Add(-(5 * time.Minute)), // In the past.
		HTTPOnly: true,
		Secure:   httpsOn,
		SameSite: "Strict",
		Path:     "/",
	})

	// Unauthorized Route Specific Security Headers
	c.Set("Cache-Control", "no-store, no-cache, must-revalidate")
	c.Set("Pragma", "no-cache")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Successfully logged out & revoked session",
	})

}
