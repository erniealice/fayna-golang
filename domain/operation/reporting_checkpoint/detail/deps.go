package detail

import (
	"context"

	reporting_checkpoint "github.com/erniealice/fayna-golang/domain/operation/reporting_checkpoint"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	checkpointpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/reporting_checkpoint"
)

// DetailViewDeps holds view dependencies for the reporting checkpoint detail views.
type DetailViewDeps struct {
	auditlog.AuditOps

	Routes       reporting_checkpoint.Routes
	Labels       reporting_checkpoint.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Reporting checkpoint read
	ReadReportingCheckpoint func(ctx context.Context, req *checkpointpb.ReadReportingCheckpointRequest) (*checkpointpb.ReadReportingCheckpointResponse, error)
}
