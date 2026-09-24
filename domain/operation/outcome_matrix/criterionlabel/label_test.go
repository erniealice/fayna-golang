package criterionlabel

import "testing"

func TestCriterionLabelLetterFromActivityOrder(t *testing.T) {
	for i, want := range []string{"Objective A: Knowing", "Objective B: Investigating", "Objective C: Applying"} {
		if got := Label([]string{"Knowing", "Investigating", "Applying"}[i], i, "Objective {letter}: ", ""); got != want {
			t.Errorf("index %d: got %q, want %q", i, got, want)
		}
	}
	if got := Label("Beyond", 26, "Objective {letter}: ", ""); got != "Beyond" {
		t.Errorf("out-of-range index: %q", got)
	}
}

func TestCriterionLabelSharedCriterionDifferentPositions(t *testing.T) {
	if a, b := Label("Investigating", 0, "Objective {letter}: ", ""), Label("Investigating", 1, "Objective {letter}: ", ""); a != "Objective A: Investigating" || b != "Objective B: Investigating" {
		t.Fatalf("shared criterion labels: %q, %q", a, b)
	}
}

func TestCriterionLabelEmptyPrefixIsOff(t *testing.T) {
	if got := Label("Investigating", 1, "", ""); got != "Investigating" {
		t.Fatalf("empty prefix: %q", got)
	}
}

func TestCriterionLabelBreakAfter(t *testing.T) {
	for _, tc := range []struct{ name, pattern, chars, want string }{
		{"Knowing and Understanding", "Objective {letter}: ", ":", "Objective A:\nKnowing and Understanding"},
		{"One / Two", "", "/", "One /\nTwo"},
		{"No match", "", ":", "No match"},
		{"A: B / C", "", ":/", "A:\nB / C"},
	} {
		if got := Label(tc.name, 0, tc.pattern, tc.chars); got != tc.want {
			t.Errorf("Label(%q, %q, %q) = %q, want %q", tc.name, tc.pattern, tc.chars, got, tc.want)
		}
	}
}
