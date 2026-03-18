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
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/config"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/crypto"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/miro"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/oauth"
	echo "github.com/labstack/echo/v4"
)

func NewHeaderAuthMiddleware(
	config *config.Config,
	jwtService crypto.Signer,
	translator service.TranslationProvider,
	logger service.Logger,
	headerName string,
) *AuthMiddleware {
	return NewAuthMiddleware(
		config,
		CompositeTokenExtractor(headerName),
		NoOpRefresher(),
		jwtService,
		translator,
		logger,
	)
}

func NewMiroAuthMiddleware(
	config *config.Config,
	jwtService crypto.Signer,
	translator service.TranslationProvider,
	logger service.Logger,
) *AuthMiddleware {
	middleware := NewAuthMiddleware(
		config,
		MiroSignatureExtractor(),
		NoOpRefresher(),
		jwtService,
		translator,
		logger,
	)

	extractor := func(c echo.Context) (string, error) {
		tokenString, err := MiroSignatureExtractor()(c)
		if err != nil {
			return "", err
		}

		_, err = middleware.ValidateToken(tokenString)
		if err != nil {
			return "", err
		}

		return tokenString, nil
	}

	middleware.extractor = extractor
	return middleware
}

func NewCookieOAuthMiddleware(
	config *config.Config,
	oauthService oauth.OAuthService[miro.AuthenticationResponse],
	jwtService crypto.Signer,
	translator service.TranslationProvider,
	logger service.Logger,
) *AuthMiddleware {
	middleware := NewAuthMiddleware(
		config,
		MiroSignatureExtractor(),
		nil,
		jwtService,
		translator,
		logger,
	)

	middleware.refresher = MiroOAuthTokenRefresher(middleware, oauthService)
	return middleware
}

func NewTokenAuthMiddleware(
	config *config.Config,
	jwtService crypto.Signer,
	translator service.TranslationProvider,
	logger service.Logger,
) *AuthMiddleware {
	return NewHeaderAuthMiddleware(config, jwtService, translator, logger, miroSignature)
}

func NewCookieAuthMiddleware(
	config *config.Config,
	oauthService oauth.OAuthService[miro.AuthenticationResponse],
	jwtService crypto.Signer,
	translator service.TranslationProvider,
	logger service.Logger,
) *AuthMiddleware {
	return NewCookieOAuthMiddleware(config, oauthService, jwtService, translator, logger)
}

func NewEditorAuthMiddleware(
	config *config.Config,
	oauthService oauth.OAuthService[miro.AuthenticationResponse],
	jwtService crypto.Signer,
	translator service.TranslationProvider,
	logger service.Logger,
) *AuthMiddleware {
	return NewCookieOAuthMiddleware(config, oauthService, jwtService, translator, logger)
}
