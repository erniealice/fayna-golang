package block

// wiring.go copies the typed *UseCases closures into each view module's
// ModuleDeps.
//
// Phase 2 (20260531-composition-and-dependency-audit, Q-WIRE-1): this replaces
// the prior reflection-based wiring. The fayna/block sub-package does not import
// espyna-golang (that would create a dependency cycle: fayna/block → espyna →
// fayna for domain types). Instead, service-admin's composition layer builds a
// typed *UseCases (see usecases.go) from espyna's consumer container and passes
// it via WithUseCases(); these helpers just assign the already-typed closures.
//
// Every assignment is nil-safe at the struct level: a nil group field copies a
// nil closure, and the view modules tolerate nil closures (empty-state render).
// RequireFor() (usecases.go) is the deterministic completeness gate for the
// REQUIRED closures — drift there is a startup error, not a silent nil.

import (
	fulfillmentdomain "github.com/erniealice/fayna-golang/domain/fulfillment"
	operation "github.com/erniealice/fayna-golang/domain/operation"
)

// ---------------------------------------------------------------------------
// Job module wiring
// ---------------------------------------------------------------------------

func wireJobDeps(deps *operation.JobModuleDeps, u *UseCases) {
	j := &u.Operation.Job
	deps.CreateJob = j.CreateJob
	deps.ReadJob = j.ReadJob
	deps.UpdateJob = j.UpdateJob
	deps.DeleteJob = j.DeleteJob
	deps.ListJobs = j.ListJobs

	// Phases tab — list only (Read/Update wired to the job_phase module).
	deps.ListJobPhases = u.Operation.JobPhase.ListJobPhases

	// Tasks tab.
	deps.ListJobTasks = u.Operation.JobTask.ListJobTasks

	// Activities tab + Actuals tab (cost rollup).
	deps.ListJobActivities = u.Operation.JobActivity.ListJobActivities
	deps.GetJobActivityRollup = u.Operation.JobActivity.GetJobActivityRollup

	// Budget tab — template-derived phase/task hours.
	deps.ReadJobTemplate = u.Operation.JobTemplate.ReadJobTemplate
	deps.ListJobTemplatePhasesByTemplate = u.Operation.JobTemplatePhase.ListByJobTemplate
	deps.ListJobTemplateTasksByPhase = u.Operation.JobTemplateTask.ListByPhase

	// Cross-domain Subscription read for the Job detail origin breadcrumb
	// (2026-04-29 auto-spawn-jobs-from-subscription plan §5.4).
	deps.ReadSubscription = u.Subscription.Subscription.ReadSubscription
}

// ---------------------------------------------------------------------------
// JobTemplate module wiring
// ---------------------------------------------------------------------------

func wireJobTemplateDeps(deps *operation.JobTemplateModuleDeps, u *UseCases) {
	jt := &u.Operation.JobTemplate
	deps.CreateJobTemplate = jt.CreateJobTemplate
	deps.ReadJobTemplate = jt.ReadJobTemplate
	deps.UpdateJobTemplate = jt.UpdateJobTemplate
	deps.DeleteJobTemplate = jt.DeleteJobTemplate
	deps.GetJobTemplateListPageData = jt.GetJobTemplateListPageData

	deps.ListPhasesByJobTemplate = u.Operation.JobTemplatePhase.ListByJobTemplate

	// P6.template-children — wire when present; nil is safe (empty-state panels).
	deps.ListTasksByPhase = u.Operation.JobTemplateTask.ListByPhase
	deps.ListCriteriaByTask = u.Operation.TemplateTaskCriteria.ListByTemplateTask
}

// ---------------------------------------------------------------------------
// JobActivity module wiring
// ---------------------------------------------------------------------------

func wireJobActivityDeps(deps *operation.JobActivityModuleDeps, u *UseCases) {
	ja := &u.Operation.JobActivity
	deps.GetJobActivityListPageData = ja.GetJobActivityListPageData
	deps.ReadJobActivity = ja.ReadJobActivity
	deps.CreateJobActivity = ja.CreateJobActivity
	deps.UpdateJobActivity = ja.UpdateJobActivity
	deps.DeleteJobActivity = ja.DeleteJobActivity
	deps.SubmitForApproval = ja.SubmitForApproval
	deps.ApproveActivity = ja.ApproveActivity
	deps.RejectActivity = ja.RejectActivity
}

