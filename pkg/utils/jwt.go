package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken         = errors.New("invalid or expired token")
	ErrInvalidSigningMethod = errors.New("unexpected signing method")
)

type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	RoleID   uuid.UUID `json:"role_id"`
	RoleCode string    `json:"role_code"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uuid.UUID, email, roleCode string, roleID uuid.UUID, secret string, expHours int) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID:   userID,
		Email:    email,
		RoleID:   roleID,
		RoleCode: roleCode,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "proceive-api",
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func ValidateToken(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
