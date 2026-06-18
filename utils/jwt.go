package utils

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// getJWTKey resolves the signing key at call time (not package-init time) so
// that values loaded from a .env file by godotenv in main() are picked up.
func getJWTKey() []byte {
	key := os.Getenv("JWT_SECRET")
	if key == "" {
		log.Println("WARNING: JWT_SECRET is not set; falling back to an insecure default key")
		return []byte("super_secret_key_change_me_in_prod")
	}
	return []byte(key)
}

// GenerateToken issues a short-lived access token bound to the user's current
// token version, so bumping the version (logout / password change) revokes it.
func GenerateToken(userID uint, role string, tokenVersion int) (string, error) {
	claims := &jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"ver":     tokenVersion,
		"type":    "access",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTKey())
}

// GenerateRefreshToken issues a long-lived refresh token, also version-bound.
func GenerateRefreshToken(userID uint, tokenVersion int) (string, error) {
	claims := &jwt.MapClaims{
		"user_id": userID,
		"ver":     tokenVersion,
		"type":    "refresh",
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTKey())
}

func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Pin the signing method to prevent algorithm-confusion attacks.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return getJWTKey(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
