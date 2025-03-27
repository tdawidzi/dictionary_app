package utils

import (
	"dictionary-app/config"
	"dictionary-app/models"
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
)

// DB represents the database connection.
var DB *gorm.DB

// ConnectDB establishes a connection to the PostgreSQL database.
func ConnectDB(config *config.Config) error {
	var err error
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DB_Host, config.DB_Port, config.DB_User, config.DB_Password, config.DB_Name)
	DB, err = gorm.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Ensures that database has all necessary tables
	err = migrateTables(DB)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	return nil
}

func migrateTables(db *gorm.DB) error {
	// db.AutoMigrate does not return error - it panics. To protect API from fatal error:
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("Panic during tables migration: %v", r)
		}
	}()

	err := db.AutoMigrate(&models.Word{}, &models.Translation{}, &models.Example{})
	if err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}
	fmt.Println("Successfully created tables")
	return nil
}
