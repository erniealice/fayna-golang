package block

import (
	"context"
	"strings"
	"testing"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplateTaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	templatetaskcriteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
	pyeza "github.com/erniealice/pyeza-golang"
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

	appCtx := &pyeza.AppContext{
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

	appCtx := &pyeza.AppContext{
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

	appCtx := &pyeza.AppContext{
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
