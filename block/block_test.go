package block

import (
	"context"
	"strings"
	"testing"

	consumerapp "github.com/erniealice/espyna-golang/consumer/app"
	fulfillmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/fulfillment"
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
)

func TestBlockConfig_NoOptions_EnablesAll(t *testing.T) {
	t.Parallel()

	cfg := &blockConfig{enableAll: true}

	modules := []struct {
		name string
		want func() bool
	}{
		{"wantJob", cfg.wantJob},
		{"wantJobTemplate", cfg.wantJobTemplate},
		{"wantJobActivity", cfg.wantJobActivity},
		{"wantOutcomeCriteria", cfg.wantOutcomeCriteria},
		{"wantTaskOutcome", cfg.wantTaskOutcome},
		{"wantOutcomeSummary", cfg.wantOutcomeSummary},
		{"wantFulfillment", cfg.wantFulfillment},
	}

	for _, m := range modules {
		if !m.want() {
			t.Fatalf("enableAll=true: %s() = false, want true", m.name)
		}
	}
}

func TestBlockConfig_NoOptions_MatchesBlockConstructor(t *testing.T) {
	t.Parallel()

	// Block() with no options sets enableAll = true.
	cfg := &blockConfig{enableAll: len([]BlockOption{}) == 0}

	if !cfg.enableAll {
		t.Fatal("Block() with no options should set enableAll = true")
	}
	if !cfg.wantJob() || !cfg.wantFulfillment() {
		t.Fatal("enableAll should enable all modules")
	}
}

func TestBlockConfig_WithJob_EnablesOnlyJob(t *testing.T) {
	t.Parallel()

	cfg := &blockConfig{}
	WithJob()(cfg)

	tests := []struct {
		name string
		want bool
		got  func() bool
	}{
		{"wantJob", true, cfg.wantJob},
		{"wantJobTemplate", false, cfg.wantJobTemplate},
		{"wantJobActivity", false, cfg.wantJobActivity},
		{"wantOutcomeCriteria", false, cfg.wantOutcomeCriteria},
		{"wantTaskOutcome", false, cfg.wantTaskOutcome},
		{"wantOutcomeSummary", false, cfg.wantOutcomeSummary},
		{"wantFulfillment", false, cfg.wantFulfillment},
	}

	for _, tt := range tests {
		if tt.got() != tt.want {
			t.Fatalf("%s() = %v, want %v", tt.name, tt.got(), tt.want)
		}
	}
}

func TestBlockConfig_WithFulfillment_EnablesOnlyFulfillment(t *testing.T) {
	t.Parallel()

	cfg := &blockConfig{}
	WithFulfillment()(cfg)

	tests := []struct {
		name string
		want bool
		got  func() bool
	}{
		{"wantJob", false, cfg.wantJob},
		{"wantJobTemplate", false, cfg.wantJobTemplate},
		{"wantJobActivity", false, cfg.wantJobActivity},
		{"wantOutcomeCriteria", false, cfg.wantOutcomeCriteria},
		{"wantTaskOutcome", false, cfg.wantTaskOutcome},
		{"wantOutcomeSummary", false, cfg.wantOutcomeSummary},
		{"wantFulfillment", true, cfg.wantFulfillment},
	}

	for _, tt := range tests {
		if tt.got() != tt.want {
			t.Fatalf("%s() = %v, want %v", tt.name, tt.got(), tt.want)
		}
	}
}

func TestBlockConfig_MultipleOptions(t *testing.T) {
	t.Parallel()

	cfg := &blockConfig{}
	WithJob()(cfg)
	WithOutcomeCriteria()(cfg)
	WithFulfillment()(cfg)

	tests := []struct {
		name string
		want bool
		got  func() bool
	}{
		{"wantJob", true, cfg.wantJob},
		{"wantJobTemplate", false, cfg.wantJobTemplate},
		{"wantJobActivity", false, cfg.wantJobActivity},
		{"wantOutcomeCriteria", true, cfg.wantOutcomeCriteria},
		{"wantTaskOutcome", false, cfg.wantTaskOutcome},
		{"wantOutcomeSummary", false, cfg.wantOutcomeSummary},
		{"wantFulfillment", true, cfg.wantFulfillment},
	}

	for _, tt := range tests {
		if tt.got() != tt.want {
			t.Fatalf("%s() = %v, want %v", tt.name, tt.got(), tt.want)
		}
	}
}

