package list

import (
	"context"
	"testing"

	"github.com/erniealice/pyeza-golang/types"

	enumspb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// withAggregation stamps an aggregation method onto every criterion of a phase.
func withAggregation(ph *matrixpb.PhaseColumn, m enumspb.AggregationMethod) {
	for _, task := range ph.GetTasks() {
		for _, c := range task.GetCriteria() {
			c.Criteria = &criteriapb.OutcomeCriteria{Id: "cr", AggregationMethod: m}
		}
	}
}

func compositeLeaf(t *testing.T, grid *types.CellGridConfig, phaseID string) types.CellGridLevel3 {
	t.Helper()
	for _, l1 := range grid.Columns {
		if l1.Key != phaseID {
			continue
		}
		for _, l2 := range l1.Level2 {
			if l2.Key == totalKeyPrefix+phaseID {
				return l2.Level3[0]
			}
		}
	}
	t.Fatalf("composite leaf for %s missing", phaseID)
	return types.CellGridLevel3{}
}

func augmentedAverageGrid(t *testing.T, aggA, aggB enumspb.AggregationMethod) (*types.CellGridConfig, *PageViewDeps) {
	t.Helper()
	resp := ratingsResp()
	withAggregation(resp.Phases[0], aggA)
	withAggregation(resp.Phases[1], aggB)
	deps := ratingsDeps(func(ctx context.Context, req *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error) {
		return &matrixpb.GetOutcomeSummaryRosterResponse{
			Rows: []*matrixpb.OutcomeSummaryRosterRow{withTotals(rosterRow("c1", "O", "7", "7"), f64(99), f64(28))},
		}, nil
	})
	grid := buildGrid(context.Background(), deps, types.NewEmptyUserPermissions(), resp, true, "tmpl-1", nil, ratingsViewCtx(), nil)
	augmentRatingColumns(context.Background(), deps, grid, resp, true, "", "all", "")
	return grid, deps
}

// TestAugmentRatingColumns_AverageHeaderForAveragedPhase: a phase whose every
// criterion aggregates by AVERAGE gets the average header + cell tooltip; the
// sibling MAXIMUM phase keeps the total wording.
func TestAugmentRatingColumns_AverageHeaderForAveragedPhase(t *testing.T) {
	grid, deps := augmentedAverageGrid(t, enumspb.AggregationMethod_AGGREGATION_METHOD_AVERAGE, enumspb.AggregationMethod_AGGREGATION_METHOD_MAXIMUM)
	g := deps.Labels.Grid
	if got := compositeLeaf(t, grid, "pA").Label; got != g.AverageColumn {
		t.Errorf("averaged phase header = %q, want %q", got, g.AverageColumn)
	}
	if got := compositeLeaf(t, grid, "pB").Label; got != g.TotalColumn {
		t.Errorf("maximum phase header = %q, want %q", got, g.TotalColumn)
	}
	for _, row := range grid.Rows {
		if row.ID != "c1" {
			continue
		}
		if got := row.Cells[totalKeyPrefix+"pA"].ReadOnlyTooltip; got != g.AverageTooltip {
			t.Errorf("averaged cell tooltip = %q, want %q", got, g.AverageTooltip)
		}
		if got := row.Cells[totalKeyPrefix+"pB"].ReadOnlyTooltip; got != g.TotalTooltip {
			t.Errorf("total cell tooltip = %q, want %q", got, g.TotalTooltip)
		}
	}
}

// TestAugmentRatingColumns_TotalHeaderOtherwise: unspecified (legacy rows) and
// mixed phases never read as averaged.
func TestAugmentRatingColumns_TotalHeaderOtherwise(t *testing.T) {
	grid, deps := augmentedAverageGrid(t, enumspb.AggregationMethod_AGGREGATION_METHOD_UNSPECIFIED, enumspb.AggregationMethod_AGGREGATION_METHOD_MAXIMUM)
	for _, id := range []string{"pA", "pB"} {
		if got := compositeLeaf(t, grid, id).Label; got != deps.Labels.Grid.TotalColumn {
			t.Errorf("phase %s header = %q, want total", id, got)
		}
	}

	mixed := ratingsResp()
	mixed.Phases[0].Tasks = append(mixed.Phases[0].Tasks, &matrixpb.TaskColumn{
		JobTemplateTaskId: "tA2",
		Criteria:          []*matrixpb.CriterionColumn{{ColumnKey: "tA2:crB"}},
	})
	withAggregation(mixed.Phases[0], enumspb.AggregationMethod_AGGREGATION_METHOD_AVERAGE)
	mixed.Phases[0].Tasks[1].Criteria[0].Criteria = &criteriapb.OutcomeCriteria{Id: "crB", AggregationMethod: enumspb.AggregationMethod_AGGREGATION_METHOD_MAXIMUM}
	if averagedPhases(mixed)["pA"] {
		t.Error("mixed AVERAGE/MAXIMUM phase must not read as averaged")
	}
	if averagedPhases(&matrixpb.GetOutcomeMatrixResponse{Phases: []*matrixpb.PhaseColumn{{JobTemplatePhaseId: "empty"}}})["empty"] {
		t.Error("phase with no criteria must not read as averaged")
	}
}
