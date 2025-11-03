package general

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Load environment variables from .env file
var _ = godotenv.Load(".env")

func GetEnv[T string | int](key string, defaultValue T) T {

	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var result any
	switch any(defaultValue).(type) {
	case string:
		result = value
	case int:
		if parsed, err := strconv.ParseInt(value, 10, 32); err == nil {
			result = int(parsed)
		} else {
			result = defaultValue
		}
	}

	return result.(T)
}