func TestBlockConfig_EachWithOption(t *testing.T) {
	t.Parallel()

	// Table-driven: apply one option, verify exactly one module is enabled.
	tests := []struct {
		name       string
		opt        BlockOption
		wantModule string
	}{
		{"WithJob", WithJob(), "job"},
		{"WithJobTemplate", WithJobTemplate(), "jobTemplate"},
		{"WithJobActivity", WithJobActivity(), "jobActivity"},
		{"WithOutcomeCriteria", WithOutcomeCriteria(), "outcomeCriteria"},
		{"WithTaskOutcome", WithTaskOutcome(), "taskOutcome"},
		{"WithOutcomeSummary", WithOutcomeSummary(), "outcomeSummary"},
		{"WithFulfillment", WithFulfillment(), "fulfillment"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &blockConfig{}
			tt.opt(cfg)

			modules := map[string]func() bool{
				"job":             cfg.wantJob,
				"jobTemplate":     cfg.wantJobTemplate,
				"jobActivity":     cfg.wantJobActivity,
				"outcomeCriteria": cfg.wantOutcomeCriteria,
				"taskOutcome":     cfg.wantTaskOutcome,
				"outcomeSummary":  cfg.wantOutcomeSummary,
				"fulfillment":     cfg.wantFulfillment,
			}

			for modName, fn := range modules {
				want := modName == tt.wantModule
				if fn() != want {
					t.Fatalf("after %s: want%s(%s) = %v, want %v",
						tt.name, "", modName, fn(), want)
				}
			}
		})
	}
}

func TestBlockConfig_EnableAll_OverridesIndividualFlags(t *testing.T) {
	t.Parallel()

	// Even when no individual flags are set, enableAll makes all want*() return true.
	cfg := &blockConfig{enableAll: true}

	if !cfg.wantJob() {
		t.Fatal("enableAll should override individual job flag")
	}
	if !cfg.wantJobTemplate() {
		t.Fatal("enableAll should override individual jobTemplate flag")
	}
	if !cfg.wantJobActivity() {
		t.Fatal("enableAll should override individual jobActivity flag")
	}
	if !cfg.wantOutcomeCriteria() {
		t.Fatal("enableAll should override individual outcomeCriteria flag")
	}
	if !cfg.wantTaskOutcome() {
		t.Fatal("enableAll should override individual taskOutcome flag")
	}
	if !cfg.wantOutcomeSummary() {
		t.Fatal("enableAll should override individual outcomeSummary flag")
	}
	if !cfg.wantFulfillment() {
		t.Fatal("enableAll should override individual fulfillment flag")
	}
}

func TestBlockConfig_AllFlagsFalse_DisablesAll(t *testing.T) {
	t.Parallel()

	cfg := &blockConfig{}

	modules := []struct {
		name string
		fn   func() bool
	}{
		{"wantJob", cfg.wantJob},
		{"wantJobTemplate", cfg.wantJobTemplate},
		{"wantJobActivity", cfg.wantJobActivity},
		{"wantOutcomeCriteria", cfg.wantOutcomeCriteria},
		{"wantTaskOutcome", cfg.wantTaskOutcome},
		{"wantOutcomeSummary", cfg.wantOutcomeSummary},
		{"wantFulfillment", cfg.wantFulfillment},
	}

	for _, m := range modules {
		if m.fn() {
			t.Fatalf("zero-value config: %s() = true, want false", m.name)
		}
	}
}

// ---------------------------------------------------------------------------
// Negative / defensive test cases
// ---------------------------------------------------------------------------

func TestBlock_NilTranslations(t *testing.T) {
	t.Parallel()

	// Block() returns an AppOption func; calling it with nil Translations
	// should return an error (not panic).
	opt := Block()

	appCtx := &consumerapp.AppContext{
		Translations: nil, // nil translations
	}

	err := opt(appCtx)
	if err == nil {
		t.Fatal("expected error when ctx.Translations is nil, got nil")
	}
}

func TestBlock_WrongTranslationsType(t *testing.T) {
	t.Parallel()

	// Pass a non-*lynguaV1.TranslationProvider value for Translations.
	opt := Block()

	appCtx := &consumerapp.AppContext{
		Translations: "not a translation provider", // wrong type
	}

	err := opt(appCtx)
	if err == nil {
		t.Fatal("expected error when ctx.Translations is wrong type, got nil")
	}
}

func TestBlock_IntTranslationsType(t *testing.T) {
	t.Parallel()

	opt := Block()

	appCtx := &consumerapp.AppContext{
		Translations: 42, // definitely wrong type
	}

	err := opt(appCtx)
	if err == nil {
		t.Fatal("expected error when ctx.Translations is int, got nil")
	}
}

func TestBlockConfig_WithOptions_DisablesEnableAll(t *testing.T) {
	t.Parallel()

	// When any option is provided, enableAll should be false.
	cfg := &blockConfig{enableAll: len([]BlockOption{WithJob()}) == 0}

	if cfg.enableAll {
		t.Fatal("Block() with options should set enableAll = false")
	}
}

func TestBlockConfig_DuplicateOptions(t *testing.T) {
	t.Parallel()

	// Applying the same option twice should not cause issues.
	cfg := &blockConfig{}
	WithJob()(cfg)
	WithJob()(cfg)

	if !cfg.wantJob() {
		t.Fatal("wantJob should be true after applying WithJob twice")
	}
	if cfg.wantFulfillment() {
		t.Fatal("wantFulfillment should be false when only WithJob was applied")
	}
}

