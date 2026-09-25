package action

// rating_description.go — the rating-description → outcome narrative bridge
// (owner decision 2026-09-21; resolver cut-over 2026-09-25, schema-proposal
// §4/§5, interfaces.md §2).
//
// A template_task_criteria binding in RATING_MODE_NUMERIC_WITH_DESCRIPTION
// carries per-band wording sourced from a rating_description_set LINKED to the
// cell's offering (product_plan) × academic year (price_schedule). When a
// grader records a numeric value for such a cell, the matching description is
// SNAPSHOTTED into the outcome's determination_note — the one canonical
// narrative field the per-cell drawer already edits — so the grid cell stays a
// bare number and the wording travels with the recorded grade (a later edit of
// the set's entry, or a relink to a different set, never rewrites history).
//
// Q18 (no legacy fallback): the description text for a description-mode cell
// comes ONLY from Deps.ResolveCellRatingDescriptions, batched once per save
// request over every description-mode cell's SERVER-DERIVED identity (job_id,
// job_task_id, outcome_criteria_id — the POST body never supplies it). The
// column's own (legacy per-slot) RatingDescriptions are never read on this
// write path any more.

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// ratingMatcher is one resolved band: an exact numeric match or a half-open
// [min, max) range, plus the set entry's wording for it.
type ratingMatcher struct {
	kind        enums.ScaleKind
	min, max    *float64
	match       *float64
	description string
}

// ratingDescriptions is the resolved description set for one cell. Empty means
// either the cell is not in description mode, or the resolver returned no
// usable entry for the value at hand (a RESOLVED status with no matching band,
// or NO_LINK) — in every such case describe() returns "" ("no entry", Q20/Q21).
type ratingDescriptions []ratingMatcher

// ratingDescriptionsFromList projects a resolver response's per-cell
// descriptions (interfaces.md §2 CellRatingResolution.descriptions — the same
// RatingDescription message the matrix column used to carry) into matchers.
// Blank wording and unsupported scale kinds are dropped (never guessed).
func ratingDescriptionsFromList(descs []*matrixpb.RatingDescription) ratingDescriptions {
	out := make(ratingDescriptions, 0, len(descs))
	for _, d := range descs {
		if d == nil {
			continue
		}
		text := strings.TrimSpace(d.GetDescription())
		if text == "" {
			continue
		}
		m := ratingMatcher{kind: d.GetScaleKind(), min: d.InputMin, max: d.InputMax, description: text}
		switch d.GetScaleKind() {
		case enums.ScaleKind_SCALE_KIND_EXACT_MAP:
			raw := strings.TrimSpace(d.GetInputMatch())
			v, err := strconv.ParseFloat(raw, 64)
			if raw == "" || err != nil {
				continue
			}
			m.match = &v
		case enums.ScaleKind_SCALE_KIND_RANGE_MAP:
		default:
			continue
		}
		out = append(out, m)
	}
	return out
}

// describe returns the description for v: an exact match wins, then the first
// half-open range containing v. "" when nothing matches (an intentional
// absence, e.g. level 0 — Q20).
func (r ratingDescriptions) describe(v float64) string {
	for _, m := range r {
		if m.kind == enums.ScaleKind_SCALE_KIND_EXACT_MAP && m.match != nil && math.Abs(*m.match-v) < 1e-9 {
			return m.description
		}
	}
	for _, m := range r {
		if m.kind != enums.ScaleKind_SCALE_KIND_RANGE_MAP {
			continue
		}
		if (m.min == nil || v >= *m.min) && (m.max == nil || v < *m.max) {
			return m.description
		}
	}
	return ""
}

// ── batch resolution (Q18: the resolver is the ONLY source) ────────────────

// ratingCellKey is the full cell identity the resolver addresses by
// (schema-proposal §4; interfaces.md §2 CellRatingRef) — job_id, job_task_id
// and outcome_criteria_id, all read from the server-derived matrix (srvCell),
// never the POST body.
type ratingCellKey struct {
	jobID, jobTaskID, criteriaID string
}

