package outcome_matrix

import "testing"

// TestRowGroupValueRank pins the grade-sheet band-order grammar (symmetry with
// outcome_summary.RowOptions.GroupValueRank): listed values lead in list
// order (case-insensitive, trimmed); unlisted report ok=false; empty order =
// nothing listed (fail-safe → value-asc elsewhere).
func TestRowGroupValueRank(t *testing.T) {
	o := Options{RowGroupValueOrder: []string{"male", "female"}}
	cases := []struct {
		value    string
		wantRank int
		wantOK   bool
	}{
		{"male", 0, true},
		{"female", 1, true},
		{"MALE", 0, true},
		{" female ", 1, true},
		{"nonbinary", 0, false},
		{"", 0, false},
	}
	for _, c := range cases {
		rank, ok := o.RowGroupValueRank(c.value)
		if ok != c.wantOK || (ok && rank != c.wantRank) {
			t.Errorf("RowGroupValueRank(%q) = (%d,%v), want (%d,%v)", c.value, rank, ok, c.wantRank, c.wantOK)
		}
	}
	if _, ok := (Options{}).RowGroupValueRank("male"); ok {
		t.Errorf("empty RowGroupValueOrder must report every value unlisted")
	}
}
