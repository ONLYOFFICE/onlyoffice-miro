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
package service

import "context"

type Fields map[string]any

type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Fields)
	Info(ctx context.Context, msg string, fields ...Fields)
	Warn(ctx context.Context, msg string, fields ...Fields)
	Error(ctx context.Context, msg string, fields ...Fields)
	Fatal(ctx context.Context, msg string, fields ...Fields)
	WithFields(fields Fields) Logger
	WithContext(ctx context.Context) Logger
}
