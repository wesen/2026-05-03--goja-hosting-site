package dbguard

import (
	"context"
	"database/sql"
)

type MeteredDB struct {
	inner *sql.DB
	guard *Guard
}

func NewMeteredDB(inner *sql.DB, guard *Guard) *MeteredDB {
	return &MeteredDB{inner: inner, guard: guard}
}

func (m *MeteredDB) Query(query string, args ...any) (*sql.Rows, error) {
	return m.QueryContext(context.Background(), query, args...)
}

func (m *MeteredDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return m.inner.QueryContext(ctx, query, args...)
}

func (m *MeteredDB) Exec(query string, args ...any) (sql.Result, error) {
	return m.ExecContext(context.Background(), query, args...)
}

func (m *MeteredDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if m.guard != nil {
		if err := m.guard.BeforeExec(query); err != nil {
			return nil, err
		}
	}
	result, err := m.inner.ExecContext(ctx, query, args...)
	if err == nil && m.guard != nil {
		check, checkErr := m.guard.AfterExec(query)
		if checkErr != nil {
			return result, checkErr
		}
		if err := m.guard.ErrorAfterExec(query, check); err != nil {
			return result, err
		}
	}
	return result, err
}
