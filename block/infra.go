package block

import (
	"context"

	"github.com/erniealice/espyna-golang/ports"
	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	documenttemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/template"
)

// Infra carries the subset of AppContext that view modules need beyond
// the typed UseCases: attachment ops, reference checker, DB for search
// handlers, and cross-package URL patterns. Built once by service-admin
// and passed into each catalog binder.
type Infra struct {
	UploadFile            func(context.Context, string, string, []byte, string) error
	ListAttachments       func(context.Context, string, string) (*attachmentpb.ListAttachmentsResponse, error)
	CreateAttachment      func(context.Context, *attachmentpb.CreateAttachmentRequest) (*attachmentpb.CreateAttachmentResponse, error)
	DeleteAttachment      func(context.Context, *attachmentpb.DeleteAttachmentRequest) (*attachmentpb.DeleteAttachmentResponse, error)
	NewAttachmentID       func() string
	RefChecker            ports.Checker
	DB                    listSimpler
	SubscriptionDetailURL string
	// GenerateDoc wraps the fycha DocumentService.ProcessBytes closure (template
	// bytes + data map → processed .docx), injected from the app AppContext
	// (ctx.GenerateDoc). Nil when the app did not wire document generation —
	// consuming units degrade the download surface to a 503, never a panic.
	GenerateDoc func(templateData []byte, data map[string]any) ([]byte, error)

	// ResolveTemplateBytes resolves the applicable published report-card template
	// binding for a card's price_schedule and returns the bound template's storage
	// bytes (binding resolver ∘ storage download, built in the app container).
	// Returns (nil, nil) when no binding is configured or the object is
	// unavailable → the document unit falls back to its embedded template (no
	// download regression). Nil when the app did not wire the resolver.
	ResolveTemplateBytes func(ctx context.Context, priceScheduleID string) ([]byte, error)

	// Report-card template settings (TB3) artifact closures, sourced from the app
	// AppContext (ctx.UploadTemplate / ctx.ListDocTemplates / ctx.CreateDocTemplate,
	// same as GenerateDoc/ResolveTemplateBytes). All optional/nil-safe — a nil
	// closure degrades the upload path to "not configured".
	UploadTemplate    func(ctx context.Context, bucket, key string, content []byte, contentType string) error
	ListDocTemplates  func(ctx context.Context, req *documenttemplatepb.ListDocumentTemplatesRequest) (*documenttemplatepb.ListDocumentTemplatesResponse, error)
	CreateDocTemplate func(ctx context.Context, req *documenttemplatepb.CreateDocumentTemplateRequest) (*documenttemplatepb.CreateDocumentTemplateResponse, error)
}
