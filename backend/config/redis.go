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
package config

import (
	"fmt"
	"os"
	"time"

	validator "github.com/go-playground/validator/v10"
)

type RedisConfig struct {
	Host     string        `yaml:"host" env:"REDIS_HOST" validate:"required"`
	Port     int           `yaml:"port" env:"REDIS_PORT" validate:"gt=0"`
	Password string        `yaml:"password" env:"REDIS_PASSWORD"`
	DB       int           `yaml:"db" env:"REDIS_DB" validate:"min=0"`
	Timeout  time.Duration `yaml:"timeout" env:"REDIS_TIMEOUT" validate:"required"`
}

func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
		Timeout:  5 * time.Second,
	}
}

func (c *RedisConfig) loadEnv() error {
	if host := os.Getenv("REDIS_HOST"); host != "" {
		c.Host = host
	}

	if port := os.Getenv("REDIS_PORT"); port != "" {
		var portInt int
		if _, err := fmt.Sscanf(port, "%d", &portInt); err != nil {
			return fmt.Errorf("invalid port number: %w", err)
		}
		c.Port = portInt
	}

	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		c.Password = password
	}

	if db := os.Getenv("REDIS_DB"); db != "" {
		var dbInt int
		if _, err := fmt.Sscanf(db, "%d", &dbInt); err != nil {
			return fmt.Errorf("invalid db number: %w", err)
		}
		c.DB = dbInt
	}

	if timeout := os.Getenv("REDIS_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err != nil {
			return fmt.Errorf("invalid timeout duration: %w", err)
		} else {
			c.Timeout = duration
		}
	}

	return nil
}

func (c *RedisConfig) Validate() error {
	validate := validator.New()

	if err := validate.Struct(c); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				switch e.Field() {
				case "Host":
					return fmt.Errorf("host is required")
				case "Port":
					return fmt.Errorf("port must be positive")
				case "DB":
					return fmt.Errorf("db must be non-negative")
				case "Timeout":
					return fmt.Errorf("timeout is required")
				default:
					return fmt.Errorf("validation error on field %s: %s", e.Field(), e.Tag())
				}
			}
		}
		return err
	}

	return nil
}
