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
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
	joboutcomesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	phaseoutcomesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	subscriptionpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription"

	fulfillmentmod "github.com/erniealice/fayna-golang/views/fulfillment"
	fulfillmentdashboard "github.com/erniealice/fayna-golang/views/fulfillment/dashboard"
	jobmod "github.com/erniealice/fayna-golang/views/job"
	jobdashboard "github.com/erniealice/fayna-golang/views/job/dashboard"
	jobactivitymod "github.com/erniealice/fayna-golang/views/job_activity"
	jobtemplatemod "github.com/erniealice/fayna-golang/views/job_template"
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
		// 2026-04-29 milestone-billing plan §4 — phase set-status surface
		// needs Read (for the existing-name lookup) and Update (which fires
		// the OnJobPhaseCompleted hook).
		if fn, ok := execFn(jp, "ReadJobPhase").(func(context.Context, *jobphasepb.ReadJobPhaseRequest) (*jobphasepb.ReadJobPhaseResponse, error)); ok {
			deps.ReadJobPhase = fn
		}
		if fn, ok := execFn(jp, "UpdateJobPhase").(func(context.Context, *jobphasepb.UpdateJobPhaseRequest) (*jobphasepb.UpdateJobPhaseResponse, error)); ok {
			deps.UpdateJobPhase = fn
		}
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
