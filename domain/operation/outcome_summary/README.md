# Outcome Summary Presentation

This package is Fayna's reusable presentation unit for the canonical
`outcome_summary` concept. In the education vertical, Lyngua exposes that concept
as report cards; the Go API and permission names remain generic.

## Key files

- `descriptor.go` declares the composition unit, navigation contributions, label
  bindings, route bindings, and embedded templates.
- `routes.go` defines generic route defaults and the route map that applications
  can override through their Lyngua tier.
- `labels.go`, `options.go`, and `permissions.go` define view text, app-selected
  presentation behavior, and permission helpers.
- `embed.go` publishes the package-owned HTML templates as an embedded filesystem.

Subpackages implement the landing list, section and client views, document
rendering, exports, and template-management actions. Espyna supplies injected use
cases and storage operations; Pyeza supplies the shared UI and view contracts.

Host applications such as `school-admin` compose this unit with education routes,
labels, capabilities, and provider wiring without introducing an education-only
domain model.
