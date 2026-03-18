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
	"time"

	validator "github.com/go-playground/validator/v10"
)

type OAuthConfig struct {
	ClientID         string        `yaml:"client_id" env:"OAUTH_CLIENT_ID" validate:"required"`
	ClientSecret     string        `yaml:"client_secret" env:"OAUTH_CLIENT_SECRET" validate:"required"`
	EncryptionSecret string        `yaml:"encryption_secret" env:"OAUTH_ENCRYPTION_SECRET" validate:"required"`
	RedirectURI      string        `yaml:"redirect_uri" env:"OAUTH_REDIRECT_URI" validate:"required,http_address"`
	TokenURI         string        `yaml:"token_uri" env:"OAUTH_TOKEN_URI" validate:"required,http_address"`
	Timeout          time.Duration `yaml:"timeout" env:"OAUTH_TIMEOUT" validate:"required"`
}

func DefaultOAuthConfig() *OAuthConfig {
	return &OAuthConfig{
		TokenURI: "https://api.miro.com/v1/oauth/token",
		Timeout:  4 * time.Second,
	}
}

func (c *OAuthConfig) loadEnv() error {
	if clientID := os.Getenv("OAUTH_CLIENT_ID"); clientID != "" {
		c.ClientID = clientID
	}

	if clientSecret := os.Getenv("OAUTH_CLIENT_SECRET"); clientSecret != "" {
		c.ClientSecret = clientSecret
	}

	if encryptionSecret := os.Getenv("OAUTH_ENCRYPTION_SECRET"); encryptionSecret != "" {
		c.EncryptionSecret = encryptionSecret
	}

	if redirectURI := os.Getenv("OAUTH_REDIRECT_URI"); redirectURI != "" {
		c.RedirectURI = redirectURI
	}

	if tokenURI := os.Getenv("OAUTH_TOKEN_URI"); tokenURI != "" {
		c.TokenURI = tokenURI
	}

	if timeout := os.Getenv("OAUTH_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err != nil {
			return fmt.Errorf("invalid timeout duration: %w", err)
		} else {
			c.Timeout = duration
		}
	}

	return nil
}

func (c *OAuthConfig) Validate() error {
	validate := validator.New()

	validate.RegisterValidation("http_address", func(fl validator.FieldLevel) bool {
		uri := fl.Field().String()
		if !strings.HasPrefix(uri, "http://") && !strings.HasPrefix(uri, "https://") {
			return false
		}

		if strings.HasSuffix(uri, "/") {
			return false
		}

		return true
	})

	if err := validate.Struct(c); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				switch e.Field() {
				case "ClientID":
					return fmt.Errorf("client_id is required")
				case "ClientSecret":
					return fmt.Errorf("client_secret is required")
				case "EncryptionSecret":
					return fmt.Errorf("encryption_secret is required")
				case "RedirectURI":
					if e.Tag() == "http_address" {
						return fmt.Errorf("redirect_uri must be an HTTP/HTTPS URL without trailing slash")
					}

					return fmt.Errorf("redirect_uri is required")
				case "TokenURI":
					if e.Tag() == "http_address" {
						return fmt.Errorf("token_uri must be an HTTP/HTTPS URL without trailing slash")
					}

					return fmt.Errorf("token_uri is required")
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
