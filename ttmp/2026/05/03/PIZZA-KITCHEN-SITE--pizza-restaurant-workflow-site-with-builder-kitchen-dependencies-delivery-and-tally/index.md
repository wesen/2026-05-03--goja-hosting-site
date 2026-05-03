---
Title: Pizza restaurant workflow site with builder, kitchen dependencies, delivery, and tally
Ticket: PIZZA-KITCHEN-SITE
Status: active
Topics:
    - goja-site
    - javascript
    - kanban
    - sqlite
    - documentation
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: deploy/sites.local.yaml
      Note: Local multi-site config includes pizza
    - Path: deploy/sites.yaml
      Note: Production multi-site config includes pizza
    - Path: sites/pizza/README.md
      Note: Runnable README for the Pizza Ops example
    - Path: sites/pizza/scripts/00_domain.js
      Note: Pizza menu
    - Path: sites/pizza/scripts/02_store.js
      Note: SQLite model
    - Path: sites/pizza/scripts/03_views.js
      Note: UI DSL rendering and Kanban board builders
ExternalSources: []
Summary: ""
LastUpdated: 2026-05-03T19:11:42.264149304-04:00
WhatFor: ""
WhenToUse: ""
---


# Pizza restaurant workflow site with builder, kitchen dependencies, delivery, and tally

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- goja-site
- javascript
- kanban
- sqlite
- documentation

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
