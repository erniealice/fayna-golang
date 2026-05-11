package block

// wiring.go provides reflection-based use-case wiring helpers for Block().
//
// The fayna/block sub-package does not import espyna-golang (that would create a
// dependency cycle: fayna/block → espyna → fayna for domain types). Instead, we
// use reflect to navigate the opaque *usecases.Aggregate and extract each
// .Execute method, then type-assert it to the concrete function signature
// expected by each ModuleDeps. This exactly mirrors domain_operations.go but
// without a compile-time espyna import.
//
// Struct field path convention (mirrors espyna's usecases.Aggregate):
//
//	UseCases (Aggregate)
//	  └─ Operation (*OperationUseCases)
//	       ├─ Job (*job.UseCases)
//	       │    ├─ CreateJob (*CreateJobUseCase) → .Execute(ctx, req) (resp, error)
//	       │    ├─ ReadJob   ...
//	       │    └─ ...
//	       ├─ JobTemplate, JobActivity, OutcomeCriteria, TaskOutcome,
//	       │  JobOutcomeSummary, PhaseOutcomeSummary, ...
//	       └─ ...
//	  └─ Fulfillment (*fulfillment.UseCases)
//	       ├─ CreateFulfillment (*CreateFulfillmentUseCase) → .Execute(ctx, req) (resp, error)
//	       └─ ...

import (
	"context"
	"reflect"
	"time"

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

	activityexpensemod "github.com/erniealice/fayna-golang/views/activity_expense"
	activitylabormod "github.com/erniealice/fayna-golang/views/activity_labor"
	activitymaterialmod "github.com/erniealice/fayna-golang/views/activity_material"
	fulfillmentmod "github.com/erniealice/fayna-golang/views/fulfillment"
	fulfillmentdashboard "github.com/erniealice/fayna-golang/views/fulfillment/dashboard"
	jobmod "github.com/erniealice/fayna-golang/views/job"
	jobdashboard "github.com/erniealice/fayna-golang/views/job/dashboard"
	jobactivitymod "github.com/erniealice/fayna-golang/views/job_activity"
	jobphasemod "github.com/erniealice/fayna-golang/views/job_phase"
	jobtaskmod "github.com/erniealice/fayna-golang/views/job_task"
	jobtemplatemod "github.com/erniealice/fayna-golang/views/job_template"
	jobtemplatePhasemod "github.com/erniealice/fayna-golang/views/job_template_phase"
	jobtemplateTaskmod "github.com/erniealice/fayna-golang/views/job_template_task"
	outcomecriteriaMod "github.com/erniealice/fayna-golang/views/outcome_criteria"
	outcomesummaryMod "github.com/erniealice/fayna-golang/views/outcome_summary"
	taskoutcomeMod "github.com/erniealice/fayna-golang/views/task_outcome"
)

// ---------------------------------------------------------------------------
// Reflection helpers
// ---------------------------------------------------------------------------

// ucAggregate wraps the opaque ctx.UseCases value for safe field navigation.
type ucAggregate struct {
	v reflect.Value // dereferenced *usecases.Aggregate struct
}

// assertUseCases wraps ctx.UseCases in a reflection accessor.
// Returns nil if ctx.UseCases is nil or not a pointer-to-struct.
func assertUseCases(raw any) *ucAggregate {
	if raw == nil {
		return nil
	}
	v := reflect.ValueOf(raw)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	return &ucAggregate{v: v}
}

// ptrField safely dereferences a pointer-typed struct field by name.
// Returns zero Value if the field is not found or is nil.
func ptrField(v reflect.Value, name string) reflect.Value {
	if !v.IsValid() {
		return reflect.Value{}
	}
	f := v.FieldByName(name)
	if !f.IsValid() {
		return reflect.Value{}
	}
	if f.Kind() == reflect.Ptr {
		if f.IsNil() {
			return reflect.Value{}
		}
		return f.Elem()
	}
	return f
}

// execFn extracts the Execute method from a pointer-typed use-case leaf field
// (e.g. parent.CreateJob) and returns it as interface{} for type-assertion.
// Returns nil if the field is not found, is nil, or has no Execute method.
func execFn(parent reflect.Value, fieldName string) any {
	if !parent.IsValid() {
		return nil
	}
	f := parent.FieldByName(fieldName)
	if !f.IsValid() {
		return nil
	}
	if f.Kind() == reflect.Ptr {
		if f.IsNil() {
			return nil
		}
		m := f.MethodByName("Execute")
		if m.IsValid() {
			return m.Interface()
		}
		// Try on dereferenced value
		m = f.Elem().MethodByName("Execute")
		if m.IsValid() {
			return m.Interface()
		}
		return nil
	}
	m := f.MethodByName("Execute")
	if !m.IsValid() {
		return nil
	}
	return m.Interface()
}

// ---------------------------------------------------------------------------
// Job module wiring
// ---------------------------------------------------------------------------

