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

	// GeneratePDF wraps the fycha DocumentService.ProcessBytesToPDF closure
	// (template bytes + data map → rendered DOCX → PDF via LibreOffice), injected
	// from the app AppContext (ctx.GeneratePDF) — a SECOND closure mirroring
	// GenerateDoc exactly (fayna does NOT import fycha; this is a bare-signature
	// injected closure). Nil when the app did not wire PDF generation — the
	// report-card download's ?format=pdf branch then fails closed with a 503,
	// while the DOCX baseline (GenerateDoc) is unaffected.
	GeneratePDF func(templateData []byte, data map[string]any) ([]byte, error)

	// ComputePhaseOutcome / ComputeJobOutcome are the inline grade-recompute
	// closures (W2 grade-sheet edit mode, Q-GSE-5), injected from the app
	// AppContext (ctx.ComputePhaseOutcome / .ComputeJobOutcome — the GenerateDoc
	// precedent). The outcome_matrix record action calls them after a successful
	// ACADEMIC cell write to refresh the affected phase_outcome_summary then
	// job_outcome_summary, using the SERVER-DERIVED job_phase_id / job_id (never
	// a browser value). Return contract: (true,nil)=recomputed → ratingFresh;
	// (false,nil)=frozen/authoritative skip → ratingFresh (not stale);
	// (false,err)=compute failed → ratingFresh:false (grade saved, rating stale
	// + retryable). Nil when the app did not wire recompute (non-postgres /
	// service-admin without grade use-cases) — the save still succeeds with
	// ratingFresh:false, never a panic or 500.
	ComputePhaseOutcome func(ctx context.Context, jobPhaseID string) (bool, error)
	ComputeJobOutcome   func(ctx context.Context, jobID string) (bool, error)

	// RecomputeEligibility classifies whether a saved numeric cell drives a
	// scaled-summary recompute (graph-derived: the phase's scheme resolves a score
	// scale and the cell's criterion is in that scheme's active component graph),
	// returning eligibility + the in-scope criterion id set. Injected from the app
	// AppContext (ctx.RecomputeEligibility). Nil-safe: when unwired (or on a lookup
	// error) the record action falls back to numeric-type classification.
	RecomputeEligibility func(ctx context.Context, jobPhaseID string) (bool, map[string]bool, error)

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
