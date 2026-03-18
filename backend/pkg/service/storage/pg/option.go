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
package pg

import (
	"runtime"
	"time"
)

type Options struct {
	MaxConnections              int
	MinConnections              int
	ConnectionTimeout           time.Duration
	ConnectionIdleTimeout       time.Duration
	ConnectionHealthcheckPeriod time.Duration
	MaxConnLifetime             time.Duration
	MaxRetries                  int
	RetryInterval               time.Duration
}

type Option func(*Options)

func WithMaxConnections(val int) Option {
	return func(o *Options) {
		if val <= 0 || val > runtime.NumCPU() {
			o.MaxConnections = runtime.NumCPU()
			return
		}

		o.MaxConnections = val
	}
}

func WithMinConnections(val int) Option {
	return func(o *Options) {
		if val <= 0 || val > runtime.NumCPU() {
			o.MinConnections = 1
			return
		}

		o.MinConnections = val
	}
}

func WithConnectionTimeout(val time.Duration) Option {
	return func(o *Options) {
		o.ConnectionTimeout = val
	}
}

func WithConnectionIdleTimeout(val time.Duration) Option {
	return func(o *Options) {
		o.ConnectionIdleTimeout = val
	}
}

func WithConnectionHealthcheckPeriod(val time.Duration) Option {
	return func(o *Options) {
		o.ConnectionHealthcheckPeriod = val
	}
}

func WithMaxConnLifetime(val time.Duration) Option {
	return func(o *Options) {
		o.MaxConnLifetime = val
	}
}

func WithMaxRetries(val int) Option {
	return func(o *Options) {
		if val < 0 {
			o.MaxRetries = 0
			return
		}

		o.MaxRetries = val
	}
}

func WithRetryInterval(val time.Duration) Option {
	return func(o *Options) {
		if val < 100*time.Millisecond {
			o.RetryInterval = 100 * time.Millisecond
			return
		}

		o.RetryInterval = val
	}
}