func wireJobDeps(deps *jobmod.ModuleDeps, uc *ucAggregate) {
	op := ptrField(uc.v, "Operation")
	if !op.IsValid() {
		return
	}

	j := ptrField(op, "Job")
	if j.IsValid() {
		if fn, ok := execFn(j, "CreateJob").(func(context.Context, *jobpb.CreateJobRequest) (*jobpb.CreateJobResponse, error)); ok {
			deps.CreateJob = fn
		}
		if fn, ok := execFn(j, "ReadJob").(func(context.Context, *jobpb.ReadJobRequest) (*jobpb.ReadJobResponse, error)); ok {
			deps.ReadJob = fn
		}
		if fn, ok := execFn(j, "UpdateJob").(func(context.Context, *jobpb.UpdateJobRequest) (*jobpb.UpdateJobResponse, error)); ok {
			deps.UpdateJob = fn
		}
		if fn, ok := execFn(j, "DeleteJob").(func(context.Context, *jobpb.DeleteJobRequest) (*jobpb.DeleteJobResponse, error)); ok {
			deps.DeleteJob = fn
		}
		if fn, ok := execFn(j, "ListJobs").(func(context.Context, *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error)); ok {
			deps.ListJobs = fn
		}
	}

	jp := ptrField(op, "JobPhase")
	if jp.IsValid() {
		if fn, ok := execFn(jp, "ListJobPhases").(func(context.Context, *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error)); ok {
			deps.ListJobPhases = fn
		}
		// Note: ReadJobPhase and UpdateJobPhase are now wired to the job_phase
		// module (wireJobPhaseDeps), not the job module. The job module only needs
		// ListJobPhases for the Phases tab display.
	}

	jt := ptrField(op, "JobTask")
	if jt.IsValid() {
		if fn, ok := execFn(jt, "ListJobTasks").(func(context.Context, *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error)); ok {
			deps.ListJobTasks = fn
		}
	}

	ja := ptrField(op, "JobActivity")
	if ja.IsValid() {
		if fn, ok := execFn(ja, "ListJobActivities").(func(context.Context, *jobactivitypb.ListJobActivitiesRequest) (*jobactivitypb.ListJobActivitiesResponse, error)); ok {
			deps.ListJobActivities = fn
		}
		// Actuals tab — GetJobActivityRollup is the aggregated cost rollup RPC.
		if fn, ok := execFn(ja, "GetJobActivityRollup").(func(context.Context, *jobactivitypb.GetJobActivityRollupRequest) (*jobactivitypb.GetJobActivityRollupResponse, error)); ok {
			deps.GetJobActivityRollup = fn
		}
	}

	// Budget tab — ReadJobTemplate + ListByJobTemplate + ListByPhase.
	// Mirrors the wireJobTemplateDeps pattern but wired onto job ModuleDeps
	// so the detail budget tab can load template-derived phase/task hours.
	jt2 := ptrField(op, "JobTemplate")
	if jt2.IsValid() {
		if fn, ok := execFn(jt2, "ReadJobTemplate").(func(context.Context, *jobtemplatepb.ReadJobTemplateRequest) (*jobtemplatepb.ReadJobTemplateResponse, error)); ok {
			deps.ReadJobTemplate = fn
		}
	}
	jtp2 := ptrField(op, "JobTemplatePhase")
	if jtp2.IsValid() {
		if fn, ok := execFn(jtp2, "ListByJobTemplate").(func(context.Context, *jobtemplatephasepb.ListByJobTemplateRequest) (*jobtemplatephasepb.ListByJobTemplateResponse, error)); ok {
			deps.ListJobTemplatePhasesByTemplate = fn
		}
	}
	jtt2 := ptrField(op, "JobTemplateTask")
	if jtt2.IsValid() {
		if fn, ok := execFn(jtt2, "ListByPhase").(func(context.Context, *jobtemplateTaskpb.ListJobTemplateTasksByPhaseRequest) (*jobtemplateTaskpb.ListJobTemplateTasksByPhaseResponse, error)); ok {
			deps.ListJobTemplateTasksByPhase = fn
		}
	}

	// 2026-04-29 auto-spawn-jobs-from-subscription plan §5.4 — wire the
	// cross-domain Subscription read for the Job detail's origin breadcrumb.
	sub := ptrField(uc.v, "Subscription")
	if sub.IsValid() {
		s := ptrField(sub, "Subscription")
		if s.IsValid() {
			if fn, ok := execFn(s, "ReadSubscription").(func(context.Context, *subscriptionpb.ReadSubscriptionRequest) (*subscriptionpb.ReadSubscriptionResponse, error)); ok {
				deps.ReadSubscription = fn
			}
		}
	}
}

// ---------------------------------------------------------------------------
// JobTemplate module wiring
// ---------------------------------------------------------------------------

func wireJobTemplateDeps(deps *jobtemplatemod.ModuleDeps, uc *ucAggregate) {
	op := ptrField(uc.v, "Operation")
	if !op.IsValid() {
		return
	}

	jt := ptrField(op, "JobTemplate")
	if jt.IsValid() {
		if fn, ok := execFn(jt, "CreateJobTemplate").(func(context.Context, *jobtemplatepb.CreateJobTemplateRequest) (*jobtemplatepb.CreateJobTemplateResponse, error)); ok {
			deps.CreateJobTemplate = fn
		}
		if fn, ok := execFn(jt, "ReadJobTemplate").(func(context.Context, *jobtemplatepb.ReadJobTemplateRequest) (*jobtemplatepb.ReadJobTemplateResponse, error)); ok {
			deps.ReadJobTemplate = fn
		}
		if fn, ok := execFn(jt, "UpdateJobTemplate").(func(context.Context, *jobtemplatepb.UpdateJobTemplateRequest) (*jobtemplatepb.UpdateJobTemplateResponse, error)); ok {
			deps.UpdateJobTemplate = fn
		}
		if fn, ok := execFn(jt, "DeleteJobTemplate").(func(context.Context, *jobtemplatepb.DeleteJobTemplateRequest) (*jobtemplatepb.DeleteJobTemplateResponse, error)); ok {
			deps.DeleteJobTemplate = fn
		}
		if fn, ok := execFn(jt, "GetJobTemplateListPageData").(func(context.Context, *jobtemplatepb.GetJobTemplateListPageDataRequest) (*jobtemplatepb.GetJobTemplateListPageDataResponse, error)); ok {
			deps.GetJobTemplateListPageData = fn
		}
	}

	jtp := ptrField(op, "JobTemplatePhase")
	if jtp.IsValid() {
		if fn, ok := execFn(jtp, "ListByJobTemplate").(func(context.Context, *jobtemplatephasepb.ListByJobTemplateRequest) (*jobtemplatephasepb.ListByJobTemplateResponse, error)); ok {
			deps.ListPhasesByJobTemplate = fn
		}
	}

	// P6.template-children stubs — wire ListByPhase and ListByTemplateTask when
	// the JobTemplateTask / TemplateTaskCriteria use-case leafs are present.
	// Nil is safe: the detail loaders show empty-state panels.
	jtt := ptrField(op, "JobTemplateTask")
	if jtt.IsValid() {
		if fn, ok := execFn(jtt, "ListByPhase").(func(context.Context, *jobtemplateTaskpb.ListJobTemplateTasksByPhaseRequest) (*jobtemplateTaskpb.ListJobTemplateTasksByPhaseResponse, error)); ok {
			deps.ListTasksByPhase = fn
		}
	}

	ttc := ptrField(op, "TemplateTaskCriteria")
	if ttc.IsValid() {
		if fn, ok := execFn(ttc, "ListByTemplateTask").(func(context.Context, *templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskRequest) (*templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskResponse, error)); ok {
			deps.ListCriteriaByTask = fn
		}
	}
}

