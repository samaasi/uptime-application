package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/samaasi/uptime-application/services/api-services/internal/config"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresClient implements the Client interface for PostgreSQL with enhanced features
type PostgresClient struct {
	db      *gorm.DB
	options *PostgresClientOptions
	mu      sync.RWMutex
	closed  bool
}

// PostgresClientOptions holds comprehensive configuration for the Postgres client
type PostgresClientOptions struct {
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
	StatementCacheSize int
	PreparedStatements bool
}

// NewPostgresClient creates a new PostgreSQL client with enhanced initialization
func NewPostgresClient(cfg config.PostgresConfig, opts *PostgresClientOptions) (Client, error) {
	if opts == nil {
		opts = DefaultPostgresClientOptions()
	}

	client := &PostgresClient{
		options: opts,
	}

	if !cfg.Enable {
		return client, nil
	}

	if err := client.initialize(cfg); err != nil {
		return nil, fmt.Errorf("postgres initialization failed: %w", err)
	}

	return client, nil
}

// initialize sets up the database connection with retry logic
func (c *PostgresClient) initialize(cfg config.PostgresConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	var db *gorm.DB
	var err error

	// Connection with retry logic
	for attempt := 0; attempt <= c.options.MaxRetries; attempt++ {
		db, err = c.createConnection(cfg)
		if err == nil {
			break
		}

		if attempt < c.options.MaxRetries {
			log.Printf("Postgres connection attempt %d/%d failed: %v - retrying in %v",
				attempt+1, c.options.MaxRetries+1, err, c.options.RetryInterval)
			time.Sleep(c.options.RetryInterval)
			continue
		}
		return fmt.Errorf("failed to connect after %d attempts: %w", c.options.MaxRetries+1, err)
	}

	// Configure a connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(c.options.MaxIdleConns)
	sqlDB.SetMaxOpenConns(c.options.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(c.options.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(c.options.ConnMaxIdleTime)

	// Perform migrations if needed
	if len(c.options.AutoMigrateModels) > 0 {
		if err := c.performSafeMigrations(db); err != nil {
			_ = sqlDB.Close()
			return fmt.Errorf("migrations failed: %w", err)
		}
	}

	c.db = db
	return nil
}

// createConnection establishes a new database connection
func (c *PostgresClient) createConnection(cfg config.PostgresConfig) (*gorm.DB, error) {
	gormConfig := &gorm.Config{
		Logger:                 c.configureGormLogger(),
		DisableAutomaticPing:   true,
		PrepareStmt:            false,
		SkipDefaultTransaction: true,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  cfg.DSN(),
		PreferSimpleProtocol: true,
	}), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open gorm connection: %w", err)
	}

	// Verify connection
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

// configureGormLogger sets up appropriate logging based on the environment
func (c *PostgresClient) configureGormLogger() logger.Interface {
	logLevel := logger.Warn
	if gin.Mode() == gin.DebugMode || c.options.EnableDebugLogs {
		logLevel = logger.Info
	}

	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             c.options.SlowQueryThreshold,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  gin.Mode() == gin.DebugMode,
		},
	)
}

// performSafeMigrations handles database schema changes safely with retries
func (c *PostgresClient) performSafeMigrations(db *gorm.DB) error {
	const maxMigrationRetries = 3
	var lastErr error

	for i := 0; i < maxMigrationRetries; i++ {
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("DISCARD PLANS").Error; err != nil {
				return fmt.Errorf("failed to discard plans: %w", err)
			}

			if err := c.createDatabaseExtensions(tx); err != nil {
				return err
			}

			return tx.AutoMigrate(c.options.AutoMigrateModels...)
		})

		if err == nil {
			return nil
		}

		lastErr = err
		time.Sleep(time.Second * time.Duration(i+1))
	}

	return fmt.Errorf("migration failed after %d attempts: %w", maxMigrationRetries, lastErr)
}

// createDatabaseExtensions ensures required extensions are available
func (c *PostgresClient) createDatabaseExtensions(tx *gorm.DB) error {
	extensions := []SQLObjectToCreate{
		{
			Query:       "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"",
			Description: "uuid-ossp extension",
		},
		{
			Query:       "CREATE EXTENSION IF NOT EXISTS \"pgcrypto\"",
			Description: "pgcrypto extension",
		},
	}

	for _, ext := range extensions {
		if err := tx.Exec(ext.Query).Error; err != nil {
			return fmt.Errorf("failed to create %s: %w", ext.Description, err)
		}
	}
	return nil
}

// DB returns the underlying GORM DB instance with thread-safe access
func (c *PostgresClient) DB() *gorm.DB {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed || c.db == nil {
		return nil
	}
	return c.db
}

// WithContext returns a DB instance with context
func (c *PostgresClient) WithContext(ctx context.Context) *gorm.DB {
	if db := c.DB(); db != nil {
		return db.WithContext(ctx)
	}
	return nil
}

// Transaction executes a function within a database transaction with proper context handling
func (c *PostgresClient) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	db := c.DB()
	if db == nil {
		return fmt.Errorf("database connection not established")
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SET LOCAL statement_timeout = ?",
			c.options.TransactionTimeout.Milliseconds()).Error; err != nil {
			return fmt.Errorf("failed to set transaction timeout: %w", err)
		}
		return fn(tx)
	})
}

// DebugDbInfo outputs detailed information about the database connection
func (c *PostgresClient) DebugDbInfo(ctx context.Context) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		log.Println("Postgres client is closed")
		return
	}

	if c.db == nil {
		log.Println("Postgres client is disabled")
		return
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		log.Printf("Failed to get DB info: %v", err)
		return
	}

	stats := sqlDB.Stats()
	log.Printf("Postgres Connection Pool Stats:\n"+
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

// HealthCheck verifies database connectivity with timeout
func (c *PostgresClient) HealthCheck(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("postgres client is closed")
	}

	if c.db == nil {
		return fmt.Errorf("postgres client is not initialized")
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

// Close safely shuts down the database connection with thread safety
func (c *PostgresClient) Close() error {
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

// DefaultPostgresClientOptions provides optimized default configuration values
func DefaultPostgresClientOptions() *PostgresClientOptions {
	return &PostgresClientOptions{
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
		StatementCacheSize: 100,
		PreparedStatements: true,
	}
}
