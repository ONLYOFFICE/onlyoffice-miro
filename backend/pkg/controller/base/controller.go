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
package base

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/config"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/core/component"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/crypto"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/miro"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/common"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/middleware/authentication"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/document"
	oauthService "github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/oauth"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/settings"
	echo "github.com/labstack/echo/v4"
	errgroup "golang.org/x/sync/errgroup"
)

type BaseController struct {
	*common.BaseHandler
	Config             *config.Config
	MiroClient         miro.Client
	JwtService         crypto.Signer
	BuilderService     document.BuilderService
	OAuthService       oauthService.OAuthService[miro.AuthenticationResponse]
	SettingsService    settings.SettingsService
	TranslationService service.TranslationProvider
	Logger             service.Logger
}

func NewBaseController(
	config *config.Config,
	miroClient miro.Client,
	jwtService crypto.Signer,
	builderService document.BuilderService,
	oauthService oauthService.OAuthService[miro.AuthenticationResponse],
	settingsService settings.SettingsService,
	translationService service.TranslationProvider,
	logger service.Logger,
) *BaseController {
	return &BaseController{
		BaseHandler:        &common.BaseHandler{},
		Config:             config,
		MiroClient:         miroClient,
		JwtService:         jwtService,
		BuilderService:     builderService,
		OAuthService:       oauthService,
		SettingsService:    settingsService,
		TranslationService: translationService,
		Logger:             logger,
	}
}

func (c *BaseController) ExtractUserToken(ctx echo.Context) (*authentication.TokenClaims, error) {
	token, ok := ctx.Get(common.ContextKeyUser).(*authentication.TokenClaims)
	if !ok {
		return nil, ErrMissingOpenIdToken
	}

	return token, nil
}

func (c *BaseController) GetQueryParam(ctx echo.Context, key string) (string, error) {
	value := ctx.QueryParam(key)
	if value == "" {
		return "", fmt.Errorf("missing query parameter: %s", key)
	}

	return value, nil
}

func (c *BaseController) FetchAuthenticationWithSettings(ctx context.Context, uid, tid, bid string) (*component.Settings, *component.Authentication, error) {
	g, ctx := errgroup.WithContext(ctx)
	var settings component.Settings
	var auth component.Authentication
	var settingsErr, authErr error

	g.Go(func() error {
		var err error
		settings, err = c.SettingsService.Find(ctx, tid, bid)
		if err != nil {
			settingsErr = err
			return err
		}

		if settings.Demo.Enabled && (settings.Address == "" || settings.Header == "" || settings.Secret == "") {
			if settings.Demo.Started == nil || !settings.Demo.Started.Add(time.Duration(c.Config.DemoServer.Days)*24*time.Hour).After(time.Now()) {
				settingsErr = ErrSettingsNotConfigured
				return ErrSettingsNotConfigured
			}
		} else if settings.Address == "" && settings.Header == "" && settings.Secret == "" {
			settingsErr = ErrSettingsNotConfigured
			return ErrSettingsNotConfigured
		}

		return nil
	})

	g.Go(func() error {
		var err error
		auth, err = c.OAuthService.Find(ctx, tid, uid)
		if err != nil {
			authErr = err
			return err
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		if authErr != nil && errors.Is(authErr, oauthService.ErrTokenMissing) {
			return nil, nil, ErrMissingAuthentication
		}
		if settingsErr != nil && errors.Is(settingsErr, ErrSettingsNotConfigured) {
			return nil, nil, ErrSettingsNotConfigured
		}
		return nil, nil, err
	}

	return &settings, &auth, nil
}

func (c *BaseController) SendError(ctx echo.Context, status int, message string) error {
	return ctx.JSON(status, common.ErrorResponse{Error: message})
}

func (c *BaseController) SendJSON(ctx echo.Context, data any) error {
	return ctx.JSON(http.StatusOK, map[string]any{"data": data})
}

func (c *BaseController) HandleError(ctx echo.Context, err error, status int, message string) error {
	if err != nil {
		c.Logger.Error(ctx.Request().Context(), message, service.Fields{"error": err})
		return c.SendError(ctx, status, message)
	}
	return nil
}

func (c *BaseController) HandleWarning(ctx echo.Context, err error, status int, message string) error {
	if err != nil {
		c.Logger.Warn(ctx.Request().Context(), message, service.Fields{"error": err})
		return c.SendError(ctx, status, message)
	}
	return nil
}

func (c *BaseController) withTimeout(duration time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), duration)
}

func (c *BaseController) ExecuteWithTimeout(ctx echo.Context, duration time.Duration, fn func(context.Context) error) error {
	tctx, cancel := c.withTimeout(duration)
	defer cancel()

	return fn(tctx)
}
