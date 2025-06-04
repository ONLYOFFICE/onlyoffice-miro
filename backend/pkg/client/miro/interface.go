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

import "context"

type Client interface {
	GetBoard(ctx context.Context, req GetBoardRequest) (*BoardResponse, error)
	GetBoardMember(ctx context.Context, req GetBoardMemberRequest) (*BoardMemberResponse, error)
	GetFileInfo(ctx context.Context, req GetFileInfoRequest) (*FileInfoResponse, error)
	GetFilesInfo(ctx context.Context, req GetFilesInfoRequest) (*FilesInfoResponse, error)
	GetFilePublicURL(ctx context.Context, req GetFilePublicURLRequest) (*FileLocationResponse, error)
	GetUserInfo(ctx context.Context, req GetUserInfoRequest) (*UserInfoResponse, error)

	CreateFile(ctx context.Context, req CreateFileRequest) (*FileCreatedResponse, error)
	UploadFile(ctx context.Context, req UploadFileRequest) (*FileLocationResponse, error)
}
