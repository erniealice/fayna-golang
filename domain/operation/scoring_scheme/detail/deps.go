package detail

import (
	"context"

	scoring_scheme "github.com/erniealice/fayna-golang/domain/operation/scoring_scheme"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	schemepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/scoring_scheme"
)

// DetailViewDeps holds view dependencies for the scoring scheme detail views.
type DetailViewDeps struct {
	auditlog.AuditOps

	Routes       scoring_scheme.Routes
	Labels       scoring_scheme.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Scoring scheme read
	ReadScoringScheme func(ctx context.Context, req *schemepb.ReadScoringSchemeRequest) (*schemepb.ReadScoringSchemeResponse, error)
}
