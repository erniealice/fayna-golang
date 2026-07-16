package outcome_summary

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
)

// markEvidencePageLimit chunks ListFilter(IN) id sets so each call's result set
// stays under the adapter's default cap. markEvidenceMaxPages bounds every
// offset page-loop independently of the adapter's own termination (which relies
// on a short final page) — a section's phase/task set (roster × subjects × ~2
// phases ≈ 600+) far exceeds the default row caps, and an uncapped single call
// silently truncates the evidence, so every walk pages explicitly.
const (
	markEvidencePageLimit = 100
	markEvidenceMaxPages  = 100
)

// EnrollmentEvidence is the per-job task-mark evidence that drives the
// non-enrolled-placeholder predicate. It is the generic, surface-agnostic
// distillation of "does this job carry real grading marks" — the same signal
// the report-card DOCX (B1) uses to suppress untaken-elective placeholder rows.
type EnrollmentEvidence struct {
	// HasMarks is true iff the job has ≥1 numeric task_outcome (an all-zero
	// scaffold OR a positive mark). Distinguishes an all-zero active scaffold
	// (an untaken parallel track — a placeholder) from a subject with no
	// task_outcome at all (a historical import — a real, kept subject).
	HasMarks bool
	// HasPositiveMark is true iff the job has ≥1 task_outcome with a numeric
	// value > 0. The authoritative enrollment discriminator: a genuinely-taken
	// subject always carries a positive per-criterion mark somewhere, even when
	// its composite year-final floors to "0"/"1".
	HasPositiveMark bool
}

