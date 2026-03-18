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
package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/config"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	redis "github.com/redis/go-redis/v9"
)

type RateLimiter interface {
	Allow(identifier string) (bool, error)
}

type RedisStore struct {
	client      *redis.Client
	config      *config.RateLimitConfig
	logger      service.Logger
	leakyScript *redis.Script
}

const leakyBucketScript = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local window = tonumber(ARGV[4])

local bucket = redis.call('HMGET', key, 'last_update', 'water_level')
local last_update = bucket[1] and tonumber(bucket[1]) or now
local water_level = bucket[2] and tonumber(bucket[2]) or 0

local time_passed = now - last_update
local leak_rate = rate / window
local leaked = time_passed * leak_rate
water_level = math.max(0, water_level - leaked)

if water_level + 1 > capacity then
    redis.call('HMSET', key, 'last_update', now, 'water_level', water_level)
    redis.call('EXPIRE', key, window / 1000 * 2)
    return 0
end

water_level = water_level + 1
redis.call('HMSET', key, 'last_update', now, 'water_level', water_level)
redis.call('EXPIRE', key, window / 1000 * 2)
return 1
`

func NewRedisStore(config *config.Config, logger service.Logger) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), config.Redis.Timeout)
	defer cancel()

	logger.Info(ctx, "connecting to Redis",
		service.Fields{
			"host":     config.Redis.Host,
			"port":     config.Redis.Port,
			"database": config.Redis.DB,
		})

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Error(ctx, "failed to connect to Redis",
			service.Fields{
				"error": err.Error(),
			})
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info(ctx, "successfully connected to Redis")

	return &RedisStore{
		client:      client,
		config:      config.RateLimit,
		logger:      logger,
		leakyScript: redis.NewScript(leakyBucketScript),
	}, nil
}

func (s *RedisStore) Allow(identifier string) (bool, error) {
	if identifier == "" {
		return false, fmt.Errorf("identifier cannot be empty")
	}

	tctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	key := fmt.Sprintf("rate_limit:%s", identifier)
	wms := int64(s.config.Window.Milliseconds())
	nms := time.Now().UnixMilli()

	capacity := s.config.Rate

	allowed, err := s.leakyScript.Run(tctx, s.client,
		[]string{key},
		s.config.Rate,
		capacity,
		nms,
		wms,
	).Int()

	if err != nil {
		s.logger.Error(tctx, "error processing rate limit",
			service.Fields{
				"identifier": identifier,
				"error":      err.Error(),
			})
		return false, err
	}

	if allowed == 0 {
		s.logger.Warn(tctx, "rate limit exceeded",
			service.Fields{
				"identifier": identifier,
				"rate":       s.config.Rate,
				"window":     s.config.Window.String(),
			})
		return false, nil
	}

	return true, nil
}
