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
	if m.guard != nil {
		if err := m.guard.BeforeExec(query); err != nil {
			return nil, err
		}
	}
	result, err := m.inner.Exec(query, args...)
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
