package dbguard

import (
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
	return m.inner.Query(query, args...)
}

func (m *MeteredDB) Exec(query string, args ...any) (sql.Result, error) {
	result, err := m.inner.Exec(query, args...)
	if err == nil && m.guard != nil {
		_, _ = m.guard.AfterExec(query)
	}
	return result, err
}
