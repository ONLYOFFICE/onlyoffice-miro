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
	"strconv"
	"strings"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/common"
)

type BaseError struct {
	message string
	err     error
}

func (e *BaseError) Error() string {
	if e.err == nil {
		return e.message
	}

	var b strings.Builder
	b.Grow(len(e.message) + 2 + len(e.err.Error()))
	b.WriteString(e.message)
	b.WriteString(": ")
	b.WriteString(e.err.Error())
	return b.String()
}

func (e *BaseError) Unwrap() error {
	return e.err
}

type CreateRequestError struct{ BaseError }
type SendRequestError struct{ BaseError }
type DecodeResponseError struct{ BaseError }
type RequestFailedError struct{ BaseError }

type GetFileInfoError struct{ BaseError }
type GetFilePublicURLError struct{ BaseError }
type UploadFileError struct{ BaseError }
type MarshalRequestError struct{ BaseError }
type ReadResponseError struct{ BaseError }

type GetBoardError struct{ BaseError }
type GetBoardMemberError struct{ BaseError }
type ReadFileError struct{ BaseError }
type CreateFormFileError struct{ BaseError }
type WriteFileDataError struct{ BaseError }
type CloseWriterError struct{ BaseError }

type Errors struct {
	FailedToCreateRequest  func(err error) error
	FailedToSendRequest    func(err error) error
	FailedToDecodeResponse func(err error) error
	FailedToReadResponse   func(err error) error
	RequestFailed          func(statusCode int) error
	FailedToGetFileInfo    func(err error) error
	FailedToGetFileURL     func(err error) error
	FailedToUploadFile     func(err error) error
	FailedToMarshalRequest func(err error) error
	FailedToGetBoard       func(err error) error
	FailedToGetBoardMember func(err error) error
	FailedToReadFile       func(err error) error
	FailedToCreateFormFile func(err error) error
	FailedToWriteFileData  func(err error) error
	FailedToCloseWriter    func(err error) error
}

func NewErrors() *Errors {
	return &Errors{
		FailedToCreateRequest: func(err error) error {
			return &CreateRequestError{BaseError{
				message: common.Concat("Failed to create request"),
				err:     err,
			}}
		},
		FailedToSendRequest: func(err error) error {
			return &SendRequestError{BaseError{
				message: common.Concat("Failed to send request"),
				err:     err,
			}}
		},
		FailedToDecodeResponse: func(err error) error {
			return &DecodeResponseError{BaseError{
				message: common.Concat("Failed to decode response"),
				err:     err,
			}}
		},
		FailedToReadResponse: func(err error) error {
			return &ReadResponseError{BaseError{
				message: common.Concat("Failed to read response"),
				err:     err,
			}}
		},
		RequestFailed: func(statusCode int) error {
			return &RequestFailedError{BaseError{
				message: common.Concat("Request failed with status ", strconv.Itoa(statusCode)),
			}}
		},
		FailedToGetFileInfo: func(err error) error {
			return &GetFileInfoError{BaseError{
				message: common.Concat("Failed to get file info"),
				err:     err,
			}}
		},
		FailedToGetFileURL: func(err error) error {
			return &GetFilePublicURLError{BaseError{
				message: common.Concat("Failed to get file public URL"),
				err:     err,
			}}
		},
		FailedToUploadFile: func(err error) error {
			return &UploadFileError{BaseError{
				message: common.Concat("Failed to upload file"),
				err:     err,
			}}
		},
		FailedToMarshalRequest: func(err error) error {
			return &MarshalRequestError{BaseError{
				message: common.Concat("Failed to marshal request"),
				err:     err,
			}}
		},
		FailedToGetBoard: func(err error) error {
			return &GetBoardError{BaseError{
				message: common.Concat("Failed to get board"),
				err:     err,
			}}
		},
		FailedToGetBoardMember: func(err error) error {
			return &GetBoardMemberError{BaseError{
				message: common.Concat("Failed to get board member"),
				err:     err,
			}}
		},
		FailedToReadFile: func(err error) error {
			return &ReadFileError{BaseError{
				message: common.Concat("Failed to read file"),
				err:     err,
			}}
		},
		FailedToCreateFormFile: func(err error) error {
			return &CreateFormFileError{BaseError{
				message: common.Concat("Failed to create form file"),
				err:     err,
			}}
		},
		FailedToWriteFileData: func(err error) error {
			return &WriteFileDataError{BaseError{
				message: common.Concat("Failed to write file data"),
				err:     err,
			}}
		},
		FailedToCloseWriter: func(err error) error {
			return &CloseWriterError{BaseError{
				message: common.Concat("Failed to close writer"),
				err:     err,
			}}
		},
	}
}
