package client_card

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
)

// R9 W-A6 — the per-student client card's job-category bands (dedicated
// Options.ClientCard). These pin: banding-on (ordered bands, rows under the
// right band), the <2-category degrade → flat byte-equal, the Uncategorized
// band, H2 preservation on the non-lift path, and the CSS-inert bulk-state of
// the band header (no data-bulk-enabled on a bulk-less table).

func i32ptr(v int32) *int32 { return &v }

// categoriesFn returns the same category set for EVERY request — it ignores the
// filter, so it serves both fetchBandCategories (no filter → all) and
// ResolveCategoryID (code filter → the loop finds the matching Code).
func categoriesFn(cats ...*jobcategorypb.JobCategory) func(context.Context, *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error) {
	return func(context.Context, *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error) {
		return &jobcategorypb.ListJobCategoriesResponse{Data: cats}, nil
	}
}

// job builds a subscription-origin active job for sub1 with the given template.
func job(id, tmpl string) *jobpb.Job {
	return &jobpb.Job{Id: id, JobTemplateId: sptr(tmpl), OriginId: sptr("sub1"), OriginType: enums.OriginType_ORIGIN_TYPE_SUBSCRIPTION, Active: true}
}

// bandTitles returns the group titles in render order.
func bandTitles(table *types.TableConfig) []string {
	out := make([]string, 0, len(table.Groups))
	for _, g := range table.Groups {
		out = append(out, g.Title)
	}
	return out
}

// groupByTitle returns the group with the given title (or fails).
func groupByTitle(t *testing.T, table *types.TableConfig, title string) types.TableRowGroup {
	t.Helper()
	for _, g := range table.Groups {
		if g.Title == title {
			return g
		}
	}
	t.Fatalf("band %q not found; have %v", title, bandTitles(table))
	return types.TableRowGroup{}
}

// groupHasSubject reports whether a group holds a row whose subject-name cell
// equals name.
func groupHasSubject(g types.TableRowGroup, name string) bool {
	for _, r := range g.Rows {
		if len(r.Cells) > 0 && r.Cells[0].Value == name {
			return true
		}
	}
	return false
}

// threeCategories: academic(10) < subject_deportment(20) < homeroom_deportment(30).
func threeCategories() []*jobcategorypb.JobCategory {
	return []*jobcategorypb.JobCategory{
		// Deliberately supplied OUT of sort order to prove sortBandCategories.
		{Id: "cat-home", Name: "Homeroom", Code: sptr("homeroom_deportment"), SortOrder: i32ptr(30), Active: true},
		{Id: "cat-acad", Name: "Academic", Code: sptr("academic"), SortOrder: i32ptr(10), Active: true},
		{Id: "cat-cond", Name: "Conduct", Code: sptr("subject_deportment"), SortOrder: i32ptr(20), Active: true},
	}
}

// bandDeps wires a banding-enabled card: jobs + templates (with category FKs) +
// year grades (so the table is non-nil) + the category read.
func bandDeps(jobs []*jobpb.Job, tmpls []*jobtemplatepb.JobTemplate, cats []*jobcategorypb.JobCategory) *Deps {
	years := make([]*jobsumpb.JobOutcomeSummary, 0, len(jobs))
	for _, j := range jobs {
		years = append(years, &jobsumpb.JobOutcomeSummary{JobId: j.GetId(), ScaledLabel: sptr("6"), Active: true})
	}
	return &Deps{
		BandByCategory:         true,
		IncludeAllCategories:   true,
		ListJobCategories:      categoriesFn(cats...),
		ListJobs:               jobsFn(jobs...),
		ListJobTemplates:       templatesFn(tmpls...),
		ListJobOutcomeSummarys: yearSummariesFn(years...),
	}
}

