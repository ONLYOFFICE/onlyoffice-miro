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
	"strconv"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
)

type RateLimitConfig struct {
	Rate      int           `yaml:"rate" env:"RATE_LIMIT_RATE" validate:"gt=0"`
	Window    time.Duration `yaml:"window" env:"RATE_LIMIT_WINDOW" validate:"gt=0"`
	SkipPaths []string      `yaml:"skip_paths" env:"RATE_LIMIT_SKIP_PATHS"`
}

func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Rate:      100,
		Window:    time.Minute,
		SkipPaths: []string{"/health", "/metrics"},
	}
}

func (c *RateLimitConfig) loadEnv() error {
	if rate := os.Getenv("RATE_LIMIT_RATE"); rate != "" {
		if r, err := strconv.Atoi(rate); err == nil {
			c.Rate = r
		}
	}

	if window := os.Getenv("RATE_LIMIT_WINDOW"); window != "" {
		if w, err := time.ParseDuration(window); err == nil {
			c.Window = w
		}
	}

	if skipPaths := os.Getenv("RATE_LIMIT_SKIP_PATHS"); skipPaths != "" {
		c.SkipPaths = strings.Split(skipPaths, ",")
	}

	return nil
}

func (c *RateLimitConfig) Validate() error {
	validate := validator.New()

	if err := validate.Struct(c); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				switch e.Field() {
				case "Rate":
					return fmt.Errorf("rate limit rate must be greater than 0")
				case "Window":
					return fmt.Errorf("rate limit window must be greater than 0")
				default:
					return fmt.Errorf("validation error on field %s: %s", e.Field(), e.Tag())
				}
			}
		}

		return err
	}

	return nil
}
