package database

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/samaasi/uptime-application/services/api-services/internal/config"
	"github.com/samaasi/uptime-application/services/api-services/pkg/logger"

	"github.com/go-redis/redis/v8"
)

const (
	CircuitClosed = iota
	CircuitOpen
	CircuitHalfOpen
)

// RedisClientOptions holds comprehensive configuration for the Redis client.
type RedisClientOptions struct {
	ConnectTimeout          time.Duration
	OperationTimeout        time.Duration
	MaxRetries              int
	RetryInterval           time.Duration
	HealthCheckTimeout      time.Duration
	EnableCircuitBreaker    bool
	CircuitBreakerThreshold int
	CircuitBreakerTimeout   time.Duration
	MetricsEnabled          bool
}

// DefaultRedisClientOptions provides optimized default configuration values.
func DefaultRedisClientOptions() *RedisClientOptions {
	return &RedisClientOptions{
		ConnectTimeout:          5 * time.Second,
		OperationTimeout:        3 * time.Second,
		MaxRetries:              5,
		RetryInterval:           2 * time.Second,
		HealthCheckTimeout:      2 * time.Second,
		EnableCircuitBreaker:    true,
		CircuitBreakerThreshold: 5,
		CircuitBreakerTimeout:   30 * time.Second,
		MetricsEnabled:          true,
	}
}

// redisMetrics holds metrics for Redis operations.
type redisMetrics struct {
	requests      int64
	errors        int64
	hits          int64
	misses        int64
	latency       time.Duration
	lastResetTime time.Time
}

// RedisClient implements the CacheClient interface for Redis with enhanced features.
type RedisClient struct {
	client          *redis.Client
	options         *RedisClientOptions
	mu              sync.RWMutex
	closed          bool
	metrics         *redisMetrics
	circuitState    int
	failureCount    int
	lastFailureTime time.Time
}

// NewRedisClient creates a new RedisClient instance.
func NewRedisClient(cfg config.RedisConfig, opts *RedisClientOptions) (*RedisClient, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxConnAge:   cfg.MaxConnAge,
		DialTimeout:  opts.ConnectTimeout,
		ReadTimeout:  opts.OperationTimeout,
		WriteTimeout: opts.OperationTimeout,
	})

	client := &RedisClient{
		client:  redisClient,
		options: opts,
		metrics: &redisMetrics{
			lastResetTime: time.Now(),
		},
		circuitState: CircuitClosed,
		failureCount: 0,
	}

	return client, nil
}

// Set stores a value in Redis with a specified key and expiration duration.
func (c *RedisClient) Set(ctx context.Context, key string, value []byte, duration time.Duration) error {
	start := time.Now()
	var err error

	if c.options.EnableCircuitBreaker && c.isCircuitOpen() {
		err = errors.New("circuit breaker open")
	} else {
		cmd := c.client.Set(ctx, key, value, duration)
		err = cmd.Err()
	}

	if err != nil {
		c.recordMetrics(time.Since(start), "Set_Error")
		c.handleCircuitBreaker(err)
		logger.Error("Redis Set failed",
			logger.String("key", key),
			logger.Duration("duration", duration),
			logger.ErrorField(err),
			logger.String("op", "Set"),
		)
		return fmt.Errorf("redis set operation failed for key %s: %w", key, err)
	}

	c.recordMetrics(time.Since(start), "Set_Success")
	c.resetCircuitBreaker()
	return nil
}

// Get retrieves a value from Redis.
func (c *RedisClient) Get(ctx context.Context, key string) ([]byte, error) {
	start := time.Now()
	var err error
	var val []byte

	if c.options.EnableCircuitBreaker && c.isCircuitOpen() {
		err = errors.New("circuit breaker open")
	} else {
		cmd := c.client.Get(ctx, key)
		val, err = cmd.Bytes()
	}

	if err != nil {
		if errors.Is(err, redis.Nil) {
			c.recordMetrics(time.Since(start), "Get_Miss")
			c.resetCircuitBreaker()
			return nil, errors.New("key not found in mock cache")
		}
		c.recordMetrics(time.Since(start), "Get_Error")
		c.handleCircuitBreaker(err)
		logger.Error("Redis Get failed",
			logger.String("key", key),
			logger.ErrorField(err),
			logger.String("op", "Get"),
		)
		return nil, fmt.Errorf("redis get operation failed for key %s: %w", key, err)
	}

	c.recordMetrics(time.Since(start), "Get_Hit")
	c.resetCircuitBreaker()
	return val, nil
}

