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
package file

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/config"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/crypto"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/miro"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/common"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/controller/base"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/document"
	oauthService "github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/oauth"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/settings"
	echo "github.com/labstack/echo/v4"
)

type fileManagementController struct {
	base.BaseController
}

func NewFileManagementController(
	config *config.Config,
	miroClient miro.Client,
	jwtService crypto.Signer,
	builderService document.BuilderService,
	oauthService oauthService.OAuthService[miro.AuthenticationResponse],
	settingsService settings.SettingsService,
	translationService service.TranslationProvider,
	logger service.Logger,
) common.Handler {
	controller := &fileManagementController{
		BaseController: *base.NewBaseController(
			config,
			miroClient,
			jwtService,
			builderService,
			oauthService,
			settingsService,
			translationService,
			logger,
		),
	}

	return common.NewHandler(map[common.HTTPMethod]echo.HandlerFunc{
		common.MethodGet:  controller.handleGet,
		common.MethodPost: controller.handlePost,
	})
}

func (c *fileManagementController) handleGet(ctx echo.Context) error {
	return c.BaseController.ExecuteWithTimeout(ctx, 4*time.Second, func(tctx context.Context) error {
		boardAuth, err := PrepareRequest(ctx, tctx, &c.BaseController)
		if err != nil {
			return err
		}

		if fid, ferr := c.BaseController.GetQueryParam(ctx, "fid"); ferr == nil {
			file, err := GetFileInfo(ctx, tctx, &c.BaseController, boardAuth.BoardID, fid, boardAuth.Authentication.AccessToken)
			if err != nil {
				return err
			}
			return ctx.JSON(200, file)
		}

		var cursor string
		if c := ctx.QueryParam("cursor"); c != "" {
			cursor = c
		}

		files, err := GetFilesInfo(ctx, tctx, &c.BaseController, boardAuth.BoardID, cursor, boardAuth.Authentication.AccessToken)
		if err != nil {
			return err
		}

		return ctx.JSON(200, files)
	})
}

func (c *fileManagementController) handlePost(ctx echo.Context) error {
	return c.BaseController.ExecuteWithTimeout(ctx, 15*time.Second, func(tctx context.Context) error {
		var body createBody
		if err := json.NewDecoder(ctx.Request().Body).Decode(&body); err != nil {
			return c.BaseController.HandleError(ctx, err, http.StatusBadRequest, "failed to decode request body")
		}

		token, err := c.BaseController.ExtractUserToken(ctx)
		if err != nil {
			return c.BaseController.HandleError(ctx, err, http.StatusBadRequest, "failed to extract authentication parameters")
		}

		_, auth, err := c.BaseController.FetchAuthenticationWithSettings(tctx, token.User, token.Team, body.BoardId)
		if err != nil {
			return c.BaseController.HandleError(ctx, err, http.StatusBadRequest, "failed to fetch required data")
		}

		req := miro.CreateFileRequest{
			BoardID:  body.BoardId,
			Name:     body.FileName,
			Type:     common.ToDocumentType(body.FileType),
			Language: common.ToTemplateLanguage(body.FileLang),
			Token:    auth.AccessToken,
		}

		response, err := CreateFile(ctx, tctx, &c.BaseController, req)
		if err != nil {
			return err
		}

		return c.BaseController.SendJSON(ctx, response)
	})
}
