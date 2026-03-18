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
	"context"
	"io"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/config"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/deployments"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/crypto"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/docserver"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/miro"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/oauth"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/controller"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/controller/auth"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/controller/callback"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/controller/editor"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/controller/file"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/controller/settings"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/cache"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/document"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/logger"
	oauthService "github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/oauth"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/processor"
	settingsService "github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/settings"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/storage/pg"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/translation"
	echo "github.com/labstack/echo/v4"
	fx "go.uber.org/fx"
)

// Module defines the application dependency graph using fx.
// Components are initialized in the following order:
// 1. Infrastructure (Logger, Database)
// 2. External clients
// 3. Application services
// 4. Web Framework components
// 5. Application container
var Module = fx.Options(
	fx.Provide(
		// Infrastructure layer - fundamental services
		NewLogger,   // Base logging service first
		NewDatabase, // Database connection and storage services
		NewCache,    // Caching service

		// External clients layer
		NewClients, // External API client services (Miro, OAuth, DocServer)

		// Application services layer - business logic
		NewServices, // Core application services

		// Web layer - HTTP components
		NewControllers, // Controller layer for handling requests
		NewEcho,        // Web framework setup
		NewRouter,      // Routing configuration

		// Application container - wires everything together
		NewApp, // Main application container
	),
)

//
// INFRASTRUCTURE LAYER
//

// NewLogger creates a logger service based on configuration.
// This should be initialized first as it's used by most other components.
func NewLogger(config *config.LoggerConfig) service.Logger {
	switch config.LoggerType {
	case "zap":
		return logger.NewZapLogger(config)
	default:
		return logger.NewNoopLogger()
	}
}

// NewCache creates a caching service.
// It initializes a Redis-based cache with the application configuration.
func NewCache(config *config.Config, logger service.Logger) (service.Cache, error) {
	// Create a Redis cache with default options
	return cache.NewRedisCache(
		config.Redis,
		logger,
		cache.WithKeyPrefix("app:cache:"),
		cache.WithDefaultExpiration(5*time.Minute),
	)
}

// NewDatabase initializes the database connection pool and storage services.
// It handles database migration and creates storage repositories.
func NewDatabase(config *config.Config, logger service.Logger) (*Database, error) {
	pool, err := pg.NewPostgresPool(config.Database.DatasourceURL())
	if err != nil {
		return nil, err
	}

	migrator, err := deployments.NewMigrator(pool, config.Database)
	if err != nil {
		return nil, err
	}

	defer migrator.Close()
	if err := migrator.Up(); err != nil {
		return nil, err
	}

	authStorage, err := pg.NewPostgresStorage(pool, processor.NewAuthenticationProcessor(), logger)
	if err != nil {
		return nil, err
	}

	settingsStorage, err := pg.NewPostgresStorage(pool, processor.NewSettingsProcessor(), logger)
	if err != nil {
		return nil, err
	}

	return &Database{
		Pool:            pool,
		AuthStorage:     authStorage,
		SettingsStorage: settingsStorage,
	}, nil
}

//
// EXTERNAL CLIENTS LAYER
//

// NewClients initializes external API clients.
// These are used to communicate with external services like Miro.
func NewClients(config *config.Config, logger service.Logger) (*Clients, error) {
	oauthClient, err := oauth.NewOAuthClient[miro.AuthenticationResponse](config.OAuth, logger)
	if err != nil {
		return nil, err
	}

	return &Clients{
		DocServer:   docserver.NewClient(logger),
		MiroClient:  miro.NewMiroClient(config.Miro, logger),
		OAuthClient: oauthClient,
	}, nil
}

//
// APPLICATION SERVICES LAYER
//

