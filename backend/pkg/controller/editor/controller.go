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
package editor

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/controller/base"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/document"
	oauthService "github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/oauth"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/settings"
	echo "github.com/labstack/echo/v4"
	errgroup "golang.org/x/sync/errgroup"
)

type editorController struct {
	base.BaseController
}

func NewEditorController(
	config *config.Config,
	miroClient miro.Client,
	jwtService crypto.Signer,
	builderService document.BuilderService,
	oauthService oauthService.OAuthService[miro.AuthenticationResponse],
	settingsService settings.SettingsService,
	translationService service.TranslationProvider,
	logger service.Logger,
) common.Handler {
	controller := &editorController{
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
		common.MethodGet: controller.handleGet,
	})
}

func buildCallbackURL(base, fid, uid, tid, bid, filename string) string {
	return fmt.Sprintf("%s?fid=%s&uid=%s&tid=%s&bid=%s&filename=%s", base, fid, uid, tid, bid, url.QueryEscape(filename))
}

func (c *editorController) extractAndValidateParams(ctx echo.Context) (editorRequestParams, error) {
	token, err := c.BaseController.ExtractUserToken(ctx)
	if err != nil {
		return editorRequestParams{}, err
	}

	bid, err := c.BaseController.GetQueryParam(ctx, "bid")
	if err != nil {
		return editorRequestParams{}, err
	}

	fid, err := c.BaseController.GetQueryParam(ctx, "fid")
	if err != nil {
		return editorRequestParams{}, err
	}

	lang, err := c.BaseController.GetQueryParam(ctx, "lang")
	if err != nil {
		lang = "en"
	}

	return editorRequestParams{token.User, token.Team, bid, fid, lang}, nil
}

func (c *editorController) fetchMiroData(
	ctx context.Context,
	params editorRequestParams,
	accessToken string,
) (*miro.UserInfoResponse, *miro.BoardResponse, *miro.FileInfoResponse, error) {
	g, _ := errgroup.WithContext(ctx)
	var userInfo *miro.UserInfoResponse
	var boardInfo *miro.BoardResponse
	var fileInfo *miro.FileInfoResponse

	g.Go(func() error {
		var err error
		userInfo, err = c.BaseController.MiroClient.GetUserInfo(ctx, miro.GetUserInfoRequest{
			Token: accessToken,
		})

		if err != nil {
			return fmt.Errorf("failed to fetch user info: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		boardInfo, err = c.BaseController.MiroClient.GetBoard(ctx, miro.GetBoardRequest{
			BoardID: params.bid,
			Token:   accessToken,
		})

		if err != nil {
			return fmt.Errorf("failed to fetch board info: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		fileInfo, err = c.BaseController.MiroClient.GetFileInfo(ctx, miro.GetFileInfoRequest{
			BoardID: params.bid,
			ItemID:  params.fid,
			Token:   accessToken,
		})

		if err != nil {
			return fmt.Errorf("failed to fetch file info: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, nil, nil, err
	}

	publicFile, err := c.BaseController.MiroClient.GetFilePublicURL(ctx, miro.GetFilePublicURLRequest{
		URL:   fileInfo.Data.DocumentURL,
		Token: accessToken,
	})

	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get public URL: %w", err)
	}

	fileInfo.Data.DocumentURL = publicFile.URL
	return userInfo, boardInfo, fileInfo, nil
}

func (c *editorController) resolveServerSettings(settings *component.Settings) (address, secret string, err error) {
	address = settings.Address
	secret = settings.Secret

	if settings.Demo.Enabled && address == "" && secret == "" {
		if settings.Demo.Started != nil {
			demoExpiry := settings.Demo.Started.Add(time.Duration(c.BaseController.Config.DemoServer.Days) * 24 * time.Hour)
			if demoExpiry.After(time.Now()) {
				address = c.BaseController.Config.DemoServer.Address
				secret = c.BaseController.Config.DemoServer.Secret
			}
		}
	}

	if address == "" || secret == "" {
		return "", "", base.ErrSettingsNotConfigured
	}

	return address, secret, nil
}

func (c *editorController) buildEditorConfig(
	ctx context.Context,
	callbackURL string,
	boardID string,
	user *miro.BoardMemberResponse,
	file *miro.FileInfoResponse,
	secret string,
) (*document.Config, error) {
	config, err := c.BaseController.BuilderService.Build(
		ctx,
		callbackURL,
		builderRequest{Board: boardID, File: *file},
		document.WithKey([]byte(secret)),
		document.WithUserConfigurer(user),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to build configuration: %w", err)
	}

	return config, nil
}

func (c *editorController) handleGet(ctx echo.Context) error {
	return c.BaseController.ExecuteWithTimeout(ctx, 3*time.Second, func(tctx context.Context) error {
		handleRequestError := func(err error, message string) error {
			if err != nil {
				if errors.Is(err, base.ErrSettingsNotConfigured) {
					return ctx.Render(http.StatusOK, "unauthorized", map[string]string{
						"authorizationError": c.BaseController.TranslationService.Translate(tctx, "en", "editor.errors.settings_not_configured"),
					})
				}
				return ctx.Render(http.StatusOK, "unauthorized", map[string]string{
					"authorizationError": message,
				})
			}

			return nil
		}

		params, err := c.extractAndValidateParams(ctx)
		if err := handleRequestError(err, c.BaseController.TranslationService.Translate(tctx, "en", "editor.errors.unauthorized")); err != nil {
			return err
		}

		settings, auth, err := c.BaseController.FetchAuthenticationWithSettings(tctx, params.uid, params.tid, params.bid)
		if err := handleRequestError(err, c.BaseController.TranslationService.Translate(tctx, params.lang, "editor.errors.fetch_required_data")); err != nil {
			return err
		}

		address, secret, err := c.resolveServerSettings(settings)
		if err := handleRequestError(err, c.BaseController.TranslationService.Translate(tctx, params.lang, "editor.errors.invalid_configuration")); err != nil {
			return err
		}

		uinfo, board, file, err := c.fetchMiroData(tctx, params, auth.AccessToken)
		if err := handleRequestError(err, c.BaseController.TranslationService.Translate(tctx, params.lang, "editor.errors.fetch_miro_data")); err != nil {
			return err
		}

		callbackURL := buildCallbackURL(c.BaseController.Config.Server.CallbackURL, params.fid, params.uid, params.tid, params.bid, file.Data.Title)
		config, err := c.buildEditorConfig(tctx, callbackURL, board.ID, &miro.BoardMemberResponse{
			MemberID:   uinfo.User.ID,
			MemberName: uinfo.User.Name,
			Role:       "member",
			Lang:       params.lang,
		}, file, secret)
		if err := handleRequestError(err, c.BaseController.TranslationService.Translate(tctx, params.lang, "editor.errors.build_editor_configuration")); err != nil {
			return err
		}

		configJSON, err := json.Marshal(config)
		if err := handleRequestError(err, c.BaseController.TranslationService.Translate(tctx, params.lang, "editor.errors.encode_configuration")); err != nil {
			return err
		}

		return ctx.Render(http.StatusOK, "editor", map[string]any{
			"apijs":       address + "/web-apps/apps/api/documents/api.js",
			"config":      string(configJSON),
			"favicon":     config.DocumentType,
			"editorError": c.BaseController.TranslationService.Translate(tctx, params.lang, "editor.errors.failed_to_load"),
			"closeWindow": c.BaseController.TranslationService.Translate(tctx, params.lang, "editor.close_window"),
		})
	})
}

func (c *editorController) SupportedMethods() []common.HTTPMethod {
	return []common.HTTPMethod{common.MethodGet}
}