// TestBuildTable_Banded_ThreeBands_Ordered: banding on with subjects across all
// three categories → three TableRowGroup bands, ordered by sort_order, each
// holding its own subjects; flat Rows is empty.
func TestBuildTable_Banded_ThreeBands_Ordered(t *testing.T) {
	jobs := []*jobpb.Job{
		job("jEng", "tEng"),   // academic
		job("jKor", "tKor"),   // academic
		job("jCon", "tCon"),   // subject_deportment
		job("jHome", "tHome"), // homeroom_deportment
	}
	tmpls := []*jobtemplatepb.JobTemplate{
		{Id: "tEng", Name: "English", JobCategoryId: sptr("cat-acad")},
		{Id: "tKor", Name: "Korean", JobCategoryId: sptr("cat-acad")},
		{Id: "tCon", Name: "Conduct", JobCategoryId: sptr("cat-cond")},
		{Id: "tHome", Name: "Homeroom", JobCategoryId: sptr("cat-home")},
	}
	table := buildTable(context.Background(), bandDeps(jobs, tmpls, threeCategories()), "sub1", false)
	if table == nil {
		t.Fatal("buildTable returned nil")
	}
	if len(table.Rows) != 0 {
		t.Errorf("banded table must use Groups, not flat Rows; got %d rows", len(table.Rows))
	}
	if got := bandTitles(table); !reflect.DeepEqual(got, []string{"Academic", "Conduct", "Homeroom"}) {
		t.Fatalf("band order = %v, want [Academic Conduct Homeroom] (sort_order 10/20/30)", got)
	}
	// Rows land under the correct band.
	acad := groupByTitle(t, table, "Academic")
	if !groupHasSubject(acad, "English") || !groupHasSubject(acad, "Korean") {
		t.Errorf("Academic band must hold English + Korean; got %d rows", len(acad.Rows))
	}
	if !groupHasSubject(groupByTitle(t, table, "Conduct"), "Conduct") {
		t.Error("Conduct band must hold the Conduct subject")
	}
	if !groupHasSubject(groupByTitle(t, table, "Homeroom"), "Homeroom") {
		t.Error("Homeroom band must hold the Homeroom subject")
	}
	// Collision-proof band ID/testid via the category CODE.
	if acad.ID != "rc-band-academic" || acad.DataAttrs["testid"] != "rc-band-academic" {
		t.Errorf("Academic band ID/testid = %q/%q, want rc-band-academic", acad.ID, acad.DataAttrs["testid"])
	}
}

// TestBuildTable_DegradeFewerThanTwoCategories_FlatByteEqual: banding on but the
// student has subjects in only ONE category → degrade to flat rows, byte-equal
// to a plainly-unbanded render of the same jobs.
func TestBuildTable_DegradeFewerThanTwoCategories_FlatByteEqual(t *testing.T) {
	jobs := []*jobpb.Job{job("jEng", "tEng"), job("jKor", "tKor")}
	tmpls := []*jobtemplatepb.JobTemplate{
		{Id: "tEng", Name: "English", JobCategoryId: sptr("cat-acad")},
		{Id: "tKor", Name: "Korean", JobCategoryId: sptr("cat-acad")},
	}

	banded := buildTable(context.Background(), bandDeps(jobs, tmpls, threeCategories()), "sub1", false)
	if banded == nil {
		t.Fatal("banded buildTable returned nil")
	}
	if len(banded.Groups) != 0 {
		t.Errorf("single-category card must degrade to flat rows, got %d bands", len(banded.Groups))
	}

	// A plainly-unbanded card (no banding config, no H2 filter) over the SAME jobs.
	flatDeps := &Deps{
		ListJobs:               jobsFn(jobs...),
		ListJobTemplates:       templatesFn(tmpls...),
		ListJobOutcomeSummarys: yearSummariesFn(&jobsumpb.JobOutcomeSummary{JobId: "jEng", ScaledLabel: sptr("6"), Active: true}, &jobsumpb.JobOutcomeSummary{JobId: "jKor", ScaledLabel: sptr("6"), Active: true}),
	}
	flat := buildTable(context.Background(), flatDeps, "sub1", false)
	if flat == nil {
		t.Fatal("flat buildTable returned nil")
	}
	if !reflect.DeepEqual(banded.Rows, flat.Rows) {
		t.Errorf("degrade path must be byte-equal to the unbanded flat rows.\nbanded=%+v\nflat=%+v", banded.Rows, flat.Rows)
	}
}

