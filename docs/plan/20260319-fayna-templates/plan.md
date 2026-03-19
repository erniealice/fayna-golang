# Fayna Templates — Design Plan

**Date:** 2026-03-19
**Branch:** `dev/20260319-fayna-templates`
**Status:** Draft
**Package:** fayna-golang-ryta

---

## Overview

Create all missing HTML templates for the 6 fayna view modules (job, job_template, job_activity, outcome_criteria, task_outcome, outcome_summary), fix existing templates that have CSS prefix violations and hardcoded strings, and create the corresponding CSS files. All Go view code is already fully implemented — only the template + CSS + label layers remain.

---

## Motivation

The fayna package has complete Go scaffolding (routes, labels, modules, PageData views) but 4 of 6 modules have **zero HTML templates**. The 2 modules that have templates (job, job_template) contain centymo CSS class leaks (`sale-*` prefixes) and hardcoded English strings that violate the lyngua audit. Without templates, the views return content template names that pyeza can't resolve — rendering "Page content not available" for all operations pages.

---

## Architecture

### Template Dual-Render Pattern (pyeza convention)

Every page needs two `{{define}}` blocks:

```html
{{define "page-name"}}         {{/* Full page — wraps app-shell */}}
{{template "app-shell" .}}
{{end}}

{{define "page-name-content"}}  {{/* HTMX partial — content only */}}
<div class="page-content ...">
    ...
</div>
{{end}}
```

### CSS Class Prefix Convention

| Module | CSS Prefix | Example |
|--------|-----------|---------|
| job | `job-` | `job-detail-layout`, `job-info-grid` |
| job_template | `jt-` | `jt-detail-layout`, `jt-info-grid` |
| job_activity | `act-` | `act-detail-layout`, `act-info-grid` |
| outcome_criteria | `oc-` | `oc-detail-layout`, `oc-info-grid` |
| task_outcome | `to-` | `to-detail-layout`, `to-info-grid` |
| outcome_summary | `os-` | `os-summary-layout`, `os-score-bar` |

### Label Access Pattern

- Domain labels: `{{.Labels.Detail.FieldName}}`, `{{.Labels.Columns.Name}}`
- Empty states: `{{.Labels.Empty.Title}}`, `{{.Labels.Empty.Message}}`
- Common labels: `{{.CommonLabels.Buttons.Save}}`
- **Zero hardcoded English strings** — all text through labels

### PageData Fields Available (from Go views)

**List views**: `.Table` (TableConfig) — use `{{template "table-card" .Table}}`

**Job detail**: `.Job` (map), `.Labels`, `.ActiveTab`, `.TabItems`, `.PhasesTable`, `.ActivitiesTable`, `.SettlementTable`

**Outcome criteria detail**: `.Criteria` (map), `.Labels`, `.ActiveTab`, `.TabItems`

**Task outcome detail**: `.Outcome` (map), `.Labels`

**Job outcome summary**: `.Summary` (map), `.Labels`

---

## Implementation Steps

### Phase 1: Foundation — Labels + Embed Files

Add missing label fields to `labels.go` and create `embed.go` files for modules that lack them.

- Add tab empty state labels to `JobLabels`: `Empty.PhasesTitle`, `Empty.PhasesMessage`, `Empty.ActivitiesTitle`, `Empty.ActivitiesMessage`, `Empty.SettlementTitle`, `Empty.SettlementMessage`
- Add "Approval" to `JobDetailLabels` (currently hardcoded in template)
- Add `Empty` sub-struct to `OutcomeSummaryLabels` (currently missing)
- Create `views/job_activity/embed.go` with `//go:embed templates/*.html` + `TemplatesFS`
- Create `views/job_activity/templates/` directory
- Create `views/outcome_criteria/embed.go` with `//go:embed templates/*.html` + `TemplatesFS`
- Create `views/outcome_criteria/templates/` directory
- Create `views/task_outcome/embed.go` with `//go:embed templates/*.html` + `TemplatesFS`
- Create `views/task_outcome/templates/` directory
- Create `views/outcome_summary/embed.go` with `//go:embed templates/*.html` + `TemplatesFS`
- Create `views/outcome_summary/templates/` directory

### Phase 2: Fix Existing Templates (Job + Job Template)

Fix CSS prefix violations, replace hardcoded strings, add ARIA, add outcomes tab.

