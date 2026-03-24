# Fayna Templates — Progress Log

**Plan:** [plan.md](./plan.md)
**Started:** 2026-03-19
**Branch:** `dev/20260319-fayna-templates`

---

## Phase 1: Foundation — Labels + Embed Files — COMPLETE

- [x] Add tab empty state labels to `JobLabels` in `labels.go` (PhasesTitle, PhasesMessage, ActivitiesTitle, ActivitiesMessage, SettlementTitle, SettlementMessage, OutcomesTitle, OutcomesMessage)
- [x] Add `Approval` and `SectionInfo` to `JobDetailLabels`
- [x] Add `Outcomes` to `JobTabLabels`
- [x] Add `Form` sub-struct to `JobLabels` (NamePlaceholder, ClientPlaceholder, LocationPlaceholder)
- [x] Add `Columns` and `Empty` sub-structs to `OutcomeSummaryLabels`
- [x] Create `views/job_activity/embed.go`
- [x] Create `views/job_activity/templates/` directory
- [x] Create `views/outcome_criteria/embed.go`
- [x] Create `views/outcome_criteria/templates/` directory
- [x] Create `views/task_outcome/embed.go`
- [x] Create `views/task_outcome/templates/` directory
- [x] Create `views/outcome_summary/embed.go`
- [x] Create `views/outcome_summary/templates/` directory

---

## Phase 2: Fix Existing Templates — COMPLETE

- [x] `job/templates/detail.html` — Replace `sale-*` CSS classes with `job-*` (6 occurrences)
- [x] `job/templates/detail.html` — Replace "Job Information" with `{{.Labels.Detail.SectionInfo}}`
- [x] `job/templates/detail.html` — Replace "Approval" with `{{.Labels.Detail.Approval}}`
- [x] `job/templates/detail.html` — Replace all empty state hardcoded strings with labels
- [x] `job/templates/detail.html` — Add `job-tab-outcomes` for Layer 7
- [x] `job/templates/detail.html` — Add `aria-label` to info section
- [x] `job/templates/drawer-form.html` — Replace placeholder strings with labels
- [x] `job_template/templates/detail.html` — Normalize to info-grid pattern with `jt-*` prefix
- [x] `job_template/templates/detail.html` — Add phases, attachments tabs
- [x] `job_template/templates/detail.html` — Add sheet-form + sheet.js
- [x] `job_template/templates/list.html` — Add sheet-form for drawer support

---

## Phase 3: New Templates — Job Activity — COMPLETE

- [x] Create `job_activity/templates/list.html` (dual-render + table-card)
- [x] Create `job_activity/templates/detail.html` (info grid with `act-*` classes, type sections)
- [x] Create `job_activity/templates/drawer-form.html` (dynamic per entry type)

---

## Phase 4: New Templates — Outcome Criteria — COMPLETE

- [x] Create `outcome_criteria/templates/list.html` (dual-render + table-card)
- [x] Create `outcome_criteria/templates/detail.html` (5 tabs: info, thresholds, options, templates, versions)
- [x] Create `outcome_criteria/templates/drawer-form.html` (criteria create/edit)

---

## Phase 5: New Templates — Task Outcome — COMPLETE

- [x] Create `task_outcome/templates/list.html` (dual-render + table-card)
- [x] Create `task_outcome/templates/detail.html` (determination badges, type-aware value)
- [x] Create `task_outcome/templates/recording-form.html` (6 evaluator form variants)

---

## Phase 6: New Templates — Outcome Summary — COMPLETE

- [x] Create `outcome_summary/templates/list.html` (dual-render + table-card)
- [x] Create `outcome_summary/templates/job-summary.html` (score bar, determination, breakdown)
- [x] Create `outcome_summary/templates/phase-summary.html` (phase-scoped report)

---

## Phase 7: CSS Files — COMPLETE

- [x] Create `apps/service-admin/assets/css/fayna/fayna-job-detail.css`
- [x] Create `apps/service-admin/assets/css/fayna/fayna-job-template-detail.css`
- [x] Create `apps/service-admin/assets/css/fayna/fayna-activity-detail.css`
- [x] Create `apps/service-admin/assets/css/fayna/fayna-criteria-detail.css`
- [x] Create `apps/service-admin/assets/css/fayna/fayna-outcome-detail.css`
- [x] Create `apps/service-admin/assets/css/fayna/fayna-summary.css`

---

## Summary

- **Phases complete:** 7 / 7
- **Files modified:** 5 (labels.go, job detail.html, job drawer-form.html, job_template detail.html, job_template list.html)
- **Files created:** 22 (4 embed.go + 12 new templates + 6 CSS files)

---

## Codex Review

Codex cross-review requested after plan creation. Review notes to be incorporated post-implementation.

| Review | Status | Notes File |
|--------|--------|-----------|
| Plan review | IN PROGRESS | `codex-plan-review.md` (pending) |
| Post-implementation review | NOT STARTED | — |

---

## Skipped / Deferred (update as you work)

| Item | Reason |
|------|--------|
| `JobActivityLabels.Form` sub-struct | No Form sub-struct exists; drawer-form uses Detail/Columns labels. Add in future label audit pass. |
| job_template drawer-form.html | Not in original 5 existing templates; job_template module needs action Go code first |

---

## How to Resume

All 7 phases are COMPLETE. Remaining work:
1. Wait for Codex plan review to complete → read `codex-plan-review.md`
2. Address any findings from Codex review
3. Verify `go build` passes with all build tags
4. Run E2E tests if applicable
5. Commit

Key references:
- **Existing Go views**: `packages/fayna-golang/views/*/list/page.go` (check ContentTemplate names)
- **Labels struct**: `packages/fayna-golang/labels.go` (all 6 modules defined)
- **Centymo CSS pattern**: `apps/service-admin/assets/css/centymo/centymo-sales-detail.css`
- **Pyeza base**: `apps/service-admin/assets/css/pyeza/detail-layout.css`
- **Template pattern**: `packages/centymo-golang/views/inventory/templates/detail.html`
