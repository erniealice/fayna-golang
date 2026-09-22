package template_task_criteria

import (
	"context"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplateTaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	scorescalepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale"
	scorescalebandpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale_band"
	ttcpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
	ttcrdpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria_rating_description"
)

// ModuleDeps holds the typed closures that action builders and sub-packages need.
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// TemplateTaskCriteria CRUD
	CreateTemplateTaskCriteria func(ctx context.Context, req *ttcpb.CreateTemplateTaskCriteriaRequest) (*ttcpb.CreateTemplateTaskCriteriaResponse, error)
	ReadTemplateTaskCriteria   func(ctx context.Context, req *ttcpb.ReadTemplateTaskCriteriaRequest) (*ttcpb.ReadTemplateTaskCriteriaResponse, error)
	UpdateTemplateTaskCriteria func(ctx context.Context, req *ttcpb.UpdateTemplateTaskCriteriaRequest) (*ttcpb.UpdateTemplateTaskCriteriaResponse, error)
	DeleteTemplateTaskCriteria func(ctx context.Context, req *ttcpb.DeleteTemplateTaskCriteriaRequest) (*ttcpb.DeleteTemplateTaskCriteriaResponse, error)
	ListTemplateTaskCriterias  func(ctx context.Context, req *ttcpb.ListTemplateTaskCriteriasRequest) (*ttcpb.ListTemplateTaskCriteriasResponse, error)

	// ListOutcomeCriterias populates the Outcome Criteria picker. Optional —
	// nil-safe (falls back to a raw-id text input; see form/options.go).
	ListOutcomeCriterias func(ctx context.Context, req *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error)

	// ListPhasesByJobTemplate + ListTasksByPhase back the Job Template Task
	// picker's template-scoped walk when the drawer is opened from a
	// job_template detail Standards tab (?job_template_id=). Both optional —
	// nil-safe (falls back to a raw-id text input).
	ListPhasesByJobTemplate func(ctx context.Context, req *jobtemplatephasepb.ListByJobTemplateRequest) (*jobtemplatephasepb.ListByJobTemplateResponse, error)
	ListTasksByPhase        func(ctx context.Context, req *jobtemplateTaskpb.ListJobTemplateTasksByPhaseRequest) (*jobtemplateTaskpb.ListJobTemplateTasksByPhaseResponse, error)

	// Rating configuration sources and binding-specific description CRUD.
	ListScoreScales                                                  func(ctx context.Context, req *scorescalepb.ListScoreScalesRequest) (*scorescalepb.ListScoreScalesResponse, error)
	ListScoreScaleBands                                              func(ctx context.Context, req *scorescalebandpb.ListScoreScaleBandsRequest) (*scorescalebandpb.ListScoreScaleBandsResponse, error)
	CreateTemplateTaskCriteriaRatingDescription                      func(ctx context.Context, req *ttcrdpb.CreateTemplateTaskCriteriaRatingDescriptionRequest) (*ttcrdpb.CreateTemplateTaskCriteriaRatingDescriptionResponse, error)
	ReadTemplateTaskCriteriaRatingDescription                        func(ctx context.Context, req *ttcrdpb.ReadTemplateTaskCriteriaRatingDescriptionRequest) (*ttcrdpb.ReadTemplateTaskCriteriaRatingDescriptionResponse, error)
	UpdateTemplateTaskCriteriaRatingDescription                      func(ctx context.Context, req *ttcrdpb.UpdateTemplateTaskCriteriaRatingDescriptionRequest) (*ttcrdpb.UpdateTemplateTaskCriteriaRatingDescriptionResponse, error)
	DeleteTemplateTaskCriteriaRatingDescription                      func(ctx context.Context, req *ttcrdpb.DeleteTemplateTaskCriteriaRatingDescriptionRequest) (*ttcrdpb.DeleteTemplateTaskCriteriaRatingDescriptionResponse, error)
	ListTemplateTaskCriteriaRatingDescriptionsByTemplateTaskCriteria func(ctx context.Context, req *ttcrdpb.ListTemplateTaskCriteriaRatingDescriptionsByTemplateTaskCriteriaRequest) (*ttcrdpb.ListTemplateTaskCriteriaRatingDescriptionsByTemplateTaskCriteriaResponse, error)
}
