---
Title: Help Docs Implementation Diary
Slug: help-docs-implementation-diary
Short: Chronological notes for adding built-in Glazed help entries to goja-site.
Topics:
- documentation
- glazed
- goja-site
DocType: reference
Ticket: GOJA-SITE-HELP-DOCS
Status: active
Intent: long-term
Created: 2026-05-03
Updated: 2026-05-03
---

# Help Docs Implementation Diary

## Step 1: Created ticket and read authoring guidance

Created ticket `GOJA-SITE-HELP-DOCS` for the built-in help work. Read the local skills for Glazed help authoring, textbook-style technical writing, and docmgr workflow. Also captured the canonical Glazed help guidance with:

```bash
glaze help how-to-write-good-documentation-pages
glaze help writing-help-entries
```

The important constraints are: use exact Glazed frontmatter keys, keep slugs unique, avoid a top-level Markdown heading inside help content, include troubleshooting and see-also sections, and wire the root Cobra command through `help_cmd.SetupCobraRootCommand(...)`.

## Step 2: Mapped docs to source APIs

Inspected the existing CLI and module code to ground the docs in real APIs:

- `cmd/goja-site/main.go`, `serve.go`, `serve_multi.go`
- `pkg/web/express_module.go`
- `pkg/web/request_response.go`
- `pkg/uidsl/module.go`
- `pkg/kanbanddsl/builder.go`
- `pkg/dbguard/registrar.go`
- `pkg/app/multi_config.go`
- `examples/kanban/README.md`

Then wrote `analysis/01-help-entry-content-plan-and-source-map.md`, which defines the four-entry help set and the source/API map each entry should use.

## Step 3: Added embedded help pages

Added a new `pkg/doc` package with embedded Markdown help files and the standard `AddDocToHelpSystem(...)` function. Added four initial help entries:

- `getting-started`
- `user-guide`
- `js-api-reference`
- `developer-guide`

The entries are intentionally written in textbook style: they begin with mental models, then move to concrete code, then finish with troubleshooting and cross-references.

## Step 4: Wired Glazed help into the CLI

Updated `cmd/goja-site/main.go` to create a Glazed `HelpSystem`, load `pkg/doc`, and call `help_cmd.SetupCobraRootCommand(helpSystem, root)`. This adds the rich `goja-site help ...` command tree while preserving the existing Glazed/Cobra `serve` and `serve-multi` commands.

This introduced new transitive `go.sum` entries for the Glazed help renderer and TUI packages. Running the initial test exposed missing sums for packages such as `glamour`, `frontmatter`, `bubbletea`, and `lipgloss`; `go get ...` plus `go mod tidy` fixed the module metadata.

## Step 5: Validation

Validated with:

```bash
go test ./... -count=1
go run ./cmd/goja-site help getting-started
go run ./cmd/goja-site help user-guide
go run ./cmd/goja-site help js-api-reference
go run ./cmd/goja-site help developer-guide
```

All package tests passed, and every new help slug rendered through the CLI.
