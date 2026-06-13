// Package block — typed wiring contract for fayna.Block.
//
// This file declares what fayna's Block() needs from outside. Service-admin's
// composition layer constructs a *UseCases value from espyna's consumer
// container; fayna's Block() consumes only this typed shape.
//
// Shape this struct by what FAYNA needs, NOT by mirroring espyna's
// *consumer.UseCases. Service-admin's adapter is the only place that knows
// both vocabularies. If espyna restructures its container, only that adapter
// changes.
//
// Phase 2 (20260531-composition-and-dependency-audit, Q-WIRE-1): this replaces
// the prior reflection-based wiring (wiring.go's reflect over the opaque
// *usecases.Aggregate). Drift is now a compile error, not a silent runtime nil.
//
// Naming conventions:
//  1. Field names are SINGULAR matching the proto folder / entity name.
//  2. Group struct types use the `<Entity>UseCases` suffix.
//  3. Closure signatures use proto request/response types (no block-local
//     transport types), EXCEPT the two dashboard slots, which return the fayna
//     VIEW types (the proto→view translation lives in service-admin's
//     adapters.go, where both vocabularies are visible — Round 2).
package block

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	staffpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/staff"
	fulfillmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/fulfillment"
	activityexpensepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_expense"
	activitylaborpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_labor"
	activitymaterialpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_material"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
	joboutcomesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplateTaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	phaseoutcomesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	templatetaskcriteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
	subscriptionpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription"
	activitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/workflow/activity"

	fulfillmentdashboard "github.com/erniealice/fayna-golang/domain/fulfillment/fulfillment/dashboard"
	jobdashboard "github.com/erniealice/fayna-golang/domain/operation/job/dashboard"
)

// UseCases declares everything fayna's Block() needs from outside.
// Construction is service-admin's job; fayna only declares the shape.
type UseCases struct {
	// GetWorkspaceIDFromCtx extracts the workspace ID from a request context.
	// Wired by service-admin as consumer.GetWorkspaceIDFromContext. Used by the
	// dashboard wrappers (Round 2: in adapters.go) for the empty-workspace
	// fallback so postgres dashboard queries don't render cross-workspace
	// aggregates when the view-layer Request omits a workspace_id.
	GetWorkspaceIDFromCtx func(ctx context.Context) string

	// Operation — the operation domain entity use-case groups (jobs, templates,
	// phases, tasks, activities, outcomes, summaries).
	Operation OperationUseCases

	// Fulfillment — the fulfillment domain use cases.
	Fulfillment FulfillmentUseCases

	// Subscription — cross-domain read used by the Job detail origin breadcrumb
	// (auto-spawn-jobs-from-subscription plan §5.4).
	Subscription SubscriptionUseCases

	// Entity — cross-domain reads used by the job/activity drawer search pickers
	// (client + staff auto-complete). Optional: nil → flat-filter fallback.
	Entity EntityUseCases

	// Service — service-driven dashboard closures. These return the fayna VIEW
	// types directly (NOT proto types): the proto→view translation lives in
	// service-admin's adapters.go (Round 2), where both the proto Response and
	// the fayna view Response are importable. Until Round 2 wires them, the
	// dashboard func slots are nil and the dashboard views render empty-state.
	Service ServiceUseCases
}

// OperationUseCases mirrors espyna's Operation aggregate sub-groups. Field
// names match the proto entity nesting (Job, JobPhase, JobTask, ...).
type OperationUseCases struct {
	Job              JobUseCases
	JobPhase         JobPhaseUseCases
	JobTask          JobTaskUseCases
	JobActivity      JobActivityUseCases
	JobTemplate      JobTemplateUseCases
	JobTemplatePhase JobTemplatePhaseUseCases
	JobTemplateTask  JobTemplateTaskUseCases
	OutcomeCriteria  OutcomeCriteriaUseCases
	TaskOutcome      TaskOutcomeUseCases
	// TemplateTaskCriteria — JobTemplate detail criteria-by-task list.
	TemplateTaskCriteria TemplateTaskCriteriaUseCases
	JobOutcomeSummary    JobOutcomeSummaryUseCases
	PhaseOutcomeSummary  PhaseOutcomeSummaryUseCases

	// OPTIONAL — NOT in RequireFor (leave nil-able). These sibling charge-detail
	// use cases are not yet in espyna's OperationUseCases (TODO P5); they
	// silently no-op today, so their typed fields stay nil-able.
	ActivityLabor    ActivityLaborUseCases
	ActivityMaterial ActivityMaterialUseCases
	ActivityExpense  ActivityExpenseUseCases
}

