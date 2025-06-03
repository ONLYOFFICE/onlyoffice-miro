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
	_ "embed"
	"encoding/json"
	"path/filepath"
	"strings"
	"sync"
)

//go:embed formats/onlyoffice-docs-formats.json
var rawFormatsData []byte

var (
	cache map[string]Format
	once  sync.Once
)

type Format struct {
	Name    string            `json:"name"`
	Type    string            `json:"type"`
	Actions map[string]string `json:"-"`
	Convert map[string]string `json:"-"`
	Mime    []string          `json:"mime"`
}

func (f Format) IsLossyEditable() bool {
	_, exists := f.Actions["lossy-edit"]
	return exists
}

func (f Format) IsEditable() bool {
	_, exists := f.Actions["edit"]
	return exists
}

func (f Format) IsViewable() bool {
	_, exists := f.Actions["view"]
	return exists
}

func (f Format) IsViewOnly() bool {
	_, exists := f.Actions["view"]
	return exists && len(f.Actions) == 1
}

func (f Format) IsFillable() bool {
	_, exists := f.Actions["fill"]
	return exists
}

func (f Format) IsAutoConvertable() bool {
	_, exists := f.Actions["auto-convert"]
	return exists
}

func (f Format) IsOpenXMLConvertable() bool {
	_, word := f.Convert["docx"]
	_, slide := f.Convert["pptx"]
	_, cell := f.Convert["xlsx"]
	return word || slide || cell
}

func (f Format) GetOpenXMLExtension() string {
	if f.Type == "cell" {
		return "xlsx"
	} else if f.Type == "slide" {
		return "pptx"
	} else {
		return "docx"
	}
}

type MapFormatManager struct {
	formats map[string]Format
}

func NewMapFormatManager() (FormatManager, error) {
	var manager MapFormatManager

	once.Do(func() {
		var rawFormats []struct {
			Name    string   `json:"name"`
			Type    string   `json:"type"`
			Actions []string `json:"actions"`
			Convert []string `json:"convert"`
			Mime    []string `json:"mime"`
		}

		if err := json.Unmarshal(rawFormatsData, &rawFormats); err != nil {
			return
		}

		cache = make(map[string]Format)
		for _, rawFormat := range rawFormats {
			actionsSet := make(map[string]string)
			for _, action := range rawFormat.Actions {
				actionsSet[action] = action
			}

			if _, exists := actionsSet["view"]; !exists {
				continue
			}

			convertSet := make(map[string]string)
			for _, conv := range rawFormat.Convert {
				convertSet[conv] = conv
			}

			cache[rawFormat.Name] = Format{
				Name:    rawFormat.Name,
				Type:    rawFormat.Type,
				Actions: actionsSet,
				Convert: convertSet,
				Mime:    rawFormat.Mime,
			}
		}
	})

	if cache == nil {
		var rawFormats []struct {
			Name    string   `json:"name"`
			Type    string   `json:"type"`
			Actions []string `json:"actions"`
			Convert []string `json:"convert"`
			Mime    []string `json:"mime"`
		}

		if err := json.Unmarshal(rawFormatsData, &rawFormats); err != nil {
			return manager, err
		}

		manager.formats = make(map[string]Format)
		for _, rawFormat := range rawFormats {
			actionsSet := make(map[string]string)
			for _, action := range rawFormat.Actions {
				actionsSet[action] = action
			}

			if _, exists := actionsSet["view"]; !exists {
				continue
			}

			convertSet := make(map[string]string)
			for _, conv := range rawFormat.Convert {
				convertSet[conv] = conv
			}

			manager.formats[rawFormat.Name] = Format{
				Name:    rawFormat.Name,
				Type:    rawFormat.Type,
				Actions: actionsSet,
				Convert: convertSet,
				Mime:    rawFormat.Mime,
			}
		}
	} else {
		manager.formats = cache
	}

	return manager, nil
}

type FormatManager interface {
	GetFileExt(filename string) string
	EscapeFileName(filename string) string
	GetFormatByName(name string) (Format, bool)
	GetAllFormats() map[string]Format
}

func (m MapFormatManager) GetFileExt(filename string) string {
	return strings.ReplaceAll(strings.ToLower(filepath.Ext(filename)), ".", "")
}

func (m MapFormatManager) EscapeFileName(filename string) string {
	f := strings.ReplaceAll(filename, "\\", ":")
	f = strings.ReplaceAll(f, "/", ":")
	return f
}

func (m MapFormatManager) GetFormatByName(name string) (Format, bool) {
	format, exists := m.formats[strings.ToLower(name)]
	return format, exists
}

func (m MapFormatManager) GetAllFormats() map[string]Format {
	return m.formats
}
