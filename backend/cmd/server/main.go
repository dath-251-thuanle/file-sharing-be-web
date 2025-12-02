package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/admin"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/config"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/controllers"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/database"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/routes"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/services"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/storage"
	"github.com/gin-gonic/gin"
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
	statsService := services.NewStatisticsService(database.GetDB())
	historyService := services.NewDownloadHistoryService(database.GetDB())

	// Initialize controllers
	fileController := controllers.NewFileController(fileService, statsService, historyService)

	// Setup router
	router := gin.Default()
	router.Use(corsMiddleware())
	routes.SetupRoutes(router, fileController)

	admin.Setup(router, database.GetDB(), store) // Pass the router and the DB instance directly to your single-file admin manager

	addr := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)
	go func() {
		log.Printf("Server running on %s (storage=%T)", addr, store)
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

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Cron-Secret")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
