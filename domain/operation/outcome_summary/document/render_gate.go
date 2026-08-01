package document

import (
	"context"
	"fmt"
	"sort"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// render_gate.go — the D5 report-render gate (plan §4.4 / codex-rereview.md
// "Published-return and lock overlay contract"; hardened per codex-p3-review.md
// §B4; sheet grain made a composition-time declaration per
// docs/plan/20260729-report-card-render-gate-group-grain). Because every
// download is a LIVE render (collectCard reads live jobs/phases/outcomes/
// summaries), a card drawn from a sheet whose grades have entered the approval
// workflow but are not yet fully PUBLISHED must NOT be re-issued: the handler
// returns 409 (proven unsafe) or 503 (cannot prove safe). Never-workflowed
// sheets (the historical backfill: IN_PROGRESS with every approval audit stamp
// null) keep rendering exactly as today.
//
// THE SHEET GRAIN IS OPTION-SELECTED (DocumentOptions.GateGrain) — it is a
// deployment declaration, NOT a fixed design property of this file:
//
//   - "" (zero value) — template-phase grain: the sheet S = every active
//     job_phase sharing one of the card's template_phase_ids, workspace-wide.
//     The stricter, over-blocking gate (a sibling group mid-workflow blocks
//     this card too), byte-identical to the pre-option gate: same requests,
//     same call counts, same predicate.
//   - GateGrainSubscriptionGroup — (template_phase × subscription_group)
//     grain: the sheet S = the route group's OWN rows per template phase,
//     proven through the fail-closed GetPhaseApprovalGateRollup port (exact
//     applied-group echo + coverage of EVERY expected card template phase +
//     the card's own proven membership as the ≥1-member anchor). A sibling
//     group no longer blocks this one. The four approval transitions already
//     operate at this grain (the espyna group-membership transition
//     predicate), so this is the grain the write side has always used.
//
// An unrecognized grain never reaches this file: the block's EngineBlock
// refuses to boot (outcome_summary.ValidateGateGrain). There is NO runtime
// grain fallback in either direction — a configured group grain whose port is
// unpublished, or whose response cannot be proven (transport error, nil/failed
// response, coverage gap, echo mismatch, missing card anchor, duplicate rows)
// returns an ERROR → 503, never a silent widen back to the template-phase
// grain and never a card-seeded singleton floor.
//
// codex-p3-review.md §B4 rulings, in force under BOTH grains:
//   - **FAIL CLOSED**, not fail-open. D5 is a document-issuance INTEGRITY
//     boundary, not a UX nicety: inability to prove a sheet safe (nil
//     dependency, list/rollup error, permission denial on a reused read)
//     returns an ERROR → 503, never a silent render. A proven-unsafe sheet
//     returns blocked=true → 409.
//   - **Full-S aggregate.** A sheet blocks when BOOL_OR(any member
//     workflow-entered) AND NOT BOOL_AND(every member PUBLISHED) AND
//     data-present anywhere in S — so a PUBLISHED member with data plus a late
//     pristine IN_PROGRESS member blocks, and the workflow-entered member and
//     the data-bearing member may differ. Only the MEMBERSHIP of S varies by
//     grain; the block predicate never does.
//   - **Complete reads.** On the zero-grain path every list is chunked over
//     its id set AND paged to exhaustion (the reads are UNNARROWED, so a short
//     batch really is exhaustion — a generic-list default page cap can never
//     omit a sheet member). On the group-grain path the sheet walk is replaced
//     by ONE SQL aggregate whose group predicate is applied before GROUP BY
//     with no LIMIT/OFFSET — a page cap cannot exist there by construction.
//
// The local reads reuse the approval-projected espyna ListJobPhases (P1
// projection carries approval_status + the four audit stamps, workspace-
// ancestry-enforced by FIX-4) and the job_task → task_outcome data seam the
// transcript already walks. A phase with a NULL template_phase_id is not part
// of a multi-member sheet, so it is evaluated as its own singleton sheet
// under BOTH grains — the singleton branch is reached ONLY by that
// no-template-relation partition (Q2), never by skipping a failed or empty
// group step.
//
// Known over-refusal edge (group grain, documented + tested): a HISTORICAL
// section (inactive group) carries inactive membership rows, which the
// active-membership transition predicate would resolve to target_count = 0.
// In practice a historical card's phase ancestry is inactive too, so the
// card-phase active filter empties the template-phase set and the group path
// makes NO rollup call (the card renders exactly as today, via the
// never-workflowed carve-out). If live data ever presents ACTIVE phases under
// inactive membership, the gate 503s — the fail-SAFE direction.

// reportRenderStatus returns (blocked, err). groupID is the card's
// subscription group — the route group the download was requested under.
// collectCard has already proven the (group, client) pair before the gate
// runs: fetchSection resolved the group by id and memberSubscription proved
// the client's own membership row, 404ing otherwise (the outcome_matrix
// ResolveGroupScope pair-validation precedent). The zero-grain path ignores
// groupID entirely (byte-identical pre-option behavior).
//
//   - (true,  nil)  a data-present, workflow-entered, not-fully-published sheet feeds
//     this card → the handler returns 409.
//   - (false, nil)  every feeding sheet is either never-workflowed, fully published,
//     or empty → render.
//   - (_,     err)  the gate could not be evaluated (nil dependency / list or rollup
//     error / unprovable group narrow / permission denial) → the handler returns 503
//     (fail closed — never render on an unproven sheet).
func reportRenderStatus(ctx context.Context, d *Deps, jobIDs []string, groupID string) (bool, error) {
	if d == nil || d.ListJobPhases == nil {
		return false, fmt.Errorf("render gate: job_phase read not wired — cannot prove sheet safe (fail closed)")
	}
	if len(jobIDs) == 0 {
		return false, nil
	}

	// (1) The card's own phases → the set of target sheets (distinct
	// template_phase_id) the card draws from, plus any NULL-template-phase
	// singletons. This read is IDENTICAL under both grains and is NOT
	// group-narrowed: the card's jobs were already resolved through the route
	// group's membership by collectCard, so narrowing here could only ever
	// DROP one of the card's own phases — fewer sheets evaluated, a weaker
	// gate. Byte-identical request to the pre-option gate (the additive
	// ListJobPhasesRequest group field stays unconsumed).
	cardPhases, err := listPhasesByChunkedIDs(ctx, d, "job_id", jobIDs)
	if err != nil {
		return false, fmt.Errorf("render gate: list card phases: %w", err)
	}
	templatePhaseSet := map[string]struct{}{}
	var singletons []*jobphasepb.JobPhase // NULL template_phase_id — evaluated alone
	for _, p := range cardPhases {
		if !p.GetActive() {
			continue
		}
		if tp := p.GetTemplatePhaseId(); tp != "" {
			templatePhaseSet[tp] = struct{}{}
		} else {
			singletons = append(singletons, p)
		}
	}

	if d.DocOptions.GateGrainGroup() {
		return groupGrainRenderStatus(ctx, d, groupID, templatePhaseSet, singletons)
	}

	// ---- Zero grain: the template-phase gate, byte-identical to before the
	// option existed (same requests, same call counts, same predicate). ----

	// (2) Full sheet S: every active phase sharing one of the target template_phase_ids
	// (all members across the workspace, FIX-4 workspace-scoped), grouped per sheet.
	sheetMembers := map[string][]*jobphasepb.JobPhase{}
	if len(templatePhaseSet) > 0 {
		tpIDs := make([]string, 0, len(templatePhaseSet))
		for id := range templatePhaseSet {
			tpIDs = append(tpIDs, id)
		}
		members, err := listPhasesByChunkedIDs(ctx, d, "template_phase_id", tpIDs)
		if err != nil {
			return false, fmt.Errorf("render gate: list sheet members: %w", err)
		}
		for _, p := range members {
			if !p.GetActive() {
				continue
			}
			if tp := p.GetTemplatePhaseId(); tp != "" {
				sheetMembers[tp] = append(sheetMembers[tp], p)
			}
		}
	}

	// (3) Sheet-grain aggregate: a sheet is a block CANDIDATE when it is
	// BOOL_OR(workflow-entered) AND NOT BOOL_AND(PUBLISHED). Collect the candidate
	// sheets' member phase ids for the data-presence probe.
	var candidatePhaseIDs []string
	addCandidate := func(members []*jobphasepb.JobPhase) {
		anyEntered := false
		allPublished := true
		for _, p := range members {
			if phaseWorkflowEntered(p) {
				anyEntered = true
			}
			if p.GetApprovalStatus() != jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED {
				allPublished = false
			}
		}
		if anyEntered && !allPublished {
			for _, p := range members {
				if id := p.GetId(); id != "" {
					candidatePhaseIDs = append(candidatePhaseIDs, id)
				}
			}
		}
	}
	for _, members := range sheetMembers {
		addCandidate(members)
	}
	for _, p := range singletons {
		addCandidate([]*jobphasepb.JobPhase{p})
	}
	if len(candidatePhaseIDs) == 0 {
		return false, nil // every feeding sheet is never-workflowed or fully published
	}

	// (4) Data-presence over the candidate sheets.
	return probeCandidateData(ctx, d, candidatePhaseIDs)
}

// groupGrainRenderStatus is the GateGrainSubscriptionGroup path: template-
// backed sheets are proven through the fail-closed gate-rollup port; the
// NULL-template-phase singletons keep TODAY's local evaluation (Q2 — that
// branch is reached only by the no-template-relation partition, never by a
// failed or empty group step). blocked = OR(group sheets) OR OR(singletons).
func groupGrainRenderStatus(ctx context.Context, d *Deps, groupID string, templatePhaseSet map[string]struct{}, singletons []*jobphasepb.JobPhase) (bool, error) {
	if len(templatePhaseSet) > 0 {
		blocked, err := groupSheetsBlocked(ctx, d, groupID, templatePhaseSet)
		if err != nil {
			return false, err
		}
		if blocked {
			return true, nil
		}
	}
	return singletonSheetsBlocked(ctx, d, singletons)
}

// groupSheetsBlocked resolves the card's template-backed sheets at
// (template_phase × subscription_group) grain through the rollup port and
// applies the unchanged block predicate per covered sheet. EVERY unprovable
// state is an ERROR (→ 503): unpublished port, empty group, transport error,
// nil/failed response, unrequested or duplicate rollup rows, group-echo
// mismatch, missing coverage, or a missing card anchor. There is no fallback
// of any kind here — not to template grain, not to a card-seeded singleton.
func groupSheetsBlocked(ctx context.Context, d *Deps, groupID string, templatePhaseSet map[string]struct{}) (bool, error) {
	// Grain configured but the port unpublished (a build without the
	// specialized query, or unwired composition) is unprovable — never a
	// silent degrade to another grain.
	if d.GetPhaseApprovalGateRollup == nil {
		return false, fmt.Errorf("render gate: group grain configured but no gate-rollup implementation is published — cannot prove sheet safe (fail closed)")
	}
	// Unreachable in practice — the handler 404s an empty route group before
	// collectCard runs — but the gate re-checks: an empty group cannot scope a
	// group-grain read, and it must NOT degrade to a singleton evaluation.
	if groupID == "" {
		return false, fmt.Errorf("render gate: empty group id on a group-grain gate — cannot prove sheet safe (fail closed)")
	}

	tpIDs := make([]string, 0, len(templatePhaseSet))
	for id := range templatePhaseSet {
		tpIDs = append(tpIDs, id)
	}
	sort.Strings(tpIDs) // deterministic request shape

	resp, err := d.GetPhaseApprovalGateRollup(ctx, &matrixpb.GetPhaseApprovalGateRollupRequest{
		SubscriptionGroupId: groupID,
		JobTemplatePhaseIds: tpIDs,
	})
	if err != nil {
		return false, fmt.Errorf("render gate: gate rollup: %w", err)
	}
	if resp == nil || !resp.GetSuccess() {
		return false, fmt.Errorf("render gate: gate rollup returned no provable result (fail closed)")
	}

	rollups := map[string]*matrixpb.PhaseApprovalGateRollup{}
	for _, r := range resp.GetRollups() {
		tp := r.GetJobTemplatePhaseId()
		if _, wanted := templatePhaseSet[tp]; !wanted {
			return false, fmt.Errorf("render gate: gate rollup carries an unrequested template phase — cannot prove sheet safe (fail closed)")
		}
		if _, dup := rollups[tp]; dup {
			// Also the dedupe pin: one evaluation per template phase — a
			// response that presents the same sheet twice is ambiguous
			// coverage, never double-counted.
			return false, fmt.Errorf("render gate: duplicate gate rollup for a template phase — ambiguous coverage (fail closed)")
		}
		// The exact applied-group echo is the applied-narrow proof: an
		// implementation that ignored the narrow (or applied a different
		// group) cannot echo this group id.
		if r.GetAppliedSubscriptionGroupId() != groupID {
			return false, fmt.Errorf("render gate: gate rollup group echo mismatch — narrow not proven applied (fail closed)")
		}
		// Card anchor: collectCard proved this card's own membership in the
		// group, so a correct narrow cannot resolve an EMPTY sheet for a
		// phase this card feeds. Zero members = an under-resolved read.
		if r.GetTargetCount() < 1 {
			return false, fmt.Errorf("render gate: gate rollup missing the card's own membership anchor (fail closed)")
		}
		rollups[tp] = r
	}
	// Coverage: EVERY expected card template phase must be present (the
	// provider omits zero-member phases; absence here is unprovable, not safe).
	for tp := range templatePhaseSet {
		if _, ok := rollups[tp]; !ok {
			return false, fmt.Errorf("render gate: gate rollup coverage gap — cannot prove sheet safe (fail closed)")
		}
	}

	// Block predicate per covered sheet — UNCHANGED semantics, now from the
	// rollup fields. The never-workflowed carve-out is exactly
	// any_workflow_entered == false (audit-stamp-derived in SQL,
	// parity-pinned against phaseWorkflowEntered).
	for _, r := range rollups {
		if r.GetAnyWorkflowEntered() && !r.GetAllPublished() && r.GetHasData() {
			return true, nil
		}
	}
	return false, nil
}

// singletonSheetsBlocked evaluates the NULL-template-phase singletons exactly
// as the zero-grain path does (each phase its own sheet: workflow-entered AND
// not published AND data-present blocks). Shared by both grains — Q2.
func singletonSheetsBlocked(ctx context.Context, d *Deps, singletons []*jobphasepb.JobPhase) (bool, error) {
	var candidatePhaseIDs []string
	for _, p := range singletons {
		if phaseWorkflowEntered(p) &&
			p.GetApprovalStatus() != jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED {
			if id := p.GetId(); id != "" {
				candidatePhaseIDs = append(candidatePhaseIDs, id)
			}
		}
	}
	if len(candidatePhaseIDs) == 0 {
		return false, nil
	}
	return probeCandidateData(ctx, d, candidatePhaseIDs)
}

// probeCandidateData is step (4) of the gate: the data-presence probe over the
// candidate sheets' member phase ids — any active task under a candidate phase
// that carries an active task_outcome (task_outcome has no job_phase_id
// column — reach it via job_task). Missing data-seam deps FAIL CLOSED.
func probeCandidateData(ctx context.Context, d *Deps, candidatePhaseIDs []string) (bool, error) {
	if d.ListJobTasks == nil || d.ListTaskOutcomes == nil {
		return false, fmt.Errorf("render gate: data-seam read not wired — cannot prove sheet safe (fail closed)")
	}
	taskIDs, err := listActiveIDsByChunkedIDs(ctx, d, "phase", candidatePhaseIDs)
	if err != nil {
		return false, fmt.Errorf("render gate: list sheet tasks: %w", err)
	}
	if len(taskIDs) == 0 {
		return false, nil // candidate sheets carry no data → safe to render
	}
	hasData, err := anyActiveOutcome(ctx, d, taskIDs)
	if err != nil {
		return false, fmt.Errorf("render gate: probe sheet outcomes: %w", err)
	}
	return hasData, nil
}

// listPhasesByChunkedIDs lists active+inactive job_phase rows whose `field` is in
// `ids`, chunking the IN() set AND paging each chunk to exhaustion so a generic-list
// default page cap can never truncate a sheet (codex §B4 pagination fix). The
// requests are UNNARROWED (no group field), so short-batch exhaustion is valid.
func listPhasesByChunkedIDs(ctx context.Context, d *Deps, field string, ids []string) ([]*jobphasepb.JobPhase, error) {
	var out []*jobphasepb.JobPhase
	for start := 0; start < len(ids); start += pageLimit {
		end := start + pageLimit
		if end > len(ids) {
			end = len(ids)
		}
		chunk := ids[start:end]
		for page := int32(1); ; page++ {
			resp, err := d.ListJobPhases(ctx, &jobphasepb.ListJobPhasesRequest{
				Filters:    &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{listIn(field, chunk)}},
				Pagination: gateOffsetPage(page),
				Sort:       gateIDSort(),
			})
			if err != nil {
				return nil, err
			}
			batch := resp.GetData()
			out = append(out, batch...)
			if len(batch) < pageLimit {
				break
			}
		}
	}
	return out, nil
}

