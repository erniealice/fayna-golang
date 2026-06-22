package scoring_component

import (
	"context"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	scoringpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/scoring_component"
)

// ModuleDeps holds the typed closures that action builders and sub-packages need.
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// ScoringComponent CRUD
	CreateScoringComponent func(ctx context.Context, req *scoringpb.CreateScoringComponentRequest) (*scoringpb.CreateScoringComponentResponse, error)
	ReadScoringComponent   func(ctx context.Context, req *scoringpb.ReadScoringComponentRequest) (*scoringpb.ReadScoringComponentResponse, error)
	UpdateScoringComponent func(ctx context.Context, req *scoringpb.UpdateScoringComponentRequest) (*scoringpb.UpdateScoringComponentResponse, error)
	DeleteScoringComponent func(ctx context.Context, req *scoringpb.DeleteScoringComponentRequest) (*scoringpb.DeleteScoringComponentResponse, error)
	ListScoringComponents  func(ctx context.Context, req *scoringpb.ListScoringComponentsRequest) (*scoringpb.ListScoringComponentsResponse, error)
}
