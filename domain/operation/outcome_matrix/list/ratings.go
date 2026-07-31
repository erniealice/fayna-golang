package list

// ratings.go — the derived read-only SUMMARY columns (20260725/20260729): a
// per-phase TOTAL + RATING pair appended under each phase header, plus one
// whole-row Final column at the far right, plus the per-column download
// affordances stamped onto the L1 header action slots.
//
// DATA SOURCE. Everything renders STORED values read VERBATIM from the P2
// roster composite read (GetOutcomeSummaryRoster — per phase both
// phase_outcome_summary.summary_score (the raw composite the scoring scheme
// combined, pre-transmutation) and phase_outcome_summary.scaled_label (that
// composite after band transmutation), and job_outcome_summary.scaled_label
// for the year final). Nothing here computes, derives, rounds or re-aggregates;
// ComputePhaseOutcome / ComputeJobOutcome own the write path and are not
// consulted. The total and the rating come off the SAME summary row written by
// ONE upsert, so they can never be mutually inconsistent — which is also why
// the existing freshness/ack machinery needs no extension: one row, one
// staleness (the stale indicator attaches to the rating cell of the same row).
//
// ABSENCE vs ZERO (tandem-lock contract). summary_score is presence-tracked
// (optional double): UNSET renders blank ("—"), a stored 0 renders "0". A
// scheme with no transmutation scale stores a label but no composite, so its
// phases render a populated Rating beside a blank Total — presence degrades
// through the NULL-ness of DATA, never through a category or business-type
// check in code.
//
// COLUMN-TREE DESIGN (deliberate; the alternative was a first-class pyeza
// "trailer column" notion). The rating columns are SYNTHETIC LEAVES woven into
// the existing 3-level tree in this CONSUMER:
//   - a phase summary = two trailing L2+L3 pairs appended to that phase's L1
//     (total first, then rating — raw composite before transmuted output), so
//     the L1 colspan (summed in-template from len(Level3)) grows by exactly
//     two and every colspan/BandColSpan/LeafColumnCount derivation stays
//     correct BY CONSTRUCTION;
//   - the Final column = one synthetic trailing L1 ("rating-final") with a
//     single L2+L3, whose header carries the period=final download slot.
// A pyeza trailer-column concept was REJECTED: it would teach the generic
// component a second column axis (new struct + head/body template surgery +
// a second colspan path to keep correct) for something the existing tree
// already expresses, and the L1-slot download button would still need the
// tree form anyway.
//
// WHY THE OTHER WALKERS STAY CORRECT:
//   - pruneColumns / resolveHidden run BEFORE this augmentation (inside
//     buildGrid), and resolveHidden's known-token set comes from the proto
//     response tree — a synthetic "rating:*"/"total:*" key can never be hidden
//     and never appears in the columns menu (page.go builds that menu from a
//     fresh proto-tree buildColumns, not from this augmented grid), and a
//     phase hidden via ?hide= is pruned before it can grow either leaf.
//   - assignAutoSaveCoords also ran inside buildGrid, so the W2 keyboard
//     coordinates are IDENTICAL to the pre-augment grid; total and rating cells
//     render as plain read-only value spans (Editable=false), never inputs, and
//     cell-grid.js only ever walks `.cell-grid-input` elements.
//   - the CSV export path (export.go) builds its own grid and never calls
//     this augmentation, so the exported grid CSV is byte-identical to before;
//     the composite values already have their own first-class export
//     (period=final), which is exactly what the Final header's button
//     pre-selects in the download drawer.
//   - the record action re-derives the full matrix per POST and never sees
//     render-side columns; rating cells emit no inputs, so the batch/auto
//     save POST body is unchanged.
//
// ROSTER FILTER (hard constraint). GetOutcomeSummaryRosterRequest carries ONLY
// job_template_id + scope — NO group narrowing — so on a group-scoped page it
// returns every student under the template (87 for the AY2026-27 grade-scoped
// templates), not the rendered roster (29). The filter is the GRID ITSELF:
// cells are attached by iterating cfg.Rows (the rendered, already group-scoped
// roster) and looking each row's client up in the roster response; roster rows
// with no rendered grid row are simply never read. No proto change.