// Delete removes a key from Redis.
func (c *RedisClient) Delete(ctx context.Context, key string) error {
	start := time.Now()
	var err error

	if c.options.EnableCircuitBreaker && c.isCircuitOpen() {
		err = errors.New("circuit breaker open")
	} else {
		cmd := c.client.Del(ctx, key)
		err = cmd.Err()
	}

	if err != nil {
		c.recordMetrics(time.Since(start), "Delete_Error")
		c.handleCircuitBreaker(err)
		logger.Error("Redis Delete failed",
			logger.String("key", key),
			logger.ErrorField(err),
			logger.String("op", "Delete"),
		)
		return fmt.Errorf("redis delete operation failed for key %s: %w", key, err)
	}

	c.recordMetrics(time.Since(start), "Delete_Success")
	c.resetCircuitBreaker()
	return nil
}

// Update updates the value of an existing key in Redis without altering its TTL.
// It returns an error if the key does not exist or if the update fails.
func (c *RedisClient) Update(ctx context.Context, key string, value []byte) error {
	start := time.Now()
	var err error

	if c.options.EnableCircuitBreaker && c.isCircuitOpen() {
		err = errors.New("circuit breaker open")
	} else {
		cmd := c.client.Do(ctx, "SET", key, value, "KEEPTTL")
		err = cmd.Err()
	}

	if err != nil {
		c.recordMetrics(time.Since(start), "Update_Error")
		c.handleCircuitBreaker(err)
		logger.Error("Redis Update failed",
			logger.String("key", key),
			logger.ErrorField(err),
			logger.String("op", "Update"),
		)
		return fmt.Errorf("redis update operation failed for key %s: %w", key, err)
	}

	c.recordMetrics(time.Since(start), "Update_Success")
	c.resetCircuitBreaker()
	return nil
}

// Increment atomically increments the value of a key by 1 and returns the new value.
// If the key does not exist, it is set to 1.
func (c *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	start := time.Now()
	var result int64
	var err error

	if c.options.EnableCircuitBreaker && c.isCircuitOpen() {
		err = errors.New("circuit breaker open")
	} else {
		cmd := c.client.Incr(ctx, key)
		result, err = cmd.Result()
	}

	if err != nil {
		c.recordMetrics(time.Since(start), "Increment_Error")
		c.handleCircuitBreaker(err)
		logger.Error("Redis Increment failed",
			logger.String("key", key),
			logger.ErrorField(err),
			logger.String("op", "Increment"),
		)
		return 0, fmt.Errorf("redis increment operation failed for key %s: %w", key, err)
	}

	c.recordMetrics(time.Since(start), "Increment_Success")
	c.resetCircuitBreaker()
	logger.Debug("Redis Increment successful",
		logger.String("key", key),
		logger.Int64("new_value", result),
	)
	return result, nil
}

// Decrement atomically decrements the value of a key by 1 and returns the new value.
// If the key does not exist, it is set to -1.
func (c *RedisClient) Decrement(ctx context.Context, key string) (int64, error) {
	start := time.Now()
	var result int64
	var err error

	if c.options.EnableCircuitBreaker && c.isCircuitOpen() {
		err = errors.New("circuit breaker open")
	} else {
		cmd := c.client.Decr(ctx, key)
		result, err = cmd.Result()
	}

	if err != nil {
		c.recordMetrics(time.Since(start), "Decrement_Error")
		c.handleCircuitBreaker(err)
		logger.Error("Redis Decrement failed",
			logger.String("key", key),
			logger.ErrorField(err),
			logger.String("op", "Decrement"),
		)
		return 0, fmt.Errorf("redis decrement operation failed for key %s: %w", key, err)
	}

	c.recordMetrics(time.Since(start), "Decrement_Success")
	c.resetCircuitBreaker()
	logger.Debug("Redis Decrement successful",
		logger.String("key", key),
		logger.Int64("new_value", result),
	)
	return result, nil
}

// HealthCheck pings the Redis server to check its availability.
func (c *RedisClient) HealthCheck(ctx context.Context) error {
	start := time.Now()
	var err error

	if c.options.EnableCircuitBreaker && c.isCircuitOpen() {
		err = errors.New("circuit breaker open, skipping health check")
	} else {
		pingCmd := c.client.Ping(ctx)
		err = pingCmd.Err()
	}

	if err != nil {
		c.recordMetrics(time.Since(start), "HealthCheck_Error")
		c.handleCircuitBreaker(err)
		logger.Error("Redis HealthCheck failed", logger.ErrorField(err))
		return fmt.Errorf("redis health check failed: %w", err)
	}

	c.recordMetrics(time.Since(start), "HealthCheck_Success")
	c.resetCircuitBreaker()
	return nil
}

