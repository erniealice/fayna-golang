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

	// Template-grain delivery summary (education tier; List view — 20260710
	// staff-class-list plan, S6). The former ~76-fetch Go aggregation (jobs +
	// seats + group/plan/staff lookups) is now ONE server-side GROUP-BY read;
	// nil-safe (a nil closure renders the education-tier list empty).
	deps.ListJobTemplateSummaries = u.Operation.JobTemplateSummary.ListJobTemplateSummaries

	// "/classes" job_category tab-split (20260714). ListJobCategories supplies
	// the tab rows; ListJobTemplates builds the job_template→job_category map the
	// education-tier filter needs. Both nil-safe → flat list when absent.
	deps.ListJobCategories = u.Operation.JobCategory.ListJobCategories
	deps.ListJobTemplates = u.Operation.JobTemplate.ListJobTemplates

	// "/classes" tabstrip single-statement support read (20260718 courses-list-
	// perf Rank-1). ONE call per page load returns all categories + active
	// template stubs, replacing the 12 generic-List statements above (the two
	// ListJobCategories closures + the paged ListJobTemplates). Nil-safe → the
	// list falls back to no tabs.
	deps.ListJobListTabSupport = u.Operation.JobListTabSupport.ListJobListTabSupport
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

func wireJobCategoryDeps(deps *operation.JobCategoryModuleDeps, u *UseCases) {
	jc := &u.Operation.JobCategory
	deps.CreateJobCategory = jc.CreateJobCategory
	deps.ReadJobCategory = jc.ReadJobCategory
	deps.UpdateJobCategory = jc.UpdateJobCategory
	deps.DeleteJobCategory = jc.DeleteJobCategory
	deps.ListJobCategories = jc.ListJobCategories
}

// ---------------------------------------------------------------------------
// Education grading module wiring (20260616 v1)
// ---------------------------------------------------------------------------

func wireScoringSchemeDeps(deps *operation.ScoringSchemeModuleDeps, u *UseCases) {
	ss := &u.Operation.ScoringScheme
	deps.CreateScoringScheme = ss.CreateScoringScheme
	deps.ReadScoringScheme = ss.ReadScoringScheme
	deps.UpdateScoringScheme = ss.UpdateScoringScheme
	deps.DeleteScoringScheme = ss.DeleteScoringScheme
	deps.ListScoringSchemes = ss.ListScoringSchemes
}

func wireScoringComponentDeps(deps *operation.ScoringComponentModuleDeps, u *UseCases) {
	sc := &u.Operation.ScoringComponent
	deps.CreateScoringComponent = sc.CreateScoringComponent
	deps.ReadScoringComponent = sc.ReadScoringComponent
	deps.UpdateScoringComponent = sc.UpdateScoringComponent
	deps.DeleteScoringComponent = sc.DeleteScoringComponent
	deps.ListScoringComponents = sc.ListScoringComponents
}

func wireTemplateTaskCriteriaDeps(deps *operation.TemplateTaskCriteriaModuleDeps, u *UseCases) {
	ttc := &u.Operation.TemplateTaskCriteria
	deps.CreateTemplateTaskCriteria = ttc.CreateTemplateTaskCriteria
	deps.ReadTemplateTaskCriteria = ttc.ReadTemplateTaskCriteria
	deps.UpdateTemplateTaskCriteria = ttc.UpdateTemplateTaskCriteria
	deps.DeleteTemplateTaskCriteria = ttc.DeleteTemplateTaskCriteria
	deps.ListTemplateTaskCriterias = ttc.ListTemplateTaskCriterias
}

func wireScoringComponentCriteriaDeps(deps *operation.ScoringComponentCriteriaModuleDeps, u *UseCases) {
	scc := &u.Operation.ScoringComponentCriteria
	deps.CreateScoringComponentCriteria = scc.CreateScoringComponentCriteria
	deps.ReadScoringComponentCriteria = scc.ReadScoringComponentCriteria
	deps.UpdateScoringComponentCriteria = scc.UpdateScoringComponentCriteria
	deps.DeleteScoringComponentCriteria = scc.DeleteScoringComponentCriteria
	deps.ListScoringComponentCriterias = scc.ListScoringComponentCriterias
}

func wireScoreScaleDeps(deps *operation.ScoreScaleModuleDeps, u *UseCases) {
	ss := &u.Operation.ScoreScale
	deps.CreateScoreScale = ss.CreateScoreScale
	deps.ReadScoreScale = ss.ReadScoreScale
	deps.UpdateScoreScale = ss.UpdateScoreScale
	deps.DeleteScoreScale = ss.DeleteScoreScale
	deps.ListScoreScales = ss.ListScoreScales
}