// TestBuildTable_Uncategorized_Band: a subject whose template category is NULL
// (and one whose category is foreign/out-of-corpus) folds into a single
// Uncategorized band alongside the academic band — never dropped, never
// duplicated.
func TestBuildTable_Uncategorized_Band(t *testing.T) {
	jobs := []*jobpb.Job{
		job("jEng", "tEng"),     // academic
		job("jNull", "tNull"),   // NULL category
		job("jFor", "tForeign"), // category id not in the corpus
	}
	tmpls := []*jobtemplatepb.JobTemplate{
		{Id: "tEng", Name: "English", JobCategoryId: sptr("cat-acad")},
		{Id: "tNull", Name: "Mystery"}, // JobCategoryId nil → ""
		{Id: "tForeign", Name: "Ghost", JobCategoryId: sptr("cat-gone")},
	}
	cats := []*jobcategorypb.JobCategory{
		{Id: "cat-acad", Name: "Academic", Code: sptr("academic"), SortOrder: i32ptr(10), Active: true},
	}
	table := buildTable(context.Background(), bandDeps(jobs, tmpls, cats), "sub1", false)
	if table == nil {
		t.Fatal("buildTable returned nil")
	}
	if got := bandTitles(table); !reflect.DeepEqual(got, []string{"Academic", "Uncategorized"}) {
		t.Fatalf("bands = %v, want [Academic Uncategorized]", got)
	}
	un := groupByTitle(t, table, "Uncategorized")
	if !groupHasSubject(un, "Mystery") || !groupHasSubject(un, "Ghost") {
		t.Errorf("Uncategorized band must fold both the NULL and the foreign subject; got %d rows", len(un.Rows))
	}
	if un.ID != "rc-band-uncategorized" {
		t.Errorf("Uncategorized band ID = %q, want rc-band-uncategorized", un.ID)
	}
	// Never duplicated: exactly one subject per band row total = 3.
	total := len(groupByTitle(t, table, "Academic").Rows) + len(un.Rows)
	if total != 3 {
		t.Errorf("every subject appears exactly once; total banded rows = %d, want 3", total)
	}
}

// TestBuildTable_H2Preserved_WhenNotBanding: with the card NOT lifting H2 (the
// default / no IncludeAllCategories) the academic-only CategoryFilter still
// drops same-origin deportment jobs — deportment appears ONLY when the card
// opts into IncludeAllCategories. (The report-card DOCUMENT download keeps its
// own independent H2 filter — document/data.go — unaffected by this card.)
func TestBuildTable_H2Preserved_WhenNotBanding(t *testing.T) {
	jobs := []*jobpb.Job{
		{Id: "jAcad", JobTemplateId: sptr("tEng"), OriginId: sptr("sub1"), OriginType: enums.OriginType_ORIGIN_TYPE_SUBSCRIPTION, Active: true, JobCategoryId: sptr("cat-acad")},
		{Id: "jDep", JobTemplateId: sptr("tCon"), OriginId: sptr("sub1"), OriginType: enums.OriginType_ORIGIN_TYPE_SUBSCRIPTION, Active: true, JobCategoryId: sptr("cat-cond")},
	}
	deps := &Deps{
		// Banding NOT lifted: BandByCategory false → H2 applies.
		CategoryFilter:         "academic",
		ListJobCategories:      categoriesFn(&jobcategorypb.JobCategory{Id: "cat-acad", Name: "Academic", Code: sptr("academic"), Active: true}),
		ListJobs:               jobsFn(jobs...),
		ListJobTemplates:       templatesFn(&jobtemplatepb.JobTemplate{Id: "tEng", Name: "English", JobCategoryId: sptr("cat-acad")}, &jobtemplatepb.JobTemplate{Id: "tCon", Name: "Conduct", JobCategoryId: sptr("cat-cond")}),
		ListJobOutcomeSummarys: yearSummariesFn(&jobsumpb.JobOutcomeSummary{JobId: "jAcad", ScaledLabel: sptr("6"), Active: true}, &jobsumpb.JobOutcomeSummary{JobId: "jDep", ScaledLabel: sptr("6"), Active: true}),
	}
	table := buildTable(context.Background(), deps, "sub1", false)
	if table == nil {
		t.Fatal("buildTable returned nil")
	}
	if len(table.Groups) != 0 {
		t.Errorf("H2 (non-lift) card must stay flat, got %d bands", len(table.Groups))
	}
	if len(table.Rows) != 1 || table.Rows[0].Cells[0].Value != "English" {
		t.Errorf("H2 must drop the deportment subject, keeping only English; got %d rows", len(table.Rows))
	}
}