// JobUseCases — Job CRUD + list + the cross-tab reads the Job module needs.
type JobUseCases struct {
	CreateJob func(context.Context, *jobpb.CreateJobRequest) (*jobpb.CreateJobResponse, error)
	ReadJob   func(context.Context, *jobpb.ReadJobRequest) (*jobpb.ReadJobResponse, error)
	UpdateJob func(context.Context, *jobpb.UpdateJobRequest) (*jobpb.UpdateJobResponse, error)
	DeleteJob func(context.Context, *jobpb.DeleteJobRequest) (*jobpb.DeleteJobResponse, error)
	ListJobs  func(context.Context, *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error)
}

// JobPhaseUseCases — JobPhase CRUD + list (standalone module + Job detail tab).
type JobPhaseUseCases struct {
	CreateJobPhase func(context.Context, *jobphasepb.CreateJobPhaseRequest) (*jobphasepb.CreateJobPhaseResponse, error)
	ReadJobPhase   func(context.Context, *jobphasepb.ReadJobPhaseRequest) (*jobphasepb.ReadJobPhaseResponse, error)
	UpdateJobPhase func(context.Context, *jobphasepb.UpdateJobPhaseRequest) (*jobphasepb.UpdateJobPhaseResponse, error)
	DeleteJobPhase func(context.Context, *jobphasepb.DeleteJobPhaseRequest) (*jobphasepb.DeleteJobPhaseResponse, error)
	ListJobPhases  func(context.Context, *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error)
}

// JobTaskUseCases — JobTask CRUD + list + ListByPhase.
type JobTaskUseCases struct {
	CreateJobTask       func(context.Context, *jobtaskpb.CreateJobTaskRequest) (*jobtaskpb.CreateJobTaskResponse, error)
	ReadJobTask         func(context.Context, *jobtaskpb.ReadJobTaskRequest) (*jobtaskpb.ReadJobTaskResponse, error)
	UpdateJobTask       func(context.Context, *jobtaskpb.UpdateJobTaskRequest) (*jobtaskpb.UpdateJobTaskResponse, error)
	DeleteJobTask       func(context.Context, *jobtaskpb.DeleteJobTaskRequest) (*jobtaskpb.DeleteJobTaskResponse, error)
	ListJobTasks        func(context.Context, *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error)
	ListJobTasksByPhase func(context.Context, *jobtaskpb.ListJobTasksByPhaseRequest) (*jobtaskpb.ListJobTasksByPhaseResponse, error)
}

// JobActivityUseCases — JobActivity CRUD + list + rollup + approval workflow.
type JobActivityUseCases struct {
	GetJobActivityListPageData func(context.Context, *jobactivitypb.GetJobActivityListPageDataRequest) (*jobactivitypb.GetJobActivityListPageDataResponse, error)
	ReadJobActivity            func(context.Context, *jobactivitypb.ReadJobActivityRequest) (*jobactivitypb.ReadJobActivityResponse, error)
	CreateJobActivity          func(context.Context, *jobactivitypb.CreateJobActivityRequest) (*jobactivitypb.CreateJobActivityResponse, error)
	UpdateJobActivity          func(context.Context, *jobactivitypb.UpdateJobActivityRequest) (*jobactivitypb.UpdateJobActivityResponse, error)
	DeleteJobActivity          func(context.Context, *jobactivitypb.DeleteJobActivityRequest) (*jobactivitypb.DeleteJobActivityResponse, error)
	ListJobActivities          func(context.Context, *jobactivitypb.ListJobActivitiesRequest) (*jobactivitypb.ListJobActivitiesResponse, error)
	GetJobActivityRollup       func(context.Context, *jobactivitypb.GetJobActivityRollupRequest) (*jobactivitypb.GetJobActivityRollupResponse, error)
	SubmitForApproval          func(context.Context, *jobactivitypb.SubmitForApprovalRequest) (*jobactivitypb.SubmitForApprovalResponse, error)
	ApproveActivity            func(context.Context, *jobactivitypb.ApproveJobActivityRequest) (*jobactivitypb.ApproveJobActivityResponse, error)
	RejectActivity             func(context.Context, *jobactivitypb.RejectJobActivityRequest) (*jobactivitypb.RejectJobActivityResponse, error)
}

