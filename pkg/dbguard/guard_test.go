package dbguard_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/goja-site/pkg/dbguard"
	_ "github.com/mattn/go-sqlite3"
)

func TestStatsIncludesDatabaseFile(t *testing.T) {
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
	stats, err := guard.Stats()
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if stats.FileBytes <= 0 || stats.TotalBytes <= 0 {
		t.Fatalf("expected positive file sizes: %+v", stats)
	}
	if stats.PageSize <= 0 || stats.PageCount <= 0 {
		t.Fatalf("expected pragma page stats: %+v", stats)
	}
}

func TestMeteredExecTriggersManualLimitStateWithoutCallback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	guard := dbguard.New(db, path)
	guard.Configure(dbguard.Options{MaxBytes: 1, CheckEveryWrites: 1, IncludeWAL: true})
	metered := dbguard.NewMeteredDB(db, guard)
	if _, err := metered.Exec("CREATE TABLE cards(id INTEGER PRIMARY KEY, title TEXT)"); err != nil {
		t.Fatalf("exec: %v", err)
	}
	last := guard.LastResult()
	if !last.Triggered {
		t.Fatalf("expected triggered over-limit result: %+v", last)
	}
	if last.SkippedReason != "no callback registered" {
		t.Fatalf("expected no callback skip, got %+v", last)
	}
}
