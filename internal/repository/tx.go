package repository

import (
	"context"
	"database/sql"

	"github.com/fazriegi/netbase-be/internal/domain"
	"github.com/jmoiron/sqlx"
)

type txKey struct{}

type sqlxTxManager struct {
	db *sqlx.DB
}

func NewTransactionManager(db *sqlx.DB) domain.TransactionManager {
	return &sqlxTxManager{db: db}
}

func (m *sqlxTxManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txCtx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(txCtx); err != nil {
		return err
	}

	return tx.Commit()
}

// sqlxQueryer defines methods implemented by both *sqlx.DB and *sqlx.Tx
type sqlxQueryer interface {
	sqlx.ExtContext
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// Queryer represents the methods common to *sqlx.DB and *sqlx.Tx with transaction support
type Queryer interface {
	sqlx.ExtContext
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type queryerWrapper struct {
	sqlxQueryer
}

func (w queryerWrapper) NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error) {
	return sqlx.NamedQueryContext(ctx, w.sqlxQueryer, query, arg)
}

// getQueryer returns the transaction if present in the context, otherwise it returns the base db connection.
func getQueryer(ctx context.Context, db *sqlx.DB) Queryer {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return queryerWrapper{tx}
	}
	return queryerWrapper{db}
}