// renderTableCard parses the shared pyeza template tree and executes the
// "table-card" template — the real app render path — returning the HTML.
func renderTableCard(t *testing.T, cfg types.TableConfig) string {
	t.Helper()
	r := pyeza.NewHTMLRendererFromFS(pyeza.SharedFS)
	if err := r.Init(); err != nil {
		t.Fatalf("init shared templates: %v", err)
	}
	var buf bytes.Buffer
	// Pass a POINTER — RowCount() (used by the footer) has a pointer receiver, so
	// a value config is missing it from its template method set. Apps render
	// .Table (a *TableConfig), so this matches the real path.
	if err := r.GetTemplates().ExecuteTemplate(&buf, "table-card", &cfg); err != nil {
		t.Fatalf("execute table-card: %v", err)
	}
	return buf.String()
}

// TestBandedTable_BulkControlsInert is the CSS-inert bulk-state regression
// (Q-R9-8 / plan §3.9): the client-card table sets NO BulkActions, so the card
// wrapper emits no data-bulk-enabled="true" (table.html:74). The band header's
// bulk-select controls are still in the markup but hidden by CSS
// (table.css:3427/3456 — .table-card:not([data-bulk-enabled="true"])
// .table-group-selection{display:none}), so they render INERTLY, not absent.
func TestBandedTable_BulkControlsInert(t *testing.T) {
	cfg := types.TableConfig{
		ID: "report-cards-student",
		Groups: []types.TableRowGroup{
			{ID: "rc-band-academic", Title: "Academic", DataAttrs: map[string]string{"testid": "rc-band-academic"}, Rows: []types.TableRow{{ID: "j1", Cells: []types.TableCell{{Value: "English"}}}}},
			{ID: "rc-band-uncategorized", Title: "Uncategorized", DataAttrs: map[string]string{"testid": "rc-band-uncategorized"}, Rows: []types.TableRow{{ID: "j2", Cells: []types.TableCell{{Value: "Mystery"}}}}},
		},
		// BulkActions deliberately nil (the client card never bulk-selects).
	}
	out := renderTableCard(t, cfg)

	// 1. No bulk-enabled attribute → the CSS :not() rules hide the controls.
	if strings.Contains(out, `data-bulk-enabled="true"`) {
		t.Errorf("bulk-less banded card must NOT emit data-bulk-enabled; got: %s", out)
	}
	// 2. The bands DID render (band header + titles present).
	if !strings.Contains(out, "table-group-header") {
		t.Errorf("band header must render; got: %s", out)
	}
	if !strings.Contains(out, ">Academic<") || !strings.Contains(out, ">Uncategorized<") {
		t.Errorf("both band titles must render; got: %s", out)
	}
	// 3. The selection controls are PRESENT in the markup (inert-by-CSS, not
	//    inert-by-absence) — proving the regression is guarded by the CSS rule,
	//    not by the header lacking the controls.
	if !strings.Contains(out, "table-group-selection") {
		t.Errorf("group-selection controls must be present (CSS-hidden), proving inert-by-CSS; got: %s", out)
	}
}
