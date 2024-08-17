package util

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WrapGRPCError wraps a given error with a gRPC status error if the error is not
// already a gRPC status error. It returns the original error if it is already
// a gRPC status error.
func WrapGRPCError(err error, code codes.Code, msg string) error {
	if err == nil {
		return nil
	}

	_, ok := status.FromError(err)
	if ok {
		return err
	}
	return status.Errorf(code, msg+": %v", err)
}

// MustLoadEnvFile loads the .env file if it exists and panics if it does not.
func MustLoadEnvFile() {
	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		// Log an error if the .env file cannot be loaded
		pwd, _ := os.Getwd()
		log.Panicf("error loading .env file: %v, current working directory: %v", err, pwd)
	}
}

// HashPassword hashes a password using the bcrypt algorithm
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPasswordHash checks if a password matches a hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
