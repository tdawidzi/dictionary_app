package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config struct -  stores configuration info
type Config struct {
	DB_Host     string
	DB_Port     string
	DB_User     string
	DB_Password string
	DB_Name     string
}

// Load config from .env file - returns pointer to config struct and error
func Load(envpath string) (*Config, error) {
	path := envpath + ".env"
	err := godotenv.Load(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from environment file: %w", err)
	}
	config := &Config{
		DB_Host:     os.Getenv("DB_HOST"),
		DB_Port:     os.Getenv("DB_PORT"),
		DB_User:     os.Getenv("POSTGRES_USER"),
		DB_Password: os.Getenv("POSTGRES_PASSWORD"),
		DB_Name:     os.Getenv("POSTGRES_DB"),
	}
	return config, nil
}

func LoadTest(envpath string) (*Config, error) {
	path := envpath + ".env"
	err := godotenv.Load(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from environment file: %w", err)
	}
	config := &Config{
		DB_Host:     os.Getenv("DB_HOST_TEST"),
		DB_Port:     os.Getenv("DB_PORT_TEST"),
		DB_User:     os.Getenv("POSTGRES_USER_TEST"),
		DB_Password: os.Getenv("POSTGRES_PASSWORD_TEST"),
		DB_Name:     os.Getenv("POSTGRES_DB_TEST"),
	}
	return config, nil
}
