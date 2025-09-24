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
package authentication

import (
	"net/http"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/config"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/crypto"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/common"
	jwt "github.com/golang-jwt/jwt/v5"
	echo "github.com/labstack/echo/v4"
)

type TokenClaims struct {
	User string `json:"user"`
	Team string `json:"team"`
	jwt.RegisteredClaims
}

type TokenExtractor func(c echo.Context) (string, error)

type TokenRefresher func(c echo.Context, token *TokenClaims) error

type AuthMiddleware struct {
	config      *config.Config
	extractor   TokenExtractor
	refresher   TokenRefresher
	jwtService  crypto.Signer
	translator  service.TranslationProvider
	logger      service.Logger
	defaultLang string
}

func NewAuthMiddleware(
	config *config.Config,
	extractor TokenExtractor,
	refresher TokenRefresher,
	jwtService crypto.Signer,
	translator service.TranslationProvider,
	logger service.Logger,
) *AuthMiddleware {
	return &AuthMiddleware{
		config:      config,
		extractor:   extractor,
		refresher:   refresher,
		jwtService:  jwtService,
		translator:  translator,
		logger:      logger,
		defaultLang: "en",
	}
}

func (m *AuthMiddleware) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString, err := m.extractor(c)
		lang := c.QueryParam("lang")
		if lang == "" {
			lang = m.defaultLang
		}

		if err != nil {
			return c.Render(http.StatusOK, "unauthorized", map[string]string{
				"language":           lang,
				"authorizationError": m.translator.Translate(c.Request().Context(), lang, "errors.authentication.missing_authentication"),
			})
		}

		m.logger.Info(c.Request().Context(), "authenticating request",
			service.Fields{
				"token": tokenString,
			})

		token, err := m.ValidateToken(tokenString)
		if err != nil {
			return c.Render(http.StatusOK, "unauthorized", map[string]string{
				"language":           lang,
				"authorizationError": m.translator.Translate(c.Request().Context(), lang, "errors.authentication.missing_authentication"),
			})
		}

		if m.refresher != nil {
			if err := m.refresher(c, token); err != nil {
				return c.Render(http.StatusOK, "unauthorized", map[string]string{
					"language":           lang,
					"authorizationError": m.translator.Translate(c.Request().Context(), lang, "errors.authentication.missing_authentication"),
				})
			}
		}

		m.logger.Info(c.Request().Context(), "authenticated request",
			service.Fields{
				"token": tokenString,
			})

		c.Set(common.ContextKeyUser, token)
		return next(c)
	}
}

func (m *AuthMiddleware) ValidateToken(tokenString string) (*TokenClaims, error) {
	var token TokenClaims
	if err := m.jwtService.ValidateTarget(tokenString, []byte(m.config.OAuth.ClientSecret), &token); err != nil {
		return nil, err
	}

	return &token, nil
}

func (m *AuthMiddleware) CreateAuthToken(uid, tid string, expiresAt int) (string, error) {
	claims := &TokenClaims{
		User: uid,
		Team: tid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(int64(expiresAt), 0)),
		},
	}

	return m.jwtService.Create(claims, []byte(m.config.OAuth.ClientSecret))
}

func (m *AuthMiddleware) GetTokenAuthorization(c echo.Context) error {
	tokenString, err := m.extractor(c)
	lang := c.QueryParam("lang")
	if lang == "" {
		lang = m.defaultLang
	}

	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": m.translator.Translate(c.Request().Context(), lang, "errors.authentication.missing_authentication"),
		})
	}

	token, err := m.ValidateToken(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": m.translator.Translate(c.Request().Context(), lang, "errors.authentication.invalid_token"),
		})
	}

	expiresAt := time.Now().Add(5 * time.Minute).Unix()
	jwtToken, err := m.CreateAuthToken(token.User, token.Team, int(expiresAt))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate authorization token",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"token":      jwtToken,
		"expires_at": expiresAt,
	})
}
