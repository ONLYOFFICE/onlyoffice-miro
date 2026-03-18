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
package translation

import (
	"context"
	"encoding/json"
	"io/fs"
	"path"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/assets"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	i18n "github.com/nicksnyder/go-i18n/v2/i18n"
	language "golang.org/x/text/language"
)

type Translation struct {
	bundle      *i18n.Bundle
	defaultLang string
	langs       []string
	logger      service.Logger
}

func NewTranslation(defaultLang string, logger service.Logger) (service.TranslationProvider, error) {
	bundle := i18n.NewBundle(language.Make(defaultLang))
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	langs := []string{}

	entries, err := fs.ReadDir(assets.Translations, "translations")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		ext := path.Ext(filename)
		if ext != ".json" {
			continue
		}

		lang := filename[:len(filename)-len(ext)]
		langs = append(langs, lang)

		logger.Debug(context.Background(), "Loading translation file", service.Fields{
			"filename": filename,
			"language": lang,
		})

		file, err := fs.ReadFile(assets.Translations, path.Join("translations", filename))
		if err != nil {
			logger.Error(context.Background(), "Failed to read translation file", service.Fields{
				"filename": filename,
				"error":    err,
			})
			return nil, err
		}

		_, err = bundle.ParseMessageFileBytes(file, filename)
		if err != nil {
			logger.Error(context.Background(), "Failed to parse translation file", service.Fields{
				"filename": filename,
				"error":    err,
			})
			return nil, err
		}

		logger.Debug(context.Background(), "Successfully loaded translation file", service.Fields{
			"language": lang,
		})
	}

	return &Translation{
		bundle:      bundle,
		defaultLang: defaultLang,
		langs:       langs,
		logger:      logger,
	}, nil
}

func (t *Translation) Translate(ctx context.Context, lang, id string) string {
	localizer := i18n.NewLocalizer(t.bundle, lang, t.defaultLang)

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: id,
	})

	if err != nil {
		t.logger.Debug(ctx, "Translation not found", service.Fields{
			"id":       id,
			"language": lang,
		})
		return id
	}

	return msg
}
