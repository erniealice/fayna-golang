# fayna-golang

**Operational work execution domain for the Ichizen OS monorepo.**

`fayna-golang` owns the job lifecycle — templates, execution, activities, cost
capture, and outcome assessment. It answers "what work was done?" and "was it
any good?"

## Etymology

**fayna** — from Filipino *faena* (Spanish origin), meaning labor or work that
demands effort. In Ichizen OS, fayna is where work gets done and measured.

## Package layout (Option B — entity-first)

```
fayna-golang/
  assets.go                    embed.FS for CSS/JS (pyeza.CopyNamespacedAssets)
  assets/css|js/               placeholder asset dirs

  domain/
    operation/                 package operation — domain facade
      operation.go             facade: re-exports entity labels/routes as type aliases
      routes.go                aggregated route constructors
      <entity>_module.go       per-entity module assembler (block deps → Module)
      job/                     package job
      job_phase/               package job_phase
      job_task/                package job_task
      job_template/            package job_template
      job_template_phase/      package job_template_phase
      job_template_task/       package job_template_task
      job_activity/            package job_activity
      activity_labor/          package activity_labor
      activity_material/       package activity_material
      activity_expense/        package activity_expense
      outcome_criteria/        package outcome_criteria
      task_outcome/            package task_outcome
      outcome_summary/         aggregate view (job+phase summaries) — legacyAllow
        job_summary/
        phase_summary/
        list/

    fulfillment/               package fulfillment — domain facade
      fulfillment.go           facade: re-exports FulfillmentLabels/Routes aliases
      fulfillment_module.go    module assembler
      fulfillment/             package fulfillment (entity package)

  block/
    block.go                   Block() entry point; MustValidate fail-closed gate
    usecases.go                UseCases struct; RequireFor + MustValidate
    wiring.go                  per-entity wireXxxModule helpers

  tests/e2e/                   Playwright E2E specs (job lifecycle, templates, etc.)
```

### Entity package shape

Every entity dir (`domain/<d>/<e>/`) follows the same layout:

```
<entity>/
  embed.go        embed.FS (TemplatesFS)
  labels.go       Labels struct + DefaultLabels()
  routes.go       Routes struct + DefaultRoutes() + RouteMap()
  list/           list view + action handler
  detail/         detail view
  action/         mutation handlers (create/update/delete)
  form/           form.Data/FormData + option builders
  templates/      (job_activity only) extra template packages
  dashboard/      (job, fulfillment only) dashboard view
```

### Facade (`<d>.go`)

`domain/operation/operation.go` and `domain/fulfillment/fulfillment.go` are
hand-written facades that re-export every entity's `Labels` and `Routes` types
as named aliases (`type JobLabels = job.Labels`) so consumers write
`operation.JobLabels` rather than importing each entity package directly. Build
enforcement: a missing alias is a compile error in the consumer.

### Module assemblers (`<e>_module.go`)

Each `<entity>_module.go` in `domain/<d>/` is a function `wire<Entity>Module`
that wires block deps → `espyna.Module` (routes, labels, typed use cases). The
block assembler (`block/block.go`) calls these functions; the result is passed
to `pyeza.AppOption` as the registered sidebar module.

### Block assembler (`block/`)

`block.go` is the composition root for fayna modules. `UseCases` in
`usecases.go` carries all function-field ports. `MustValidate` (fail-closed
wrapper around `RequireFor`) panics in dev/test when a REQUIRED closure is nil
and logs + returns an error in prod — mirrors the AUTHZ_ENFORCE boot-guard.

## Domains

| Domain | Entities |
|--------|----------|
| `operation` | job, job_phase, job_task, job_template, job_template_phase, job_template_task, job_activity, activity_labor, activity_material, activity_expense, outcome_criteria, task_outcome |
| `fulfillment` | fulfillment |

## Dependencies

fayna imports only downward (framework + hybra). It never imports peer domain
packages (centymo, entydad, fycha).

```
fayna-golang
  ├── pyeza-golang        (view interface, components, asset hosting)
  ├── espyna-golang       (use cases, ports, Module type)
  ├── hybra-golang        (cross-cutting: attachment, audit_trail views)
  └── esqyma              (operation + fulfillment protobuf types)
```

## Module

```
module github.com/erniealice/fayna-golang
```
