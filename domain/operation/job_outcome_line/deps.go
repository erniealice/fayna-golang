package job_outcome_line

import (
	"context"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	joboutcomelinepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_line"
)

// ModuleDeps holds the typed closures that action builders and sub-packages need.
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// JobOutcomeLine CRUD
	CreateJobOutcomeLine func(ctx context.Context, req *joboutcomelinepb.CreateJobOutcomeLineRequest) (*joboutcomelinepb.CreateJobOutcomeLineResponse, error)
	ReadJobOutcomeLine   func(ctx context.Context, req *joboutcomelinepb.ReadJobOutcomeLineRequest) (*joboutcomelinepb.ReadJobOutcomeLineResponse, error)
	UpdateJobOutcomeLine func(ctx context.Context, req *joboutcomelinepb.UpdateJobOutcomeLineRequest) (*joboutcomelinepb.UpdateJobOutcomeLineResponse, error)
	DeleteJobOutcomeLine func(ctx context.Context, req *joboutcomelinepb.DeleteJobOutcomeLineRequest) (*joboutcomelinepb.DeleteJobOutcomeLineResponse, error)
	ListJobOutcomeLines  func(ctx context.Context, req *joboutcomelinepb.ListJobOutcomeLinesRequest) (*joboutcomelinepb.ListJobOutcomeLinesResponse, error)
}
