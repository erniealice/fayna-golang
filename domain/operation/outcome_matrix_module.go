package operation

import (
	"context"

	outcomematrixpkg "github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"
	outcomematrixaction "github.com/erniealice/fayna-golang/domain/operation/outcome_matrix/action"
	outcomematrixlist "github.com/erniealice/fayna-golang/domain/operation/outcome_matrix/list"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/view"

	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// OutcomeMatrixModuleDeps holds all dependencies for the outcome matrix module.
//
// GetOutcomeMatrix is the new espyna use case (typed against the generated
// esqyma request/response). The TaskOutcome CRUD closures back the batch-save
// write path (create + update), and ResolveStaff supplies the acting staff_id
// for both the read-only gate (view) and the IDOR guard (record action).
type OutcomeMatrixModuleDeps struct {
	Routes       outcomematrixpkg.Routes
	Labels       outcomematrixpkg.Labels
	CommonLabels pyeza.CommonLabels

	GetOutcomeMatrix func(ctx context.Context, req *matrixpb.GetOutcomeMatrixRequest) (*matrixpb.GetOutcomeMatrixResponse, error)

	CreateTaskOutcome func(ctx context.Context, req *taskoutcomepb.CreateTaskOutcomeRequest) (*taskoutcomepb.CreateTaskOutcomeResponse, error)
	UpdateTaskOutcome func(ctx context.Context, req *taskoutcomepb.UpdateTaskOutcomeRequest) (*taskoutcomepb.UpdateTaskOutcomeResponse, error)
	ReadTaskOutcome   func(ctx context.Context, req *taskoutcomepb.ReadTaskOutcomeRequest) (*taskoutcomepb.ReadTaskOutcomeResponse, error)

	ResolveStaff func(ctx context.Context) (string, error)

	// ListClients hydrates the roster's display names (the matrix's client_id
	// rows are otherwise opaque — see list/page.go's PageViewDeps.ListClients
	// doc comment). Same closure the job drawer's client search picker
	// already uses; optional/nil-safe.
	ListClients func(ctx context.Context, req *clientpb.ListClientsRequest) (*clientpb.ListClientsResponse, error)

	// Page-header delivery-group resolution (round 4 item 2) — see
	// list/page.go's PageViewDeps doc comment for the full chain. All three
	// are already-wired top-level closures reused from elsewhere in the
	// block; optional/nil-safe.
	ListJobs                     func(ctx context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error)
	ListSubscriptionGroupMembers func(ctx context.Context, req *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error)
	ListSubscriptionGroups       func(ctx context.Context, req *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error)
}

// OutcomeMatrixModule holds the constructed outcome matrix views.
type OutcomeMatrixModule struct {
	routes outcomematrixpkg.Routes
	Matrix view.View // GET — the grid page
	Record view.View // POST — batch save
}

// NewOutcomeMatrixModule creates the outcome matrix module with all views wired.
func NewOutcomeMatrixModule(deps *OutcomeMatrixModuleDeps) *OutcomeMatrixModule {
	matrixView := outcomematrixlist.NewView(&outcomematrixlist.PageViewDeps{
		Routes:                       deps.Routes,
		Labels:                       deps.Labels,
		CommonLabels:                 deps.CommonLabels,
		GetOutcomeMatrix:             deps.GetOutcomeMatrix,
		ResolveStaff:                 deps.ResolveStaff,
		ListClients:                  deps.ListClients,
		ListJobs:                     deps.ListJobs,
		ListSubscriptionGroupMembers: deps.ListSubscriptionGroupMembers,
		ListSubscriptionGroups:       deps.ListSubscriptionGroups,
	})

	recordView := outcomematrixaction.NewRecordAction(&outcomematrixaction.Deps{
		Routes:            deps.Routes,
		Labels:            deps.Labels,
		CreateTaskOutcome: deps.CreateTaskOutcome,
		UpdateTaskOutcome: deps.UpdateTaskOutcome,
		ReadTaskOutcome:   deps.ReadTaskOutcome,
		GetOutcomeMatrix:  deps.GetOutcomeMatrix,
		ResolveStaff:      deps.ResolveStaff,
	})

	return &OutcomeMatrixModule{
		routes: deps.Routes,
		Matrix: matrixView,
		Record: recordView,
	}
}

// RegisterRoutes registers the outcome matrix routes.
func (m *OutcomeMatrixModule) RegisterRoutes(r view.RouteRegistrar) {
	if m.Matrix != nil {
		r.GET(m.routes.MatrixURL, m.Matrix)
	}
	if m.Record != nil {
		r.POST(m.routes.RecordURL, m.Record)
	}
}
