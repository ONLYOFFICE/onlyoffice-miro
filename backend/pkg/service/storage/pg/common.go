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
	"time"

	pgx "github.com/jackc/pgx/v5"
	pgxpool "github.com/jackc/pgx/v5/pgxpool"
)

func executeTransactionally(
	ctx context.Context,
	pool *pgxpool.Pool,
	opts pgx.TxOptions,
	statementName, statementQuery string,
	fn func(tx pgx.Tx) error,
) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}

	defer conn.Release()

	tx, err := conn.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	if _, err = tx.Prepare(ctx, statementName, statementQuery); err != nil {
		tctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		return tx.Rollback(tctx)
	}

	if err = fn(tx); err != nil {
		tctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		return tx.Rollback(tctx)
	}

	return tx.Commit(ctx)
}
