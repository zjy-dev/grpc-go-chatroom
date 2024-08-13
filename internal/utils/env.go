package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func MustLoadEnvFile() {
	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		// Log an error if the .env file cannot be loaded
		pwd, _ := os.Getwd()
		log.Panicf("error loading .env file: %v, current working directory: %v", err, pwd)
	}
}
