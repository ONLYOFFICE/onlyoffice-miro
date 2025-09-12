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
package document

import (
	"context"
	"encoding/json"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
)

type BuilderService interface {
	Build(ctx context.Context, callbackUrl string,
		configurer DocumentConfigurer, opts ...BuilderOption) (*Config, error)
}

type builderService struct {
	keyGenerator       KeyGenerator
	signatureGenerator SignatureGenerator
	formatManager      FormatManager
	logger             service.Logger
}

func NewBuilderService(
	keyGenerator KeyGenerator,
	signatureGenerator SignatureGenerator,
	formatManager FormatManager,
	logger service.Logger,
) BuilderService {
	return &builderService{
		keyGenerator:       keyGenerator,
		signatureGenerator: signatureGenerator,
		formatManager:      formatManager,
		logger:             logger,
	}
}

func (s *builderService) Build(
	ctx context.Context,
	callbackUrl string,
	configurer DocumentConfigurer,
	opts ...BuilderOption,
) (*Config, error) {
	s.logger.Debug(ctx, "Building document config", service.Fields{"callbackUrl": callbackUrl})

	options := BuilderOptions{mode: Desktop}
	for _, option := range opts {
		option(&options)
	}

	title := configurer.Title()
	ext := s.formatManager.GetFileExt(title)

	s.logger.Debug(ctx, "Document info", service.Fields{"title": title, "extension": ext})
	key, err := s.keyGenerator.Generate(ctx, configurer)
	if err != nil {
		s.logger.Error(ctx, "Failed to generate key", service.Fields{"error": err.Error()})
		return nil, err
	}

	format, exists := s.formatManager.GetFormatByName(ext)
	if !exists {
		s.logger.Error(ctx, "Unsupported format", service.Fields{"extension": ext})
		return nil, ErrUnsupportedFormat
	}

	s.logger.Debug(ctx, "Format determined", service.Fields{"documentType": format.Type, "isEditable": format.IsEditable()})
	config := &Config{
		Document: Document{
			Key:      key,
			Title:    title,
			URL:      configurer.URL(),
			FileType: ext,
			Permissions: Permissions{
				Edit:      format.IsEditable(),
				Protect:   false,
				Chat:      false,
				Comment:   false,
				FillForms: false,
				Review:    false,
			},
		},
		Editor: Editor{
			CallbackURL: callbackUrl,
		},
		DocumentType: format.Type,
		Type:         string(options.mode),
	}

	if options.userConfigurer != nil {
		uconfigurer := options.userConfigurer
		config.Editor.User = User{
			ID:   uconfigurer.ID(),
			Name: uconfigurer.Name(),
		}

		config.Editor.Lang = uconfigurer.Language()
		s.logger.Debug(ctx, "User configuration applied", service.Fields{
			"userId":   uconfigurer.ID(),
			"language": uconfigurer.Language(),
		})
	}

	if len(options.key) > 0 {
		s.logger.Debug(ctx, "Signing configuration")
		if err := s.signConfig(config, options.key); err != nil {
			s.logger.Error(ctx, "Failed to sign configuration", service.Fields{"error": err.Error()})
			return nil, err
		}
	}

	s.logger.Debug(ctx, "Document config built successfully", service.Fields{"key": key})
	return config, nil
}

func (s *builderService) signConfig(config *Config, secret []byte) error {
	s.logger.Debug(context.Background(), "Signing config")
	buf, err := json.Marshal(config)
	if err != nil {
		s.logger.Error(context.Background(), "Failed to marshal config", service.Fields{"error": err.Error()})
		return err
	}

	token, err := s.signatureGenerator.Sign(secret, buf)
	if err != nil {
		s.logger.Error(context.Background(), "Failed to sign config", service.Fields{"error": err.Error()})
		return err
	}

	config.Token = token
	s.logger.Debug(context.Background(), "Config signed successfully")
	return nil
}
