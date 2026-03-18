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
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/core"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/core/component"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	pgx "github.com/jackc/pgx/v5"
)

const (
	settingsSelectQuery = `SELECT s.address, s.header, s.secret, s.demo_detached,
	d.enabled, d.started
	FROM settings s
	LEFT JOIN demos d ON s.team_id = d.team_id
	WHERE s.team_id = $1 AND s.board_id = $2;`

	settingsUpdateQuery = `UPDATE settings
SET address = $3,
    header = $4,
    secret = $5,
    demo_detached = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE team_id = $1 AND board_id = $2;`

	settingsDeleteQuery = `DELETE FROM settings
WHERE team_id = $1 AND board_id = $2;`
)

type settingsProcessor struct{}

func settingsScanner(row pgx.Row) (*component.Settings, error) {
	result := &component.Settings{}
	var enabled *bool
	var started *time.Time
	var demoDetached bool

	if err := row.Scan(
		&result.Address,
		&result.Header,
		&result.Secret,
		&demoDetached,
		&enabled,
		&started,
	); err != nil {
		return nil, err
	}

	result.DemoDetached = demoDetached
	if !demoDetached && enabled != nil && started != nil {
		result.Demo = component.Demo{
			Enabled: *enabled,
			Started: started,
		}
	}

	return result, nil
}

func NewSettingsProcessor() service.StorageProcessor[core.SettingsCompositeKey, component.Settings, pgx.Row] {
	return &settingsProcessor{}
}

func (s settingsProcessor) TableName() string {
	return "settings"
}

func (s settingsProcessor) BuildSelectQuery(id core.SettingsCompositeKey) (string, []any, func(pgx.Row) (component.Settings, error)) {
	return settingsSelectQuery, []any{id.TeamID, id.BoardID}, func(row pgx.Row) (component.Settings, error) {
		settings, err := settingsScanner(row)
		if err != nil {
			return component.Settings{}, err
		}

		if settings.Demo != (component.Demo{}) {
			settings.Demo.TeamID = id.TeamID
		}

		return *settings, nil
	}
}

func (s settingsProcessor) BuildInsertQuery(id core.SettingsCompositeKey, settings component.Settings) (string, []any) {
	if settings.Demo.Enabled {
		var started *time.Time
		if !settings.Demo.Started.IsZero() {
			started = settings.Demo.Started
		}

		return `
            WITH settings_update AS (
                INSERT INTO settings (team_id, board_id, address, header, secret, demo_detached)
                VALUES ($1, $2, $3, $4, $5, $6)
                ON CONFLICT (team_id, board_id) DO UPDATE
                SET address = EXCLUDED.address,
                    header = EXCLUDED.header,
                    secret = EXCLUDED.secret,
                    demo_detached = EXCLUDED.demo_detached,
                    updated_at = CURRENT_TIMESTAMP
                RETURNING team_id
            )
            INSERT INTO demos (team_id, enabled, started)
            VALUES ($1, $7, $8)
            ON CONFLICT (team_id) DO NOTHING
            RETURNING team_id
        `, []any{
				id.TeamID,
				id.BoardID,
				settings.Address,
				settings.Header,
				settings.Secret,
				settings.DemoDetached,
				settings.Demo.Enabled,
				started,
			}
	}

	return `
        INSERT INTO settings (team_id, board_id, address, header, secret, demo_detached)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (team_id, board_id) DO UPDATE
        SET address = EXCLUDED.address,
            header = EXCLUDED.header,
            secret = EXCLUDED.secret,
            demo_detached = EXCLUDED.demo_detached,
            updated_at = CURRENT_TIMESTAMP
        RETURNING team_id
    `, []any{
			id.TeamID,
			id.BoardID,
			settings.Address,
			settings.Header,
			settings.Secret,
			settings.DemoDetached,
		}
}

func (s settingsProcessor) BuildUpdateQuery(id core.SettingsCompositeKey, settings component.Settings) (string, []any) {
	return settingsUpdateQuery, []any{
		id.TeamID,
		id.BoardID,
		settings.Address,
		settings.Header,
		settings.Secret,
		settings.DemoDetached,
	}
}

func (s settingsProcessor) BuildDeleteQuery(id core.SettingsCompositeKey) (string, []any) {
	return settingsDeleteQuery, []any{id.TeamID, id.BoardID}
}
