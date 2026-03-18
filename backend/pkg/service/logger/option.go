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
package logger

import (
	"io"
	"os"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/config"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
)

type Option func(*config.LoggerConfig)

func WithServiceName(name string) Option {
	return func(c *config.LoggerConfig) {
		c.ServiceName = name
	}
}

func WithEnvironment(env string) Option {
	return func(c *config.LoggerConfig) {
		c.Environment = env
	}
}

func WithLevel(level config.Level) Option {
	return func(c *config.LoggerConfig) {
		c.Level = string(level)
	}
}

func WithPrettyPrint(enabled bool) Option {
	return func(c *config.LoggerConfig) {
		c.PrettyPrint = enabled
	}
}

func WithOutput(w io.Writer) Option {
	return func(c *config.LoggerConfig) {
		c.Output = w
	}
}

func NewLoggerWithOptions(loggerType string, opts ...Option) service.Logger {
	config := config.DefaultLoggerConfig()

	for _, opt := range opts {
		opt(config)
	}

	if config.Output == nil {
		config.Output = os.Stdout
	}

	switch loggerType {
	case "zap":
		return NewZapLogger(config)
	case "noop":
		return NewNoopLogger()
	default:
		return NewZapLogger(config)
	}
}
