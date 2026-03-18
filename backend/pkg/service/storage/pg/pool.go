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
	"context"
	"runtime"
	"time"

	pgx "github.com/jackc/pgx/v5"
	pgxpool "github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(connection string, opts ...Option) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connection)
	if err != nil {
		return nil, err
	}

	maxConns := runtime.NumCPU() * 4
	if maxConns < 5 {
		maxConns = 5
	}

	options := Options{
		MaxConnections:              maxConns,
		MinConnections:              2,
		ConnectionTimeout:           5 * time.Second,
		ConnectionIdleTimeout:       5 * time.Minute,
		ConnectionHealthcheckPeriod: 15 * time.Second,
		MaxConnLifetime:             1 * time.Hour,
		MaxRetries:                  3,
		RetryInterval:               500 * time.Millisecond,
	}

	for _, opt := range opts {
		opt(&options)
	}

	config.MaxConns = int32(options.MaxConnections)
	config.MinConns = int32(options.MinConnections)
	config.MaxConnIdleTime = options.ConnectionIdleTimeout
	config.HealthCheckPeriod = options.ConnectionHealthcheckPeriod
	config.MaxConnLifetime = options.MaxConnLifetime
	config.ConnConfig.ConnectTimeout = options.ConnectionTimeout
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET statement_timeout = '5s'")
		return err
	}

	config.ConnConfig.RuntimeParams = map[string]string{
		"application_name": "onlyoffice-miro",
	}

	var pool *pgxpool.Pool
	var retryErr error

	for i := 0; i <= options.MaxRetries; i++ {
		if i > 0 {
			time.Sleep(options.RetryInterval)
		}

		pool, retryErr = pgxpool.NewWithConfig(context.Background(), config)
		if retryErr == nil {
			ctx, cancel := context.WithTimeout(context.Background(), options.ConnectionTimeout)
			err = pool.Ping(ctx)
			cancel()

			if err == nil {
				return pool, nil
			}

			pool.Close()
			retryErr = err
		}
	}

	if retryErr != nil {
		return nil, retryErr
	}

	return pool, nil
}