**job/templates/detail.html:**
- Replace all `sale-*` CSS classes with `job-*` prefix (6 occurrences)
- Replace "Job Information" → `{{.Labels.Tabs.Info}}` or `{{.Labels.Detail.SectionTitle}}`
- Replace "Approval" → `{{.Labels.Detail.Approval}}`
- Replace "No phases" / "This job has no phases defined yet." → `{{.Labels.Empty.PhasesTitle}}` / `{{.Labels.Empty.PhasesMessage}}`
- Replace "No activities" / message → `{{.Labels.Empty.ActivitiesTitle}}` / `{{.Labels.Empty.ActivitiesMessage}}`
- Replace "No settlements" / message → `{{.Labels.Empty.SettlementTitle}}` / `{{.Labels.Empty.SettlementMessage}}`
- Add `{{else if eq .ActiveTab "outcomes"}}{{template "job-tab-outcomes" .}}` for Layer 7 tab
- Add `aria-label` to info sections

**job/templates/drawer-form.html:**
- Replace placeholder "Job name" → use Labels.Form field
- Replace placeholder "Client ID" → use Labels.Form field
- Replace placeholder "Location ID" → use Labels.Form field

**job_template/templates/detail.html:**
- Normalize to match centymo detail pattern (info-grid instead of dt/dd)
- Add phases tab content
- Add criteria tab content (Layer 7 `template_task_criteria` junction)
- Add attachments tab
- Add sheet-form + sheet.js for add/edit actions
- Replace any hardcoded strings with label references

**job_template/templates/list.html:**
- Add `{{template "sheet-form" .}}` and sheet.js for drawer form support

### Phase 3: New Templates — Job Activity (Layer 4)

**job_activity/templates/list.html:**
- Dual-render: `job-activity-list` + `job-activity-list-content`
- Content: `{{template "table-card" .Table}}`
- Sheet form + sheet.js for add action
- Columns (from Go): date, job, entry_type, description, quantity, amount, status

**job_activity/templates/detail.html:**
- Dual-render: `job-activity-detail` + `job-activity-detail-content`
- Info grid with `act-*` prefix classes
- Fields: date, job, task, entry type, description, quantity, unit cost, amount, billable, approval status
- Type-specific section (labor: hours/rate, material: qty/product, expense: category/receipt)

**job_activity/templates/drawer-form.html:**
- Dynamic form based on entry type (labor/material/expense)
- Common fields: job (select), task (select), description, billable toggle
- Labor fields: hours, hourly_rate
- Material fields: product, quantity, unit_cost
- Expense fields: amount, category, receipt upload

### Phase 4: New Templates — Outcome Criteria (Layer 7)

**outcome_criteria/templates/list.html:**
- Dual-render: `outcome-criteria-list` + `outcome-criteria-list-content`
- Content: `{{template "table-card" .Table}}`
- Sheet form + sheet.js
- Columns (from Go): name, type (badge), scope, version, status

**outcome_criteria/templates/detail.html:**
- Dual-render: `outcome-criteria-detail` + `outcome-criteria-detail-content`
- Tabs: Info, Thresholds, Options, Templates, Versions (from Go)
- Info tab: name, type, scope, version, status, required, weight, description
- Thresholds tab: table of min/max/unit/role rows
- Options tab: table of label/value/weight rows
- Templates tab: table of linked template tasks
- Versions tab: version history list
- CSS: `oc-*` prefix classes

**outcome_criteria/templates/drawer-form.html:**
- Create/edit criterion form
- Fields: name, type (select), scope (select), description, required toggle, weight

### Phase 5: New Templates — Task Outcome (Layer 7)

**task_outcome/templates/list.html:**
- Dual-render: `task-outcome-list` + `task-outcome-list-content`
- Content: `{{template "table-card" .Table}}`
- Columns (from Go): task, criteria, value, determination, recorded_by, date

**task_outcome/templates/detail.html:**
- Dual-render: `task-outcome-detail` + `task-outcome-detail-content`
- Info grid with `to-*` prefix classes
- Fields: task, criteria ref, value (type-aware display), determination badge, recorded_by, notes, timestamp
- Determination badge colors: PASS=success, FAIL=danger, CONDITIONAL=warning, PENDING=info, NOT_EVALUATED=muted

**task_outcome/templates/recording-form.html:**
- Dynamic form that adapts per criteria_type
- NUMERIC_RANGE: number input + unit label, auto-determination
- NUMERIC_SCORE: number input (0-100), auto-determination
- PASS_FAIL: toggle switch
- CATEGORICAL: select dropdown from criteria options
- TEXT: textarea + manual determination select
- MULTI_CHECK: checkbox list from criteria options + pass_rule display
- Common fields: notes textarea

