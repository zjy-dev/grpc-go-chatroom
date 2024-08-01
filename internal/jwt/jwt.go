package jwt

import (
	"log"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// jwtKey is a global variable that stores the JWT key
var jwtKey string

// init function is used to load the .env file and set the jwtKey variable
func mustLoadJwtKey() {

	if os.Getenv("JWT_KEY") != "" {
		jwtKey = os.Getenv("JWT_KEY")
		return
	}
	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		// Log an error if the .env file cannot be loaded
		pwd, _ := os.Getwd()
		log.Panicf("error loading .env file: %v, current working directory: %v", err, pwd)
	}

	if jwtKey = os.Getenv("JWT_KEY"); jwtKey == "" {
		log.Panicf("no JWT_KEY in .env")
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
		log.Printf("jwtKey: %s", jwtKey)
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
		return nil, status.Errorf(codes.InvalidArgument, "token is empty")
	}
	// Parse the token with the jwtKey
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})

	// Check if the token is valid
	if err != nil || !token.Valid {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse token: %v", err)
	}

	// Parse the claims from the token
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to parse claims")
	}

	// Return the claims
	return claims, nil
}