// FetchJobMarkEvidence walks job_phase → job_task → task_outcome for the given
// jobIDs and returns per-job mark evidence (existence + any-positive), keyed by
// job id. It is the bulk (no N+1), existence-only counterpart to the DOCX's
// per-criterion fetchCriteriaByJob: every walk pages explicitly and chunks its
// IN-filter so a large section's evidence is never silently truncated.
//
// PRINCIPAL PREREQUISITE (report cards are admin/registrar-only today): the
// task walk uses listJobTasks, which staff-narrows to tasks assigned to the
// acting principal when that principal is a workspace STAFF member. For a
// non-staff principal (admin/registrar — the only ones the Layer-3
// job_outcome_summary:list gate currently admits) the walk sees every task, so
// the evidence is authoritative. If report-card access is EVER granted to staff
// (the deferred teacher-access decision), a teacher viewing a multi-teacher job
// would see only their own tasks and could blank a co-teacher's genuine mark —
// so before lifting that gate, this evidence read MUST switch to a
// non-staff-scoped aggregate over the already-authorized job ids.
//
// Fully nil-safe: any nil closure (a tier that never wired the walk, e.g.
// service-admin's flat surfaces) or an empty jobIDs yields an empty map, so
// IsNonEnrolledCell then keeps every cell (no behavior change off-education).
//
// FAIL-CLOSED evidence contract: a suppression decision must rest on COMPLETE
// evidence, so any list-read error aborts the whole walk and returns (nil, err)
// — never a truncated map. Callers MUST treat a non-nil error as "evidence
// unavailable → blank nothing, keep every displayed grade": partial evidence
// (e.g. a page of all-zero marks read before the page carrying the positive
// mark errored) would otherwise blank a genuinely-earned "0"/"1" on an official
// document. Keeping a phantom "1" is a cosmetic miss; blanking a real grade is a
// data error — so we always fail toward keeping the grade.
func FetchJobMarkEvidence(
	ctx context.Context,
	listJobPhases func(ctx context.Context, req *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error),
	listJobTasks func(ctx context.Context, req *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error),
	listTaskOutcomes func(ctx context.Context, req *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error),
	jobIDs []string,
) (map[string]EnrollmentEvidence, error) {
	out := map[string]EnrollmentEvidence{}
	if listJobPhases == nil || listJobTasks == nil || listTaskOutcomes == nil || len(jobIDs) == 0 {
		return out, nil
	}

	// 1. job_phase → owning job. (phase_id -> job_id)
	jobByPhase := map[string]string{}
	phaseIDs := make([]string, 0, len(jobIDs)*2)
	for start := 0; start < len(jobIDs); start += markEvidencePageLimit {
		end := start + markEvidencePageLimit
		if end > len(jobIDs) {
			end = len(jobIDs)
		}
		chunk := jobIDs[start:end]
		for page := int32(1); page <= markEvidenceMaxPages; page++ {
			resp, err := listJobPhases(ctx, &jobphasepb.ListJobPhasesRequest{
				Filters:    &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{markEvidenceListIn("job_id", chunk)}},
				Pagination: markEvidencePage(page),
				Sort:       markEvidenceSortByID(),
			})
			if err != nil {
				// Fail-closed: incomplete evidence must not drive suppression.
				return nil, fmt.Errorf("outcome summary: mark evidence list job phases (page %d): %w", page, err)
			}
			data := resp.GetData()
			for _, p := range data {
				pid, jid := p.GetId(), p.GetJobId()
				if pid == "" || jid == "" {
					continue
				}
				if _, seen := jobByPhase[pid]; seen {
					continue
				}
				jobByPhase[pid] = jid
				phaseIDs = append(phaseIDs, pid)
			}
			if len(data) < markEvidencePageLimit {
				break
			}
		}
	}
	if len(phaseIDs) == 0 {
		return out, nil
	}

	// 2. job_task → owning job (via phase). (task_id -> job_id)
	jobByTask := map[string]string{}
	taskIDs := make([]string, 0, len(phaseIDs))
	for start := 0; start < len(phaseIDs); start += markEvidencePageLimit {
		end := start + markEvidencePageLimit
		if end > len(phaseIDs) {
			end = len(phaseIDs)
		}
		chunk := phaseIDs[start:end]
		for page := int32(1); page <= markEvidenceMaxPages; page++ {
			resp, err := listJobTasks(ctx, &jobtaskpb.ListJobTasksRequest{
				Filters:    &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{markEvidenceListIn("job_phase_id", chunk)}},
				Pagination: markEvidencePage(page),
				Sort:       markEvidenceSortByID(),
			})
			if err != nil {
				// Fail-closed: incomplete evidence must not drive suppression.
				return nil, fmt.Errorf("outcome summary: mark evidence list job tasks (page %d): %w", page, err)
			}
			data := resp.GetData()
			for _, tk := range data {
				id := tk.GetId()
				jid := jobByPhase[tk.GetJobPhaseId()]
				if id == "" || jid == "" || !tk.GetActive() {
					continue
				}
				if _, seen := jobByTask[id]; seen {
					continue
				}
				jobByTask[id] = jid
				taskIDs = append(taskIDs, id)
			}
			if len(data) < markEvidencePageLimit {
				break
			}
		}
	}
	if len(taskIDs) == 0 {
		return out, nil
	}

	// 3. task_outcome → per-job evidence (any numeric mark; any positive mark).
	for start := 0; start < len(taskIDs); start += markEvidencePageLimit {
		end := start + markEvidencePageLimit
		if end > len(taskIDs) {
			end = len(taskIDs)
		}
		chunk := taskIDs[start:end]
		for page := int32(1); page <= markEvidenceMaxPages; page++ {
			resp, err := listTaskOutcomes(ctx, &taskoutcomepb.ListTaskOutcomesRequest{
				Filters:    &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{markEvidenceListIn("job_task_id", chunk)}},
				Pagination: markEvidencePage(page),
				Sort:       markEvidenceSortByID(),
			})
			if err != nil {
				// Fail-closed: incomplete evidence must not drive suppression.
				return nil, fmt.Errorf("outcome summary: mark evidence list task outcomes (page %d): %w", page, err)
			}
			data := resp.GetData()
			for _, t := range data {
				if !t.GetActive() || t.NumericValue == nil {
					continue
				}
				jid := jobByTask[t.GetJobTaskId()]
				if jid == "" {
					continue
				}
				ev := out[jid]
				ev.HasMarks = true
				if t.GetNumericValue() > 0 {
					ev.HasPositiveMark = true
				}
				out[jid] = ev
			}
			if len(data) < markEvidencePageLimit {
				break
			}
		}
	}
	return out, nil
}

