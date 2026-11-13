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

import "testing"

func TestSmokeToTemplateLanguage(t *testing.T) {
	cases := map[string]string{
		"en": "en-US", "de": "de-DE", "es": "es-ES", "fr": "fr-FR",
		"ja_JP": "ja-JP", "ko": "ko-KR", "pl": "pl-PL", "pt_BR": "pt-BR",
		"": "en-US",
	}
	for in, want := range cases {
		if got := ToTemplateLanguage(in); got != want {
			t.Errorf("ToTemplateLanguage(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSmokeToDocumentType(t *testing.T) {
	cases := map[string]DocumentType{"docx": DOCX, "xlsx": XLSX, "pptx": PPTX, "unknown": DOCX}
	for in, want := range cases {
		if got := ToDocumentType(in); got != want {
			t.Errorf("ToDocumentType(%q) = %q, want %q", in, got, want)
		}
	}
}
