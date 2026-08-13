package outcome_summary

import (
	"testing"

	bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/subscription_group_document_template"
)

// TestGroupValueRank pins the owner-locked band-order grammar: listed values
// lead in list order (case-insensitive, trimmed), unlisted report ok=false.
func TestGroupValueRank(t *testing.T) {
	o := RowOptions{GroupValueOrder: []string{"male", "female"}}

	cases := []struct {
		value    string
		wantRank int
		wantOK   bool
	}{
		{"male", 0, true},
		{"female", 1, true},
		{"MALE", 0, true},      // case-insensitive
		{"  Female ", 1, true}, // trimmed
		{"other", 0, false},    // unlisted
		{"", 0, false},         // no value
	}
	for _, c := range cases {
		rank, ok := o.GroupValueRank(c.value)
		if ok != c.wantOK || (ok && rank != c.wantRank) {
			t.Errorf("GroupValueRank(%q) = (%d,%v), want (%d,%v)", c.value, rank, ok, c.wantRank, c.wantOK)
		}
	}

	// Empty order → nothing is listed (fail-safe default = value-asc elsewhere).
	empty := RowOptions{}
	if _, ok := empty.GroupValueRank("male"); ok {
		t.Errorf("empty GroupValueOrder must report every value unlisted")
	}
}

// TestGateGrain pins the DocumentOptions.GateGrain contract: exactly two
// recognized values ("" = template grain, the const = group grain); anything
// else is a validation ERROR (the deliberate boot-error divergence from the
// sibling options' fail-safe ignore grammar — an integrity switch must not be
// typo-able into a different security posture).
func TestGateGrain(t *testing.T) {
	cases := []struct {
		name      string
		grain     string
		wantGroup bool
		wantValid bool
	}{
		{"zero_value_is_template_grain", "", false, true},
		{"const_is_group_grain", GateGrainSubscriptionGroup, true, true},
		{"const_literal", "subscription_group", true, true},
		{"whitespace_only_is_zero", "   ", false, true},
		{"padded_const_is_group_grain", "  subscription_group  ", true, true},
		{"garbage_is_invalid", "sideways", false, false},
		{"casing_is_invalid", "Subscription_Group", false, false},
		{"navigation_ref_is_invalid", "subscription_group.name", false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := DocumentOptions{GateGrain: c.grain}
			if got := d.GateGrainGroup(); got != c.wantGroup {
				t.Errorf("GateGrainGroup() = %v, want %v", got, c.wantGroup)
			}
			err := d.ValidateGateGrain()
			if c.wantValid && err != nil {
				t.Errorf("ValidateGateGrain() = %v, want nil", err)
			}
			if !c.wantValid && err == nil {
				t.Errorf("ValidateGateGrain() = nil, want error (unknown grain must be a boot error, never ignored)")
			}
		})
	}

	// An invalid grain must NEVER read as group grain (no runtime guess).
	if (DocumentOptions{GateGrain: "sideways"}).GateGrainGroup() {
		t.Errorf("an unrecognized grain must not report group grain")
	}
}

func TestSectionExportEnabledIndependentFromGroupedPresentation(t *testing.T) {
	for _, tc := range []struct {
		name          string
		entity        string
		exportEnabled bool
		wantGrouped   bool
	}{
		{name: "zero"},
		{name: "grouped only", entity: ListEntitySubscriptionGroup, wantGrouped: true},
		{name: "export only", exportEnabled: true},
		{name: "grouped and export", entity: ListEntitySubscriptionGroup, exportEnabled: true, wantGrouped: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			options := Options{List: ListOptions{Entity: tc.entity}, SectionExport: SectionExportOptions{Enabled: tc.exportEnabled}}
			if got := options.SectionExportEnabled(); got != tc.exportEnabled {
				t.Fatalf("SectionExportEnabled() = %v, want %v", got, tc.exportEnabled)
			}
			if got := options.List.SubscriptionGroups(); got != tc.wantGrouped {
				t.Fatalf("SubscriptionGroups() = %v, want %v", got, tc.wantGrouped)
			}
		})
	}
}

func TestSectionExportProfileAndRowBandConfiguration(t *testing.T) {
	const profile = bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1
	opts := Options{
		Row: RowOptions{GroupByField: "client_attributes.gender"},
		SectionExport: SectionExportOptions{
			ProfileByCategoryCode:  map[string]bindingpb.RenderProfile{"academic": profile, "bad": 99},
			GroupByAttributeModule: " entity ",
		},
	}
	if got, ok := opts.SectionExport.ProfileForCategoryCode("academic"); !ok || got != profile {
		t.Fatalf("trusted profile lookup = (%v,%v), want (%v,true)", got, ok, profile)
	}
	for _, code := range []string{"missing", "bad"} {
		if got, ok := opts.SectionExport.ProfileForCategoryCode(code); ok || got != bindingpb.RenderProfile_RENDER_PROFILE_UNSPECIFIED {
			t.Fatalf("profile lookup %q = (%v,%v), want fail-closed UNSPECIFIED", code, got, ok)
		}
	}
	code, module, configured, err := opts.ExportRowBandConfig()
	if err != nil || !configured || code != "gender" || module != "entity" {
		t.Fatalf("ExportRowBandConfig() = (%q,%q,%v,%v)", code, module, configured, err)
	}

	for _, tc := range []struct {
		name           string
		field          string
		module         string
		wantConfigured bool
		wantErr        bool
	}{
		{name: "empty disables", field: "", module: "", wantConfigured: false},
		{name: "foreign reference", field: "job_category", module: "entity", wantConfigured: true, wantErr: true},
		{name: "missing module", field: "client_attributes.gender", module: "", wantConfigured: true, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candidate := Options{Row: RowOptions{GroupByField: tc.field}, SectionExport: SectionExportOptions{GroupByAttributeModule: tc.module}}
			_, _, configured, err := candidate.ExportRowBandConfig()
			if configured != tc.wantConfigured || (err != nil) != tc.wantErr {
				t.Fatalf("configured/error = %v/%v, want %v/%v", configured, err, tc.wantConfigured, tc.wantErr)
			}
		})
	}
}
