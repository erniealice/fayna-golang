package detail

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"

	pyeza "github.com/erniealice/pyeza-golang"

	activityexpensepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_expense"
	activitylaborpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_labor"
	activitymaterialpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_material"
	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
)

// Deps holds view dependencies for the job activity detail views.
type Deps struct {
	Routes       fayna.JobActivityRoutes
	Labels       fayna.JobActivityLabels
	CommonLabels pyeza.CommonLabels

	// Job activity read
	ReadJobActivity func(ctx context.Context, req *jobactivitypb.ReadJobActivityRequest) (*jobactivitypb.ReadJobActivityResponse, error)

	// Activity subtype read functions (for displaying detail by entry type)
	ReadActivityLabor    func(ctx context.Context, req *activitylaborpb.ReadActivityLaborRequest) (*activitylaborpb.ReadActivityLaborResponse, error)
	ReadActivityMaterial func(ctx context.Context, req *activitymaterialpb.ReadActivityMaterialRequest) (*activitymaterialpb.ReadActivityMaterialResponse, error)
	ReadActivityExpense  func(ctx context.Context, req *activityexpensepb.ReadActivityExpenseRequest) (*activityexpensepb.ReadActivityExpenseResponse, error)
}
