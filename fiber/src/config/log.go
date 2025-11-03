package config

import (
	"errors"

	"api/src/lib/general"
	"api/src/models"

	"github.com/gofiber/fiber/v2/log"
)

var env = general.GetEnv("NODE_ENV", "development")

// @level: can be be between the ranges 0-3 inclusive, representing: DEBUG, NOTICE, WARNING & ERROR.
// Any log at level 0 will be ignored in produciton.
// @isFatal: if decalred as true it will return an error OR if the log level is 4 & the ENV is developement, it will OS exit instead.
// @saveLog: determines whether to log the output to the database
func Log(msg string, level uint8, isFatal bool, saveLog bool) error {

	if env == "" {
		return errors.New("no ENV found")
	}

	// Ignore DEBUG logs in produciton
	if env == "production" && level == 0 {
		return nil
	}

	if saveLog && DB != nil {
		logx := models.Logs{
			Message: msg,
			Level:   int16(level),
		}
		if err := DB.Create(&logx).Error; err != nil {
			log.Error("[ERROR] | Could not create database entry for log")
		}
	}

	switch level {
	case 0:
		log.Debug(msg)
	case 1:
		log.Info(msg)
	case 2:
		log.Warn(msg)
	case 3:
		if isFatal {
			if env == "development" {
				log.Fatal(msg)
				return nil
			} else {
				log.Error(msg)
				return errors.New(msg)
			}
		} else {
			log.Error(msg)
		}
	default:
		log.Debug(msg)
	}

	return nil
}
