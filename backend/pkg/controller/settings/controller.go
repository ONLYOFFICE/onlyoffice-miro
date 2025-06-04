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
package settings

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/miro"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/common"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/middleware/authentication"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/oauth"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/settings"
	echo "github.com/labstack/echo/v4"
)

type settingsController struct {
	miroClient      miro.Client
	settingsService settings.SettingsService
	oauthService    oauth.OAuthService[miro.AuthenticationResponse]
	timeout         time.Duration
	logger          service.Logger
}

func NewSettingsController(
	miroClient miro.Client,
	settingsService settings.SettingsService,
	oauthService oauth.OAuthService[miro.AuthenticationResponse],
	timeout time.Duration,
	logger service.Logger,
) common.Handler {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	controller := &settingsController{
		miroClient:      miroClient,
		settingsService: settingsService,
		oauthService:    oauthService,
		timeout:         timeout,
		logger:          logger,
	}

	return common.NewHandler(map[common.HTTPMethod]echo.HandlerFunc{
		common.MethodGet:  controller.handleGet,
		common.MethodPost: controller.handlePost,
	})
}

func validateRequest(ctx echo.Context) (string, *authentication.TokenClaims, error) {
	bid := ctx.QueryParam("bid")
	if bid == "" {
		return "", nil, ErrMissingBoardParameter
	}

	token, ok := ctx.Get("user").(*authentication.TokenClaims)
	if !ok {
		return "", nil, ErrMissingOpenIdToken
	}

	return bid, token, nil
}

func validatePersistRequest(ctx echo.Context) (*persistSettingsRequest, *authentication.TokenClaims, error) {
	var body persistSettingsRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&body); err != nil {
		return nil, nil, ErrInvalidRequestBody
	}

	if body.BoardID == "" {
		return nil, nil, ErrMissingBoardParameter
	}

	token, ok := ctx.Get("user").(*authentication.TokenClaims)
	if !ok {
		return nil, nil, ErrMissingOpenIdToken
	}

	return &body, token, nil
}

func (c *settingsController) handleGet(ctx echo.Context) error {
	tctx, cancel := context.WithTimeout(ctx.Request().Context(), c.timeout)
	defer cancel()

	bid, token, err := validateRequest(ctx)
	if err != nil {
		c.logger.Error(ctx.Request().Context(), "Invalid request", service.Fields{"error": err})
		return ctx.JSON(http.StatusBadRequest, common.ErrorResponse{Error: err.Error()})
	}

	user, err := c.oauthService.Find(tctx, token.Team, token.User)
	if err != nil {
		if errors.Is(err, oauth.ErrTokenMissing) {
			c.logger.Error(ctx.Request().Context(), "Authentication error", service.Fields{"error": err, "user_id": token.User, "team_id": token.Team})
			return ctx.JSON(http.StatusUnauthorized, common.ErrorResponse{Error: err.Error()})
		}

		c.logger.Error(ctx.Request().Context(), "Failed to fetch user authentication", service.Fields{"error": err, "user_id": token.User, "team_id": token.Team})
		return ctx.JSON(http.StatusInternalServerError, common.ErrorResponse{Error: err.Error()})
	}

	member, err := c.miroClient.GetBoardMember(tctx, miro.GetBoardMemberRequest{
		BoardID:  bid,
		MemberID: token.User,
		Token:    user.AccessToken,
	})

	if err != nil {
		c.logger.Error(ctx.Request().Context(), "Failed to get board member", service.Fields{"error": err, "board_id": bid, "user_id": token.User})
		return ctx.JSON(http.StatusInternalServerError, common.ErrorResponse{Error: err.Error()})
	}

	if strings.ToLower(member.Role) != "owner" {
		c.logger.Warn(ctx.Request().Context(), "Access denied: not board owner", service.Fields{"user_id": token.User, "board_id": bid, "role": member.Role})
		return ctx.JSON(http.StatusForbidden, common.ErrorResponse{Error: "Only owners can access this endpoint"})
	}

	settings, err := c.settingsService.Find(tctx, token.Team, bid)
	if err != nil {
		c.logger.Error(ctx.Request().Context(), "Failed to fetch settings", service.Fields{"error": err, "board_id": bid, "team_id": token.Team})
		return ctx.JSON(http.StatusBadRequest, common.ErrorResponse{Error: err.Error()})
	}

	c.logger.Info(ctx.Request().Context(), "Settings retrieved successfully", service.Fields{"board_id": bid, "user_id": token.User, "team_id": token.Team})
	return ctx.JSON(http.StatusOK, settings)
}

func (c *settingsController) handlePost(ctx echo.Context) error {
	tctx, cancel := context.WithTimeout(ctx.Request().Context(), c.timeout)
	defer cancel()

	body, token, err := validatePersistRequest(ctx)
	if err != nil {
		c.logger.Error(ctx.Request().Context(), "Invalid request body", service.Fields{"error": err})
		return ctx.JSON(http.StatusBadRequest, common.ErrorResponse{Error: err.Error()})
	}

	user, err := c.oauthService.Find(tctx, token.Team, token.User)
	if err != nil {
		if errors.Is(err, oauth.ErrTokenMissing) {
			c.logger.Error(ctx.Request().Context(), "Authentication error", service.Fields{"error": err, "user_id": token.User, "team_id": token.Team})
			return ctx.JSON(http.StatusUnauthorized, common.ErrorResponse{Error: err.Error()})
		}

		c.logger.Error(ctx.Request().Context(), "Failed to fetch user authentication", service.Fields{"error": err, "user_id": token.User, "team_id": token.Team})
		return ctx.JSON(http.StatusInternalServerError, common.ErrorResponse{Error: err.Error()})
	}

	member, err := c.miroClient.GetBoardMember(tctx, miro.GetBoardMemberRequest{
		BoardID:  body.BoardID,
		MemberID: token.User,
		Token:    user.AccessToken,
	})

	if err != nil {
		c.logger.Error(ctx.Request().Context(), "Failed to get board member", service.Fields{"error": err, "board_id": body.BoardID, "user_id": token.User})
		return ctx.JSON(http.StatusInternalServerError, common.ErrorResponse{Error: err.Error()})
	}

	if strings.ToLower(member.Role) != "owner" {
		c.logger.Warn(ctx.Request().Context(), "Access denied: not board owner", service.Fields{"user_id": token.User, "board_id": body.BoardID, "role": member.Role})
		return ctx.JSON(http.StatusForbidden, common.ErrorResponse{Error: "Only owners can access this endpoint"})
	}

	if err := c.settingsService.Save(
		tctx,
		token.Team,
		body.BoardID,
		settings.WithAddress(body.Address),
		settings.WithHeader(body.Header),
		settings.WithSecret(body.Secret),
		settings.WithDemo(body.Demo),
	); err != nil {
		c.logger.Error(ctx.Request().Context(), "Failed to save settings", service.Fields{"error": err, "board_id": body.BoardID, "team_id": token.Team})
		return ctx.JSON(http.StatusInternalServerError, common.ErrorResponse{Error: err.Error()})
	}

	c.logger.Info(ctx.Request().Context(), "Settings updated successfully", service.Fields{"board_id": body.BoardID, "user_id": token.User, "team_id": token.Team})
	return ctx.JSON(http.StatusOK, nil)
}
