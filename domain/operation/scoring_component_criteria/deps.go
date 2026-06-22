package scoring_component_criteria

import (
	"context"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	sccpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/scoring_component_criteria"
)

// ModuleDeps holds the typed closures that action builders and sub-packages need.
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// ScoringComponentCriteria CRUD
	CreateScoringComponentCriteria func(ctx context.Context, req *sccpb.CreateScoringComponentCriteriaRequest) (*sccpb.CreateScoringComponentCriteriaResponse, error)
	ReadScoringComponentCriteria   func(ctx context.Context, req *sccpb.ReadScoringComponentCriteriaRequest) (*sccpb.ReadScoringComponentCriteriaResponse, error)
	UpdateScoringComponentCriteria func(ctx context.Context, req *sccpb.UpdateScoringComponentCriteriaRequest) (*sccpb.UpdateScoringComponentCriteriaResponse, error)
	DeleteScoringComponentCriteria func(ctx context.Context, req *sccpb.DeleteScoringComponentCriteriaRequest) (*sccpb.DeleteScoringComponentCriteriaResponse, error)
	ListScoringComponentCriterias  func(ctx context.Context, req *sccpb.ListScoringComponentCriteriasRequest) (*sccpb.ListScoringComponentCriteriasResponse, error)
}
