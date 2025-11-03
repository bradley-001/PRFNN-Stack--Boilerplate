package middleware

import (
	"context"
	"fmt"
	"time"

	"api/src/config"
	"api/src/constants"
	"api/src/lib/caching"
	"api/src/lib/general"
	"api/src/lib/security"
	"api/src/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// Get secret from .env
var jwtSecret = general.GetEnv("JWT_SECRET", "")
var httpsOn = general.GetEnv("NODE_ENV", "") == "production"

func CoreMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqIP := c.IP() // Get request IP

		// Extract JWT from request cookie ------------------------------
		jwtTokenString := c.Cookies("jwt_token")
		if jwtTokenString == "" {
			config.Log("No JWT Token in request cookies, can not authorise", 1, false, false)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "No JWT Token found",
			})
		}

		// Verify JWT -----------------------------------------------------
		parser := jwt.Parser{}

		// Parse token
		token, tokenErr := parser.ParseWithClaims(jwtTokenString, &security.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// If the signing methods is not HMAC, then this is not a token we issued - don't bother continuing, log error & block req
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Incorrect/unexpected signing method - %v", token.Header["alg"])
			}

			return []byte(jwtSecret), nil
		})

		// If there was an error, or the token is invlaid - log error & block req
		if tokenErr != nil || !token.Valid {
			config.Log(fmt.Sprintf("Token failed to parse: %s. Invalid or manipulated token (IP: %s)", tokenErr, reqIP), 2, false, true)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token failed to parse",
			})
		}

		// Get Token Claims (data) ----------------------------------------
		claims, claimsOk := token.Claims.(*security.JWTClaims)

		if !claimsOk {
			config.Log(fmt.Sprintf("Invalid token claims. Invalid or manipulated token (IP: %s)", reqIP), 2, false, true)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token failed to parse",
			})
		} else if claims == nil || claims.ExpiresAt == nil {

			config.Log(fmt.Sprintf("No expiry found in token claims, or claims is null. Invalid or manipulated token (IP: %s)", reqIP), 2, false, true)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token failed to parse",
			})
		}

		// TODO: Check claims.UID against rate limiting.

		// Start async check if session already exists  ----------------------
		type awaitSessionReturn struct {
			session models.Sessions
			errMsg  string
		}
		awaitSession := make(chan awaitSessionReturn, 1)
		go func() {
			// Try Redis cache first
			if cachedSession, err := caching.GetCachedSession(claims.SID); err == nil {
				// Cache hit! Return cached session
				awaitSession <- awaitSessionReturn{*cachedSession, ""}
				return
			} else if err != redis.Nil {
				// Redis error (not a cache miss), log it but continue to database
				config.Log(fmt.Sprintf("Redis could not fetch session %s: %v", claims.SID, err), 3, false, false)
			}

			// Cache miss or Redis error - fetch from database
			var existingSession models.Sessions
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(constants.DEFAULT_TIMEOUT)*time.Second)
			defer cancel()

			if err := config.DB.WithContext(ctx).First(&existingSession, "id = ?", claims.SID).Error; err != nil {
				awaitSession <- awaitSessionReturn{existingSession, fmt.Sprintf("Could not find session (%s), likely expired", claims.SID)}
				return
			}

			// Cache the session for future requests
			if cacheErr := caching.CacheSession(claims.SID, existingSession); cacheErr != nil {
				config.Log(fmt.Sprintf("Failed to cache session %s: %v", claims.SID, cacheErr), 1, false, false)
			}

			awaitSession <- awaitSessionReturn{existingSession, ""}
		}()

		// Start async check if user already exists  ----------------------
		type awaitUserReturn struct {
			user   models.Users
			errMsg string
		}
		awaitUser := make(chan awaitUserReturn, 1)
		go func() {
			// Try Redis cache first
			if cachedUser, err := caching.GetCachedUser(claims.UID); err == nil {
				// Cache hit! Return cached user
				awaitUser <- awaitUserReturn{*cachedUser, ""}
				return
			} else if err != redis.Nil {
				// Redis error (not a cache miss), log it but continue to database
				config.Log(fmt.Sprintf("Redis could not fetch user %s: %v", claims.UID, err), 3, false, false)
			}

			// Cache miss or Redis error - fetch from database
			var existingUser models.Users
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(constants.DEFAULT_TIMEOUT)*time.Second)
			defer cancel()

			if err := config.DB.WithContext(ctx).First(&existingUser, "id = ?", claims.UID).Error; err != nil {
				awaitUser <- awaitUserReturn{existingUser, fmt.Sprintf("Could not find user (%s) attached to request", claims.UID)}
				return
			}

			// Cache the user for future requests
			if cacheErr := caching.CacheUser(claims.UID, existingUser); cacheErr != nil {
				config.Log(fmt.Sprintf("Failed to cache user %s: %v", claims.UID, cacheErr), 1, false, false)
			}

			awaitUser <- awaitUserReturn{existingUser, ""}
		}()

		// Check if JWT needs refreshing  ---------------------------------
		var jwtRequiresRefresh bool = false
		if time.Until(claims.ExpiresAt.Time) <= constants.JWT_REFRESH_THRESHOLD {
			jwtRequiresRefresh = true
		}

		// If JWT does need to refresh, generate new one. -----------------
		var newToken string = ""
		if jwtRequiresRefresh {

			if token, err := security.GenerateJWT(claims.UID, claims.SID); err != nil {
				errMsg := fmt.Sprintf("Internal Server Error when trying to refresh JWT for UserId: %s", claims.UID)
				config.Log(errMsg, 2, false, true)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": errMsg,
				})
			} else {
				newToken = token
			}
		}

		// Await session & verify  ----------------------------------------
		var session models.Sessions
		if sessionRes := <-awaitSession; sessionRes.errMsg != "" {
			// An error here means no session, so wipe the cookie (logout).
			c.Cookie(&fiber.Cookie{
				Name:     "jwt_token",
				Value:    "",
				Expires:  time.Now().Add(-(5 * time.Minute)),
				HTTPOnly: true,
				Secure:   httpsOn,
				SameSite: "Strict",
				Path:     "/",
			})
			config.Log(sessionRes.errMsg, 2, false, true)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": sessionRes.errMsg,
			})
		} else {
			// Expired sessions are deleted from the database every minute, so it's best to just allow a minute
			// of tolerance instead of preforming a series of checks against expiry and edge conditons.
			session = sessionRes.session
		}

		// Await user & verify  ----------------------------------------
		var user models.Users
		if userRes := <-awaitUser; userRes.errMsg != "" {
			config.Log(userRes.errMsg, 2, false, true)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": userRes.errMsg,
			})
		} else {
			user = userRes.user
		}

		// Attach new JWT to cookie response, if needed -------------------
		if jwtRequiresRefresh {
			c.Cookie(&fiber.Cookie{
				Name:     "jwt_token",
				Value:    newToken,
				Expires:  time.Now().Add(constants.SESSION_DURATION),
				HTTPOnly: true,
				Secure:   httpsOn,
				SameSite: "Strict",
				Path:     "/",
			})
		}

		c.Locals("user", user)
		c.Locals("session", session)

		return c.Next()
	}
}

func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Global security headers for all responses
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		return c.Next()
	}
}