// JobTemplateUseCases — JobTemplate CRUD + list page data.
type JobTemplateUseCases struct {
	CreateJobTemplate          func(context.Context, *jobtemplatepb.CreateJobTemplateRequest) (*jobtemplatepb.CreateJobTemplateResponse, error)
	ReadJobTemplate            func(context.Context, *jobtemplatepb.ReadJobTemplateRequest) (*jobtemplatepb.ReadJobTemplateResponse, error)
	UpdateJobTemplate          func(context.Context, *jobtemplatepb.UpdateJobTemplateRequest) (*jobtemplatepb.UpdateJobTemplateResponse, error)
	DeleteJobTemplate          func(context.Context, *jobtemplatepb.DeleteJobTemplateRequest) (*jobtemplatepb.DeleteJobTemplateResponse, error)
	GetJobTemplateListPageData func(context.Context, *jobtemplatepb.GetJobTemplateListPageDataRequest) (*jobtemplatepb.GetJobTemplateListPageDataResponse, error)
}

// JobTemplatePhaseUseCases — JobTemplatePhase CRUD + ListByJobTemplate.
type JobTemplatePhaseUseCases struct {
	CreateJobTemplatePhase func(context.Context, *jobtemplatephasepb.CreateJobTemplatePhaseRequest) (*jobtemplatephasepb.CreateJobTemplatePhaseResponse, error)
	ReadJobTemplatePhase   func(context.Context, *jobtemplatephasepb.ReadJobTemplatePhaseRequest) (*jobtemplatephasepb.ReadJobTemplatePhaseResponse, error)
	UpdateJobTemplatePhase func(context.Context, *jobtemplatephasepb.UpdateJobTemplatePhaseRequest) (*jobtemplatephasepb.UpdateJobTemplatePhaseResponse, error)
	DeleteJobTemplatePhase func(context.Context, *jobtemplatephasepb.DeleteJobTemplatePhaseRequest) (*jobtemplatephasepb.DeleteJobTemplatePhaseResponse, error)
	ListByJobTemplate      func(context.Context, *jobtemplatephasepb.ListByJobTemplateRequest) (*jobtemplatephasepb.ListByJobTemplateResponse, error)
}

// JobTemplateTaskUseCases — JobTemplateTask CRUD + ListByPhase.
type JobTemplateTaskUseCases struct {
	CreateJobTemplateTask func(context.Context, *jobtemplateTaskpb.CreateJobTemplateTaskRequest) (*jobtemplateTaskpb.CreateJobTemplateTaskResponse, error)
	ReadJobTemplateTask   func(context.Context, *jobtemplateTaskpb.ReadJobTemplateTaskRequest) (*jobtemplateTaskpb.ReadJobTemplateTaskResponse, error)
	UpdateJobTemplateTask func(context.Context, *jobtemplateTaskpb.UpdateJobTemplateTaskRequest) (*jobtemplateTaskpb.UpdateJobTemplateTaskResponse, error)
	DeleteJobTemplateTask func(context.Context, *jobtemplateTaskpb.DeleteJobTemplateTaskRequest) (*jobtemplateTaskpb.DeleteJobTemplateTaskResponse, error)
	ListByPhase           func(context.Context, *jobtemplateTaskpb.ListJobTemplateTasksByPhaseRequest) (*jobtemplateTaskpb.ListJobTemplateTasksByPhaseResponse, error)
}

// TemplateTaskCriteriaUseCases — criteria-by-task list for the JobTemplate detail.
type TemplateTaskCriteriaUseCases struct {
	ListByTemplateTask func(context.Context, *templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskRequest) (*templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskResponse, error)
}

// OutcomeCriteriaUseCases — OutcomeCriteria CRUD + list.
type OutcomeCriteriaUseCases struct {
	CreateOutcomeCriteria func(context.Context, *criteriapb.CreateOutcomeCriteriaRequest) (*criteriapb.CreateOutcomeCriteriaResponse, error)
	ReadOutcomeCriteria   func(context.Context, *criteriapb.ReadOutcomeCriteriaRequest) (*criteriapb.ReadOutcomeCriteriaResponse, error)
	UpdateOutcomeCriteria func(context.Context, *criteriapb.UpdateOutcomeCriteriaRequest) (*criteriapb.UpdateOutcomeCriteriaResponse, error)
	DeleteOutcomeCriteria func(context.Context, *criteriapb.DeleteOutcomeCriteriaRequest) (*criteriapb.DeleteOutcomeCriteriaResponse, error)
	ListOutcomeCriterias  func(context.Context, *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error)
}

