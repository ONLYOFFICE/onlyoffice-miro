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
package miro

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strings"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/assets"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/config"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/common"
)

type client struct {
	baseUrl    string
	httpClient *http.Client
	errors     *Errors
	logger     service.Logger
}

func NewMiroClient(config *config.MiroConfig, logger service.Logger) Client {
	return &client{
		baseUrl: config.BaseURL,
		errors:  NewErrors(),
		httpClient: &http.Client{
			Timeout: config.Timeout,
			Transport: common.NewRetryableTransport(&http.Transport{
				MaxIdleConnsPerHost:    100,
				IdleConnTimeout:        90 * time.Second,
				MaxResponseHeaderBytes: 1 << 20,
				DisableCompression:     false,
				ForceAttemptHTTP2:      true,
			},
			),
		},
		logger: logger,
	}
}

func (c *client) buildURL(paths ...string) string {
	var sb strings.Builder
	sb.WriteString(c.baseUrl)
	sb.WriteString("/")
	sb.WriteString(strings.Join(paths, "/"))
	return sb.String()
}

func (c *client) sendRequest(
	ctx context.Context,
	method, url, token string,
	body io.Reader,
	headers map[string]string,
	result any,
) error {
	c.logger.Info(ctx, fmt.Sprintf("Sending %s request to %s", method, url))

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("Failed to create request: %v", err))
		return c.errors.FailedToCreateRequest(err)
	}

	req.Header.Set("Authorization", common.Concat("Bearer ", token))
	req.Header.Set("Accept", "application/json")
	if body != nil && headers == nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("Failed to send request: %v", err))
		return c.errors.FailedToSendRequest(err)
	}

	defer res.Body.Close()
	if res.StatusCode >= 300 || res.StatusCode < 200 {
		c.logger.Error(ctx, fmt.Sprintf("Request failed with status code: %d", res.StatusCode))
		return c.errors.RequestFailed(res.StatusCode)
	}

	c.logger.Debug(ctx, fmt.Sprintf("Request successful with status code: %d", res.StatusCode))

	if result != nil {
		if err := json.NewDecoder(res.Body).Decode(result); err != nil {
			c.logger.Error(ctx, fmt.Sprintf("Failed to decode response: %v", err))
			return c.errors.FailedToDecodeResponse(err)
		}
	}

	return nil
}

func (c *client) GetBoard(ctx context.Context, req GetBoardRequest) (*BoardResponse, error) {
	c.logger.Info(ctx, "Getting board info", service.Fields{
		"boardId": req.BoardID,
	})

	if err := req.Validate(); err != nil {
		c.logger.Error(ctx, fmt.Sprintf("Invalid get board request: %v", err))
		return nil, err
	}

	var response BoardResponse
	url := c.buildURL("boards", req.BoardID)
	if err := c.sendRequest(ctx, http.MethodGet, url, req.Token, nil, nil, &response); err != nil {
		return nil, c.errors.FailedToGetBoard(err)
	}

	c.logger.Debug(ctx, "Successfully retrieved board info", service.Fields{
		"boardId": req.BoardID,
	})
	return &response, nil
}

func (c *client) GetBoardMember(ctx context.Context, req GetBoardMemberRequest) (*BoardMemberResponse, error) {
	c.logger.Info(ctx, "Getting board member info", service.Fields{
		"boardId":  req.BoardID,
		"memberId": req.MemberID,
	})

	if err := req.Validate(); err != nil {
		c.logger.Error(ctx, fmt.Sprintf("Invalid get board member request: %v", err))
		return nil, err
	}

	var response BoardMemberResponse
	url := c.buildURL("boards", req.BoardID, "members", req.MemberID)
	if err := c.sendRequest(ctx, http.MethodGet, url, req.Token, nil, nil, &response); err != nil {
		return nil, c.errors.FailedToGetBoardMember(err)
	}

	c.logger.Debug(ctx, "Successfully retrieved board member info", service.Fields{
		"boardId":  req.BoardID,
		"memberId": req.MemberID,
	})
	return &response, nil
}

