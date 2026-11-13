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
package assets

import (
	"encoding/json"
	"testing"
)

func TestSmokeDefaultTemplatesEmbedded(t *testing.T) {
	for _, lang := range []string{"default", "en-US", "de-DE", "ja-JP", "pt-BR"} {
		for _, ext := range []string{"docx", "xlsx", "pptx"} {
			p := "templates/" + lang + "/new." + ext
			data, err := Templates.ReadFile(p)
			if err != nil || len(data) == 0 {
				t.Errorf("template %s is missing (submodule not initialized?): %v", p, err)
			}
		}
	}
}

func TestSmokeTranslationsAreValidJSON(t *testing.T) {
	entries, err := Translations.ReadDir("translations")
	if err != nil || len(entries) == 0 {
		t.Fatalf("no embedded translations: %v", err)
	}
	for _, e := range entries {
		data, err := Translations.ReadFile("translations/" + e.Name())
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		var v map[string]any
		if err := json.Unmarshal(data, &v); err != nil {
			t.Errorf("%s is not valid JSON: %v", e.Name(), err)
		}
	}
}

func TestSmokeViewsEmbedded(t *testing.T) {
	for _, v := range []string{"editor.html", "exchange.html", "unauthorized.html"} {
		if _, err := Views.ReadFile("views/" + v); err != nil {
			t.Errorf("view %s missing: %v", v, err)
		}
	}
}
