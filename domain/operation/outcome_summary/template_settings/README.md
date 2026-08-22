# Outcome Summary Template Settings

This package implements the management surface for outcome-summary document
templates. It lists schedule-scoped bindings and handles upload, replacement,
publish, and delete actions through injected application services.

## Key files

- `page.go` owns the list view, drawer form, permission gates, multipart upload
  validation, binding lifecycle, storage calls, and compensation behavior.
- `page_test.go` pins permission-family alignment, DOCX archive hardening, upload
  ordering, cleanup, replacement, publish, and delete contracts.

Uploaded artifacts are constrained to bounded, structurally valid DOCX archives.
The code works with the canonical `job_outcome_summary_document_template` binding;
vertical names such as "Report Card Templates" come from Lyngua labels.

Fayna owns the HTTP/view orchestration here. Espyna owns the permission-gated use
cases and persistence, while the host application supplies its selected storage
provider and container configuration.
