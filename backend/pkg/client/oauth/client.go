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
package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/config"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
)

type client[T any] struct {
	config     *config.OAuthConfig
	httpClient *http.Client
	errors     *Errors
	logger     service.Logger
}

func NewOAuthClient[T any](config *config.OAuthConfig, logger service.Logger) (OAuthClient[T], error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &client[T]{
		config: config,
		errors: NewErrors(),
		httpClient: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				MaxIdleConnsPerHost:    50,
				IdleConnTimeout:        90 * time.Second,
				ResponseHeaderTimeout:  1500 * time.Millisecond,
				MaxResponseHeaderBytes: 1 << 20,
				DisableCompression:     false,
				ForceAttemptHTTP2:      true,
			},
		},
		logger: logger,
	}, nil
}

func (c *client[T]) buildFormData(params map[string]string) url.Values {
	data := url.Values{}
	for key, value := range params {
		data.Set(key, value)
	}

	return data
}

func (c *client[T]) doRequest(ctx context.Context, url string, data url.Values) (T, error) {
	var response T
	c.logger.Debug(ctx, "Making OAuth request", service.Fields{
		"url": url,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(data.Encode()))
	if err != nil {
		c.logger.Error(ctx, "Failed to create request", service.Fields{
			"error": err.Error(),
		})
		return response, c.errors.FailedToCreateRequest(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error(ctx, "Failed to send request", service.Fields{
			"error": err.Error(),
		})
		return response, c.errors.FailedToSendRequest(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		c.logger.Error(ctx, "Request failed", service.Fields{
			"status_code": resp.StatusCode,
		})
		return response, c.errors.RequestFailed(resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.logger.Error(ctx, "Failed to decode response", service.Fields{
			"error": err.Error(),
		})
		return response, c.errors.FailedToDecodeResponse(err)
	}

	c.logger.Debug(ctx, "OAuth request successful", nil)
	return response, nil
}

func (c *client[T]) Exchange(ctx context.Context, code string) (T, error) {
	var zero T
	c.logger.Debug(ctx, "Exchanging authorization code for token", service.Fields{
		"code":         code,
		"client_id":    c.config.ClientID,
		"redirect_uri": c.config.RedirectURI,
		"grant_type":   "authorization_code",
	})

	data := c.buildFormData(map[string]string{
		"grant_type":    "authorization_code",
		"code":          code,
		"client_id":     c.config.ClientID,
		"client_secret": c.config.ClientSecret,
		"redirect_uri":  c.config.RedirectURI,
	})

	address, err := url.Parse(c.config.TokenURI)
	if err != nil {
		c.logger.Error(ctx, "Failed to parse token URI", service.Fields{
			"error": err.Error(),
		})
		return zero, c.errors.FailedToExchangeToken(err)
	}

	query := address.Query()
	query.Set("client_id", c.config.ClientID)
	query.Set("code", code)
	query.Set("redirect_uri", c.config.RedirectURI)
	address.RawQuery = query.Encode()

	response, err := c.doRequest(ctx, address.String(), data)
	if err != nil {
		c.logger.Error(ctx, "Token exchange failed", service.Fields{
			"error": err.Error(),
		})
		return zero, c.errors.FailedToExchangeToken(err)
	}

	c.logger.Debug(ctx, "Token exchange successful", nil)
	return response, nil
}

func (c *client[T]) Refresh(ctx context.Context, refreshToken string) (T, error) {
	var zero T
	c.logger.Info(ctx, "Refreshing token", nil)

	data := c.buildFormData(map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     c.config.ClientID,
		"client_secret": c.config.ClientSecret,
	})

	address, err := url.Parse(c.config.TokenURI)
	if err != nil {
		c.logger.Error(ctx, "Failed to parse token URI", service.Fields{
			"error": err.Error(),
		})
		return zero, c.errors.FailedToRefreshToken(err)
	}

	query := address.Query()
	query.Set("grant_type", "refresh_token")
	query.Set("refresh_token", refreshToken)
	address.RawQuery = query.Encode()

	response, err := c.doRequest(ctx, address.String(), data)
	if err != nil {
		c.logger.Error(ctx, "Token refresh failed", service.Fields{
			"error": err.Error(),
		})
		return zero, c.errors.FailedToRefreshToken(err)
	}

	c.logger.Debug(ctx, "Token refresh successful", nil)
	return response, nil
}