func wireScoreScaleBandDeps(deps *operation.ScoreScaleBandModuleDeps, u *UseCases) {
	ssb := &u.Operation.ScoreScaleBand
	deps.CreateScoreScaleBand = ssb.CreateScoreScaleBand
	deps.ReadScoreScaleBand = ssb.ReadScoreScaleBand
	deps.UpdateScoreScaleBand = ssb.UpdateScoreScaleBand
	deps.DeleteScoreScaleBand = ssb.DeleteScoreScaleBand
	deps.ListScoreScaleBands = ssb.ListScoreScaleBands
}

func wireJobOutcomeLineDeps(deps *operation.JobOutcomeLineModuleDeps, u *UseCases) {
	jol := &u.Operation.JobOutcomeLine
	deps.CreateJobOutcomeLine = jol.CreateJobOutcomeLine
	deps.ReadJobOutcomeLine = jol.ReadJobOutcomeLine
	deps.UpdateJobOutcomeLine = jol.UpdateJobOutcomeLine
	deps.DeleteJobOutcomeLine = jol.DeleteJobOutcomeLine
	deps.ListJobOutcomeLines = jol.ListJobOutcomeLines
}

func wireReportingCheckpointDeps(deps *operation.ReportingCheckpointModuleDeps, u *UseCases) {
	rc := &u.Operation.ReportingCheckpoint
	deps.CreateReportingCheckpoint = rc.CreateReportingCheckpoint
	deps.ReadReportingCheckpoint = rc.ReadReportingCheckpoint
	deps.UpdateReportingCheckpoint = rc.UpdateReportingCheckpoint
	deps.DeleteReportingCheckpoint = rc.DeleteReportingCheckpoint
	deps.ListReportingCheckpoints = rc.ListReportingCheckpoints
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
// OutcomeMatrix module wiring
// ---------------------------------------------------------------------------

func wireOutcomeMatrixDeps(deps *operation.OutcomeMatrixModuleDeps, u *UseCases) {
	om := &u.Operation.OutcomeMatrix
	deps.GetOutcomeMatrix = om.GetOutcomeMatrix
	deps.GetOutcomeSummaryRoster = om.GetOutcomeSummaryRoster
	deps.ResolveStaff = om.ResolveStaff

	// Per-phase approval transitions backing the approval bar (plan §4.2). These
	// route through the espyna job_phase transition use cases (full-set authz, D7
	// ownership / admin override, ancestry+workspace, hard-freeze, exact-set
	// compare) — never raw SQL from the view. Nil-safe (a build without them
	// leaves the bar actions unwired / fail-closed).
	jp := &u.Operation.JobPhase
	deps.SubmitJobPhaseApproval = jp.SubmitJobPhaseApproval
	deps.VerifyJobPhaseApproval = jp.VerifyJobPhaseApproval
	deps.PublishJobPhaseApproval = jp.PublishJobPhaseApproval
	deps.ReturnJobPhaseApproval = jp.ReturnJobPhaseApproval

	// Batch-save write path routes through the typed task_outcome CRUD use cases
	// (Read backs the IDOR ownership check; Update/Create apply the values).
	to := &u.Operation.TaskOutcome
	deps.CreateTaskOutcome = to.CreateTaskOutcome
	deps.UpdateTaskOutcome = to.UpdateTaskOutcome
	deps.ReadTaskOutcome = to.ReadTaskOutcome

	// NOTE: deps.ComputePhaseOutcome / deps.ComputeJobOutcome (inline grade
	// recompute, W2) are NOT sourced here — they come from the app AppContext via
	// infra (the GenerateDoc precedent), set on the deps by OutcomeMatrixUnit's
	// Mount. This wiring helper only sources use-case closures off *UseCases.

	// Roster display-name hydration (the same closure the job drawer's client
	// search picker already uses — already wired in engineblock.go, no new
	// espyna surface needed here).
	deps.ListClients = u.Entity.Client.ListClients

	// Page-header delivery-group resolution (round 4 item 2) — all three are
	// already-wired top-level closures reused from elsewhere in the block
	// (the job list's template-grain summary uses the same three).
	deps.ListJobs = u.Operation.Job.ListJobs
	deps.ListSubscriptionGroupMembers = u.Subscription.SubscriptionGroupMember.ListSubscriptionGroupMembers
	deps.ListSubscriptionGroups = u.Subscription.SubscriptionGroup.ListSubscriptionGroups

	// Row-attribute hydration backing deps.Options (set by the unit from the
	// app's EngineBlock option). Nil-safe end to end.
	deps.ListClientAttributes = u.Entity.ClientAttribute.ListClientAttributes
	deps.ResolveAttributeIDByCode = u.Entity.ClientAttribute.ResolveAttributeIDByCode
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

	// Report-cards navigation deps (view-1 landing tabbed section list + view-2
	// per-section grid). All reused from already-wired top-level closures — no
	// new espyna surface. Nil-safe end to end (a nil closure degrades the
	// affected surface to empty/flat, never panics).
	deps.ListPriceSchedules = u.Subscription.PriceSchedule.ListPriceSchedules
	deps.ListSubscriptionGroups = u.Subscription.SubscriptionGroup.ListSubscriptionGroups
	// H2 category filter: resolve Options.CategoryFilter (a job_category code,
	// e.g. "academic") to its id once per request so the three grade surfaces
	// drop same-origin deportment jobs. Reuses the same closure the "/classes"
	// tab-split already consumes. Nil-safe → no filter (service-admin unaffected).
	deps.ListJobCategories = u.Operation.JobCategory.ListJobCategories
	deps.ListSubscriptionGroupMembers = u.Subscription.SubscriptionGroupMember.ListSubscriptionGroupMembers
	deps.ListSubscriptionGroupWorkspaceUsers = u.Subscription.SubscriptionGroupWorkspaceUser.ListSubscriptionGroupWorkspaceUsers
	deps.ListWorkspaceUsers = u.Entity.WorkspaceUser.ListWorkspaceUsers
	deps.ListJobs = u.Operation.Job.ListJobs
	// view-3 per-student card: maps each phase_outcome_summary to its Sem 1 / Sem 2
	// column via job_phase.phase_order.
	deps.ListJobPhases = u.Operation.JobPhase.ListJobPhases
	// Block-layout report-card tree: resolves job_template_phase.code (projected by
	// the specialized ListByJobTemplate SQL) so per-phase leaves key by phase code.
	deps.ListJobTemplatePhasesByTemplate = u.Operation.JobTemplatePhase.ListByJobTemplate
	deps.ListJobTemplates = u.Operation.JobTemplate.ListJobTemplates
	deps.ListClients = u.Entity.Client.ListClients
	deps.ListClientAttributes = u.Entity.ClientAttribute.ListClientAttributes
	deps.ResolveAttributeIDByCode = u.Entity.ClientAttribute.ResolveAttributeIDByCode
	deps.ListJobTemplateSummaries = u.Operation.JobTemplateSummary.ListJobTemplateSummaries
	// Landing dynamic category columns (R9 W-A2): the SAME single-statement
	// tab-support UNION read the "/classes" job list consumes — category
	// headers + the ACTIVE template→category map. Nil-safe → static columns.
	deps.ListJobListTabSupport = u.Operation.JobListTabSupport.ListJobListTabSupport
	// Report-card document download: the per-criterion transcript fetch. The
	// authoritative per-criterion marks live on task_outcome, reached through
	// job_task and A/B/C/D-ordered via template_task_criteria.sequence_order.
	// job_outcome_line is retained for a potential per-subject fallback.
	// Optional/nil-safe — a nil closure leaves the criterion columns blank.
	deps.ListJobOutcomeLines = u.Operation.JobOutcomeLine.ListJobOutcomeLines
	deps.ListJobTasks = u.Operation.JobTask.ListJobTasks
	deps.ListTaskOutcomes = u.Operation.TaskOutcome.ListTaskOutcomes
	// Ownership-joined latest-cell read (phase/task/criterion codes) backing the
	// coded-cell surface of the outcome-summary document. Optional/nil-safe.
	deps.ListCodedTaskOutcomeValuesByJob = u.Operation.TaskOutcome.ListCodedTaskOutcomeValuesByJob
	// Past-AY sibling: admits the inactive historical ancestry so past report
	// cards resolve their attendance/coded cells instead of rendering blank.
	deps.ListCodedTaskOutcomeValuesByJobHistorical = u.Operation.TaskOutcome.ListCodedTaskOutcomeValuesByJobHistorical
	deps.ListTemplateTaskCriterias = u.Operation.TemplateTaskCriteria.ListTemplateTaskCriterias
	// v2 block-layout document enrichments: criterion display names + the
	// User-hydrating staff read (bare ListStaffs never populates Staff.User).
	deps.ListOutcomeCriterias = u.Operation.OutcomeCriteria.ListOutcomeCriterias
	deps.GetStaffListPageData = u.Entity.Staff.GetStaffListPageData

	// TB3 template settings — the report-card template binding lifecycle
	// (list/create/delete/publish). Nil-safe: a nil aggregate leaves the settings
	// surface at "not configured".
	binding := &u.Operation.JobOutcomeSummaryDocumentTemplate
	deps.ListTemplateBindings = binding.ListJobOutcomeSummaryDocumentTemplates
	deps.CreateTemplateBinding = binding.CreateJobOutcomeSummaryDocumentTemplate
	deps.DeleteTemplateBinding = binding.DeleteJobOutcomeSummaryDocumentTemplate
	deps.PublishTemplateBinding = binding.PublishJobOutcomeSummaryDocumentTemplate
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
// Performance-Evaluation module wiring (20260604) — OPTIONAL / nil-able.
//
// Every assignment is nil-safe: a nil closure copies through and the eval view
// modules tolerate it (empty-state render). These groups are NOT in RequireFor,
// so a missing closure never refuses boot. service-admin's buildFaynaUseCases
// (adapters_fayna.go) is the espyna→block adapter that populates them; until it
// lands they stay nil and the eval surfaces render empty-state.
// ---------------------------------------------------------------------------

func wireEvaluationDeps(deps *operation.EvaluationModuleDeps, u *UseCases) {
	e := &u.Operation.Evaluation
	deps.CreateEvaluation = e.CreateEvaluation
	deps.ReadEvaluation = e.ReadEvaluation
	deps.UpdateEvaluation = e.UpdateEvaluation
	deps.DeleteEvaluation = e.DeleteEvaluation
	deps.ListEvaluations = e.ListEvaluations
	deps.GetListPageData = e.GetListPageData
	deps.GetItemPageData = e.GetItemPageData
	deps.GetPortalPageData = e.GetPortalPageData
	deps.SignOffEvaluation = e.SignOffEvaluation
	deps.ArchiveEvaluation = e.ArchiveEvaluation
	deps.ListEvaluationResponses = e.ListEvaluationResponses
	deps.CreateEvaluationResponse = e.CreateEvaluationResponse
	deps.ListEvaluationTemplateItems = e.ListEvaluationTemplateItems
	// deps.NewID is set in the Unit Mount closure from infra.NewAttachmentID
	// (a generic UUIDv7 generator) — NewID is infra, not a use case.
}

func wireEvaluationTemplateDeps(deps *operation.EvaluationTemplateModuleDeps, u *UseCases) {
	t := &u.Operation.EvaluationTemplate
	deps.CreateEvaluationTemplate = t.CreateEvaluationTemplate
	deps.ReadEvaluationTemplate = t.ReadEvaluationTemplate
	deps.UpdateEvaluationTemplate = t.UpdateEvaluationTemplate
	deps.DeleteEvaluationTemplate = t.DeleteEvaluationTemplate
	deps.ListEvaluationTemplates = t.ListEvaluationTemplates
	deps.ListEvaluationTemplateItems = t.ListEvaluationTemplateItems
	deps.ListOutcomeCriterias = t.ListOutcomeCriterias
	// deps.NewID is set in the Unit Mount closure from infra.NewAttachmentID.
}

func wireEvaluationTemplateItemDeps(deps *operation.EvaluationTemplateItemModuleDeps, u *UseCases) {
	i := &u.Operation.EvaluationTemplateItem
	deps.CreateEvaluationTemplateItem = i.CreateEvaluationTemplateItem
	deps.ReadEvaluationTemplateItem = i.ReadEvaluationTemplateItem
	deps.UpdateEvaluationTemplateItem = i.UpdateEvaluationTemplateItem
	deps.DeleteEvaluationTemplateItem = i.DeleteEvaluationTemplateItem
	deps.ListEvaluationTemplateItems = i.ListEvaluationTemplateItems
	deps.ListOutcomeCriterias = i.ListOutcomeCriterias
	// deps.NewID is set in the Unit Mount closure from infra.NewAttachmentID.
}

func wireEvaluationCycleDeps(deps *operation.EvaluationCycleModuleDeps, u *UseCases) {
	c := &u.Operation.EvaluationCycle
	deps.CreateEvaluationCycle = c.CreateEvaluationCycle
	deps.ReadEvaluationCycle = c.ReadEvaluationCycle
	deps.ListEvaluationCycles = c.ListEvaluationCycles
	deps.OpenEvaluationCycle = c.OpenEvaluationCycle
	deps.CloseEvaluationCycle = c.CloseEvaluationCycle
	deps.ListEvaluationCycleMembers = c.ListEvaluationCycleMembers
	deps.ListEvaluations = c.ListEvaluations
	// X-of-Y banner: prefer the view-typed espyna read-UC adapter (Service.
	// Performance.GetCycleProgress); nil → the detail/list compute inline from
	// ListEvaluationCycleMembers + ListEvaluations.
	deps.GetCycleProgress = u.Service.Performance.GetCycleProgress
	// deps.NewID is set in the Unit Mount closure from infra.NewAttachmentID.
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
