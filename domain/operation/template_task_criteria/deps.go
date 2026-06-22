package template_task_criteria

import (
	"context"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	ttcpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
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
}
