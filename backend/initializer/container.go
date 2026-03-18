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
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/config"
	core "github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/core"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/core/component"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/crypto"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/docserver"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/miro"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/oauth"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/common"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/controller"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/document"
	oauthService "github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/oauth"
	settingsService "github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/settings"
	pgxpool "github.com/jackc/pgx/v5/pgxpool"
	echo "github.com/labstack/echo/v4"
)

var _ core.AuthCompositeKey

// App is the main application container that holds all dependencies.
type App struct {
	Config      *config.Config
	Clients     *Clients
	Controllers *Controllers
	Database    *Database
	Echo        *echo.Echo
	Router      *Router
	Services    *Services
}

// SetupRoutes calls the router's route setup method.
func (a *App) SetupRoutes(logger service.Logger) {
	// Route setup is now implemented in router.go
	a.Router.SetupRoutes(a.Controllers, logger)
}

// Database holds all database-related components and repositories.
type Database struct {
	Pool            *pgxpool.Pool
	AuthStorage     service.Storage[core.AuthCompositeKey, component.Authentication]
	SettingsStorage service.Storage[core.SettingsCompositeKey, component.Settings]
}

// Clients contains all external API client instances.
type Clients struct {
	DocServer   docserver.Client
	MiroClient  miro.Client
	OAuthClient oauth.OAuthClient[miro.AuthenticationResponse]
}

// Services contains all application service instances.
type Services struct {
	AuthService     oauthService.OAuthService[miro.AuthenticationResponse]
	Builder         document.BuilderService
	FormatManager   document.FormatManager
	JwtService      crypto.Signer
	Renderer        *controller.TemplateRenderer
	SettingsService settingsService.SettingsService
	Translator      service.TranslationProvider
}

// Controllers contains all HTTP request handlers.
type Controllers struct {
	Auth           common.Handler
	Callback       common.Handler
	Editor         common.Handler
	FileConversion common.Handler
	FileManagement common.Handler
	Settings       common.Handler
}

// Router provides access to the Echo instance and configuration.
type Router struct {
	Config   *config.Config
	Echo     *echo.Echo
	Services *Services
}