// Close closes the Redis client connection pool.
func (c *RedisClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return errors.New("redis client already closed")
	}
	c.closed = true
	logger.Info("Closing Redis client")
	return c.client.Close()
}

// isCircuitOpen checks if the circuit breaker is currently open.
func (c *RedisClient) isCircuitOpen() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.circuitState == CircuitOpen {
		if time.Since(c.lastFailureTime) > c.options.CircuitBreakerTimeout {
			c.mu.RUnlock()
			c.mu.Lock()
			defer c.mu.Unlock()
			if c.circuitState == CircuitOpen {
				c.circuitState = CircuitHalfOpen
				logger.Warn("Circuit breaker moved to Half-Open state", logger.String("last_failure_time", c.lastFailureTime.String()))
			}
			return false
		}
		return true
	}
	return false
}

// handleCircuitBreaker increments failure count and opens the circuit if threshold is met.
func (c *RedisClient) handleCircuitBreaker(err error) {
	if !c.options.EnableCircuitBreaker {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err != nil && !errors.Is(err, redis.Nil) {
		c.failureCount++
		c.lastFailureTime = time.Now()
		logger.Debug("Circuit breaker: failure counted", logger.Int("failure_count", c.failureCount), logger.ErrorField(err))

		if c.circuitState == CircuitHalfOpen && c.failureCount > 0 {
			c.circuitState = CircuitOpen
			c.failureCount = 0
			logger.Error("Circuit breaker moved from Half-Open back to Open state due to failure")
			return
		}

		if c.circuitState == CircuitClosed && c.failureCount >= c.options.CircuitBreakerThreshold {
			c.circuitState = CircuitOpen
			logger.Error("Circuit breaker opened due to consecutive failures",
				logger.Int("threshold", c.options.CircuitBreakerThreshold),
				logger.Int("failure_count", c.failureCount),
			)
		}
	}
}

// resetCircuitBreaker resets the circuit breaker to closed state on success.
func (c *RedisClient) resetCircuitBreaker() {
	if !c.options.EnableCircuitBreaker {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.circuitState != CircuitClosed {
		logger.Info("Circuit breaker reset to Closed state")
	}
	c.circuitState = CircuitClosed
	c.failureCount = 0
}

// recordMetrics updates operation metrics.
func (c *RedisClient) recordMetrics(latency time.Duration, op string) {
	if !c.options.MetricsEnabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics.requests++
	c.metrics.latency += latency

	switch op {
	case "Get_Hit":
		c.metrics.hits++
	case "Get_Miss":
		c.metrics.misses++
	case "Set_Error", "Get_Error", "Delete_Error", "Update_Error",
		"HealthCheck_Error":
		c.metrics.errors++
	}

	if c.metrics.requests%1000 == 0 || time.Since(c.metrics.lastResetTime) > time.Minute {
		avgLatency := time.Duration(0)
		if c.metrics.requests > 0 {
			avgLatency = c.metrics.latency / time.Duration(c.metrics.requests)
		}

		logger.Info("Redis metrics",
			logger.Int64("requests", c.metrics.requests),
			logger.Int64("errors", c.metrics.errors),
			logger.Int64("hits", c.metrics.hits),
			logger.Int64("misses", c.metrics.misses),
			logger.Duration("avg_latency", avgLatency),
		)

		// Reset metrics after logging
		c.metrics.latency = 0
		c.metrics.requests = 0
		c.metrics.errors = 0
		c.metrics.hits = 0
		c.metrics.misses = 0
		c.metrics.lastResetTime = time.Now()
	}
}

// GetMetrics returns current Redis operation metrics (for monitoring).
func (c *RedisClient) GetMetrics() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	avgLatency := time.Duration(0)
	if c.metrics.requests > 0 {
		avgLatency = c.metrics.latency / time.Duration(c.metrics.requests)
	}

	return map[string]interface{}{
		"requests":      c.metrics.requests,
		"errors":        c.metrics.errors,
		"hits":          c.metrics.hits,
		"misses":        c.metrics.misses,
		"avg_latency":   avgLatency.String(),
		"circuit_state": c.circuitState,
		"failure_count": c.failureCount,
	}
}
