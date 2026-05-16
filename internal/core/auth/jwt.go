package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"go-marketplace/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GenerateAccessToken(userID uuid.UUID, role domain.UserRole, secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID.String(),
		"role": string(role),
		"exp":  time.Now().Add(time.Minute * 15).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

type TokenClaims struct {
	UserID uuid.UUID
	Role   domain.UserRole
}

func ValidateAccessToken(tokenString string, secret string) (*TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		sub, ok := claims["sub"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid sub claim")
		}
		userID, err := uuid.Parse(sub)
		if err != nil {
			return nil, fmt.Errorf("invalid userID format")
		}

		role, ok := claims["role"].(string)
		if !ok || role == "" {
			return nil, fmt.Errorf("missing or invalid role claim")
		}

		return &TokenClaims{
			UserID: userID,
			Role:   domain.UserRole(role),
		}, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func GenerateVerificationCode() (string, error) {
	b := make([]byte, 3)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// Simple numeric code from bytes
	code := (uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])) % 1000000
	return fmt.Sprintf("%06d", code), nil
}
