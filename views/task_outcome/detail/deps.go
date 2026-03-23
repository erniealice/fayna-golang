package detail

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/hybra-golang/views/auditlog"

	outcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
)

// DetailViewDeps holds view dependencies for the task outcome detail views.
type DetailViewDeps struct {
	auditlog.AuditOps

	Routes       fayna.TaskOutcomeRoutes
	Labels       fayna.TaskOutcomeLabels
	CommonLabels pyeza.CommonLabels

	// Task outcome read
	ReadTaskOutcome func(ctx context.Context, req *outcomepb.ReadTaskOutcomeRequest) (*outcomepb.ReadTaskOutcomeResponse, error)
}
