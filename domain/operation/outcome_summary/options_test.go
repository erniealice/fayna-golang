package outcome_summary

import "testing"

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
