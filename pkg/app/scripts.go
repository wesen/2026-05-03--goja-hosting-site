package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func scriptFiles(dirs []string) ([]string, error) {
	if len(dirs) == 0 {
		return nil, fmt.Errorf("at least one scripts directory is required")
	}
	seen := map[string]struct{}{}
	var files []string
	for _, dir := range dirs {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}
		dirFiles, err := scriptFilesInDir(dir)
		if err != nil {
			return nil, err
		}
		for _, file := range dirFiles {
			abs, err := filepath.Abs(file)
			if err != nil {
				return nil, fmt.Errorf("resolve script path %s: %w", file, err)
			}
			abs = filepath.Clean(abs)
			if _, ok := seen[abs]; ok {
				continue
			}
			seen[abs] = struct{}{}
			files = append(files, abs)
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no JavaScript files found in scripts directories: %s", strings.Join(dirs, ", "))
	}
	return files, nil
}

func scriptFilesInDir(dir string) ([]string, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("stat scripts directory %s: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("scripts path %s is not a directory", dir)
	}
	var files []string
	if err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".js") {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk scripts directory %s: %w", dir, err)
	}
	sort.Strings(files)
	return files, nil
}