// IsNonEnrolledCell reports whether a rendered grade cell (or subject row) is a
// non-enrolled placeholder that must render BLANK — the grid/card mirror of the
// DOCX's isNonEnrolledPlaceholder row suppression. bands are the cell's stored
// labels (year-final + any semester bands). A row is a placeholder when it
// carries NO positive grade evidence:
//
//   - no per-criterion mark is > 0 (ev.HasPositiveMark — the authoritative
//     discriminator: a genuinely-taken subject, including one whose composite
//     floors to "0"/"1", always has ≥1 positive task_outcome), AND
//   - no REAL (> 1) stored year-final / semester band (the transmute-of-zero
//     floor is "0"/"1" and is not evidence of enrollment), AND
//   - it either HAS task_outcome marks (the all-zero active scaffold) OR has no
//     summary at all (a fully-blank row).
//
// A genuinely-enrolled subject with a real 0 keeps rendering: it has a positive
// mark somewhere, a real (> 1) stored band, or — for historical imports — no
// task_outcome but a real stored band (HasMarks=false + a band present). A
// rendered grid cell always has a band, so hasSummary is normally true and the
// decision reduces to "all-zero scaffold with no > 1 band".
func IsNonEnrolledCell(ev EnrollmentEvidence, bands ...string) bool {
	if ev.HasPositiveMark {
		return false // real per-criterion mark → enrolled
	}
	hasSummary := false
	for _, b := range bands {
		if NumGreaterThan(b, 1) {
			return false // a real (non-floor) stored band → keep
		}
		if strings.TrimSpace(b) != "" {
			hasSummary = true
		}
	}
	// All-zero scaffold (HasMarks) or fully-blank (no summary) → placeholder.
	// No task_outcome BUT a summary present → historical real subject → keep.
	return ev.HasMarks || !hasSummary
}

// NumGreaterThan reports whether s parses as a number strictly greater than n.
// Non-numeric or blank values are treated as not-greater (false). Shared by
// IsNonEnrolledCell and the DOCX placeholder predicate so the numeric-band test
// is defined once.
func NumGreaterThan(s string, n float64) bool {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return err == nil && f > n
}

// markEvidenceListIn builds a LIST_IN filter for the mark-evidence walk.
func markEvidenceListIn(field string, values []string) *commonpb.TypedFilter {
	return &commonpb.TypedFilter{
		Field: field,
		FilterType: &commonpb.TypedFilter_ListFilter{
			ListFilter: &commonpb.ListFilter{Values: values, Operator: commonpb.ListOperator_LIST_IN},
		},
	}
}

// markEvidencePage builds an offset pagination request for the given 1-based page.
func markEvidencePage(page int32) *commonpb.PaginationRequest {
	return &commonpb.PaginationRequest{
		Limit:  int32(markEvidencePageLimit),
		Method: &commonpb.PaginationRequest_Offset{Offset: &commonpb.OffsetPagination{Page: page}},
	}
}

// markEvidenceSortByID sorts each paged list by the primary key (id ASC), a
// UNIQUE column, so OFFSET pagination is deterministic across pages. Without it
// the base adapter falls back to `ORDER BY date_created DESC`; education1's
// bulk-seed stamps every phase/task/outcome row of a section with an IDENTICAL
// date_created, and OFFSET paging over a fully-tied sort key returns an
// overlapping/gapped subset per page — silently dropping whole jobs' mark rows,
// which flips a taken job to HasMarks=false and leaves its phantom cell showing
// "1" (an under-blank). A unique id-sort makes every row land on exactly one
// page. The base List REPLACES the default order with this Sort; existence-only
// walks don't care about row order, only that no row is dropped or duplicated.
func markEvidenceSortByID() *commonpb.SortRequest {
	return &commonpb.SortRequest{
		Fields: []*commonpb.SortField{{Field: "id"}}, // Direction zero-value = ASC
	}
}
