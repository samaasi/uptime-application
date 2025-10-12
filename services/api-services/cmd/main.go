package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/samaasi/uptime-application/services/api-services/internal/api/models"
	"github.com/samaasi/uptime-application/services/api-services/internal/api/router"
	"github.com/samaasi/uptime-application/services/api-services/internal/config"
	"github.com/samaasi/uptime-application/services/api-services/internal/database"
	"github.com/samaasi/uptime-application/services/api-services/internal/seeder"
	"github.com/samaasi/uptime-application/services/api-services/internal/utils"
	"github.com/samaasi/uptime-application/services/api-services/pkg/cache"
	"github.com/samaasi/uptime-application/services/api-services/pkg/logger"
	"github.com/samaasi/uptime-application/services/api-services/pkg/notifier/email"
	"github.com/samaasi/uptime-application/services/api-services/pkg/storage"

	"github.com/gin-gonic/gin"
)

type ServiceContainer struct {
	PostgresClient   database.Client
	ClickHouseClient database.Client
	CacheService     *cache.Service
	StorageDriver    storage.Driver
	EmailService     email.Service
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	appConfig, err := config.GetConfig()
	if err != nil {
		fmt.Printf("FATAL: %v\n", err)
		os.Exit(1)
	}

	if appConfig.App.Mode == config.AppModeProduction {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	if err := logger.InitFromConfig(appConfig.Logging); err != nil {
		fmt.Printf("FATAL: failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	defer utils.CheckError(logger.Sync())

	isDevMode := appConfig.App.Mode == config.AppModeDevelopment
	utils.CheckError(utils.InitResponseUtil(appConfig, isDevMode))

	logger.Info("Application starting",
		logger.String("version", appConfig.App.Version),
		logger.String("environment", appConfig.App.Mode),
		logger.String("port", appConfig.App.Port),
	)

	services, err := initializeServices(appConfig)
	if err != nil {
		logger.Fatal("failed to initialize services", logger.ErrorField(err))
	}
	go runHealthChecks(ctx, services)

	ginRouter, err := router.SetupRoutes(
		appConfig,
		services.PostgresClient,
		services.ClickHouseClient,
		services.CacheService,
		services.StorageDriver,
		services.EmailService,
	)
	if err != nil {
		logger.Fatal("Failed to setup routes", logger.ErrorField(err))
	}

	srv := &http.Server{
		Addr:    ":" + appConfig.App.Port,
		Handler: ginRouter,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Failed to start HTTP server", logger.ErrorField(err))
		}
	}()

	<-sigChan
	logger.Info("Shutting down application...")

	// Attempt to sync logger one last time before graceful shutdown
	// This is important to ensure all logs are flushed.
	if err := logger.Sync(); err != nil && !logger.IsBrokenPipeError(err) {
		logger.Error("Failed to sync logger during shutdown", logger.ErrorField(err))
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown failed", logger.ErrorField(err))
	} else {
		logger.Info("HTTP server gracefully stopped")
	}

	shutdownServices(shutdownCtx, services)

	logger.Info("Application shutdown complete.")
}

// initializeServices initializes and returns a ServiceContainer
func initializeServices(appConfig *config.Config) (*ServiceContainer, error) {
	services := &ServiceContainer{}

	if appConfig.Redis.Enable {
		redisClient, err := database.NewRedisClient(appConfig.Redis, database.DefaultRedisClientOptions())
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Redis client: %w", err)
		}
		services.CacheService = cache.NewCacheService(redisClient)
		logger.Info("Redis client and CacheService initialized")
	}

	if appConfig.Postgres.Enable {
		postgresOpts := database.DefaultPostgresClientOptions()
		postgresOpts.AutoMigrateModels = []interface{}{
			&models.User{},
			&models.Organization{},
			&models.OrganizationUser{},
			// Authorizaton models
			&models.Role{},
			&models.Permission{},
			&models.RolePermission{},
			&models.UserRole{},
			&models.UserPermission{},
			&models.Policy{},
		}

		pgClient, err := database.NewPostgresClient(appConfig.Postgres, postgresOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize PostgreSQL client: %w", err)
		}
		services.PostgresClient = pgClient
		logger.Info("PostgreSQL client initialized")

		// Seed default data including permissions
		ctx := context.Background()
		if err := seeder.SeedDefaultData(ctx, pgClient.DB()); err != nil {
			logger.Warn("Failed to seed default data", logger.ErrorField(err))
		}
	}

	// Initialize ClickHouse (GORM-based client)
	if appConfig.ClickHouse.Enable {
		chOpts := database.DefaultClickHouseClientOptions()
		chOpts.AutoMigrateModels = []interface{}{
			//&models.ClickHouseEvent{},
		}

		chClient, err := database.NewClickHouseClient(appConfig.ClickHouse, chOpts)
		if err != nil {
			logger.Error("Failed to initialize ClickHouse client", logger.ErrorField(err))
			return nil, fmt.Errorf("failed to initialize ClickHouse client: %w", err)
		}
		services.ClickHouseClient = chClient
		logger.Info("ClickHouse client initialized")
		chClient.DebugDbInfo(context.Background())
	}

	// Initialize Storage
	storageDriver, err := storage.NewLocalStorageDriver(appConfig.LocalStorage.Path, appConfig.LocalStorage.BaseURL)
	if err != nil {
		logger.Error("Failed to initialize storage driver", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to initialize storage driver: %w", err)
	}
	services.StorageDriver = storageDriver
	logger.Info("Storage driver initialized")

	// Initialize Email Service
	emailService, err := email.NewEmailService(&appConfig.Email)
	if err != nil {
		logger.Error("Failed to initialize email service", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to initialize email service: %w", err)
	}
	services.EmailService = emailService
	logger.Info("Email service initialized")

	return services, nil
}

// runHealthChecks periodically checks the health of various services
func runHealthChecks(ctx context.Context, services *ServiceContainer) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if services.PostgresClient != nil {
				if err := services.PostgresClient.HealthCheck(ctx); err != nil {
					logger.Error("PostgreSQL health check failed", logger.ErrorField(err))
				}
			}

			if services.ClickHouseClient != nil {
				if err := services.ClickHouseClient.HealthCheck(ctx); err != nil {
					logger.Error("ClickHouse health check failed", logger.ErrorField(err))
				}
			}

			if services.CacheService != nil {
				if err := services.CacheService.HealthCheck(ctx); err != nil {
					logger.Error("Redis (CacheService) health check failed", logger.ErrorField(err))
				}
			}

			if services.StorageDriver != nil {
				// Add storage health check if applicable
				if err := checkStorageHealth(services.StorageDriver); err != nil {
					logger.Error("Storage health check failed", logger.ErrorField(err))
				}
			}

			if services.EmailService != nil {
				if err := services.EmailService.HealthCheck(ctx); err != nil {
					logger.Error("Email service health check failed", logger.ErrorField(err))
				}
			}

		case <-ctx.Done():
			logger.Info("Health checks stopped as context was cancelled")
			return
		}
	}
}

