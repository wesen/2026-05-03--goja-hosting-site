package dbguard_test

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-go-golems/goja-site/pkg/dbguard"
	_ "github.com/mattn/go-sqlite3"
)

func TestHardLimitRejectsGrowthButAllowsCleanup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec("CREATE TABLE cards(id INTEGER PRIMARY KEY, title TEXT)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	guard := dbguard.New(db, path)
	guard.Configure(dbguard.Options{HardMaxBytes: 1, MaxBytes: 1, FailOverHardLimit: true, CheckEveryWrites: 1, IncludeWAL: true})
	metered := dbguard.NewMeteredDB(db, guard)

	if _, err := metered.Exec("INSERT INTO cards(title) VALUES ('blocked')"); err == nil {
		t.Fatalf("expected hard-limit error for insert")
	} else {
		var hard *dbguard.HardLimitError
		if !errors.As(err, &hard) {
			t.Fatalf("expected HardLimitError, got %T %v", err, err)
		}
	}

	if _, err := metered.Exec("DELETE FROM cards WHERE id < 0"); err != nil {
		t.Fatalf("cleanup delete should be allowed over hard limit: %v", err)
	}
	if _, err := metered.Exec("VACUUM"); err != nil {
		t.Fatalf("maintenance vacuum should be allowed over hard limit: %v", err)
	}
}

func TestSoftLimitDoesNotRejectWhenFailHardDisabled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec("CREATE TABLE cards(id INTEGER PRIMARY KEY, title TEXT)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	guard := dbguard.New(db, path)
	guard.Configure(dbguard.Options{HardMaxBytes: 1, MaxBytes: 1, FailOverHardLimit: false, CheckEveryWrites: 1, IncludeWAL: true})
	metered := dbguard.NewMeteredDB(db, guard)
	if _, err := metered.Exec("INSERT INTO cards(title) VALUES ('allowed')"); err != nil {
		t.Fatalf("insert should be allowed when hard failure disabled: %v", err)
	}
}

func TestCleanupCallbackCanWriteWhileOverHardLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec("CREATE TABLE cards(id INTEGER PRIMARY KEY, title TEXT)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	guard := dbguard.New(db, path)
	// Enable hard failure, but don't preflight before the write by setting hard high
	// initially. The callback then lowers the hard limit and verifies cleanup writes
	// are still allowed while inCleanup is true.
	guard.Configure(dbguard.Options{MaxBytes: 1, HardMaxBytes: 1 << 60, FailOverHardLimit: true, CheckEveryWrites: 1, Cooldown: time.Millisecond, IncludeWAL: true})
	metered := dbguard.NewMeteredDB(db, guard)
	if _, err := metered.Exec("INSERT INTO cards(title) VALUES ('seed')"); err != nil {
		t.Fatalf("seed insert: %v", err)
	}
	guard.Configure(dbguard.Options{MaxBytes: 1, HardMaxBytes: 1, FailOverHardLimit: true, CheckEveryWrites: 1, Cooldown: time.Millisecond, IncludeWAL: true})
	// Directly exercise callback-in-cleanup path through CheckNow would require a goja
	// callback; runtime coverage lives in registrar_test. Here we assert cleanup SQL is
	// allowed by policy over hard limit.
	if err := guard.BeforeExec("DELETE FROM cards WHERE id = 1"); err != nil {
		t.Fatalf("delete preflight should be allowed: %v", err)
	}
}
