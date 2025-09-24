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

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/miro"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/oauth"
	echo "github.com/labstack/echo/v4"
)

func CompositeTokenExtractor(headerName string) TokenExtractor {
	return func(c echo.Context) (string, error) {
		token := c.Request().Header.Get(headerName)
		if token == "" {
			token = c.QueryParam("token")
			if token == "" {
				return "", echo.NewHTTPError(http.StatusUnauthorized, "Missing authentication token")
			}
		}

		return token, nil
	}
}

func MiroSignatureExtractor() TokenExtractor {
	return CompositeTokenExtractor(miroSignature)
}

func NoOpRefresher() TokenRefresher {
	return func(c echo.Context, token *TokenClaims) error {
		return nil
	}
}

func MiroOAuthTokenRefresher(middleware *AuthMiddleware, oauthService oauth.OAuthService[miro.AuthenticationResponse]) TokenRefresher {
	return func(c echo.Context, token *TokenClaims) error {
		if token.RegisteredClaims.ExpiresAt.Before(time.Now()) || time.Until(token.RegisteredClaims.ExpiresAt.Time) < time.Hour {
			_, err := oauthService.Find(c.Request().Context(), token.Team, token.User)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Failed to refresh token")
			}
		}

		return nil
	}
}
