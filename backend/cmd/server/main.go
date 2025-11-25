package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gin-gonic/gin"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/config"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/controllers"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/database"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/routes"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/services"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := database.Connect(&cfg.Database); err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("error closing database: %v", err)
		}
	}()

	store, err := buildStorage(cfg)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}

	// Initialize services
	fileService := services.NewFileService(database.GetDB(), store)

	// Initialize controllers
	fileController := controllers.NewFileController(fileService)

	// Setup router
	router := gin.Default()
	routes.SetupRoutes(router, fileController)

	addr := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)
	go func() {
		log.Printf("🚀 Server running on %s (storage=%T)", addr, store)
		if err := router.Run(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to run server: %v", err)
		}
	}()

	waitForShutdown()
}

func buildStorage(cfg *config.Config) (storage.Storage, error) {
	if cfg.CloudStorage.Enabled {
		azStorage, err := storage.NewAzureBlobStorage(
			cfg.CloudStorage.Endpoint,
			cfg.CloudStorage.AccessKey,
			cfg.CloudStorage.SecretKey,
			cfg.CloudStorage.PublicContainer,
			cfg.CloudStorage.PrivateContainer,
		)
		if err != nil {
			log.Printf("Azure Blob init failed, falling back to LocalStorage: %v", err)
			basePath := cfg.Storage.Path
			if basePath == "" {
				basePath = "./storage/uploads"
			}
			return storage.NewLocalStorage(basePath), nil
		}
		return azStorage, nil
	}

	basePath := cfg.Storage.Path
	if basePath == "" {
		basePath = "./storage/uploads"
	}
	return storage.NewLocalStorage(basePath), nil
}

func waitForShutdown() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("Shutting down server...")
}
