package section

import (
	"strings"
	"testing"

	outcome_summary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/types"
)

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