### Phase 6: New Templates — Outcome Summary (Layer 7)

**outcome_summary/templates/list.html:**
- Dual-render: `outcome-summary-list` + `outcome-summary-list-content`
- Content: `{{template "table-card" .Table}}`
- Columns (from Go): job, determination, score, scoring_method, total, pass, fail, issued_by

**outcome_summary/templates/job-summary.html:**
- Dual-render: `job-outcome-summary` + `job-outcome-summary-content`
- Overall determination badge (large, prominent)
- Score display with visual progress bar (`os-score-bar`)
- Criteria breakdown: total, passed, failed, conditional, pending (stat cards or summary grid)
- Narrative section (rich text block)
- Metadata: issued_by, valid_from, valid_until, timestamps
- CSS: `os-*` prefix classes

**outcome_summary/templates/phase-summary.html:**
- Dual-render: `phase-outcome-summary` + `phase-outcome-summary-content`
- Same structure as job summary but scoped to one phase
- Phase name + job reference in header

### Phase 7: CSS Files

Create CSS files in `apps/service-admin/assets/css/fayna/`:

**fayna-job-detail.css:**
- `job-detail-layout`, `job-detail-tabs`, `job-detail-body`
- `job-info-grid`, `job-info-item`, `job-info-label`, `job-info-value`
- `job-section-title`, `job-section-spacer`
- All values via design tokens
- Responsive breakpoints at 768px and 480px

**fayna-job-template-detail.css:**
- `jt-detail-layout`, `jt-info-grid`, `jt-info-item`
- Follows same pattern as job detail

**fayna-activity-detail.css:**
- `act-detail-layout`, `act-info-grid`
- Type-specific sections (labor, material, expense visual distinction)

**fayna-criteria-detail.css:**
- `oc-detail-layout`, `oc-info-grid`
- Threshold/option table styling

**fayna-outcome-detail.css:**
- `to-detail-layout`, `to-info-grid`
- Determination badge prominence styling

**fayna-summary.css:**
- `os-summary-layout`, `os-score-bar`, `os-criteria-breakdown`
- Score progress bar styling
- Determination badge (large variant)
- Phase accordion styling

---

## File References

| File | Change | Phase |
|------|--------|-------|
| `packages/fayna-golang-ryta/labels.go` | Add missing Empty labels for job tabs, Approval label, OutcomeSummary Empty struct | 1 |
| `packages/fayna-golang-ryta/views/job_activity/embed.go` | **New file** — template embedding | 1 |
| `packages/fayna-golang-ryta/views/outcome_criteria/embed.go` | **New file** — template embedding | 1 |
| `packages/fayna-golang-ryta/views/task_outcome/embed.go` | **New file** — template embedding | 1 |
| `packages/fayna-golang-ryta/views/outcome_summary/embed.go` | **New file** — template embedding | 1 |
| `packages/fayna-golang-ryta/views/job/templates/detail.html` | Fix `sale-*` → `job-*` CSS, replace hardcoded strings, add outcomes tab | 2 |
| `packages/fayna-golang-ryta/views/job/templates/drawer-form.html` | Replace hardcoded placeholders with labels | 2 |
| `packages/fayna-golang-ryta/views/job_template/templates/detail.html` | Normalize to info-grid pattern, add tabs, add sheet | 2 |
| `packages/fayna-golang-ryta/views/job_template/templates/list.html` | Add sheet-form for drawer support | 2 |
| `packages/fayna-golang-ryta/views/job_activity/templates/list.html` | **New file** — activity list page | 3 |
| `packages/fayna-golang-ryta/views/job_activity/templates/detail.html` | **New file** — activity detail with type sections | 3 |
| `packages/fayna-golang-ryta/views/job_activity/templates/drawer-form.html` | **New file** — dynamic form per entry type | 3 |
| `packages/fayna-golang-ryta/views/outcome_criteria/templates/list.html` | **New file** — criteria library list | 4 |
| `packages/fayna-golang-ryta/views/outcome_criteria/templates/detail.html` | **New file** — criteria detail with 5 tabs | 4 |
| `packages/fayna-golang-ryta/views/outcome_criteria/templates/drawer-form.html` | **New file** — criteria create/edit form | 4 |
| `packages/fayna-golang-ryta/views/task_outcome/templates/list.html` | **New file** — outcomes list | 5 |
| `packages/fayna-golang-ryta/views/task_outcome/templates/detail.html` | **New file** — outcome detail with determination | 5 |
| `packages/fayna-golang-ryta/views/task_outcome/templates/recording-form.html` | **New file** — dynamic recording form per criteria type | 5 |
| `packages/fayna-golang-ryta/views/outcome_summary/templates/list.html` | **New file** — report cards list | 6 |
| `packages/fayna-golang-ryta/views/outcome_summary/templates/job-summary.html` | **New file** — job report card with score bar | 6 |
| `packages/fayna-golang-ryta/views/outcome_summary/templates/phase-summary.html` | **New file** — phase report | 6 |
| `apps/service-admin/assets/css/fayna/fayna-job-detail.css` | **New file** — job detail styling | 7 |
| `apps/service-admin/assets/css/fayna/fayna-job-template-detail.css` | **New file** — job template detail styling | 7 |
| `apps/service-admin/assets/css/fayna/fayna-activity-detail.css` | **New file** — activity detail styling | 7 |
| `apps/service-admin/assets/css/fayna/fayna-criteria-detail.css` | **New file** — criteria detail styling | 7 |
| `apps/service-admin/assets/css/fayna/fayna-outcome-detail.css` | **New file** — task outcome detail styling | 7 |
| `apps/service-admin/assets/css/fayna/fayna-summary.css` | **New file** — summary/report card styling | 7 |