// ---------------------------------------------------------------------------
// JobPhase module wiring
// ---------------------------------------------------------------------------

func wireJobPhaseDeps(deps *operation.JobPhaseModuleDeps, u *UseCases) {
	jp := &u.Operation.JobPhase
	deps.CreateJobPhase = jp.CreateJobPhase
	deps.ReadJobPhase = jp.ReadJobPhase
	deps.UpdateJobPhase = jp.UpdateJobPhase
	deps.DeleteJobPhase = jp.DeleteJobPhase
	deps.ListJobPhases = jp.ListJobPhases

	// Activities tab — in-memory filter of all job activities by the phase's job_id.
	deps.ListJobActivities = u.Operation.JobActivity.ListJobActivities

	// Tasks tab — ListJobTasksByPhase from the JobTask use case.
	deps.ListJobTasksByPhase = u.Operation.JobTask.ListJobTasksByPhase
}

// ---------------------------------------------------------------------------
// OutcomeCriteria module wiring
// ---------------------------------------------------------------------------

func wireOutcomeCriteriaDeps(deps *operation.OutcomeCriteriaModuleDeps, u *UseCases) {
	oc := &u.Operation.OutcomeCriteria
	deps.CreateOutcomeCriteria = oc.CreateOutcomeCriteria
	deps.ReadOutcomeCriteria = oc.ReadOutcomeCriteria
	deps.UpdateOutcomeCriteria = oc.UpdateOutcomeCriteria
	deps.DeleteOutcomeCriteria = oc.DeleteOutcomeCriteria
	deps.ListOutcomeCriterias = oc.ListOutcomeCriterias
}

// ---------------------------------------------------------------------------
// TaskOutcome module wiring
// ---------------------------------------------------------------------------

func wireTaskOutcomeDeps(deps *operation.TaskOutcomeModuleDeps, u *UseCases) {
	to := &u.Operation.TaskOutcome
	deps.CreateTaskOutcome = to.CreateTaskOutcome
	deps.ReadTaskOutcome = to.ReadTaskOutcome
	deps.UpdateTaskOutcome = to.UpdateTaskOutcome
	deps.DeleteTaskOutcome = to.DeleteTaskOutcome
	deps.ListTaskOutcomes = to.ListTaskOutcomes

	// ReadOutcomeCriteria cross-dep (from OutcomeCriteria use case).
	deps.ReadOutcomeCriteria = u.Operation.OutcomeCriteria.ReadOutcomeCriteria
}

// ---------------------------------------------------------------------------
// OutcomeSummary module wiring
// ---------------------------------------------------------------------------

func wireOutcomeSummaryDeps(deps *operation.OutcomeSummaryModuleDeps, u *UseCases) {
	jos := &u.Operation.JobOutcomeSummary
	deps.GetJobOutcomeSummaryByJob = jos.GetByJob
	deps.ListJobOutcomeSummarys = jos.ListJobOutcomeSummaries

	pos := &u.Operation.PhaseOutcomeSummary
	deps.GetPhaseOutcomeSummaryByJobPhase = pos.GetByJobPhase
	deps.ListPhaseOutcomeSummarysByJob = pos.ListByJob
}

// ---------------------------------------------------------------------------
// Fulfillment module wiring
// ---------------------------------------------------------------------------

func wireFulfillmentDeps(deps *fulfillmentdomain.FulfillmentModuleDeps, u *UseCases) {
	ff := &u.Fulfillment
	deps.GetFulfillmentListPageData = ff.GetFulfillmentListPageData
	deps.GetFulfillmentItemPageData = ff.GetFulfillmentItemPageData
	deps.CreateFulfillment = ff.CreateFulfillment
	deps.UpdateFulfillment = ff.UpdateFulfillment
	deps.DeleteFulfillment = ff.DeleteFulfillment
	deps.TransitionStatus = ff.TransitionStatus
}

// ---------------------------------------------------------------------------
// Dashboard wiring
// ---------------------------------------------------------------------------
//
// The dashboard func slots return the fayna VIEW types directly (the proto→view
// translation lives in service-admin's adapters.go — Round 2). These helpers
// just copy the typed closure into the module deps. nil → empty-state render.

