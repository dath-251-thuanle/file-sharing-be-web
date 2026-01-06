package services_test

import (
	"fmt"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "postgres"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}
	
	dbname := os.Getenv("TEST_DB_NAME")
	if dbname == "" {
		dbname = "file_sharing_db" 
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open postgres DB: %v", err)
	}

	return db
}

