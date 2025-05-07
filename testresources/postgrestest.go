package testresources

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/tdawidzi/dictionary_app/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewSingleTestConnection(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := NewInMemTestDB(t)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	err = db.Start()
	if err != nil {
		t.Fatalf("failed to start embedded postgres: %v", err)
	}
	t.Cleanup(func() {
		db.Stop()
	})

	// Load config
	cfg, err := config.Load("..\\testresources\\")
	if err != nil {
		t.Fatalf("Error while loading configuration: %v", err)
	}

	postgresDB, err := ConnectDB(cfg)
	if err != nil {
		t.Fatalf("failed to connect to embedded_postgres database: %v", err)
	}

	return postgresDB
}

func NewInMemTestDB(t *testing.T) (*embeddedpostgres.EmbeddedPostgres, error) {
	// Load config
	cfg, err := config.Load("..\\testresources\\")
	if err != nil {
		t.Fatalf("Error while loading configuration: %v,", err)
	}
	port, _ := strconv.ParseUint(cfg.DB_Port, 0, 32)

	embedCfg := embeddedpostgres.DefaultConfig().
		Username(cfg.DB_User).
		Password(cfg.DB_Password).
		Database(cfg.DB_Name).
		Port(uint32(port)).
		StartTimeout(45 * time.Second)

	postgres := embeddedpostgres.NewDatabase(embedCfg)

	return postgres, nil
}

func ConnectDB(config *config.Config) (*gorm.DB, error) {

	var err error
	var Test_DB *gorm.DB

	// Get connection details from config
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DB_Host, config.DB_Port, config.DB_User, config.DB_Password, config.DB_Name)

	Test_DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return Test_DB, nil

}
