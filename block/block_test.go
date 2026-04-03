package block

import (
	"testing"

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

func TestAssertUseCases_NilInput(t *testing.T) {
	t.Parallel()

	uc := assertUseCases(nil)
	if uc != nil {
		t.Fatal("assertUseCases(nil) should return nil")
	}
}

func TestAssertUseCases_NonStructInput(t *testing.T) {
	t.Parallel()

	// A string value is not a pointer-to-struct.
	uc := assertUseCases("not a struct")
	if uc != nil {
		t.Fatal("assertUseCases(string) should return nil")
	}
}

func TestAssertUseCases_IntInput(t *testing.T) {
	t.Parallel()

	uc := assertUseCases(42)
	if uc != nil {
		t.Fatal("assertUseCases(int) should return nil")
	}
}

func TestAssertUseCases_NilPointer(t *testing.T) {
	t.Parallel()

	var p *struct{ Name string }
	uc := assertUseCases(p)
	if uc != nil {
		t.Fatal("assertUseCases(nil pointer) should return nil")
	}
}

func TestAssertUseCases_ValidStruct(t *testing.T) {
	t.Parallel()

	s := struct{ Name string }{Name: "test"}
	uc := assertUseCases(&s)
	if uc == nil {
		t.Fatal("assertUseCases(valid struct pointer) should return non-nil")
	}
}

func TestAssertUseCases_SliceInput(t *testing.T) {
	t.Parallel()

	uc := assertUseCases([]string{"a", "b"})
	if uc != nil {
		t.Fatal("assertUseCases(slice) should return nil")
	}
}

func TestAssertUseCases_MapInput(t *testing.T) {
	t.Parallel()

	uc := assertUseCases(map[string]int{"a": 1})
	if uc != nil {
		t.Fatal("assertUseCases(map) should return nil")
	}
}
