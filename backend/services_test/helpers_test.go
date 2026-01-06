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

	resetTestDatabase(t, db)
	t.Cleanup(func() {
		resetTestDatabase(t, db)
	})

	return db
}

func resetTestDatabase(t *testing.T, db *gorm.DB) {
	t.Helper()

	truncateStmt := `
TRUNCATE TABLE 
	download_history,
	file_statistics,
	files,
	login_sessions,
	system_policy,
	users
RESTART IDENTITY CASCADE`
	if err := db.Exec(truncateStmt).Error; err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	seedDefaultSystemPolicy(t, db)
}

func seedDefaultSystemPolicy(t *testing.T, db *gorm.DB) {
	t.Helper()

	stmt := `
INSERT INTO system_policy (
	id,
	max_file_size_mb,
	min_validity_hours,
	max_validity_days,
	default_validity_days,
	require_password_min_length
) VALUES (1, 50, 1, 30, 7, 6)
ON CONFLICT (id) DO UPDATE SET
	max_file_size_mb = EXCLUDED.max_file_size_mb,
	min_validity_hours = EXCLUDED.min_validity_hours,
	max_validity_days = EXCLUDED.max_validity_days,
	default_validity_days = EXCLUDED.default_validity_days,
	require_password_min_length = EXCLUDED.require_password_min_length`

	if err := db.Exec(stmt).Error; err != nil {
		t.Fatalf("failed to seed system policy: %v", err)
	}
}

