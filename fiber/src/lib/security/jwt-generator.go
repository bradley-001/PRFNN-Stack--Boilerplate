package security

import (
	"log"
	"os"
	"time"

	"api/src/constants"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UID string `json:"uid"`
	SID string `json:"sid"`
	jwt.RegisteredClaims
}

func GenerateJWT(uid, sid string) (string, error) {
	return GenerateJWTWithDuration(uid, sid, constants.JWT_DURATION)
}

func GenerateJWTWithDuration(uid, sid string, duration time.Duration) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("[ERROR] No JWT_SECRET found in .env")
		return "", os.ErrNotExist
	}

	claims := JWTClaims{
		UID: uid,
		SID: sid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil

}
