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

import (
	"fmt"
	"strings"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/common"
)

type CreateFileRequest struct {
	BoardID  string
	Name     string
	Type     common.DocumentType
	Language string
	Token    string
}

func (r *CreateFileRequest) Validate() error {
	if strings.TrimSpace(string(r.Type)) == "" {
		return fmt.Errorf("documentType is required")
	}

	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("name is required")
	}

	if strings.TrimSpace(r.BoardID) == "" {
		return fmt.Errorf("boardID is required")
	}

	if strings.TrimSpace(r.Token) == "" {
		return fmt.Errorf("token is required")
	}

	if strings.TrimSpace(r.Language) == "" {
		r.Language = "en-US"
	}

	return nil
}

type FileUploadRequest struct {
	Data FileUploadRequestData `json:"data"`
}

type FileUploadRequestData struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type GetBoardMemberRequest struct {
	BoardID  string
	MemberID string
	Token    string
}

func (r *GetBoardMemberRequest) Validate() error {
	if strings.TrimSpace(r.BoardID) == "" {
		return fmt.Errorf("boardID is required")
	}

	if strings.TrimSpace(r.MemberID) == "" {
		return fmt.Errorf("memberID is required")
	}

	if strings.TrimSpace(r.Token) == "" {
		return fmt.Errorf("token is required")
	}

	return nil
}

type GetBoardRequest struct {
	BoardID string
	Token   string
}

func (r *GetBoardRequest) Validate() error {
	if strings.TrimSpace(r.BoardID) == "" {
		return fmt.Errorf("boardID is required")
	}

	if strings.TrimSpace(r.Token) == "" {
		return fmt.Errorf("token is required")
	}

	return nil
}

type GetFileInfoRequest struct {
	BoardID string `json:"board_id"`
	ItemID  string `json:"item_id"`
	Token   string `json:"token"`
}

func (r *GetFileInfoRequest) Validate() error {
	if strings.TrimSpace(r.BoardID) == "" {
		return fmt.Errorf("boardID is required")
	}

	if strings.TrimSpace(r.ItemID) == "" {
		return fmt.Errorf("itemID is required")
	}

	if strings.TrimSpace(r.Token) == "" {
		return fmt.Errorf("token is required")
	}

	return nil
}

type GetFilePublicURLRequest struct {
	URL   string
	Token string
}

func (r *GetFilePublicURLRequest) Validate() error {
	if strings.TrimSpace(r.URL) == "" {
		return fmt.Errorf("url is required")
	}

	if strings.TrimSpace(r.Token) == "" {
		return fmt.Errorf("token is required")
	}

	return nil
}

type GetFilesInfoRequest struct {
	Cursor  string `json:"cursor"`
	BoardID string `json:"board_id"`
	Token   string `json:"token"`
}

func (r *GetFilesInfoRequest) Validate() error {
	if strings.TrimSpace(r.BoardID) == "" {
		return fmt.Errorf("boardID is required")
	}

	if strings.TrimSpace(r.Token) == "" {
		return fmt.Errorf("token is required")
	}

	return nil
}

type GetUserInfoRequest struct {
	Token string
}

func (r *GetUserInfoRequest) Validate() error {
	if strings.TrimSpace(r.Token) == "" {
		return fmt.Errorf("token is required")
	}

	return nil
}

type UploadFileRequest struct {
	BoardID  string
	ItemID   string
	Filename string
	FileURL  string
	Token    string
}

func (r *UploadFileRequest) Validate() error {
	if strings.TrimSpace(r.BoardID) == "" {
		return fmt.Errorf("boardID is required")
	}

	if strings.TrimSpace(r.ItemID) == "" {
		return fmt.Errorf("itemID is required")
	}

	if strings.TrimSpace(r.FileURL) == "" {
		return fmt.Errorf("fileURL is required")
	}

	if strings.TrimSpace(r.Token) == "" {
		return fmt.Errorf("token is required")
	}

	return nil
}
