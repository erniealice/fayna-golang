# Fayna E2E Test Bugs

Discovered during initial E2E test run on 2026-03-19 against service-admin (localhost:8081, professional business type).

---

## BUG-FAY-001: Matter SOPs list returns HTTP 500 (SQL error)

**Severity:** P0 — page completely broken
**Route:** `/app/matter-sops/list/active` (professional override for `/app/job-templates/list/active`)
**Error:** `failed to load job templates: failed to query job template list page data: pq: missing FROM-clause entry for table "jt"`
**Root cause:** The postgres adapter for `GetJobTemplateListPageData` has a broken SQL query. The query references table alias `"jt"` in a clause but the alias is not defined in the FROM clause.
**Impact:** The entire Matter SOPs (Job Templates) section is inaccessible. All 6 tests for FAY-TPL-001 skip.
**Fix needed:** Fix the SQL query in the job template postgres adapter — ensure the `"jt"` alias is defined in the FROM or JOIN clause.

---

## BUG-FAY-002: Timesheet list returns HTTP 500 (Internal Server Error)

**Severity:** P0 — page completely broken
**Route:** `/app/timesheet/list` (professional override for `/app/activities`)
**Error:** `Internal Server Error` (no detailed error message in response body)
**Root cause:** The job activity list handler crashes. Likely causes: (1) missing or unimplemented DB adapter for `GetJobActivityListPageData`, (2) nil pointer in view code, or (3) missing template registration. This view was recently scaffolded and may not be fully wired.
**Impact:** The entire Timesheet (Activities) section is inaccessible. All 5 tests for FAY-ACT-001 skip.
**Fix needed:** Check server logs for the full stack trace. Wire the `GetJobActivityListPageData` use case and ensure the postgres adapter query is implemented.

---

## BUG-FAY-003: Matters table is empty (no seed data)

**Severity:** P2 — tests limited, not a code bug
**Route:** `/app/matters/list/active`
**Observation:** The matters list page loads correctly but shows "No jobs found" with 0 entries. This means:
- Edit/action button tests cannot run (FAY-JOB-003 skips both tests)
- CRUD create tests would need to create data first
**Impact:** 2 tests in FAY-JOB-003 skip due to empty table.
**Fix needed:** Add seed data for jobs in the service1 database, or write a create-first test that seeds a matter before testing edit.

---

## Summary

| Test Suite | Tests | Passed | Skipped | Failed |
|-----------|-------|--------|---------|--------|
| FAY-JOB-001: Matters List | 5 | 5 | 0 | 0 |
| FAY-JOB-002: Matter Add | 3 | 3 | 0 | 0 |
| FAY-JOB-003: Matter Edit | 2 | 0 | 2 | 0 |
| FAY-TPL-001: Matter SOPs | 6 | 0 | 6 | 0 |
| FAY-ACT-001: Timesheet | 5 | 0 | 5 | 0 |
| **Total** | **21** | **8** | **13** | **0** |

### Skip reasons breakdown

| Reason | Count |
|--------|-------|
| HTTP 500: Matter SOPs SQL error (BUG-FAY-001) | 6 |
| HTTP 500: Timesheet handler crash (BUG-FAY-002) | 5 |
| Empty table: no matter seed data (BUG-FAY-003) | 2 |
