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
package cache

import (
	"time"
)

type CacheOptions struct {
	KeyPrefix         string
	DefaultExpiration time.Duration
}

func DefaultCacheOptions() *CacheOptions {
	return &CacheOptions{
		KeyPrefix:         "app:cache:",
		DefaultExpiration: 5 * time.Minute,
	}
}

func (o *CacheOptions) Validate() error {
	return nil
}

type Option func(*CacheOptions)

func WithKeyPrefix(prefix string) Option {
	return func(o *CacheOptions) {
		o.KeyPrefix = prefix
	}
}

func WithDefaultExpiration(duration time.Duration) Option {
	return func(o *CacheOptions) {
		o.DefaultExpiration = duration
	}
}

func ApplyOptions(o *CacheOptions, opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}
