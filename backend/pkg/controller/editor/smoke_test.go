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
package editor

import (
	"errors"
	"testing"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/config"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/core/component"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/controller/base"
)

func smokeEditor() *editorController {
	return &editorController{BaseController: base.BaseController{Config: &config.Config{
		DemoServer: &config.DemoServerConfig{Address: "https://demo", Header: "Authorization", Secret: "demo", Days: 30},
	}}}
}

func TestSmokeResolveServerSettings(t *testing.T) {
	c := smokeEditor()
	recent := time.Now().Add(-24 * time.Hour)
	expired := time.Now().Add(-31 * 24 * time.Hour)

	addr, secret, err := c.resolveServerSettings(&component.Settings{Address: "https://own", Secret: "s"})
	if err != nil || addr != "https://own" || secret != "s" {
		t.Errorf("own server: %s %s %v", addr, secret, err)
	}

	addr, _, err = c.resolveServerSettings(&component.Settings{Demo: component.Demo{Enabled: true, Started: &recent}})
	if err != nil || addr != "https://demo" {
		t.Errorf("active demo: %s %v", addr, err)
	}

	_, _, err = c.resolveServerSettings(&component.Settings{Demo: component.Demo{Enabled: true, Started: &expired}})
	if !errors.Is(err, base.ErrSettingsNotConfigured) {
		t.Errorf("expired demo must be rejected, got %v", err)
	}

	_, _, err = c.resolveServerSettings(&component.Settings{})
	if !errors.Is(err, base.ErrSettingsNotConfigured) {
		t.Errorf("empty settings must be rejected, got %v", err)
	}
}

func TestSmokeCallbackURLEscapesFilename(t *testing.T) {
	got := buildCallbackURL("https://b/api/callback", "f", "u", "t", "b", "my file&x.docx")
	want := "https://b/api/callback?fid=f&uid=u&tid=t&bid=b&filename=my+file%26x.docx"
	if got != want {
		t.Errorf("got %s", got)
	}
}
