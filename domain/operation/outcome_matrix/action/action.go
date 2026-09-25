package action

import (
	"context"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"

	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// Deps holds dependencies for the outcome matrix batch-save action handler.
//
// All writes route through task_outcome:create / task_outcome:update / task_outcome:delete
// use cases (never raw SQL). ResolveStaff supplies the acting staff_id for the IDOR
// guard (a cell may only be updated or deleted by its recorded_by owner).
type Deps struct {
	Routes outcome_matrix.Routes
	Labels outcome_matrix.Labels

	CreateTaskOutcome func(ctx context.Context, req *taskoutcomepb.CreateTaskOutcomeRequest) (*taskoutcomepb.CreateTaskOutcomeResponse, error)
	UpdateTaskOutcome func(ctx context.Context, req *taskoutcomepb.UpdateTaskOutcomeRequest) (*taskoutcomepb.UpdateTaskOutcomeResponse, error)
	ReadTaskOutcome   func(ctx context.Context, req *taskoutcomepb.ReadTaskOutcomeRequest) (*taskoutcomepb.ReadTaskOutcomeResponse, error)
	DeleteTaskOutcome func(ctx context.Context, req *taskoutcomepb.DeleteTaskOutcomeRequest) (*taskoutcomepb.DeleteTaskOutcomeResponse, error)

	// Q26 conditional writes (schema-proposal §9.1, interfaces.md §2b). The
	// record action writes every value cell through these when wired:
	//   - UpdateTaskOutcomeIfUnchanged lands only if the row still carries the
	//     snapshot read earlier in the SAME request (expected: numeric_value,
	//     determination_note, date_modified).
	//   - CreateTaskOutcomeIfAbsent inserts only when the (job_task,
	//     criterion) cell still has no active outcome (job_task row lock).
	// conflict == true ⇒ nothing was written; the cell is rejected with
	// `cell_changed_retry` (value + note untouched). When nil, a
	// description-mode cell is REJECTED (fail closed — its note decision
	// depends on the snapshot); other cells fall back to the plain
	// Create/UpdateTaskOutcome closures (pre-Q26 behaviour).
	UpdateTaskOutcomeIfUnchanged func(ctx context.Context, req *taskoutcomepb.UpdateTaskOutcomeRequest, expected *taskoutcomepb.TaskOutcome) (resp *taskoutcomepb.UpdateTaskOutcomeResponse, conflict bool, err error)
	CreateTaskOutcomeIfAbsent    func(ctx context.Context, req *taskoutcomepb.CreateTaskOutcomeRequest) (resp *taskoutcomepb.CreateTaskOutcomeResponse, conflict bool, err error)

	// GetOutcomeMatrix re-derives the acting principal's MINE-scoped matrix on
	// POST so the batch save only touches cells the server itself says are
	// addressable — the POST body's cell addresses are attacker-controlled and
	// must never be trusted as a scope.
	GetOutcomeMatrix func(ctx context.Context, req *matrixpb.GetOutcomeMatrixRequest) (*matrixpb.GetOutcomeMatrixResponse, error)

	// ResolveStaff maps the acting session user → staff_id ("" == fail-closed).
	ResolveStaff func(ctx context.Context) (string, error)

	// ResolveCellRatingDescriptions batch-resolves the numeric-with-description
	// cells of ONE save request to their linked rating_description_set entries
	// (schema-proposal §4/§5; interfaces.md §2 OutcomeMatrixService RPC). The
	// record action calls it exactly once per request with every description-
	// mode cell's SERVER-DERIVED identity (job_id, job_task_id,
	// outcome_criteria_id — never the POST body; a POST-supplied identity is
	// never trusted for resolution any more than it is for authority).
	//
	// Q18 (no legacy fallback): this is the ONLY source of description text for
	// a description-mode cell — the column's own (legacy per-slot)
	// RatingDescriptions are never read on the write path any more. A nil
	// closure (unwired) REJECTS every description-mode cell's save in the
	// request (fail-closed — never silently "no entry"); see
	// rating_description.go's resolveRatingDescriptions.
	//
	// NOTE (espyna wiring, added ahead of the esqyma proto regen this depends
	// on — matrixpb.ResolveCellRatingDescriptions{Request,Response} do not
	// exist yet as of this write; this field will not compile until the W1
	// proto regen (CHECKPOINT-1) lands): the espyna agent wires this from
	// OutcomeMatrixModuleDeps (see fayna-golang.md "Resolver pass-through") —
	// this package only declares and calls the seam.
	ResolveCellRatingDescriptions func(ctx context.Context, req *matrixpb.ResolveCellRatingDescriptionsRequest) (*matrixpb.ResolveCellRatingDescriptionsResponse, error)

	// ComputePhaseOutcome / ComputeJobOutcome are the inline grade-recompute
	// closures (Q-GSE-5) called after a successful ACADEMIC cell write in
	// save_mode=cell. Both keyed off the SERVER-DERIVED job_phase_id / job_id
	// from the re-derived matrix (never a POST value). Both optional/nil-safe: a
	// nil closure → the cell still saves, reported ratingFresh:false (grade
	// persisted, rating stale). Return contract:
	//   (true,  nil) → recomputed        → ratingFresh (true)
	//   (false, nil) → frozen/authoritative skip → ratingFresh (true), rating
	//                  not stale (the authoritative grade stands)
	//   (false, err) → compute failed    → ratingFresh:false (stale, retryable)
	ComputePhaseOutcome func(ctx context.Context, jobPhaseID string) (bool, error)
	ComputeJobOutcome   func(ctx context.Context, jobID string) (bool, error)

	// RecomputeEligibility classifies, from the scoring graph, whether a saved
	// numeric cell on a job phase drives a scaled-summary recompute: eligible is
	// true only when the phase's scheme resolves a score scale and has scoped
	// criteria, and inScope is that scheme's active component-graph criterion id
	// set. A numeric cell whose phase is ineligible (e.g. a ledger scheme with no
	// scale) or whose criterion is outside inScope saves normally and is acked
	// not-applicable — never enqueued for a roll-up that would fail loud, never
	// reported stale. Optional/nil-safe: a nil closure (or a lookup error) falls
	// back to numeric-type classification (the prior behavior), so a save never
	// silently stops refreshing summaries.
	RecomputeEligibility func(ctx context.Context, jobPhaseID string) (eligible bool, inScope map[string]bool, err error)
}
