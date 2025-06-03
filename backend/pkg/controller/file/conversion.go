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
	"crypto/md5"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/config"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/crypto"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/docserver"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/miro"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/common"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/controller/base"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/document"
	oauthService "github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/oauth"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/settings"
	jwt "github.com/golang-jwt/jwt/v5"
	echo "github.com/labstack/echo/v4"
)

type fileConversionController struct {
	base.BaseController
	DocserverClient docserver.Client
}

func NewFileConversionController(
	config *config.Config,
	miroClient miro.Client,
	docserverClient docserver.Client,
	jwtService crypto.Signer,
	builderService document.BuilderService,
	oauthService oauthService.OAuthService[miro.AuthenticationResponse],
	settingsService settings.SettingsService,
	translationService service.TranslationProvider,
	logger service.Logger,
) common.Handler {
	controller := &fileConversionController{
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
		DocserverClient: docserverClient,
	}

	return common.NewHandler(map[common.HTTPMethod]echo.HandlerFunc{
		common.MethodGet: controller.handleGet,
	})
}

func (c *fileConversionController) handleGet(ctx echo.Context) error {
	return c.BaseController.ExecuteWithTimeout(ctx, 2*time.Second, func(tctx context.Context) error {
		boardAuth, err := PrepareRequest(ctx, tctx, &c.BaseController)
		if err != nil {
			return err
		}

		if boardAuth == nil {
			return c.BaseController.HandleError(ctx, ErrInvalidBoardAuthentication, http.StatusInternalServerError, ErrInvalidBoardAuthentication.Error())
		}

		if boardAuth.Authentication == nil {
			return c.BaseController.HandleError(ctx, ErrMissingAuthenticationData, http.StatusInternalServerError, ErrMissingAuthenticationData.Error())
		}

		if fid, ferr := c.BaseController.GetQueryParam(ctx, "fid"); ferr == nil {
			file, err := GetFileInfo(ctx, tctx, &c.BaseController, boardAuth.BoardID, fid, boardAuth.Authentication.AccessToken)
			if err != nil {
				return err
			}

			location, err := c.BaseController.MiroClient.GetFilePublicURL(tctx, miro.GetFilePublicURLRequest{
				URL:   file.Data.DocumentURL,
				Token: boardAuth.Authentication.AccessToken,
			})

			if err != nil {
				return c.BaseController.HandleError(ctx, err, http.StatusInternalServerError, ErrFailedToFetchMiroFile.Error())
			}

			token, err := c.BaseController.ExtractUserToken(ctx)
			if err != nil {
				return c.BaseController.HandleError(ctx, err, http.StatusForbidden, ErrFailedToExtractToken.Error())
			}

			settings, err := c.BaseController.SettingsService.Find(tctx, token.Team, boardAuth.BoardID)
			if err != nil {
				return c.BaseController.HandleError(ctx, err, http.StatusBadRequest, ErrFailedToFetchSettings.Error())
			}

			address := settings.Address
			secret := settings.Secret
			if settings.Demo.Enabled && settings.Demo.Started.Add(time.Duration(c.BaseController.Config.DemoServer.Days)*24*time.Hour).After(time.Now()) && (address == "" || secret == "") {
				address = c.BaseController.Config.DemoServer.Address
				secret = c.BaseController.Config.DemoServer.Secret
			}

			fileExt := path.Ext(file.Data.Title)
			convReq := convertClaims{
				Async:      true,
				FileType:   string(common.ToDocumentType(fileExt)),
				Key:        fmt.Sprintf("%x", md5.Sum([]byte(file.Data.DocumentURL))),
				OutputType: "pdf",
				Title:      file.Data.Title,
				URL:        location.URL,
				RegisteredClaims: jwt.RegisteredClaims{
					IssuedAt:  jwt.NewNumericDate(time.Now()),
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
				},
			}

			jwtToken, err := c.BaseController.JwtService.Create(convReq, []byte(secret))
			if err != nil {
				return c.BaseController.HandleError(ctx, err, http.StatusBadRequest, "failed to create token")
			}

			response, err := c.DocserverClient.ConvertFile(
				tctx,
				address,
				jwtToken,
				docserver.WithHeader(settings.Header),
				docserver.WithToken(settings.Secret),
			)

			if err != nil {
				return c.BaseController.HandleError(ctx, err, http.StatusBadRequest, "failed to convert file")
			}

			return ctx.JSON(200, convertResponse{
				URL:     response.FileURL,
				Percent: response.Percent,
			})
		}

		return ctx.JSON(200, nil)
	})
}
