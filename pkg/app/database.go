package app

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/go-go-golems/go-go-goja/engine"
	databasemod "github.com/go-go-golems/go-go-goja/modules/database"
	"github.com/go-go-golems/goja-site/pkg/dbguard"
)

type databaseRuntimeConfig struct {
	moduleSpecs []engine.ModuleSpec
	registrars  []engine.RuntimeModuleRegistrar
}

func buildDatabaseRuntimeConfig(cfg Config, db *sql.DB) (databaseRuntimeConfig, error) {
	policy, err := normalizeDBPolicy(cfg.DBPolicy)
	if err != nil {
		return databaseRuntimeConfig{}, err
	}

	var queryExecer databasemod.QueryExecer
	var registrars []engine.RuntimeModuleRegistrar
	switch policy {
	case DBPolicySimple:
		queryExecer = &simpleDB{db: db, allowWrites: cfg.AllowWrites && !cfg.ReadOnly}
	case DBPolicyGuarded:
		guard := dbguard.New(db, cfg.DBPath)
		queryExecer = dbguard.NewMeteredDB(db, guard)
		registrars = append(registrars, dbguard.NewRegistrar(guard))
	default:
		return databaseRuntimeConfig{}, fmt.Errorf("unsupported database policy %q", policy)
	}

	databaseModule := databasemod.New(
		databasemod.WithPreconfiguredDB(queryExecer),
		databasemod.WithConfigureEnabled(false),
	)
	dbAliasModule := databasemod.New(
		databasemod.WithName("db"),
		databasemod.WithPreconfiguredDB(queryExecer),
		databasemod.WithConfigureEnabled(false),
	)

	return databaseRuntimeConfig{
		moduleSpecs: []engine.ModuleSpec{
			engine.NativeModuleSpec{ModuleID: "database:app", ModuleName: databaseModule.Name(), Loader: databaseModule.Loader},
			engine.NativeModuleSpec{ModuleID: "database:db-alias", ModuleName: dbAliasModule.Name(), Loader: dbAliasModule.Loader},
		},
		registrars: registrars,
	}, nil
}

type simpleDB struct {
	db          *sql.DB
	allowWrites bool
}

func (s *simpleDB) Query(query string, args ...any) (*sql.Rows, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("database is not configured")
	}
	if !s.allowWrites && !isReadOnlySQL(query) {
		return nil, writesDisabledError()
	}
	return s.db.Query(query, args...)
}

func (s *simpleDB) Exec(query string, args ...any) (sql.Result, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("database is not configured")
	}
	if !s.allowWrites {
		return nil, writesDisabledError()
	}
	return s.db.Exec(query, args...)
}

func writesDisabledError() error {
	return fmt.Errorf("database writes are disabled; restart with --db-policy simple --allow-writes and without --readonly")
}

func isReadOnlySQL(query string) bool {
	switch firstSQLToken(query) {
	case "SELECT", "WITH", "PRAGMA", "EXPLAIN":
		return true
	default:
		return false
	}
}

func firstSQLToken(query string) string {
	s := strings.TrimSpace(query)
	for {
		switch {
		case strings.HasPrefix(s, "--"):
			idx := strings.IndexByte(s, '\n')
			if idx < 0 {
				return ""
			}
			s = strings.TrimSpace(s[idx+1:])
		case strings.HasPrefix(s, "/*"):
			idx := strings.Index(s, "*/")
			if idx < 0 {
				return ""
			}
			s = strings.TrimSpace(s[idx+2:])
		default:
			if s == "" {
				return ""
			}
			for i, r := range s {
				if !isSQLTokenChar(r) {
					return strings.ToUpper(s[:i])
				}
			}
			return strings.ToUpper(s)
		}
	}
}

func isSQLTokenChar(r rune) bool {
	return r == '_' || r == '-' || r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z'
}
