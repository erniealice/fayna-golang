// Package action contains HTTP/HTMX handlers for the job_template_phase drawer-only module.
// Dependency-bearing helpers that need full Deps live here; pure-function builders live
// in the sibling form/ package.
package action

import (
	"context"

	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	scoringschemepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/scoring_scheme"
	"github.com/erniealice/fayna-golang/domain/operation/job_template_phase"
)

// Deps holds dependencies shared across all job_template_phase action handlers.
type Deps struct {
	Routes job_template_phase.Routes
	Labels job_template_phase.Labels

	// Job template phase CRUD
	CreateJobTemplatePhase func(ctx context.Context, req *jobtemplatephasepb.CreateJobTemplatePhaseRequest) (*jobtemplatephasepb.CreateJobTemplatePhaseResponse, error)
	ReadJobTemplatePhase   func(ctx context.Context, req *jobtemplatephasepb.ReadJobTemplatePhaseRequest) (*jobtemplatephasepb.ReadJobTemplatePhaseResponse, error)
	UpdateJobTemplatePhase func(ctx context.Context, req *jobtemplatephasepb.UpdateJobTemplatePhaseRequest) (*jobtemplatephasepb.UpdateJobTemplatePhaseResponse, error)
	DeleteJobTemplatePhase func(ctx context.Context, req *jobtemplatephasepb.DeleteJobTemplatePhaseRequest) (*jobtemplatephasepb.DeleteJobTemplatePhaseResponse, error)

	// ListScoringSchemes populates the Grading/Scoring Scheme picker (optional
	// FK). Nil-safe: the picker renders with no options.
	ListScoringSchemes func(ctx context.Context, req *scoringschemepb.ListScoringSchemesRequest) (*scoringschemepb.ListScoringSchemesResponse, error)

	// ResourceSearchURL for the resource picker in the Add/Edit drawer.
	ResourceSearchURL string
}
