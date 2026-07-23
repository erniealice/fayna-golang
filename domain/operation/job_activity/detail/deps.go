package detail

import (
	"context"

	"github.com/erniealice/fayna-golang/domain/operation/activity_expense"
	"github.com/erniealice/fayna-golang/domain/operation/activity_labor"
	"github.com/erniealice/fayna-golang/domain/operation/activity_material"
	"github.com/erniealice/fayna-golang/domain/operation/job_activity"

	"github.com/erniealice/hybra-golang/views/attachment"
	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"

	activityexpensepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_expense"
	activitylaborpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_labor"
	activitymaterialpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_material"
	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
)

// DetailViewDeps holds view dependencies for the job activity detail views.
type DetailViewDeps struct {
	attachment.AttachmentOps
	auditlog.AuditOps

	Routes       job_activity.Routes
	Labels       job_activity.Labels
	CommonLabels pyeza.CommonLabels

	// ActivityLaborRoutes is used by the charge tab to resolve the labor edit URL.
	ActivityLaborRoutes activity_labor.Routes

	// ActivityMaterialRoutes is used by the charge tab to resolve the material edit URL.
	ActivityMaterialRoutes activity_material.Routes

	// ActivityExpenseRoutes is used by the charge tab to resolve the expense edit URL.
	ActivityExpenseRoutes activity_expense.Routes

	// Job activity read
	ReadJobActivity func(ctx context.Context, req *jobactivitypb.ReadJobActivityRequest) (*jobactivitypb.ReadJobActivityResponse, error)

	// Activity subtype read functions (for displaying detail by entry type)
	ReadActivityLabor    func(ctx context.Context, req *activitylaborpb.ReadActivityLaborRequest) (*activitylaborpb.ReadActivityLaborResponse, error)
	ReadActivityMaterial func(ctx context.Context, req *activitymaterialpb.ReadActivityMaterialRequest) (*activitymaterialpb.ReadActivityMaterialResponse, error)
	ReadActivityExpense  func(ctx context.Context, req *activityexpensepb.ReadActivityExpenseRequest) (*activityexpensepb.ReadActivityExpenseResponse, error)
}
