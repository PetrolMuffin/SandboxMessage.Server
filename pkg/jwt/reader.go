package jwt

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	Role   string `json:"role,omitempty"`
	UserID int64  `json:"uid" validate:"required"`
}

func ParseToken(tokenString string, secret []byte) (*Claims, error) {
	keyFunc := func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return secret, nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, keyFunc)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, err
		}

		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("could not parse claims: %w", err)
}
