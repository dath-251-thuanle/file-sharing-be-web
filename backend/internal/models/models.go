package models

import (
	"crypto/rand"
	"encoding/hex"
)

func AllModels() []interface{} {
	return []interface{}{
		&User{},
		&File{},
		&LoginSession{},
		&SystemPolicy{},
		&FileStatistics{},
		&DownloadHistory{},
	}
}

func GenerateSecureToken(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

