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
package processor

import (
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/core"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/core/component"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	pgx "github.com/jackc/pgx/v5"
)

const (
	authSelectQuery = `SELECT token_type, access_token, refresh_token, expires_at, scope
FROM authentications
WHERE team_id = $1 AND user_id = $2;`

	authInsertQuery = `INSERT INTO authentications (team_id, user_id, token_type, access_token, refresh_token, expires_at, scope)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (team_id, user_id) DO UPDATE
SET token_type = EXCLUDED.token_type,
    access_token = EXCLUDED.access_token,
    refresh_token = EXCLUDED.refresh_token,
    expires_at = EXCLUDED.expires_at,
    scope = EXCLUDED.scope,
    updated_at = CURRENT_TIMESTAMP;`

	authUpdateQuery = `UPDATE authentications
SET token_type = $3,
    access_token = $4,
    refresh_token = $5,
    expires_at = $6,
    scope = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE team_id = $1 AND user_id = $2;`

	authDeleteQuery = `DELETE FROM authentications
WHERE team_id = $1 AND user_id = $2;`
)

type authenticationProcessor struct{}

func authenticationScanner(row pgx.Row) (*component.Authentication, error) {
	result := &component.Authentication{}

	if err := row.Scan(
		&result.TokenType,
		&result.AccessToken,
		&result.RefreshToken,
		&result.ExpiresAt,
		&result.Scope,
	); err != nil {
		return nil, err
	}

	return result, nil
}

func NewAuthenticationProcessor() service.StorageProcessor[core.AuthCompositeKey, component.Authentication, pgx.Row] {
	return &authenticationProcessor{}
}

func (s authenticationProcessor) TableName() string {
	return "authentications"
}

func (s authenticationProcessor) BuildSelectQuery(id core.AuthCompositeKey) (string, []any, func(row pgx.Row) (component.Authentication, error)) {
	return authSelectQuery, []any{id.TeamID, id.UserID}, func(row pgx.Row) (component.Authentication, error) {
		auth, err := authenticationScanner(row)
		if err != nil {
			return component.Authentication{}, err
		}

		return *auth, nil
	}
}

func (s authenticationProcessor) BuildInsertQuery(id core.AuthCompositeKey, authentication component.Authentication) (string, []any) {
	return authInsertQuery, []any{
		id.TeamID,
		id.UserID,
		authentication.TokenType,
		authentication.AccessToken,
		authentication.RefreshToken,
		authentication.ExpiresAt,
		authentication.Scope,
	}
}

func (s authenticationProcessor) BuildUpdateQuery(id core.AuthCompositeKey, authentication component.Authentication) (string, []any) {
	return authUpdateQuery, []any{
		id.TeamID,
		id.UserID,
		authentication.TokenType,
		authentication.AccessToken,
		authentication.RefreshToken,
		authentication.ExpiresAt,
		authentication.Scope,
	}
}

func (s authenticationProcessor) BuildDeleteQuery(id core.AuthCompositeKey) (string, []any) {
	return authDeleteQuery, []any{id.TeamID, id.UserID}
}