---

## Context & Sub-Agent Strategy

**Estimated files to read:** ~25 (existing templates, centymo references, labels, Go views)
**Estimated files to modify/create:** 27
**Estimated context usage:** Medium (30-60 files total)

**Sub-agent plan:**
- Phase 1 (labels + embed): Single session, straightforward
- Phase 2 (fix existing): Single session, needs careful label mapping
- Phases 3-6 (new templates): **4 parallel agents** — each module is independent
  - Agent A: job_activity templates (Phase 3)
  - Agent B: outcome_criteria templates (Phase 4)
  - Agent C: task_outcome templates (Phase 5)
  - Agent D: outcome_summary templates (Phase 6)
- Phase 7 (CSS): **Can run in parallel** with template creation, or after
- **Codex cross-review** after plan creation — feedback incorporated post-implementation

---

## Risk & Dependencies

| Risk | Impact | Mitigation |
|------|--------|------------|
| Go views expect specific template names | Templates won't render | Match ContentTemplate strings exactly from Go page.go files |
| Missing label fields in labels.go | Template panics on `{{.Labels.X.Y}}` | Phase 1 adds all needed fields before templates reference them |
| Tab content template names must match | Detail tabs show blank | Cross-reference job detail.go `loadTabData` with template `{{define}}` names |
| Outcome recording form complexity | Dynamic form may need JS | Start with server-rendered form per type, add JS later if needed |

**Dependencies:**
- Phase 1 must complete before Phase 2 (labels needed for template string replacement)
- Phases 3-6 are fully independent (can run in parallel)
- Phase 7 is independent (CSS files don't block template rendering)

---

## Acceptance Criteria

- [ ] All 6 modules have `embed.go` + `templates/` directory
- [ ] 17 new HTML template files created
- [ ] 5 existing templates fixed (no hardcoded strings, no `sale-*` classes)
- [ ] Zero hardcoded English text in any template (all via `{{.Labels.*}}`)
- [ ] All CSS classes use module-specific prefixes (`job-`, `jt-`, `act-`, `oc-`, `to-`, `os-`)
- [ ] All CSS values use design tokens (`var(--*)`) — no hardcoded colors/spacing/radius
- [ ] All detail pages have `role="tabpanel"` + `aria-labelledby` on tab content
- [ ] All forms use pyeza `form-group` template (inherits ARIA attributes)
- [ ] All empty states use label-driven title + message
- [ ] 6 CSS files created in `apps/service-admin/assets/css/fayna/`
- [ ] `go build` passes (templates compile with embed directives)
- [ ] Template names match Go ContentTemplate strings exactly

---

## Design Decisions

1. **Single drawer-form per module** (not per-subtype) — job_activity uses one form with conditional sections shown/hidden via entry type select, rather than 3 separate forms. Simpler to maintain, matches centymo pattern.

2. **Outcome recording form is server-rendered** — each criteria_type gets a different form section rendered server-side based on the criteria's type. No client-side JS switching needed initially (the Go view already knows the type).

3. **CSS in service-admin only** — fayna CSS lives in `apps/service-admin/assets/css/fayna/` since that's the primary consumer. Retail-admin gets it when fayna is deployed there.

4. **Phase summary reuses job summary structure** — same template layout, just scoped to phase. Avoids creating a third summary pattern.
