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
package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/config"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/core/component"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/miro"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/oauth"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/common"
	oauthService "github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/oauth"
	echo "github.com/labstack/echo/v4"
)

type authController struct {
	installationLocation string
	oauthClient          oauth.OAuthClient[miro.AuthenticationResponse]
	oauthService         oauthService.OAuthService[miro.AuthenticationResponse]
	logger               service.Logger
}

func NewAuthController(
	config *config.Config,
	oauthClient oauth.OAuthClient[miro.AuthenticationResponse],
	oauthService oauthService.OAuthService[miro.AuthenticationResponse],
	logger service.Logger,
) common.Handler {
	controller := &authController{
		installationLocation: fmt.Sprintf(
			"%s?response_type=code&client_id=%s&redirect_uri=%s",
			miroInstallationBase, config.OAuth.ClientID, config.OAuth.RedirectURI,
		),
		oauthClient:  oauthClient,
		oauthService: oauthService,
		logger:       logger,
	}

	logger.Info(context.Background(), "Auth controller initialized", service.Fields{
		"redirect_uri": config.OAuth.RedirectURI,
		"client_id":    config.OAuth.ClientID,
	})

	return common.NewHandler(map[common.HTTPMethod]echo.HandlerFunc{
		common.MethodGet: controller.handleGet,
	})
}

func (c *authController) extractParams(ctx echo.Context) (authQueryParams, error) {
	c.logger.Debug(ctx.Request().Context(), "Extractig auth query params")

	params := authQueryParams{
		Code: ctx.QueryParam("code"),
	}

	if params.Code == "" {
		c.logger.Warn(ctx.Request().Context(), "Missing authorization code")
		return params, fmt.Errorf("missing authorization code")
	}

	c.logger.Debug(ctx.Request().Context(), "Successfully extracted auth code")
	return params, nil
}

func (c *authController) handleError(ctx echo.Context, msg string, err error, args ...any) error {
	fields := service.Fields{"error": err.Error()}
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			fields[args[i].(string)] = args[i+1]
		}
	}

	c.logger.Error(ctx.Request().Context(), msg, fields)
	return ctx.Render(http.StatusBadRequest, "exchange", map[string]any{
		"installationLocation": c.installationLocation,
	})
}

func (c *authController) handleGet(ctx echo.Context) error {
	requestID := ctx.Response().Header().Get(echo.HeaderXRequestID)
	c.logger.Info(ctx.Request().Context(), "Handling auth request", service.Fields{
		"method":      ctx.Request().Method,
		"remote_addr": ctx.Request().RemoteAddr,
		"request_id":  requestID,
		"user_agent":  ctx.Request().UserAgent(),
	})

	tctx, cancel := context.WithTimeout(ctx.Request().Context(), 3*time.Second)
	defer cancel()

	params, err := c.extractParams(ctx)
	if err != nil {
		return c.handleError(ctx, "Failed to extract authorization code", err)
	}

	c.logger.Info(tctx, "Exchanging authorization code for token")
	token, err := c.oauthClient.Exchange(tctx, params.Code)
	if err != nil {
		return c.handleError(ctx, "Failed to exchange authorization code", err,
			"code", params.Code)
	}

	c.logger.Debug(tctx, "Token exchange successful", service.Fields{
		"user_id":    token.UserID,
		"team_id":    token.TeamID,
		"expires_in": token.ExpiresIn,
		"token_type": token.TokenType,
		"scope":      token.Scope,
	})

	expiresAt := time.Now().Add(time.Second * time.Duration(token.ExpiresIn-10)).Unix()
	auth := component.Authentication{
		TokenType:    token.TokenType,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    int(expiresAt),
		Scope:        token.Scope,
	}

	c.logger.Info(tctx, "Saving authentication token", service.Fields{
		"user_id":    token.UserID,
		"team_id":    token.TeamID,
		"expires_at": expiresAt,
		"scope":      token.Scope,
	})

	if err := c.oauthService.Save(tctx, token.TeamID, token.UserID, auth); err != nil {
		return c.handleError(ctx, "Failed to persist authentication token", err,
			"user_id", token.UserID,
			"team_id", token.TeamID)
	}

	c.logger.Info(ctx.Request().Context(), "Successfully authenticated user",
		service.Fields{
			"user_id": token.UserID,
			"team_id": token.TeamID,
		},
	)

	c.logger.Debug(ctx.Request().Context(), "Redirecting to Miro application",
		service.Fields{
			"redirect_url": miroApplicationBase,
			"status_code":  http.StatusPermanentRedirect,
		},
	)

	return ctx.Redirect(http.StatusPermanentRedirect, miroApplicationBase)
}
