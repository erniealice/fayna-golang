package list

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// ratingsResp builds a 2-phase (s1/s2), 1-task, 1-criterion response with two
// client rows — the smallest tree that exercises every augmentation seam.
func ratingsResp() *matrixpb.GetOutcomeMatrixResponse {
	num := func() *matrixpb.CriterionColumn {
		return &matrixpb.CriterionColumn{ColumnKey: "tA:crA"}
	}
	return &matrixpb.GetOutcomeMatrixResponse{
		JobTemplateId:   "tmpl-1",
		JobTemplateName: "Subject",
		Phases: []*matrixpb.PhaseColumn{
			{JobTemplatePhaseId: "pA", Label: "Semester 1", Code: "s1",
				Tasks: []*matrixpb.TaskColumn{{JobTemplateTaskId: "tA", Label: "Task A", Criteria: []*matrixpb.CriterionColumn{num()}}}},
			{JobTemplatePhaseId: "pB", Label: "Semester 2", Code: "s2",
				Tasks: []*matrixpb.TaskColumn{{JobTemplateTaskId: "tB", Label: "Task B",
					Criteria: []*matrixpb.CriterionColumn{{ColumnKey: "tB:crA"}}}}},
		},
		Rows: []*matrixpb.OutcomeRow{
			{ClientId: "c1"},
			{ClientId: "c2"},
		},
	}
}

func rosterRow(clientID, s1, s2, final string) *matrixpb.OutcomeSummaryRosterRow {
	return &matrixpb.OutcomeSummaryRosterRow{
		ClientId: clientID,
		Phases: []*matrixpb.OutcomeSummaryPhaseEntry{
			{JobTemplatePhaseId: "pA", Code: "s1", Label: "Semester 1", SequenceOrder: 1, ScaledLabel: s1},
			{JobTemplatePhaseId: "pB", Code: "s2", Label: "Semester 2", SequenceOrder: 2, ScaledLabel: s2},
		},
		YearFinalLabel: final,
	}
}

// TestAugmentRatingColumns pins the whole augmentation contract: synthetic
// per-phase + final columns, verbatim stored values, the grid-rows roster
// filter (a roster row with no rendered grid row is never read), read-only
// cells, and the per-column download stamping incl. the group-scoped base.
func TestAugmentRatingColumns(t *testing.T) {
	resp := ratingsResp()
	deps := &PageViewDeps{
		Labels: outcome_matrix.DefaultLabels(),
		Routes: outcome_matrix.DefaultRoutes(),
		GetOutcomeSummaryRoster: func(ctx context.Context, req *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error) {
			if req.GetJobTemplateId() != "tmpl-1" {
				t.Fatalf("roster read got template %q", req.GetJobTemplateId())
			}
			if req.GetScope() != matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL {
				t.Fatalf("roster read must thread the resolved scope, got %v", req.GetScope())
			}
			return &matrixpb.GetOutcomeSummaryRosterResponse{
				JobTemplateId: "tmpl-1",
				Rows: []*matrixpb.OutcomeSummaryRosterRow{
					rosterRow("c1", "7", "6", "7"),
					rosterRow("c2", "5", "", ""),
					// The template-wide roster carries students the (group-scoped)
					// grid does NOT render — the grid rows are the filter.
					rosterRow("c-foreign", "4", "4", "4"),
				},
			}, nil
		},
	}
	perms := types.NewEmptyUserPermissions()
	viewCtx := &view.ViewContext{Request: httptest.NewRequest("GET", "/outcome-matrix/tmpl-1", nil)}
	grid := buildGrid(context.Background(), deps, perms, resp, true, "tmpl-1", nil, viewCtx, nil)
	if got := grid.LeafColumnCount(); got != 2 {
		t.Fatalf("pre-augment leaf count = %d, want 2", got)
	}

	base := "/outcome-matrix/tmpl-1/subscription-group/g1/export"
	augmentRatingColumns(context.Background(), deps, grid, resp, true, base, "all", "")

	// Column tree: each phase gained ONE trailing rating leaf; one final L1.
	if got := grid.LeafColumnCount(); got != 5 {
		t.Fatalf("post-augment leaf count = %d, want 5 (2 criteria + 2 phase ratings + final)", got)
	}
	if len(grid.Columns) != 3 {
		t.Fatalf("L1 count = %d, want 3 (2 phases + final)", len(grid.Columns))
	}
	pA := grid.Columns[0]
	lastL2 := pA.Level2[len(pA.Level2)-1]
	if len(lastL2.Level3) != 1 || lastL2.Level3[0].ColumnKey != "rating:pA" {
		t.Fatalf("phase pA trailing leaf = %+v, want ColumnKey rating:pA", lastL2)
	}
	if lastL2.Level3[0].Label != "Rating" {
		t.Errorf("rating leaf label = %q, want Rating", lastL2.Level3[0].Label)
	}
	fin := grid.Columns[2]
	if fin.Key != finalL1Key || fin.Label != "Final" {
		t.Fatalf("final L1 = {%q %q}, want {rating-final Final}", fin.Key, fin.Label)
	}
	if len(fin.Level2) != 1 || fin.Level2[0].Level3[0].ColumnKey != finalColumnKey {
		t.Fatalf("final leaf = %+v, want ColumnKey %s", fin.Level2, finalColumnKey)
	}

	// Cells: verbatim stored values, read-only, roster filtered to grid rows.
	if len(grid.Rows) != 2 {
		t.Fatalf("row count changed: %d, want 2 (roster c-foreign must not add a row)", len(grid.Rows))
	}
	byID := map[string]types.CellGridRow{}
	for _, r := range grid.Rows {
		byID[r.ID] = r
	}
	c1 := byID["c1"]
	if c := c1.Cells["rating:pA"]; c.Value != "7" || c.Editable {
		t.Errorf("c1 rating:pA = {%q editable=%v}, want {7 false}", c.Value, c.Editable)
	}
	if c := c1.Cells[finalColumnKey]; c.Value != "7" || c.Editable {
		t.Errorf("c1 final = {%q editable=%v}, want {7 false}", c.Value, c.Editable)
	}
	if c := c1.Cells["rating:pA"]; c.TestID != "om-rating-c1-s1" {
		t.Errorf("c1 rating testid = %q, want om-rating-c1-s1", c.TestID)
	}
	c2 := byID["c2"]
	if c := c2.Cells["rating:pB"]; c.Value != "" {
		t.Errorf("c2 rating:pB = %q, want empty (no stored summary)", c.Value)
	}
	for _, r := range grid.Rows {
		for k := range r.Cells {
			if strings.HasPrefix(k, ratingKeyPrefix) && k != finalColumnKey &&
				k != "rating:pA" && k != "rating:pB" {
				t.Errorf("unexpected rating cell key %q (foreign roster data leaked)", k)
			}
		}
	}

	// Download stamping: per-phase period token on the group-scoped base;
	// final header carries period=final. Aria composed from the label template.
	paActions, ok := pA.Actions.(PhaseActions)
	if !ok {
		t.Fatalf("phase pA Actions = %T, want PhaseActions", pA.Actions)
	}
	wantURL := base + "?scope=all&period=s1"
	if paActions.DownloadURL != wantURL {
		t.Errorf("pA download URL = %q, want %q", paActions.DownloadURL, wantURL)
	}
	if paActions.DownloadTestID != "om-dl-s1" {
		t.Errorf("pA download testid = %q, want om-dl-s1", paActions.DownloadTestID)
	}
	if paActions.DownloadAria != "Download Semester 1" {
		t.Errorf("pA download aria = %q", paActions.DownloadAria)
	}
	finActions, ok := fin.Actions.(PhaseActions)
	if !ok {
		t.Fatalf("final Actions = %T, want PhaseActions", fin.Actions)
	}
	if want := base + "?scope=all&period=final"; finActions.DownloadURL != want {
		t.Errorf("final download URL = %q, want %q", finActions.DownloadURL, want)
	}
	if finActions.DownloadTestID != "om-dl-final" {
		t.Errorf("final download testid = %q", finActions.DownloadTestID)
	}
	if !finActions.Any() {
		t.Errorf("final PhaseActions.Any() = false — the header slot would not render")
	}
}