// TaskOutcomeUseCases — TaskOutcome CRUD + list.
type TaskOutcomeUseCases struct {
	CreateTaskOutcome func(context.Context, *taskoutcomepb.CreateTaskOutcomeRequest) (*taskoutcomepb.CreateTaskOutcomeResponse, error)
	ReadTaskOutcome   func(context.Context, *taskoutcomepb.ReadTaskOutcomeRequest) (*taskoutcomepb.ReadTaskOutcomeResponse, error)
	UpdateTaskOutcome func(context.Context, *taskoutcomepb.UpdateTaskOutcomeRequest) (*taskoutcomepb.UpdateTaskOutcomeResponse, error)
	DeleteTaskOutcome func(context.Context, *taskoutcomepb.DeleteTaskOutcomeRequest) (*taskoutcomepb.DeleteTaskOutcomeResponse, error)
	ListTaskOutcomes  func(context.Context, *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error)
}

// JobOutcomeSummaryUseCases — job-level outcome summary reads.
type JobOutcomeSummaryUseCases struct {
	GetByJob                func(context.Context, *joboutcomesumpb.GetJobOutcomeSummaryByJobRequest) (*joboutcomesumpb.GetJobOutcomeSummaryByJobResponse, error)
	ListJobOutcomeSummaries func(context.Context, *joboutcomesumpb.ListJobOutcomeSummarysRequest) (*joboutcomesumpb.ListJobOutcomeSummarysResponse, error)
}

// PhaseOutcomeSummaryUseCases — phase-level outcome summary reads.
type PhaseOutcomeSummaryUseCases struct {
	GetByJobPhase func(context.Context, *phaseoutcomesumpb.GetPhaseOutcomeSummaryByJobPhaseRequest) (*phaseoutcomesumpb.GetPhaseOutcomeSummaryByJobPhaseResponse, error)
	ListByJob     func(context.Context, *phaseoutcomesumpb.ListPhaseOutcomeSummarysByJobRequest) (*phaseoutcomesumpb.ListPhaseOutcomeSummarysByJobResponse, error)
}

// ActivityLaborUseCases — OPTIONAL (not in RequireFor; nil-able until espyna P5).
type ActivityLaborUseCases struct {
	CreateActivityLabor func(context.Context, *activitylaborpb.CreateActivityLaborRequest) (*activitylaborpb.CreateActivityLaborResponse, error)
	ReadActivityLabor   func(context.Context, *activitylaborpb.ReadActivityLaborRequest) (*activitylaborpb.ReadActivityLaborResponse, error)
	UpdateActivityLabor func(context.Context, *activitylaborpb.UpdateActivityLaborRequest) (*activitylaborpb.UpdateActivityLaborResponse, error)
	DeleteActivityLabor func(context.Context, *activitylaborpb.DeleteActivityLaborRequest) (*activitylaborpb.DeleteActivityLaborResponse, error)
	ListActivityLabors  func(context.Context, *activitylaborpb.ListActivityLaborsRequest) (*activitylaborpb.ListActivityLaborsResponse, error)
}

// ActivityMaterialUseCases — OPTIONAL (not in RequireFor; nil-able until espyna P5).
type ActivityMaterialUseCases struct {
	CreateActivityMaterial func(context.Context, *activitymaterialpb.CreateActivityMaterialRequest) (*activitymaterialpb.CreateActivityMaterialResponse, error)
	ReadActivityMaterial   func(context.Context, *activitymaterialpb.ReadActivityMaterialRequest) (*activitymaterialpb.ReadActivityMaterialResponse, error)
	UpdateActivityMaterial func(context.Context, *activitymaterialpb.UpdateActivityMaterialRequest) (*activitymaterialpb.UpdateActivityMaterialResponse, error)
	DeleteActivityMaterial func(context.Context, *activitymaterialpb.DeleteActivityMaterialRequest) (*activitymaterialpb.DeleteActivityMaterialResponse, error)
	ListActivityMaterials  func(context.Context, *activitymaterialpb.ListActivityMaterialsRequest) (*activitymaterialpb.ListActivityMaterialsResponse, error)
}

// ActivityExpenseUseCases — OPTIONAL (not in RequireFor; nil-able until espyna P5).
type ActivityExpenseUseCases struct {
	CreateActivityExpense func(context.Context, *activityexpensepb.CreateActivityExpenseRequest) (*activityexpensepb.CreateActivityExpenseResponse, error)
	ReadActivityExpense   func(context.Context, *activityexpensepb.ReadActivityExpenseRequest) (*activityexpensepb.ReadActivityExpenseResponse, error)
	UpdateActivityExpense func(context.Context, *activityexpensepb.UpdateActivityExpenseRequest) (*activityexpensepb.UpdateActivityExpenseResponse, error)
	DeleteActivityExpense func(context.Context, *activityexpensepb.DeleteActivityExpenseRequest) (*activityexpensepb.DeleteActivityExpenseResponse, error)
	ListActivityExpenses  func(context.Context, *activityexpensepb.ListActivityExpensesRequest) (*activityexpensepb.ListActivityExpensesResponse, error)
}

