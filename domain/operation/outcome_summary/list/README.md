# Outcome Summary List View

This directory implements the list entry point for outcome summaries. It supports
both the generic flat summary table and the app-enabled, tabbed section landing
used by school-admin's report-card routes.

## Key files

- `page.go` defines dependencies, permission checks, schedule tabs, section rows,
  historical reads, and the flat-versus-landing render path.
- `categories.go` builds ordered category columns and the optional uncategorized
  bucket from workspace data.
- `chips.go` converts group-level approval rollups into status chips for composite
  landing cells.

The view consumes Espyna use-case closures and Esqyma messages, then shapes them
into Pyeza tables and tabs. It does not query storage directly.

Application options preserve the generic service-admin table while allowing
school-admin to opt into subscription-group sections, category columns, scoped
visibility, current/past schedule facets, and download actions.
