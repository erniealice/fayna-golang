# fayna-golang

Operational work execution domain package for Ichizen OS. Owns **two esqyma domains -- `operation` (13 entity packages) and `fulfillment` (1 entity package)** -- making it the largest multi-domain vertical slice in the monorepo.

**Module path:** `github.com/erniealice/fayna-golang`

## Domain ownership

fayna maps to two esqyma proto domains (`proto/v1/domain/operation/` and `proto/v1/domain/fulfillment/`). The operation domain covers the full job lifecycle -- templates (design-time), execution (runtime), activities (cost capture), and outcome assessment (quality). The fulfillment domain covers delivery tracking. Together they answer "what work was done?" and "was it any good?"

### Universal job model

Service, project, maintenance, and production jobs are all unified under one schema. A job is the root aggregate; phases and tasks decompose it. Templates define reusable job blueprints. Activities capture cost (labor, material, expense). Outcome criteria define pass/fail targets; task outcomes record actuals; summaries roll up to phase and job level.

## Package structure (Option B)

Under Option B the ENTITY is the contract package. Each `domain/<d>/<e>/` directory is one esqyma entity. The domain facade (`domain/<d>/<d>.go`) re-exports entity-local types as Go type aliases so consumers never change their import paths.

```
fayna-golang/
  placement_test.go            # B-STRICT placement gate -- the ONLY test at root
  routes_config_test.go        # route contract tests (all fields non-empty, RouteMap coverage)
  assets.go                    # embed.FS for pyeza.CopyNamespacedAssets (legacyAllow)
  go.mod / go.sum
  domain/
    operation/                 # package operation -- facade for the operation domain
      operation.go             # facade: type JobLabels = job.Labels, etc. (aliases only)
      routes.go                # route type aliases + default URL constants
      job_module.go            # NewModule() assembler for the job entity
      job_activity_module.go   # NewJobActivityModule() assembler
      job_phase_module.go      # NewJobPhaseModule() assembler
      job_task_module.go       # NewJobTaskModule() assembler
      job_template_module.go   # NewJobTemplateModule() assembler
      job_template_phase_module.go
      job_template_task_module.go
      activity_labor_module.go
      activity_material_module.go
      activity_expense_module.go
      outcome_criteria_module.go
      task_outcome_module.go
      outcome_summary_module.go
      job/                     # entity: esqyma operation/job
        descriptor.go          # Describe() -> compose.Unit (Nav: "Operations")
        labels.go              # Labels struct (JobLabels in the facade)
        routes.go              # Routes struct + DefaultRoutes()
        embed.go               # template embed.FS
        list/page.go           # list page handler
        detail/                # detail page + tabs
        action/                # add / delete handlers
        form/                  # form.Data + option builders
        dashboard/page.go      # operations dashboard handler
        templates/             # HTML templates
      job_activity/            # entity: esqyma operation/job_activity
        descriptor.go, labels.go, routes.go, embed.go
        action/ detail/ form/ list/ templates/
      job_phase/               # entity: esqyma operation/job_phase
      job_task/                # entity: esqyma operation/job_task
      job_template/            # entity: esqyma operation/job_template
        (same shape as job, plus dashboard/ for template listing)
      job_template_phase/      # entity: esqyma operation/job_template_phase
        (no list/ or detail/ -- inline sub-entity)
      job_template_task/       # entity: esqyma operation/job_template_task
        (no list/ or detail/ -- inline sub-entity)
      activity_labor/          # entity: esqyma operation/activity_labor
      activity_material/       # entity: esqyma operation/activity_material
      activity_expense/        # entity: esqyma operation/activity_expense
      outcome_criteria/        # entity: esqyma operation/outcome_criteria
        (includes deps.go + actions.go for inline management)
      task_outcome/            # entity: esqyma operation/task_outcome
        (includes deps.go + actions.go for outcome recording)
      outcome_summary/         # aggregate view: job_outcome_summary + phase_outcome_summary (legacyAllow)
        labels.go              # Labels struct for the combined summary view
        routes.go, embed.go, descriptor.go
        job_summary/           # job-level outcome summary partial
        phase_summary/         # phase-level outcome summary partial
        list/                  # list page handler
        templates/             # HTML templates

    fulfillment/               # package fulfillment -- facade for the fulfillment domain
      fulfillment.go           # facade: type FulfillmentLabels = fulfillment.Labels, etc.
      fulfillment_module.go    # module assembler
      fulfillment/             # entity: esqyma fulfillment/fulfillment
        descriptor.go          # Describe() -> compose.Unit (Nav: "Fulfillment")
        labels.go              # Labels struct (10 sub-structs: Status, DeliveryMode, Column, Tab, Action, Button, Empty, Error, Dashboard)
        routes.go              # Routes struct + DefaultRoutes()
        embed.go               # template embed.FS
        action/ detail/ form/ list/ dashboard/ templates/

  block/
    block.go                   # Block() constructor -- pyeza.AppOption entry point (1,023 LoC)
    catalog.go                 # per-entity compose.Unit binders (Describe -> Mount wiring)
    usecases.go                # *UseCases typed wiring contract + RequireFor + MustValidate
    wiring.go                  # assigns *UseCases closures onto view ModuleDeps
    infra.go                   # Infra struct (attachment ops, ref checker, DB, cross-package URLs)
    block_test.go              # MustValidate fail-closed wiring tests

  assets/css|js/               # placeholder asset dirs (embed.FS via assets.go)
  tests/e2e/                   # Playwright E2E specs (job lifecycle, templates, fulfillment)
```

