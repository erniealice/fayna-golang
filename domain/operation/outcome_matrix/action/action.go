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
}
