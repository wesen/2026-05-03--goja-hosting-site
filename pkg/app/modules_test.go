package app

import (
	"context"
	"path/filepath"
	"testing"
)

func TestServerRegistersYAMLModule(t *testing.T) {
	root := t.TempDir()
	scripts := writeSiteScript(t, root, `
		const express = require("express");
		const yaml = require("yaml");
		const app = express.app();
		app.get("/", (req, res) => {
		  const parsed = yaml.parse("name: yaml-module");
		  res.type("text/plain").send(parsed.name);
		});
	`)

	srv, err := NewServer(Config{DBPath: filepath.Join(root, "app.db"), ScriptDirs: []string{scripts}, DBPolicy: DBPolicySimple, ReadOnly: true})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	defer func() { _ = srv.Close(context.Background()) }()

	if got := getServerBody(t, srv, "/"); got != "yaml-module" {
		t.Fatalf("YAML route body = %q, want yaml-module", got)
	}
}
