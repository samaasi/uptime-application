package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samaasi/uptime-application/services/api-services/internal/config"
	"github.com/samaasi/uptime-application/services/api-services/pkg/logger"

	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// ClickHouseClient implements the Client interface for ClickHouse using GORM
type ClickHouseClient struct {
	db      *gorm.DB
	options *ClickHouseClientOptions
	mu      sync.RWMutex
	closed  bool
}

// ClickHouseClientOptions holds comprehensive configuration for the ClickHouse client
type ClickHouseClientOptions struct {
	MaxIdleConns       int
	MaxOpenConns       int
	ConnMaxLifetime    time.Duration
	ConnMaxIdleTime    time.Duration
	ConnectTimeout     time.Duration
	HealthCheckTimeout time.Duration
	TransactionTimeout time.Duration
	MaxRetries         int
	RetryInterval      time.Duration
	EnableDebugLogs    bool
	SlowQueryThreshold time.Duration
	AutoMigrateModels  []interface{}
}

// NewClickHouseClient creates a new ClickHouse client with enhanced initialization
func NewClickHouseClient(cfg config.ClickHouseConfig, opts *ClickHouseClientOptions) (Client, error) {
	if opts == nil {
		opts = DefaultClickHouseClientOptions()
	}

	client := &ClickHouseClient{options: opts}

	if !cfg.Enable {
		return client, nil
	}

	if err := client.initialize(cfg); err != nil {
		return nil, fmt.Errorf("clickhouse initialization failed: %w", err)
	}
	return client, nil
}

// initialize sets up the database connection with retry logic
func (c *ClickHouseClient) initialize(cfg config.ClickHouseConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	var db *gorm.DB
	var err error

	for attempt := 0; attempt <= c.options.MaxRetries; attempt++ {
		db, err = c.createConnection(cfg)
		if err == nil {
			break
		}

		if attempt < c.options.MaxRetries {
			logger.Warn("ClickHouse connection attempt failed, retrying",
				logger.Int("attempt", attempt+1),
				logger.Int("max_retries", c.options.MaxRetries+1),
				logger.ErrorField(err),
				logger.Duration("retry_interval", c.options.RetryInterval),
			)
			time.Sleep(c.options.RetryInterval)
			continue
		}
		return fmt.Errorf("failed to connect after %d attempts: %w", c.options.MaxRetries+1, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(c.options.MaxIdleConns)
	sqlDB.SetMaxOpenConns(c.options.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(c.options.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(c.options.ConnMaxIdleTime)

	if len(c.options.AutoMigrateModels) > 0 {
		if err := db.AutoMigrate(c.options.AutoMigrateModels...); err != nil {
			_ = sqlDB.Close()
			return fmt.Errorf("migrations failed: %w", err)
		}
	}

	c.db = db
	return nil
}

func (c *ClickHouseClient) createConnection(cfg config.ClickHouseConfig) (*gorm.DB, error) {
	gormConfig := &gorm.Config{
		Logger:               c.configureGormLogger(),
		DisableAutomaticPing: true,
		NowFunc:              func() time.Time { return time.Now().UTC() },
	}

	db, err := gorm.Open(clickhouse.Open(cfg.DSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open gorm connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.options.ConnectTimeout)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("connection ping failed: %w", err)
	}

	return db, nil
}

// configureGormLogger sets up a GORM logger based on options
func (c *ClickHouseClient) configureGormLogger() gormLogger.Interface {
	logLevel := gormLogger.Warn
	if gin.Mode() == gin.DebugMode || c.options.EnableDebugLogs {
		logLevel = gormLogger.Info
	}

	// Use Zap-backed GORM logger adapter for unified structured logging
	return logger.NewGormLogger(gormLogger.Config{
		SlowThreshold:             c.options.SlowQueryThreshold,
		LogLevel:                  logLevel,
		IgnoreRecordNotFoundError: true,
		ParameterizedQueries:      true,
		Colorful:                  false,
	})
}

// DB returns the underlying GORM DB instance with thread-safe access
func (c *ClickHouseClient) DB() *gorm.DB {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed || c.db == nil {
		return nil
	}
	return c.db
}

// WithContext returns a DB instance with context
func (c *ClickHouseClient) WithContext(ctx context.Context) *gorm.DB {
	if db := c.DB(); db != nil {
		return db.WithContext(ctx)
	}
	return nil
}

// Transaction executes a function within a database transaction
func (c *ClickHouseClient) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	db := c.DB()
	if db == nil {
		return fmt.Errorf("database connection not established")
	}
	// ClickHouse transactions are limited; emulate transactional behavior as needed
	return db.WithContext(ctx).Transaction(fn)
}

// HealthCheck verifies database connectivity with timeout
func (c *ClickHouseClient) HealthCheck(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("clickhouse client is closed")
	}

	if c.db == nil {
		return fmt.Errorf("clickhouse client is not initialized")
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.options.HealthCheckTimeout)
		defer cancel()
	}

	return sqlDB.PingContext(ctx)
}

// DebugDbInfo outputs detailed information about the database connection
func (c *ClickHouseClient) DebugDbInfo(ctx context.Context) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		log.Println("ClickHouse client is closed")
		return
	}

	if c.db == nil {
		log.Println("ClickHouse client is disabled")
		return
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		log.Printf("Failed to get DB info: %v", err)
		return
	}

	stats := sqlDB.Stats()
	log.Printf("ClickHouse Connection Pool Stats:\n"+
		"  Open Connections: %d\n"+
		"  In Use: %d\n"+
		"  Idle: %d\n"+
		"  Wait Count: %d\n"+
		"  Wait Duration: %v\n"+
		"  Max Idle Closed: %d\n"+
		"  Max Lifetime Closed: %d",
		stats.OpenConnections,
		stats.InUse,
		stats.Idle,
		stats.WaitCount,
		stats.WaitDuration,
		stats.MaxIdleClosed,
		stats.MaxLifetimeClosed,
	)
}

// Close safely shuts down the database connection with thread safety
func (c *ClickHouseClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed || c.db == nil {
		return nil
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	c.closed = true
	return sqlDB.Close()
}

// DefaultClickHouseClientOptions provides optimized default configuration values
func DefaultClickHouseClientOptions() *ClickHouseClientOptions {
	return &ClickHouseClientOptions{
		MaxIdleConns:       10,
		MaxOpenConns:       50,
		ConnMaxLifetime:    30 * time.Minute,
		ConnMaxIdleTime:    5 * time.Minute,
		ConnectTimeout:     5 * time.Second,
		HealthCheckTimeout: 3 * time.Second,
		TransactionTimeout: 30 * time.Second,
		MaxRetries:         5,
		RetryInterval:      2 * time.Second,
		EnableDebugLogs:    false,
		SlowQueryThreshold: 200 * time.Millisecond,
		AutoMigrateModels:  []interface{}{},
	}
}
