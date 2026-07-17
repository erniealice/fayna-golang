package document

import (
	"context"
	"testing"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
)

// Formation-page (DOCX v2) builder tests. They pin the MMIS "Student Formation"
// contract at the DATA layer: one table per non-academic job_category (titled by
// the category NAME, ordered by sort_order — Subject Deportment before Homeroom
// Deportment), each strand shown with its FROZEN authoritative average
// (job_outcome_summary.scaled_label), and untaken strands (the transmute-of-zero
// floor, e.g. a Korean parallel track) suppressed via the shared enrollment
// predicate. NEVER blank a real deportment grade; NEVER render an untaken strand.

func strp(s string) *string { return &s }
func i32p(i int32) *int32   { return &i }

func catsFn(cats ...*jobcategorypb.JobCategory) func(context.Context, *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error) {
	return func(context.Context, *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error) {
		return &jobcategorypb.ListJobCategoriesResponse{Data: cats}, nil
	}
}

func templatesFn(t ...*jobtemplatepb.JobTemplate) func(context.Context, *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error) {
	return func(context.Context, *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error) {
		return &jobtemplatepb.ListJobTemplatesResponse{Data: t}, nil
	}
}

func summariesFn(s ...*jobsumpb.JobOutcomeSummary) func(context.Context, *jobsumpb.ListJobOutcomeSummarysRequest) (*jobsumpb.ListJobOutcomeSummarysResponse, error) {
	return func(context.Context, *jobsumpb.ListJobOutcomeSummarysRequest) (*jobsumpb.ListJobOutcomeSummarysResponse, error) {
		return &jobsumpb.ListJobOutcomeSummarysResponse{Data: s}, nil
	}
}

func job(id, tmpl, cat string) *jobpb.Job {
	return &jobpb.Job{Id: id, JobTemplateId: strp(tmpl), JobCategoryId: strp(cat)}
}

// depsForFormation wires the three closures collectFormationGroups consumes.
func depsForFormation() *Deps {
	return &Deps{
		ListJobCategories: catsFn(
			&jobcategorypb.JobCategory{Id: "cat-acad", Name: "Academic", SortOrder: i32p(10)},
			&jobcategorypb.JobCategory{Id: "cat-subj", Name: "Subject Deportment", SortOrder: i32p(20)},
			&jobcategorypb.JobCategory{Id: "cat-home", Name: "Homeroom Deportment", SortOrder: i32p(30)},
		),
		ListJobTemplates: templatesFn(
			&jobtemplatepb.JobTemplate{Id: "t-math", Name: "Mathematics — AY 2025-2026"},
			&jobtemplatepb.JobTemplate{Id: "t-sci", Name: "Sciences — AY 2025-2026"},
			&jobtemplatepb.JobTemplate{Id: "t-kor", Name: "Language and Literature: Korean — AY 2025-2026"},
			&jobtemplatepb.JobTemplate{Id: "t-home", Name: "Grade 8 Mercury — AY 2025-2026"},
		),
		ListJobOutcomeSummarys: summariesFn(
			&jobsumpb.JobOutcomeSummary{JobId: "j-math", Active: true, ScaledLabel: strp("98")},
			&jobsumpb.JobOutcomeSummary{JobId: "j-sci", Active: true, ScaledLabel: strp("96.5")},
			&jobsumpb.JobOutcomeSummary{JobId: "j-kor", Active: true, ScaledLabel: strp("0")}, // untaken → suppress
			&jobsumpb.JobOutcomeSummary{JobId: "j-home", Active: true, ScaledLabel: strp("95")},
		),
	}
}

func TestCollectFormationGroups_OrderSuppressAndAverage(t *testing.T) {
	d := depsForFormation()
	deportJobs := []*jobpb.Job{
		job("j-math", "t-math", "cat-subj"),
		job("j-sci", "t-sci", "cat-subj"),
		job("j-kor", "t-kor", "cat-subj"), // Korean, frozen avg 0 → suppressed
		job("j-home", "t-home", "cat-home"),
	}
	groups := collectFormationGroups(context.Background(), d, deportJobs, false)

	if len(groups) != 2 {
		t.Fatalf("want 2 groups (Subject Deportment, Homeroom Deportment), got %d: %+v", len(groups), groups)
	}
	// sort_order: Subject Deportment (20) before Homeroom Deportment (30).
	if groups[0].Title != "Subject Deportment" || groups[1].Title != "Homeroom Deportment" {
		t.Fatalf("group order wrong: %q, %q", groups[0].Title, groups[1].Title)
	}
	// Subject group: Korean suppressed; Mathematics + Sciences kept, subject ASC,
	// cleaned of the "— AY …" suffix, average = the frozen scaled_label verbatim.
	sub := groups[0].Rows
	if len(sub) != 2 {
		t.Fatalf("Subject Deportment: want 2 rows (Korean suppressed), got %d: %+v", len(sub), sub)
	}
	if sub[0].Subject != "Mathematics" || sub[0].Average != "98" {
		t.Fatalf("row0 want {Mathematics 98}, got %+v", sub[0])
	}
	if sub[1].Subject != "Sciences" || sub[1].Average != "96.5" {
		t.Fatalf("row1 want {Sciences 96.5}, got %+v", sub[1])
	}
	// Homeroom group: single row, cleaned name, frozen average.
	home := groups[1].Rows
	if len(home) != 1 || home[0].Subject != "Grade 8 Mercury" || home[0].Average != "95" {
		t.Fatalf("Homeroom Deportment want [{Grade 8 Mercury 95}], got %+v", home)
	}
}

