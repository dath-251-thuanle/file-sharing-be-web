package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/config"
	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations executes SQL migrations using golang-migrate
func RunMigrations(cfg *config.DatabaseConfig) error {
	path := os.Getenv("MIGRATIONS_PATH")
	if path == "" {
		path = "migrations"
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve migrations path: %w", err)
	}

	sourceURL := fmt.Sprintf("file://%s", filepath.ToSlash(absPath))

	m, err := migrate.New(sourceURL, cfg.GetURL())
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrate: %w", err)
	}

	return nil
}