import (
	"context"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"
	"github.com/erniealice/pyeza-golang/types"

	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

const (
	// ratingKeyPrefix namespaces the synthetic ColumnKeys ("rating:{phase_id}").
	// Collision-proof against real cell keys, which are always
	// "{job_template_task_id}:{outcome_criteria_id}" (two UUIDs).
	ratingKeyPrefix = "rating:"
	// totalKeyPrefix namespaces the synthetic ColumnKeys of the RAW COMPOSITE
	// leaf ("total:{phase_id}") that precedes each phase's rating leaf. Same
	// collision-proofness as ratingKeyPrefix: real cell keys are always two
	// UUIDs, and the ?hide= axis only ever knows phase ids and real cell keys.
	totalKeyPrefix = "total:"
	// finalColumnKey / finalL1Key address the trailing whole-row column.
	finalColumnKey = "rating:final"
	finalL1Key     = "rating-final"
	finalSlug      = "final"
)

// augmentRatingColumns appends the derived rating columns and stamps the
// per-column download affordances onto the already-built grid. View path only
// (NewView) — the export handler's buildGrid never calls it. Fail-safe: a nil
// or failed roster read leaves the column tree exactly as built (downloads are
// still stamped — the drawer works independently of the composite read).
//
// drawerBase is the RESOLVED download-drawer GET path (group-scoped when the
// page is); each column's trigger appends the live ?scope=/?hide= pair plus its
// own ?period= token, which the drawer pre-selects.
func augmentRatingColumns(
	ctx context.Context,
	deps *PageViewDeps,
	cfg *types.CellGridConfig,
	resp *matrixpb.GetOutcomeMatrixResponse,
	effectiveAll bool,
	drawerBase, scopeActive, hideCSV string,
) {
	l := deps.Labels

	codeByPhase := map[string]string{}
	labelByPhase := map[string]string{}
	for _, ph := range resp.GetPhases() {
		codeByPhase[ph.GetJobTemplatePhaseId()] = ph.GetCode()
		labelByPhase[ph.GetJobTemplatePhaseId()] = ph.GetLabel()
	}

	// (a) Per-phase download buttons — an icon-only trigger in each phase's L1
	// header slot opening the export drawer with THAT phase's period token
	// pre-selected, so the operator still chooses CSV vs PDF (20260726).
	// Phases with no stable code get no button: an empty period token would
	// pre-select "All periods" under a per-phase affordance — a mislabelled
	// control, worse than none. It is also exactly the token set the drawer's
	// own select and export.go's periodKnown guard recognise (phase CODE, never
	// the phase id — ?hide= is the id-keyed axis), so the two agree by
	// construction.
	if drawerBase != "" {
		for i := range cfg.Columns {
			l1 := &cfg.Columns[i]
			code := codeByPhase[l1.Key]
			if code == "" {
				continue
			}
			stampDownload(l1, slug(l1.Key), l.Approval,
				drawerPeriodURL(drawerBase, scopeActive, hideCSV, code),
				subColumn(l.Export.DownloadAria, labelByPhase[l1.Key]),
				"om-dl-"+slug(code))
		}
	}

	// (b) The rating columns. Stored composites, read verbatim; no roster read
	// ⇒ no rating columns (the grid renders exactly as before).
	roster := fetchRatingRoster(ctx, deps, resp.GetJobTemplateId(), effectiveAll)
	if roster == nil {
		return
	}
	byClient := make(map[string]*matrixpb.OutcomeSummaryRosterRow, len(roster.GetRows()))
	for _, row := range roster.GetRows() {
		byClient[row.GetClientId()] = row
	}

	// Trailing summary leaves per (visible) phase L1: TOTAL then RATING, in that
	// fixed order — the raw composite sits between the last criterion column and
	// the transmuted output it was transmuted into. testKey prefers the phase
	// CODE ("s1") for a stable, readable testid; slug(id) is the fallback.
	testKeyByPhase := map[string]string{}
	for i := range cfg.Columns {
		l1 := &cfg.Columns[i]
		if _, isPhase := codeByPhase[l1.Key]; !isPhase {
			continue
		}
		tk := codeByPhase[l1.Key]
		if tk == "" {
			tk = slug(l1.Key)
		}
		testKeyByPhase[l1.Key] = slug(tk)
		l1.Level2 = append(l1.Level2,
			types.CellGridLevel2{
				Key: totalKeyPrefix + l1.Key,
				Level3: []types.CellGridLevel3{{
					ColumnKey: totalKeyPrefix + l1.Key,
					Label:     l.Grid.TotalColumn,
					CellInput: types.CellInputDescriptor{Type: "text"},
				}},
			},
			types.CellGridLevel2{
				Key: ratingKeyPrefix + l1.Key,
				Level3: []types.CellGridLevel3{{
					ColumnKey: ratingKeyPrefix + l1.Key,
					Label:     l.Grid.RatingColumn,
					CellInput: types.CellInputDescriptor{Type: "text"},
				}},
			})
	}

	// Trailing whole-row Final column, its header carrying the period=final
	// download slot. Rendered through the SAME L1ActionsTemplate as the phase
	// controls — the payload's approval gates are all zero, so only the
	// download anchor renders.
	finalActions := PhaseActions{Phase: ApprovalPhase{Slug: finalSlug}, Labels: l.Approval}
	if drawerBase != "" {
		finalActions.DownloadDrawerURL = drawerPeriodURL(drawerBase, scopeActive, hideCSV, "final")
		finalActions.DownloadAria = subColumn(l.Export.DownloadAria, l.Export.PeriodFinal)
		finalActions.DownloadTestID = "om-dl-final"
	}
	var finalPayload any
	if finalActions.Any() {
		finalPayload = finalActions
	}
	cfg.Columns = append(cfg.Columns, types.CellGridLevel1{
		Key:     finalL1Key,
		Label:   l.Export.PeriodFinal,
		Actions: finalPayload,
		Level2: []types.CellGridLevel2{{
			Key: finalColumnKey,
			Level3: []types.CellGridLevel3{{
				ColumnKey: finalColumnKey,
				Label:     l.Grid.RatingColumn,
				CellInput: types.CellInputDescriptor{Type: "text"},
			}},
		}},
	})

	// Cells — attached by walking the RENDERED grid rows (the group-scoped
	// roster filter, see the file comment). A client with no roster row, or a
	// phase with no stored summary ("" scaled_label / unset summary_score),
	// renders the standard "—".
	for i := range cfg.Rows {
		row := &cfg.Rows[i]
		rr := byClient[row.ID]
		if rr == nil {
			continue
		}
		for _, pe := range rr.GetPhases() {
			phaseID := pe.GetJobTemplatePhaseId()
			tk, isCol := testKeyByPhase[phaseID]
			if !isCol {
				continue // phase pruned from this view (or unknown) — no column
			}
			row.Cells[totalKeyPrefix+phaseID] = ratingCell(
				formatSummaryScore(pe), l.Grid.TotalTooltip, "om-total-"+row.ID+"-"+tk)
			row.Cells[ratingKeyPrefix+phaseID] = ratingCell(
				pe.GetScaledLabel(), l.Grid.RatingTooltip, "om-rating-"+row.ID+"-"+tk)
		}
		row.Cells[finalColumnKey] = ratingCell(
			rr.GetYearFinalLabel(), l.Grid.RatingTooltip, "om-rating-"+row.ID+"-final")
	}
}

// fetchRatingRoster wraps the P2 composite read, threading the SAME resolved
// MINE/ALL scope the matrix read used (a MINE-scoped teacher never receives
// the full-workspace composite roster). Nil-safe and fail-safe: any miss
// returns nil and the page renders without rating columns, never a 500.
func fetchRatingRoster(ctx context.Context, deps *PageViewDeps, templateID string, effectiveAll bool) *matrixpb.GetOutcomeSummaryRosterResponse {
	if deps.GetOutcomeSummaryRoster == nil || templateID == "" {
		return nil
	}
	scope := matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_MINE
	if effectiveAll {
		scope = matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL
	}
	roster, err := deps.GetOutcomeSummaryRoster(ctx, &matrixpb.GetOutcomeSummaryRosterRequest{
		JobTemplateId: templateID,
		Scope:         scope,
	})
	if err != nil || roster == nil {
		if err != nil {
			log.Printf("outcome matrix: summary roster read failed (rating columns omitted): %v", err)
		}
		return nil
	}
	return roster
}

// formatSummaryScore renders the stored raw composite for display. PRESENCE,
// not value, decides blankness: an UNSET summary_score returns "" (the
// component renders the standard "—"), while a stored 0 returns "0" — the two
// are semantically different (nothing computed vs a computed zero) and must
// never collapse into each other.
//
// 'f' with precision -1 emits the shortest round-tripping decimal, so an
// integral composite prints without a decimal tail (28 → "28") and a
// non-integral one prints verbatim, with no rounding and no decimal-places
// assumption. Rounding is a write-path duty of the scoring scheme, never a
// display decision taken here.
func formatSummaryScore(pe *matrixpb.OutcomeSummaryPhaseEntry) string {
	if pe == nil || pe.SummaryScore == nil {
		return ""
	}
	return strconv.FormatFloat(pe.GetSummaryScore(), 'f', -1, 64)
}

// ratingCell builds one derived read-only cell. Editable=false + no ReadOnly
// flag: the component's $showRO derives read-only presentation from the value
// itself, and an empty value renders the standard "—" with no tooltip.
func ratingCell(value, tooltip, testID string) types.CellGridCell {
	return types.CellGridCell{
		Value:           value,
		Editable:        false,
		ReadOnlyTooltip: tooltip,
		TestID:          testID,
	}
}

// stampDownload attaches the download affordance to an L1 header's action
// slot, preserving an existing approval payload (the two render from ONE
// PhaseActions so the header cell stays a single slot).
func stampDownload(l1 *types.CellGridLevel1, phaseSlug string, al outcome_matrix.ApprovalLabels, dlURL, aria, testID string) {
	if pa, ok := l1.Actions.(PhaseActions); ok {
		pa.DownloadDrawerURL, pa.DownloadAria, pa.DownloadTestID = dlURL, aria, testID
		l1.Actions = pa
		return
	}
	l1.Actions = PhaseActions{
		Phase:             ApprovalPhase{Slug: phaseSlug},
		Labels:            al,
		DownloadDrawerURL: dlURL,
		DownloadAria:      aria,
		DownloadTestID:    testID,
	}
}

// drawerPeriodURL composes the download-drawer GET target: the resolved base
// (group-scoped when the page is), the live ?scope=/?hide= pair the page's
// other navigation URLs carry, and the period token the drawer pre-selects.
func drawerPeriodURL(base, scope, hide, period string) string {
	u := base + "?scope=" + scope
	if hide != "" {
		u += "&hide=" + hide
	}
	return u + "&period=" + url.QueryEscape(period)
}

// subColumn substitutes the "{column}" placeholder (once) — the download
// aria template's only token.
func subColumn(tmpl, column string) string {
	return strings.Replace(tmpl, "{column}", column, 1)
}
