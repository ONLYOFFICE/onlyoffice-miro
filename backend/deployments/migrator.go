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
package deployments

import (
	"embed"
	"fmt"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/config"
	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	iofs "github.com/golang-migrate/migrate/v4/source/iofs"
	pgxpool "github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Migrator struct {
	migrate *migrate.Migrate
}

func NewMigrator(pool *pgxpool.Pool, config *config.DataSourceConfig) (*Migrator, error) {
	d, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to create migrations source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, config.DatasourceURL())
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return &Migrator{
		migrate: m,
	}, nil
}

func (m *Migrator) Up() error {
	if err := m.migrate.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func (m *Migrator) Down() error {
	if err := m.migrate.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	return nil
}

func (m *Migrator) Close() error {
	if _, err := m.migrate.Close(); err != nil {
		return fmt.Errorf("failed to close migrator: %w", err)
	}

	return nil
}
