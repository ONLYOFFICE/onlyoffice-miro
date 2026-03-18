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
	"errors"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/core"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/core/component"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/crypto"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/oauth"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/storage/pg"
)

type OAuthable[T any] interface {
	Convert(T) (component.Authentication, error)
}

type oauthService[T any] struct {
	cipher         crypto.Cipher
	oauthClient    oauth.OAuthClient[T]
	oauthConverter OAuthable[T]
	storageService service.Storage[core.AuthCompositeKey, component.Authentication]
	logger         service.Logger
}

func NewOAuthService[T any](
	cipher crypto.Cipher,
	oauthClient oauth.OAuthClient[T],
	oauthConverter OAuthable[T],
	storageService service.Storage[core.AuthCompositeKey, component.Authentication],
	logger service.Logger,
) OAuthService[T] {
	return &oauthService[T]{
		cipher:         cipher,
		oauthClient:    oauthClient,
		oauthConverter: oauthConverter,
		storageService: storageService,
		logger:         logger,
	}
}

func (s *oauthService[T]) encryptTokens(token component.Authentication) (string, string, error) {
	encAccess, err := s.cipher.Encrypt(token.AccessToken)
	if err != nil {
		return "", "", err
	}

	encRefresh, err := s.cipher.Encrypt(token.RefreshToken)
	if err != nil {
		return "", "", err
	}

	return encAccess, encRefresh, nil
}

func (s *oauthService[T]) decryptTokens(token component.Authentication) (string, string, error) {
	decAccess, err := s.cipher.Decrypt(token.AccessToken)
	if err != nil {
		return "", "", err
	}

	decRefresh, err := s.cipher.Decrypt(token.RefreshToken)
	if err != nil {
		return "", "", err
	}

	return decAccess, decRefresh, nil
}

func (s *oauthService[T]) createEncryptedAuth(token component.Authentication) (component.Authentication, error) {
	encAccess, encRefresh, err := s.encryptTokens(token)
	if err != nil {
		return component.Authentication{}, err
	}

	return component.Authentication{
		TokenType:    token.TokenType,
		AccessToken:  encAccess,
		RefreshToken: encRefresh,
		ExpiresAt:    token.ExpiresAt,
		Scope:        token.Scope,
	}, nil
}

func (s *oauthService[T]) createDecryptedAuth(token component.Authentication) (component.Authentication, error) {
	decAccess, decRefresh, err := s.decryptTokens(token)
	if err != nil {
		return component.Authentication{}, err
	}

	return component.Authentication{
		TokenType:    token.TokenType,
		AccessToken:  decAccess,
		RefreshToken: decRefresh,
		ExpiresAt:    token.ExpiresAt,
		Scope:        token.Scope,
	}, nil
}

func (s *oauthService[T]) Save(ctx context.Context, teamID, userID string, token component.Authentication) error {
	s.logger.Info(ctx, "Saving OAuth token", service.Fields{
		"teamID": teamID,
		"userID": userID,
	})

	auth, err := s.createEncryptedAuth(token)
	if err != nil {
		s.logger.Error(ctx, "Failed to encrypt OAuth token", service.Fields{
			"error":  err.Error(),
			"teamID": teamID,
			"userID": userID,
		})
		return err
	}

	_, err = s.storageService.Insert(ctx, core.AuthCompositeKey{
		TeamID: teamID,
		UserID: userID,
	}, auth)

	if err != nil {
		s.logger.Error(ctx, "Failed to save OAuth token", service.Fields{
			"error":  err.Error(),
			"teamID": teamID,
			"userID": userID,
		})
	} else {
		s.logger.Info(ctx, "Successfully saved OAuth token", service.Fields{
			"teamID": teamID,
			"userID": userID,
		})
	}

	return err
}

func (s *oauthService[T]) Find(ctx context.Context, teamID, userID string) (component.Authentication, error) {
	s.logger.Info(ctx, "Finding OAuth token", service.Fields{
		"teamID": teamID,
		"userID": userID,
	})

	key := core.AuthCompositeKey{
		TeamID: teamID,
		UserID: userID,
	}

	storedAuth, err := s.storageService.Find(ctx, key)
	if err != nil {
		if errors.Is(err, pg.ErrNoRowsAffected) {
			s.logger.Warn(ctx, "OAuth token not found", service.Fields{
				"teamID": teamID,
				"userID": userID,
			})
			return storedAuth, ErrTokenMissing
		}

		s.logger.Error(ctx, "Error finding OAuth token", service.Fields{
			"error":  err.Error(),
			"teamID": teamID,
			"userID": userID,
		})
		return storedAuth, err
	}

	if storedAuth.AccessToken == "" {
		s.logger.Warn(ctx, "OAuth token is empty", service.Fields{
			"teamID": teamID,
			"userID": userID,
		})
		return storedAuth, ErrTokenMissing
	}

	if time.Now().Unix() <= int64(storedAuth.ExpiresAt) {
		s.logger.Info(ctx, "Using existing OAuth token", service.Fields{
			"teamID":    teamID,
			"userID":    userID,
			"expiresAt": storedAuth.ExpiresAt,
		})
		return s.createDecryptedAuth(storedAuth)
	}

	s.logger.Info(ctx, "OAuth token expired, refreshing", service.Fields{
		"teamID": teamID,
		"userID": userID,
	})

	refreshToken, err := s.cipher.Decrypt(storedAuth.RefreshToken)
	if err != nil {
		s.logger.Error(ctx, "Failed to decrypt refresh token", service.Fields{
			"error":  err.Error(),
			"teamID": teamID,
			"userID": userID,
		})
		return component.Authentication{}, ErrTokenMissing
	}

	token, err := s.oauthClient.Refresh(ctx, refreshToken)
	if err != nil {
		s.logger.Error(ctx, "Failed to refresh OAuth token", service.Fields{
			"error":  err.Error(),
			"teamID": teamID,
			"userID": userID,
		})
		return component.Authentication{}, ErrTokenMissing
	}

	refreshedToken, err := s.oauthConverter.Convert(token)
	if err != nil {
		s.logger.Error(ctx, "Failed to convert refreshed OAuth token", service.Fields{
			"error":  err.Error(),
			"teamID": teamID,
			"userID": userID,
		})
		return component.Authentication{}, ErrTokenMissing
	}

	updatedAuth, err := s.createEncryptedAuth(refreshedToken)
	if err != nil {
		s.logger.Error(ctx, "Failed to encrypt refreshed OAuth token", service.Fields{
			"error":  err.Error(),
			"teamID": teamID,
			"userID": userID,
		})
		return component.Authentication{}, ErrTokenMissing
	}

	if _, err = s.storageService.Update(ctx, key, updatedAuth); err != nil {
		s.logger.Error(ctx, "Failed to update OAuth token in storage", service.Fields{
			"error":  err.Error(),
			"teamID": teamID,
			"userID": userID,
		})
		return component.Authentication{}, ErrTokenMissing
	}

	s.logger.Info(ctx, "Successfully refreshed and updated OAuth token", service.Fields{
		"teamID": teamID,
		"userID": userID,
	})
	return refreshedToken, nil
}
