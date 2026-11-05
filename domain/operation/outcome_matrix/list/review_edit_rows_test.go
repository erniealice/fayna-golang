package list

import (
	"testing"

	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"
)

// Review edit (docs/plan/20261105-reviewer-edit-grades): a cell recorded by
// another staff renders editable only when the server marks it Editable; with
// Editable=false it stays read-only exactly as before; a locked phase still
// forces read-only.
func TestBuildRows_ReviewEditReadOnlyRule(t *testing.T) {
	nl := outcome_matrix.DefaultLabels().Narrative
	labels := map[string]string{"tt1:cr1": "Knowing"}
	build := func(cell *matrixpb.OutcomeCell, allow func(string) bool) (editable, readOnly bool) {
		row := buildRows(narrativeRows(cell), "staff-1", "ro", map[string]clientName{"c1": {display: "Ana"}}, allow, "", labels, nl)
		c := row[0].Cells["tt1:cr1"]
		return c.Editable, c.ReadOnly
	}
	locked := func(string) bool { return false }

	cases := []struct {
		name         string
		cell         *matrixpb.OutcomeCell
		allow        func(string) bool
		wantEditable bool
		wantRO       bool
	}{
		{"other recorder, server editable", &matrixpb.OutcomeCell{OutcomeId: "o1", RecordedBy: "staff-2", Editable: true}, nil, true, false},
		{"other recorder, not editable (no role)", &matrixpb.OutcomeCell{OutcomeId: "o1", RecordedBy: "staff-2", Editable: false}, nil, false, true},
		{"own cell", &matrixpb.OutcomeCell{OutcomeId: "o1", RecordedBy: "staff-1", Editable: true}, nil, true, false},
		{"other recorder, editable but phase locked", &matrixpb.OutcomeCell{OutcomeId: "o1", RecordedBy: "staff-2", Editable: true}, locked, false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ed, ro := build(c.cell, c.allow)
			if ed != c.wantEditable || ro != c.wantRO {
				t.Fatalf("editable=%t readOnly=%t, want %t/%t", ed, ro, c.wantEditable, c.wantRO)
			}
		})
	}
}