func wireJobDashboard(deps *operation.JobModuleDeps, u *UseCases) {
	deps.GetJobDashboardPageData = u.Service.Dashboard.Job
}

func wireFulfillmentDashboard(deps *fulfillmentdomain.FulfillmentModuleDeps, u *UseCases) {
	deps.GetFulfillmentDashboardPageData = u.Service.Dashboard.Fulfillment
}

// ---------------------------------------------------------------------------
// ActivityLabor module wiring (OPTIONAL — nil-able until espyna P5)
// ---------------------------------------------------------------------------

func wireActivityLaborDeps(deps *operation.ActivityLaborModuleDeps, u *UseCases) {
	al := &u.Operation.ActivityLabor
	deps.CreateActivityLabor = al.CreateActivityLabor
	deps.ReadActivityLabor = al.ReadActivityLabor
	deps.UpdateActivityLabor = al.UpdateActivityLabor
	deps.DeleteActivityLabor = al.DeleteActivityLabor
	deps.ListActivityLabors = al.ListActivityLabors
}

// ---------------------------------------------------------------------------
// ActivityMaterial module wiring (OPTIONAL — nil-able until espyna P5)
// ---------------------------------------------------------------------------

func wireActivityMaterialDeps(deps *operation.ActivityMaterialModuleDeps, u *UseCases) {
	am := &u.Operation.ActivityMaterial
	deps.CreateActivityMaterial = am.CreateActivityMaterial
	deps.ReadActivityMaterial = am.ReadActivityMaterial
	deps.UpdateActivityMaterial = am.UpdateActivityMaterial
	deps.DeleteActivityMaterial = am.DeleteActivityMaterial
	deps.ListActivityMaterials = am.ListActivityMaterials
}

// ---------------------------------------------------------------------------
// ActivityExpense module wiring (OPTIONAL — nil-able until espyna P5)
// ---------------------------------------------------------------------------

func wireActivityExpenseDeps(deps *operation.ActivityExpenseModuleDeps, u *UseCases) {
	ae := &u.Operation.ActivityExpense
	deps.CreateActivityExpense = ae.CreateActivityExpense
	deps.ReadActivityExpense = ae.ReadActivityExpense
	deps.UpdateActivityExpense = ae.UpdateActivityExpense
	deps.DeleteActivityExpense = ae.DeleteActivityExpense
	deps.ListActivityExpenses = ae.ListActivityExpenses
}

// ---------------------------------------------------------------------------
// JobTemplatePhase module wiring
// ---------------------------------------------------------------------------

func wireJobTemplatePhaseDeps(deps *operation.JobTemplatePhaseModuleDeps, u *UseCases) {
	jtp := &u.Operation.JobTemplatePhase
	deps.CreateJobTemplatePhase = jtp.CreateJobTemplatePhase
	deps.ReadJobTemplatePhase = jtp.ReadJobTemplatePhase
	deps.UpdateJobTemplatePhase = jtp.UpdateJobTemplatePhase
	deps.DeleteJobTemplatePhase = jtp.DeleteJobTemplatePhase
}

// ---------------------------------------------------------------------------
// JobTemplateTask module wiring
// ---------------------------------------------------------------------------

func wireJobTemplateTaskDeps(deps *operation.JobTemplateTaskModuleDeps, u *UseCases) {
	jtt := &u.Operation.JobTemplateTask
	deps.CreateJobTemplateTask = jtt.CreateJobTemplateTask
	deps.ReadJobTemplateTask = jtt.ReadJobTemplateTask
	deps.UpdateJobTemplateTask = jtt.UpdateJobTemplateTask
	deps.DeleteJobTemplateTask = jtt.DeleteJobTemplateTask
}

// ---------------------------------------------------------------------------
// JobTask module wiring
// ---------------------------------------------------------------------------

func wireJobTaskDeps(deps *operation.JobTaskModuleDeps, u *UseCases) {
	jt := &u.Operation.JobTask
	deps.CreateJobTask = jt.CreateJobTask
	deps.ReadJobTask = jt.ReadJobTask
	deps.UpdateJobTask = jt.UpdateJobTask
	deps.DeleteJobTask = jt.DeleteJobTask
	deps.ListJobTasks = jt.ListJobTasks

	// Activities tab — filtered by job_task_id from all job activities.
	deps.ListJobActivities = u.Operation.JobActivity.ListJobActivities
}