func TestCollectFormationGroups_EmptyJobs_Nil(t *testing.T) {
	if g := collectFormationGroups(context.Background(), depsForFormation(), nil, false); g != nil {
		t.Fatalf("no deportment jobs must yield nil groups, got %+v", g)
	}
}

func TestCollectFormationGroups_AllUntaken_NoGroups(t *testing.T) {
	d := &Deps{
		ListJobCategories: catsFn(&jobcategorypb.JobCategory{Id: "cat-subj", Name: "Subject Deportment", SortOrder: i32p(20)}),
		ListJobTemplates:  templatesFn(&jobtemplatepb.JobTemplate{Id: "t-kor", Name: "Language and Literature: Korean — AY 2025-2026"}),
		ListJobOutcomeSummarys: summariesFn(
			&jobsumpb.JobOutcomeSummary{JobId: "j-kor", Active: true, ScaledLabel: strp("0")},
		),
	}
	// A strand with only the transmute-of-zero floor (0/1) or blank is fully
	// suppressed → no group emitted (not an empty-titled table).
	if g := collectFormationGroups(context.Background(), d, []*jobpb.Job{job("j-kor", "t-kor", "cat-subj")}, false); g != nil {
		t.Fatalf("all-untaken deportment must yield nil groups, got %+v", g)
	}
}

func TestCollectFormationGroups_FloorSuppressedRealKept(t *testing.T) {
	d := &Deps{
		ListJobCategories: catsFn(&jobcategorypb.JobCategory{Id: "cat-subj", Name: "Subject Deportment", SortOrder: i32p(20)}),
		ListJobTemplates: templatesFn(
			&jobtemplatepb.JobTemplate{Id: "t-floor", Name: "Floor"},
			&jobtemplatepb.JobTemplate{Id: "t-real", Name: "Real"},
		),
		ListJobOutcomeSummarys: summariesFn(
			&jobsumpb.JobOutcomeSummary{JobId: "j-floor", Active: true, ScaledLabel: strp("1")}, // floor → suppress
			&jobsumpb.JobOutcomeSummary{JobId: "j-real", Active: true, ScaledLabel: strp("2")},  // real >1 → keep
		),
	}
	groups := collectFormationGroups(context.Background(), d,
		[]*jobpb.Job{job("j-floor", "t-floor", "cat-subj"), job("j-real", "t-real", "cat-subj")}, false)
	if len(groups) != 1 || len(groups[0].Rows) != 1 || groups[0].Rows[0].Subject != "Real" {
		t.Fatalf("floor (avg 1) must suppress, real (avg 2) must keep; got %+v", groups)
	}
}

func TestFormationData_Shape(t *testing.T) {
	// Empty groups → empty slice (never a raw {{formation_groups}} leak; the engine
	// removes the loop block on an empty slice).
	if got := formationData(nil); got == nil || len(got) != 0 {
		t.Fatalf("nil groups must map to a non-nil empty slice, got %#v", got)
	}
	groups := []formationGroup{{Title: "Subject Deportment", Rows: []formationRow{{Subject: "Mathematics", Average: "98"}}}}
	data := formationData(groups)
	if len(data) != 1 {
		t.Fatalf("want 1 group item, got %d", len(data))
	}
	item := data[0].(map[string]any)
	if item["category_title"] != "Subject Deportment" {
		t.Fatalf("category_title = %v", item["category_title"])
	}
	rows := item["rows"].([]any)
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	r := rows[0].(map[string]any)
	if r["row_subject"] != "Mathematics" || r["row_average"] != "98" {
		t.Fatalf("row = %+v", r)
	}
}

// TestBuildReportCardData_IncludesFormation pins that the top-level data map always
// carries a formation_groups key (a subject-only card still emits an empty slice,
// so the template's {{#formation_groups}} loop resolves rather than leaking).
func TestBuildReportCardData_IncludesFormation(t *testing.T) {
	rc := reportCard{
		Subjects:        []itemRow{{Name: "Science", Sem1Band: "6", Sem2Band: "7", YearFinal: "7"}},
		FormationGroups: []formationGroup{{Title: "Subject Deportment", Rows: []formationRow{{Subject: "Mathematics", Average: "98"}}}},
	}
	m := buildReportCardData(rc)
	fg, ok := m["formation_groups"].([]any)
	if !ok {
		t.Fatalf("formation_groups must be []any, got %T", m["formation_groups"])
	}
	if len(fg) != 1 {
		t.Fatalf("want 1 formation group in the data map, got %d", len(fg))
	}
}
