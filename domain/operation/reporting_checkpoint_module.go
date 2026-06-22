package operation

import (
	"context"

	reportingcheckpointpkg "github.com/erniealice/fayna-golang/domain/operation/reporting_checkpoint"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	checkpointpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/reporting_checkpoint"

	reportingcheckpointdetail "github.com/erniealice/fayna-golang/domain/operation/reporting_checkpoint/detail"
	reportingcheckpointlist "github.com/erniealice/fayna-golang/domain/operation/reporting_checkpoint/list"
)

// ReportingCheckpointModuleDeps holds all dependencies for the reporting checkpoint module.
type ReportingCheckpointModuleDeps struct {
	Routes       reportingcheckpointpkg.Routes
	Labels       reportingcheckpointpkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Reporting checkpoint CRUD
	CreateReportingCheckpoint func(ctx context.Context, req *checkpointpb.CreateReportingCheckpointRequest) (*checkpointpb.CreateReportingCheckpointResponse, error)
	ReadReportingCheckpoint   func(ctx context.Context, req *checkpointpb.ReadReportingCheckpointRequest) (*checkpointpb.ReadReportingCheckpointResponse, error)
	UpdateReportingCheckpoint func(ctx context.Context, req *checkpointpb.UpdateReportingCheckpointRequest) (*checkpointpb.UpdateReportingCheckpointResponse, error)
	DeleteReportingCheckpoint func(ctx context.Context, req *checkpointpb.DeleteReportingCheckpointRequest) (*checkpointpb.DeleteReportingCheckpointResponse, error)
	ListReportingCheckpoints  func(ctx context.Context, req *checkpointpb.ListReportingCheckpointsRequest) (*checkpointpb.ListReportingCheckpointsResponse, error)

	// Audit history (optional — nil = history tab hidden/empty)
	auditlog.AuditOps
}

// ReportingCheckpointModule holds all constructed reporting checkpoint views.
type ReportingCheckpointModule struct {
	routes     reportingcheckpointpkg.Routes
	List       view.View
	Detail     view.View
	TabAction  view.View
	Add        view.View
	Edit       view.View
	Delete     view.View
	BulkDelete view.View
}

// NewReportingCheckpointModule creates a new reporting checkpoint module with all views wired.
func NewReportingCheckpointModule(deps *ReportingCheckpointModuleDeps) *ReportingCheckpointModule {
	detailDeps := &reportingcheckpointdetail.DetailViewDeps{
		AuditOps:                deps.AuditOps,
		Routes:                  deps.Routes,
		Labels:                  deps.Labels,
		CommonLabels:            deps.CommonLabels,
		TableLabels:             deps.TableLabels,
		ReadReportingCheckpoint: deps.ReadReportingCheckpoint,
	}

	// Build entity-package ModuleDeps for the exported action builders.
	entityDeps := &reportingcheckpointpkg.ModuleDeps{
		Routes:                    deps.Routes,
		Labels:                    deps.Labels,
		CommonLabels:              deps.CommonLabels,
		TableLabels:               deps.TableLabels,
		CreateReportingCheckpoint: deps.CreateReportingCheckpoint,
		ReadReportingCheckpoint:   deps.ReadReportingCheckpoint,
		UpdateReportingCheckpoint: deps.UpdateReportingCheckpoint,
		DeleteReportingCheckpoint: deps.DeleteReportingCheckpoint,
		ListReportingCheckpoints:  deps.ListReportingCheckpoints,
	}

	return &ReportingCheckpointModule{
		routes: deps.Routes,
		List: reportingcheckpointlist.NewView(&reportingcheckpointlist.ListViewDeps{
			Routes:                   deps.Routes,
			ListReportingCheckpoints: deps.ListReportingCheckpoints,
			Labels:                   deps.Labels,
			CommonLabels:             deps.CommonLabels,
			TableLabels:              deps.TableLabels,
		}),
		Detail:     reportingcheckpointdetail.NewView(detailDeps),
		TabAction:  reportingcheckpointdetail.NewTabAction(detailDeps),
		Add:        reportingcheckpointpkg.NewAddAction(entityDeps),
		Edit:       reportingcheckpointpkg.NewEditAction(entityDeps),
		Delete:     reportingcheckpointpkg.NewDeleteAction(entityDeps),
		BulkDelete: reportingcheckpointpkg.NewBulkDeleteAction(entityDeps),
	}
}

// RegisterRoutes registers all reporting checkpoint routes.
func (m *ReportingCheckpointModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)
	r.GET(m.routes.TabActionURL, m.TabAction)

	// CRUD actions (GET = drawer form, POST = process submission)
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.DeleteURL, m.Delete)
	r.POST(m.routes.BulkDeleteURL, m.BulkDelete)
}
