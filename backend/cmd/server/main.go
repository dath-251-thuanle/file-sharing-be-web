package main

import (
//	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/admin"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/config"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/controllers"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/database"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/middleware"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/repositories"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/routes"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/services"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file from root folder (../.env) or backend folder (./.env)
	// Try root folder first (when running from backend/ directory)
	if err := godotenv.Load("../.env"); err != nil {
		// Fallback to backend/.env if root .env doesn't exist
		if err := godotenv.Load(".env"); err != nil {
			log.Printf("Warning: Error loading .env file: %v (using environment variables only)", err)
		}
	}

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

	// Initialize repositories
	userRepo := repositories.NewUserRepository(database.GetDB())
	loginSessionRepo := repositories.NewLoginSessionRepository(database.GetDB())

	// Initialize services
	authService := services.NewAuthServiceWithLoginSessions(userRepo, loginSessionRepo, cfg)
	fileService := services.NewFileService(database.GetDB(), store)
	statsService := services.NewStatisticsService(database.GetDB())
	historyService := services.NewDownloadHistoryService(database.GetDB())

	// Initialize controllers
	authController := controllers.NewAuthController(authService)
	fileController := controllers.NewFileController(fileService, statsService, historyService)

	// Middlewares
	authMiddleware := middleware.AuthMiddleware(cfg)

	// Setup router
	router := gin.Default()
	router.Use(corsMiddleware(&cfg.CORS))

	// Application routes
	routes.SetupRoutes(router, fileController, authController, authMiddleware)

	// Admin routes
	admin.Setup(router, database.GetDB(), store)

	// Start server using config
	addr := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)
	log.Printf("Server running on %s (storage=%T)", addr, store)

	if err := router.Run(addr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to run server: %v", err)
	}

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

func corsMiddleware(corsCfg *config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Backdoor: allow everything
		if os.Getenv("ALL_HOST") == "true" {
			if origin != "" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			}
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Cron-Secret")
			
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}
			c.Next()
			return
		}
		
		// Normal whitelist logic
		allowed := false
		if len(corsCfg.AllowedOrigins) > 0 {
			for _, allowedOrigin := range corsCfg.AllowedOrigins {
				if strings.TrimSuffix(origin, "/") == strings.TrimSuffix(allowedOrigin, "/") {
					allowed = true
					break
				}
			}
			// Debug logging: log when origin doesn't match
			if !allowed && origin != "" {
				log.Printf("[CORS] Origin '%s' not in allowed list: %v", origin, corsCfg.AllowedOrigins)
			} else if allowed {
				log.Printf("[CORS] Origin '%s' matched, allowing request", origin)
			}
		} else {
			allowed = true
		}
		
		if allowed && origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Cron-Secret")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
