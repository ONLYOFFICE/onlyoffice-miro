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

const (
	Desktop  EditorMode = "desktop"
	Embedded EditorMode = "embedded"
	Mobile   EditorMode = "mobile"
)

type EditorMode string

type DocumentConfigurer interface {
	ID() string
	FolderID() string
	Title() string
	URL() string
	ModifiedAt() string
}

type UserConfigurer interface {
	ID() string
	Name() string
	Language() string
}

type Permissions struct {
	Chat                    bool `json:"chat"`
	Comment                 bool `json:"comment"`
	Copy                    bool `json:"copy"`
	DeleteCommentAuthorOnly bool `json:"deleteCommentAuthorOnly"`
	Download                bool `json:"download"`
	Edit                    bool `json:"edit"`
	EditCommentAuthorOnly   bool `json:"editCommentAuthorOnly"`
	FillForms               bool `json:"fillForms"`
	ModifyContentControl    bool `json:"modifyContentControl"`
	ModifyFilter            bool `json:"modifyFilter"`
	Print                   bool `json:"print"`
	Review                  bool `json:"review"`
	Protect                 bool `json:"protect,omitempty"`
}

type Document struct {
	Key         string      `json:"key"`
	Title       string      `json:"title"`
	URL         string      `json:"url"`
	FileType    string      `json:"fileType"`
	Permissions Permissions `json:"permissions"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Customization struct {
	Goback struct {
		RequestClose bool `json:"requestClose"`
	} `json:"goback"`
	Plugins       bool `json:"plugins"`
	HideRightMenu bool `json:"hideRightMenu"`
}

type Editor struct {
	User        User   `json:"user"`
	CallbackURL string `json:"callbackUrl"`
	Lang        string `json:"lang"`
}

type Config struct {
	Document     Document `json:"document"`
	DocumentType string   `json:"documentType"`
	Editor       Editor   `json:"editorConfig"`
	Type         string   `json:"type"`
	Token        string   `json:"token"`
}