// FulfillmentUseCases — Fulfillment CRUD + status transition + page data.
type FulfillmentUseCases struct {
	GetFulfillmentListPageData func(context.Context, *fulfillmentpb.GetFulfillmentListPageDataRequest) (*fulfillmentpb.GetFulfillmentListPageDataResponse, error)
	GetFulfillmentItemPageData func(context.Context, *fulfillmentpb.GetFulfillmentItemPageDataRequest) (*fulfillmentpb.GetFulfillmentItemPageDataResponse, error)
	CreateFulfillment          func(context.Context, *fulfillmentpb.CreateFulfillmentRequest) (*fulfillmentpb.CreateFulfillmentResponse, error)
	UpdateFulfillment          func(context.Context, *fulfillmentpb.UpdateFulfillmentRequest) (*fulfillmentpb.UpdateFulfillmentResponse, error)
	DeleteFulfillment          func(context.Context, *fulfillmentpb.DeleteFulfillmentRequest) (*fulfillmentpb.DeleteFulfillmentResponse, error)
	TransitionStatus           func(context.Context, *fulfillmentpb.TransitionStatusRequest) (*fulfillmentpb.TransitionStatusResponse, error)
}

// SubscriptionUseCases — cross-domain Subscription read (Job origin breadcrumb).
// Mirrors the aggregate's Subscription.Subscription.ReadSubscription path.
type SubscriptionUseCases struct {
	Subscription SubscriptionSubscriptionUseCases
}

type SubscriptionSubscriptionUseCases struct {
	ReadSubscription func(context.Context, *subscriptionpb.ReadSubscriptionRequest) (*subscriptionpb.ReadSubscriptionResponse, error)
}

// EntityUseCases — cross-domain entity reads for the drawer search pickers.
// Optional: nil → the auto-complete drawer falls back to flat-filter mode.
type EntityUseCases struct {
	Client EntityClientUseCases
	Staff  EntityStaffUseCases
}

type EntityClientUseCases struct {
	SearchClientsByName func(context.Context, *clientpb.SearchClientsByNameRequest) (*clientpb.SearchClientsByNameResponse, error)
	ListClients         func(context.Context, *clientpb.ListClientsRequest) (*clientpb.ListClientsResponse, error)
}

type EntityStaffUseCases struct {
	ListStaffs func(context.Context, *staffpb.ListStaffsRequest) (*staffpb.ListStaffsResponse, error)
}

// ServiceUseCases — service-driven dashboards + engine identity bridge. The
// Dashboard func slots return the fayna VIEW types: proto→view translation is
// service-admin's job (Round 2, adapters.go), where both the proto Response and
// the fayna view Response are importable without a dependency cycle. nil until
// Round 2 wires them.
type ServiceUseCases struct {
	Dashboard DashboardUseCases

	// Workflow carries the engine identity bridge (read-only). The closure
	// wraps espyna's WorkflowAssigneeQueryService.ListPendingActivitiesForAssignee
	// so that view modules can render a "My Approvals" / "Assigned to Me"
	// queue without importing espyna or knowing the bridge SQL shape. The
	// two identity inputs (workspace_user_id, workspace_id) are sourced
	// from session context by service-admin's adapter closure — the view
	// layer passes them through the request struct but NEVER reads them
	// from wire / form params.
	//
	// OPTIONAL — nil until the engine identity bridge is wired (Phase 5).
	// nil → "My Approvals" view renders empty-state gracefully.
	Workflow WorkflowUseCases
}

// WorkflowUseCases — engine identity bridge read surface. Fields are OPTIONAL
// (nil-able); fayna's RequireFor does not gate on them because the bridge is a
// downstream capability that degrades gracefully to empty-state.
type WorkflowUseCases struct {
	// ListPendingActivitiesForAssignee returns engine activities assigned to
	// the logged-in human, scoped to the active workspace. nil → empty-state.
	ListPendingActivitiesForAssignee func(ctx context.Context, req *WorkflowAssigneeQueryRequest) (*WorkflowAssigneeQueryResponse, error)
}

// WorkflowAssigneeQueryRequest carries the identity inputs for the engine
// identity bridge query. Both WorkspaceUserID and WorkspaceID are sourced
// from session context by the adapter closure — NEVER from request params.
type WorkflowAssigneeQueryRequest struct {
	WorkspaceUserID string
	WorkspaceID     string
	Limit           int
	Offset          int
}