// listActiveIDsByChunkedIDs pages active job_task rows whose job_phase_id is in
// `ids` and returns their ids. `kind` is currently always "phase".
func listActiveIDsByChunkedIDs(ctx context.Context, d *Deps, kind string, ids []string) ([]string, error) {
	_ = kind
	var out []string
	for start := 0; start < len(ids); start += pageLimit {
		end := start + pageLimit
		if end > len(ids) {
			end = len(ids)
		}
		chunk := ids[start:end]
		for page := int32(1); ; page++ {
			resp, err := d.ListJobTasks(ctx, &jobtaskpb.ListJobTasksRequest{
				Filters:    &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{listIn("job_phase_id", chunk)}},
				Pagination: gateOffsetPage(page),
				Sort:       gateIDSort(),
			})
			if err != nil {
				return nil, err
			}
			batch := resp.GetData()
			for _, jt := range batch {
				if jt.GetActive() {
					if id := jt.GetId(); id != "" {
						out = append(out, id)
					}
				}
			}
			if len(batch) < pageLimit {
				break
			}
		}
	}
	return out, nil
}

// anyActiveOutcome reports whether any active task_outcome exists under the given
// job_task ids (chunked + paged; short-circuits on the first hit).
func anyActiveOutcome(ctx context.Context, d *Deps, taskIDs []string) (bool, error) {
	for start := 0; start < len(taskIDs); start += pageLimit {
		end := start + pageLimit
		if end > len(taskIDs) {
			end = len(taskIDs)
		}
		chunk := taskIDs[start:end]
		for page := int32(1); ; page++ {
			resp, err := d.ListTaskOutcomes(ctx, &taskoutcomepb.ListTaskOutcomesRequest{
				Filters:    &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{listIn("job_task_id", chunk)}},
				Pagination: gateOffsetPage(page),
				Sort:       gateIDSort(),
			})
			if err != nil {
				return false, err
			}
			batch := resp.GetData()
			for _, t := range batch {
				if t.GetActive() {
					return true, nil
				}
			}
			if len(batch) < pageLimit {
				break
			}
		}
	}
	return false, nil
}

