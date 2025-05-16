package config

import (
	"fmt"
	"os"
	"strings"

	validator "github.com/go-playground/validator/v10"
)

type ServerConfig struct {
	Domain      string `yaml:"domain" env:"SERVER_DOMAIN" validate:"required"`
	CallbackURL string `yaml:"callback_url" env:"CALLBACK_URL" validate:"required,http_address"`
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Domain:      "localhost",
		CallbackURL: "http://localhost:8080/api/callback",
	}
}

func (c *ServerConfig) loadEnv() error {
	if domain := os.Getenv("SERVER_DOMAIN"); domain != "" {
		c.Domain = domain
	}

	if callbackUrl := os.Getenv("CALLBACK_URL"); callbackUrl != "" {
		c.CallbackURL = callbackUrl
	}

	return nil
}

func (c *ServerConfig) Validate() error {
	validate := validator.New()

	validate.RegisterValidation("http_address", func(fl validator.FieldLevel) bool {
		url := fl.Field().String()
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			return false
		}

		if strings.HasSuffix(url, "/") {
			return false
		}

		return true
	})

	if err := validate.Struct(c); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				switch e.Field() {
				case "Domain":
					return fmt.Errorf("server domain is required")
				case "CallbackURL":
					if e.Tag() == "http_address" {
						return fmt.Errorf("callback_url must be an HTTP/HTTPS URL without trailing slash")
					}

					return fmt.Errorf("callback_url is required")
				default:
					return fmt.Errorf("validation error on field %s: %s", e.Field(), e.Tag())
				}
			}
		}

		return err
	}

	return nil
}
