package outcome_criteria

import (
	"context"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
)

// ModuleDeps holds the typed closures that action builders and sub-packages need.
// Defined here (not in module.go) to avoid a self-import cycle: module.go cannot
// import its own package path.
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Outcome criteria CRUD
	CreateOutcomeCriteria func(ctx context.Context, req *criteriapb.CreateOutcomeCriteriaRequest) (*criteriapb.CreateOutcomeCriteriaResponse, error)
	ReadOutcomeCriteria   func(ctx context.Context, req *criteriapb.ReadOutcomeCriteriaRequest) (*criteriapb.ReadOutcomeCriteriaResponse, error)
	UpdateOutcomeCriteria func(ctx context.Context, req *criteriapb.UpdateOutcomeCriteriaRequest) (*criteriapb.UpdateOutcomeCriteriaResponse, error)
	DeleteOutcomeCriteria func(ctx context.Context, req *criteriapb.DeleteOutcomeCriteriaRequest) (*criteriapb.DeleteOutcomeCriteriaResponse, error)
	ListOutcomeCriterias  func(ctx context.Context, req *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error)

	// Attachment operations
	UploadFile       func(ctx context.Context, bucket, key string, content []byte, contentType string) error
	ListAttachments  func(ctx context.Context, moduleKey, foreignKey string) (*attachmentpb.ListAttachmentsResponse, error)
	CreateAttachment func(ctx context.Context, req *attachmentpb.CreateAttachmentRequest) (*attachmentpb.CreateAttachmentResponse, error)
	DeleteAttachment func(ctx context.Context, req *attachmentpb.DeleteAttachmentRequest) (*attachmentpb.DeleteAttachmentResponse, error)
	NewID            func() string
}
