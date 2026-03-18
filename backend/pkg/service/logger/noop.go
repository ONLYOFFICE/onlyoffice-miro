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
	"context"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
)

type NoopLogger struct{}

func NewNoopLogger() service.Logger {
	return &NoopLogger{}
}

func (l *NoopLogger) Debug(_ context.Context, _ string, _ ...service.Fields) {}
func (l *NoopLogger) Info(_ context.Context, _ string, _ ...service.Fields)  {}
func (l *NoopLogger) Warn(_ context.Context, _ string, _ ...service.Fields)  {}
func (l *NoopLogger) Error(_ context.Context, _ string, _ ...service.Fields) {}
func (l *NoopLogger) Fatal(_ context.Context, _ string, _ ...service.Fields) {}

func (l *NoopLogger) WithFields(_ service.Fields) service.Logger {
	return l
}

func (l *NoopLogger) WithContext(_ context.Context) service.Logger {
	return l
}
