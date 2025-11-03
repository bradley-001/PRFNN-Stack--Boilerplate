package config

import (
	"fmt"

	"api/src/lib/general"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectToDatabase() {
	// Load in enviornment variables

	err := godotenv.Load(".env")
	if err != nil {
		Log("Could not find .env file", 3, true, false)
		return
	}

	pgAddr := general.GetEnv("POSTGRES_ADDRESS", "localhost")
	pgUser := general.GetEnv("POSTGRES_USER", "dev")
	pgPasswd := general.GetEnv("POSTGRES_PASSWORD", "")
	pgDB := general.GetEnv("POSTGRES_DB", "postgres")
	pgPort := general.GetEnv("POSTGRES_PORT", 5432)
	pgSSL := general.GetEnv("SSLMODE", "disable")

	dbConnectionString := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		pgAddr,
		pgUser,
		pgPasswd,
		pgDB,
		pgPort,
		pgSSL,
	)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Database connection
	DB, err = gorm.Open(postgres.Open(dbConnectionString), gormConfig)
	if err != nil {
		Log("Could not open connection to Postgres", 3, true, false)
		return
	}

	notice := fmt.Sprintf("Database connection successful to %s:%d | via user (%s)",
		pgAddr, pgPort, pgUser,
	)
	Log(notice, 1, false, false)
}
