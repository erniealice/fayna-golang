# fayna-golang

**Operational work execution domain for the Ichizen OS monorepo.**

`fayna-golang` owns the job lifecycle — templates, execution, activities, cost capture, and outcome assessment. It's the domain package for everything that answers "what work was done?" and "was it any good?"

## Etymology

**fayna** — from Filipino *faena* (Spanish origin), meaning labor, work, or a task that demands effort. Grittier than "operation" — it evokes the production floor, the job site, the hands-on work. In bullfighting, *faena* is the final act — the decisive performance. In Ichizen OS, fayna is where the work gets done and measured.

## Architecture

```
  Domain Packages          centymo (commerce), entydad (identity), fycha (accounting), fayna (operations)
        │                  Business logic, entity-specific views, route wiring
        ▼
  Shared Features          hybra (cross-cutting: attachments, comments, audit)
        │                  Application patterns used by all domains
        ▼
  Framework Layer          pyeza (UI), espyna (backend), esqyma (proto schemas)
                           Presentation primitives, infrastructure, data contracts
```

### What fayna owns

| Layer | Concern | Entities |
|-------|---------|----------|
| **Templates** (design-time) | Job blueprints | job_template, job_template_phase, job_template_task |
| **Execution** (runtime) | Active work | job, job_phase, job_task |
| **Activities** (cost capture) | Time & materials | job_activity, activity_labor, activity_material, activity_expense |
| **Settlement** (allocation) | Cost distribution | job_settlement, inventory_movement |
| **Outcomes** (Layer 7) | Quality assessment | outcome_criteria, criteria_threshold, criteria_option, template_task_criteria, task_outcome, task_outcome_check, phase_outcome_summary, job_outcome_summary |

### What does NOT live in fayna

| Concern | Where it belongs | Why |
|---------|-----------------|-----|
| Sales/revenue/invoicing | centymo | Commerce domain — billing the work |
| Client/user/role management | entydad | Identity domain — who did the work |
| Financial reporting | fycha | Accounting domain — what it cost |
| Attachment handling | hybra | Cross-cutting — shared by all domains |
| Proto schemas | esqyma (`domain/operation/`) | Data contracts — schema layer |
| Postgres adapters | espyna | Infrastructure — backend framework |

## View Modules

### Layers 2-6 (migrated from centymo)

- **`views/job/`** — Job lifecycle: list, detail, status transitions, phases/tasks tabs
- **`views/job_template/`** — Template management: list, detail, phase/task hierarchy
- **`views/job_activity/`** — Activity log: CRUD, approval workflow, labor/material/expense subtypes

### Layer 7 (new)

- **`views/outcome_criteria/`** — Criteria library: versioned definitions, thresholds, options
- **`views/task_outcome/`** — Outcome recording: type-adaptive forms, auto-evaluation
- **`views/outcome_summary/`** — Report cards: phase/job summaries, determination badges

## Dependencies

```
fayna-golang
  ├── pyeza-golang/view      (View interface, ViewFunc, ViewResult)
  ├── pyeza-golang/types      (TableConfig, TableColumn, TableRow)
  ├── pyeza-golang/route      (URL resolution helpers)
  ├── hybra-golang/views/attachment  (generic attachment handler)
  └── esqyma                  (operation protobuf types)
```

fayna depends **only** on framework packages and hybra (downward). It never imports peer domain packages (centymo, entydad, fycha).

## Module

```
module github.com/erniealice/fayna-golang
```
