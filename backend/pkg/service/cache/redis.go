/**
 *
 * (c) Copyright Ascensio System SIA 2025
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/config"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	redis "github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client  *redis.Client
	options *CacheOptions
	logger  service.Logger
}

// TODO: Chained cache
func NewRedisCache(cfg *config.RedisConfig, logger service.Logger, opts ...Option) (*RedisCache, error) {
	options := DefaultCacheOptions()

	for _, opt := range opts {
		opt(options)
	}

	if err := options.Validate(); err != nil {
		return nil, fmt.Errorf("invalid cache options: %w", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     10,
		MinIdleConns: 2,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	logger.Info(ctx, "Connecting to Redis cache",
		service.Fields{
			"host":       cfg.Host,
			"port":       cfg.Port,
			"db":         cfg.DB,
			"key_prefix": options.KeyPrefix,
		})

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Error(ctx, "Failed to connect to Redis cache",
			service.Fields{
				"error": err.Error(),
			})
		return nil, fmt.Errorf("failed to connect to Redis cache: %w", err)
	}

	logger.Info(ctx, "Successfully connected to Redis cache")

	return &RedisCache{
		client:  client,
		options: options,
		logger:  logger,
	}, nil
}

func (c *RedisCache) buildKey(key string) string {
	return c.options.KeyPrefix + key
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	cacheKey := c.buildKey(key)

	val, err := c.client.Get(ctx, cacheKey).Bytes()
	if err == redis.Nil {
		c.logger.Debug(ctx, "Cache miss",
			service.Fields{
				"key": cacheKey,
			})
		return nil, nil
	}

	if err != nil {
		c.logger.Error(ctx, "Failed to get value from cache",
			service.Fields{
				"key":   cacheKey,
				"error": err.Error(),
			})
		return nil, fmt.Errorf("failed to get value from cache: %w", err)
	}

	c.logger.Debug(ctx, "Cache hit",
		service.Fields{
			"key":  cacheKey,
			"size": len(val),
		})

	return val, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	cacheKey := c.buildKey(key)

	if expiration == 0 {
		expiration = c.options.DefaultExpiration
	}

	err := c.client.Set(ctx, cacheKey, value, expiration).Err()
	if err != nil {
		c.logger.Error(ctx, "Failed to set value in cache",
			service.Fields{
				"key":        cacheKey,
				"expiration": expiration.String(),
				"error":      err.Error(),
			})
		return fmt.Errorf("failed to set value in cache: %w", err)
	}

	c.logger.Debug(ctx, "Value stored in cache",
		service.Fields{
			"key":        cacheKey,
			"size":       len(value),
			"expiration": expiration.String(),
		})

	return nil
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	cacheKey := c.buildKey(key)

	err := c.client.Del(ctx, cacheKey).Err()
	if err != nil {
		c.logger.Error(ctx, "Failed to delete key from cache",
			service.Fields{
				"key":   cacheKey,
				"error": err.Error(),
			})
		return fmt.Errorf("failed to delete key from cache: %w", err)
	}

	c.logger.Debug(ctx, "Key deleted from cache",
		service.Fields{
			"key": cacheKey,
		})

	return nil
}
