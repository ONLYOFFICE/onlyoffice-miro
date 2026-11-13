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
package callback

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/config"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/crypto"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/common"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/logger"
	echo "github.com/labstack/echo/v4"
)

func callCallback(t *testing.T, query, body string) *httptest.ResponseRecorder {
	t.Helper()
	h := NewCallbackController(&config.Config{}, nil, crypto.NewJwtService(), nil, nil, logger.NewNoopLogger()).
		Handlers()[common.MethodPost]
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/callback"+query, strings.NewReader(body))
	rec := httptest.NewRecorder()
	if err := h(e.NewContext(req, rec)); err != nil {
		t.Fatal(err)
	}
	return rec
}

const fullQuery = "?fid=f&uid=u&tid=t&bid=b&filename=a.docx"

func TestSmokeCallbackNonSaveStatusAcknowledged(t *testing.T) {
	rec := callCallback(t, fullQuery, `{"status":1,"token":"x"}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"error":"0"`) {
		t.Errorf("got %d %s", rec.Code, rec.Body.String())
	}
}

func TestSmokeCallbackRejectsBadRequests(t *testing.T) {
	cases := map[string][2]string{
		"missing params": {"?fid=f", `{"status":1,"token":"x"}`},
		"missing token":  {fullQuery, `{"status":2}`},
		"bad json":       {fullQuery, `{`},
	}
	for name, c := range cases {
		rec := callCallback(t, c[0], c[1])
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), `"error":"1"`) {
			t.Errorf("%s: got %d %s", name, rec.Code, rec.Body.String())
		}
	}
}
