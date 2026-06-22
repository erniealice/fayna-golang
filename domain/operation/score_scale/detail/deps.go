package detail

import (
	"context"

	score_scale "github.com/erniealice/fayna-golang/domain/operation/score_scale"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	scalepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale"
)

// DetailViewDeps holds view dependencies for the score scale detail views.
type DetailViewDeps struct {
	Routes       score_scale.Routes
	Labels       score_scale.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Score scale read
	ReadScoreScale func(ctx context.Context, req *scalepb.ReadScoreScaleRequest) (*scalepb.ReadScoreScaleResponse, error)
}
