package detail

import (
	"context"

	scc "github.com/erniealice/fayna-golang/domain/operation/scoring_component_criteria"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	sccpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/scoring_component_criteria"
)

// DetailViewDeps holds view dependencies for the scoring component criteria detail views.
type DetailViewDeps struct {
	auditlog.AuditOps

	Routes       scc.Routes
	Labels       scc.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// ScoringComponentCriteria read
	ReadScoringComponentCriteria func(ctx context.Context, req *sccpb.ReadScoringComponentCriteriaRequest) (*sccpb.ReadScoringComponentCriteriaResponse, error)
}