// ---------------------------------------------------------------------------
// JobActivity module wiring
// ---------------------------------------------------------------------------

func wireJobActivityDeps(deps *jobactivitymod.ModuleDeps, uc *ucAggregate) {
	op := ptrField(uc.v, "Operation")
	if !op.IsValid() {
		return
	}

	ja := ptrField(op, "JobActivity")
	if !ja.IsValid() {
		return
	}

	if fn, ok := execFn(ja, "GetJobActivityListPageData").(func(context.Context, *jobactivitypb.GetJobActivityListPageDataRequest) (*jobactivitypb.GetJobActivityListPageDataResponse, error)); ok {
		deps.GetJobActivityListPageData = fn
	}
	if fn, ok := execFn(ja, "ReadJobActivity").(func(context.Context, *jobactivitypb.ReadJobActivityRequest) (*jobactivitypb.ReadJobActivityResponse, error)); ok {
		deps.ReadJobActivity = fn
	}
	if fn, ok := execFn(ja, "CreateJobActivity").(func(context.Context, *jobactivitypb.CreateJobActivityRequest) (*jobactivitypb.CreateJobActivityResponse, error)); ok {
		deps.CreateJobActivity = fn
	}
	if fn, ok := execFn(ja, "UpdateJobActivity").(func(context.Context, *jobactivitypb.UpdateJobActivityRequest) (*jobactivitypb.UpdateJobActivityResponse, error)); ok {
		deps.UpdateJobActivity = fn
	}
	if fn, ok := execFn(ja, "DeleteJobActivity").(func(context.Context, *jobactivitypb.DeleteJobActivityRequest) (*jobactivitypb.DeleteJobActivityResponse, error)); ok {
		deps.DeleteJobActivity = fn
	}
	if fn, ok := execFn(ja, "SubmitForApproval").(func(context.Context, *jobactivitypb.SubmitForApprovalRequest) (*jobactivitypb.SubmitForApprovalResponse, error)); ok {
		deps.SubmitForApproval = fn
	}
	if fn, ok := execFn(ja, "ApproveActivity").(func(context.Context, *jobactivitypb.ApproveJobActivityRequest) (*jobactivitypb.ApproveJobActivityResponse, error)); ok {
		deps.ApproveActivity = fn
	}
	if fn, ok := execFn(ja, "RejectActivity").(func(context.Context, *jobactivitypb.RejectJobActivityRequest) (*jobactivitypb.RejectJobActivityResponse, error)); ok {
		deps.RejectActivity = fn
	}
}

// ---------------------------------------------------------------------------
// JobPhase module wiring
// ---------------------------------------------------------------------------

// wireJobPhaseDeps wires all CRUD + ListJobActivities use-case functions into
// the job_phase ModuleDeps from the espyna Operation aggregate.
func wireJobPhaseDeps(deps *jobphasemod.ModuleDeps, uc *ucAggregate) {
	op := ptrField(uc.v, "Operation")
	if !op.IsValid() {
		return
	}

	jp := ptrField(op, "JobPhase")
	if jp.IsValid() {
		if fn, ok := execFn(jp, "CreateJobPhase").(func(context.Context, *jobphasepb.CreateJobPhaseRequest) (*jobphasepb.CreateJobPhaseResponse, error)); ok {
			deps.CreateJobPhase = fn
		}
		if fn, ok := execFn(jp, "ReadJobPhase").(func(context.Context, *jobphasepb.ReadJobPhaseRequest) (*jobphasepb.ReadJobPhaseResponse, error)); ok {
			deps.ReadJobPhase = fn
		}
		if fn, ok := execFn(jp, "UpdateJobPhase").(func(context.Context, *jobphasepb.UpdateJobPhaseRequest) (*jobphasepb.UpdateJobPhaseResponse, error)); ok {
			deps.UpdateJobPhase = fn
		}
		if fn, ok := execFn(jp, "DeleteJobPhase").(func(context.Context, *jobphasepb.DeleteJobPhaseRequest) (*jobphasepb.DeleteJobPhaseResponse, error)); ok {
			deps.DeleteJobPhase = fn
		}
		if fn, ok := execFn(jp, "ListJobPhases").(func(context.Context, *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error)); ok {
			deps.ListJobPhases = fn
		}
	}

	// Activities tab — in-memory filter of all job activities by the phase's job_id.
	// TODO(WaveN): replace with dedicated ListJobActivitiesByPhase RPC once available.
	ja := ptrField(op, "JobActivity")
	if ja.IsValid() {
		if fn, ok := execFn(ja, "ListJobActivities").(func(context.Context, *jobactivitypb.ListJobActivitiesRequest) (*jobactivitypb.ListJobActivitiesResponse, error)); ok {
			deps.ListJobActivities = fn
		}
	}

	// Tasks tab — ListJobTasksByPhase from the JobTask use case.
	jt := ptrField(op, "JobTask")
	if jt.IsValid() {
		if fn, ok := execFn(jt, "ListJobTasksByPhase").(func(context.Context, *jobtaskpb.ListJobTasksByPhaseRequest) (*jobtaskpb.ListJobTasksByPhaseResponse, error)); ok {
			deps.ListJobTasksByPhase = fn
		}
	}
}