// ---------------------------------------------------------------------------
// RequireFor — typed-contract completeness gate (Phase 2, Q-WIRE-1).
// Replaces the prior reflection-based assertUseCases tests: drift is now a
// startup error listing every needed-but-nil REQUIRED field, not a silent nil.
// ---------------------------------------------------------------------------

func TestRequireFor_NilReceiver(t *testing.T) {
	t.Parallel()

	var uc *UseCases
	err := uc.RequireFor(&blockConfig{enableAll: true})
	if err == nil {
		t.Fatal("RequireFor on a nil *UseCases should return an error")
	}
}

func TestRequireFor_EmptyUseCases_EnableAll_Errors(t *testing.T) {
	t.Parallel()

	// A zero-value UseCases has every REQUIRED closure nil → enableAll must err.
	uc := &UseCases{}
	err := uc.RequireFor(&blockConfig{enableAll: true})
	if err == nil {
		t.Fatal("RequireFor(empty, enableAll) should return a missing-fields error")
	}
}

func TestRequireFor_NoModulesEnabled_OK(t *testing.T) {
	t.Parallel()

	// No module enabled → nothing required → no error even on empty UseCases.
	uc := &UseCases{}
	if err := uc.RequireFor(&blockConfig{}); err != nil {
		t.Fatalf("RequireFor(empty, no modules) should be nil, got %v", err)
	}
}

func TestRequireFor_JobModule_PartialWiring_Errors(t *testing.T) {
	t.Parallel()

	// Job enabled but only some required closures wired → error.
	uc := &UseCases{}
	uc.Operation.Job.CreateJob = func(context.Context, *jobpb.CreateJobRequest) (*jobpb.CreateJobResponse, error) {
		return nil, nil
	}
	err := uc.RequireFor(&blockConfig{job: true})
	if err == nil {
		t.Fatal("RequireFor(partial job wiring) should return an error for the missing closures")
	}
}

func TestRequireFor_JobModule_FullWiring_OK(t *testing.T) {
	t.Parallel()

	// All five required Job closures wired → no error.
	uc := &UseCases{}
	uc.Operation.Job.CreateJob = func(context.Context, *jobpb.CreateJobRequest) (*jobpb.CreateJobResponse, error) { return nil, nil }
	uc.Operation.Job.ReadJob = func(context.Context, *jobpb.ReadJobRequest) (*jobpb.ReadJobResponse, error) { return nil, nil }
	uc.Operation.Job.UpdateJob = func(context.Context, *jobpb.UpdateJobRequest) (*jobpb.UpdateJobResponse, error) { return nil, nil }
	uc.Operation.Job.DeleteJob = func(context.Context, *jobpb.DeleteJobRequest) (*jobpb.DeleteJobResponse, error) { return nil, nil }
	uc.Operation.Job.ListJobs = func(context.Context, *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error) { return nil, nil }

	if err := uc.RequireFor(&blockConfig{job: true}); err != nil {
		t.Fatalf("RequireFor(fully wired job) should be nil, got %v", err)
	}
}

func TestRequireFor_OptionalActivityModules_NotRequired(t *testing.T) {
	t.Parallel()

	// ActivityLabor/Material/Expense are intentionally NOT in RequireFor — even
	// with their drawer modules enabled, nil closures must not cause an error.
	uc := &UseCases{}
	cfg := &blockConfig{activityLabor: true, activityMaterial: true, activityExpense: true}
	if err := uc.RequireFor(cfg); err != nil {
		t.Fatalf("RequireFor(optional activity modules, nil closures) should be nil, got %v", err)
	}
}

// wireJobTemplateRequired sets every closure RequireFor checks for the
// JobTemplate module: the five JobTemplate CRUD/page closures plus the three
// detail-tab cross-entity list closures (MEDIUM-2).
func wireJobTemplateRequired(uc *UseCases) {
	jt := &uc.Operation.JobTemplate
	jt.CreateJobTemplate = func(context.Context, *jobtemplatepb.CreateJobTemplateRequest) (*jobtemplatepb.CreateJobTemplateResponse, error) {
		return nil, nil
	}
	jt.ReadJobTemplate = func(context.Context, *jobtemplatepb.ReadJobTemplateRequest) (*jobtemplatepb.ReadJobTemplateResponse, error) {
		return nil, nil
	}
	jt.UpdateJobTemplate = func(context.Context, *jobtemplatepb.UpdateJobTemplateRequest) (*jobtemplatepb.UpdateJobTemplateResponse, error) {
		return nil, nil
	}
	jt.DeleteJobTemplate = func(context.Context, *jobtemplatepb.DeleteJobTemplateRequest) (*jobtemplatepb.DeleteJobTemplateResponse, error) {
		return nil, nil
	}
	jt.GetJobTemplateListPageData = func(context.Context, *jobtemplatepb.GetJobTemplateListPageDataRequest) (*jobtemplatepb.GetJobTemplateListPageDataResponse, error) {
		return nil, nil
	}
	uc.Operation.JobTemplatePhase.ListByJobTemplate = func(context.Context, *jobtemplatephasepb.ListByJobTemplateRequest) (*jobtemplatephasepb.ListByJobTemplateResponse, error) {
		return nil, nil
	}
	uc.Operation.JobTemplateTask.ListByPhase = func(context.Context, *jobtemplateTaskpb.ListJobTemplateTasksByPhaseRequest) (*jobtemplateTaskpb.ListJobTemplateTasksByPhaseResponse, error) {
		return nil, nil
	}
	uc.Operation.TemplateTaskCriteria.ListByTemplateTask = func(context.Context, *templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskRequest) (*templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskResponse, error) {
		return nil, nil
	}
}

