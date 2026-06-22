package detail

import (
	"context"

	score_scale_band "github.com/erniealice/fayna-golang/domain/operation/score_scale_band"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	ssbpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale_band"
)

// DetailViewDeps holds view dependencies for the score scale band detail views.
type DetailViewDeps struct {
	Routes       score_scale_band.Routes
	Labels       score_scale_band.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Score scale band read
	ReadScoreScaleBand func(ctx context.Context, req *ssbpb.ReadScoreScaleBandRequest) (*ssbpb.ReadScoreScaleBandResponse, error)
}
