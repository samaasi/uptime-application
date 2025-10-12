package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/samaasi/uptime-application/services/api-services/internal/database"
	"github.com/samaasi/uptime-application/services/api-services/pkg/logger"
	"golang.org/x/sync/singleflight"
)

const (
	cachedErrorPrefix = "ERR:"
	cachedErrorTTL    = 1 * time.Minute
)

// Service defines the application-level caching operations.
type Service struct {
	cacheClient database.CacheClient
	sfGroup     *singleflight.Group
	randSource  rand.Source
}

// NewCacheService creates a new instance of Service.
func NewCacheService(client database.CacheClient) *Service {
	return &Service{
		cacheClient: client,
		sfGroup:     &singleflight.Group{},
		randSource:  rand.NewSource(time.Now().UnixNano()),
	}
}

// Set stores a value in the cache with a specified key and expiration duration.
func (s *Service) Set(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		logger.Error("failed to marshal cache value", logger.String("key", key), logger.ErrorField(err))
		return fmt.Errorf("failed to marshal cache value for key %s: %w", key, err)
	}

	jitterDuration := s.addJitter(duration, 0.1)
	logger.Debug("setting cache value with jitter", logger.String("key", key), logger.Duration("duration", duration), logger.Duration("jittered_duration", jitterDuration))

	return s.cacheClient.Set(ctx, key, data, jitterDuration)
}

// Get retrieves a value from the cache and unmarshal it into the provided destination.
func (s *Service) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := s.cacheClient.Get(ctx, key)
	if err != nil {
		logger.Debug("cache get miss or error", logger.String("key", key), logger.ErrorField(err))
		return err
	}

	if len(data) >= len(cachedErrorPrefix) && string(data[:len(cachedErrorPrefix)]) == cachedErrorPrefix {
		errMsg := string(data[len(cachedErrorPrefix):])
		logger.Warn("retrieved cached error for key", logger.String("key", key), logger.String("cached_error", errMsg))
		return fmt.Errorf("cached error: %s", errMsg)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		logger.Error("failed to unmarshal cache value", logger.String("key", key), logger.ErrorField(err))
		return fmt.Errorf("failed to unmarshal cache value for key %s: %w", key, err)
	}
	logger.Debug("cache hit", logger.String("key", key))
	return nil
}

// Delete removes a value from the cache by its key.
func (s *Service) Delete(ctx context.Context, key string) error {
	logger.Info("deleting cache key", logger.String("key", key))
	return s.cacheClient.Delete(ctx, key)
}

// Update updates the value of an existing key in the cache without altering its TTL.
func (s *Service) Update(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		logger.Error("failed to marshal cache value for key for update", logger.String("key", key), logger.ErrorField(err))
		return fmt.Errorf("failed to marshal cache value for key %s for update: %w", key, err)
	}

	logger.Debug("updating cache key with new value", logger.String("key", key))
	return s.cacheClient.Update(ctx, key, data)
}

// HealthCheck performs a health check on the underlying cache client.
func (s *Service) HealthCheck(ctx context.Context) error {
	return s.cacheClient.HealthCheck(ctx)
}

// Close closes the cache client connection.
func (s *Service) Close() error {
	return s.cacheClient.Close()
}

// Increment atomically increments the value of a key by 1 and returns the new value.
// If the key does not exist, it is set to 1.
func (s *Service) Increment(ctx context.Context, key string) (int64, error) {
	logger.Debug("incrementing cache key", logger.String("key", key))
	return s.cacheClient.Increment(ctx, key)
}

// Decrement atomically decrements the value of a key by 1.
func (s *Service) Decrement(ctx context.Context, key string) (int64, error) {
	return s.cacheClient.Decrement(ctx, key)
}

// GetOrSet retrieves a value from the cache by key. If not found or expired,
// it executes the provided `fetchFunc`, stores the result, and returns it.
// It uses `singleflight` to prevent cache stampedes and can cache errors.
func (s *Service) GetOrSet(ctx context.Context, key string, dest interface{}, duration time.Duration, fetchFunc func() (interface{}, error)) error {
	err := s.Get(ctx, key, dest)
	if err == nil {
		return nil
	}

	if err.Error() == fmt.Sprintf("cached error: %s", err.Error()[len("cached error:"):]) {
		return err
	}

	logger.Info("cache miss, attempting to fetch with singleflight", logger.String("key", key))

	value, fetchErr, _ := s.sfGroup.Do(key, func() (interface{}, error) {
		fetchedVal, innerFetchErr := fetchFunc()
		if innerFetchErr != nil {
			logger.Error("fetch function failed for cache key", logger.String("key", key), logger.ErrorField(innerFetchErr))
			if setErr := s.cacheClient.Set(ctx, key, []byte(cachedErrorPrefix+innerFetchErr.Error()), cachedErrorTTL); setErr != nil {
				logger.Error("failed to cache error result", logger.String("key", key), logger.ErrorField(setErr))
			}
			return nil, innerFetchErr
		}

		_, marshalErr := json.Marshal(fetchedVal)
		if marshalErr != nil {
			logger.Error("failed to marshal fetched value for cache", logger.String("key", key), logger.ErrorField(marshalErr))
			return nil, fmt.Errorf("failed to marshal fetched value: %w", marshalErr)
		}

		if setErr := s.Set(ctx, key, fetchedVal, duration); setErr != nil {
			logger.Error("failed to set cache value after successful fetch", logger.String("key", key), logger.ErrorField(setErr))
			return nil, fmt.Errorf("failed to set cache value: %w", setErr)
		}

		logger.Info("successfully fetched and cached value", logger.String("key", key))
		return fetchedVal, nil
	})

	if fetchErr != nil {
		return fetchErr
	}

	// Unmarshal the fetched value (from singleflight result) into dest
	// The `value` returned by sfGroup.Do is the `interface{}` returned by the `fetchFunc`.
	// We need to marshal and unmarshal it again into the target `dest` type.
	dataToUnmarshal, marshalErr := json.Marshal(value)
	if marshalErr != nil {
		logger.Error("failed to marshal value from singleflight result", logger.String("key",
			key), logger.ErrorField(marshalErr))
		return fmt.Errorf("failed to marshal singleflight result for key %s: %w", key, marshalErr)
	}

	if err := json.Unmarshal(dataToUnmarshal, dest); err != nil {
		logger.Error("failed to unmarshal singleflight result into destination", logger.String("key", key), logger.ErrorField(err))
		return fmt.Errorf("failed to unmarshal singleflight result for key %s: %w", key, err)
	}

	return nil
}

// addJitter adds a random percentage to the duration to prevent cache stampedes due to
// simultaneous expirations (expiration synchronization).
// This uses a thread-safe random source.
func (s *Service) addJitter(duration time.Duration, percentage float64) time.Duration {
	if duration <= 0 {
		return duration
	}

	r := rand.New(s.randSource)

	jitter := (2 * percentage * r.Float64()) - percentage
	return duration + time.Duration(float64(duration)*jitter)
}
