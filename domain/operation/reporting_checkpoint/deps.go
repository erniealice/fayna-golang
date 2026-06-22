package reporting_checkpoint

import (
	"context"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	checkpointpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/reporting_checkpoint"
)

// ModuleDeps holds the typed closures that action builders and sub-packages need.
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Reporting checkpoint CRUD
	CreateReportingCheckpoint func(ctx context.Context, req *checkpointpb.CreateReportingCheckpointRequest) (*checkpointpb.CreateReportingCheckpointResponse, error)
	ReadReportingCheckpoint   func(ctx context.Context, req *checkpointpb.ReadReportingCheckpointRequest) (*checkpointpb.ReadReportingCheckpointResponse, error)
	UpdateReportingCheckpoint func(ctx context.Context, req *checkpointpb.UpdateReportingCheckpointRequest) (*checkpointpb.UpdateReportingCheckpointResponse, error)
	DeleteReportingCheckpoint func(ctx context.Context, req *checkpointpb.DeleteReportingCheckpointRequest) (*checkpointpb.DeleteReportingCheckpointResponse, error)
	ListReportingCheckpoints  func(ctx context.Context, req *checkpointpb.ListReportingCheckpointsRequest) (*checkpointpb.ListReportingCheckpointsResponse, error)
}