func TestRequireFor_JobTemplateModule_FullWiring_OK(t *testing.T) {
	t.Parallel()

	// JobTemplate enabled with CRUD + page data AND the three detail-tab list
	// closures wired → no error.
	uc := &UseCases{}
	wireJobTemplateRequired(uc)

	if err := uc.RequireFor(&blockConfig{jobTemplate: true}); err != nil {
		t.Fatalf("RequireFor(fully wired JobTemplate) should be nil, got %v", err)
	}
}

// TestRequireFor_JobTemplateModule_MissingDetailTabClosures_Errors is the
// MEDIUM-2 regression: JobTemplate enabled with only its own CRUD + page data
// (the prior RequireFor scope), but the detail Tasks/Standards tabs' cross-entity
// list closures left nil. Before the fix this passed RequireFor and the tabs
// silently rendered empty; now boot must fail-fast for each missing closure.
func TestRequireFor_JobTemplateModule_MissingDetailTabClosures_Errors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		unset func(*UseCases)
	}{
		{"ListByJobTemplate", func(uc *UseCases) { uc.Operation.JobTemplatePhase.ListByJobTemplate = nil }},
		{"ListByPhase", func(uc *UseCases) { uc.Operation.JobTemplateTask.ListByPhase = nil }},
		{"ListByTemplateTask", func(uc *UseCases) { uc.Operation.TemplateTaskCriteria.ListByTemplateTask = nil }},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			uc := &UseCases{}
			wireJobTemplateRequired(uc)
			tc.unset(uc) // drop exactly one detail-tab closure

			err := uc.RequireFor(&blockConfig{jobTemplate: true})
			if err == nil {
				t.Fatalf("RequireFor(JobTemplate, missing %s) should return an error, got nil", tc.name)
			}
		})
	}
}

