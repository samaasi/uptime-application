package router

import (
    "time"

    "github.com/gin-contrib/cors"
    "github.com/samaasi/uptime-application/services/api-services/internal/api/controllers"
    "github.com/samaasi/uptime-application/services/api-services/internal/api/middleware"
    "github.com/samaasi/uptime-application/services/api-services/internal/config"
    "github.com/samaasi/uptime-application/services/api-services/internal/database"
    "github.com/samaasi/uptime-application/services/api-services/pkg/cache"
    "github.com/samaasi/uptime-application/services/api-services/pkg/notifier/email"
    "github.com/samaasi/uptime-application/services/api-services/pkg/storage"

    //"github.com/samaasi/uptime-application/services/api-services/pkg/urlsigner"

    "github.com/gin-gonic/gin"
)

func SetupRoutes(
	appConfig *config.Config,
	postgresClient database.Client,
	clickhouseClient database.Client,
	cacheService *cache.Service,
	storageDriver storage.Driver,
	emailService email.Service,
) (*gin.Engine, error) {


	// Initialize the signer with a secret
	// urlSigner := urlsigner.New(appConfig.App.Key,
	// 	urlsigner.WithExpiresParam("exp"),
	// 	urlsigner.WithSignatureParam("sig"),
	// 	urlsigner.WithClockSkewGrace(30*time.Second),
	// )
    // protected.Use(URLSignatureMiddleware(urlSigner))

	// Initialize repositories
	// (unused) repo := repositories.NewOTPRepository(cacheService)

	// Initialize services
	// (unused) otpService := services.NewUserOTPManagerService(repo, otp.NewOTPService(otp.DefaultOTPConfig()))

    // Initialize controllers
    healthController := controllers.NewHealthController(
        postgresClient,
        clickhouseClient,
        cacheService,
        storageDriver,
        emailService,
    )

	// --- Create Gin Router ---
	router := gin.New()

	// --- Global Middlewares ---
	router.Use(gin.Recovery())
	router.Use(middleware.LoggingMiddleware())
	router.Use(cors.New(getCORSConfig(appConfig)))

	// --- Routes ---
	// Health routes (public)
	router.GET("/health", healthController.GetHealth)
	router.GET("/livez", healthController.GetLiveness)
	router.GET("/readyz", healthController.GetReadiness)

	// API routes
	// api := router.Group("/api/v1")
	// {
	// 	// New authentication routes
	// 	auth := api.Group("/auth")
	// 	{
	// 	}

	// 	// Protected routes group
	// }

	return router, nil
}

func getCORSConfig(appConfig *config.Config) cors.Config {
	baseConfig := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	if appConfig.App.Mode == "production" {
		if appConfig.App.FrontendURL == "" {
			panic("CORS configuration error: APP_FRONTEND_URL cannot be empty in production mode")
		}
		baseConfig.AllowOrigins = []string{appConfig.App.FrontendURL}
	} else {
		baseConfig.AllowOrigins = []string{"http://localhost:3000"}
	}

	return baseConfig
}
