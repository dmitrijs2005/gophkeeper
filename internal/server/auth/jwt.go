// Package auth provides helpers for issuing and verifying JWT access tokens
// used by the GophKeeper server.
package auth

import (
	"errors"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/common"
	"github.com/golang-jwt/jwt/v5"
)

// Claims wraps jwt.RegisteredClaims and adds the application-specific UserID.
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

// GenerateToken creates an HS256-signed JWT containing the given userID and an
// expiration set to now + validityDuration. The token is signed with secretKey.
func GenerateToken(userID string, secretKey []byte, validityDuration time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(validityDuration)),
		},
		UserID: userID,
	})
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// GetUserIDFromToken parses and validates a JWT using secretKey and returns the
// embedded UserID. If the token is expired it returns common.ErrTokenExpired;
// if the token is otherwise invalid it returns common.ErrInvalidToken.
func GetUserIDFromToken(tokenString string, secretKey []byte) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", common.ErrTokenExpired
		}
		return "", err
	}
	if !token.Valid {
		return "", common.ErrInvalidToken
	}
	return claims.UserID, nil
}
