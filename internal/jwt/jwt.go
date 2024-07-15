package jwt

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

// TODO: use environment variable as jwt key
const jwtKey string = "小xinzi"

func GenerateJwt(username string) (string, error) {
	// Generate jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			// ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			// IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject: username,
		},
	)

	return token.SignedString([]byte(jwtKey))
}

func ParseJwt(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil, errors.New("parse claims failed")
	}

	return claims, nil
}
