package observability

import "testing"

func TestSQLKindLabel(t *testing.T) {
	tests := map[string]string{
		"SELECT * FROM items":                        "select",
		" insert into items values (1)":              "insert",
		"-- comment\nUPDATE items SET name = ?":      "update",
		"/* comment */ DELETE FROM items":            "delete",
		"PRAGMA table_info(items)":                   "pragma",
		"WITH rows AS (SELECT 1) SELECT * FROM rows": "with",
		"VACUUM": "vacuum",
		"":       "unknown",
		"definitely_not_sql but user controlled value": "other",
	}
	for query, want := range tests {
		if got := SQLKindLabel(query); got != want {
			t.Fatalf("SQLKindLabel(%q) = %q, want %q", query, got, want)
		}
	}
}

func TestErrorClass(t *testing.T) {
	if got := ErrorClass(nil); got != "none" {
		t.Fatalf("nil error class = %q", got)
	}
}
