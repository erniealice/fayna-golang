package list

import (
	"testing"

	enumspb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"
)

// A numeric cell descriptor never carries description data: the rating
// description is snapshotted into the outcome narrative on save (see the record
// action), not previewed inside the cell.
func numericCriteria() *criteriapb.OutcomeCriteria {
	return &criteriapb.OutcomeCriteria{CriteriaType: enumspb.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE}
}

func TestBuildCellInputNumericIsJustANumber(t *testing.T) {
	d := buildCellInput(&matrixpb.CriterionColumn{Criteria: numericCriteria()})
	if d.Type != "numeric" {
		t.Fatalf("descriptor type = %q, want numeric", d.Type)
	}
}

func narrativeRows(cell *matrixpb.OutcomeCell) []*matrixpb.OutcomeRow {
	return []*matrixpb.OutcomeRow{{ClientId: "c1", Cells: map[string]*matrixpb.OutcomeCell{"tt1:cr1": cell}}}
}

func TestBuildRowsNarrativeIconStates(t *testing.T) {
	nl := outcome_matrix.DefaultLabels().Narrative
	labels := map[string]string{"tt1:cr1": "Knowing"}
	const base = "/action/grade-sheet/T1/narrative"
	build := func(cell *matrixpb.OutcomeCell, baseURL string) (got struct{ url, base, add, edit, aria string }) {
		row := buildRows(narrativeRows(cell), "staff-1", "ro", map[string]clientName{"c1": {display: "Ana"}}, nil, baseURL, labels, nl)
		c := row[0].Cells["tt1:cr1"]
		got.url, got.base, got.add, got.edit, got.aria = c.NarrativeURL, c.NarrativeBaseURL, c.NarrativeAriaAdd, c.NarrativeAriaEdit, c.NarrativeAria
		return got
	}

	// Editable, nothing recorded yet → dormant icon: base only, both verbs.
	g := build(&matrixpb.OutcomeCell{JobTaskId: "jt1", Editable: true}, base)
	if g.url != "" || g.base != base || g.add == "" || g.edit == "" || g.add == g.edit {
		t.Errorf("unrecorded editable cell must carry base + both verbs and no live URL: %+v", g)
	}

	// Editable + recorded → live URL and base.
	g = build(&matrixpb.OutcomeCell{JobTaskId: "jt1", Editable: true, OutcomeId: "o1", RecordedBy: "staff-1"}, base)
	if g.url != base+"?outcome_id=o1" || g.base != base {
		t.Errorf("recorded editable cell must carry live URL + base: %+v", g)
	}

	// Recorded by someone else (read-only) → live URL for viewing, never a base.
	g = build(&matrixpb.OutcomeCell{JobTaskId: "jt1", Editable: false, OutcomeId: "o1", RecordedBy: "other"}, base)
	if g.url == "" || g.base != "" || g.add != "" {
		t.Errorf("read-only recorded cell must be view-only: %+v", g)
	}

	// Not editable and nothing recorded → no icon at all.
	if g = build(&matrixpb.OutcomeCell{JobTaskId: "jt1", Editable: false}, base); g.url != "" || g.base != "" || g.aria != "" {
		t.Errorf("non-editable unrecorded cell must have no icon: %+v", g)
	}

	// Feature off (no drawer route) → no icon anywhere.
	if g = build(&matrixpb.OutcomeCell{JobTaskId: "jt1", Editable: true, OutcomeId: "o1", RecordedBy: "staff-1"}, ""); g.url != "" || g.base != "" {
		t.Errorf("narrative feature off must emit no icon: %+v", g)
	}
}
