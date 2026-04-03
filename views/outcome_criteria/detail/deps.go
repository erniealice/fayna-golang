package detail

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
)

// DetailViewDeps holds view dependencies for the outcome criteria detail views.
type DetailViewDeps struct {
	auditlog.AuditOps

	Routes       fayna.OutcomeCriteriaRoutes
	Labels       fayna.OutcomeCriteriaLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Outcome criteria read
	ReadOutcomeCriteria func(ctx context.Context, req *criteriapb.ReadOutcomeCriteriaRequest) (*criteriapb.ReadOutcomeCriteriaResponse, error)
}
