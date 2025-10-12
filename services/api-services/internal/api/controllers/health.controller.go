package controllers

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/samaasi/uptime-application/services/api-services/pkg/cache"
	"github.com/samaasi/uptime-application/services/api-services/pkg/notifier/email"
	"github.com/samaasi/uptime-application/services/api-services/pkg/storage"

	"github.com/samaasi/uptime-application/services/api-services/internal/database"

	"github.com/gin-gonic/gin"
)

// ServiceStatus defines the health of an individual dependency.
type ServiceStatus struct {
	Status        string `json:"status"`
	Error         string `json:"error,omitempty"`
	ActiveClients int    `json:"active_clients,omitempty"`
}

// HealthResponse defines the structured response for the health check endpoint.
type HealthResponse struct {
	OverallStatus           string                   `json:"overall_status"`
	OverallHealthPercentage float64                  `json:"overall_health_percentage"`
	Services                map[string]ServiceStatus `json:"services"`
}

// HealthController handles health-related API requests.
type HealthController struct {
	PostgresClient   database.Client
	ClickHouseClient database.Client
	CacheService     *cache.Service
	StorageDriver    storage.Driver
	EmailService     email.Service
}

// NewHealthController creates a new instance of HealthController.
func NewHealthController(
	postgresClient database.Client,
	clickhouseClient database.Client,
	cacheService *cache.Service,
	storageDriver storage.Driver,
	emailService email.Service,
) *HealthController {
	return &HealthController{
		PostgresClient:   postgresClient,
		ClickHouseClient: clickhouseClient,
		CacheService:     cacheService,
		StorageDriver:    storageDriver,
		EmailService:     emailService,
	}
}

// GetHealth handles the GET /health endpoint by running concurrent checks.
func (ctrl *HealthController) GetHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	resultsChan := make(chan struct {
		Name   string
		Status ServiceStatus
	}, 5)

	serviceChecks := 0

	if ctrl.PostgresClient != nil {
		serviceChecks++
		wg.Add(1)
		go ctrl.checkPostgres(ctx, &wg, resultsChan)
	}

	if ctrl.ClickHouseClient != nil {
		serviceChecks++
		wg.Add(1)
		go ctrl.checkClickHouse(ctx, &wg, resultsChan)
	}

	if ctrl.CacheService != nil {
		serviceChecks++
		wg.Add(1)
		go ctrl.checkCache(ctx, &wg, resultsChan)
	}

	if ctrl.StorageDriver != nil {
		serviceChecks++
		wg.Add(1)
		go ctrl.checkStorage(ctx, &wg, resultsChan)
	}

	if ctrl.EmailService != nil {
		serviceChecks++
		wg.Add(1)
		go ctrl.checkEmail(ctx, &wg, resultsChan)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	services := make(map[string]ServiceStatus)
	healthyChecks := 0
	isDegraded := false

	for res := range resultsChan {
		services[res.Name] = res.Status
		if res.Status.Status == "up" {
			healthyChecks++
		} else {
			isDegraded = true
		}
	}

	var overallHealth float64
	if serviceChecks > 0 {
		overallHealth = (float64(healthyChecks) / float64(serviceChecks)) * 100
	}

	response := HealthResponse{
		OverallStatus:           "healthy",
		OverallHealthPercentage: overallHealth,
		Services:                services,
	}
	httpStatus := http.StatusOK

	if isDegraded {
		response.OverallStatus = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, response)
}

// checkPostgres performs the health check for the PostgreSQL database.
func (ctrl *HealthController) checkPostgres(ctx context.Context, wg *sync.WaitGroup, resChan chan<- struct {
	Name   string
	Status ServiceStatus
}) {
	defer wg.Done()
	if err := ctrl.PostgresClient.HealthCheck(ctx); err != nil {
		resChan <- struct {
			Name   string
			Status ServiceStatus
		}{"database", ServiceStatus{Status: "down", Error: err.Error()}}
	} else {
		resChan <- struct {
			Name   string
			Status ServiceStatus
		}{"database", ServiceStatus{Status: "up"}}
	}
}

// checkClickHouse performs the health check for the ClickHouse database.
func (ctrl *HealthController) checkClickHouse(ctx context.Context, wg *sync.WaitGroup, resChan chan<- struct {
	Name   string
	Status ServiceStatus
}) {
	defer wg.Done()
	if err := ctrl.ClickHouseClient.HealthCheck(ctx); err != nil {
		resChan <- struct {
			Name   string
			Status ServiceStatus
		}{"clickhouse", ServiceStatus{Status: "down", Error: err.Error()}}
	} else {
		resChan <- struct {
			Name   string
			Status ServiceStatus
		}{"clickhouse", ServiceStatus{Status: "up"}}
	}
}

// GetLiveness provides a simple liveness probe for Kubernetes.
func (ctrl *HealthController) GetLiveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "alive"})
}

// GetReadiness provides a readiness probe for Kubernetes.
// Checks critical dependencies required to serve traffic.
func (ctrl *HealthController) GetReadiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	errors := []string{}

	if ctrl.PostgresClient != nil {
		if err := ctrl.PostgresClient.HealthCheck(ctx); err != nil {
			errors = append(errors, "postgres: "+err.Error())
		}
	}

	if ctrl.ClickHouseClient != nil {
		if err := ctrl.ClickHouseClient.HealthCheck(ctx); err != nil {
			errors = append(errors, "clickhouse: "+err.Error())
		}
	}

	if ctrl.CacheService != nil {
		if err := ctrl.CacheService.HealthCheck(ctx); err != nil {
			errors = append(errors, "redis: "+err.Error())
		}
	}

	if len(errors) > 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "errors": errors})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

// checkCache performs the health check for the Redis cache.
func (ctrl *HealthController) checkCache(ctx context.Context, wg *sync.WaitGroup, resChan chan<- struct {
	Name   string
	Status ServiceStatus
}) {
	defer wg.Done()
	if err := ctrl.CacheService.HealthCheck(ctx); err != nil {
		resChan <- struct {
			Name   string
			Status ServiceStatus
		}{"redis", ServiceStatus{Status: "down", Error: err.Error()}}
	} else {
		resChan <- struct {
			Name   string
			Status ServiceStatus
		}{"redis", ServiceStatus{Status: "up"}}
	}
}

// checkStorage performs the health check for the storage driver.
func (ctrl *HealthController) checkStorage(ctx context.Context, wg *sync.WaitGroup, resChan chan<- struct {
	Name   string
	Status ServiceStatus
}) {
	defer wg.Done()
	// Implement storage health check logic
	// For local storage, we'll assume it's up if we can access it; adjust as needed
	_, err := ctrl.StorageDriver.Exists(ctx, "health_check")
	if err != nil {
		resChan <- struct {
			Name   string
			Status ServiceStatus
		}{"storage", ServiceStatus{Status: "down", Error: err.Error()}}
	} else {
		resChan <- struct {
			Name   string
			Status ServiceStatus
		}{"storage", ServiceStatus{Status: "up"}}
	}
}

// checkEmail performs the health check for the email service.
func (ctrl *HealthController) checkEmail(ctx context.Context, wg *sync.WaitGroup, resChan chan<- struct {
	Name   string
	Status ServiceStatus
}) {
	defer wg.Done()
	if err := ctrl.EmailService.HealthCheck(ctx); err != nil {
		resChan <- struct {
			Name   string
			Status ServiceStatus
		}{"email", ServiceStatus{Status: "down", Error: err.Error()}}
	} else {
		resChan <- struct {
			Name   string
			Status ServiceStatus
		}{"email", ServiceStatus{Status: "up"}}
	}
}
