package db

import (
	"chill-db/internal/domain"
	"context"
)

type Repository interface {
	ListDatabases(ctx context.Context) ([]string, error)

	CreateDatabase(ctx context.Context, name string) error

	CreateTable(ctx context.Context, dbName string, table domain.TableMetaData) error

	InsertRow(ctx context.Context, dbName, tableName string, row domain.Row) error

	Query(ctx context.Context, dbName, tableName string) ([]domain.Row, error)

	DropDatabase(ctx context.Context, dbName string) error
}
