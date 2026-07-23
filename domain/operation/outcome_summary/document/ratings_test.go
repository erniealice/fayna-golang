package document

import (
	"sort"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
)

// The rotation-pair merge: the period-1 strand (identified by its period-1
// conduct summary) titles first — "Arts: Visual Arts / Arts: Music".
func TestMergeRotationPairsPeriodOrder(t *testing.T) {
	c := &ratingContext{
		nameOf: map[string]string{
			"jva": "Arts: Visual Arts",
			"jmu": "Arts: Music",
		},
		pos: map[string]map[int32]string{
			"jva": {1: "100"},
			"jmu": {2: "98"},
		},
		avg: map[string]string{"jva": "100", "jmu": "98"},
	}
	c.strandJobs = jobsFromNames(c.nameOf)

	m := mergeRotationPairs(c, map[string]bool{"arts": true}, nil)
	got := m.titleFor("Arts")
	want := "Arts: Visual Arts / Arts: Music"
	if got != want {
		t.Fatalf("titleFor(Arts) = %q, want %q", got, want)
	}
}

// Without conduct summaries the INACTIVE-academic-strand fallback decides the
// period-2 strand.
func TestMergeRotationPairsInactiveFallback(t *testing.T) {
	c := &ratingContext{
		nameOf: map[string]string{
			"jtle": "Design: TLE",
			"jcom": "Design: Computer",
		},
		pos: map[string]map[int32]string{},
		avg: map[string]string{"jtle": "100", "jcom": "97"},
	}
	c.strandJobs = jobsFromNames(c.nameOf)

	m := mergeRotationPairs(c, map[string]bool{"design": true},
		map[string]bool{"design: computer": true})
	got := m.titleFor("Design")
	want := "Design: TLE / Design: Computer"
	if got != want {
		t.Fatalf("titleFor(Design) = %q, want %q", got, want)
	}
}

// A prefix that is NOT an academic subject never merges; a lone strand never
// merges.
func TestMergeRotationPairsEligibility(t *testing.T) {
	c := &ratingContext{
		nameOf: map[string]string{
			"j1": "Arts: Visual Arts",
			"j2": "Arts: Music",
			"j3": "Mathematics",
		},
		pos: map[string]map[int32]string{},
		avg: map[string]string{},
	}
	c.strandJobs = jobsFromNames(c.nameOf)

	m := mergeRotationPairs(c, map[string]bool{"mathematics": true}, nil)
	if got := m.titleFor("Arts"); got != "" {
		t.Fatalf("non-academic prefix merged: %q", got)
	}
	if got := m.titleFor("Mathematics"); got != "" {
		t.Fatalf("lone subject merged: %q", got)
	}
}

// Conduct rows: the merged pair reads each strand's own period column; the
// non-enrolled strand side stays blank; unpaired strands keep their per-period
// summaries; rows sort by title.
func TestBuildItemRatings(t *testing.T) {
	c := &ratingContext{
		nameOf: map[string]string{
			"jva":  "Arts: Visual Arts",
			"jmu":  "Arts: Music",
			"jsci": "Sciences",
			"jkor": "Language and Literature: Korean",
			"jfil": "Language Acquisition: Filipino",
		},
		pos: map[string]map[int32]string{
			"jva":  {1: "100"},
			"jmu":  {2: "98"},
			"jsci": {1: "100", 2: "80"},
			"jkor": {1: "100", 2: "100"},
			// jfil non-enrolled: no summaries
		},
		avg: map[string]string{
			"jva": "100", "jmu": "98", "jsci": "90", "jkor": "100",
			"jfil": "0", // transmute-of-zero floor → suppressed
		},
	}
	c.strandJobs = jobsFromNames(c.nameOf)

	m := mergedPairs{byCanonical: map[string]rotationPair{
		"arts": {sem1Name: "Arts: Visual Arts", sem2Name: "Arts: Music", sem1Job: "jva", sem2Job: "jmu"},
	}}
	rows, _, _ := buildItemRatings(c, m)

	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3 (merged pair + Sciences + Korean; Filipino suppressed)", len(rows))
	}
	if rows[0].Title != "Arts: Visual Arts / Arts: Music" || rows[0].Phase1 != "100" || rows[0].Phase2 != "98" {
		t.Fatalf("merged row = %+v", rows[0])
	}
	if rows[1].Title != "Language and Literature: Korean" {
		t.Fatalf("sort order: rows[1] = %+v", rows[1])
	}
	if rows[2].Title != "Sciences" || rows[2].Phase1 != "100" || rows[2].Phase2 != "80" {
		t.Fatalf("Sciences row = %+v", rows[2])
	}
}

