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

import "errors"

var (
	ErrSettingsAddressRequired             = errors.New("features.settings.form.errors.address_required")
	ErrSettingsSecretRequired              = errors.New("features.settings.form.errors.secret_required")
	ErrSettingsHeaderRequired              = errors.New("features.settings.form.errors.header_required")
	ErrSettingsInvalidURL                  = errors.New("features.settings.form.errors.invalid_url")
	ErrSettingsInvalidProtocol             = errors.New("features.settings.form.errors.invalid_protocol")
	ErrSettingsTrailingSlash               = errors.New("features.settings.form.errors.trailing_slash")
	ErrSettingsHeaderTooLong               = errors.New("features.settings.form.errors.header_too_long")
	ErrSettingsSecretTooLong               = errors.New("features.settings.form.errors.secret_too_long")
	ErrSettingsRetrievalError              = errors.New("features.settings.form.errors.retrieval_error")
	ErrSettingsBadJwtError                 = errors.New("features.settings.form.errors.bad_jwt")
	ErrDocumentServerVersionRetrievalError = errors.New("features.settings.form.errors.document_server_version_retrieval_error")
	ErrDocumentServerUnsupportedVersion    = errors.New("features.settings.form.errors.document_server_unsupported_version")
	ErrSettingsBuildNewSettingsError       = errors.New("features.settings.form.errors.settings_initialization_error")
	ErrSettingsPersistenceError            = errors.New("features.settings.form.errors.settings_persistence_error")
)
