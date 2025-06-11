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
package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/config"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/core"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/core/component"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/crypto"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/docserver"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/storage/pg"
	jwt "github.com/golang-jwt/jwt/v5"
)

const defaultCacheExpiration = 5 * time.Minute

type settingsService struct {
	config          *config.Config
	docServerClient docserver.Client
	cache           service.Cache
	cipher          crypto.Cipher
	jwtService      crypto.Signer
	storageService  service.Storage[core.SettingsCompositeKey, component.Settings]
	logger          service.Logger
}

func NewSettingsService(
	config *config.Config,
	docServerClient docserver.Client,
	cache service.Cache,
	cipher crypto.Cipher,
	jwtService crypto.Signer,
	storageService service.Storage[core.SettingsCompositeKey, component.Settings],
	logger service.Logger,
) SettingsService {
	return &settingsService{
		config:          config,
		docServerClient: docServerClient,
		cache:           cache,
		cipher:          cipher,
		jwtService:      jwtService,
		storageService:  storageService,
		logger:          logger,
	}
}

func (s *settingsService) buildCacheKey(teamID, boardID string) string {
	return fmt.Sprintf("settings:%s:%s", teamID, boardID)
}

func (s *settingsService) createCompositeKey(teamID, boardID string) core.SettingsCompositeKey {
	return core.SettingsCompositeKey{
		TeamID:  teamID,
		BoardID: boardID,
	}
}

func (s *settingsService) logEvent(ctx context.Context, level config.Level, message string, teamID, boardID string, err error) {
	fields := service.Fields{
		"team_id":  teamID,
		"board_id": boardID,
	}

	if err != nil {
		fields["error"] = err.Error()
	}

	switch level {
	case config.Debug:
		s.logger.Debug(ctx, message, fields)
	case config.Warn:
		s.logger.Warn(ctx, message, fields)
	case config.Error:
		s.logger.Error(ctx, message, fields)
	}
}

func (s *settingsService) encryptSecret(secret string) (string, error) {
	if secret == "" {
		return "", nil
	}

	encrypted, err := s.cipher.Encrypt(secret)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt secret: %w", err)
	}

	return encrypted, nil
}

func (s *settingsService) decryptSecret(encryptedSecret string) (string, error) {
	if encryptedSecret == "" {
		return "", nil
	}

	decrypted, err := s.cipher.Decrypt(encryptedSecret)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secret: %w", err)
	}

	return decrypted, nil
}

func (s *settingsService) createDemoSettings(teamID string, existingStarted *time.Time) component.Demo {
	demoSettings := component.Demo{
		TeamID:  teamID,
		Enabled: true,
	}

	if existingStarted != nil {
		demoSettings.Started = existingStarted
	} else {
		now := time.Now()
		demoSettings.Started = &now
	}

	return demoSettings
}

func (s *settingsService) invalidateCache(ctx context.Context, teamID, boardID string) {
	cacheKey := s.buildCacheKey(teamID, boardID)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logEvent(ctx, config.Warn, "Failed to invalidate settings cache", teamID, boardID, err)
	}
}

func (s *settingsService) cacheSettings(ctx context.Context, teamID, boardID string, settings component.Settings) {
	cacheKey := s.buildCacheKey(teamID, boardID)
	settingsJson, err := json.Marshal(settings)
	if err != nil {
		s.logEvent(ctx, config.Warn, "Failed to marshal settings for caching", teamID, boardID, err)
		return
	}

	if err := s.cache.Set(ctx, cacheKey, settingsJson, defaultCacheExpiration); err != nil {
		s.logEvent(ctx, config.Warn, "Failed to cache settings", teamID, boardID, err)
	}
}

func validateDocServerVersion(version string) error {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid docserver version format: %s", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid docserver major version: %s", version)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid docserver minor version: %s", version)
	}

	if major < 8 || (major == 8 && minor < 2) {
		return fmt.Errorf("docserver version is not supported: %s. Required version 8.2 or newer", version)
	}

	return nil
}

func (s *settingsService) Save(ctx context.Context, teamID, boardID string, opts ...Option) error {
	settings := &SaveOptions{}
	for _, opt := range opts {
		opt(settings)
	}

	if err := settings.Validate(); err != nil {
		s.logEvent(ctx, config.Error, "Invalid settings options", teamID, boardID, err)
		return err
	}

	compositeKey := s.createCompositeKey(teamID, boardID)
	existingSettings, err := s.storageService.Find(ctx, compositeKey)
	if err != nil && !errors.Is(err, pg.ErrNoRowsAffected) {
		s.logEvent(ctx, config.Error, "Failed to retrieve existing settings", teamID, boardID, err)
		return ErrSettingsRetrievalError
	}

	s.logEvent(ctx, config.Debug, "Validating document server", teamID, boardID, nil)
	if settings.Address != "" && settings.Header != "" && settings.Secret != "" {
		token, err := s.jwtService.Create(jwt.MapClaims{
			"c":   "version",
			"exp": jwt.NewNumericDate(time.Now().Add(time.Minute * 1)),
			"iat": jwt.NewNumericDate(time.Now()),
		}, []byte(settings.Secret))

		if err != nil {
			s.logEvent(ctx, config.Error, "Failed to create JWT token", teamID, boardID, err)
			return ErrSettingsBadJwtError
		}

		response, err := s.docServerClient.GetServerVersion(ctx, settings.Address, docserver.WithHeader(settings.Header), docserver.WithToken(token))
		if err != nil {
			s.logEvent(ctx, config.Error, "Failed to connect to document server", teamID, boardID, err)
			return ErrDocumentServerVersionRetrievalError
		}

		if response.Error != 0 {
			err := fmt.Errorf("received non-zero error code from docserver: %d", response.Error)
			s.logEvent(ctx, config.Error, "Document server returned error", teamID, boardID, err)
			return ErrDocumentServerVersionRetrievalError
		}

		if err := validateDocServerVersion(response.Version); err != nil {
			s.logEvent(ctx, config.Error, "Unsupported document server version", teamID, boardID, err)
			return ErrDocumentServerUnsupportedVersion
		}

		s.logEvent(ctx, config.Debug, fmt.Sprintf("Valid document server detected, version: %s", response.Version), teamID, boardID, nil)
	}

	newSettings, err := s.buildNewSettings(teamID, settings, existingSettings)
	if err != nil {
		s.logEvent(ctx, config.Error, "Failed to build new settings", teamID, boardID, err)
		return ErrSettingsBuildNewSettingsError
	}

	s.invalidateCache(ctx, teamID, boardID)
	if _, err := s.storageService.Insert(ctx, compositeKey, newSettings); err != nil {
		s.logEvent(ctx, config.Error, "Failed to store settings", teamID, boardID, err)
		return ErrSettingsPersistenceError
	}

	s.logEvent(ctx, config.Debug, "Settings saved successfully", teamID, boardID, nil)
	return nil
}

