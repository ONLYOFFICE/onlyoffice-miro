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
	"strings"

	validator "github.com/go-playground/validator/v10"
)

type CORSConfig struct {
	AllowOrigins     []string `yaml:"allow_origins" env:"CORS_ALLOW_ORIGINS" validate:"required,min=1"`
	AllowHeaders     []string `yaml:"allow_headers" env:"CORS_ALLOW_HEADERS" validate:"required,min=1"`
	AllowMethods     []string `yaml:"allow_methods" env:"CORS_ALLOW_METHODS" validate:"required,min=1"`
	AllowCredentials bool     `yaml:"allow_credentials" env:"CORS_ALLOW_CREDENTIALS"`
	MaxAge           int      `yaml:"max_age" env:"CORS_MAX_AGE" validate:"min=0"`
}

func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins: []string{
			"*",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Miro-Signature",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},
		AllowCredentials: true,
		MaxAge:           300,
	}
}

func (c *CORSConfig) loadEnv() error {
	if origins := os.Getenv("CORS_ALLOW_ORIGINS"); origins != "" {
		c.AllowOrigins = strings.Split(origins, ",")
	}

	if headers := os.Getenv("CORS_ALLOW_HEADERS"); headers != "" {
		c.AllowHeaders = strings.Split(headers, ",")
	}

	if methods := os.Getenv("CORS_ALLOW_METHODS"); methods != "" {
		c.AllowMethods = strings.Split(methods, ",")
	}

	if credentials := os.Getenv("CORS_ALLOW_CREDENTIALS"); credentials != "" {
		c.AllowCredentials = credentials == "true"
	}

	if maxAge := os.Getenv("CORS_MAX_AGE"); maxAge != "" {
		var age int
		if _, err := fmt.Sscanf(maxAge, "%d", &age); err != nil {
			return fmt.Errorf("invalid max age value: %w", err)
		}
		c.MaxAge = age
	}

	return nil
}

func (c *CORSConfig) Validate() error {
	validate := validator.New()

	if err := validate.Struct(c); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				switch e.Field() {
				case "AllowOrigins":
					return fmt.Errorf("at least one origin must be specified")
				case "AllowHeaders":
					return fmt.Errorf("at least one header must be specified")
				case "AllowMethods":
					return fmt.Errorf("at least one method must be specified")
				case "MaxAge":
					return fmt.Errorf("max age must be non-negative")
				default:
					return fmt.Errorf("validation error on field %s: %s", e.Field(), e.Tag())
				}
			}
		}

		return err
	}

	return nil
}
