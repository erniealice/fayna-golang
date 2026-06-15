# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0-alpha] - 2026-06-15

Operations / field-work domain.

### Added
- Jobs (with phases and tasks), job templates, job activities/timesheets (submit/approve/reject), outcome criteria library, task outcomes and summaries, and fulfillment (transition/return) workflows.

### Changed
- `go.mod` now references published tags (`v0.1.0-alpha`) instead of local `replace` directives; local development continues via `go.work`.

[Unreleased]: https://github.com/erniealice/fayna-golang/compare/v0.1.0-alpha...HEAD
[0.1.0-alpha]: https://github.com/erniealice/fayna-golang/releases/tag/v0.1.0-alpha
