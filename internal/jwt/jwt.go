package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/zjy-dev/grpc-go-chatroom/internal/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GenerateJwt function generates a JWT token with the given username
func GenerateJwt(username string) (string, error) {
	if username == "" {
		return "", status.Errorf(codes.InvalidArgument, "username is empty")
	}

	// Create a new JWT token with the given username
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Subject: username,
		},
	)

	// Sign the token with the jwtKey
	tokenString, err := token.SignedString([]byte(config.JWT.JWTKey))
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to sign token: %v", err)
	}
	return tokenString, nil
}

// ParseJwt function parses a JWT token and returns the claims
func ParseJwt(tokenString string) (*jwt.RegisteredClaims, error) {
	if tokenString == "" {
		return nil, status.Errorf(codes.Unauthenticated, "token is empty")
	}
	// Parse the token with the jwtKey
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWT.JWTKey), nil
	})

	// Check if the token is valid
	if err != nil || !token.Valid {
		return nil, status.Errorf(codes.Unauthenticated, "failed to parse token: %v", err)
	}

	// Parse the claims from the token
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "failed to parse claims")
	}

	// Return the claims
	return claims, nil
}