// ---------------------------------------------------------------------------
// OutcomeCriteria module wiring
// ---------------------------------------------------------------------------

func wireOutcomeCriteriaDeps(deps *outcomecriteriaMod.ModuleDeps, uc *ucAggregate) {
	op := ptrField(uc.v, "Operation")
	if !op.IsValid() {
		return
	}

	oc := ptrField(op, "OutcomeCriteria")
	if !oc.IsValid() {
		return
	}

	if fn, ok := execFn(oc, "CreateOutcomeCriteria").(func(context.Context, *criteriapb.CreateOutcomeCriteriaRequest) (*criteriapb.CreateOutcomeCriteriaResponse, error)); ok {
		deps.CreateOutcomeCriteria = fn
	}
	if fn, ok := execFn(oc, "ReadOutcomeCriteria").(func(context.Context, *criteriapb.ReadOutcomeCriteriaRequest) (*criteriapb.ReadOutcomeCriteriaResponse, error)); ok {
		deps.ReadOutcomeCriteria = fn
	}
	if fn, ok := execFn(oc, "UpdateOutcomeCriteria").(func(context.Context, *criteriapb.UpdateOutcomeCriteriaRequest) (*criteriapb.UpdateOutcomeCriteriaResponse, error)); ok {
		deps.UpdateOutcomeCriteria = fn
	}
	if fn, ok := execFn(oc, "DeleteOutcomeCriteria").(func(context.Context, *criteriapb.DeleteOutcomeCriteriaRequest) (*criteriapb.DeleteOutcomeCriteriaResponse, error)); ok {
		deps.DeleteOutcomeCriteria = fn
	}
	if fn, ok := execFn(oc, "ListOutcomeCriteria").(func(context.Context, *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error)); ok {
		deps.ListOutcomeCriterias = fn
	}
}

// ---------------------------------------------------------------------------
// TaskOutcome module wiring
// ---------------------------------------------------------------------------

func wireTaskOutcomeDeps(deps *taskoutcomeMod.ModuleDeps, uc *ucAggregate) {
	op := ptrField(uc.v, "Operation")
	if !op.IsValid() {
		return
	}

	to := ptrField(op, "TaskOutcome")
	if to.IsValid() {
		if fn, ok := execFn(to, "CreateTaskOutcome").(func(context.Context, *taskoutcomepb.CreateTaskOutcomeRequest) (*taskoutcomepb.CreateTaskOutcomeResponse, error)); ok {
			deps.CreateTaskOutcome = fn
		}
		if fn, ok := execFn(to, "ReadTaskOutcome").(func(context.Context, *taskoutcomepb.ReadTaskOutcomeRequest) (*taskoutcomepb.ReadTaskOutcomeResponse, error)); ok {
			deps.ReadTaskOutcome = fn
		}
		if fn, ok := execFn(to, "UpdateTaskOutcome").(func(context.Context, *taskoutcomepb.UpdateTaskOutcomeRequest) (*taskoutcomepb.UpdateTaskOutcomeResponse, error)); ok {
			deps.UpdateTaskOutcome = fn
		}
		if fn, ok := execFn(to, "DeleteTaskOutcome").(func(context.Context, *taskoutcomepb.DeleteTaskOutcomeRequest) (*taskoutcomepb.DeleteTaskOutcomeResponse, error)); ok {
			deps.DeleteTaskOutcome = fn
		}
		if fn, ok := execFn(to, "ListTaskOutcomes").(func(context.Context, *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error)); ok {
			deps.ListTaskOutcomes = fn
		}
	}

	// ReadOutcomeCriteria cross-dep (from OutcomeCriteria use case)
	oc := ptrField(op, "OutcomeCriteria")
	if oc.IsValid() {
		if fn, ok := execFn(oc, "ReadOutcomeCriteria").(func(context.Context, *criteriapb.ReadOutcomeCriteriaRequest) (*criteriapb.ReadOutcomeCriteriaResponse, error)); ok {
			deps.ReadOutcomeCriteria = fn
		}
	}
}

// ---------------------------------------------------------------------------
// OutcomeSummary module wiring
// ---------------------------------------------------------------------------

