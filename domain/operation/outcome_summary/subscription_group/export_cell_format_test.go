package subscription_group

import (
	"context"
	"reflect"
	"testing"
)

// TestExplicitExport_CompositeValues: the explicit matrix (which feeds the group
// CSV and the PDF document matrix) renders cells through the configured cell
// template; the default keeps the scaled values byte-identical.
func TestExplicitExport_CompositeValues(t *testing.T) {
	build := func(format string) []string {
		resp := exportFixture()
		score := 28.0
		cells := resp.ClientRows[0].Cells
		cells[0] = exportCell("job-b", exportString("7"), nil, true, true)
		cells[0].SummaryScore = &score
		deps, _ := exportDeps(resp)
		deps.Options.SubscriptionGroupExport.CellFormat = format
		matrix, err := normalizeExplicitMatrix(context.Background(), deps, resp)
		if err != nil {
			t.Fatalf("normalize (%q): %v", format, err)
		}
		return matrix.rows[0].values
	}
	// Columns are English (job-a) then Math (job-b). job-a: label "0", no
	// composite, positive mark → collapses to its label under every template.
	if got, want := build(""), []string{"0", "7"}; !reflect.DeepEqual(got, want) {
		t.Errorf("default values = %v, want %v", got, want)
	}
	if got, want := build("{composite}"), []string{"0", "28"}; !reflect.DeepEqual(got, want) {
		t.Errorf("composite values = %v, want %v", got, want)
	}
	if got, want := build("{composite} / {scaled}"), []string{"0", "28 / 7"}; !reflect.DeepEqual(got, want) {
		t.Errorf("both values = %v, want %v", got, want)
	}
}