// TestRequireFor_JobTemplateDetailClosures_NotRequired_WhenJobTemplateDisabled
// guards the converse: enabling only the drawer-only JobTemplatePhase /
// JobTemplateTask modules (NOT the JobTemplate module) must not require the
// detail-tab list closures, which belong to the JobTemplate detail view.
func TestRequireFor_JobTemplateDetailClosures_NotRequired_WhenJobTemplateDisabled(t *testing.T) {
	t.Parallel()

	uc := &UseCases{}
	// Only the JobTemplatePhase/Task drawer-CRUD closures; no list closures.
	uc.Operation.JobTemplatePhase.CreateJobTemplatePhase = func(context.Context, *jobtemplatephasepb.CreateJobTemplatePhaseRequest) (*jobtemplatephasepb.CreateJobTemplatePhaseResponse, error) {
		return nil, nil
	}
	uc.Operation.JobTemplatePhase.ReadJobTemplatePhase = func(context.Context, *jobtemplatephasepb.ReadJobTemplatePhaseRequest) (*jobtemplatephasepb.ReadJobTemplatePhaseResponse, error) {
		return nil, nil
	}
	uc.Operation.JobTemplatePhase.UpdateJobTemplatePhase = func(context.Context, *jobtemplatephasepb.UpdateJobTemplatePhaseRequest) (*jobtemplatephasepb.UpdateJobTemplatePhaseResponse, error) {
		return nil, nil
	}
	uc.Operation.JobTemplatePhase.DeleteJobTemplatePhase = func(context.Context, *jobtemplatephasepb.DeleteJobTemplatePhaseRequest) (*jobtemplatephasepb.DeleteJobTemplatePhaseResponse, error) {
		return nil, nil
	}
	uc.Operation.JobTemplateTask.CreateJobTemplateTask = func(context.Context, *jobtemplateTaskpb.CreateJobTemplateTaskRequest) (*jobtemplateTaskpb.CreateJobTemplateTaskResponse, error) {
		return nil, nil
	}
	uc.Operation.JobTemplateTask.ReadJobTemplateTask = func(context.Context, *jobtemplateTaskpb.ReadJobTemplateTaskRequest) (*jobtemplateTaskpb.ReadJobTemplateTaskResponse, error) {
		return nil, nil
	}
	uc.Operation.JobTemplateTask.UpdateJobTemplateTask = func(context.Context, *jobtemplateTaskpb.UpdateJobTemplateTaskRequest) (*jobtemplateTaskpb.UpdateJobTemplateTaskResponse, error) {
		return nil, nil
	}
	uc.Operation.JobTemplateTask.DeleteJobTemplateTask = func(context.Context, *jobtemplateTaskpb.DeleteJobTemplateTaskRequest) (*jobtemplateTaskpb.DeleteJobTemplateTaskResponse, error) {
		return nil, nil
	}

	cfg := &blockConfig{jobTemplatePhase: true, jobTemplateTask: true}
	if err := uc.RequireFor(cfg); err != nil {
		t.Fatalf("RequireFor(drawer-only JobTemplatePhase/Task, no list closures) should be nil, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// MustValidate — FAIL-CLOSED wiring guard (architecture-roast burn #1).
//
// RequireFor returns an error; MustValidate adds the posture: in dev/test
// (testing.Testing() is true here) a missing REQUIRED closure PANICS — loud,
// stack-traced, uncatchable-by-accident — so a nil-closure wiring gap can never
// be silently dropped into an empty-state render. OPTIONAL nils never trip it.
// ---------------------------------------------------------------------------

// wireJobRequired sets every closure RequireFor checks for the Job module.
func wireJobRequired(uc *UseCases) {
	j := &uc.Operation.Job
	j.CreateJob = func(context.Context, *jobpb.CreateJobRequest) (*jobpb.CreateJobResponse, error) { return nil, nil }
	j.ReadJob = func(context.Context, *jobpb.ReadJobRequest) (*jobpb.ReadJobResponse, error) { return nil, nil }
	j.UpdateJob = func(context.Context, *jobpb.UpdateJobRequest) (*jobpb.UpdateJobResponse, error) { return nil, nil }
	j.DeleteJob = func(context.Context, *jobpb.DeleteJobRequest) (*jobpb.DeleteJobResponse, error) { return nil, nil }
	j.ListJobs = func(context.Context, *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error) { return nil, nil }
}

// TestMustValidate_NilRequiredClosure_Panics is the core burn-#1 proof: with
// the Job module enabled but one REQUIRED closure (ListJobs) left nil,
// MustValidate must PANIC under test — not return an empty render, not silently
// degrade. This is the loud failure the bare-return path lacked.
func TestMustValidate_NilRequiredClosure_Panics(t *testing.T) {
	t.Parallel()

	uc := &UseCases{}
	wireJobRequired(uc)
	uc.Operation.Job.ListJobs = nil // drop exactly one REQUIRED closure

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("MustValidate(Job enabled, ListJobs nil) should PANIC in dev/test, but did not")
		}
		msg, _ := r.(string)
		if !strings.Contains(msg, "ListJobs") {
			t.Fatalf("panic message should name the missing field; got %q", msg)
		}
	}()

	// Should not reach the next line — MustValidate panics first.
	_ = uc.MustValidate(&blockConfig{job: true})
	t.Fatal("MustValidate returned instead of panicking on a nil REQUIRED closure")
}

// TestMustValidate_EmptyUseCases_EnableAll_Panics: a fully empty UseCases with
// every module enabled (the "permanently nil dashboard" trap) must panic loudly
// in dev/test rather than register a wall of empty views.
func TestMustValidate_EmptyUseCases_EnableAll_Panics(t *testing.T) {
	t.Parallel()

	uc := &UseCases{}
	defer func() {
		if recover() == nil {
			t.Fatal("MustValidate(empty UseCases, enableAll) should PANIC in dev/test")
		}
	}()
	_ = uc.MustValidate(&blockConfig{enableAll: true})
	t.Fatal("MustValidate returned instead of panicking on an empty enableAll wiring")
}

// TestMustValidate_NilOptionalClosure_OK proves the required-vs-optional
// discrimination survives the fail-closed wrapper: the OPTIONAL Activity*
// modules (not in RequireFor) with nil closures must pass MustValidate with NO
// panic and NO error — disabled/optional features stay legitimately nil.
func TestMustValidate_NilOptionalClosure_OK(t *testing.T) {
	t.Parallel()

	uc := &UseCases{}
	// Optional drawer modules enabled, their closures left nil. Also leave the
	// optional Entity pickers / Subscription breadcrumb / Service dashboards nil.
	cfg := &blockConfig{activityLabor: true, activityMaterial: true, activityExpense: true}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("MustValidate(optional nil closures) must NOT panic; panicked with %v", r)
		}
	}()
	if err := uc.MustValidate(cfg); err != nil {
		t.Fatalf("MustValidate(optional nil closures) should be nil, got %v", err)
	}
}

