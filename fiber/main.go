package main

import (
	"fmt"

	"api/src/config"
	"api/src/lib/general"
	"api/src/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {

	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		config.Log(fmt.Sprintf("No .env file found in current directory: %v", err), 3, true, false)
	}

	// Enviornment loading
	nodeEnv := general.GetEnv("NODE_ENV", "development")
	hostingPort := general.GetEnv("PORT", 8080)
	apiVersion := general.GetEnv("VERSION", "n/a")
	frontendUrl := general.GetEnv("FRONTEND_URL", "http://localhost:3000")
	socketIoUrl := general.GetEnv("SOCKETIO_URL", "ws://localhost:4000")

	if nodeEnv == "development" {
		log.SetLevel(log.LevelTrace)
		config.Log("You are in development mode!", 0, false, false)
	} else {
		log.SetLevel(log.LevelInfo) // Changed from LevelWarn to LevelInfo so we can see INFO logs
		config.Log("You are in production mode!", 1, false, false)
	}

	config.ConnectToDatabase()
	config.ConnectToRedis()

	app := fiber.New(fiber.Config{
		Prefork:       nodeEnv == "production",
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Accord /w Fiber",
		AppName:       fmt.Sprintf("Accord API v%s", apiVersion),
	})

	// Middleware setup
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: fmt.Sprintf("%s, %s", frontendUrl, socketIoUrl),
	}))

	routes.SetupRoutes(app)

	config.Log(fmt.Sprintf("Server started on port %d", hostingPort), 1, false, false)

	if err := app.Listen(fmt.Sprintf(":%d", hostingPort)); err != nil {
		config.Log(fmt.Sprintf("Failed to start server: %v", err), 3, true, false)
	}

}
