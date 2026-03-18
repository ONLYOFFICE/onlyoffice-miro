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
package document

type BuilderOptions struct {
	key            []byte
	userConfigurer UserConfigurer
	mode           EditorMode
}

type BuilderOption func(*BuilderOptions)

func WithKey(val []byte) BuilderOption {
	return func(o *BuilderOptions) {
		if len(val) > 0 {
			o.key = val
		}
	}
}

func WithUserConfigurer(val UserConfigurer) BuilderOption {
	return func(o *BuilderOptions) {
		if val != nil {
			o.userConfigurer = val
		}
	}
}

func WithEditorMode(val EditorMode) BuilderOption {
	return func(o *BuilderOptions) {
		o.mode = val
	}
}
