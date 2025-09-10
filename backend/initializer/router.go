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
package initializer

import (
	"log"
	"net/http"
	"strings"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/assets"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/common"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/middleware"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/middleware/authentication"
	echo "github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

// SetupRoutes configures all application routes and middleware.
// This is extracted to a separate file for better organization.
func (r *Router) SetupRoutes(
	controllers *Controllers,
	logger service.Logger,
) {
	// Configure global middleware
	setupGlobalMiddleware(r, logger)

	// Setup custom error handler
	setupErrorHandler(r, logger)

	// Setup authentication middleware
	authMiddleware, miroAuthMiddleware, editorMiddleware := setupAuthMiddleware(r, logger)

	// Setup routes by category
	setupEditorRoutes(r, controllers, editorMiddleware)
	setupCallbackRoutes(r, controllers)
	setupAuthRoutes(r, controllers)
	setupProtectedRoutes(r, controllers, authMiddleware)
	setupMiroAuthRoutes(r, miroAuthMiddleware)
	setupFileStoreRoutes(r)
}

// setupGlobalMiddleware configures global middleware for all routes
func setupGlobalMiddleware(r *Router, logger service.Logger) {
	// Add cancellation middleware first to handle client disconnections
	cancellationMiddleware := middleware.NewCancellationMiddleware(logger)
	r.Echo.Use(cancellationMiddleware.HandleRequestCancellation)

	// Basic panic recovery middleware
	r.Echo.Use(echomiddleware.Recover())

	// CORS configuration
	r.Echo.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins:     r.Config.CORS.AllowOrigins,
		AllowHeaders:     r.Config.CORS.AllowHeaders,
		AllowMethods:     r.Config.CORS.AllowMethods,
		AllowCredentials: r.Config.CORS.AllowCredentials,
		MaxAge:           r.Config.CORS.MaxAge,
	}))

	// Rate limiting
	store, err := middleware.NewRedisStore(r.Config, logger)
	if err != nil {
		log.Fatalf("failed to initialize Redis store: %v", err)
		return
	}

	r.Echo.Use(echomiddleware.RateLimiterWithConfig(echomiddleware.RateLimiterConfig{
		Skipper: echomiddleware.DefaultSkipper,
		Store:   store,
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			return ctx.RealIP(), nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return &echo.HTTPError{
				Code:     echomiddleware.ErrRateLimitExceeded.Code,
				Message:  echomiddleware.ErrRateLimitExceeded.Message,
				Internal: err,
			}
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return &echo.HTTPError{
				Code:     echomiddleware.ErrRateLimitExceeded.Code,
				Message:  echomiddleware.ErrRateLimitExceeded.Message,
				Internal: err,
			}
		},
	}))
}

func setupErrorHandler(r *Router, logger service.Logger) {
	r.Echo.HTTPErrorHandler = func(err error, c echo.Context) {
		var code int

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		} else {
			c.Echo().DefaultHTTPErrorHandler(err, c)
			return
		}

		if code == http.StatusNotFound {
			logger.Debug(c.Request().Context(), "404 Not Found - redirecting to miro.com", service.Fields{
				"path":   c.Request().URL.Path,
				"method": c.Request().Method,
			})

			if !c.Response().Committed {
				_ = c.Redirect(http.StatusFound, "https://miro.com")
			}
			return
		}

		c.Echo().DefaultHTTPErrorHandler(err, c)
	}
}

// setupAuthMiddleware creates and configures auth middleware instances
func setupAuthMiddleware(r *Router, logger service.Logger) (
	*authentication.AuthMiddleware,
	*authentication.AuthMiddleware,
	*authentication.AuthMiddleware,
) {
	authMiddleware := authentication.NewTokenAuthMiddleware(
		r.Config,
		r.Services.JwtService,
		r.Services.Translator,
		logger,
	)

	miroAuthMiddleware := authentication.NewMiroAuthMiddleware(
		r.Config,
		r.Services.JwtService,
		r.Services.Translator,
		logger,
	)

	editorMiddleware := authentication.NewEditorAuthMiddleware(
		r.Config,
		r.Services.AuthService,
		r.Services.JwtService,
		r.Services.Translator,
		logger,
	)

	return authMiddleware, miroAuthMiddleware, editorMiddleware
}

// setupEditorRoutes configures editor-related routes
func setupEditorRoutes(r *Router, controllers *Controllers, editorMiddleware *authentication.AuthMiddleware) {
	handlers := controllers.Editor.Handlers()
	r.Echo.GET("/api/editor", editorMiddleware.Authenticate(handlers[common.MethodGet]))
}

// setupCallbackRoutes configures callback-related routes
func setupCallbackRoutes(r *Router, controllers *Controllers) {
	handlers := controllers.Callback.Handlers()
	r.Echo.POST("/api/callback", handlers[common.MethodPost])
}

// setupAuthRoutes configures authentication-related routes
func setupAuthRoutes(r *Router, controllers *Controllers) {
	handlers := controllers.Auth.Handlers()
	r.Echo.GET("/api/oauth", handlers[common.MethodGet])
}

// setupProtectedRoutes configures routes that require authentication
func setupProtectedRoutes(r *Router, controllers *Controllers, authMiddleware *authentication.AuthMiddleware) {
	protected := r.Echo.Group("/api")
	protected.Use(authMiddleware.Authenticate)

	// Settings routes
	handlers := controllers.Settings.Handlers()
	protected.GET("/settings", handlers[common.MethodGet])
	protected.POST("/settings", handlers[common.MethodPost])

	// File management routes
	handlers = controllers.FileManagement.Handlers()
	protected.GET("/files", handlers[common.MethodGet])
	protected.POST("/files/create", handlers[common.MethodPost])

	// File conversion routes
	handlers = controllers.FileConversion.Handlers()
	protected.GET("/files/convert", handlers[common.MethodGet])
}

// setupMiroAuthRoutes configures Miro-specific authentication routes
func setupMiroAuthRoutes(r *Router, miroAuthMiddleware *authentication.AuthMiddleware) {
	r.Echo.GET("/api/authorize", miroAuthMiddleware.Authenticate(miroAuthMiddleware.GetCookieExpiration))
}

// setupFileStoreRoutes configures file store routes to serve embedded assets
func setupFileStoreRoutes(r *Router) {
	r.Echo.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "UP"})
	})

	r.Echo.GET("/filestore/*", func(c echo.Context) error {
		reqPath := c.Param("*")
		if strings.HasPrefix(reqPath, "icons/") {
			fileData, err := assets.Icons.ReadFile(reqPath)
			if err != nil {
				return c.NoContent(http.StatusNotFound)
			}

			if strings.HasSuffix(reqPath, ".svg") {
				return c.Blob(http.StatusOK, "image/svg+xml", fileData)
			}

			contentType := http.DetectContentType(fileData)
			return c.Blob(http.StatusOK, contentType, fileData)
		}

		return c.NoContent(http.StatusNotFound)
	})
}