func wireOutcomeSummaryDeps(deps *outcomesummaryMod.ModuleDeps, uc *ucAggregate) {
	op := ptrField(uc.v, "Operation")
	if !op.IsValid() {
		return
	}

	jos := ptrField(op, "JobOutcomeSummary")
	if jos.IsValid() {
		if fn, ok := execFn(jos, "GetByJob").(func(context.Context, *joboutcomesumpb.GetJobOutcomeSummaryByJobRequest) (*joboutcomesumpb.GetJobOutcomeSummaryByJobResponse, error)); ok {
			deps.GetJobOutcomeSummaryByJob = fn
		}
		if fn, ok := execFn(jos, "ListJobOutcomeSummaries").(func(context.Context, *joboutcomesumpb.ListJobOutcomeSummarysRequest) (*joboutcomesumpb.ListJobOutcomeSummarysResponse, error)); ok {
			deps.ListJobOutcomeSummarys = fn
		}
	}

	pos := ptrField(op, "PhaseOutcomeSummary")
	if pos.IsValid() {
		if fn, ok := execFn(pos, "GetByJobPhase").(func(context.Context, *phaseoutcomesumpb.GetPhaseOutcomeSummaryByJobPhaseRequest) (*phaseoutcomesumpb.GetPhaseOutcomeSummaryByJobPhaseResponse, error)); ok {
			deps.GetPhaseOutcomeSummaryByJobPhase = fn
		}
		if fn, ok := execFn(pos, "ListByJob").(func(context.Context, *phaseoutcomesumpb.ListPhaseOutcomeSummarysByJobRequest) (*phaseoutcomesumpb.ListPhaseOutcomeSummarysByJobResponse, error)); ok {
			deps.ListPhaseOutcomeSummarysByJob = fn
		}
	}
}

// ---------------------------------------------------------------------------
// Fulfillment module wiring
// ---------------------------------------------------------------------------

func wireFulfillmentDeps(deps *fulfillmentmod.ModuleDeps, uc *ucAggregate) {
	ff := ptrField(uc.v, "Fulfillment")
	if !ff.IsValid() {
		return
	}

	if fn, ok := execFn(ff, "GetFulfillmentListPageData").(func(context.Context, *fulfillmentpb.GetFulfillmentListPageDataRequest) (*fulfillmentpb.GetFulfillmentListPageDataResponse, error)); ok {
		deps.GetFulfillmentListPageData = fn
	}
	if fn, ok := execFn(ff, "GetFulfillmentItemPageData").(func(context.Context, *fulfillmentpb.GetFulfillmentItemPageDataRequest) (*fulfillmentpb.GetFulfillmentItemPageDataResponse, error)); ok {
		deps.GetFulfillmentItemPageData = fn
	}
	if fn, ok := execFn(ff, "CreateFulfillment").(func(context.Context, *fulfillmentpb.CreateFulfillmentRequest) (*fulfillmentpb.CreateFulfillmentResponse, error)); ok {
		deps.CreateFulfillment = fn
	}
	if fn, ok := execFn(ff, "UpdateFulfillment").(func(context.Context, *fulfillmentpb.UpdateFulfillmentRequest) (*fulfillmentpb.UpdateFulfillmentResponse, error)); ok {
		deps.UpdateFulfillment = fn
	}
	if fn, ok := execFn(ff, "DeleteFulfillment").(func(context.Context, *fulfillmentpb.DeleteFulfillmentRequest) (*fulfillmentpb.DeleteFulfillmentResponse, error)); ok {
		deps.DeleteFulfillment = fn
	}
	if fn, ok := execFn(ff, "TransitionStatus").(func(context.Context, *fulfillmentpb.TransitionStatusRequest) (*fulfillmentpb.TransitionStatusResponse, error)); ok {
		deps.TransitionStatus = fn
	}
}

// ---------------------------------------------------------------------------
// Dashboard wiring helpers (reflection-based; avoids espyna internal import)
// ---------------------------------------------------------------------------

// callDashboardExecute invokes a *GetXxxDashboardPageDataUseCase.Execute via
// reflection. The use-case pointer is obtained via ptrField. The request struct
// is created reflectively with WorkspaceID and Now set by field name.
// Returns the dereferenced response struct value and an error.
func callDashboardExecute(ucPtr reflect.Value, ctx context.Context, workspaceID string, now time.Time) (reflect.Value, error) {
	if !ucPtr.IsValid() {
		return reflect.Value{}, nil
	}
	if ucPtr.Kind() != reflect.Ptr || ucPtr.IsNil() {
		return reflect.Value{}, nil
	}
	m := ucPtr.MethodByName("Execute")
	if !m.IsValid() {
		return reflect.Value{}, nil
	}
	reqType := m.Type().In(1).Elem()
	reqPtr := reflect.New(reqType)
	if f := reqPtr.Elem().FieldByName("WorkspaceID"); f.IsValid() && f.CanSet() {
		f.SetString(workspaceID)
	}
	if f := reqPtr.Elem().FieldByName("Now"); f.IsValid() && f.CanSet() {
		f.Set(reflect.ValueOf(now))
	}
	results := m.Call([]reflect.Value{reflect.ValueOf(ctx), reqPtr})
	if len(results) < 2 {
		return reflect.Value{}, nil
	}
	if !results[1].IsNil() {
		return reflect.Value{}, results[1].Interface().(error)
	}
	resp := results[0]
	if resp.Kind() == reflect.Ptr && !resp.IsNil() {
		return resp.Elem(), nil
	}
	return resp, nil
}

// ---------------------------------------------------------------------------
// Job dashboard wiring
// ---------------------------------------------------------------------------