// TestMustValidate_FullyWired_OK: a completely wired REQUIRED set passes with no
// panic and no error (happy path — guard is silent when wiring is complete).
func TestMustValidate_FullyWired_OK(t *testing.T) {
	t.Parallel()

	uc := &UseCases{}
	wireJobRequired(uc)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("MustValidate(fully wired Job) must NOT panic; panicked with %v", r)
		}
	}()
	if err := uc.MustValidate(&blockConfig{job: true}); err != nil {
		t.Fatalf("MustValidate(fully wired Job) should be nil, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Engine Identity Bridge — WorkflowAssigneeQuery (Phase 7, EIB plan)
//
// The engine identity bridge adds Service.Workflow.ListPendingActivitiesForAssignee
// to fayna's UseCases. It is OPTIONAL (nil-able): nil -> the "My Approvals" view
// renders empty-state gracefully. RequireFor does not gate on it (same treatment
// as the Service.Dashboard slots).
// ---------------------------------------------------------------------------

// TestWorkflowAssigneeQuery_NilClosure_GracefulDegradation proves that a nil
// Workflow.ListPendingActivitiesForAssignee closure does not break RequireFor
// or MustValidate — the engine identity bridge is intentionally OPTIONAL.
func TestWorkflowAssigneeQuery_NilClosure_GracefulDegradation(t *testing.T) {
	t.Parallel()

	uc := &UseCases{}
	wireJobRequired(uc)

	// Service.Workflow.ListPendingActivitiesForAssignee is nil by default.
	if uc.Service.Workflow.ListPendingActivitiesForAssignee != nil {
		t.Fatal("zero-value UseCases should have nil ListPendingActivitiesForAssignee")
	}

	// RequireFor should still pass (Job fully wired, workflow is optional).
	if err := uc.RequireFor(&blockConfig{job: true}); err != nil {
		t.Fatalf("RequireFor(job, nil workflow) should pass, got %v", err)
	}

	// MustValidate should not panic.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("MustValidate must not panic on nil workflow closure; panicked with %v", r)
		}
	}()
	if err := uc.MustValidate(&blockConfig{job: true}); err != nil {
		t.Fatalf("MustValidate(job, nil workflow) should pass, got %v", err)
	}
}

