package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScriptFilesLoadsDirsInOrderAndSortsWithinDir(t *testing.T) {
	root := t.TempDir()
	dirA := filepath.Join(root, "a")
	dirB := filepath.Join(root, "b")
	writeScriptFile(t, dirA, "20-b.js")
	writeScriptFile(t, dirA, "10-a.js")
	writeScriptFile(t, dirB, "01-c.js")
	writeDataFile(t, dirB, "ignore.txt")

	files, err := scriptFiles([]string{dirA, dirB})
	if err != nil {
		t.Fatalf("scriptFiles() error = %v", err)
	}
	want := []string{
		filepath.Join(dirA, "10-a.js"),
		filepath.Join(dirA, "20-b.js"),
		filepath.Join(dirB, "01-c.js"),
	}
	if len(files) != len(want) {
		t.Fatalf("files = %#v, want %#v", files, want)
	}
	for i := range want {
		abs, err := filepath.Abs(want[i])
		if err != nil {
			t.Fatal(err)
		}
		if files[i] != filepath.Clean(abs) {
			t.Fatalf("files[%d] = %q, want %q", i, files[i], filepath.Clean(abs))
		}
	}
}

func TestScriptFilesDedupesRepeatedDirs(t *testing.T) {
	dir := t.TempDir()
	writeScriptFile(t, dir, "app.js")
	files, err := scriptFiles([]string{dir, dir})
	if err != nil {
		t.Fatalf("scriptFiles() error = %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected one deduped file, got %#v", files)
	}
}

func TestScriptFilesRejectsMissingDirAndNonDir(t *testing.T) {
	if _, err := scriptFiles([]string{filepath.Join(t.TempDir(), "missing")}); err == nil || !strings.Contains(err.Error(), "stat scripts directory") {
		t.Fatalf("expected missing dir error, got %v", err)
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "file.js")
	writeDataFile(t, dir, "file.js")
	if _, err := scriptFiles([]string{file}); err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("expected non-dir error, got %v", err)
	}
}

func writeScriptFile(t *testing.T, dir, name string) {
	t.Helper()
	writeDataFile(t, dir, name)
}

func writeDataFile(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte("// test\n"), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}
