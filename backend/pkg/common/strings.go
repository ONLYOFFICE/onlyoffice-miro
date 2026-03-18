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
package common

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	language "golang.org/x/text/language"
)

func Concat(strs ...string) string { return strings.Join(strs, "") }

func ToTemplateLanguage(lang string) string {
	tag, err := language.Parse(lang)
	if err != nil {
		return "en-US"
	}

	base, _ := tag.Base()
	baseStr := base.String()

	region, _ := tag.Region()
	regionStr := region.String()

	if baseStr == "zh" {
		if regionStr == "TW" || regionStr == "HK" {
			return "zh-TW"
		}

		return "zh-CN"
	}

	if regionStr != "" {
		return fmt.Sprintf("%s-%s", baseStr, regionStr)
	}

	if likely, err := language.Parse(lang + "-US"); err == nil {
		if region, _ := likely.Region(); region.String() != "" {
			return fmt.Sprintf("%s-%s", baseStr, region.String())
		}
	}

	return "en-US"
}

func GenerateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}

func ToDocumentType(ftype string) DocumentType {
	switch ftype {
	case string(DOCX):
		return DOCX
	case string(PPTX):
		return PPTX
	case string(XLSX):
		return XLSX
	default:
		return DOCX
	}
}