func (c *client) GetFileInfo(ctx context.Context, req GetFileInfoRequest) (*FileInfoResponse, error) {
	c.logger.Info(ctx, "Getting file info", service.Fields{
		"boardId": req.BoardID,
		"itemId":  req.ItemID,
	})

	if err := req.Validate(); err != nil {
		c.logger.Error(ctx, fmt.Sprintf("Invalid get file info request: %v", err))
		return nil, err
	}

	var response FileInfoResponse
	url := c.buildURL("boards", req.BoardID, "items", req.ItemID)
	if err := c.sendRequest(ctx, http.MethodGet, url, req.Token, nil, nil, &response); err != nil {
		return nil, c.errors.FailedToGetFileInfo(err)
	}

	c.logger.Debug(ctx, "Successfully retrieved file info", service.Fields{
		"boardId": req.BoardID,
		"itemId":  req.ItemID,
	})
	return &response, nil
}

func (c *client) GetFilesInfo(ctx context.Context, req GetFilesInfoRequest) (*FilesInfoResponse, error) {
	c.logger.Info(ctx, "Getting files info", service.Fields{
		"boardId": req.BoardID,
	})

	if err := req.Validate(); err != nil {
		c.logger.Error(ctx, fmt.Sprintf("Invalid get files info request: %v", err))
		return nil, err
	}

	var response FilesInfoResponse
	url := c.buildURL("boards", req.BoardID, "items?type=document&limit=50")
	if req.Cursor != "" {
		var sb strings.Builder
		sb.WriteString(url)
		sb.WriteString("&cursor=")
		sb.WriteString(req.Cursor)
		url = sb.String()
		c.logger.Debug(ctx, "Using cursor for pagination", service.Fields{
			"cursor": req.Cursor,
		})
	}

	if err := c.sendRequest(ctx, http.MethodGet, url, req.Token, nil, nil, &response); err != nil {
		return nil, c.errors.FailedToGetFileInfo(err)
	}

	c.logger.Debug(ctx, "Successfully retrieved files info", service.Fields{
		"boardId":   req.BoardID,
		"itemCount": len(response.Data),
	})
	return &response, nil
}

func (c *client) GetFilePublicURL(ctx context.Context, req GetFilePublicURLRequest) (*FileLocationResponse, error) {
	c.logger.Info(ctx, "Getting file public URL", service.Fields{
		"url": req.URL,
	})

	if err := req.Validate(); err != nil {
		c.logger.Error(ctx, fmt.Sprintf("Invalid get file public URL request: %v", err))
		return nil, err
	}

	var response FileLocationResponse
	if err := c.sendRequest(ctx, http.MethodGet, req.URL, req.Token, nil, nil, &response); err != nil {
		return nil, c.errors.FailedToGetFileURL(err)
	}

	c.logger.Debug(ctx, "Successfully retrieved file public URL")
	return &response, nil
}

func (c *client) GetUserInfo(ctx context.Context, req GetUserInfoRequest) (*UserInfoResponse, error) {
	c.logger.Info(ctx, "Getting user info", service.Fields{
		"token": req.Token,
	})

	if err := req.Validate(); err != nil {
		c.logger.Error(ctx, "Invalid get user info request", service.Fields{"error": err.Error()})
		return nil, err
	}

	var response UserInfoResponse
	url := "https://api.miro.com/v1/oauth-token"
	if err := c.sendRequest(ctx, http.MethodGet, url, req.Token, nil, nil, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *client) createMultipartForm(data map[string]any, filePath string, position map[string]any) (*bytes.Buffer, string, error) {
	// Context is not available in this method, using background context for logging
	ctx := context.Background()
	c.logger.Debug(ctx, "Creating multipart form", service.Fields{
		"filePath": filePath,
	})

	if position != nil {
		data["position"] = position
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	payload, err := json.Marshal(data)
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("Failed to marshal form data: %v", err))
		return nil, "", c.errors.FailedToMarshalRequest(err)
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, "data"))
	h.Set("Content-Type", "application/json")
	part, err := writer.CreatePart(h)
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("Failed to create form part: %v", err))
		return nil, "", c.errors.FailedToCreateFormFile(err)
	}

	if _, err := part.Write(payload); err != nil {
		c.logger.Error(ctx, fmt.Sprintf("Failed to write form data: %v", err))
		return nil, "", c.errors.FailedToWriteFileData(err)
	}

	fileData, err := assets.Templates.ReadFile(filePath)
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("Failed to read template file: %v", err), service.Fields{
			"filePath": filePath,
		})
		return nil, "", c.errors.FailedToReadFile(err)
	}

	fileWriter, err := writer.CreateFormFile("resource", filepath.Base(filePath))
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("Failed to create form file: %v", err))
		return nil, "", c.errors.FailedToCreateFormFile(err)
	}

	if _, err := fileWriter.Write(fileData); err != nil {
		c.logger.Error(ctx, fmt.Sprintf("Failed to write file data: %v", err))
		return nil, "", c.errors.FailedToWriteFileData(err)
	}

	if err := writer.Close(); err != nil {
		c.logger.Error(ctx, fmt.Sprintf("Failed to close form writer: %v", err))
		return nil, "", c.errors.FailedToCloseWriter(err)
	}

	c.logger.Debug(ctx, "Successfully created multipart form")
	return body, writer.FormDataContentType(), nil
}

