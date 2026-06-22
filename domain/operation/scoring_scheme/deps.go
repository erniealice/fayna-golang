package scoring_scheme

import (
	"context"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	schemepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/scoring_scheme"
)

// ModuleDeps holds the typed closures that action builders and sub-packages need.
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Scoring scheme CRUD
	CreateScoringScheme func(ctx context.Context, req *schemepb.CreateScoringSchemeRequest) (*schemepb.CreateScoringSchemeResponse, error)
	ReadScoringScheme   func(ctx context.Context, req *schemepb.ReadScoringSchemeRequest) (*schemepb.ReadScoringSchemeResponse, error)
	UpdateScoringScheme func(ctx context.Context, req *schemepb.UpdateScoringSchemeRequest) (*schemepb.UpdateScoringSchemeResponse, error)
	DeleteScoringScheme func(ctx context.Context, req *schemepb.DeleteScoringSchemeRequest) (*schemepb.DeleteScoringSchemeResponse, error)
	ListScoringSchemes  func(ctx context.Context, req *schemepb.ListScoringSchemesRequest) (*schemepb.ListScoringSchemesResponse, error)
}
