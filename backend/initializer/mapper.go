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
package initializer

import (
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/core/component"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/miro"
)

type authenticationMapper struct{}

func NewAuthenticationMapper() *authenticationMapper {
	return &authenticationMapper{}
}

func (m *authenticationMapper) Convert(token miro.AuthenticationResponse) (component.Authentication, error) {
	expiresAt := time.Now().Add(time.Second * time.Duration(token.ExpiresIn)).Unix()
	return component.Authentication{
		TokenType:    token.TokenType,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    int(expiresAt),
		Scope:        token.Scope,
	}, nil
}
