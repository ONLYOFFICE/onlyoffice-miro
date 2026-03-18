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
	"crypto/md5"
	"errors"
	"fmt"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	pgx "github.com/jackc/pgx/v5"
	pgxpool "github.com/jackc/pgx/v5/pgxpool"
)

func generateStatementName(query string) string {
	h := md5.Sum([]byte(query))
	return fmt.Sprintf("stmt_%x", h)
}

type postgresStorage[ID comparable, T any] struct {
	pool      *pgxpool.Pool
	processor service.StorageProcessor[ID, T, pgx.Row]
	logger    service.Logger
}

func NewPostgresStorage[ID comparable, T any](
	pool *pgxpool.Pool,
	processor service.StorageProcessor[ID, T, pgx.Row],
	logger service.Logger,
) (service.Storage[ID, T], error) {
	if pool == nil {
		return nil, ErrNilPool
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	logger.Debug(ctx, "Testing database connection")
	if err := pool.Ping(ctx); err != nil {
		logger.Error(ctx, "Failed to connect to database", service.Fields{"error": err.Error()})
		return nil, err
	}

	logger.Debug(ctx, "Database connection successful")
	return &postgresStorage[ID, T]{
		pool:      pool,
		processor: processor,
		logger:    logger,
	}, nil
}

func (s *postgresStorage[ID, T]) Find(ctx context.Context, id ID) (T, error) {
	var result T

	s.logger.Debug(ctx, "Finding record by ID", service.Fields{"id": id})

	query, args, scanner := s.processor.BuildSelectQuery(id)
	opts := pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadOnly,
	}

	sName := generateStatementName(query)
	err := executeTransactionally(ctx, s.pool, opts, sName, query, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, sName, args...)
		var err error
		result, err = scanner(row)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				s.logger.Debug(ctx, "No record found", service.Fields{"id": id})
				return ErrNoRowsAffected
			}

			s.logger.Error(ctx, "Error scanning record", service.Fields{
				"id":    id,
				"error": err.Error(),
			})
			return err
		}

		s.logger.Debug(ctx, "Record found successfully", service.Fields{"id": id})
		return nil
	})

	if err != nil && !errors.Is(err, ErrNoRowsAffected) {
		s.logger.Error(ctx, "Error finding record", service.Fields{
			"id":    id,
			"error": err.Error(),
		})
	}

	return result, err
}

func (s *postgresStorage[ID, T]) Insert(ctx context.Context, id ID, value T) (T, error) {
	s.logger.Debug(ctx, "Inserting new record", service.Fields{"id": id})

	query, args := s.processor.BuildInsertQuery(id, value)
	opts := pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	}

	sName := generateStatementName(query)
	err := executeTransactionally(ctx, s.pool, opts, sName, query, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, sName, args...)
		if err != nil {
			s.logger.Error(ctx, "Error inserting record", service.Fields{
				"id":    id,
				"error": err.Error(),
			})
			return err
		}

		s.logger.Debug(ctx, "Record inserted successfully", service.Fields{"id": id})
		return nil
	})

	return value, err
}

func (s *postgresStorage[ID, T]) Update(ctx context.Context, id ID, value T) (T, error) {
	s.logger.Debug(ctx, "Updating record", service.Fields{"id": id})

	query, args := s.processor.BuildUpdateQuery(id, value)
	opts := pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	}

	sName := generateStatementName(query)
	err := executeTransactionally(ctx, s.pool, opts, sName, query, func(tx pgx.Tx) error {
		res, err := tx.Exec(ctx, sName, args...)
		if err != nil {
			s.logger.Error(ctx, "Error executing update query", service.Fields{
				"id":    id,
				"error": err.Error(),
			})
			return err
		}

		if res.RowsAffected() == 0 {
			s.logger.Debug(ctx, "No record found to update", service.Fields{"id": id})
			return ErrNoRowsAffected
		}

		s.logger.Debug(ctx, "Record updated successfully", service.Fields{
			"id":            id,
			"rows_affected": res.RowsAffected(),
		})
		return nil
	})

	if err != nil && !errors.Is(err, ErrNoRowsAffected) {
		s.logger.Error(ctx, "Error updating record", service.Fields{
			"id":    id,
			"error": err.Error(),
		})
	}

	return value, err
}

func (s *postgresStorage[ID, T]) Delete(ctx context.Context, id ID) error {
	s.logger.Debug(ctx, "Deleting record", service.Fields{"id": id})

	query, args := s.processor.BuildDeleteQuery(id)
	opts := pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	}

	sName := generateStatementName(query)
	err := executeTransactionally(ctx, s.pool, opts, sName, query, func(tx pgx.Tx) error {
		res, err := tx.Exec(ctx, sName, args...)
		if err != nil {
			s.logger.Error(ctx, "Error executing delete query", service.Fields{
				"id":    id,
				"error": err.Error(),
			})
			return err
		}

		if res.RowsAffected() == 0 {
			s.logger.Debug(ctx, "No record found to delete", service.Fields{"id": id})
			return ErrNoRowsAffected
		}

		s.logger.Debug(ctx, "Record deleted successfully", service.Fields{
			"id":            id,
			"rows_affected": res.RowsAffected(),
		})
		return nil
	})

	if err != nil && !errors.Is(err, ErrNoRowsAffected) {
		s.logger.Error(ctx, "Error deleting record", service.Fields{
			"id":    id,
			"error": err.Error(),
		})
	}

	return err
}
