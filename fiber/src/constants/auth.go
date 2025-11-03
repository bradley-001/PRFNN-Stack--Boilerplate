package constants

import "time"

const (
	JWT_DURATION          = 5 * time.Minute     // Short lived JWT expiry (2 Minute)
	JWT_REFRESH_THRESHOLD = 30 * time.Second    // If the JWT is due to expire in <= (30 Seconds) then re-issue
	SESSION_DURATION      = 28 * 24 * time.Hour // Long-lived session (28 Days)
)
