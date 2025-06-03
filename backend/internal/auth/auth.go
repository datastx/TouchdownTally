package auth

import (
	"fmt"
	"time"

	"touchdown-tally/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// CheckPassword verifies a password against its hash
func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// GenerateJWT generates a JWT token for a user
func GenerateJWT(user models.UserProfile, secret string) (string, error) {
	// Create token claims
	claims := jwt.MapClaims{
		"email_id":    user.EmailID,
		"user_id":     user.UserID,
		"username":    user.Username,
		"display_name": user.DisplayName,
		"exp":         time.Now().Add(24 * time.Hour).Unix(), // Token expires in 24 hours
		"iat":         time.Now().Unix(),
		"iss":         "touchdown-tally",
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// RefreshJWT generates a new JWT token from an existing valid token
func RefreshJWT(tokenString, secret string) (string, error) {
	claims, err := ValidateJWT(tokenString, secret)
	if err != nil {
		return "", fmt.Errorf("invalid token for refresh: %w", err)
	}

	// Create new token with updated expiration
	claims["exp"] = time.Now().Add(24 * time.Hour).Unix()
	claims["iat"] = time.Now().Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	newTokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign refreshed token: %w", err)
	}

	return newTokenString, nil
}
