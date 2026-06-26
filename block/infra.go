package block

import (
	"context"

	"github.com/erniealice/espyna-golang/ports"
	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
)

// Infra carries the subset of AppContext that view modules need beyond
// the typed UseCases: attachment ops, reference checker, DB for search
// handlers, and cross-package URL patterns. Built once by service-admin
// and passed into each catalog binder.
type Infra struct {
	UploadFile       func(context.Context, string, string, []byte, string) error
	ListAttachments  func(context.Context, string, string) (*attachmentpb.ListAttachmentsResponse, error)
	CreateAttachment func(context.Context, *attachmentpb.CreateAttachmentRequest) (*attachmentpb.CreateAttachmentResponse, error)
	DeleteAttachment func(context.Context, *attachmentpb.DeleteAttachmentRequest) (*attachmentpb.DeleteAttachmentResponse, error)
	NewAttachmentID  func() string
	RefChecker       ports.Checker
	DB               listSimpler
	SubscriptionDetailURL string
}