// ratingResolution is one cell's resolved outcome, folded into the shape
// updateCell/createCell need to apply the note rules (schema-proposal §5). A
// nil *ratingResolution means the column is not in description mode: the
// legacy no-touch behaviour applies — the note is never read or written by
// this bridge at all.
type ratingResolution struct {
	// reject is true for UNRESOLVED_IDENTITY / AMBIGUOUS / INVALID_CONFIG /
	// PLACEHOLDER_UNRESOLVED, a
	// batch/transport error, an unwired resolver, or a resolver response that
	// omits or duplicates a requested cell (Astra r5 / RD-66). The cell's save
	// is REJECTED entirely — no write is attempted, value and note stay
	// exactly as stored (Q22).
	reject bool
	// reason is the bounded, non-value-echoing ack error code for a rejected
	// cell (lyngua outcome_matrix.errors.rating_description_*, interfaces.md
	// §6).
	reason string
	// logMsg is the full (still non-sensitive) diagnostic for the server log
	// only — never surfaced to the client ack.
	logMsg string
	// ratings is the matcher over the resolved descriptions. Empty for a
	// RESOLVED status with no matching entry, or for NO_LINK (Q21: both behave
	// as "no entry" — describe() returns "" for every value).
	ratings ratingDescriptions
}

// resolveRatingDescriptions batches every description-mode cell in cells into
// ONE Deps.ResolveCellRatingDescriptions call (Q18/RD-66) and returns each
// cell's resolution keyed by its full server-derived identity. Cells whose
// column is NOT in description mode are skipped entirely — no ref is built and
// no key is present in the result; callers must treat a missing key as "not
// applicable" (never as a rejection).
//
// An unwired resolver, a transport/decode error, or a resolver response
// missing/duplicating a requested cell rejects EVERY description-mode cell in
// THIS batch: never silently downgrade to "no entry" when the resolver itself
// could not be trusted (Q22-class fail-closed behaviour extended to
// infrastructure failure, not just a per-cell integrity status).
func resolveRatingDescriptions(ctx context.Context, fn func(context.Context, *matrixpb.ResolveCellRatingDescriptionsRequest) (*matrixpb.ResolveCellRatingDescriptionsResponse, error), cells []srvCell) map[ratingCellKey]ratingResolution {
	out := make(map[ratingCellKey]ratingResolution)

	seen := make(map[ratingCellKey]bool)
	var refs []*matrixpb.CellRatingRef
	for _, sc := range cells {
		if !sc.descriptionMode {
			continue
		}
		k := ratingCellKey{sc.jobID, sc.jobTaskID, sc.criteriaID}
		if seen[k] {
			continue
		}
		seen[k] = true
		refs = append(refs, &matrixpb.CellRatingRef{
			JobId:             sc.jobID,
			JobTaskId:         sc.jobTaskID,
			OutcomeCriteriaId: sc.criteriaID,
		})
	}
	if len(refs) == 0 {
		return out
	}

	rejectAll := func(reason, logMsg string) map[ratingCellKey]ratingResolution {
		for k := range seen {
			out[k] = ratingResolution{reject: true, reason: reason, logMsg: logMsg}
		}
		return out
	}

	if fn == nil {
		log.Printf("[outcome-matrix] rating resolution rejected %d cell(s): ResolveCellRatingDescriptions is unwired", len(refs))
		return rejectAll("rating_description_resolution_failed", "ResolveCellRatingDescriptions is unwired")
	}

	resp, err := fn(ctx, &matrixpb.ResolveCellRatingDescriptionsRequest{Cells: refs})
	if err != nil || resp == nil || !resp.GetSuccess() {
		log.Printf("[outcome-matrix] rating resolution batch failed for %d cell(s): %v", len(refs), err)
		return rejectAll("rating_description_resolution_failed", fmt.Sprintf("batch error: %v", err))
	}

	got := make(map[ratingCellKey]bool, len(resp.GetResults()))
	dup := make(map[ratingCellKey]bool)
	for _, res := range resp.GetResults() {
		ref := res.GetCell()
		if ref == nil {
			continue
		}
		k := ratingCellKey{ref.GetJobId(), ref.GetJobTaskId(), ref.GetOutcomeCriteriaId()}
		if !seen[k] {
			// The resolver returned an address we never asked for — never
			// trust it (Astra r5 / RD-66); it addresses no pending write.
			log.Printf("[outcome-matrix] rating resolution: unexpected result for job=%s task=%s criteria=%s", k.jobID, k.jobTaskID, k.criteriaID)
			continue
		}
		if got[k] {
			// The same requested address came back twice: which result is
			// authoritative is unknowable, so the cell is REJECTED (fail
			// closed — codex impl1 #7), never "first result wins".
			log.Printf("[outcome-matrix] rating resolution: duplicate result for job=%s task=%s criteria=%s — cell rejected", k.jobID, k.jobTaskID, k.criteriaID)
			dup[k] = true
			continue
		}
		got[k] = true
		out[k] = ratingResolutionFrom(res)
	}
	for k := range dup {
		out[k] = ratingResolution{reject: true, reason: "rating_description_resolution_failed", logMsg: "resolver returned a duplicate result for a requested cell"}
	}
	for k := range seen {
		if !got[k] {
			log.Printf("[outcome-matrix] rating resolution: resolver omitted requested cell job=%s task=%s criteria=%s", k.jobID, k.jobTaskID, k.criteriaID)
			out[k] = ratingResolution{reject: true, reason: "rating_description_resolution_failed", logMsg: "resolver omitted a requested cell"}
		}
	}
	return out
}

