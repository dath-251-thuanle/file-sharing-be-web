package models

// This file provides a central import point for all models
// and helper functions for database operations

import (
	"crypto/rand"
	"encoding/hex"
)

// AllModels returns all model types for GORM operations
// Note: Migrations are handled by golang-migrate, not GORM AutoMigrate
func AllModels() []interface{} {
	return []interface{}{
		&User{},
		&File{},
		&SharedWith{},
		&SystemPolicy{},
		&FileStatistics{},
		&DownloadHistory{},
	}
}

// GenerateSecureToken generates a secure random token
// Used for API keys, share tokens, etc.
func GenerateSecureToken(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to less secure method if crypto/rand fails
		return ""
	}
	return hex.EncodeToString(bytes)
}