// gateOffsetPage builds a 1-based offset pagination request of pageLimit rows.
func gateOffsetPage(page int32) *commonpb.PaginationRequest {
	return &commonpb.PaginationRequest{
		Limit:  pageLimit,
		Method: &commonpb.PaginationRequest_Offset{Offset: &commonpb.OffsetPagination{Page: page}},
	}
}

// gateIDSort pins a deterministic id sort so OFFSET paging over tied timestamps is
// stable (never drops/duplicates a member across pages).
func gateIDSort() *commonpb.SortRequest {
	return &commonpb.SortRequest{
		Fields: []*commonpb.SortField{{Field: "id", Direction: commonpb.SortDirection_ASC}},
	}
}

// phaseWorkflowEntered reports whether a job_phase has entered the approval
// workflow: its status advanced beyond IN_PROGRESS, OR any approval audit stamp
// is set (submit/verify/publish/return). A pristine backfill row (IN_PROGRESS +
// all audit null) is NOT workflow-entered.
func phaseWorkflowEntered(p *jobphasepb.JobPhase) bool {
	switch p.GetApprovalStatus() {
	case jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_UNSPECIFIED,
		jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_IN_PROGRESS:
		// only workflow-entered if an audit stamp is present (e.g. returned)
		return p.GetSubmittedBy() != "" || p.GetVerifiedBy() != "" ||
			p.GetPublishedBy() != "" || p.GetReturnedBy() != ""
	default:
		return true
	}
}
