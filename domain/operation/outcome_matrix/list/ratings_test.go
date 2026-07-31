package list

import (
	"context"
	"errors"
	"net/http/httptest"
	"reflect"
	"strconv"
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

// f64 makes a presence-tracked summary_score. A nil pointer is the UNSET
// (nothing computed) case — deliberately distinct from f64(0).
func f64(v float64) *float64 { return &v }

// withTotals sets the raw composite on the pA/pB entries of a roster row.
func withTotals(r *matrixpb.OutcomeSummaryRosterRow, tA, tB *float64) *matrixpb.OutcomeSummaryRosterRow {
	r.Phases[0].SummaryScore = tA
	r.Phases[1].SummaryScore = tB
	return r
}

func ratingsDeps(roster func(context.Context, *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error)) *PageViewDeps {
	return &PageViewDeps{
		Labels:                  outcome_matrix.DefaultLabels(),
		Routes:                  outcome_matrix.DefaultRoutes(),
		GetOutcomeSummaryRoster: roster,
	}
}

func ratingsViewCtx() *view.ViewContext {
	return &view.ViewContext{Request: httptest.NewRequest("GET", "/outcome-matrix/tmpl-1", nil)}
}

// TestAugmentRatingColumns pins the whole augmentation contract: synthetic
// per-phase total+rating pairs + the final column, verbatim stored values, the
// grid-rows roster filter (a roster row with no rendered grid row is never
// read), read-only cells, and the per-column download stamping incl. the
// group-scoped base.
func TestAugmentRatingColumns(t *testing.T) {
	resp := ratingsResp()
	deps := ratingsDeps(func(ctx context.Context, req *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error) {
		if req.GetJobTemplateId() != "tmpl-1" {
			t.Fatalf("roster read got template %q", req.GetJobTemplateId())
		}
		if req.GetScope() != matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL {
			t.Fatalf("roster read must thread the resolved scope, got %v", req.GetScope())
		}
		return &matrixpb.GetOutcomeSummaryRosterResponse{
			JobTemplateId: "tmpl-1",
			Rows: []*matrixpb.OutcomeSummaryRosterRow{
				withTotals(rosterRow("c1", "7", "6", "7"), f64(28), f64(25)),
				rosterRow("c2", "5", "", ""),
				// The template-wide roster carries students the (group-scoped)
				// grid does NOT render — the grid rows are the filter.
				withTotals(rosterRow("c-foreign", "4", "4", "4"), f64(16), f64(16)),
			},
		}, nil
	})
	perms := types.NewEmptyUserPermissions()
	grid := buildGrid(context.Background(), deps, perms, resp, true, "tmpl-1", nil, ratingsViewCtx(), nil)
	if got := grid.LeafColumnCount(); got != 2 {
		t.Fatalf("pre-augment leaf count = %d, want 2", got)
	}

	base := "/action/outcome-matrix/tmpl-1/subscription-group/g1/download"
	augmentRatingColumns(context.Background(), deps, grid, resp, true, base, "all", "")

	// Column tree: each phase gained a trailing TOTAL + RATING pair; one final L1.
	if got := grid.LeafColumnCount(); got != 7 {
		t.Fatalf("post-augment leaf count = %d, want 7 (2 criteria + 2x(total+rating) + final)", got)
	}
	if len(grid.Columns) != 3 {
		t.Fatalf("L1 count = %d, want 3 (2 phases + final)", len(grid.Columns))
	}
	pA := grid.Columns[0]
	if len(pA.Level2) != 3 {
		t.Fatalf("phase pA L2 count = %d, want 3 (task + total + rating)", len(pA.Level2))
	}
	totalL2, ratingL2 := pA.Level2[1], pA.Level2[2]
	if len(totalL2.Level3) != 1 || totalL2.Level3[0].ColumnKey != "total:pA" {
		t.Fatalf("phase pA total leaf = %+v, want ColumnKey total:pA", totalL2)
	}
	if len(ratingL2.Level3) != 1 || ratingL2.Level3[0].ColumnKey != "rating:pA" {
		t.Fatalf("phase pA rating leaf = %+v, want ColumnKey rating:pA", ratingL2)
	}
	if totalL2.Level3[0].Label != "Total" {
		t.Errorf("total leaf label = %q, want Total", totalL2.Level3[0].Label)
	}
	if ratingL2.Level3[0].Label != "Rating" {
		t.Errorf("rating leaf label = %q, want Rating", ratingL2.Level3[0].Label)
	}
	fin := grid.Columns[2]
	if fin.Key != finalL1Key || fin.Label != "Final" {
		t.Fatalf("final L1 = {%q %q}, want {rating-final Final}", fin.Key, fin.Label)
	}
	// D5: the year-final column carries the RATING only — no total twin in v1.
	if len(fin.Level2) != 1 || fin.Level2[0].Level3[0].ColumnKey != finalColumnKey {
		t.Fatalf("final leaf = %+v, want exactly one leaf with ColumnKey %s", fin.Level2, finalColumnKey)
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
	if c := c1.Cells["total:pA"]; c.Value != "28" || c.Editable {
		t.Errorf("c1 total:pA = {%q editable=%v}, want {28 false}", c.Value, c.Editable)
	}
	if c := c1.Cells["total:pA"]; c.ReadOnlyTooltip != "Computed total — read only" {
		t.Errorf("c1 total:pA tooltip = %q, want the total tooltip", c.ReadOnlyTooltip)
	}
	if c := c1.Cells[finalColumnKey]; c.Value != "7" || c.Editable {
		t.Errorf("c1 final = {%q editable=%v}, want {7 false}", c.Value, c.Editable)
	}
	if c := c1.Cells["rating:pA"]; c.TestID != "om-rating-c1-s1" {
		t.Errorf("c1 rating testid = %q, want om-rating-c1-s1", c.TestID)
	}
	if c := c1.Cells["total:pA"]; c.TestID != "om-total-c1-s1" {
		t.Errorf("c1 total testid = %q, want om-total-c1-s1", c.TestID)
	}
	c2 := byID["c2"]
	if c := c2.Cells["rating:pB"]; c.Value != "" {
		t.Errorf("c2 rating:pB = %q, want empty (no stored summary)", c.Value)
	}
	// c2's roster row leaves summary_score UNSET on both phases: a populated
	// rating beside a blank total is the scale-less-scheme shape.
	if c := c2.Cells["total:pA"]; c.Value != "" {
		t.Errorf("c2 total:pA = %q, want empty (summary_score unset)", c.Value)
	}
	for _, r := range grid.Rows {
		for k := range r.Cells {
			synthetic := strings.HasPrefix(k, ratingKeyPrefix) || strings.HasPrefix(k, totalKeyPrefix)
			known := k == finalColumnKey ||
				k == "rating:pA" || k == "rating:pB" || k == "total:pA" || k == "total:pB"
			if synthetic && !known {
				t.Errorf("unexpected synthetic cell key %q (foreign roster data leaked)", k)
			}
		}
	}

	// Download stamping: the trigger opens the DRAWER (never the CSV directly)
	// on the group-scoped drawer base, carrying this phase's period token for
	// the drawer to pre-select; the final header carries period=final. Aria
	// composed from the label template.
	paActions, ok := pA.Actions.(PhaseActions)
	if !ok {
		t.Fatalf("phase pA Actions = %T, want PhaseActions", pA.Actions)
	}
	wantURL := base + "?scope=all&period=s1"
	if paActions.DownloadDrawerURL != wantURL {
		t.Errorf("pA download URL = %q, want %q", paActions.DownloadDrawerURL, wantURL)
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
	if want := base + "?scope=all&period=final"; finActions.DownloadDrawerURL != want {
		t.Errorf("final download URL = %q, want %q", finActions.DownloadDrawerURL, want)
	}
	if finActions.DownloadTestID != "om-dl-final" {
		t.Errorf("final download testid = %q", finActions.DownloadTestID)
	}
	if !finActions.Any() {
		t.Errorf("final PhaseActions.Any() = false — the header slot would not render")
	}
}

// TestAugmentRatingColumns_TotalPrecedesRating pins the ORDER contract on every
// phase L1: the raw-composite leaf sits between the last criterion column and
// the transmuted rating leaf, never after it and never before the criteria.
func TestAugmentRatingColumns_TotalPrecedesRating(t *testing.T) {
	resp := ratingsResp()
	deps := ratingsDeps(func(ctx context.Context, req *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error) {
		return &matrixpb.GetOutcomeSummaryRosterResponse{
			Rows: []*matrixpb.OutcomeSummaryRosterRow{withTotals(rosterRow("c1", "7", "6", "7"), f64(28), f64(25))},
		}, nil
	})
	grid := buildGrid(context.Background(), deps, types.NewEmptyUserPermissions(), resp, true, "tmpl-1", nil, ratingsViewCtx(), nil)
	augmentRatingColumns(context.Background(), deps, grid, resp, true, "", "all", "")

	for _, phaseID := range []string{"pA", "pB"} {
		var l1 *types.CellGridLevel1
		for i := range grid.Columns {
			if grid.Columns[i].Key == phaseID {
				l1 = &grid.Columns[i]
			}
		}
		if l1 == nil {
			t.Fatalf("phase %s L1 missing", phaseID)
		}
		var order []string
		for _, l2 := range l1.Level2 {
			for _, l3 := range l2.Level3 {
				order = append(order, l3.ColumnKey)
			}
		}
		if len(order) < 3 {
			t.Fatalf("phase %s leaf order = %v, want criteria + total + rating", phaseID, order)
		}
		gotTotal, gotRating := order[len(order)-2], order[len(order)-1]
		if gotTotal != totalKeyPrefix+phaseID || gotRating != ratingKeyPrefix+phaseID {
			t.Errorf("phase %s tail order = [%q %q], want [%s %s]",
				phaseID, gotTotal, gotRating, totalKeyPrefix+phaseID, ratingKeyPrefix+phaseID)
		}
		// And no criterion leaf may follow the summary pair.
		for _, k := range order[:len(order)-2] {
			if strings.HasPrefix(k, totalKeyPrefix) || strings.HasPrefix(k, ratingKeyPrefix) {
				t.Errorf("phase %s: synthetic leaf %q before the criterion leaves (order %v)", phaseID, k, order)
			}
		}
		// L2 keys mirror the leaf keys (the in-template colspan summation).
		if k := l1.Level2[len(l1.Level2)-2].Key; k != totalKeyPrefix+phaseID {
			t.Errorf("phase %s second-to-last L2 key = %q, want %s", phaseID, k, totalKeyPrefix+phaseID)
		}
	}
}

// TestFormatSummaryScore pins the display contract: PRESENCE decides blankness
// (unset ≠ zero), and the shortest round-tripping decimal is emitted verbatim —
// no rounding, no fixed decimal places.
func TestFormatSummaryScore(t *testing.T) {
	cases := []struct {
		name string
		in   *float64
		want string
	}{
		{"unset renders blank", nil, ""},
		{"stored zero renders 0", f64(0), "0"},
		{"integral composite drops the tail", f64(28), "28"},
		{"integral composite (float literal)", f64(28.0), "28"},
		{"top of scale", f64(32), "32"},
		{"half", f64(17.5), "17.5"},
		{"repeating tail verbatim, unrounded", f64(17.333333333333332), "17.333333333333332"},
		{"small fraction verbatim", f64(0.5), "0.5"},
		{"negative verbatim", f64(-3.25), "-3.25"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pe := &matrixpb.OutcomeSummaryPhaseEntry{JobTemplatePhaseId: "pA", SummaryScore: tc.in}
			if got := formatSummaryScore(pe); got != tc.want {
				t.Errorf("formatSummaryScore(%s) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
	if got := formatSummaryScore(nil); got != "" {
		t.Errorf("formatSummaryScore(nil entry) = %q, want empty", got)
	}
}

// TestAugmentRatingColumns_TotalCellValues walks the same presence contract
// through the real cell walk: unset → blank, 0 → "0", 28 → "28", non-integer
// verbatim — beside an always-populated rating from the SAME summary row.
func TestAugmentRatingColumns_TotalCellValues(t *testing.T) {
	resp := ratingsResp()
	resp.Rows = append(resp.Rows, &matrixpb.OutcomeRow{ClientId: "c3"})
	deps := ratingsDeps(func(ctx context.Context, req *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error) {
		return &matrixpb.GetOutcomeSummaryRosterResponse{
			Rows: []*matrixpb.OutcomeSummaryRosterRow{
				// unset total beside a populated rating (scale-less scheme).
				rosterRow("c1", "4", "4", "4"),
				// a real stored zero must not collapse into blankness.
				withTotals(rosterRow("c2", "1", "1", "1"), f64(0), f64(0)),
				withTotals(rosterRow("c3", "7", "5", "6"), f64(28), f64(17.333333333333332)),
			},
		}, nil
	})
	grid := buildGrid(context.Background(), deps, types.NewEmptyUserPermissions(), resp, true, "tmpl-1", nil, ratingsViewCtx(), nil)
	augmentRatingColumns(context.Background(), deps, grid, resp, true, "", "all", "")

	byID := map[string]types.CellGridRow{}
	for _, r := range grid.Rows {
		byID[r.ID] = r
	}
	want := map[string]map[string]string{
		"c1": {"total:pA": "", "rating:pA": "4", "total:pB": "", "rating:pB": "4"},
		"c2": {"total:pA": "0", "rating:pA": "1", "total:pB": "0", "rating:pB": "1"},
		"c3": {"total:pA": "28", "rating:pA": "7", "total:pB": "17.333333333333332", "rating:pB": "5"},
	}
	for rowID, cells := range want {
		row, ok := byID[rowID]
		if !ok {
			t.Fatalf("row %s missing", rowID)
		}
		for key, wantVal := range cells {
			c, ok := row.Cells[key]
			if !ok {
				t.Fatalf("row %s: cell %q missing", rowID, key)
			}
			if c.Value != wantVal {
				t.Errorf("row %s cell %q = %q, want %q", rowID, key, c.Value, wantVal)
			}
			if c.Editable {
				t.Errorf("row %s cell %q is editable — synthetic summary cells are read-only", rowID, key)
			}
		}
	}
}

// TestSummaryColumns_NotInColumnsMenu pins that neither synthetic namespace can
// reach the ?hide= columns menu: page.go builds that menu from a FRESH
// proto-tree buildColumns, never from the augmented grid.
func TestSummaryColumns_NotInColumnsMenu(t *testing.T) {
	resp := ratingsResp()
	deps := ratingsDeps(func(ctx context.Context, req *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error) {
		return &matrixpb.GetOutcomeSummaryRosterResponse{
			Rows: []*matrixpb.OutcomeSummaryRosterRow{withTotals(rosterRow("c1", "7", "6", "7"), f64(28), f64(25))},
		}, nil
	})
	grid := buildGrid(context.Background(), deps, types.NewEmptyUserPermissions(), resp, true, "tmpl-1", nil, ratingsViewCtx(), nil)
	augmentRatingColumns(context.Background(), deps, grid, resp, true, "", "all", "")

	// The exact menu construction page.go performs, AFTER augmentation.
	fullCols := buildColumns(resp.GetPhases(), nil, nil)
	menu, hiddenLeaves := buildColsSelector(fullCols, nil, func(map[string]bool) string { return "/x" })
	if hiddenLeaves != 0 {
		t.Errorf("hidden leaf count = %d, want 0", hiddenLeaves)
	}
	leaves := 0
	for _, g := range menu {
		if strings.HasPrefix(g.Key, totalKeyPrefix) || strings.HasPrefix(g.Key, ratingKeyPrefix) {
			t.Errorf("columns menu group %q is synthetic — must never be offered", g.Key)
		}
		for _, task := range g.Tasks {
			for _, leaf := range task.Leaves {
				leaves++
				if strings.HasPrefix(leaf.ColumnKey, totalKeyPrefix) || strings.HasPrefix(leaf.ColumnKey, ratingKeyPrefix) {
					t.Errorf("columns menu leaf %q is synthetic — must never be offered", leaf.ColumnKey)
				}
			}
		}
	}
	if leaves != 2 {
		t.Errorf("columns menu leaf count = %d, want 2 (the real criteria only)", leaves)
	}
}

// TestSummaryColumns_NotHideable pins that a hand-crafted ?hide=total:… can
// never remove the column: resolveHidden's known set is proto-tree-derived, so
// synthetic tokens are dropped before pruning ever runs.
func TestSummaryColumns_NotHideable(t *testing.T) {
	resp := ratingsResp()
	if h := resolveHidden("total:pA,rating:pA,total:pB,rating:pB,rating:final", resp); h != nil {
		t.Errorf("resolveHidden(synthetic-only) = %v, want nil (all tokens unknown)", h)
	}
	mixed := resolveHidden("total:pA,tA:crA", resp)
	if len(mixed) != 1 || !mixed["tA:crA"] {
		t.Fatalf("resolveHidden(mixed) = %v, want only the real criterion token", mixed)
	}

	// End-to-end: the hostile hide set resolves to nil ⇒ full grid ⇒ the total
	// column is still woven for both phases.
	deps := ratingsDeps(func(ctx context.Context, req *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error) {
		return &matrixpb.GetOutcomeSummaryRosterResponse{
			Rows: []*matrixpb.OutcomeSummaryRosterRow{withTotals(rosterRow("c1", "7", "6", "7"), f64(28), f64(25))},
		}, nil
	})
	hidden := resolveHidden("total:pA,rating:pA", resp)
	grid := buildGrid(context.Background(), deps, types.NewEmptyUserPermissions(), resp, true, "tmpl-1", hidden, ratingsViewCtx(), nil)
	augmentRatingColumns(context.Background(), deps, grid, resp, true, "", "all", "")
	found := map[string]bool{}
	for _, l1 := range grid.Columns {
		for _, l2 := range l1.Level2 {
			for _, l3 := range l2.Level3 {
				found[l3.ColumnKey] = true
			}
		}
	}
	for _, want := range []string{"total:pA", "rating:pA", "total:pB", "rating:pB"} {
		if !found[want] {
			t.Errorf("column %q was hidden by a synthetic ?hide= token", want)
		}
	}

	// Hiding a PHASE prunes it before it can grow either summary leaf.
	hidden = resolveHidden("pA", resp)
	grid = buildGrid(context.Background(), deps, types.NewEmptyUserPermissions(), resp, true, "tmpl-1", hidden, ratingsViewCtx(), nil)
	augmentRatingColumns(context.Background(), deps, grid, resp, true, "", "all", "")
	for _, l1 := range grid.Columns {
		for _, l2 := range l1.Level2 {
			for _, l3 := range l2.Level3 {
				if l3.ColumnKey == "total:pA" || l3.ColumnKey == "rating:pA" {
					t.Errorf("hidden phase pA still grew leaf %q", l3.ColumnKey)
				}
			}
		}
	}
}

// TestAugmentRatingColumns_AutoSaveCoordsUnchanged pins that the augmentation
// cannot disturb the W2 keyboard grid: assignAutoSaveCoords ran inside
// buildGrid, so every REAL cell's coordinate/input identity is byte-identical
// pre- and post-augment, and the synthetic summary cells carry no input
// identity at all (cell-grid.js only walks `.cell-grid-input`).
func TestAugmentRatingColumns_AutoSaveCoordsUnchanged(t *testing.T) {
	resp := ratingsResp()
	cell := func(v float64) *matrixpb.OutcomeCell {
		return &matrixpb.OutcomeCell{
			OutcomeId: "o-" + strconv.FormatFloat(v, 'f', -1, 64),
			JobTaskId: "jt-1", NumericValue: f64(v), Editable: true,
		}
	}
	resp.Rows = []*matrixpb.OutcomeRow{
		{ClientId: "c1", Cells: map[string]*matrixpb.OutcomeCell{"tA:crA": cell(7), "tB:crA": cell(6)}},
		{ClientId: "c2", Cells: map[string]*matrixpb.OutcomeCell{"tA:crA": cell(5), "tB:crA": cell(4)}},
	}
	deps := ratingsDeps(func(ctx context.Context, req *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error) {
		return &matrixpb.GetOutcomeSummaryRosterResponse{
			Rows: []*matrixpb.OutcomeSummaryRosterRow{
				withTotals(rosterRow("c1", "7", "6", "7"), f64(28), f64(25)),
				withTotals(rosterRow("c2", "5", "4", "5"), f64(0), nil),
			},
		}, nil
	})
	// AutoSave (and therefore assignAutoSaveCoords) needs an edit grant.
	perms := types.NewUserPermissions([]string{"task_outcome:update"})
	grid := buildGrid(context.Background(), deps, perms, resp, true, "tmpl-1", nil, ratingsViewCtx(), nil)
	if !grid.AutoSave {
		t.Fatalf("AutoSave off — this test would not exercise the coordinate path")
	}

	// Snapshot every REAL cell, coordinates and input identity included.
	before := map[string]types.CellGridCell{}
	for _, r := range grid.Rows {
		for k, c := range r.Cells {
			before[r.ID+"|"+k] = c
		}
	}
	if len(before) != 4 {
		t.Fatalf("pre-augment cell count = %d, want 4", len(before))
	}

	augmentRatingColumns(context.Background(), deps, grid, resp, true, "", "all", "")

	for _, r := range grid.Rows {
		for k, c := range r.Cells {
			id := r.ID + "|" + k
			prior, existed := before[id]
			if !existed {
				// A synthetic summary cell: it must carry NO input identity, so
				// cell-grid.js (which walks only `.cell-grid-input`) can never
				// pick it up and the keyboard grid never routes into it.
				if !strings.HasPrefix(k, totalKeyPrefix) && !strings.HasPrefix(k, ratingKeyPrefix) {
					t.Errorf("unexpected new cell %q", id)
					continue
				}
				if c.Editable || c.InputID != "" || c.StatusID != "" || c.SavedValue != "" {
					t.Errorf("synthetic cell %q carries input identity: %+v", id, c)
				}
				continue
			}
			if !reflect.DeepEqual(prior, c) {
				t.Errorf("real cell %q changed across augmentation:\n before %+v\n after  %+v", id, prior, c)
			}
		}
	}

	// And the real leaf columns kept their left-to-right indices: the synthetic
	// leaves were appended AFTER coords were assigned, so index 0/1 still name
	// the two criteria.
	for _, r := range grid.Rows {
		if got := r.Cells["tA:crA"].ColIndex; got != 0 {
			t.Errorf("row %s tA:crA ColIndex = %d, want 0", r.ID, got)
		}
		if got := r.Cells["tB:crA"].ColIndex; got != 1 {
			t.Errorf("row %s tB:crA ColIndex = %d, want 1", r.ID, got)
		}
	}
}

// TestAugmentRatingColumns_NoRoster pins the fail-safe: a nil roster closure
// leaves the column tree untouched — NEITHER total NOR rating columns, no final
// — while the download stamping, independent of the composite read, still
// applies.
func TestAugmentRatingColumns_NoRoster(t *testing.T) {
	resp := ratingsResp()
	deps := &PageViewDeps{Labels: outcome_matrix.DefaultLabels(), Routes: outcome_matrix.DefaultRoutes()}
	perms := types.NewEmptyUserPermissions()
	grid := buildGrid(context.Background(), deps, perms, resp, true, "tmpl-1", nil, ratingsViewCtx(), nil)

	augmentRatingColumns(context.Background(), deps, grid, resp, true, "/action/outcome-matrix/tmpl-1/download", "mine", "h1")

	assertNoSummaryColumns(t, grid)
	pa, ok := grid.Columns[0].Actions.(PhaseActions)
	if !ok || pa.DownloadDrawerURL != "/action/outcome-matrix/tmpl-1/download?scope=mine&hide=h1&period=s1" {
		t.Errorf("download stamp = %+v ok=%v", pa, ok)
	}
}

// TestAugmentRatingColumns_RosterError pins the same fail-safe on a FAILING
// roster read (not merely an unwired one): the page renders without either
// summary column, never a 500.
func TestAugmentRatingColumns_RosterError(t *testing.T) {
	resp := ratingsResp()
	deps := ratingsDeps(func(ctx context.Context, req *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error) {
		return nil, errors.New("roster read exploded")
	})
	grid := buildGrid(context.Background(), deps, types.NewEmptyUserPermissions(), resp, true, "tmpl-1", nil, ratingsViewCtx(), nil)

	augmentRatingColumns(context.Background(), deps, grid, resp, true, "", "all", "")

	assertNoSummaryColumns(t, grid)
}

// assertNoSummaryColumns pins the fail-safe shape: the original 2-leaf grid,
// no final L1, and not one synthetic column or cell of either namespace.
func assertNoSummaryColumns(t *testing.T, grid *types.CellGridConfig) {
	t.Helper()
	if got := grid.LeafColumnCount(); got != 2 {
		t.Fatalf("leaf count = %d, want 2 (no total and no rating columns without a roster read)", got)
	}
	if len(grid.Columns) != 2 {
		t.Fatalf("L1 count = %d, want 2 (no final column)", len(grid.Columns))
	}
	for _, l1 := range grid.Columns {
		for _, l2 := range l1.Level2 {
			if strings.HasPrefix(l2.Key, totalKeyPrefix) || strings.HasPrefix(l2.Key, ratingKeyPrefix) {
				t.Errorf("synthetic L2 %q survived a failed roster read", l2.Key)
			}
		}
	}
	for _, r := range grid.Rows {
		for k := range r.Cells {
			if strings.HasPrefix(k, totalKeyPrefix) || strings.HasPrefix(k, ratingKeyPrefix) {
				t.Errorf("synthetic cell %q survived a failed roster read", k)
			}
		}
	}
}

// TestAugmentRatingColumns_NoDrawerBase pins that an unwired drawer route
// yields NO download affordances (a dead icon trigger is worse than none) while
// the summary columns still build from the roster read.
func TestAugmentRatingColumns_NoDrawerBase(t *testing.T) {
	resp := ratingsResp()
	deps := ratingsDeps(func(ctx context.Context, req *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error) {
		return &matrixpb.GetOutcomeSummaryRosterResponse{
			Rows: []*matrixpb.OutcomeSummaryRosterRow{rosterRow("c1", "7", "6", "7")},
		}, nil
	})
	perms := types.NewEmptyUserPermissions()
	grid := buildGrid(context.Background(), deps, perms, resp, true, "tmpl-1", nil, ratingsViewCtx(), nil)

	augmentRatingColumns(context.Background(), deps, grid, resp, true, "", "all", "")

	if got := grid.LeafColumnCount(); got != 7 {
		t.Fatalf("leaf count = %d, want 7", got)
	}
	if pa, ok := grid.Columns[0].Actions.(PhaseActions); ok && pa.DownloadDrawerURL != "" {
		t.Errorf("unexpected download URL with no drawer base: %q", pa.DownloadDrawerURL)
	}
	if fin := grid.Columns[2]; fin.Actions != nil {
		t.Errorf("final column should carry no action payload with no drawer base, got %+v", fin.Actions)
	}
}
