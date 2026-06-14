package evaluation_cycle

import (
	"context"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	cyclepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_cycle"
	memberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_cycle_member"
	evalpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation"
)

// CycleProgress is the X-of-Y READ PROJECTION over evaluation_cycle_member.
//
//	Y = COUNT(member WHERE evaluation_cycle_id = cycle)            (frozen denominator, SR-1)
//	X = COUNT(member ⋈ SUBMITTED/SIGNED_OFF evaluation matching
//	          (subject_staff_id, client_id, period, type))         (STR-2: NOT materialize-DRAFT)
//
// Either the espyna GetCycleProgress read-UC supplies it server-side (preferred,
// view-delta F-GATE.2) OR the view computes it inline from
// ListEvaluationCycleMembers + ListEvaluations (the closures below).
type CycleProgress struct {
	CycleID   string
	Completed int // X
	Total     int // Y (frozen member count)
}

// ModuleDeps holds the typed closures that the action builders and sub-packages
// need. Defined here (not in the module assembler) to avoid a self-import cycle.
//
// The use-case closures are INJECTED by the Integrator / block — this package
// never calls espyna directly. GetCycleProgress is OPTIONAL: when nil, the
// detail view / banner compute the projection inline from the member + eval
// list closures (no materialize-DRAFT either way).
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Evaluation cycle CRUD + lifecycle (espyna usecases, injected).
	CreateEvaluationCycle func(ctx context.Context, req *cyclepb.CreateEvaluationCycleRequest) (*cyclepb.CreateEvaluationCycleResponse, error)
	ReadEvaluationCycle   func(ctx context.Context, req *cyclepb.ReadEvaluationCycleRequest) (*cyclepb.ReadEvaluationCycleResponse, error)
	ListEvaluationCycles  func(ctx context.Context, req *cyclepb.ListEvaluationCyclesRequest) (*cyclepb.ListEvaluationCyclesResponse, error)
	// OpenEvaluationCycle / CloseEvaluationCycle map to the espyna
	// evaluation_cycle/orchestration.go OpenUseCase / CloseUseCase (idempotent
	// member enrolment over ACTIVE seats; NO DRAFT materialize).
	OpenEvaluationCycle  func(ctx context.Context, req *cyclepb.UpdateEvaluationCycleRequest) (*cyclepb.UpdateEvaluationCycleResponse, error)
	CloseEvaluationCycle func(ctx context.Context, req *cyclepb.UpdateEvaluationCycleRequest) (*cyclepb.UpdateEvaluationCycleResponse, error)

	// Members tab (count = Y, TEST-2). No standalone member list page (STR-1).
	ListEvaluationCycleMembers func(ctx context.Context, req *memberpb.ListEvaluationCycleMembersRequest) (*memberpb.ListEvaluationCycleMembersResponse, error)

	// X-of-Y banner projection.
	// Preferred: the server-side read-UC (view-delta F-GATE.2). When nil, the
	// view computes inline from ListEvaluationCycleMembers + ListEvaluations.
	GetCycleProgress func(ctx context.Context, cycleID string) (*CycleProgress, error)
	ListEvaluations  func(ctx context.Context, req *evalpb.ListEvaluationsRequest) (*evalpb.ListEvaluationsResponse, error)

	NewID func() string
}
