package detail

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/hybra-golang/views/attachment"
	"github.com/erniealice/hybra-golang/views/auditlog"
	"github.com/erniealice/pyeza-golang/types"

	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
)

// DetailViewDeps holds view dependencies.
type DetailViewDeps struct {
	attachment.AttachmentOps
	auditlog.AuditOps

	Routes                  fayna.JobTemplateRoutes
	ReadJobTemplate         func(ctx context.Context, req *jobtemplatepb.ReadJobTemplateRequest) (*jobtemplatepb.ReadJobTemplateResponse, error)
	ListPhasesByJobTemplate func(ctx context.Context, req *jobtemplatephasepb.ListByJobTemplateRequest) (*jobtemplatephasepb.ListByJobTemplateResponse, error)
	Labels                  fayna.JobTemplateLabels
	CommonLabels            pyeza.CommonLabels
	TableLabels             types.TableLabels
}
