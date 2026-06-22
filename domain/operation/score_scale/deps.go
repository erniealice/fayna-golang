package score_scale

import (
	"context"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	scalepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale"
)

// ModuleDeps holds the typed closures that action builders and sub-packages need.
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Score Scale CRUD
	CreateScoreScale func(ctx context.Context, req *scalepb.CreateScoreScaleRequest) (*scalepb.CreateScoreScaleResponse, error)
	ReadScoreScale   func(ctx context.Context, req *scalepb.ReadScoreScaleRequest) (*scalepb.ReadScoreScaleResponse, error)
	UpdateScoreScale func(ctx context.Context, req *scalepb.UpdateScoreScaleRequest) (*scalepb.UpdateScoreScaleResponse, error)
	DeleteScoreScale func(ctx context.Context, req *scalepb.DeleteScoreScaleRequest) (*scalepb.DeleteScoreScaleResponse, error)
	ListScoreScales  func(ctx context.Context, req *scalepb.ListScoreScalesRequest) (*scalepb.ListScoreScalesResponse, error)
}
