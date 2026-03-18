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
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/config"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	zap "go.uber.org/zap"
	zapcore "go.uber.org/zap/zapcore"
)

type zapLevelMap map[config.Level]zapcore.Level

var defaultZapLevelMap = zapLevelMap{
	config.Debug: zapcore.DebugLevel,
	config.Info:  zapcore.InfoLevel,
	config.Warn:  zapcore.WarnLevel,
	config.Error: zapcore.ErrorLevel,
	config.Fatal: zapcore.FatalLevel,
}

type zapLogger struct {
	logger *zap.Logger
	fields service.Fields
}

func NewZapLogger(config *config.LoggerConfig) service.Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) { enc.AppendString(t.Format(time.RFC3339Nano)) }),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if config.PrettyPrint {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	zapLevel, ok := defaultZapLevelMap[config.ToLogLevel()]
	if !ok {
		zapLevel = zapcore.InfoLevel
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(config.Output),
		zapLevel,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2))

	logger = logger.With(
		zap.String("service", config.ServiceName),
		zap.String("env", config.Environment),
	)

	return &zapLogger{
		logger: logger,
		fields: service.Fields{},
	}
}

func toZapFields(fields service.Fields) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return zapFields
}

func extractContextFields(ctx context.Context) []zap.Field {
	var fields []zap.Field

	if ctx == nil {
		return fields
	}

	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields = append(fields, zap.String("trace_id", traceID.(string)))
	}

	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, zap.String("request_id", requestID.(string)))
	}

	return fields
}

func (l *zapLogger) log(ctx context.Context, level config.Level, msg string, fields ...service.Fields) {
	contextFields := extractContextFields(ctx)

	preFields := toZapFields(l.fields)

	var mergedFields service.Fields
	if len(fields) > 0 {
		mergedFields = service.Fields{}
		for _, f := range fields {
			for k, v := range f {
				mergedFields[k] = v
			}
		}
	}

	additionalFields := toZapFields(mergedFields)

	allFields := append(contextFields, preFields...)
	allFields = append(allFields, additionalFields...)

	switch level {
	case config.Debug:
		l.logger.Debug(msg, allFields...)
	case config.Info:
		l.logger.Info(msg, allFields...)
	case config.Warn:
		l.logger.Warn(msg, allFields...)
	case config.Error:
		l.logger.Error(msg, allFields...)
	case config.Fatal:
		l.logger.Fatal(msg, allFields...)
	}
}

func (l *zapLogger) Debug(ctx context.Context, msg string, fields ...service.Fields) {
	l.log(ctx, config.Debug, msg, fields...)
}

func (l *zapLogger) Info(ctx context.Context, msg string, fields ...service.Fields) {
	l.log(ctx, config.Info, msg, fields...)
}

func (l *zapLogger) Warn(ctx context.Context, msg string, fields ...service.Fields) {
	l.log(ctx, config.Warn, msg, fields...)
}

func (l *zapLogger) Error(ctx context.Context, msg string, fields ...service.Fields) {
	l.log(ctx, config.Error, msg, fields...)
}

func (l *zapLogger) Fatal(ctx context.Context, msg string, fields ...service.Fields) {
	l.log(ctx, config.Fatal, msg, fields...)
}

func (l *zapLogger) WithFields(fields service.Fields) service.Logger {
	newFields := service.Fields{}

	for k, v := range l.fields {
		newFields[k] = v
	}

	for k, v := range fields {
		newFields[k] = v
	}

	return &zapLogger{
		logger: l.logger,
		fields: newFields,
	}
}

func (l *zapLogger) WithContext(ctx context.Context) service.Logger {
	fields := service.Fields{}

	if ctx != nil {
		if traceID := ctx.Value("trace_id"); traceID != nil {
			fields["trace_id"] = traceID
		}

		if requestID := ctx.Value("request_id"); requestID != nil {
			fields["request_id"] = requestID
		}
	}

	return l.WithFields(fields)
}
