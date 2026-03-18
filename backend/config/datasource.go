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

	validator "github.com/go-playground/validator/v10"
)

type DataSourceConfig struct {
	Host     string `yaml:"host" env:"DB_HOST" validate:"required"`
	Port     int    `yaml:"port" env:"DB_PORT" validate:"gt=0"`
	User     string `yaml:"user" env:"DB_USER" validate:"required"`
	Password string `yaml:"password" env:"DB_PASSWORD" validate:"required"`
	Database string `yaml:"database" env:"DB_NAME" validate:"required"`
}

func DefaultDataSourceConfig() *DataSourceConfig {
	return &DataSourceConfig{
		Host:     "localhost",
		Port:     6432,
		User:     "admin",
		Password: "admin",
		Database: "miro",
	}
}

func (c *DataSourceConfig) loadEnv() error {
	if host := os.Getenv("DB_HOST"); host != "" {
		c.Host = host
	}

	if port := os.Getenv("DB_PORT"); port != "" {
		var portInt int
		if _, err := fmt.Sscanf(port, "%d", &portInt); err != nil {
			return fmt.Errorf("invalid port number: %w", err)
		}

		c.Port = portInt
	}

	if user := os.Getenv("DB_USER"); user != "" {
		c.User = user
	}

	if password := os.Getenv("DB_PASSWORD"); password != "" {
		c.Password = password
	}

	if database := os.Getenv("DB_NAME"); database != "" {
		c.Database = database
	}

	return nil
}

func (c *DataSourceConfig) Validate() error {
	validate := validator.New()

	if err := validate.Struct(c); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				switch e.Field() {
				case "Host":
					return fmt.Errorf("host is required")
				case "Port":
					return fmt.Errorf("port must be positive")
				case "User":
					return fmt.Errorf("user is required")
				case "Password":
					return fmt.Errorf("password is required")
				case "Database":
					return fmt.Errorf("database name is required")
				default:
					return fmt.Errorf("validation error on field %s: %s", e.Field(), e.Tag())
				}
			}
		}

		return err
	}

	return nil
}

func (c *DataSourceConfig) DatasourceURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", c.User, c.Password, c.Host, c.Port, c.Database)
}
