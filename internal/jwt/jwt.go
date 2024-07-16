package jwt

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

// jwtKey is a global variable that stores the JWT key
var jwtKey string

// init function is used to load the .env file and set the jwtKey variable
func init() {

	// Load the .env file
	err := godotenv.Load()
	if err != nil {

		// Log an error if the .env file cannot be loaded
		log.Fatal("error loading .env file")
	}

	// Set the jwtKey variable to the value of the JWT_KEY environment variable
	jwtKey = os.Getenv("JWT_KEY")

	// Print the jwtKey variable
	fmt.Println("jwtKey: ", jwtKey)
}

// GenerateJwt function generates a JWT token with the given username
func GenerateJwt(username string) (string, error) {

	// Create a new JWT token with the given username
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.RegisteredClaims{

			Subject: username,
		},
	)

	// Sign the token with the jwtKey
	return token.SignedString([]byte(jwtKey))
}

// ParseJwt function parses a JWT token and returns the claims
func ParseJwt(tokenString string) (*jwt.RegisteredClaims, error) {

	// Parse the token with the jwtKey
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {

		// Return the jwtKey
		return []byte(jwtKey), nil
	})
	if err != nil {
		return nil, err
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Parse the claims from the token
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil, errors.New("parse claims failed")
	}

	// Return the claims
	return claims, nil
}
