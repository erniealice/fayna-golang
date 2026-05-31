package block

import (
	"context"
	"testing"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
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
