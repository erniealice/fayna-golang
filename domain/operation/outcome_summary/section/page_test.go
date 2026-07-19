package section

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	outcome_summary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/types"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
)

// --- ?jc= category-tab test helpers (R9 W-A3) ---------------------------------

func strp(s string) *string { return &s }
func i32p(v int32) *int32   { return &v }

// jcat builds an ACTIVE job_category with a code + sort_order.
func jcat(id, code, name string, order int32) *jobcategorypb.JobCategory {
	return &jobcategorypb.JobCategory{Id: id, Name: name, Active: true, Code: strp(code), SortOrder: i32p(order)}
}

// jtmpl builds a job_template with an optional (nullable) category FK.
func jtmpl(id, name, categoryID string) *jobtemplatepb.JobTemplate {
	t := &jobtemplatepb.JobTemplate{Id: id, Name: name}
	if categoryID != "" {
		t.JobCategoryId = strp(categoryID)
	}
	return t
}

// jjob builds a job with a template id + a (nullable) frozen category snapshot.
func jjob(id, templateID, categorySnapshot, clientID string) *jobpb.Job {
	j := &jobpb.Job{Id: id, Active: true}
	if templateID != "" {
		j.JobTemplateId = strp(templateID)
	}
	if categorySnapshot != "" {
		j.JobCategoryId = strp(categorySnapshot)
	}
	if clientID != "" {
		j.ClientId = strp(clientID)
	}
	return j
}

func catListStub(cats []*jobcategorypb.JobCategory, err error) func(context.Context, *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error) {
	return func(context.Context, *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error) {
		if err != nil {
			return nil, err
		}
		return &jobcategorypb.ListJobCategoriesResponse{Data: cats}, nil
	}
}

func tmplListStub(tmpls []*jobtemplatepb.JobTemplate) func(context.Context, *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error) {
	return func(context.Context, *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error) {
		return &jobtemplatepb.ListJobTemplatesResponse{Data: tmpls}, nil
	}
}

// row builds a minimal TableRow whose first cell is the (unnumbered) name.
func row(clientID, name string) types.TableRow {
	return types.TableRow{ID: clientID, Cells: []types.TableCell{{Value: name}}}
}

// TestApplyRowPresentation_BandOrderAndNumbering pins two owner-locked
// contracts on the section grid: (1) gender bands order per
// Options.Row.GroupValueOrder ["male","female"] — MALE band first regardless
// of value-alpha (which would put female first); (2) row numbering is
// CONTINUOUS across bands (male 1..N, female N+1..M), applied after banding.
func TestApplyRowPresentation_BandOrderAndNumbering(t *testing.T) {
	students := map[string]student{
		"f1": {clientID: "f1", name: "Ann Fem", lastName: "Fem", firstName: "Ann"},
		"m1": {clientID: "m1", name: "Bob Male", lastName: "Male", firstName: "Bob"},
		"m2": {clientID: "m2", name: "Al Male", lastName: "Aale", firstName: "Al"},
	}
	rows := []types.TableRow{row("f1", "Fem, Ann"), row("m1", "Male, Bob"), row("m2", "Aale, Al")}
	attrValues := map[string]map[string]string{
		"gender": {"f1": "female", "m1": "male", "m2": "male"},
	}
	opts := outcome_summary.Options{Row: outcome_summary.RowOptions{
		GroupByField:    "client_attributes.gender",
		GroupValueOrder: []string{"male", "female"},
		SortField:       "last_name",
	}}

	table := &types.TableConfig{}
	applyRowPresentation(table, rows, students, attrValues, opts)
	numberRows(table)

	if len(table.Groups) != 2 {
		t.Fatalf("want 2 bands, got %d", len(table.Groups))
	}
	if strings.ToLower(table.Groups[0].Title) != "male" {
		t.Errorf("band 0 = %q, want male first (per GroupValueOrder)", table.Groups[0].Title)
	}
	if strings.ToLower(table.Groups[1].Title) != "female" {
		t.Errorf("band 1 = %q, want female", table.Groups[1].Title)
	}

	// Male band: last-name asc (Aale before Male) → "1 Aale, Al", "2 Male, Bob".
	male := table.Groups[0].Rows
	if got := male[0].Cells[0].Value; got != "1 Aale, Al" {
		t.Errorf("male[0] = %q, want %q", got, "1 Aale, Al")
	}
	if got := male[1].Cells[0].Value; got != "2 Male, Bob" {
		t.Errorf("male[1] = %q, want %q", got, "2 Male, Bob")
	}
	// Female band continues the sequence at 3 (continuous across bands).
	female := table.Groups[1].Rows
	if got := female[0].Cells[0].Value; got != "3 Fem, Ann" {
		t.Errorf("female[0] = %q, want %q (continuous numbering)", got, "3 Fem, Ann")
	}
}

