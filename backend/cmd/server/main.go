package main

import (
	"fmt"
	"log"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Server running on %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Database: %s\n", cfg.Database.GetDSN())
	fmt.Printf("Config loaded successfully!\n")
}