// TestAugmentRatingColumns_NoRoster pins the fail-safe: a nil roster closure
// leaves the column tree untouched (no rating columns, no final), while the
// download stamping — independent of the composite read — still applies.
func TestAugmentRatingColumns_NoRoster(t *testing.T) {
	resp := ratingsResp()
	deps := &PageViewDeps{Labels: outcome_matrix.DefaultLabels(), Routes: outcome_matrix.DefaultRoutes()}
	perms := types.NewEmptyUserPermissions()
	viewCtx := &view.ViewContext{Request: httptest.NewRequest("GET", "/outcome-matrix/tmpl-1", nil)}
	grid := buildGrid(context.Background(), deps, perms, resp, true, "tmpl-1", nil, viewCtx, nil)

	augmentRatingColumns(context.Background(), deps, grid, resp, true, "/outcome-matrix/tmpl-1/export", "mine", "h1")

	if got := grid.LeafColumnCount(); got != 2 {
		t.Fatalf("leaf count = %d, want 2 (no rating columns without a roster read)", got)
	}
	if len(grid.Columns) != 2 {
		t.Fatalf("L1 count = %d, want 2 (no final column)", len(grid.Columns))
	}
	pa, ok := grid.Columns[0].Actions.(PhaseActions)
	if !ok || pa.DownloadURL != "/outcome-matrix/tmpl-1/export?scope=mine&hide=h1&period=s1" {
		t.Errorf("download stamp = %+v ok=%v", pa, ok)
	}
}

// TestAugmentRatingColumns_NoExportBase pins that an unwired export route
// yields NO download affordances (a dead icon link is worse than none) while
// the rating columns still build from the roster read.
func TestAugmentRatingColumns_NoExportBase(t *testing.T) {
	resp := ratingsResp()
	deps := &PageViewDeps{
		Labels: outcome_matrix.DefaultLabels(),
		Routes: outcome_matrix.DefaultRoutes(),
		GetOutcomeSummaryRoster: func(ctx context.Context, req *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error) {
			return &matrixpb.GetOutcomeSummaryRosterResponse{
				Rows: []*matrixpb.OutcomeSummaryRosterRow{rosterRow("c1", "7", "6", "7")},
			}, nil
		},
	}
	perms := types.NewEmptyUserPermissions()
	viewCtx := &view.ViewContext{Request: httptest.NewRequest("GET", "/outcome-matrix/tmpl-1", nil)}
	grid := buildGrid(context.Background(), deps, perms, resp, true, "tmpl-1", nil, viewCtx, nil)

	augmentRatingColumns(context.Background(), deps, grid, resp, true, "", "all", "")

	if got := grid.LeafColumnCount(); got != 5 {
		t.Fatalf("leaf count = %d, want 5", got)
	}
	if pa, ok := grid.Columns[0].Actions.(PhaseActions); ok && pa.DownloadURL != "" {
		t.Errorf("unexpected download URL with no export base: %q", pa.DownloadURL)
	}
	if fin := grid.Columns[2]; fin.Actions != nil {
		t.Errorf("final column should carry no action payload with no export base, got %+v", fin.Actions)
	}
}