// NewServices initializes all application business logic services.
// These build upon the database and clients to provide core functionality.
func NewServices(
	config *config.Config,
	database *Database,
	clients *Clients,
	cache service.Cache,
	logger service.Logger,
) (*Services, error) {
	mapper := NewAuthenticationMapper()
	cipher := crypto.NewAESCipher([]byte(config.OAuth.EncryptionSecret))
	jwt := crypto.NewJwtService()

	renderer, err := controller.NewTemplateRenderer(logger)
	if err != nil {
		return nil, err
	}

	formatManager, err := document.NewMapFormatManager()
	if err != nil {
		return nil, err
	}

	builder := document.NewBuilderService(
		document.NewModificationKeyGenerator(logger),
		document.NewJwtSignatureGenerator(logger),
		formatManager,
		logger,
	)

	authService := oauthService.NewOAuthService(
		cipher,
		clients.OAuthClient,
		mapper,
		database.AuthStorage,
		logger,
	)

	settingsService := settingsService.NewSettingsService(
		config,
		clients.DocServer,
		cache,
		cipher,
		jwt,
		database.SettingsStorage,
		logger,
	)

	translator, err := translation.NewTranslation("en", logger)
	if err != nil {
		return nil, err
	}

	return &Services{
		AuthService:     authService,
		SettingsService: settingsService,
		JwtService:      jwt,
		Builder:         builder,
		FormatManager:   formatManager,
		Renderer:        &renderer,
		Translator:      translator,
	}, nil
}

//
// WEB LAYER - CONTROLLERS AND FRAMEWORK
//

// NewControllers initializes HTTP request handlers.
// These depend on services for processing business logic.
func NewControllers(
	config *config.Config,
	clients *Clients,
	services *Services,
	logger service.Logger,
) (*Controllers, error) {
	editor := editor.NewEditorController(
		config,
		clients.MiroClient,
		services.JwtService,
		services.Builder,
		services.AuthService,
		services.SettingsService,
		services.Translator,
		logger,
	)

	auth := auth.NewAuthController(
		config,
		clients.OAuthClient,
		services.AuthService,
		logger,
	)

	callback := callback.NewCallbackController(
		config,
		clients.MiroClient,
		services.JwtService,
		services.AuthService,
		services.SettingsService,
		logger,
	)

	settings := settings.NewSettingsController(
		clients.MiroClient,
		services.SettingsService,
		services.AuthService,
		4*time.Second,
		logger,
	)

	fileManagement := file.NewFileManagementController(
		config,
		clients.MiroClient,
		services.JwtService,
		services.Builder,
		services.AuthService,
		services.SettingsService,
		services.Translator,
		logger,
	)

	fileConversion := file.NewFileConversionController(
		config,
		clients.MiroClient,
		clients.DocServer,
		services.JwtService,
		services.Builder,
		services.AuthService,
		services.SettingsService,
		services.Translator,
		logger,
	)

	return &Controllers{
		Editor:         editor,
		Auth:           auth,
		Callback:       callback,
		Settings:       settings,
		FileManagement: fileManagement,
		FileConversion: fileConversion,
	}, nil
}

// NewEcho initializes the Echo web framework.
// It disables unnecessary default logging and sets up required components.
func NewEcho(services *Services) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger.SetOutput(io.Discard)
	e.Renderer = services.Renderer
	return e
}

// NewRouter creates the routing configuration for the application.
// This component only creates the router instance - route setup happens in App.SetupRoutes.
func NewRouter(
	echo *echo.Echo,
	config *config.Config,
	services *Services,
) *Router {
	return &Router{
		Echo:     echo,
		Config:   config,
		Services: services,
	}
}

//
// APPLICATION CONTAINER
//

// NewApp creates the main application container and wires together all components.
// It registers lifecycle hooks to handle graceful startup and shutdown.
func NewApp(
	lifecycle fx.Lifecycle,
	config *config.Config,
	echo *echo.Echo,
	router *Router,
	database *Database,
	services *Services,
	clients *Clients,
	controllers *Controllers,
	logger service.Logger,
) *App {
	app := &App{
		Echo:        echo,
		Router:      router,
		Database:    database,
		Services:    services,
		Controllers: controllers,
		Clients:     clients,
		Config:      config,
	}

	app.SetupRoutes(logger)

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return echo.Shutdown(ctx)
		},
	})

	return app
}
