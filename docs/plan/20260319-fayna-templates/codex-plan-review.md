# Codex Plan Review — Fayna Templates

**Reviewed by:** Codex (gpt-5.4, read-only sandbox)
**Date:** 2026-03-19
**Tokens used:** 568,476

---

## Findings

### 1. Job outcomes tab — Go wiring missing
The template now includes a `job-tab-outcomes` block (`job/templates/detail.html:24`), but `job/detail/page.go` has no `OutcomesTable` field in PageData and doesn't add an `"outcomes"` tab item to `TabItems`. **Go view code change needed.**

### 2. job_template tab-partial name mismatch
`job_template/detail/page.go:224` renders tab partials as `job-tab-*`, but the template defines them as `jt-tab-info` and `jt-tab-phases`. **Either the Go code or the template names need alignment.**

### 3. outcome_criteria detail — PageData doesn't expose tab tables
The template expects `.ThresholdsTable`, `.OptionsTable`, `.TemplatesTable`, `.VersionsTable`, but `PageData` in `outcome_criteria/detail/page.go` doesn't have these fields yet. Templates will show empty state, which is fine for now but **Go tab data loading needed for the tabs to work.**

### 4. Drawer forms lack matching Go routes
- `job_activity/module.go` only registers POST create/update/delete — no GET form view routes
- `outcome_criteria/module.go` only registers list/detail/tab routes — no add/edit routes
- `task_outcome/module.go` only registers list/detail routes — no add/edit routes

**The drawer form templates exist but the Go route/view wiring to serve them isn't in place yet.** Templates will sit unused until routes are added.

### 5. Form label structs missing
`labels.go` has no `Form` sub-struct for `job_activity`, `outcome_criteria`, or `task_outcome`. The drawer forms use `Detail.*` and `Columns.*` labels as workaround. **Form labels should be added in a future label audit pass.**

### 6. outcome_summary list page — hardcoded Go strings
`outcome_summary/list/page.go` has hardcoded strings (title "Report Cards", etc.) that aren't sourced from `OutcomeSummaryLabels`. **Go-side lyngua compliance issue, not a template issue.**

### 7. Summary map key mismatches
Templates reference some `Summary` map keys that the Go views don't currently expose (`job_name`, `phase_name`, `valid_from`). **Templates should use available keys or Go views should be extended.**

### 8. CSS filename vs data-page-css handle mismatch
Templates use `data-page-css` attributes that may not match the CSS filenames exactly. **Minor — verify the asset pipeline resolves handles to files correctly, or rename to match.**

### 9. embed.go pattern — PASS
All 6 embed.go files match the existing pattern exactly. No issues.

### 10. CSS prefix convention — PASS
No centymo class-prefix leakage found in the plan or new templates.

---

## Classification

| Finding | Severity | Action Needed | Blocks Templates? |
|---------|----------|---------------|--------------------|
| 1. Outcomes tab Go wiring | Medium | Add to job/detail/page.go | No — shows empty state |
| 2. jt-tab-* name mismatch | **High** | Align Go or template names | **Yes — tab action breaks** |
| 3. Criteria tab tables | Low | Add to page.go when ready | No — shows coming-soon |
| 4. Drawer form routes | Medium | Add GET routes to modules | No — forms exist, not served |
| 5. Form label structs | Low | Future label audit | No — uses Detail/Columns |
| 6. Go hardcoded strings | Low | Future lyngua audit | No — template is clean |
| 7. Summary map keys | Medium | Verify keys match | Possible runtime error |
| 8. CSS handle naming | Low | Verify asset pipeline | No — CSS loads by file |
