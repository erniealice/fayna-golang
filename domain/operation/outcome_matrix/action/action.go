package action

import (
	"context"

	outcome_matrix "github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"

	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// Deps holds dependencies for the outcome matrix batch-save action handler.
//
// All writes route through the task_outcome:create / task_outcome:update use
// cases (never raw SQL). ResolveStaff supplies the acting staff_id for the IDOR
// guard (a cell may only be updated by its recorded_by owner).
type Deps struct {
	Routes outcome_matrix.Routes
	Labels outcome_matrix.Labels

	CreateTaskOutcome func(ctx context.Context, req *taskoutcomepb.CreateTaskOutcomeRequest) (*taskoutcomepb.CreateTaskOutcomeResponse, error)
	UpdateTaskOutcome func(ctx context.Context, req *taskoutcomepb.UpdateTaskOutcomeRequest) (*taskoutcomepb.UpdateTaskOutcomeResponse, error)
	ReadTaskOutcome   func(ctx context.Context, req *taskoutcomepb.ReadTaskOutcomeRequest) (*taskoutcomepb.ReadTaskOutcomeResponse, error)

	// GetOutcomeMatrix re-derives the acting principal's MINE-scoped matrix on
	// POST so the batch save only touches cells the server itself says are
	// addressable — the POST body's cell addresses are attacker-controlled and
	// must never be trusted as a scope.
	GetOutcomeMatrix func(ctx context.Context, req *matrixpb.GetOutcomeMatrixRequest) (*matrixpb.GetOutcomeMatrixResponse, error)

	// ResolveStaff maps the acting session user → staff_id ("" == fail-closed).
	ResolveStaff func(ctx context.Context) (string, error)

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