// WorkflowAssigneeQueryResponse wraps the query results.
type WorkflowAssigneeQueryResponse struct {
	Activities []*activitypb.Activity
	Total      int
}

type DashboardUseCases struct {
	// Job returns the job dashboard view payload. nil → empty-state.
	Job func(ctx context.Context, req *jobdashboard.Request) (*jobdashboard.Response, error)
	// Fulfillment returns the fulfillment dashboard view payload. nil → empty-state.
	Fulfillment func(ctx context.Context, req *fulfillmentdashboard.Request) (*fulfillmentdashboard.Response, error)
}

// RequireFor returns an error listing every needed-but-nil field for cfg's
// enabled modules. Called at Block() entry; a missing field → startup error.
//
// CRITICAL: this is the deterministic completeness check that replaces the
// prior silent-nil reflection drift. Partial wiring is a startup error, not a
// runtime nil panic.
//
// NOT checked (intentionally optional, nil-able):
//   - Operation.{ActivityLabor, ActivityMaterial, ActivityExpense} — not yet in
//     espyna's OperationUseCases (TODO P5); they silently no-op today.
//   - Entity.{Client, Staff} — drawer search; nil → flat-filter fallback.
//   - Subscription.Subscription.ReadSubscription — origin breadcrumb; nil hides it.
//   - Service.Dashboard.{Job, Fulfillment} — nil until Round 2; nil → empty-state.
func (u *UseCases) RequireFor(cfg *blockConfig) error {
	if u == nil {
		return fmt.Errorf("fayna.Block: WithUseCases(...) was not supplied")
	}

	var missing []string
	check := func(ok bool, name string) {
		if !ok {
			missing = append(missing, name)
		}
	}

	if cfg.wantJob() {
		op := &u.Operation.Job
		check(op.CreateJob != nil, "UseCases.Operation.Job.CreateJob")
		check(op.ReadJob != nil, "UseCases.Operation.Job.ReadJob")
		check(op.UpdateJob != nil, "UseCases.Operation.Job.UpdateJob")
		check(op.DeleteJob != nil, "UseCases.Operation.Job.DeleteJob")
		check(op.ListJobs != nil, "UseCases.Operation.Job.ListJobs")
	}

	if cfg.wantJobTemplate() {
		jt := &u.Operation.JobTemplate
		check(jt.CreateJobTemplate != nil, "UseCases.Operation.JobTemplate.CreateJobTemplate")
		check(jt.ReadJobTemplate != nil, "UseCases.Operation.JobTemplate.ReadJobTemplate")
		check(jt.UpdateJobTemplate != nil, "UseCases.Operation.JobTemplate.UpdateJobTemplate")
		check(jt.DeleteJobTemplate != nil, "UseCases.Operation.JobTemplate.DeleteJobTemplate")
		check(jt.GetJobTemplateListPageData != nil, "UseCases.Operation.JobTemplate.GetJobTemplateListPageData")

		// The JobTemplate detail Tasks + Standards tabs (views/job_template/detail/
		// tasks.go + standards.go, wired via wireJobTemplateDeps) walk
		// phases → tasks → criteria. These three cross-entity list closures are
		// REQUIRED whenever JobTemplate is enabled — without them the detail tabs
		// silently render empty (degraded, not a panic). They are provided by
		// service-admin's buildFaynaUseCases (adapters.go: JobTemplatePhase.
		// ListByJobTemplate, JobTemplateTask.ListByPhase, TemplateTaskCriteria.
		// ListByTemplateTask), so asserting them here is boot-safe and surfaces
		// any future wiring gap at startup instead of as an empty tab.
		check(u.Operation.JobTemplatePhase.ListByJobTemplate != nil, "UseCases.Operation.JobTemplatePhase.ListByJobTemplate")
		check(u.Operation.JobTemplateTask.ListByPhase != nil, "UseCases.Operation.JobTemplateTask.ListByPhase")
		check(u.Operation.TemplateTaskCriteria.ListByTemplateTask != nil, "UseCases.Operation.TemplateTaskCriteria.ListByTemplateTask")
	}

	if cfg.wantJobActivity() {
		ja := &u.Operation.JobActivity
		check(ja.GetJobActivityListPageData != nil, "UseCases.Operation.JobActivity.GetJobActivityListPageData")
		check(ja.ReadJobActivity != nil, "UseCases.Operation.JobActivity.ReadJobActivity")
		check(ja.CreateJobActivity != nil, "UseCases.Operation.JobActivity.CreateJobActivity")
		check(ja.UpdateJobActivity != nil, "UseCases.Operation.JobActivity.UpdateJobActivity")
		check(ja.DeleteJobActivity != nil, "UseCases.Operation.JobActivity.DeleteJobActivity")
		check(ja.ListJobActivities != nil, "UseCases.Operation.JobActivity.ListJobActivities")
	}

	if cfg.wantJobPhase() {
		jp := &u.Operation.JobPhase
		check(jp.CreateJobPhase != nil, "UseCases.Operation.JobPhase.CreateJobPhase")
		check(jp.ReadJobPhase != nil, "UseCases.Operation.JobPhase.ReadJobPhase")
		check(jp.UpdateJobPhase != nil, "UseCases.Operation.JobPhase.UpdateJobPhase")
		check(jp.DeleteJobPhase != nil, "UseCases.Operation.JobPhase.DeleteJobPhase")
		check(jp.ListJobPhases != nil, "UseCases.Operation.JobPhase.ListJobPhases")
	}

	if cfg.wantJobTask() {
		jt := &u.Operation.JobTask
		check(jt.CreateJobTask != nil, "UseCases.Operation.JobTask.CreateJobTask")
		check(jt.ReadJobTask != nil, "UseCases.Operation.JobTask.ReadJobTask")
		check(jt.UpdateJobTask != nil, "UseCases.Operation.JobTask.UpdateJobTask")
		check(jt.DeleteJobTask != nil, "UseCases.Operation.JobTask.DeleteJobTask")
		check(jt.ListJobTasks != nil, "UseCases.Operation.JobTask.ListJobTasks")
	}

	if cfg.wantJobTemplatePhase() {
		jtp := &u.Operation.JobTemplatePhase
		check(jtp.CreateJobTemplatePhase != nil, "UseCases.Operation.JobTemplatePhase.CreateJobTemplatePhase")
		check(jtp.ReadJobTemplatePhase != nil, "UseCases.Operation.JobTemplatePhase.ReadJobTemplatePhase")
		check(jtp.UpdateJobTemplatePhase != nil, "UseCases.Operation.JobTemplatePhase.UpdateJobTemplatePhase")
		check(jtp.DeleteJobTemplatePhase != nil, "UseCases.Operation.JobTemplatePhase.DeleteJobTemplatePhase")
	}

	if cfg.wantJobTemplateTask() {
		jtt := &u.Operation.JobTemplateTask
		check(jtt.CreateJobTemplateTask != nil, "UseCases.Operation.JobTemplateTask.CreateJobTemplateTask")
		check(jtt.ReadJobTemplateTask != nil, "UseCases.Operation.JobTemplateTask.ReadJobTemplateTask")
		check(jtt.UpdateJobTemplateTask != nil, "UseCases.Operation.JobTemplateTask.UpdateJobTemplateTask")
		check(jtt.DeleteJobTemplateTask != nil, "UseCases.Operation.JobTemplateTask.DeleteJobTemplateTask")
	}

	if cfg.wantOutcomeCriteria() {
		oc := &u.Operation.OutcomeCriteria
		check(oc.CreateOutcomeCriteria != nil, "UseCases.Operation.OutcomeCriteria.CreateOutcomeCriteria")
		check(oc.ReadOutcomeCriteria != nil, "UseCases.Operation.OutcomeCriteria.ReadOutcomeCriteria")
		check(oc.UpdateOutcomeCriteria != nil, "UseCases.Operation.OutcomeCriteria.UpdateOutcomeCriteria")
		check(oc.DeleteOutcomeCriteria != nil, "UseCases.Operation.OutcomeCriteria.DeleteOutcomeCriteria")
		check(oc.ListOutcomeCriterias != nil, "UseCases.Operation.OutcomeCriteria.ListOutcomeCriterias")
	}

	if cfg.wantTaskOutcome() {
		to := &u.Operation.TaskOutcome
		check(to.CreateTaskOutcome != nil, "UseCases.Operation.TaskOutcome.CreateTaskOutcome")
		check(to.ReadTaskOutcome != nil, "UseCases.Operation.TaskOutcome.ReadTaskOutcome")
		check(to.UpdateTaskOutcome != nil, "UseCases.Operation.TaskOutcome.UpdateTaskOutcome")
		check(to.DeleteTaskOutcome != nil, "UseCases.Operation.TaskOutcome.DeleteTaskOutcome")
		check(to.ListTaskOutcomes != nil, "UseCases.Operation.TaskOutcome.ListTaskOutcomes")
	}

	if cfg.wantOutcomeSummary() {
		jos := &u.Operation.JobOutcomeSummary
		check(jos.GetByJob != nil, "UseCases.Operation.JobOutcomeSummary.GetByJob")
		check(jos.ListJobOutcomeSummaries != nil, "UseCases.Operation.JobOutcomeSummary.ListJobOutcomeSummaries")
		pos := &u.Operation.PhaseOutcomeSummary
		check(pos.GetByJobPhase != nil, "UseCases.Operation.PhaseOutcomeSummary.GetByJobPhase")
		check(pos.ListByJob != nil, "UseCases.Operation.PhaseOutcomeSummary.ListByJob")
	}

	if cfg.wantFulfillment() {
		ff := &u.Fulfillment
		check(ff.GetFulfillmentListPageData != nil, "UseCases.Fulfillment.GetFulfillmentListPageData")
		check(ff.GetFulfillmentItemPageData != nil, "UseCases.Fulfillment.GetFulfillmentItemPageData")
		check(ff.CreateFulfillment != nil, "UseCases.Fulfillment.CreateFulfillment")
		check(ff.UpdateFulfillment != nil, "UseCases.Fulfillment.UpdateFulfillment")
		check(ff.DeleteFulfillment != nil, "UseCases.Fulfillment.DeleteFulfillment")
		check(ff.TransitionStatus != nil, "UseCases.Fulfillment.TransitionStatus")
	}

	if len(missing) > 0 {
		return fmt.Errorf("fayna.Block: incomplete UseCases — missing %v", missing)
	}
	return nil
}

