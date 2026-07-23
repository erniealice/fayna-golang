package detail

import (
	"context"

	jtr "github.com/erniealice/fayna-golang/domain/operation/job_template_relation"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	relationpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_relation"
)

// DetailViewDeps holds view dependencies for the job template relation detail views.
type DetailViewDeps struct {
	auditlog.AuditOps

	Routes       jtr.Routes
	Labels       jtr.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// JobTemplateRelation read
	ReadJobTemplateRelation func(ctx context.Context, req *relationpb.ReadJobTemplateRelationRequest) (*relationpb.ReadJobTemplateRelationResponse, error)
}