// checkStorageHealth is a placeholder, implement based on storage needs
func checkStorageHealth(driver storage.Driver) error {
	// For local storage, check if base path is accessible
	// Return nil if healthy, error otherwise
	return nil
}

// shutdownServices gracefully shuts down all services
func shutdownServices(ctx context.Context, services *ServiceContainer) {
	_, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if services.PostgresClient != nil {
		if err := services.PostgresClient.Close(); err != nil {
			logger.Error("failed to close PostgreSQL client", logger.ErrorField(err))
		} else {
			logger.Info("PostgreSQL client closed successfully")
		}
	}

	if services.ClickHouseClient != nil {
		if err := services.ClickHouseClient.Close(); err != nil {
			logger.Error("failed to close ClickHouse client", logger.ErrorField(err))
		} else {
			logger.Info("ClickHouse client closed successfully")
		}
	}

	if services.CacheService != nil { // Close the CacheService
		if err := services.CacheService.Close(); err != nil {
			logger.Error("failed to close Redis (CacheService) client", logger.ErrorField(err))
		} else {
			logger.Info("Redis (CacheService) client closed successfully")
		}
	}

	// Add shutdown for storage and email if they have close methods
	if services.StorageDriver != nil {
		// If storage has a Close method, call it
	}

	if services.EmailService != nil {
		// If email has a Close method, call it
	}

	// Add other service shutdowns here
}
