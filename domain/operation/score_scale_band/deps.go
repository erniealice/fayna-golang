package score_scale_band

import (
	"context"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	ssbpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale_band"
)

// ModuleDeps holds the typed closures that action builders and sub-packages need.
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Score scale band CRUD
	CreateScoreScaleBand func(ctx context.Context, req *ssbpb.CreateScoreScaleBandRequest) (*ssbpb.CreateScoreScaleBandResponse, error)
	ReadScoreScaleBand   func(ctx context.Context, req *ssbpb.ReadScoreScaleBandRequest) (*ssbpb.ReadScoreScaleBandResponse, error)
	UpdateScoreScaleBand func(ctx context.Context, req *ssbpb.UpdateScoreScaleBandRequest) (*ssbpb.UpdateScoreScaleBandResponse, error)
	DeleteScoreScaleBand func(ctx context.Context, req *ssbpb.DeleteScoreScaleBandRequest) (*ssbpb.DeleteScoreScaleBandResponse, error)
	ListScoreScaleBands  func(ctx context.Context, req *ssbpb.ListScoreScaleBandsRequest) (*ssbpb.ListScoreScaleBandsResponse, error)
}
