package list

// columns_test.go — the ?hide= column-selector helpers (plan 20260720):
// resolveHidden's unknown-token drop + all-hidden fail-safe, pruneColumns'
// subtree/leaf pruning with empty-parent collapse, hiddenCSV's stable tree
// order, and buildColsSelector's toggle-URL flip semantics.

import (
	"strings"
	"testing"

	"github.com/erniealice/pyeza-golang/types"

	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// twoPhaseResp builds a minimal response tree: phase p1 (task t1: leaves
// t1:c1, t1:c2) + phase p2 (task t2: leaf t2:c1).
func twoPhaseResp() *matrixpb.GetOutcomeMatrixResponse {
	return &matrixpb.GetOutcomeMatrixResponse{
		Phases: []*matrixpb.PhaseColumn{
			{
				JobTemplatePhaseId: "p1",
				Label:              "Phase 1",
				Tasks: []*matrixpb.TaskColumn{{
					JobTemplateTaskId: "t1",
					Label:             "Task 1",
					Criteria: []*matrixpb.CriterionColumn{
						{ColumnKey: "t1:c1"},
						{ColumnKey: "t1:c2"},
					},
				}},
			},
			{
				JobTemplatePhaseId: "p2",
				Label:              "Phase 2",
				Tasks: []*matrixpb.TaskColumn{{
					JobTemplateTaskId: "t2",
					Label:             "Task 2",
					Criteria: []*matrixpb.CriterionColumn{
						{ColumnKey: "t2:c1"},
					},
				}},
			},
		},
	}
}

func TestResolveHidden(t *testing.T) {
	resp := twoPhaseResp()

	t.Run("empty and nil-safe", func(t *testing.T) {
		if got := resolveHidden("", resp); got != nil {
			t.Fatalf("empty raw: want nil, got %v", got)
		}
		if got := resolveHidden("p1", nil); got != nil {
			t.Fatalf("nil resp: want nil, got %v", got)
		}
	})

	t.Run("unknown tokens dropped", func(t *testing.T) {
		got := resolveHidden("p1,ghost,<script>", resp)
		if len(got) != 1 || !got["p1"] {
			t.Fatalf("want {p1}, got %v", got)
		}
	})

	t.Run("only unknown tokens resolves nil", func(t *testing.T) {
		if got := resolveHidden("ghost,phantom", resp); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})

	t.Run("all-hidden fails safe to nil", func(t *testing.T) {
		if got := resolveHidden("p1,p2", resp); got != nil {
			t.Fatalf("hiding every phase must fail safe: got %v", got)
		}
		if got := resolveHidden("p1,t2:c1", resp); got != nil {
			t.Fatalf("phase+last-leaf covering the tree must fail safe: got %v", got)
		}
	})

	t.Run("leaf and phase mix survives when leaves remain", func(t *testing.T) {
		got := resolveHidden("p2,t1:c1", resp)
		if len(got) != 2 || !got["p2"] || !got["t1:c1"] {
			t.Fatalf("want {p2, t1:c1}, got %v", got)
		}
	})
}

func TestPruneColumns(t *testing.T) {
	cols := buildColumns(twoPhaseResp().GetPhases())

	t.Run("no-op on empty set", func(t *testing.T) {
		if got := pruneColumns(cols, nil); len(got) != 2 {
			t.Fatalf("want 2 phases, got %d", len(got))
		}
	})

	t.Run("L1 subtree pruned", func(t *testing.T) {
		got := pruneColumns(cols, map[string]bool{"p2": true})
		if len(got) != 1 || got[0].Key != "p1" {
			t.Fatalf("want [p1], got %+v", got)
		}
	})

	t.Run("leaf pruned, siblings kept", func(t *testing.T) {
		got := pruneColumns(cols, map[string]bool{"t1:c1": true})
		if len(got) != 2 || len(got[0].Level2[0].Level3) != 1 || got[0].Level2[0].Level3[0].ColumnKey != "t1:c2" {
			t.Fatalf("want p1 to keep only t1:c2, got %+v", got)
		}
	})

	t.Run("emptied L2 and L1 collapse", func(t *testing.T) {
		got := pruneColumns(cols, map[string]bool{"t2:c1": true})
		if len(got) != 1 || got[0].Key != "p1" {
			t.Fatalf("p2's only leaf hidden must drop p2 entirely, got %+v", got)
		}
	})

	t.Run("colspan derivations follow the pruned tree", func(t *testing.T) {
		cfg := types.CellGridConfig{Columns: pruneColumns(cols, map[string]bool{"p2": true})}
		if n := cfg.LeafColumnCount(); n != 2 {
			t.Fatalf("want 2 leaves after pruning p2, got %d", n)
		}
		if n := cfg.BandColSpan(); n != 3 {
			t.Fatalf("want band colspan 3 (2 leaves + row head), got %d", n)
		}
	})
}

func TestHiddenCSVStableOrder(t *testing.T) {
	phases := twoPhaseResp().GetPhases()
	// Map iteration order must not leak into the URL: tree order is
	// p1, t1:c1, t1:c2, p2, t2:c1.
	got := hiddenCSV(map[string]bool{"t2:c1": true, "p1": true, "t1:c2": true}, phases)
	if got != "p1,t1:c2,t2:c1" {
		t.Fatalf("want stable tree order p1,t1:c2,t2:c1 — got %q", got)
	}
	if hiddenCSV(nil, phases) != "" {
		t.Fatal("empty set must serialize to empty string")
	}
}

func TestBuildColsSelector(t *testing.T) {
	full := buildColumns(twoPhaseResp().GetPhases())
	hidden := map[string]bool{"p2": true, "t1:c1": true}
	urlFor := func(h map[string]bool) string {
		return "?hide=" + hiddenCSV(h, twoPhaseResp().GetPhases())
	}

	groups, hiddenLeaves := buildColsSelector(full, hidden, urlFor)
	if len(groups) != 2 {
		t.Fatalf("selector must list the FULL tree, got %d groups", len(groups))
	}
	// 1 leaf hidden individually + 1 leaf under the hidden p2 subtree.
	if hiddenLeaves != 2 {
		t.Fatalf("want 2 effectively hidden leaves, got %d", hiddenLeaves)
	}

	p1, p2 := groups[0], groups[1]
	if p1.Hidden || !p2.Hidden {
		t.Fatalf("hidden flags wrong: p1=%v p2=%v", p1.Hidden, p2.Hidden)
	}
	// Toggling visible p1 ADDS its token, keeping the rest.
	if !strings.Contains(p1.ToggleURL, "p1") || !strings.Contains(p1.ToggleURL, "p2") {
		t.Fatalf("p1 toggle must add p1 alongside p2: %q", p1.ToggleURL)
	}
	// Toggling hidden p2 REMOVES its token but keeps the individually hidden
	// leaf, so re-showing the phase restores the finer state.
	if strings.Contains(p2.ToggleURL, "p2") || !strings.Contains(p2.ToggleURL, "t1:c1") {
		t.Fatalf("p2 toggle must drop p2 and keep t1:c1: %q", p2.ToggleURL)
	}
	// Leaf flags: t1:c1 hidden individually, t1:c2 visible.
	leaves := p1.Tasks[0].Leaves
	if !leaves[0].Hidden || leaves[1].Hidden {
		t.Fatalf("leaf hidden flags wrong: %+v", leaves)
	}
	// Toggling the hidden leaf removes exactly its own token.
	if strings.Contains(leaves[0].ToggleURL, "t1:c1") || !strings.Contains(leaves[0].ToggleURL, "p2") {
		t.Fatalf("t1:c1 toggle must drop itself and keep p2: %q", leaves[0].ToggleURL)
	}
}

func TestResolveHiddenInputCaps(t *testing.T) {
	resp := twoPhaseResp()
	// A hostile mega-string must neither blow up nor bypass the known-token
	// filter; oversized input degrades (possibly to nil), never errors.
	huge := strings.Repeat("ghost,", 5000) + "p1"
	if got := resolveHidden(huge, resp); got != nil {
		t.Fatalf("token-capped hostile input should resolve nil (valid token past cap), got %v", got)
	}
	// Within caps, valid tokens still resolve.
	if got := resolveHidden("p1", resp); len(got) != 1 || !got["p1"] {
		t.Fatalf("sanity: want {p1}, got %v", got)
	}
}

func TestCSVSafe(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"plain", "plain"},
		{"=cmd|calc", "\t=cmd|calc"},
		{"+1", "\t+1"},
		{"-1", "\t-1"},
		{"@x", "\t@x"},
		{"\tx", "\t\tx"},
		{"\rx", "\t\rx"},
		{"\nx", "\t\nx"},
		{"＝FW equals", "\t＝FW equals"},
		{"＋FW plus", "\t＋FW plus"},
		{"－FW minus", "\t－FW minus"},
		{"＠FW at", "\t＠FW at"},
		{"7", "7"},
	}
	for _, c := range cases {
		if got := csvSafe(c.in); got != c.want {
			t.Errorf("csvSafe(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
