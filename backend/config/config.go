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

	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	Database   *DataSourceConfig `yaml:"database"`
	Miro       *MiroConfig       `yaml:"miro"`
	OAuth      *OAuthConfig      `yaml:"oauth"`
	Server     *ServerConfig     `yaml:"server"`
	Redis      *RedisConfig      `yaml:"redis"`
	RateLimit  *RateLimitConfig  `yaml:"rate_limit"`
	CORS       *CORSConfig       `yaml:"cors"`
	DemoServer *DemoServerConfig `yaml:"demo_server"`
	Logger     *LoggerConfig     `yaml:"logger"`
}

func DefaultConfig() *Config {
	return &Config{
		Database:   DefaultDataSourceConfig(),
		Miro:       DefaultMiroConfig(),
		OAuth:      DefaultOAuthConfig(),
		Server:     DefaultServerConfig(),
		Redis:      DefaultRedisConfig(),
		RateLimit:  DefaultRateLimitConfig(),
		CORS:       DefaultCORSConfig(),
		DemoServer: DefaultDemoServerConfig(),
		Logger:     DefaultLoggerConfig(),
	}
}

func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()
	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, &config); err != nil {
				return config, fmt.Errorf("failed to parse YAML config: %w", err)
			}
		}
	}

	if err := config.Database.loadEnv(); err != nil {
		return config, fmt.Errorf("failed to load database environment variables: %w", err)
	}

	if err := config.Miro.loadEnv(); err != nil {
		return config, fmt.Errorf("failed to load Miro environment variables: %w", err)
	}

	if err := config.OAuth.loadEnv(); err != nil {
		return config, fmt.Errorf("failed to load OAuth environment variables: %w", err)
	}

	if err := config.Server.loadEnv(); err != nil {
		return config, fmt.Errorf("failed to load server environment variables: %w", err)
	}

	if err := config.Redis.loadEnv(); err != nil {
		return config, fmt.Errorf("failed to load Redis environment variables: %w", err)
	}

	if err := config.RateLimit.loadEnv(); err != nil {
		return config, fmt.Errorf("failed to load rate limit environment variables: %w", err)
	}

	if err := config.CORS.loadEnv(); err != nil {
		return config, fmt.Errorf("failed to load CORS environment variables: %w", err)
	}

	if err := config.DemoServer.loadEnv(); err != nil {
		return config, fmt.Errorf("failed to load demo server environment variables: %w", err)
	}

	if err := config.Logger.loadEnv(); err != nil {
		return config, fmt.Errorf("failed to load logger environment variables: %w", err)
	}

	return config, nil
}

func (c *Config) Validate() error {
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("invalid database config: %w", err)
	}

	if err := c.Miro.Validate(); err != nil {
		return fmt.Errorf("invalid Miro config: %w", err)
	}

	if err := c.OAuth.Validate(); err != nil {
		return fmt.Errorf("invalid OAuth config: %w", err)
	}

	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("invalid server config: %w", err)
	}

	if err := c.Redis.Validate(); err != nil {
		return fmt.Errorf("invalid Redis config: %w", err)
	}

	if err := c.RateLimit.Validate(); err != nil {
		return fmt.Errorf("invalid rate limit config: %w", err)
	}

	if err := c.CORS.Validate(); err != nil {
		return fmt.Errorf("invalid CORS config: %w", err)
	}

	if err := c.DemoServer.Validate(); err != nil {
		return fmt.Errorf("invalid demo server config: %w", err)
	}

	if err := c.Logger.Validate(); err != nil {
		return fmt.Errorf("invalid logger config: %w", err)
	}

	return nil
}
