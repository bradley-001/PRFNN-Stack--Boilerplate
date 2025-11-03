package routes

import (
	"api/src/handlers"
	"api/src/lib/general"
	"api/src/middleware"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

var apiVersion = fmt.Sprintf("v%s", general.GetEnv("VERSION", "0"))

func SetupRoutes(app *fiber.App) {
	// Apply pre-request handling (i.e. IP rate limiting)
	app.Use(middleware.PreRequest())

	// Apply common security headers to all requests
	app.Use(middleware.SecurityHeaders())

	// Get API version from enviornment and apply to main route.
	apiBase := app.Group(fmt.Sprintf("/api/%s", apiVersion))

	// Seperate API into public and private segments, public avoids main middleware controls.
	apiBasePublic := apiBase.Group("/public")
	apiBasePrivate := apiBase.Group("/private", middleware.CoreMiddleware())

	// --- Main route setup ---

	// Auth routes (Public) ---
	apiBasePublic.Post("/auth/register", handlers.PostRegister)
	apiBasePublic.Post("/auth/login", handlers.PostLogin)

	// Auth routes (Private) ---
	apiBasePrivate.Delete("/auth/logout", handlers.DeleteLogout)

	// Users routes (Private) ---
	usersGroup := apiBasePrivate.Group("/users")

	usersGroup.Get("/me", handlers.GetMe) // -> Note: By default a user can only make requests regarding user data on their own data.
	usersGroup.Patch("/me", handlers.PatchMe)
	usersGroup.Delete("/me", handlers.DeleteMe)

}
