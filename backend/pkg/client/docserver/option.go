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
package docserver

type ClientOptions struct {
	Token  string
	Header string
}

func DefaultClientOptions() *ClientOptions {
	return &ClientOptions{
		Token:  "",
		Header: "",
	}
}

func (o *ClientOptions) Validate() error {
	return nil
}

type Option func(*ClientOptions)

func WithToken(token string) Option {
	return func(o *ClientOptions) {
		o.Token = token
	}
}

func WithHeader(header string) Option {
	return func(o *ClientOptions) {
		o.Header = header
	}
}

func ApplyOptions(o *ClientOptions, opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}
