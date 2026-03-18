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

type MiroConfig struct {
	BaseURL string        `yaml:"base_url" env:"MIRO_BASE_URL" validate:"required"`
	Timeout time.Duration `yaml:"timeout" env:"MIRO_TIMEOUT" validate:"required"`
}

func DefaultMiroConfig() *MiroConfig {
	return &MiroConfig{
		BaseURL: "https://api.miro.com/v2",
		Timeout: 15 * time.Second,
	}
}

func (c *MiroConfig) loadEnv() error {
	if baseURL := os.Getenv("MIRO_BASE_URL"); baseURL != "" {
		c.BaseURL = baseURL
	}

	if timeout := os.Getenv("MIRO_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err != nil {
			return fmt.Errorf("invalid timeout duration: %w", err)
		} else {
			c.Timeout = duration
		}
	}

	return nil
}

func (c *MiroConfig) Validate() error {
	validate := validator.New()

	if err := validate.Struct(c); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				switch e.Field() {
				case "BaseURL":
					return fmt.Errorf("base_url is required")
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
