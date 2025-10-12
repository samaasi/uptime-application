package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/samaasi/uptime-application/services/api-services/internal/api/controllers"
	"github.com/samaasi/uptime-application/services/api-services/internal/api/middleware"
	"github.com/samaasi/uptime-application/services/api-services/internal/api/repositories"
	"github.com/samaasi/uptime-application/services/api-services/internal/api/services"
	"github.com/samaasi/uptime-application/services/api-services/internal/config"
	"github.com/samaasi/uptime-application/services/api-services/internal/database"
	"github.com/samaasi/uptime-application/services/api-services/pkg/cache"
	"github.com/samaasi/uptime-application/services/api-services/pkg/notifier/email"
	"github.com/samaasi/uptime-application/services/api-services/pkg/otp"
	"github.com/samaasi/uptime-application/services/api-services/pkg/security"
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

	// Initialize JWT service for token creation/verification
	jwtService, err := security.NewJWTService(appConfig.App.Key, appConfig.App.JWTExpiration)
	if err != nil {
		return nil, err
	}

	// Initialize repositories
	userRepo := repositories.NewUserRepository(postgresClient.DB())
	otpRepo := repositories.NewOTPRepository(cacheService)

	// Initialize services
	otpService := services.NewUserOTPManagerService(otpRepo, otp.NewOTPService(otp.DefaultOTPConfig()))
	authService := services.NewAuthService(userRepo, otpService, emailService, jwtService)

	// Initialize controllers
	healthController := controllers.NewHealthController(
		postgresClient,
		clickhouseClient,
		cacheService,
		storageDriver,
		emailService,
	)
	authController := controllers.NewAuthController(authService)

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
	api := router.Group("/api/v1")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/signup", authController.SignUp)
			auth.POST("/signin", authController.SignIn)
		}

		// Protected routes group (add later)
	}

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
