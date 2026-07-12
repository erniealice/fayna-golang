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