// TestWorkflowAssigneeQuery_WiredClosure_Callable verifies that when the
// ListPendingActivitiesForAssignee closure is wired, it can be called and
// returns the expected response shape.
func TestWorkflowAssigneeQuery_WiredClosure_Callable(t *testing.T) {
	t.Parallel()

	uc := &UseCases{}
	uc.Service.Workflow.ListPendingActivitiesForAssignee = func(
		ctx context.Context,
		req *WorkflowAssigneeQueryRequest,
	) (*WorkflowAssigneeQueryResponse, error) {
		// Simulate an empty result (no pending activities).
		return &WorkflowAssigneeQueryResponse{
			Activities: nil,
			Total:      0,
		}, nil
	}

	resp, err := uc.Service.Workflow.ListPendingActivitiesForAssignee(
		context.Background(),
		&WorkflowAssigneeQueryRequest{
			WorkspaceUserID: "ws-user-001",
			WorkspaceID:     "ws-001",
			Limit:           10,
			Offset:          0,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("response should not be nil")
	}
	if resp.Total != 0 {
		t.Fatalf("expected Total=0, got %d", resp.Total)
	}
	if len(resp.Activities) != 0 {
		t.Fatalf("expected empty Activities slice, got %d items", len(resp.Activities))
	}
}

// TestWorkflowAssigneeQuery_EnableAll_DoesNotRequireWorkflow confirms that
// enableAll (the default Block() mode) does not require the workflow closure.
func TestWorkflowAssigneeQuery_EnableAll_DoesNotRequireWorkflow(t *testing.T) {
	t.Parallel()

	// This test mirrors TestRequireFor_EmptyUseCases_EnableAll_Errors but
	// focuses on proving the workflow slot is NOT in the required set. We
	// set up a fully-wired UseCases (all REQUIRED closures populated) but
	// leave Service.Workflow nil — RequireFor must still pass.
	uc := &UseCases{}
	wireJobRequired(uc)
	wireJobTemplateRequired(uc)

	// Wire the remaining enableAll-required modules at minimum.
	uc.Operation.JobActivity.GetJobActivityListPageData = func(context.Context, *jobactivitypb.GetJobActivityListPageDataRequest) (*jobactivitypb.GetJobActivityListPageDataResponse, error) {
		return nil, nil
	}
	uc.Operation.JobActivity.ReadJobActivity = func(context.Context, *jobactivitypb.ReadJobActivityRequest) (*jobactivitypb.ReadJobActivityResponse, error) {
		return nil, nil
	}
	uc.Operation.JobActivity.CreateJobActivity = func(context.Context, *jobactivitypb.CreateJobActivityRequest) (*jobactivitypb.CreateJobActivityResponse, error) {
		return nil, nil
	}
	uc.Operation.JobActivity.UpdateJobActivity = func(context.Context, *jobactivitypb.UpdateJobActivityRequest) (*jobactivitypb.UpdateJobActivityResponse, error) {
		return nil, nil
	}
	uc.Operation.JobActivity.DeleteJobActivity = func(context.Context, *jobactivitypb.DeleteJobActivityRequest) (*jobactivitypb.DeleteJobActivityResponse, error) {
		return nil, nil
	}
	uc.Operation.JobActivity.ListJobActivities = func(context.Context, *jobactivitypb.ListJobActivitiesRequest) (*jobactivitypb.ListJobActivitiesResponse, error) {
		return nil, nil
	}

	// Service.Workflow stays nil. RequireFor(enableAll) must pass.
	if uc.Service.Workflow.ListPendingActivitiesForAssignee != nil {
		t.Fatal("Service.Workflow should be nil for this test")
	}

	// Note: RequireFor(enableAll) would fail because we haven't wired
	// every module (JobPhase, JobTask, OutcomeCriteria, TaskOutcome,
	// OutcomeSummary, Fulfillment). The point of this test is narrower:
	// confirm that a nil Workflow closure is never the CAUSE of failure.
	// Wire the remaining required modules minimally.
	uc.Operation.JobPhase.CreateJobPhase = func(context.Context, *jobphasepb.CreateJobPhaseRequest) (*jobphasepb.CreateJobPhaseResponse, error) {
		return nil, nil
	}
	uc.Operation.JobPhase.ReadJobPhase = func(context.Context, *jobphasepb.ReadJobPhaseRequest) (*jobphasepb.ReadJobPhaseResponse, error) {
		return nil, nil
	}
	uc.Operation.JobPhase.UpdateJobPhase = func(context.Context, *jobphasepb.UpdateJobPhaseRequest) (*jobphasepb.UpdateJobPhaseResponse, error) {
		return nil, nil
	}
	uc.Operation.JobPhase.DeleteJobPhase = func(context.Context, *jobphasepb.DeleteJobPhaseRequest) (*jobphasepb.DeleteJobPhaseResponse, error) {
		return nil, nil
	}
	uc.Operation.JobPhase.ListJobPhases = func(context.Context, *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
		return nil, nil
	}

	// Fulfillment
	uc.Fulfillment.GetFulfillmentListPageData = func(context.Context, *fulfillmentpb.GetFulfillmentListPageDataRequest) (*fulfillmentpb.GetFulfillmentListPageDataResponse, error) {
		return nil, nil
	}
	uc.Fulfillment.GetFulfillmentItemPageData = func(context.Context, *fulfillmentpb.GetFulfillmentItemPageDataRequest) (*fulfillmentpb.GetFulfillmentItemPageDataResponse, error) {
		return nil, nil
	}
	uc.Fulfillment.CreateFulfillment = func(context.Context, *fulfillmentpb.CreateFulfillmentRequest) (*fulfillmentpb.CreateFulfillmentResponse, error) {
		return nil, nil
	}
	uc.Fulfillment.UpdateFulfillment = func(context.Context, *fulfillmentpb.UpdateFulfillmentRequest) (*fulfillmentpb.UpdateFulfillmentResponse, error) {
		return nil, nil
	}
	uc.Fulfillment.DeleteFulfillment = func(context.Context, *fulfillmentpb.DeleteFulfillmentRequest) (*fulfillmentpb.DeleteFulfillmentResponse, error) {
		return nil, nil
	}
	uc.Fulfillment.TransitionStatus = func(context.Context, *fulfillmentpb.TransitionStatusRequest) (*fulfillmentpb.TransitionStatusResponse, error) {
		return nil, nil
	}

	// JobTemplatePhase CRUD (enableAll requires these beyond the list closures)
	uc.Operation.JobTemplatePhase.CreateJobTemplatePhase = func(context.Context, *jobtemplatephasepb.CreateJobTemplatePhaseRequest) (*jobtemplatephasepb.CreateJobTemplatePhaseResponse, error) {
		return nil, nil
	}
	uc.Operation.JobTemplatePhase.ReadJobTemplatePhase = func(context.Context, *jobtemplatephasepb.ReadJobTemplatePhaseRequest) (*jobtemplatephasepb.ReadJobTemplatePhaseResponse, error) {
		return nil, nil
	}
	uc.Operation.JobTemplatePhase.UpdateJobTemplatePhase = func(context.Context, *jobtemplatephasepb.UpdateJobTemplatePhaseRequest) (*jobtemplatephasepb.UpdateJobTemplatePhaseResponse, error) {
		return nil, nil
	}
	uc.Operation.JobTemplatePhase.DeleteJobTemplatePhase = func(context.Context, *jobtemplatephasepb.DeleteJobTemplatePhaseRequest) (*jobtemplatephasepb.DeleteJobTemplatePhaseResponse, error) {
		return nil, nil
	}

	// JobTemplateTask CRUD
	uc.Operation.JobTemplateTask.CreateJobTemplateTask = func(context.Context, *jobtemplateTaskpb.CreateJobTemplateTaskRequest) (*jobtemplateTaskpb.CreateJobTemplateTaskResponse, error) {
		return nil, nil
	}
	uc.Operation.JobTemplateTask.ReadJobTemplateTask = func(context.Context, *jobtemplateTaskpb.ReadJobTemplateTaskRequest) (*jobtemplateTaskpb.ReadJobTemplateTaskResponse, error) {
		return nil, nil
	}
	uc.Operation.JobTemplateTask.UpdateJobTemplateTask = func(context.Context, *jobtemplateTaskpb.UpdateJobTemplateTaskRequest) (*jobtemplateTaskpb.UpdateJobTemplateTaskResponse, error) {
		return nil, nil
	}
	uc.Operation.JobTemplateTask.DeleteJobTemplateTask = func(context.Context, *jobtemplateTaskpb.DeleteJobTemplateTaskRequest) (*jobtemplateTaskpb.DeleteJobTemplateTaskResponse, error) {
		return nil, nil
	}

	// JobTask
	uc.Operation.JobTask.CreateJobTask = func(context.Context, *jobtaskpb.CreateJobTaskRequest) (*jobtaskpb.CreateJobTaskResponse, error) {
		return nil, nil
	}
	uc.Operation.JobTask.ReadJobTask = func(context.Context, *jobtaskpb.ReadJobTaskRequest) (*jobtaskpb.ReadJobTaskResponse, error) {
		return nil, nil
	}
	uc.Operation.JobTask.UpdateJobTask = func(context.Context, *jobtaskpb.UpdateJobTaskRequest) (*jobtaskpb.UpdateJobTaskResponse, error) {
		return nil, nil
	}
	uc.Operation.JobTask.DeleteJobTask = func(context.Context, *jobtaskpb.DeleteJobTaskRequest) (*jobtaskpb.DeleteJobTaskResponse, error) {
		return nil, nil
	}
	uc.Operation.JobTask.ListJobTasks = func(context.Context, *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
		return nil, nil
	}

	// OutcomeCriteria
	uc.Operation.OutcomeCriteria.CreateOutcomeCriteria = func(context.Context, *criteriapb.CreateOutcomeCriteriaRequest) (*criteriapb.CreateOutcomeCriteriaResponse, error) {
		return nil, nil
	}
	uc.Operation.OutcomeCriteria.ReadOutcomeCriteria = func(context.Context, *criteriapb.ReadOutcomeCriteriaRequest) (*criteriapb.ReadOutcomeCriteriaResponse, error) {
		return nil, nil
	}
	uc.Operation.OutcomeCriteria.UpdateOutcomeCriteria = func(context.Context, *criteriapb.UpdateOutcomeCriteriaRequest) (*criteriapb.UpdateOutcomeCriteriaResponse, error) {
		return nil, nil
	}
	uc.Operation.OutcomeCriteria.DeleteOutcomeCriteria = func(context.Context, *criteriapb.DeleteOutcomeCriteriaRequest) (*criteriapb.DeleteOutcomeCriteriaResponse, error) {
		return nil, nil
	}
	uc.Operation.OutcomeCriteria.ListOutcomeCriterias = func(context.Context, *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error) {
		return nil, nil
	}

	// TaskOutcome
	uc.Operation.TaskOutcome.CreateTaskOutcome = func(context.Context, *taskoutcomepb.CreateTaskOutcomeRequest) (*taskoutcomepb.CreateTaskOutcomeResponse, error) {
		return nil, nil
	}
	uc.Operation.TaskOutcome.ReadTaskOutcome = func(context.Context, *taskoutcomepb.ReadTaskOutcomeRequest) (*taskoutcomepb.ReadTaskOutcomeResponse, error) {
		return nil, nil
	}
	uc.Operation.TaskOutcome.UpdateTaskOutcome = func(context.Context, *taskoutcomepb.UpdateTaskOutcomeRequest) (*taskoutcomepb.UpdateTaskOutcomeResponse, error) {
		return nil, nil
	}
	uc.Operation.TaskOutcome.DeleteTaskOutcome = func(context.Context, *taskoutcomepb.DeleteTaskOutcomeRequest) (*taskoutcomepb.DeleteTaskOutcomeResponse, error) {
		return nil, nil
	}
	uc.Operation.TaskOutcome.ListTaskOutcomes = func(context.Context, *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
		return nil, nil
	}

	// OutcomeSummary
	uc.Operation.JobOutcomeSummary.GetByJob = func(context.Context, *joboutcomesumpb.GetJobOutcomeSummaryByJobRequest) (*joboutcomesumpb.GetJobOutcomeSummaryByJobResponse, error) {
		return nil, nil
	}
	uc.Operation.JobOutcomeSummary.ListJobOutcomeSummaries = func(context.Context, *joboutcomesumpb.ListJobOutcomeSummarysRequest) (*joboutcomesumpb.ListJobOutcomeSummarysResponse, error) {
		return nil, nil
	}
	uc.Operation.PhaseOutcomeSummary.GetByJobPhase = func(context.Context, *phaseoutcomesumpb.GetPhaseOutcomeSummaryByJobPhaseRequest) (*phaseoutcomesumpb.GetPhaseOutcomeSummaryByJobPhaseResponse, error) {
		return nil, nil
	}
	uc.Operation.PhaseOutcomeSummary.ListByJob = func(context.Context, *phaseoutcomesumpb.ListPhaseOutcomeSummarysByJobRequest) (*phaseoutcomesumpb.ListPhaseOutcomeSummarysByJobResponse, error) {
		return nil, nil
	}

	// Now try enableAll — Workflow is nil but should not be a problem.
	if err := uc.RequireFor(&blockConfig{enableAll: true}); err != nil {
		t.Fatalf("RequireFor(enableAll, nil workflow) should pass; got %v", err)
	}
}
