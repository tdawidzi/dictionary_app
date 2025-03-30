package utils

import (
	//"database/sql"

	"fmt"
	"log"

	"github.com/tdawidzi/dictionary_app/config"
	"github.com/tdawidzi/dictionary_app/models"

	// "github.com/jinzhu/gorm"
	// _ "github.com/jinzhu/gorm/dialects/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB represents the database connection.
var DB *gorm.DB

// ConnectDB establishes a connection to the PostgreSQL database.
func ConnectDB(config *config.Config) error {

	var err error

	// // Połączenie do PostgreSQL BEZ podawania konkretnej bazy danych
	// dsnBase := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable",
	// 	config.DB_Host, config.DB_Port, config.DB_User, config.DB_Password)

	// sqlDB, err := sql.Open("postgres", dsnBase)
	// if err != nil {
	// 	return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	// }
	// defer sqlDB.Close()

	// // Sprawdzenie, czy baza danych istnieje
	// var exists bool
	// query := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = '%s')", config.DB_Name)
	// err = sqlDB.QueryRow(query).Scan(&exists)
	// if err != nil {
	// 	return fmt.Errorf("failed to check if database exists: %w", err)
	// }

	// // Jeśli baza nie istnieje, stwórz ją
	// if !exists {
	// 	_, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE %s", config.DB_Name))
	// 	if err != nil {
	// 		return fmt.Errorf("failed to create database: %w", err)
	// 	}
	// 	log.Printf("Database %s created successfully", config.DB_Name)
	// }

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DB_Host, config.DB_Port, config.DB_User, config.DB_Password, config.DB_Name)

	// DB, err = gorm.Open("postgres", dsn)
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
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
