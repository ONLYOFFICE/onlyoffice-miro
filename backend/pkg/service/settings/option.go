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
package settings

import (
	"net/url"
	"strings"
)

type SaveOptions struct {
	Address string
	Header  string
	Secret  string
	Demo    bool
}

func (o *SaveOptions) Validate() error {
	if !o.Demo && o.Address == "" {
		return ErrSettingsAddressRequired
	}

	if !o.Demo && o.Secret == "" {
		return ErrSettingsSecretRequired
	}

	if !o.Demo && o.Header == "" {
		return ErrSettingsHeaderRequired
	}

	if o.Address != "" {
		u, err := url.Parse(o.Address)
		if err != nil {
			return ErrSettingsInvalidURL
		}

		if u.Scheme != "https" {
			return ErrSettingsInvalidProtocol
		}

		if strings.HasSuffix(o.Address, "/") {
			return ErrSettingsTrailingSlash
		}
	}

	if len(o.Header) > 255 {
		return ErrSettingsHeaderTooLong
	}

	if len(o.Secret) > 255 {
		return ErrSettingsSecretTooLong
	}

	return nil
}

type Option func(*SaveOptions)

func WithAddress(val string) Option {
	return func(o *SaveOptions) {
		o.Address = val
	}
}

func WithHeader(val string) Option {
	return func(o *SaveOptions) {
		o.Header = val
	}
}

func WithSecret(val string) Option {
	return func(o *SaveOptions) {
		o.Secret = val
	}
}

func WithDemo(val bool) Option {
	return func(o *SaveOptions) {
		o.Demo = val
	}
}