// wireJobDashboard sets deps.GetJobDashboardPageData from
// Operation.Dashboard (GetJobDashboardPageDataUseCase).
func wireJobDashboard(deps *jobmod.ModuleDeps, uc *ucAggregate) {
	op := ptrField(uc.v, "Operation")
	if !op.IsValid() {
		return
	}
	dashField := op.FieldByName("Dashboard")
	if !dashField.IsValid() || dashField.IsNil() {
		return
	}

	deps.GetJobDashboardPageData = func(ctx context.Context, req *jobdashboard.Request) (*jobdashboard.Response, error) {
		workspaceID := ""
		var now time.Time
		if req != nil {
			workspaceID = req.WorkspaceID
			now = req.Now
		}
		if now.IsZero() {
			now = time.Now()
		}
		resp, err := callDashboardExecute(dashField, ctx, workspaceID, now)
		if err != nil || !resp.IsValid() {
			return nil, err
		}

		var result jobdashboard.Response
		if s := resp.FieldByName("Stats"); s.IsValid() {
			result.ActiveJobs = s.FieldByName("ActiveJobs").Int()
			result.DoneThisMonth = s.FieldByName("DoneThisMonth").Int()
			result.OverdueJobs = s.FieldByName("OverdueJobs").Int()
			result.HoursThisWeek = s.FieldByName("HoursThisWeek").Float()
		}
		result.TrendLabels, _ = resp.FieldByName("TrendLabels").Interface().([]string)
		result.TrendValues, _ = resp.FieldByName("TrendValues").Interface().([]float64)
		result.UpcomingDeadlines, _ = resp.FieldByName("UpcomingDeadlines").Interface().([]*jobpb.Job)
		result.RecentActivity, _ = resp.FieldByName("RecentActivity").Interface().([]*jobactivitypb.JobActivity)

		// Map RiskTopRows: []JobRisk → []JobRiskRow
		if riskF := resp.FieldByName("RiskTopRows"); riskF.IsValid() && !riskF.IsNil() {
			for i := 0; i < riskF.Len(); i++ {
				s := riskF.Index(i)
				result.RiskTopRows = append(result.RiskTopRows, jobdashboard.JobRiskRow{
					JobID:         s.FieldByName("JobID").String(),
					Code:          s.FieldByName("Code").String(),
					Name:          s.FieldByName("Name").String(),
					CompletionPct: s.FieldByName("CompletionPct").Float(),
					DateEnd:       s.FieldByName("DateEnd").Interface().(time.Time),
				})
			}
		}
		return &result, nil
	}
}

// ---------------------------------------------------------------------------
// Fulfillment dashboard wiring
// ---------------------------------------------------------------------------

// wireFulfillmentDashboard sets deps.GetFulfillmentDashboardPageData from
// Fulfillment.Dashboard (GetFulfillmentDashboardPageDataUseCase).
func wireFulfillmentDashboard(deps *fulfillmentmod.ModuleDeps, uc *ucAggregate) {
	ff := ptrField(uc.v, "Fulfillment")
	if !ff.IsValid() {
		return
	}
	dashField := ff.FieldByName("Dashboard")
	if !dashField.IsValid() || dashField.IsNil() {
		return
	}

	deps.GetFulfillmentDashboardPageData = func(ctx context.Context, req *fulfillmentdashboard.Request) (*fulfillmentdashboard.Response, error) {
		workspaceID := ""
		var now time.Time
		if req != nil {
			workspaceID = req.WorkspaceID
			now = req.Now
		}
		if now.IsZero() {
			now = time.Now()
		}
		resp, err := callDashboardExecute(dashField, ctx, workspaceID, now)
		if err != nil || !resp.IsValid() {
			return nil, err
		}

		var result fulfillmentdashboard.Response
		if s := resp.FieldByName("Stats"); s.IsValid() {
			result.Pending = s.FieldByName("Pending").Int()
			result.InTransit = s.FieldByName("InTransit").Int()
			result.DeliveredToday = s.FieldByName("DeliveredToday").Int()
			result.Exceptions = s.FieldByName("Exceptions").Int()
			result.AvgFulfillDays = s.FieldByName("AvgFulfillDays").Float()
		}
		result.StatusMixLabels, _ = resp.FieldByName("StatusMixLabels").Interface().([]string)
		result.StatusMixValues, _ = resp.FieldByName("StatusMixValues").Interface().([]float64)
		result.TrendLabels, _ = resp.FieldByName("TrendLabels").Interface().([]string)
		result.TrendValues, _ = resp.FieldByName("TrendValues").Interface().([]float64)
		result.RecentExceptions, _ = resp.FieldByName("RecentExceptions").Interface().([]*fulfillmentpb.Fulfillment)
		return &result, nil
	}
}

// ---------------------------------------------------------------------------
// ActivityLabor module wiring
// ---------------------------------------------------------------------------

// wireActivityLaborDeps wires ActivityLabor use case functions from the espyna aggregate.
//
// Struct field path:
//
//	UseCases.Operation.ActivityLabor.*
//
// TODO(P5 wave 3): Add ActivityLabor to espyna's OperationUseCases struct and
// register its use cases (Create/Read/Update/Delete/List) following the same
// pattern as JobActivity. Until then all type assertions silently no-op and
// the ModuleDeps functions remain nil — handlers return clear gap error messages.
func wireActivityLaborDeps(deps *activitylabormod.ModuleDeps, uc *ucAggregate) {
	op := ptrField(uc.v, "Operation")
	if !op.IsValid() {
		return
	}

	al := ptrField(op, "ActivityLabor")
	if !al.IsValid() {
		// ActivityLabor use case not yet in OperationUseCases — expected until P5 wave 3.
		return
	}

	if fn, ok := execFn(al, "CreateActivityLabor").(func(context.Context, *activitylaborpb.CreateActivityLaborRequest) (*activitylaborpb.CreateActivityLaborResponse, error)); ok {
		deps.CreateActivityLabor = fn
	}
	if fn, ok := execFn(al, "ReadActivityLabor").(func(context.Context, *activitylaborpb.ReadActivityLaborRequest) (*activitylaborpb.ReadActivityLaborResponse, error)); ok {
		deps.ReadActivityLabor = fn
	}
	if fn, ok := execFn(al, "UpdateActivityLabor").(func(context.Context, *activitylaborpb.UpdateActivityLaborRequest) (*activitylaborpb.UpdateActivityLaborResponse, error)); ok {
		deps.UpdateActivityLabor = fn
	}
	if fn, ok := execFn(al, "DeleteActivityLabor").(func(context.Context, *activitylaborpb.DeleteActivityLaborRequest) (*activitylaborpb.DeleteActivityLaborResponse, error)); ok {
		deps.DeleteActivityLabor = fn
	}
	if fn, ok := execFn(al, "ListActivityLabors").(func(context.Context, *activitylaborpb.ListActivityLaborsRequest) (*activitylaborpb.ListActivityLaborsResponse, error)); ok {
		deps.ListActivityLabors = fn
	}
}