// MustValidate is the FAIL-CLOSED enforcement wrapper around RequireFor. It is
// the seam-level guard that makes a missing REQUIRED closure impossible to
// ignore — mirroring the AUTHZ_ENFORCE boot-guard in service-admin's
// container.go (a missing security precondition is a boot REFUSAL, never a
// silent degrade).
//
// Why a wrapper and not just `return RequireFor(...)`: a bare returned error is
// fail-OPEN by convention. A caller can drop it (`_ =`, an ignored value, a
// future app that doesn't check) and the block silently registers an empty
// feature — the exact nil-closure trap the architecture roast (burn #1) named.
// MustValidate removes that escape hatch:
//
//   - In dev/test (running under `go test`, OR FAYNA_BLOCK_STRICT truthy) a
//     missing REQUIRED closure PANICS with the full field list. A panic cannot
//     be silently dropped, prints a stack trace at the offending wiring site,
//     and fails the test/CI loudly. This is where a developer wiring a new
//     entity discovers a gap — at their desk, not in prod.
//   - In prod a missing REQUIRED closure logs a screaming FATAL line at the
//     seam (so even a caller that drops the returned error leaves an
//     unmissable log record) AND returns the error so Block() propagates it and
//     NewServiceAdmin halts boot with a clear "domain block failed" message.
//
// OPTIONAL ports (Operation.Activity{Labor,Material,Expense}, Entity pickers,
// Subscription breadcrumb, Service dashboards) are NEVER flagged — that
// required-vs-optional discrimination lives entirely in RequireFor, which only
// asserts a field when its enabling cfg.wantXxx() module is on. MustValidate
// adds posture, not policy: it changes HOW a gap fails, not WHICH fields gate.
func (u *UseCases) MustValidate(cfg *blockConfig) error {
	err := u.RequireFor(cfg)
	if err == nil {
		return nil
	}
	if blockStrictMode() {
		// Dev/test: loud, uncatchable-by-accident, stack-traced.
		panic("FATAL: " + err.Error() + " — REQUIRED block wiring is nil. " +
			"Fix the closure assignment in service-admin's buildFaynaUseCases " +
			"(adapters.go) before this reaches prod.")
	}
	// Prod: scream at the seam, then return so boot halts. The log line is the
	// belt to the returned-error's suspenders (a dropped error still screams).
	log.Printf("FATAL: %v — refusing to register fayna modules with a nil "+
		"REQUIRED closure (fail-closed wiring).", err)
	return err
}

// blockStrictMode reports whether the fail-closed wiring guard should PANIC
// (dev/test) rather than return-and-log (prod) on a missing REQUIRED closure.
//
// True when running under `go test` (testing.Testing(), Go 1.21+ — zero env
// coupling, auto-on in every test + CI run) OR when FAYNA_BLOCK_STRICT is set to
// an explicit truthy value (the dev escape hatch for `go run` smoke tests).
// The env matching mirrors container.go's authzEnforceEnabled — anything else
// (unset, "", "0", "false") is prod posture.
func blockStrictMode() bool {
	if testing.Testing() {
		return true
	}
	switch os.Getenv("FAYNA_BLOCK_STRICT") {
	case "1", "true", "TRUE", "True", "yes", "on":
		return true
	default:
		return false
	}
}
