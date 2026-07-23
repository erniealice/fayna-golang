package detail

import (
	"context"

	"github.com/erniealice/fayna-golang/domain/operation/job_outcome_line"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	joboutcomelinepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_line"
)

// DetailViewDeps holds view dependencies for the job outcome line detail views.
type DetailViewDeps struct {
	auditlog.AuditOps

	Routes       job_outcome_line.Routes
	Labels       job_outcome_line.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// JobOutcomeLine read
	ReadJobOutcomeLine func(ctx context.Context, req *joboutcomelinepb.ReadJobOutcomeLineRequest) (*joboutcomelinepb.ReadJobOutcomeLineResponse, error)
}
