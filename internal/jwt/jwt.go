package jwt

import (
	"log"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// jwtKey is a global variable that stores the JWT key
var jwtKey string

// init function is used to load the .env file and set the jwtKey variable
func mustLoadJwtKey() {
	if jwtKey = os.Getenv("JWT_KEY"); jwtKey == "" {
		log.Panicf("JWT_KEY not in environment variable")
	}
}

// GenerateJwt function generates a JWT token with the given username
func GenerateJwt(username string) (string, error) {
	if username == "" {
		return "", status.Errorf(codes.InvalidArgument, "username is empty")
	}
	// Check if the jwtKey variable is already set
	if jwtKey == "" {
		mustLoadJwtKey()
	}

	// Create a new JWT token with the given username
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Subject: username,
		},
	)

	// Sign the token with the jwtKey
	tokenString, err := token.SignedString([]byte(jwtKey))
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
		return []byte(jwtKey), nil
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
