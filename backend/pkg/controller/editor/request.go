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

import "github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/client/miro"

type editorRequestParams struct {
	uid  string
	tid  string
	bid  string
	fid  string
	lang string
}

type builderRequest struct {
	Board string
	File  miro.FileInfoResponse
}

func (r builderRequest) ID() string {
	return r.File.ID
}

func (r builderRequest) FolderID() string {
	return r.Board
}

func (r builderRequest) Title() string {
	return r.File.Data.Title
}

func (r builderRequest) URL() string {
	return r.File.Data.DocumentURL
}

func (r builderRequest) ModifiedAt() string {
	return r.File.ModifiedAt
}