// The per-period criterion rows letter by sequence order and total per period;
// the year collapse (v1 keys) takes the MAX across periods.
func TestTranscriptCriterionRowsAndYearCollapse(t *testing.T) {
	tr := &transcript{
		marks: map[string]map[int32]float64{
			"oc-investigating":       {1: 7, 2: 4},
			"oc-developing":          {1: 4, 2: 3},
			"oc-creating-performing": {1: 5, 2: 4},
			"oc-evaluating":          {1: 7, 2: 3},
		},
		seq: map[string]int32{
			"oc-investigating": 1, "oc-developing": 2,
			"oc-creating-performing": 3, "oc-evaluating": 4,
		},
	}
	names := map[string]string{
		"oc-investigating":       "Investigating",
		"oc-developing":          "Developing",
		"oc-creating-performing": "Creating/performing",
		"oc-evaluating":          "Evaluating",
	}
	rows, s1, s2 := tr.criterionRows(names)
	if len(rows) != 4 {
		t.Fatalf("rows = %d", len(rows))
	}
	if rows[0].Label != "A - Investigating" || rows[0].Phase1 != "7" || rows[0].Phase2 != "4" {
		t.Fatalf("row A = %+v", rows[0])
	}
	if rows[2].Label != "C - Creating/performing" {
		t.Fatalf("row C = %+v", rows[2])
	}
	if s1 != "23" || s2 != "14" {
		t.Fatalf("totals = %s/%s, want 23/14", s1, s2)
	}
	yc, ok := tr.yearCriteria()
	if !ok || yc.a != "7" || yc.b != "4" || yc.c != "5" || yc.d != "7" || yc.total != "23" {
		t.Fatalf("year collapse = %+v ok=%v", yc, ok)
	}
}

// The teacher line: rotation pair (tie in period 2 prefers the non-period-1
// assignee) vs single-teacher subject.
func TestStaffLine(t *testing.T) {
	labels := outcome_summary.PeriodLabels{StaffLabel: "Teacher:", StaffPluralLabel: "Teachers:"}
	names := map[string]string{"sP": "Alexis Purisima", "sC": "Darianne Cabornay"}

	pair := &transcript{teachers: map[int32]map[string]int{
		1: {"sP": 1},
		2: {"sP": 1, "sC": 1}, // tie → prefer the non-period-1 assignee
	}}
	// classFallbackID "" — a per-task assignee is present, so the class-edge
	// fallback is never consulted (the assignee override wins).
	if got := staffLine(labels, pair, names, ""); got != "Teachers: Alexis Purisima / Darianne Cabornay" {
		t.Fatalf("pair line = %q", got)
	}

	single := &transcript{teachers: map[int32]map[string]int{
		1: {"sP": 1}, 2: {"sP": 1},
	}}
	if got := staffLine(labels, single, names, ""); got != "Teacher: Alexis Purisima" {
		t.Fatalf("single line = %q", got)
	}

	if got := staffLine(labels, nil, names, ""); got != "" {
		t.Fatalf("nil transcript line = %q", got)
	}

	// D5 COALESCE: with NO per-task assignee, the servicer is DERIVED from the
	// class edge (classFallbackID resolved via names).
	if got := staffLine(labels, nil, names, "sC"); got != "Teacher: Darianne Cabornay" {
		t.Fatalf("class-edge fallback line = %q", got)
	}
	// The per-task assignee override still wins over any class-edge fallback.
	if got := staffLine(labels, single, names, "sC"); got != "Teacher: Alexis Purisima" {
		t.Fatalf("override-wins line = %q", got)
	}
}

// jobsFromNames builds the minimal strand-job slice for ratingContext tests
// (deterministic id order; only GetId() is consumed by the code under test).
func jobsFromNames(names map[string]string) []*jobpb.Job {
	ids := make([]string, 0, len(names))
	for id := range names {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	jobs := make([]*jobpb.Job, 0, len(ids))
	for _, id := range ids {
		jobs = append(jobs, &jobpb.Job{Id: id})
	}
	return jobs
}
