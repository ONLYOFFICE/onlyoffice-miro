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
package oauth

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

type ExchangeTokenError struct{ BaseError }
type RefreshTokenError struct{ BaseError }

type Errors struct {
	FailedToCreateRequest  func(err error) error
	FailedToSendRequest    func(err error) error
	FailedToDecodeResponse func(err error) error
	RequestFailed          func(statusCode int) error

	FailedToExchangeToken func(err error) error
	FailedToRefreshToken  func(err error) error
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
				message: common.Concat("Failed to perform HTTP request"),
				err:     err,
			}}
		},
		FailedToDecodeResponse: func(err error) error {
			return &DecodeResponseError{BaseError{
				message: common.Concat("Failed to decode response"),
				err:     err,
			}}
		},
		RequestFailed: func(statusCode int) error {
			return &RequestFailedError{BaseError{
				message: common.Concat("Unexpected HTTP status ", strconv.Itoa(statusCode)),
			}}
		},
		FailedToExchangeToken: func(err error) error {
			return &ExchangeTokenError{BaseError{
				message: common.Concat("Failed to exchange authorization code for token"),
				err:     err,
			}}
		},
		FailedToRefreshToken: func(err error) error {
			return &RefreshTokenError{BaseError{
				message: common.Concat("Failed to refresh token"),
				err:     err,
			}}
		},
	}
}
