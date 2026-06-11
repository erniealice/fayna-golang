package task_outcome

import (
	"context"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	outcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
)

// ModuleDeps holds the typed closures that action builders and sub-packages need.
// Defined here (not in module.go) to avoid a self-import cycle: module.go cannot
// import its own package path.
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Task outcome CRUD
	CreateTaskOutcome func(ctx context.Context, req *outcomepb.CreateTaskOutcomeRequest) (*outcomepb.CreateTaskOutcomeResponse, error)
	ReadTaskOutcome   func(ctx context.Context, req *outcomepb.ReadTaskOutcomeRequest) (*outcomepb.ReadTaskOutcomeResponse, error)
	UpdateTaskOutcome func(ctx context.Context, req *outcomepb.UpdateTaskOutcomeRequest) (*outcomepb.UpdateTaskOutcomeResponse, error)
	DeleteTaskOutcome func(ctx context.Context, req *outcomepb.DeleteTaskOutcomeRequest) (*outcomepb.DeleteTaskOutcomeResponse, error)
	ListTaskOutcomes  func(ctx context.Context, req *outcomepb.ListTaskOutcomesRequest) (*outcomepb.ListTaskOutcomesResponse, error)

	// Outcome criteria read (for linking criteria details)
	ReadOutcomeCriteria func(ctx context.Context, req *criteriapb.ReadOutcomeCriteriaRequest) (*criteriapb.ReadOutcomeCriteriaResponse, error)

	// Attachment operations
	UploadFile       func(ctx context.Context, bucket, key string, content []byte, contentType string) error
	ListAttachments  func(ctx context.Context, moduleKey, foreignKey string) (*attachmentpb.ListAttachmentsResponse, error)
	CreateAttachment func(ctx context.Context, req *attachmentpb.CreateAttachmentRequest) (*attachmentpb.CreateAttachmentResponse, error)
	DeleteAttachment func(ctx context.Context, req *attachmentpb.DeleteAttachmentRequest) (*attachmentpb.DeleteAttachmentResponse, error)
	NewID            func() string
}