func (s *settingsService) buildNewSettings(teamID string, opts *SaveOptions, existingSettings component.Settings) (component.Settings, error) {
	var newSettings component.Settings

	if opts.Demo {
		newSettings.Demo = s.createDemoSettings(teamID, existingSettings.Demo.Started)
		newSettings.DemoDetached = false

		if opts.Address != "" || opts.Header != "" || opts.Secret != "" {
			encSecret, err := s.encryptSecret(opts.Secret)
			if err != nil {
				return newSettings, err
			}

			newSettings.Address = opts.Address
			newSettings.Header = opts.Header
			newSettings.Secret = encSecret
		}

		return newSettings, nil
	}

	encSecret, err := s.encryptSecret(opts.Secret)
	if err != nil {
		return newSettings, err
	}

	newSettings = component.Settings{
		Address: opts.Address,
		Header:  opts.Header,
		Secret:  encSecret,
	}

	if opts.Address != "" || opts.Header != "" || opts.Secret != "" {
		newSettings.DemoDetached = true
	} else {
		newSettings.DemoDetached = existingSettings.DemoDetached
	}

	if !newSettings.DemoDetached && existingSettings.Demo.Enabled {
		newSettings.Demo = existingSettings.Demo
	}

	return newSettings, nil
}

func (s *settingsService) Find(ctx context.Context, teamID, boardID string) (component.Settings, error) {
	s.logEvent(ctx, config.Debug, "Looking up settings", teamID, boardID, nil)
	settings, found, err := s.getFromCache(ctx, teamID, boardID)
	if err != nil {
		s.logEvent(ctx, config.Warn, "Error processing cached settings", teamID, boardID, err)
	}

	if found {
		return settings, nil
	}

	s.logEvent(ctx, config.Debug, "Settings not found in cache, checking storage", teamID, boardID, nil)
	return s.getFromStorage(ctx, teamID, boardID)
}

func (s *settingsService) getFromCache(ctx context.Context, teamID, boardID string) (component.Settings, bool, error) {
	cacheKey := s.buildCacheKey(teamID, boardID)
	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err != nil {
		return component.Settings{}, false, err
	}

	if cachedData == nil {
		return component.Settings{}, false, nil
	}

	var settings component.Settings
	if err := json.Unmarshal(cachedData, &settings); err != nil {
		return component.Settings{}, false, err
	}

	if !settings.DemoDetached && settings.Demo.Enabled && (settings.Address == "" || settings.Header == "" || settings.Secret == "") {
		s.logEvent(ctx, config.Debug, "Demo settings retrieved from cache", teamID, boardID, nil)
		return settings, true, nil
	}

	if settings.DemoDetached {
		settings.Demo = component.Demo{}
	}

	if settings.Secret != "" {
		decSecret, err := s.decryptSecret(settings.Secret)
		if err != nil {
			return settings, false, err
		}
		settings.Secret = decSecret
	}

	s.logEvent(ctx, config.Debug, "Settings retrieved from cache", teamID, boardID, nil)
	return settings, true, nil
}

func (s *settingsService) getFromStorage(ctx context.Context, teamID, boardID string) (component.Settings, error) {
	compositeKey := s.createCompositeKey(teamID, boardID)
	settings, err := s.storageService.Find(ctx, compositeKey)

	if err != nil {
		if errors.Is(err, pg.ErrNoRowsAffected) {
			s.logEvent(ctx, config.Debug, "No settings found in storage", teamID, boardID, nil)
			return component.Settings{}, nil
		}
		s.logEvent(ctx, config.Error, "Failed to retrieve settings from storage", teamID, boardID, err)
		return component.Settings{}, err
	}

	s.logEvent(ctx, config.Debug, "Settings retrieved from storage", teamID, boardID, nil)
	s.cacheSettings(ctx, teamID, boardID, settings)

	if !settings.DemoDetached && settings.Demo.Enabled && (settings.Address == "" || settings.Header == "" || settings.Secret == "") {
		return settings, nil
	}

	if settings.DemoDetached {
		settings.Demo = component.Demo{}
	}

	if settings.Secret != "" {
		decSecret, err := s.decryptSecret(settings.Secret)
		if err != nil {
			s.logEvent(ctx, config.Error, "Failed to decrypt secret", teamID, boardID, err)
			return settings, err
		}
		settings.Secret = decSecret
	}

	return settings, nil
}
