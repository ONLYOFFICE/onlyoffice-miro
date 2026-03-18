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
package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	echo "github.com/labstack/echo/v4"
)

type CancellationMiddleware struct {
	logger service.Logger
}

func NewCancellationMiddleware(logger service.Logger) *CancellationMiddleware {
	return &CancellationMiddleware{
		logger: logger,
	}
}

func (m *CancellationMiddleware) HandleRequestCancellation(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		ctx, cancel := context.WithTimeout(req.Context(), 30*time.Second)
		defer cancel()

		c.SetRequest(req.WithContext(ctx))

		done := make(chan error)

		go func() {
			done <- next(c)
		}()

		select {
		case <-ctx.Done():
			if ctx.Err() == context.Canceled {
				m.logger.Info(ctx, "request canceled by client",
					service.Fields{
						"path":   req.URL.Path,
						"method": req.Method,
						"error":  ctx.Err().Error(),
					})
				return echo.NewHTTPError(http.StatusRequestTimeout, "Request canceled")
			} else if ctx.Err() == context.DeadlineExceeded {
				m.logger.Info(ctx, "request timeout",
					service.Fields{
						"path":   req.URL.Path,
						"method": req.Method,
						"error":  ctx.Err().Error(),
					})
				return echo.NewHTTPError(http.StatusRequestTimeout, "Request timeout")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		case err := <-done:
			return err
		}
	}
}
