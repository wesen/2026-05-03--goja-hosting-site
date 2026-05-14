package app

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/go-go-golems/go-go-goja/engine"
	databasemod "github.com/go-go-golems/go-go-goja/modules/database"
	"github.com/go-go-golems/goja-site/pkg/dbguard"
	"github.com/go-go-golems/goja-site/pkg/observability"
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
		if cfg.Observability != nil && cfg.Observability.Guard != nil {
			guard.SetObserver(observability.NewGuardObserver(cfg.SiteName, cfg.Observability.Guard))
		}
		queryExecer = dbguard.NewMeteredDB(db, guard)
		registrars = append(registrars, dbguard.NewRegistrar(guard))
	default:
		return databaseRuntimeConfig{}, fmt.Errorf("unsupported database policy %q", policy)
	}

	if cfg.Observability != nil && cfg.Observability.DB != nil {
		queryExecer = observability.InstrumentQueryExecer(queryExecer, cfg.SiteName, string(policy), cfg.Observability.DB)
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
	tokens := sqlTokens(query)
	if len(tokens) == 0 {
		return false
	}
	switch tokens[0] {
	case "SELECT", "EXPLAIN":
		return !containsSQLWriteToken(tokens)
	case "WITH":
		return !containsSQLWriteToken(tokens)
	case "PRAGMA":
		return isReadOnlyPragma(tokens)
	default:
		return false
	}
}

func containsSQLWriteToken(tokens []string) bool {
	for _, token := range tokens {
		switch token {
		case "INSERT", "UPDATE", "DELETE", "REPLACE", "CREATE", "ALTER", "DROP", "TRUNCATE", "VACUUM", "ANALYZE", "REINDEX", "ATTACH", "DETACH":
			return true
		}
	}
	return false
}

func isReadOnlyPragma(tokens []string) bool {
	if len(tokens) < 2 {
		return false
	}
	for _, token := range tokens {
		if token == "=" {
			return false
		}
	}
	switch tokens[1] {
	case "APPLICATION_ID", "DATABASE_LIST", "FOREIGN_KEY_LIST", "FREELIST_COUNT", "INDEX_INFO", "INDEX_LIST", "INDEX_XINFO", "PAGE_COUNT", "PAGE_SIZE", "QUICK_CHECK", "SCHEMA_VERSION", "TABLE_INFO", "TABLE_LIST", "TABLE_XINFO", "USER_VERSION":
		return true
	default:
		return false
	}
}

func sqlTokens(query string) []string {
	var tokens []string
	for i := 0; i < len(query); {
		ch := query[i]
		switch {
		case isSQLSpace(ch):
			i++
		case ch == '-' && i+1 < len(query) && query[i+1] == '-':
			i += 2
			for i < len(query) && query[i] != '\n' {
				i++
			}
		case ch == '/' && i+1 < len(query) && query[i+1] == '*':
			i += 2
			for i+1 < len(query) && (query[i] != '*' || query[i+1] != '/') {
				i++
			}
			if i+1 < len(query) {
				i += 2
			}
		case ch == '\'' || ch == '"' || ch == '`':
			i = skipSQLQuoted(query, i, ch)
		case ch == '[':
			i++
			for i < len(query) && query[i] != ']' {
				i++
			}
			if i < len(query) {
				i++
			}
		case isSQLTokenByte(ch):
			start := i
			i++
			for i < len(query) && isSQLTokenByte(query[i]) {
				i++
			}
			tokens = append(tokens, strings.ToUpper(query[start:i]))
		default:
			tokens = append(tokens, string(ch))
			i++
		}
	}
	return tokens
}

func skipSQLQuoted(query string, start int, quote byte) int {
	for i := start + 1; i < len(query); i++ {
		if query[i] != quote {
			continue
		}
		if i+1 < len(query) && query[i+1] == quote {
			i++
			continue
		}
		return i + 1
	}
	return len(query)
}

func isSQLSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\f'
}

func isSQLTokenByte(ch byte) bool {
	return ch == '_' || ch >= '0' && ch <= '9' || ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z'
}