func (c *client) CreateFile(ctx context.Context, req CreateFileRequest) (*FileCreatedResponse, error) {
	c.logger.Info(ctx, "Creating file", service.Fields{
		"boardId":  req.BoardID,
		"fileName": fmt.Sprintf("%s.%s", req.Name, req.Type),
		"type":     req.Type,
	})

	if err := req.Validate(); err != nil {
		c.logger.Error(ctx, fmt.Sprintf("Invalid create file request: %v", err))
		return nil, err
	}

	assetPath := fmt.Sprintf("new.%s", string(req.Type))
	templatePath := filepath.Join("templates", req.Language, assetPath)
	_, err := assets.Templates.ReadFile(templatePath)
	if err != nil {
		c.logger.Warn(ctx, "Template not found, falling back to default", service.Fields{
			"templatePath": templatePath,
			"language":     req.Language,
			"type":         req.Type,
		})
		templatePath = filepath.Join("templates", "default", assetPath)
	}

	data := map[string]any{
		"title": fmt.Sprintf("%s.%s", req.Name, string(req.Type)),
	}

	x, _ := rand.Int(rand.Reader, big.NewInt(2500))
	y, _ := rand.Int(rand.Reader, big.NewInt(2500))
	xSign, _ := rand.Int(rand.Reader, big.NewInt(2))
	ySign, _ := rand.Int(rand.Reader, big.NewInt(2))

	if xSign.Int64() == 0 {
		x = x.Neg(x)
	}

	if ySign.Int64() == 0 {
		y = y.Neg(y)
	}

	body, contentType, err := c.createMultipartForm(data, templatePath, map[string]any{
		"x": x,
		"y": y,
	})

	if err != nil {
		return nil, err
	}

	url := c.buildURL("boards", req.BoardID, "documents")
	headers := map[string]string{
		"Content-Type": contentType,
	}

	var response FileCreatedResponse
	if err := c.sendRequest(ctx, http.MethodPost, url, req.Token, body, headers, &response); err != nil {
		return nil, err
	}

	c.logger.Debug(ctx, "Successfully created file", service.Fields{
		"boardId":  req.BoardID,
		"itemId":   response.ID,
		"fileName": fmt.Sprintf("%s.%s", req.Name, req.Type),
	})
	return &response, nil
}

func (c *client) UploadFile(ctx context.Context, req UploadFileRequest) (*FileLocationResponse, error) {
	c.logger.Info(ctx, "Uploading file", service.Fields{
		"boardId": req.BoardID,
		"itemId":  req.ItemID,
	})

	if err := req.Validate(); err != nil {
		c.logger.Error(ctx, fmt.Sprintf("Invalid upload file request: %v", err))
		return nil, err
	}

	body := FileUploadRequest{
		Data: FileUploadRequestData{
			URL: req.FileURL,
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("Failed to marshal upload request: %v", err))
		return nil, c.errors.FailedToMarshalRequest(err)
	}

	url := c.buildURL("boards", req.BoardID, "documents", req.ItemID)
	headers := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}

	var response FileLocationResponse
	if err := c.sendRequest(ctx, http.MethodPatch, url, req.Token, bytes.NewBuffer(payload), headers, &response); err != nil {
		return nil, c.errors.FailedToUploadFile(err)
	}

	c.logger.Debug(ctx, "Successfully uploaded file", service.Fields{
		"boardId": req.BoardID,
		"itemId":  req.ItemID,
	})
	return &response, nil
}
