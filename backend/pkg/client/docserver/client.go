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
package docserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/common"
)

type client struct {
	httpClient *http.Client
	logger     service.Logger
}

func NewClient(logger service.Logger) Client {
	return &client{
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
			Transport: common.NewRetryableTransport(&http.Transport{
				MaxIdleConnsPerHost:    100,
				IdleConnTimeout:        90 * time.Second,
				MaxResponseHeaderBytes: 1 << 20,
				DisableCompression:     false,
				ForceAttemptHTTP2:      true,
			}),
		},
		logger: logger,
	}
}

func (c *client) createRequest(ctx context.Context, method, baseURL, path string, body any, options *ClientOptions) (*http.Request, error) {
	c.logger.Debug(ctx, "Creating DocServer request", service.Fields{
		"method":  method,
		"baseURL": baseURL,
	})

	address := strings.TrimRight(baseURL, "/")
	if _, err := url.Parse(address); err != nil {
		c.logger.Error(ctx, "Invalid DocServer address", service.Fields{
			"address": address,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("malformed address: %w", err)
	}

	if options.Token == "" {
		c.logger.Error(ctx, "DocServer token required but not provided", nil)
		return nil, ErrTokenRequired
	}

	url := common.Concat(address, path, "?shardKey=", common.GenerateRandomString(8))

	var reader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			c.logger.Error(ctx, "Failed to marshal DocServer request body", service.Fields{
				"error": err.Error(),
			})
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reader = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		c.logger.Error(ctx, "Failed to create DocServer request", service.Fields{
			"url":   url,
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if options.Header != "" {
		req.Header.Set(options.Header, options.Token)
	}

	c.logger.Debug(ctx, "DocServer request created successfully", service.Fields{
		"url": url,
	})
	return req, nil
}

func (c *client) sendRequest(req *http.Request, target any) error {
	ctx := req.Context()
	c.logger.Debug(ctx, "Sending DocServer request", service.Fields{
		"method": req.Method,
		"url":    req.URL.String(),
	})

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error(ctx, "Failed to send DocServer request", service.Fields{
			"method": req.Method,
			"url":    req.URL.String(),
			"error":  err.Error(),
		})
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		c.logger.Error(ctx, "Received non-OK status code from DocServer", service.Fields{
			"method":     req.Method,
			"url":        req.URL.String(),
			"statusCode": resp.StatusCode,
		})
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		c.logger.Error(ctx, "Failed to decode DocServer response", service.Fields{
			"method": req.Method,
			"url":    req.URL.String(),
			"error":  err.Error(),
		})
		return fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Debug(ctx, "DocServer request completed successfully", service.Fields{
		"method": req.Method,
		"url":    req.URL.String(),
	})
	return nil
}

func (c *client) GetServerVersion(ctx context.Context, base string, opts ...Option) (*ServerVersionResponse, error) {
	c.logger.Info(ctx, "Getting DocServer version", service.Fields{
		"baseURL": base,
	})

	options := DefaultClientOptions()
	ApplyOptions(options, opts...)

	body := GetServerVersionRequest{C: "version", Token: options.Token}

	req, err := c.createRequest(ctx, http.MethodPost, base, "/command", body, options)
	if err != nil {
		c.logger.Error(ctx, "Failed to create request for GetServerVersion", service.Fields{
			"baseURL": base,
			"error":   err.Error(),
		})
		return nil, err
	}

	var response ServerVersionResponse
	if err := c.sendRequest(req, &response); err != nil {
		c.logger.Error(ctx, "Failed to get server version", service.Fields{
			"baseURL": base,
			"error":   err.Error(),
		})
		return nil, err
	}

	c.logger.Info(ctx, "DocServer version retrieved successfully", service.Fields{
		"baseURL": base,
		"version": response.Version,
	})
	return &response, nil
}

func (c *client) ConvertFile(ctx context.Context, base, token string, opts ...Option) (*FileConversionResponse, error) {
	c.logger.Info(ctx, "Getting DocServer convert URL", service.Fields{
		"baseURL": base,
	})

	options := DefaultClientOptions()
	ApplyOptions(options, opts...)

	body := ConvertFileRequest{Token: token}

	req, err := c.createRequest(ctx, http.MethodPost, base, "/converter", body, options)
	if err != nil {
		c.logger.Error(ctx, "Failed to create request for ConvertFile", service.Fields{
			"baseURL": base,
			"error":   err.Error(),
		})
		return nil, err
	}

	var response FileConversionResponse
	if err := c.sendRequest(req, &response); err != nil {
		c.logger.Error(ctx, "Failed to convert file", service.Fields{
			"baseURL": base,
			"error":   err.Error(),
		})
		return nil, err
	}

	c.logger.Info(ctx, "File converted successfully", service.Fields{
		"baseURL": base,
		"url":     response.FileURL,
	})

	if response.Error != 0 {
		c.logger.Error(ctx, "Failed to convert file. Received non-zero error code", service.Fields{
			"baseURL": base,
			"error":   response.Error,
		})
		return nil, fmt.Errorf("failed to convert file: %d", response.Error)
	}

	return &response, nil
}