// ratingResolutionFrom maps one CellRatingResolution to the shape the write
// path applies. RESOLVED and NO_LINK both save the score; every other status
// (or an unrecognized future status) rejects the cell's save entirely (Q22).
func ratingResolutionFrom(res *matrixpb.CellRatingResolution) ratingResolution {
	switch res.GetStatus() {
	case enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_RESOLVED:
		return ratingResolution{ratings: ratingDescriptionsFromList(res.GetDescriptions())}
	case enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_NO_LINK:
		// Q21: the offering × AY has no linked set — treated exactly like a
		// value with no entry (cleared on change, kept on same-value re-save).
		return ratingResolution{}
	case enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_UNRESOLVED_IDENTITY:
		return ratingResolution{reject: true, reason: "rating_description_unresolved_identity", logMsg: res.GetReason()}
	case enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_AMBIGUOUS:
		return ratingResolution{reject: true, reason: "rating_description_ambiguous", logMsg: res.GetReason()}
	case enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_INVALID_CONFIG:
		return ratingResolution{reject: true, reason: "rating_description_invalid_config", logMsg: res.GetReason()}
	case enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_PLACEHOLDER_UNRESOLVED:
		// schema-proposal §10: the resolved text carries a placeholder tag
		// whose value is missing for this record — fail closed, never write a
		// half-rendered (or raw-tag) note. Value and note stay untouched.
		return ratingResolution{reject: true, reason: "rating_description_placeholder_unresolved", logMsg: res.GetReason()}
	default:
		return ratingResolution{reject: true, reason: "rating_description_resolution_failed", logMsg: fmt.Sprintf("unexpected status %v: %s", res.GetStatus(), res.GetReason())}
	}
}

// ── note rules (schema-proposal §5) ─────────────────────────────────────────

// applyDescriptionNote folds one UPDATE's resolved rating into the note the
// write should persist (schema-proposal §5). res == nil means the column is
// not in description mode: the legacy no-touch behaviour applies (nil note,
// hasNote reflects the pre-existing note untouched). Callers must never invoke
// this when res.reject is true — that cell is rejected before any write is
// attempted (value and note stay untouched entirely).
//
//   - the new value has a matching entry                    → note = entry
//     text, ALWAYS replacing the stored note (even staff-edited wording, even
//     a same-value re-save/re-pick — Q10/Q14).
//   - the new value has no entry (NO_LINK counts as "no entry", Q21) AND the
//     value changed from what was stored                     → note cleared
//     ("").
//   - the new value has no entry and the value is UNCHANGED (a re-save of a
//     value that never had an entry, e.g. 0→0)                → note kept
//     (nil: the request does not touch it — a grader may have written one).
func applyDescriptionNote(res *ratingResolution, oldNumeric *float64, newNumeric float64, existingNote string) (note *string, hasNote bool) {
	if res == nil {
		return nil, existingNote != ""
	}
	if entry := res.ratings.describe(newNumeric); entry != "" {
		text := entry
		return &text, true
	}
	if oldNumeric == nil || *oldNumeric != newNumeric {
		empty := ""
		return &empty, false
	}
	return nil, existingNote != ""
}

// createDescriptionNote folds one CREATE's resolved rating into the note the
// new outcome should be stamped with. A brand-new cell has no prior value, so
// only the "has a matching entry" branch of the note rules applies (schema-
// proposal §5, "Create path: same rules") — no entry means nothing to clear or
// keep, so no note is written.
func createDescriptionNote(res *ratingResolution, newNumeric float64) (note *string, hasNote bool) {
	if res == nil {
		return nil, false
	}
	if entry := res.ratings.describe(newNumeric); entry != "" {
		text := entry
		return &text, true
	}
	return nil, false
}
