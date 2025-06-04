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
	"errors"
	"fmt"
	"net/http"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/miro"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/controller/base"
	echo "github.com/labstack/echo/v4"
)

func PrepareRequest(
	ctx echo.Context,
	tctx context.Context,
	c *base.BaseController,
) (*boardAuthenticationResponse, error) {
	token, err := c.ExtractUserToken(ctx)
	if err != nil {
		return nil, c.HandleError(ctx, err, http.StatusForbidden, "failed to extract authentication parameters")
	}

	bid, err := c.GetQueryParam(ctx, "bid")
	if err != nil {
		return nil, c.HandleError(ctx, err, http.StatusBadRequest, "board id parameter is missing")
	}

	_, auth, err := c.FetchAuthenticationWithSettings(tctx, token.User, token.Team, bid)
	if err != nil {
		if errors.Is(err, base.ErrMissingAuthentication) {
			return nil, c.HandleWarning(ctx, err, http.StatusUnauthorized, "could not retrieve authentication")
		}

		if errors.Is(err, base.ErrSettingsNotConfigured) {
			return nil, c.HandleWarning(ctx, err, http.StatusConflict, "could not retrieve document editor settigns")
		}

		return nil, c.HandleError(ctx, err, http.StatusBadRequest, "could not retrieve required data")
	}

	return &boardAuthenticationResponse{
		BoardID:        bid,
		Authentication: auth,
	}, nil
}

func CreateFile(
	ctx echo.Context,
	tctx context.Context,
	c *base.BaseController,
	req miro.CreateFileRequest,
) (*miro.FileCreatedResponse, error) {
	response, err := c.MiroClient.CreateFile(tctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create a file: %w", err)
	}

	return response, nil
}

func GetFileInfo(
	ctx echo.Context,
	tctx context.Context,
	c *base.BaseController,
	boardID string,
	fileID string,
	accessToken string,
) (*miro.FileInfoResponse, error) {
	file, err := c.MiroClient.GetFileInfo(tctx, miro.GetFileInfoRequest{
		BoardID: boardID,
		ItemID:  fileID,
		Token:   accessToken,
	})

	if err != nil {
		return nil, c.HandleError(ctx, err, http.StatusBadRequest, "failed to fetch miro file")
	}

	return file, nil
}

func GetFilesInfo(
	ctx echo.Context,
	tctx context.Context,
	c *base.BaseController,
	boardID string,
	cursor string,
	accessToken string,
) (*miro.FilesInfoResponse, error) {
	files, err := c.MiroClient.GetFilesInfo(tctx, miro.GetFilesInfoRequest{
		Cursor:  cursor,
		BoardID: boardID,
		Token:   accessToken,
	})

	if err != nil {
		return nil, c.HandleError(ctx, err, http.StatusBadRequest, "failed to fetch miro files")
	}

	return files, nil
}