// ---------------------------------------------------------------------------
// ActivityMaterial module wiring
// ---------------------------------------------------------------------------

// wireActivityMaterialDeps wires ActivityMaterial use case functions from the espyna aggregate.
//
// Struct field path:
//
//	UseCases.Operation.ActivityMaterial.*
//
// TODO(P5): Add ActivityMaterial to espyna's OperationUseCases struct and
// register its use cases (Create/Read/Update/Delete/List) following the same
// pattern as ActivityLabor. Until then all type assertions silently no-op and
// the ModuleDeps functions remain nil — handlers return clear gap error messages.
func wireActivityMaterialDeps(deps *activitymaterialmod.ModuleDeps, uc *ucAggregate) {
	op := ptrField(uc.v, "Operation")
	if !op.IsValid() {
		return
	}

	am := ptrField(op, "ActivityMaterial")
	if !am.IsValid() {
		// ActivityMaterial use case not yet in OperationUseCases — expected until P5.
		return
	}

	if fn, ok := execFn(am, "CreateActivityMaterial").(func(context.Context, *activitymaterialpb.CreateActivityMaterialRequest) (*activitymaterialpb.CreateActivityMaterialResponse, error)); ok {
		deps.CreateActivityMaterial = fn
	}
	if fn, ok := execFn(am, "ReadActivityMaterial").(func(context.Context, *activitymaterialpb.ReadActivityMaterialRequest) (*activitymaterialpb.ReadActivityMaterialResponse, error)); ok {
		deps.ReadActivityMaterial = fn
	}
	if fn, ok := execFn(am, "UpdateActivityMaterial").(func(context.Context, *activitymaterialpb.UpdateActivityMaterialRequest) (*activitymaterialpb.UpdateActivityMaterialResponse, error)); ok {
		deps.UpdateActivityMaterial = fn
	}
	if fn, ok := execFn(am, "DeleteActivityMaterial").(func(context.Context, *activitymaterialpb.DeleteActivityMaterialRequest) (*activitymaterialpb.DeleteActivityMaterialResponse, error)); ok {
		deps.DeleteActivityMaterial = fn
	}
	if fn, ok := execFn(am, "ListActivityMaterials").(func(context.Context, *activitymaterialpb.ListActivityMaterialsRequest) (*activitymaterialpb.ListActivityMaterialsResponse, error)); ok {
		deps.ListActivityMaterials = fn
	}
}

// ---------------------------------------------------------------------------
// ActivityExpense module wiring
// ---------------------------------------------------------------------------

// wireActivityExpenseDeps wires ActivityExpense use case functions from the espyna aggregate.
//
// Struct field path:
//
//	UseCases.Operation.ActivityExpense.*
//
// TODO(P5): Add ActivityExpense to espyna's OperationUseCases struct and
// register its use cases (Create/Read/Update/Delete/List) following the same
// pattern as ActivityLabor. Until then all type assertions silently no-op and
// the ModuleDeps functions remain nil — handlers return clear gap error messages.
func wireActivityExpenseDeps(deps *activityexpensemod.ModuleDeps, uc *ucAggregate) {
	op := ptrField(uc.v, "Operation")
	if !op.IsValid() {
		return
	}

	ae := ptrField(op, "ActivityExpense")
	if !ae.IsValid() {
		// ActivityExpense use case not yet in OperationUseCases — expected until P5.
		return
	}

	if fn, ok := execFn(ae, "CreateActivityExpense").(func(context.Context, *activityexpensepb.CreateActivityExpenseRequest) (*activityexpensepb.CreateActivityExpenseResponse, error)); ok {
		deps.CreateActivityExpense = fn
	}
	if fn, ok := execFn(ae, "ReadActivityExpense").(func(context.Context, *activityexpensepb.ReadActivityExpenseRequest) (*activityexpensepb.ReadActivityExpenseResponse, error)); ok {
		deps.ReadActivityExpense = fn
	}
	if fn, ok := execFn(ae, "UpdateActivityExpense").(func(context.Context, *activityexpensepb.UpdateActivityExpenseRequest) (*activityexpensepb.UpdateActivityExpenseResponse, error)); ok {
		deps.UpdateActivityExpense = fn
	}
	if fn, ok := execFn(ae, "DeleteActivityExpense").(func(context.Context, *activityexpensepb.DeleteActivityExpenseRequest) (*activityexpensepb.DeleteActivityExpenseResponse, error)); ok {
		deps.DeleteActivityExpense = fn
	}
	if fn, ok := execFn(ae, "ListActivityExpenses").(func(context.Context, *activityexpensepb.ListActivityExpensesRequest) (*activityexpensepb.ListActivityExpensesResponse, error)); ok {
		deps.ListActivityExpenses = fn
	}
}

// ---------------------------------------------------------------------------
// JobTemplatePhase module wiring
// ---------------------------------------------------------------------------

