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
	"io"
	"os"
)

type Level string

const (
	Debug Level = "debug"
	Info  Level = "info"
	Warn  Level = "warn"
	Error Level = "error"
	Fatal Level = "fatal"
)

type LoggerConfig struct {
	ServiceName string    `yaml:"service_name"`
	Environment string    `yaml:"environment"`
	Level       string    `yaml:"level"`
	PrettyPrint bool      `yaml:"pretty_print"`
	LoggerType  string    `yaml:"logger_type"`
	Output      io.Writer `yaml:"-"`
}

func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		ServiceName: "onlyoffice-miro-service",
		Environment: "development",
		Level:       "info",
		PrettyPrint: false,
		LoggerType:  "zap",
		Output:      os.Stdout,
	}
}

func (c *LoggerConfig) loadEnv() error {
	if val, exists := os.LookupEnv("LOGGER_SERVICE_NAME"); exists {
		c.ServiceName = val
	}

	if val, exists := os.LookupEnv("LOGGER_ENVIRONMENT"); exists {
		c.Environment = val
	}

	if val, exists := os.LookupEnv("LOGGER_LEVEL"); exists {
		c.Level = val
	}

	if val, exists := os.LookupEnv("LOGGER_PRETTY_PRINT"); exists {
		c.PrettyPrint = val == "true"
	}

	if val, exists := os.LookupEnv("LOGGER_TYPE"); exists {
		c.LoggerType = val
	}

	return nil
}

func (c *LoggerConfig) Validate() error {
	return nil
}

func (c *LoggerConfig) ToLogLevel() Level {
	switch c.Level {
	case "debug":
		return Debug
	case "info":
		return Info
	case "warn":
		return Warn
	case "error":
		return Error
	case "fatal":
		return Fatal
	default:
		return Info
	}
}
