package detail

import (
	"context"

	scoring_component "github.com/erniealice/fayna-golang/domain/operation/scoring_component"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	scoringpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/scoring_component"
)

// DetailViewDeps holds view dependencies for the scoring component detail views.
type DetailViewDeps struct {
	auditlog.AuditOps

	Routes       scoring_component.Routes
	Labels       scoring_component.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// ScoringComponent read
	ReadScoringComponent func(ctx context.Context, req *scoringpb.ReadScoringComponentRequest) (*scoringpb.ReadScoringComponentResponse, error)
}
