package document

import (
	"context"
	"fmt"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
)

// render_gate.go — the D5 report-render gate (plan §4.4 / codex-rereview.md
// "Published-return and lock overlay contract"; hardened per codex-p3-review.md
// §B4). Because every download is a LIVE render (collectCard reads live
// jobs/phases/outcomes/summaries), a card drawn from a sheet whose grades have
// entered the approval workflow but are not yet fully PUBLISHED must NOT be
// re-issued: the handler returns 409 (proven unsafe) or 503 (cannot prove safe).
// Never-workflowed sheets (the historical backfill: IN_PROGRESS with every
// approval audit stamp null) keep rendering exactly as today.
//
// codex-p3-review.md §B4 rulings folded in:
//   - **FAIL CLOSED**, not fail-open. D5 is a document-issuance INTEGRITY boundary,
//     not a UX nicety: inability to prove a sheet safe (nil dependency, list error,
//     permission denial on a reused read) returns an ERROR → 503, never a silent
//     render. A proven-unsafe sheet returns blocked=true → 409.
//   - **Full-S, template-phase grain.** The target is the whole approval sheet
//     S = every active job_phase sharing a template_phase_id, NOT just the current
//     card's jobs. A sheet blocks when BOOL_OR(any member workflow-entered) AND NOT
//     BOOL_AND(every member PUBLISHED) AND data-present anywhere in S — so a
//     PUBLISHED member with data plus a late pristine IN_PROGRESS member blocks, and
//     the workflow-entered member and the data-bearing member may differ.
//   - **Complete, paged reads.** Every list is chunked over its id set AND paged to
//     exhaustion, so a generic-list default page cap can never omit a sheet member.
//
// The reads reuse the approval-projected espyna ListJobPhases (P1 projection carries
// approval_status + the four audit stamps, workspace-ancestry-enforced by FIX-4) and
// the job_task → task_outcome data seam the transcript already walks. A phase with a
// NULL template_phase_id is not part of a multi-student sheet, so it is evaluated as
// its own singleton sheet.

// reportRenderStatus returns (blocked, err):
//   - (true,  nil)  a data-present, workflow-entered, not-fully-published sheet feeds
//     this card → the handler returns 409.
//   - (false, nil)  every feeding sheet is either never-workflowed, fully published,
//     or empty → render.
//   - (_,     err)  the gate could not be evaluated (nil dependency / list error /
//     permission denial) → the handler returns 503 (fail closed — never render on an
//     unproven sheet).
func reportRenderStatus(ctx context.Context, d *Deps, jobIDs []string) (bool, error) {
	if d == nil || d.ListJobPhases == nil {
		return false, fmt.Errorf("render gate: job_phase read not wired — cannot prove sheet safe (fail closed)")
	}
	if len(jobIDs) == 0 {
		return false, nil
	}

	// (1) The card's own phases → the set of target sheets (distinct template_phase_id)
	// the card draws from, plus any NULL-template-phase singletons.
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

	// (4) Data-presence over the candidate sheets: any active task under a candidate
	// phase that carries an active task_outcome (task_outcome has no job_phase_id
	// column — reach it via job_task). Missing data-seam deps FAIL CLOSED.
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
// default page cap can never truncate a sheet (codex §B4 pagination fix).
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
