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
package miro

type AuthenticationResponse struct {
	UserID       string `json:"user_id"`
	TeamID       string `json:"team_id"`
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IssuedAt     int    `json:"issued_at"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

type BoardMemberResponse struct {
	MemberID   string `json:"id"`
	MemberName string `json:"name"`
	Role       string `json:"role"`
	Lang       string `json:"lang,omitempty"`
}

func (b *BoardMemberResponse) ID() string {
	return b.MemberID
}

func (b *BoardMemberResponse) Name() string {
	return b.MemberName
}

func (b *BoardMemberResponse) Language() string {
	if b.Lang == "" {
		return "en"
	}

	return b.Lang
}

type BoardResponse struct {
	ID string `json:"id"`
}

type FileCreatedResponse struct {
	ID         string `json:"id"`
	CreatedAt  string `json:"createdAt"`
	ModifiedAt string `json:"modifiedAt"`
	Links      struct {
		Self string `json:"self"`
	} `json:"links"`
}

type FileInfoResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data struct {
		Title       string `json:"title"`
		DocumentURL string `json:"documentUrl"`
	} `json:"data"`
	CreatedAt  string `json:"createdAt"`
	ModifiedAt string `json:"modifiedAt"`
}

type FileLocationResponse struct {
	URL string `json:"url"`
}

type FilesInfoResponse struct {
	Size   int                `json:"size"`
	Limit  int                `json:"limit"`
	Total  int                `json:"total"`
	Data   []FileInfoResponse `json:"data"`
	Cursor string             `json:"cursor,omitempty"`
}

type GenericFileResponse struct {
	ID string `json:"id"`
}

type UserInfoResponse struct {
	User struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
}
