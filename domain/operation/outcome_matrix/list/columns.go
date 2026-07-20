package list

// columns.go — the ?hide= columns-selector layer (plan 20260720 Q1/Q2:
// server-prune, URL-canonical). Split out of page.go for the fayna
// placement gate (godFileThreshold 1200): the selector menu model
// (ColsGroup/ColsTask/ColsLeaf), the hide-set resolution + fail-safe
// (resolveHidden), the deterministic URL serialization (hiddenCSV), the
// render-time tree pruning (pruneColumns), and the menu builder
// (buildColsSelector). All pure functions over the response tree — the
// record action's save authority never sees any of this.

import (
	"strings"

	"github.com/erniealice/pyeza-golang/types"

	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// ColsGroup is one L1 (phase) entry in the columns-selector menu.
type ColsGroup struct {
	Key       string
	Label     string
	Hidden    bool
	ToggleURL string
	Tasks     []ColsTask
}

// ColsTask is an L2 heading grouping its leaf entries (not itself toggleable —
// L2 granularity is deliberately out of scope; hide the leaves or the phase).
type ColsTask struct {
	Label  string
	Leaves []ColsLeaf
}

// ColsLeaf is one leaf (criterion) entry in the columns-selector menu.
type ColsLeaf struct {
	ColumnKey string
	Slug      string // testid-safe form of ColumnKey
	Label     string
	Hidden    bool
	ToggleURL string
}

// resolveHidden parses ?hide= (comma-separated L1 phase ids and/or leaf
// ColumnKeys) against the response tree. Unknown tokens are dropped (no URL
// junk accumulation, nothing attacker-controlled echoes into hrefs beyond
// known ids), and a set that would hide EVERY leaf resolves to nil — the grid
// fails safe to fully visible, never to an empty spreadsheet. Render-only:
// the record action's save authority re-derives the full matrix per POST and
// never sees this.
//
// Input caps (codex 20260720 #6): the raw param is bounded by bytes AND token
// count before any per-token work, so a hostile query string cannot amplify —
// the honest maximum is one token per tree node (a few dozen), far under
// either cap. Truncation may split a trailing token; the known-set filter
// discards the fragment.
func resolveHidden(raw string, resp *matrixpb.GetOutcomeMatrixResponse) map[string]bool {
	const (
		maxHideBytes  = 4096
		maxHideTokens = 256
	)
	if raw == "" || resp == nil {
		return nil
	}
	if len(raw) > maxHideBytes {
		raw = raw[:maxHideBytes]
	}
	known := map[string]bool{}
	for _, ph := range resp.GetPhases() {
		known[ph.GetJobTemplatePhaseId()] = true
		for _, tk := range ph.GetTasks() {
			for _, cr := range tk.GetCriteria() {
				known[cr.GetColumnKey()] = true
			}
		}
	}
	hidden := map[string]bool{}
	tokens := strings.Split(raw, ",")
	if len(tokens) > maxHideTokens {
		tokens = tokens[:maxHideTokens]
	}
	for _, tok := range tokens {
		tok = strings.TrimSpace(tok)
		if tok != "" && known[tok] {
			hidden[tok] = true
		}
	}
	if len(hidden) == 0 {
		return nil
	}
	visible := 0
	for _, ph := range resp.GetPhases() {
		if hidden[ph.GetJobTemplatePhaseId()] {
			continue
		}
		for _, tk := range ph.GetTasks() {
			for _, cr := range tk.GetCriteria() {
				if !hidden[cr.GetColumnKey()] {
					visible++
				}
			}
		}
	}
	if visible == 0 {
		return nil
	}
	return hidden
}

// hiddenCSV serializes a hidden set in stable tree order (deterministic URLs
// regardless of map iteration).
func hiddenCSV(hidden map[string]bool, phases []*matrixpb.PhaseColumn) string {
	if len(hidden) == 0 {
		return ""
	}
	parts := make([]string, 0, len(hidden))
	for _, ph := range phases {
		if hidden[ph.GetJobTemplatePhaseId()] {
			parts = append(parts, ph.GetJobTemplatePhaseId())
		}
		for _, tk := range ph.GetTasks() {
			for _, cr := range tk.GetCriteria() {
				if hidden[cr.GetColumnKey()] {
					parts = append(parts, cr.GetColumnKey())
				}
			}
		}
	}
	return strings.Join(parts, ",")
}

// pruneColumns returns the tree minus hidden L1 subtrees and hidden leaves,
// dropping any L2/L1 emptied by leaf pruning. A no-op on an empty set.
func pruneColumns(cols []types.CellGridLevel1, hidden map[string]bool) []types.CellGridLevel1 {
	if len(hidden) == 0 {
		return cols
	}
	out := make([]types.CellGridLevel1, 0, len(cols))
	for _, l1 := range cols {
		if hidden[l1.Key] {
			continue
		}
		p1 := types.CellGridLevel1{Key: l1.Key, Label: l1.Label}
		for _, l2 := range l1.Level2 {
			p2 := types.CellGridLevel2{Key: l2.Key, Label: l2.Label}
			for _, l3 := range l2.Level3 {
				if hidden[l3.ColumnKey] {
					continue
				}
				p2.Level3 = append(p2.Level3, l3)
			}
			if len(p2.Level3) > 0 {
				p1.Level2 = append(p1.Level2, p2)
			}
		}
		if len(p1.Level2) > 0 {
			out = append(out, p1)
		}
	}
	return out
}

// buildColsSelector maps the FULL column tree into the selector menu model:
// per-node hidden state + a precomputed toggle URL that flips exactly that
// node's token in the hide set. Leaves under a hidden L1 are not individually
// toggleable (the template collapses them); their individual hide tokens, if
// any, survive the L1 toggle so re-showing the phase restores the finer state.
// Returns the menu and the count of effectively hidden leaves.
func buildColsSelector(full []types.CellGridLevel1, hidden map[string]bool, urlFor func(map[string]bool) string) ([]ColsGroup, int) {
	toggled := func(tok string) map[string]bool {
		h := make(map[string]bool, len(hidden)+1)
		for k := range hidden {
			h[k] = true
		}
		if h[tok] {
			delete(h, tok)
		} else {
			h[tok] = true
		}
		return h
	}
	hiddenLeaves := 0
	groups := make([]ColsGroup, 0, len(full))
	for _, l1 := range full {
		g := ColsGroup{
			Key:       l1.Key,
			Label:     l1.Label,
			Hidden:    hidden[l1.Key],
			ToggleURL: urlFor(toggled(l1.Key)),
		}
		for _, l2 := range l1.Level2 {
			t := ColsTask{Label: l2.Label}
			for _, l3 := range l2.Level3 {
				if g.Hidden || hidden[l3.ColumnKey] {
					hiddenLeaves++
				}
				t.Leaves = append(t.Leaves, ColsLeaf{
					ColumnKey: l3.ColumnKey,
					Slug:      slug(l3.ColumnKey),
					Label:     l3.Label,
					Hidden:    hidden[l3.ColumnKey],
					ToggleURL: urlFor(toggled(l3.ColumnKey)),
				})
			}
			g.Tasks = append(g.Tasks, t)
		}
		groups = append(groups, g)
	}
	return groups, hiddenLeaves
}