// TestCSVSafe pins the spreadsheet-formula-injection guard: a value leading
// with = + - @ (or tab/CR) is tab-prefixed; benign values pass through.
func TestCSVSafe(t *testing.T) {
	cases := map[string]string{
		"=SUM(A1)":   "\t=SUM(A1)",
		"+1":         "\t+1",
		"-cmd":       "\t-cmd",
		"@import":    "\t@import",
		"Fem, Ann":   "Fem, Ann", // benign
		"6":          "6",
		"":           "", // empty untouched
		"\t=already": "\t\t=already",
	}
	for in, want := range cases {
		if got := csvSafe(in); got != want {
			t.Errorf("csvSafe(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestBuildRows_PhantomBlank_RealShown pins the phantom-blank invariant on the
// section grid cell render: an untaken-elective all-zero scaffold whose
// year-final floored to "1" renders a truly BLANK cell (empty HTML + empty CSV,
// NOT the floor and NOT the "—" no-data marker); a genuinely-enrolled subject
// scored a real 1 or 0 (positive mark evidence) KEEPS its rating link; a
// subject with no summary at all still renders "—". NEVER blank a real grade.
func TestBuildRows_PhantomBlank_RealShown(t *testing.T) {
	students := map[string]student{
		"stu1": {clientID: "stu1", name: "Doe, Jane", lastName: "Doe", firstName: "Jane"},
	}
	// Columns: phantom, real-1, real-0, no-data (no job).
	templateIDs := []string{"tP", "tR1", "tR0", "tE"}
	cellJob := map[string]string{
		"stu1\x00tP":  "jobP",
		"stu1\x00tR1": "jobR1",
		"stu1\x00tR0": "jobR0",
		// tE: no cellJob entry → no-data cell.
	}
	labelByJob := map[string]string{
		"jobP":  "1", // floored phantom
		"jobR1": "1", // real 1
		"jobR0": "0", // real 0
	}
	evByJob := map[string]outcome_summary.EnrollmentEvidence{
		"jobP":  {HasMarks: true, HasPositiveMark: false}, // all-zero scaffold → phantom
		"jobR1": {HasMarks: true, HasPositiveMark: true},  // enrolled → keep
		"jobR0": {HasMarks: true, HasPositiveMark: true},  // enrolled → keep
	}

	rows := buildRows(students, templateIDs, cellJob, labelByJob, evByJob, "sec1", outcome_summary.Routes{}, outcome_summary.Labels{})
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	cells := rows[0].Cells
	// cells: [0]=name, [1]=actions, [2]=phantom, [3]=real1, [4]=real0, [5]=no-data.
	if len(cells) != 6 {
		t.Fatalf("want 6 cells (name+actions+4 subjects), got %d", len(cells))
	}

	phantom, real1, real0, nodata := cells[2], cells[3], cells[4], cells[5]

	// Phantom: BLANK — empty CSV, and no rating anchor/label in the HTML.
	if csv := types.CellCSV(phantom); csv != "" {
		t.Errorf("phantom CSV = %q, want \"\" (blank)", csv)
	}
	ph := string(phantom.HTML)
	if strings.Contains(ph, "table-link") || strings.Contains(ph, ">1<") {
		t.Errorf("phantom cell must not render the floored rating, got HTML %q", ph)
	}
	if !strings.Contains(ph, "rc-cell-empty") {
		t.Errorf("phantom cell should carry the rc-cell-empty class, got HTML %q", ph)
	}

	// Real 1: SHOWN — CSV "1" and a rating link with the label.
	if csv := types.CellCSV(real1); csv != "1" {
		t.Errorf("real-1 CSV = %q, want \"1\" (a real grade must never blank)", csv)
	}
	if !strings.Contains(string(real1.HTML), "table-link") || !strings.Contains(string(real1.HTML), ">1<") {
		t.Errorf("real-1 cell must render its rating link, got HTML %q", string(real1.HTML))
	}

	// Real 0: SHOWN — CSV "0" and a rating link.
	if csv := types.CellCSV(real0); csv != "0" {
		t.Errorf("real-0 CSV = %q, want \"0\" (a real grade must never blank)", csv)
	}
	if !strings.Contains(string(real0.HTML), ">0<") {
		t.Errorf("real-0 cell must render its rating, got HTML %q", string(real0.HTML))
	}

	// No-data cell: the "—" marker (distinct from a phantom's true blank).
	if csv := types.CellCSV(nodata); csv != "—" {
		t.Errorf("no-data CSV = %q, want the \"—\" marker", csv)
	}
}

// TestNumberRows_Flat pins numbering on the ungrouped (no bands) path.
func TestNumberRows_Flat(t *testing.T) {
	table := &types.TableConfig{Rows: []types.TableRow{row("a", "Xi"), row("b", "Yo")}}
	numberRows(table)
	if table.Rows[0].Cells[0].Value != "1 Xi" || table.Rows[1].Cells[0].Value != "2 Yo" {
		t.Errorf("flat numbering wrong: %q, %q", table.Rows[0].Cells[0].Value, table.Rows[1].Cells[0].Value)
	}
}

// TestActionsColumnSkipContract (T8) pins the cross-function contract between the
// grid builder (buildColumns emits the frozen actionsColumnKey column) and the
// CSV export (exportSkipColumns must skip exactly that column and no data
// column). A rename of the shared const keeps both in lock-step; a divergence
// would leak raw HTML action anchors into every exported row.
func TestActionsColumnSkipContract(t *testing.T) {
	cols := buildColumns([]string{"t1", "t2"}, map[string]string{"t1": "Math", "t2": "Science"}, outcome_summary.Labels{})

	idx, count := -1, 0
	for i, c := range cols {
		if c.Key == actionsColumnKey {
			idx, count = i, count+1
		}
	}
	if count != 1 {
		t.Fatalf("buildColumns must emit exactly one %q column, got %d", actionsColumnKey, count)
	}

	skip := exportSkipColumns(cols)
	if len(skip) != 1 || !skip[idx] {
		t.Fatalf("exportSkipColumns must skip exactly the actions column (index %d), got %v", idx, skip)
	}
	for i := range cols {
		if skip[i] && cols[i].Key != actionsColumnKey {
			t.Fatalf("export skipped a data column %q (index %d) — the contract leaked", cols[i].Key, i)
		}
	}
}

// TestSectionBucket pins the strict authoritative partition (plan §3.0): the
// CURRENT template FK wins over the frozen job snapshot; a NULL FK or any
// out-of-corpus category folds into the single Uncategorized bucket; the
// snapshot is consulted ONLY when the template row is absent.
func TestSectionBucket(t *testing.T) {
	corpus := map[string]bool{"idA": true, "idD": true}
	meta := map[string]templateMeta{
		"tA":     {categoryID: "idA"}, // active FK
		"tNull":  {categoryID: ""},    // NULL FK
		"tStale": {categoryID: "idOld"},
	}
	cases := []struct {
		name       string
		snapshot   string
		templateID string
		want       string
	}{
		{"strict FK wins over disagreeing snapshot", "idD", "tA", "idA"},
		{"NULL FK folds to Uncategorized (snapshot ignored)", "idA", "tNull", uncategorizedTab},
		{"out-of-corpus FK folds to Uncategorized", "idA", "tStale", uncategorizedTab},
		{"missing template → snapshot fallback (in corpus)", "idD", "tMissing", "idD"},
		{"missing template + NULL snapshot → Uncategorized", "", "tMissing", uncategorizedTab},
		{"missing template + stale snapshot → Uncategorized", "idOld", "tMissing", uncategorizedTab},
	}
	for _, c := range cases {
		if got := sectionBucket(c.snapshot, c.templateID, meta, corpus); got != c.want {
			t.Errorf("%s: sectionBucket(%q,%q) = %q, want %q", c.name, c.snapshot, c.templateID, got, c.want)
		}
	}
}

// TestResolveSectionSelection pins the fail-closed ?jc= resolver (plan §3.3):
// unknown/foreign/inactive/empty falls back to the configured category, then
// the first present active category, then Uncategorized — NEVER all-categories,
// NEVER a raw stale id.
func TestResolveSectionSelection(t *testing.T) {
	academic := jcat("idA", "academic", "Academic", 1)
	deport := jcat("idD", "subject_deportment", "Deportment", 2)
	cats := []*jobcategorypb.JobCategory{academic, deport} // pre-sorted
	full := map[string]int{"idA": 3, "idD": 2, uncategorizedTab: 1}

	cases := []struct {
		name     string
		rawJC    string
		counts   map[string]int
		configed string
		want     string
	}{
		{"empty → configured default", "", full, "academic", "idA"},
		{"explicit present category", "idD", full, "academic", "idD"},
		{"explicit Uncategorized sentinel", uncategorizedTab, full, "academic", uncategorizedTab},
		{"foreign id → default", "id-foreign-xyz", full, "academic", "idA"},
		{"present-but-absent-count id → default", "idD", map[string]int{"idA": 3}, "academic", "idA"},
		{"configured absent → first present active", "", map[string]int{"idD": 2, uncategorizedTab: 1}, "academic", "idD"},
		{"only Uncategorized present", "", map[string]int{uncategorizedTab: 1}, "academic", uncategorizedTab},
		{"whitespace jc trimmed then defaulted", "   ", full, "academic", "idA"},
		{"nothing present → empty", "", map[string]int{}, "academic", ""},
	}
	for _, c := range cases {
		if got := resolveSectionSelection(c.rawJC, cats, c.counts, c.configed); got != c.want {
			t.Errorf("%s: resolveSectionSelection(%q) = %q, want %q", c.name, c.rawJC, got, c.want)
		}
	}
}

// TestSortSectionCategories pins the category-tab sort contract: sort_order ASC
// with NULLs LAST, then name ASC, then id (plan §3.3; mirrors
// list.sortLandingCategories).
func TestSortSectionCategories(t *testing.T) {
	alpha := jcat("a", "a", "Alpha", 2)
	beta := jcat("b", "b", "Beta", 1)
	noOrder := &jobcategorypb.JobCategory{Id: "z", Name: "Zeta", Active: true, Code: strp("z")} // NULL sort_order → last
	cats := []*jobcategorypb.JobCategory{noOrder, alpha, beta}
	sortSectionCategories(cats)
	got := []string{cats[0].GetId(), cats[1].GetId(), cats[2].GetId()}
	want := []string{"b", "a", "z"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sort order = %v, want %v (sort_order ASC, NULLs last)", got, want)
		}
	}
}

// TestCategoryIDByCode pins the in-memory configured-code → id resolution
// (case-insensitive, trimmed; no extra read).
func TestCategoryIDByCode(t *testing.T) {
	cats := []*jobcategorypb.JobCategory{jcat("idA", "academic", "Academic", 1)}
	cases := map[string]string{"academic": "idA", "ACADEMIC": "idA", "  academic  ": "idA", "": "", "missing": ""}
	for code, want := range cases {
		if got := categoryIDByCode(cats, code); got != want {
			t.Errorf("categoryIDByCode(%q) = %q, want %q", code, got, want)
		}
	}
}

// TestSectionTabKey_CollisionProof pins the stable, collision-proof tab key: the
// last-8 slug for a real id, the full Uncategorized word (which can never
// collide with a ≤8-char short() key nor with a real uuidv7 id).
func TestSectionTabKey_CollisionProof(t *testing.T) {
	if got := sectionTabKey(uncategorizedTab); got != "jc-tab-uncategorized" {
		t.Errorf("Uncategorized key = %q", got)
	}
	id := "0192f000-1111-7abc-aaaa-deadbeef1234"
	if got, want := sectionTabKey(id), "jc-tab-"+id[len(id)-8:]; got != want {
		t.Errorf("real key = %q, want %q", got, want)
	}
	if sectionTabKey("anything-eef1234") == sectionTabKey(uncategorizedTab) {
		t.Errorf("real short key collided with the Uncategorized sentinel key")
	}
	if got := sectionTabKey(""); got != "" {
		t.Errorf("empty id key = %q, want \"\"", got)
	}
}

// TestResolveSectionPartition_Tabbed pins the tabbed happy path end-to-end: the
// authoritative template→category meta, the present-only sorted tabstrip
// (Uncategorized last), the distinct-template counts, and the default-tab keep
// predicate (rawJC="" → configured "academic").
func TestResolveSectionPartition_Tabbed(t *testing.T) {
	academic := jcat("cat-academic-aaaa1111", "academic", "Academic", 1)
	deport := jcat("cat-deport-bbbb2222", "subject_deportment", "Subject Deportment", 2)
	emptyCat := jcat("cat-extra-cccc3333", "extracurricular", "Extracurricular", 3) // active but NO templates in section
	cats := []*jobcategorypb.JobCategory{deport, emptyCat, academic}                // deliberately unsorted
	tmpls := []*jobtemplatepb.JobTemplate{
		jtmpl("tA", "Math", academic.GetId()),
		jtmpl("tB", "Science", academic.GetId()),
		jtmpl("tD", "Conduct", deport.GetId()),
		jtmpl("tN", "Orphan", ""), // NULL FK → Uncategorized
	}
	jobs := []*jobpb.Job{
		jjob("j1", "tA", "", "c1"),
		jjob("j2", "tB", "", "c1"),
		jjob("j3", "tD", "", "c1"),
		jjob("j4", "tN", "", "c1"),
	}
	deps := &Deps{
		Options: outcome_summary.Options{
			List:           outcome_summary.ListOptions{ColumnsByField: "job_category"},
			CategoryFilter: "academic",
		},
		ListJobCategories: catListStub(cats, nil),
		ListJobTemplates:  tmplListStub(tmpls),
		Routes:            outcome_summary.Routes{SectionURL: "/report-cards/section/{id}"},
	}
	l := outcome_summary.Labels{}
	l.Section.CategoryTabsAriaLabel = "Report card categories"
	l.Landing.UncategorizedColumn = "Uncategorized"

	keep, meta, tabs := resolveSectionPartition(context.Background(), deps, "sec1", "", jobs, map[string]string{}, false, l)

	// Authoritative meta captured the FKs (NULL FK stays "").
	if meta["tA"].categoryID != academic.GetId() {
		t.Errorf("tA FK = %q, want %q", meta["tA"].categoryID, academic.GetId())
	}
	if meta["tN"].categoryID != "" {
		t.Errorf("tN FK = %q, want \"\" (NULL)", meta["tN"].categoryID)
	}

	// Default tab (rawJC="") = configured "academic": academic jobs kept, others dropped.
	if !keep(jobs[0]) || !keep(jobs[1]) {
		t.Errorf("academic jobs must be kept on the default (academic) tab")
	}
	if keep(jobs[2]) {
		t.Errorf("deportment job must be dropped on the academic tab (no leak)")
	}
	if keep(jobs[3]) {
		t.Errorf("uncategorized job must be dropped on the academic tab")
	}

	// Tabstrip: present-only (Extracurricular has no templates → no tab), sorted,
	// Uncategorized last, active = academic.
	if tabs == nil || len(tabs.Items) != 3 {
		t.Fatalf("want 3 present tabs, got %+v", tabs)
	}
	if tabs.Items[0].Label != "Academic" || tabs.Items[1].Label != "Subject Deportment" || tabs.Items[2].Label != "Uncategorized" {
		t.Errorf("tab order/labels = [%q %q %q]", tabs.Items[0].Label, tabs.Items[1].Label, tabs.Items[2].Label)
	}
	if tabs.Items[0].Count != 2 { // tA + tB
		t.Errorf("Academic count = %d, want 2 subjects", tabs.Items[0].Count)
	}
	if tabs.Items[2].Count != 1 { // tN
		t.Errorf("Uncategorized count = %d, want 1", tabs.Items[2].Count)
	}
	if tabs.ActiveTab != sectionTabKey(academic.GetId()) {
		t.Errorf("active tab = %q, want %q", tabs.ActiveTab, sectionTabKey(academic.GetId()))
	}
	if tabs.Aria != "Report card categories" {
		t.Errorf("aria = %q", tabs.Aria)
	}
	if !strings.Contains(tabs.Items[0].Href, "jc=cat-academic-aaaa1111") {
		t.Errorf("academic href = %q, want query-encoded ?jc=", tabs.Items[0].Href)
	}
	if !strings.Contains(tabs.Items[2].Href, "jc=uncategorized") {
		t.Errorf("uncategorized href = %q", tabs.Items[2].Href)
	}

	// Selecting the deportment tab keeps ONLY deportment jobs (labelled tab, no leak).
	keepD, _, tabsD := resolveSectionPartition(context.Background(), deps, "sec1", deport.GetId(), jobs, map[string]string{}, false, l)
	if keepD(jobs[0]) || !keepD(jobs[2]) {
		t.Errorf("deportment tab must keep only deportment jobs")
	}
	if tabsD.ActiveTab != sectionTabKey(deport.GetId()) {
		t.Errorf("deportment active tab = %q", tabsD.ActiveTab)
	}
}

// TestResolveSectionPartition_StaticWhenKnobOff pins the config-gated degrade:
// with no category-columns Options the section renders today's single static-H2
// grid, no tabs, no meta read — byte-for-byte the pre-tab behavior.
func TestResolveSectionPartition_StaticWhenKnobOff(t *testing.T) {
	academic := jcat("idA", "academic", "Academic", 1)
	deport := jcat("idD", "subject_deportment", "Deportment", 2)
	deps := &Deps{
		Options:           outcome_summary.Options{CategoryFilter: "academic"}, // no ColumnsByField → knob OFF
		ListJobCategories: catListStub([]*jobcategorypb.JobCategory{academic, deport}, nil),
	}
	jobs := []*jobpb.Job{jjob("j1", "tA", "idA", "c1"), jjob("j2", "tD", "idD", "c1")}
	keep, meta, tabs := resolveSectionPartition(context.Background(), deps, "s", "", jobs, map[string]string{}, false, outcome_summary.Labels{})
	if tabs != nil {
		t.Errorf("static path must render no tabs")
	}
	if meta != nil {
		t.Errorf("static path must not fetch template meta")
	}
	if !keep(jobs[0]) {
		t.Errorf("static H2 filter must keep the academic job")
	}
	if keep(jobs[1]) {
		t.Errorf("static H2 filter must drop the deportment job")
	}
}

// TestResolveSectionPartition_FailClosedOnCorpusError pins the fail-closed
// contract: a category read error degrades to the static H2 filter with NO tabs
// — never an all-categories render (which would strip H2 and leak deportment).
// The static filter itself fails closed (keeps no jobs) when the configured
// category cannot resolve.
func TestResolveSectionPartition_FailClosedOnCorpusError(t *testing.T) {
	deps := &Deps{
		Options: outcome_summary.Options{
			List:           outcome_summary.ListOptions{ColumnsByField: "job_category"},
			CategoryFilter: "academic",
		},
		ListJobCategories: catListStub(nil, errors.New("boom")),
	}
	jobs := []*jobpb.Job{jjob("j1", "tA", "idA", "c1")}
	keep, meta, tabs := resolveSectionPartition(context.Background(), deps, "s", "idA", jobs, map[string]string{}, false, outcome_summary.Labels{})
	if tabs != nil {
		t.Errorf("a corpus read error must degrade to NO tabs (never all-categories)")
	}
	if meta != nil {
		t.Errorf("no meta on the fail-closed path")
	}
	if keep(jobs[0]) {
		t.Errorf("fail-closed: an unresolvable configured filter must keep NO jobs")
	}
}

// TestSectionTemplate_R8ScriptAndTabsContract pins that the R8 entries-seed
// script (constant table id "report-cards-grid") survives the tabstrip addition,
// that the tabs live inside the SINGLE content define (goldens untouched), and
// that the constant Table.ID literal remains in page.go.
func TestSectionTemplate_R8ScriptAndTabsContract(t *testing.T) {
	tpl, err := os.ReadFile("../templates/section.html")
	if err != nil {
		t.Fatalf("read section.html: %v", err)
	}
	s := string(tpl)
	for _, want := range []string{
		"getElementById('report-cards-grid-entries')", // R8 seed script — constant id
		"{{if .TabItems}}",                            // tabstrip gate
		`{{template "tabs"`,                           // pyeza tabs component
		`"report-cards-category-tabs"`,                // tabstrip nav id
		`"Indicator" "#tabContent"`,                   // R2 indicator precedent
		`id="tabContent"`,
		`role="tabpanel"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("section.html must contain %q", want)
		}
	}
	if n := strings.Count(s, `{{define "outcome-summary-section-content"}}`); n != 1 {
		t.Errorf("want exactly ONE content define block (templates_golden untouched), got %d", n)
	}

	pg, err := os.ReadFile("page.go")
	if err != nil {
		t.Fatalf("read page.go: %v", err)
	}
	if !strings.Contains(string(pg), `"report-cards-grid"`) {
		t.Errorf("page.go must keep the constant Table.ID \"report-cards-grid\" across tabs (R8)")
	}
}
