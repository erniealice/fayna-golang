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

// TestNumberRows_Flat pins numbering on the ungrouped (no bands) path.
func TestNumberRows_Flat(t *testing.T) {
	table := &types.TableConfig{Rows: []types.TableRow{row("a", "Xi"), row("b", "Yo")}}
	numberRows(table)
	if table.Rows[0].Cells[0].Value != "1 Xi" || table.Rows[1].Cells[0].Value != "2 Yo" {
		t.Errorf("flat numbering wrong: %q, %q", table.Rows[0].Cells[0].Value, table.Rows[1].Cells[0].Value)
	}
}
