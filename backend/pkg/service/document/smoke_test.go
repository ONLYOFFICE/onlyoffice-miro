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
	"testing"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/crypto"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/service/logger"
)

type smokeDoc struct{ id, title, modified string }

func (d smokeDoc) ID() string         { return d.id }
func (d smokeDoc) FolderID() string   { return "board" }
func (d smokeDoc) Title() string      { return d.title }
func (d smokeDoc) URL() string        { return "https://example.com/file" }
func (d smokeDoc) ModifiedAt() string { return d.modified }

func newSmokeBuilder(t *testing.T) BuilderService {
	t.Helper()
	l := logger.NewNoopLogger()
	fm, err := NewMapFormatManager()
	if err != nil {
		t.Fatalf("format manager: %v (document-formats submodule initialized?)", err)
	}
	return NewBuilderService(NewModificationKeyGenerator(l), NewJwtSignatureGenerator(l), fm, l)
}

func TestSmokeEditorConfigPermissions(t *testing.T) {
	b := newSmokeBuilder(t)
	cases := map[string]struct {
		docType string
		edit    bool
	}{
		"a.docx": {"word", true},
		"a.xlsx": {"cell", true},
		"a.pptx": {"slide", true},
		"a.doc":  {"word", false},
	}
	for title, want := range cases {
		cfg, err := b.Build(context.Background(), "https://cb", smokeDoc{"1", title, "t1"})
		if err != nil {
			t.Fatalf("%s: %v", title, err)
		}
		if cfg.DocumentType != want.docType || cfg.Document.Permissions.Edit != want.edit {
			t.Errorf("%s: type=%s edit=%v, want %s/%v", title, cfg.DocumentType, cfg.Document.Permissions.Edit, want.docType, want.edit)
		}
		if cfg.Document.Permissions.Chat || cfg.Document.Permissions.Comment {
			t.Errorf("%s: chat/comments must stay disabled", title)
		}
	}
	if _, err := b.Build(context.Background(), "https://cb", smokeDoc{"1", "a.exe", "t1"}); err == nil {
		t.Error("unsupported format must fail")
	}
}

func TestSmokeEditorConfigSigned(t *testing.T) {
	b := newSmokeBuilder(t)
	secret := []byte("smoke-secret")
	cfg, err := b.Build(context.Background(), "https://cb", smokeDoc{"1", "a.docx", "t1"}, WithKey(secret))
	if err != nil {
		t.Fatal(err)
	}
	jwt := crypto.NewJwtService()
	if _, err := jwt.Validate(cfg.Token, secret); err != nil {
		t.Errorf("token not valid with the same secret: %v", err)
	}
	if _, err := jwt.Validate(cfg.Token, []byte("other")); err == nil {
		t.Error("token must not validate with another secret")
	}
}

func TestSmokeDocumentKey(t *testing.T) {
	g := NewModificationKeyGenerator(logger.NewNoopLogger())
	k1, _ := g.Generate(context.Background(), smokeDoc{"1", "a.docx", "t1"})
	k2, _ := g.Generate(context.Background(), smokeDoc{"1", "a.docx", "t1"})
	k3, _ := g.Generate(context.Background(), smokeDoc{"1", "a.docx", "t2"})
	if k1 != k2 || k1 == k3 {
		t.Errorf("key must depend on id+modifiedAt only: %s %s %s", k1, k2, k3)
	}
}
