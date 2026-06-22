package detail

import (
	"context"

	ttc "github.com/erniealice/fayna-golang/domain/operation/template_task_criteria"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	ttcpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
)

// DetailViewDeps holds view dependencies for the template task criteria detail views.
type DetailViewDeps struct {
	auditlog.AuditOps

	Routes       ttc.Routes
	Labels       ttc.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// TemplateTaskCriteria read
	ReadTemplateTaskCriteria func(ctx context.Context, req *ttcpb.ReadTemplateTaskCriteriaRequest) (*ttcpb.ReadTemplateTaskCriteriaResponse, error)
}
