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
package callback

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/config"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/core/component"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/crypto"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/miro"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/common"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/oauth"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/settings"
	echo "github.com/labstack/echo/v4"
	errgroup "golang.org/x/sync/errgroup"
)

type callbackController struct {
	config          *config.Config
	miroClient      miro.Client
	jwtService      crypto.Signer
	oauthService    oauth.OAuthService[miro.AuthenticationResponse]
	settingsService settings.SettingsService
	logger          service.Logger
}

func NewCallbackController(
	config *config.Config,
	miroClient miro.Client,
	jwtService crypto.Signer,
	oauthService oauth.OAuthService[miro.AuthenticationResponse],
	settingsService settings.SettingsService,
	logger service.Logger,
) common.Handler {
	controller := &callbackController{
		config:          config,
		miroClient:      miroClient,
		jwtService:      jwtService,
		oauthService:    oauthService,
		settingsService: settingsService,
		logger:          logger,
	}

	logger.Info(context.Background(), "Callback controller initialized")

	return common.NewHandler(map[common.HTTPMethod]echo.HandlerFunc{
		common.MethodPost: controller.handlePost,
	})
}

func (c *callbackController) logErrorAndRespond(ctx echo.Context, statusCode int, logMessage string, err error) error {
	c.logger.Error(ctx.Request().Context(), logMessage, service.Fields{"error": err.Error()})
	return ctx.JSON(statusCode, common.ErrorResponse{Error: callbackErrorCodeFailure})
}

func (c *callbackController) extractParams(ctx echo.Context) (callbackQueryParams, error) {
	filename := ctx.QueryParam("filename")
	if filename != "" {
		if decoded, err := url.QueryUnescape(filename); err == nil {
			filename = decoded
		}
	}

	params := callbackQueryParams{
		UID:      ctx.QueryParam("uid"),
		TID:      ctx.QueryParam("tid"),
		BID:      ctx.QueryParam("bid"),
		FID:      ctx.QueryParam("fid"),
		Filename: filename,
	}

	if params.UID == "" || params.BID == "" || params.TID == "" || params.FID == "" {
		return params, fmt.Errorf("missing required query parameters")
	}

	return params, nil
}

func (c *callbackController) fetchAuthenticationAndSettings(
	ctx context.Context,
	params callbackQueryParams,
) (component.Authentication, component.Settings, error) {
	var settings component.Settings
	var auth component.Authentication

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		settings, err = c.settingsService.Find(ctx, params.TID, params.BID)
		if err != nil {
			return fmt.Errorf("failed to fetch settings: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		auth, err = c.oauthService.Find(ctx, params.TID, params.UID)
		if err != nil {
			return fmt.Errorf("failed to fetch auth: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return auth, settings, err
	}

	return auth, settings, nil
}

func (c *callbackController) getSecretFromSettings(settings component.Settings) string {
	if settings.Demo.Enabled &&
		settings.Address == "" &&
		settings.Demo.Started.Add(time.Duration(c.config.DemoServer.Days)*24*time.Hour).After(time.Now()) {
		return c.config.DemoServer.Secret
	}
	return settings.Secret
}

func (c *callbackController) handlePost(ctx echo.Context) error {
	tctx, cancel := context.WithTimeout(ctx.Request().Context(), saveFileRequestTimeout)
	defer cancel()

	params, err := c.extractParams(ctx)
	if err != nil {
		return c.logErrorAndRespond(ctx, http.StatusBadRequest, "Failed to extract query parameters", err)
	}

	var body callbackRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&body); err != nil {
		return c.logErrorAndRespond(ctx, http.StatusBadRequest, "Failed to decode request body", err)
	}

	if err := body.Validate(); err != nil {
		return c.logErrorAndRespond(ctx, http.StatusBadRequest, "Failed to validate request body", err)
	}

	if body.Status == 2 {
		c.logger.Debug(ctx.Request().Context(), "Processing callback with save status", nil)
		if body.Token == "" {
			c.logger.Error(ctx.Request().Context(), "Failed to extract token from request body", nil)
			return ctx.JSON(http.StatusBadRequest, common.ErrorResponse{Error: callbackErrorCodeFailure})
		}

		auth, settings, err := c.fetchAuthenticationAndSettings(tctx, params)
		if err != nil {
			return c.logErrorAndRespond(ctx, http.StatusBadRequest, "Failed to extract authentication and settings", err)
		}

		c.logger.Debug(ctx.Request().Context(), "Successfully fetched authentication and settings", nil)

		secret := c.getSecretFromSettings(settings)
		c.logger.Debug(ctx.Request().Context(), "Validating token", nil)
		if err = c.jwtService.ValidateTarget(body.Token, []byte(secret), &body); err != nil {
			return c.logErrorAndRespond(ctx, http.StatusUnauthorized, "Failed to validate and map token", err)
		}

		c.logger.Debug(ctx.Request().Context(), "Token validated successfully", nil)

		c.logger.Info(ctx.Request().Context(), "Uploading file to Miro", service.Fields{
			"board_id": params.BID,
			"file_id":  params.FID,
			"file_url": body.Url,
			"filename": params.Filename,
		})

		if _, err := c.miroClient.UploadFile(tctx, miro.UploadFileRequest{
			BoardID:  params.BID,
			ItemID:   params.FID,
			Filename: params.Filename,
			FileURL:  body.Url,
			Token:    auth.AccessToken,
		}); err != nil {
			c.logger.Error(ctx.Request().Context(), "Failed to upload file",
				service.Fields{
					"error":    err.Error(),
					"board_id": params.BID,
					"file_id":  params.FID,
					"file_url": body.Url,
				},
			)
			return ctx.JSON(http.StatusInternalServerError, common.ErrorResponse{Error: callbackErrorCodeFailure})
		}

		c.logger.Info(ctx.Request().Context(), "File uploaded successfully",
			service.Fields{
				"board_id": params.BID,
				"file_id":  params.FID,
				"filename": params.Filename,
			},
		)
	} else {
		c.logger.Info(ctx.Request().Context(), "Skipping file upload for non-save callback",
			service.Fields{
				"status": body.Status,
				"bid":    params.BID,
				"fid":    params.FID,
			})
	}

	c.logger.Debug(ctx.Request().Context(), "Callback processing complete", nil)
	return ctx.JSON(http.StatusOK, common.ErrorResponse{Error: callbackErrorCodeSuccess})
}