// wireJobTemplatePhaseDeps wires CRUD use-case functions into the
// job_template_phase drawer-only ModuleDeps from the espyna Operation aggregate.
func wireJobTemplatePhaseDeps(deps *jobtemplatePhasemod.ModuleDeps, uc *ucAggregate) {
	op := ptrField(uc.v, "Operation")
	if !op.IsValid() {
		return
	}

	jtp := ptrField(op, "JobTemplatePhase")
	if !jtp.IsValid() {
		return
	}

	if fn, ok := execFn(jtp, "CreateJobTemplatePhase").(func(context.Context, *jobtemplatephasepb.CreateJobTemplatePhaseRequest) (*jobtemplatephasepb.CreateJobTemplatePhaseResponse, error)); ok {
		deps.CreateJobTemplatePhase = fn
	}
	if fn, ok := execFn(jtp, "ReadJobTemplatePhase").(func(context.Context, *jobtemplatephasepb.ReadJobTemplatePhaseRequest) (*jobtemplatephasepb.ReadJobTemplatePhaseResponse, error)); ok {
		deps.ReadJobTemplatePhase = fn
	}
	if fn, ok := execFn(jtp, "UpdateJobTemplatePhase").(func(context.Context, *jobtemplatephasepb.UpdateJobTemplatePhaseRequest) (*jobtemplatephasepb.UpdateJobTemplatePhaseResponse, error)); ok {
		deps.UpdateJobTemplatePhase = fn
	}
	if fn, ok := execFn(jtp, "DeleteJobTemplatePhase").(func(context.Context, *jobtemplatephasepb.DeleteJobTemplatePhaseRequest) (*jobtemplatephasepb.DeleteJobTemplatePhaseResponse, error)); ok {
		deps.DeleteJobTemplatePhase = fn
	}
}

// ---------------------------------------------------------------------------
// JobTemplateTask module wiring
// ---------------------------------------------------------------------------

// wireJobTemplateTaskDeps wires CRUD use-case functions into the
// job_template_task drawer-only ModuleDeps from the espyna Operation aggregate.
func wireJobTemplateTaskDeps(deps *jobtemplateTaskmod.ModuleDeps, uc *ucAggregate) {
	op := ptrField(uc.v, "Operation")
	if !op.IsValid() {
		return
	}

	jtt := ptrField(op, "JobTemplateTask")
	if !jtt.IsValid() {
		return
	}

	if fn, ok := execFn(jtt, "CreateJobTemplateTask").(func(context.Context, *jobtemplateTaskpb.CreateJobTemplateTaskRequest) (*jobtemplateTaskpb.CreateJobTemplateTaskResponse, error)); ok {
		deps.CreateJobTemplateTask = fn
	}
	if fn, ok := execFn(jtt, "ReadJobTemplateTask").(func(context.Context, *jobtemplateTaskpb.ReadJobTemplateTaskRequest) (*jobtemplateTaskpb.ReadJobTemplateTaskResponse, error)); ok {
		deps.ReadJobTemplateTask = fn
	}
	if fn, ok := execFn(jtt, "UpdateJobTemplateTask").(func(context.Context, *jobtemplateTaskpb.UpdateJobTemplateTaskRequest) (*jobtemplateTaskpb.UpdateJobTemplateTaskResponse, error)); ok {
		deps.UpdateJobTemplateTask = fn
	}
	if fn, ok := execFn(jtt, "DeleteJobTemplateTask").(func(context.Context, *jobtemplateTaskpb.DeleteJobTemplateTaskRequest) (*jobtemplateTaskpb.DeleteJobTemplateTaskResponse, error)); ok {
		deps.DeleteJobTemplateTask = fn
	}
}

// ---------------------------------------------------------------------------
// JobTask module wiring
// ---------------------------------------------------------------------------

// wireJobTaskDeps wires all CRUD + ListJobActivities use-case functions into
// the job_task ModuleDeps from the espyna Operation aggregate.
func wireJobTaskDeps(deps *jobtaskmod.ModuleDeps, uc *ucAggregate) {
	op := ptrField(uc.v, "Operation")
	if !op.IsValid() {
		return
	}

	jt := ptrField(op, "JobTask")
	if jt.IsValid() {
		if fn, ok := execFn(jt, "CreateJobTask").(func(context.Context, *jobtaskpb.CreateJobTaskRequest) (*jobtaskpb.CreateJobTaskResponse, error)); ok {
			deps.CreateJobTask = fn
		}
		if fn, ok := execFn(jt, "ReadJobTask").(func(context.Context, *jobtaskpb.ReadJobTaskRequest) (*jobtaskpb.ReadJobTaskResponse, error)); ok {
			deps.ReadJobTask = fn
		}
		if fn, ok := execFn(jt, "UpdateJobTask").(func(context.Context, *jobtaskpb.UpdateJobTaskRequest) (*jobtaskpb.UpdateJobTaskResponse, error)); ok {
			deps.UpdateJobTask = fn
		}
		if fn, ok := execFn(jt, "DeleteJobTask").(func(context.Context, *jobtaskpb.DeleteJobTaskRequest) (*jobtaskpb.DeleteJobTaskResponse, error)); ok {
			deps.DeleteJobTask = fn
		}
		if fn, ok := execFn(jt, "ListJobTasks").(func(context.Context, *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error)); ok {
			deps.ListJobTasks = fn
		}
	}

	// Activities tab — filtered by job_task_id from all job activities.
	ja := ptrField(op, "JobActivity")
	if ja.IsValid() {
		if fn, ok := execFn(ja, "ListJobActivities").(func(context.Context, *jobactivitypb.ListJobActivitiesRequest) (*jobactivitypb.ListJobActivitiesResponse, error)); ok {
			deps.ListJobActivities = fn
		}
	}
}