## Placement gate (`placement_test.go`)

fayna carries a **B-STRICT** placement gate (v2, Option B). `legacyAllow` holds three dated residuals; the target state is empty (STRICT).

| Rule | What it checks |
|------|----------------|
| **R1** Empty root | No package `.go` files at module root -- only `_test.go` permitted |
| **R2** Canonical dirs | Every first-level dir is an allowed infra surface; every `domain/<d>` is an esqyma proto domain |
| **R2'** Entity dirs | Every `domain/<d>/<child>/` DIR is an esqyma entity of domain `<d>`, `shared`, or a domain-view (name starts with `<d>`) |
| **R3'** Entity contract | No real `*Labels`/`*Routes` type declaration at the domain root -- only alias re-exports (`type X = pkg.Y`) are allowed |
| **R4** No god-files | No `.go` file (excl. `_test.go`) may exceed 1,200 lines |
| **R5** Facade exists | A facade `domain/<d>/<d>.go` must exist for every domain dir with >=1 entity subdir |
| **R6** No cycles | Enforced by `lint-no-domain-cycles.sh` (external, go-list based) |

`crossCutting = false` -- the domain variant applies. esqyma's `proto/v1/domain/` is located at test time so the rules never drift from the live proto tree.

Current `legacyAllow` residuals (all EXPIRES 2026-07-15):
- `assets.go` -- embed FS still imported by service-admin container (`fayna.AssetsFS`); pending move to pyeza-owned asset host
- `docs` -- planning markdown from the A-to-B restructure; not a Go concern
- `domain/operation/outcome_summary` -- aggregate of `job_outcome_summary` + `phase_outcome_summary`, not a 1:1 esqyma entity; pending rename to a domain-view or split per esqyma entity

## Labels

2,420 LoC across 14 `labels.go` files (133 structs: 123 operation + 10 fulfillment), split per entity. Each entity's `labels.go` declares a `Labels` struct and a `DefaultLabels()` constructor with English defaults. The facade re-exports each as a named alias (`type JobLabels = job.Labels`).

## Block assembler (`block/`)

`block.go` (1,023 LoC -- under the 1,200-line god-file threshold) is the composition root. It stays monolithic because the largest `wantX` block is under 40 LoC and the control flow is linear.

`catalog.go` bridges compose-v2: each `XxxUnit(uc, infra)` function calls the entity's `Describe()` to get a `compose.Unit`, then sets `Unit.Mount` to a closure that wires deps, routes, labels, and handlers. The block assembler calls these unit binders to build the full module set.

`infra.go` declares the `Infra` struct -- the subset of AppContext that view modules need beyond the typed UseCases: attachment operations, reference checker, DB handle for search endpoints, and cross-package URL patterns (e.g. `SubscriptionDetailURL`).

## Fail-closed wiring (`block/usecases.go`)

`*UseCases` is the typed wiring contract between service-admin's composition layer and fayna's view modules. It declares three groups: `OperationUseCases`, `FulfillmentUseCases`, and cross-domain reads (`SubscriptionUseCases`, `EntityUseCases`). `RequireFor(cfg)` lists every missing REQUIRED closure for the enabled modules. `MustValidate(cfg)` adds fail-closed posture:

- **dev/test** (`testing.Testing()` true or `FAYNA_BLOCK_STRICT` truthy): PANIC with the full field list -- uncatchable-by-accident, stack-traced, fails CI loudly.
- **prod**: `log.Printf("FATAL: ...")` at the seam AND returns the error -> `Block()` propagates -> `NewServiceAdmin` halts boot.

OPTIONAL closures (nested-entity lists, derived picker closures, dashboard aggregates) are never flagged -- they degrade gracefully to empty-state.

## Granular module selection

`Block()` with no options enables all 14 modules. `BlockOption` functions (`WithJob()`, `WithFulfillment()`, `WithJobTemplate()`, etc.) allow consumers to mount only the modules they need:

```go
app.Apply(faynablock.Block())                                      // all modules
app.Apply(faynablock.Block(faynablock.WithJob(), faynablock.WithFulfillment())) // selective
```

## Dependencies

- `github.com/erniealice/pyeza-golang` -- UI framework (view system, compose engine, template engine, types)
- `github.com/erniealice/esqyma` -- proto schemas (operation + fulfillment domains)
- `github.com/erniealice/lyngua` -- translation/i18n
- `github.com/erniealice/espyna-golang` -- typed use cases (via reference checker, consumer container)
- `github.com/erniealice/hybra-golang` -- cross-cutting views (attachment, audit trail)

fayna imports only downward (framework + cross-cutting). It never imports peer domain packages (centymo, entydad, fycha, cyta).

## Role in the monorepo

fayna sits in the domain layer above pyeza and espyna. Consumer apps (e.g., `apps/service-admin`) call `block.Block()` to mount the operations module, supplying a `*UseCases` via `block.WithUseCases(...)`. The typed contract ensures any drift between espyna and fayna is a compile error, not a silent nil.

See `docs/wiki/articles/vertical-slices.md` for the full entity trace and `docs/wiki/articles/package-map.md` for the monorepo dependency graph.
