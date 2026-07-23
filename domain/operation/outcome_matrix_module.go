package operation

import (
	"context"
	"log"
	"net/http"

	outcomematrixpkg "github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"
	outcomematrixaction "github.com/erniealice/fayna-golang/domain/operation/outcome_matrix/action"
	outcomematrixlist "github.com/erniealice/fayna-golang/domain/operation/outcome_matrix/list"
	outcomematrixtemplatesettings "github.com/erniealice/fayna-golang/domain/operation/outcome_matrix/template_settings"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	documenttemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/template"
	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	clientattributepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client_attribute"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	sheetbindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_document_template"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	priceschedulepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/price_schedule"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// OutcomeMatrixModuleDeps holds all dependencies for the outcome matrix module.
//
// GetOutcomeMatrix is the new espyna use case (typed against the generated
// esqyma request/response). The TaskOutcome CRUD closures back the batch-save
// write path (create + update), and ResolveStaff supplies the acting staff_id
// for both the read-only gate (view) and the IDOR guard (record action).
type OutcomeMatrixModuleDeps struct {
	Routes       outcomematrixpkg.Routes
	Labels       outcomematrixpkg.Labels
	CommonLabels pyeza.CommonLabels
	// TableLabels back the grade-sheet template-settings list table (Wave C / P4).
	TableLabels types.TableLabels

	GetOutcomeMatrix func(ctx context.Context, req *matrixpb.GetOutcomeMatrixRequest) (*matrixpb.GetOutcomeMatrixResponse, error)

	// GetOutcomeSummaryRoster — the roster-scoped composite read (20260720 P2)
	// backing the CSV "Final" export. Sourced from espyna's Service aggregate
	// (same seam as GetOutcomeMatrix). Optional/nil-safe: a nil closure 404s a
	// period=final export (no composite source), never a 500.
	GetOutcomeSummaryRoster func(ctx context.Context, req *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error)

	// --- Grade-sheet PDF render (20260720 P5) ---
	//
	// ReadJobTemplate resolves the template's job_category_id + name for the PDF
	// render context (category keys the binding; name headers the sheet).
	// GenerateDoc/GeneratePDF are the injected fycha closures (from the app
	// AppContext via infra, the report-card GenerateDoc precedent).
	// ResolveSheetTemplateBytes resolves the PUBLISHED sheet binding + downloads
	// its storage bytes (fail-loud on any miss — Q1, no embedded fallback). All
	// four are threaded onto the list PageViewDeps for the format=pdf export
	// branch. Optional/nil-safe: a nil closure fails the pdf export loud/closed.
	ReadJobTemplate           func(ctx context.Context, req *jobtemplatepb.ReadJobTemplateRequest) (*jobtemplatepb.ReadJobTemplateResponse, error)
	GenerateDoc               func(templateData []byte, data map[string]any) ([]byte, error)
	GeneratePDF               func(templateData []byte, data map[string]any) ([]byte, error)
	ResolveSheetTemplateBytes func(ctx context.Context, jobCategoryID, priceScheduleID string) ([]byte, error)

	// Per-phase approval transition use cases (plan §4.2). Back the approval-bar
	// POST forms; each carries only the trusted sheet identity (actor + workspace
	// from context). Optional/nil-safe: a nil closure fails the bar action closed.
	SubmitJobPhaseApproval  func(ctx context.Context, req *jobphasepb.SubmitJobPhaseApprovalRequest) (*jobphasepb.SubmitJobPhaseApprovalResponse, error)
	VerifyJobPhaseApproval  func(ctx context.Context, req *jobphasepb.VerifyJobPhaseApprovalRequest) (*jobphasepb.VerifyJobPhaseApprovalResponse, error)
	PublishJobPhaseApproval func(ctx context.Context, req *jobphasepb.PublishJobPhaseApprovalRequest) (*jobphasepb.PublishJobPhaseApprovalResponse, error)
	ReturnJobPhaseApproval  func(ctx context.Context, req *jobphasepb.ReturnJobPhaseApprovalRequest) (*jobphasepb.ReturnJobPhaseApprovalResponse, error)

	CreateTaskOutcome func(ctx context.Context, req *taskoutcomepb.CreateTaskOutcomeRequest) (*taskoutcomepb.CreateTaskOutcomeResponse, error)
	UpdateTaskOutcome func(ctx context.Context, req *taskoutcomepb.UpdateTaskOutcomeRequest) (*taskoutcomepb.UpdateTaskOutcomeResponse, error)
	ReadTaskOutcome   func(ctx context.Context, req *taskoutcomepb.ReadTaskOutcomeRequest) (*taskoutcomepb.ReadTaskOutcomeResponse, error)

	ResolveStaff func(ctx context.Context) (string, error)

	// ComputePhaseOutcome / ComputeJobOutcome are the inline grade-recompute
	// closures (W2 grade-sheet edit mode, Q-GSE-5). The record action calls them
	// after a successful ACADEMIC cell write to refresh the affected
	// phase_outcome_summary then job_outcome_summary, keyed off the
	// SERVER-DERIVED job_phase_id / job_id from the re-derived matrix (never a
	// browser value). Both optional/nil-safe: a nil closure → the save still
	// succeeds with ratingFresh:false (grade persisted, rating stale). Return
	// contract: (true,nil)=recomputed; (false,nil)=frozen/authoritative skip;
	// (false,err)=compute failed (stale + retryable).
	ComputePhaseOutcome func(ctx context.Context, jobPhaseID string) (bool, error)
	ComputeJobOutcome   func(ctx context.Context, jobID string) (bool, error)

	// RecomputeEligibility classifies whether a saved numeric cell drives a
	// scaled-summary recompute (graph-derived: the phase's scheme resolves a score
	// scale and the cell's criterion is in that scheme's active component graph).
	// Optional/nil-safe: a nil closure (or a lookup error) falls back to
	// numeric-type classification in the record action.
	RecomputeEligibility func(ctx context.Context, jobPhaseID string) (bool, map[string]bool, error)

	// ListClients hydrates the roster's display names (the matrix's client_id
	// rows are otherwise opaque — see list/page.go's PageViewDeps.ListClients
	// doc comment). Same closure the job drawer's client search picker
	// already uses; optional/nil-safe.
	ListClients func(ctx context.Context, req *clientpb.ListClientsRequest) (*clientpb.ListClientsResponse, error)

	// Page-header delivery-group resolution (round 4 item 2) — see
	// list/page.go's PageViewDeps doc comment for the full chain. All three
	// are already-wired top-level closures reused from elsewhere in the
	// block; optional/nil-safe.
	ListJobs                     func(ctx context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error)
	ListSubscriptionGroupMembers func(ctx context.Context, req *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error)
	ListSubscriptionGroups       func(ctx context.Context, req *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error)

	// Options — app-configured row presentation (sort/description/group_by
	// through "client_attributes.<code>" references). Zero value → flat
	// roster, rendering unchanged.
	Options outcomematrixpkg.Options

	// Row-attribute hydration backing Options. Both optional/nil-safe: nil or
	// a failed lookup disables the attribute-driven behaviors, never the page.
	ListClientAttributes     func(ctx context.Context, req *clientattributepb.ListClientAttributesRequest) (*clientattributepb.ListClientAttributesResponse, error)
	ResolveAttributeIDByCode func(ctx context.Context, code string) (string, error)

	// Header-breadcrumb back-link to the job list (the matrix's parent
	// surface). Both come from the job unit's RESOLVED routes/labels at mount
	// time, so they carry the tier's slug + wording. Optional: empty values
	// render no breadcrumb (header falls back to the title-only crumb).
	JobListURL   string // job list route pattern, "{status}" placeholder intact
	JobListLabel string // the job list's active-status heading

	// --- Grade-sheet template settings (Wave C / P4) ---
	//
	// The document_template artifact + storage closures come from the app
	// AppContext (via infra, the GenerateDoc precedent); the binding lifecycle
	// closures come from the espyna job_template_document_template use cases via
	// the block seam. The category + schedule dropdown reads back the same
	// closures the grid + job list already consume. All optional/nil-safe — a nil
	// write closure degrades the settings surface to "not configured" (never a
	// panic); the list page still renders.
	ListJobCategories      func(ctx context.Context, req *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error)
	ListPriceSchedules     func(ctx context.Context, req *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error)
	UploadTemplate         func(ctx context.Context, bucket, key string, content []byte, contentType string) error
	ListDocumentTemplates  func(ctx context.Context, req *documenttemplatepb.ListDocumentTemplatesRequest) (*documenttemplatepb.ListDocumentTemplatesResponse, error)
	CreateDocumentTemplate func(ctx context.Context, req *documenttemplatepb.CreateDocumentTemplateRequest) (*documenttemplatepb.CreateDocumentTemplateResponse, error)
	DeleteDocumentTemplate func(ctx context.Context, req *documenttemplatepb.DeleteDocumentTemplateRequest) (*documenttemplatepb.DeleteDocumentTemplateResponse, error)
	ListTemplateBindings   func(ctx context.Context, req *sheetbindingpb.ListJobTemplateDocumentTemplatesRequest) (*sheetbindingpb.ListJobTemplateDocumentTemplatesResponse, error)
	CreateTemplateBinding  func(ctx context.Context, req *sheetbindingpb.CreateJobTemplateDocumentTemplateRequest) (*sheetbindingpb.CreateJobTemplateDocumentTemplateResponse, error)
	DeleteTemplateBinding  func(ctx context.Context, req *sheetbindingpb.DeleteJobTemplateDocumentTemplateRequest) (*sheetbindingpb.DeleteJobTemplateDocumentTemplateResponse, error)
	PublishTemplateBinding func(ctx context.Context, req *sheetbindingpb.PublishJobTemplateDocumentTemplateRequest) (*sheetbindingpb.PublishJobTemplateDocumentTemplateResponse, error)
}

// OutcomeMatrixModule holds the constructed outcome matrix views.
type OutcomeMatrixModule struct {
	routes    outcomematrixpkg.Routes
	Matrix    view.View // GET — the grid page
	Record    view.View // POST — batch save
	Download  view.View // GET — export drawer form (safe method under /action/*)
	Narrative view.View // GET form / POST save — per-cell narrative drawer
	Submit    view.View // POST — approval: IN_PROGRESS → FOR_REVIEW
	Verify    view.View // POST — approval: FOR_REVIEW → VERIFIED
	Publish   view.View // POST — approval: VERIFIED → PUBLISHED
	Return    view.View // POST — approval: mixed/advanced → IN_PROGRESS

	// Export is the sheet-level CSV download — a raw GET handler (not a
	// view.View): it streams a file, so it bypasses the render pipeline but
	// is still wrapped with the ViewAdapter's RBAC context at registration
	// (same shape as outcome_summary's SectionExport).
	Export http.HandlerFunc

	// Grade-sheet template settings surface (Wave C / P4 — JOSDT parity).
	TemplateSettings view.View // GET — list page
	TemplateUpload   view.View // GET form / POST create-draft
	TemplatePublish  view.View // POST — publish draft
	TemplateDelete   view.View // POST — delete draft
}

// NewOutcomeMatrixModule creates the outcome matrix module with all views wired.
func NewOutcomeMatrixModule(deps *OutcomeMatrixModuleDeps) *OutcomeMatrixModule {
	pageDeps := &outcomematrixlist.PageViewDeps{
		Routes:                  deps.Routes,
		Labels:                  deps.Labels,
		CommonLabels:            deps.CommonLabels,
		GetOutcomeMatrix:        deps.GetOutcomeMatrix,
		GetOutcomeSummaryRoster: deps.GetOutcomeSummaryRoster,
		ResolveStaff:            deps.ResolveStaff,
		// Grade-sheet PDF render context (P5).
		ReadJobTemplate:              deps.ReadJobTemplate,
		GenerateDoc:                  deps.GenerateDoc,
		GeneratePDF:                  deps.GeneratePDF,
		ResolveSheetTemplateBytes:    deps.ResolveSheetTemplateBytes,
		ListClients:                  deps.ListClients,
		ListJobs:                     deps.ListJobs,
		ListSubscriptionGroupMembers: deps.ListSubscriptionGroupMembers,
		ListSubscriptionGroups:       deps.ListSubscriptionGroups,
		Options:                      deps.Options,
		ListClientAttributes:         deps.ListClientAttributes,
		ResolveAttributeIDByCode:     deps.ResolveAttributeIDByCode,
		JobListURL:                   deps.JobListURL,
		JobListLabel:                 deps.JobListLabel,
	}
	matrixView := outcomematrixlist.NewView(pageDeps)

	recordView := outcomematrixaction.NewRecordAction(&outcomematrixaction.Deps{
		Routes:               deps.Routes,
		Labels:               deps.Labels,
		CreateTaskOutcome:    deps.CreateTaskOutcome,
		UpdateTaskOutcome:    deps.UpdateTaskOutcome,
		ReadTaskOutcome:      deps.ReadTaskOutcome,
		GetOutcomeMatrix:     deps.GetOutcomeMatrix,
		ResolveStaff:         deps.ResolveStaff,
		ComputePhaseOutcome:  deps.ComputePhaseOutcome,
		ComputeJobOutcome:    deps.ComputeJobOutcome,
		RecomputeEligibility: deps.RecomputeEligibility,
	})

	// Per-phase approval transition handlers (the signed HTMX POST forms in the
	// approval bar). Share one TransitionDeps; each handler gates on its own
	// job_phase:<verb> and forwards the trusted sheet identity to the espyna use
	// case (authoritative).
	transitionDeps := &outcomematrixaction.TransitionDeps{
		Routes:  deps.Routes,
		Labels:  deps.Labels,
		Submit:  deps.SubmitJobPhaseApproval,
		Verify:  deps.VerifyJobPhaseApproval,
		Publish: deps.PublishJobPhaseApproval,
		Return:  deps.ReturnJobPhaseApproval,
	}

	// Export drawer GET view (20260720 Q3) — reuses the SAME GetOutcomeMatrix
	// closure the grid + CSV handler use, so its period options match the export.
	downloadView := outcomematrixaction.NewDownloadDrawer(&outcomematrixaction.DrawerDeps{
		Routes:           deps.Routes,
		Labels:           deps.Labels,
		GetOutcomeMatrix: deps.GetOutcomeMatrix,
	})

	// Per-cell narrative drawer (N-1 LOCKED): GET form + POST save on one route.
	// Reuses the SAME ResolveStaff + GetOutcomeMatrix the record action gates on
	// (byte-identical authority core) plus the task_outcome read/update use cases
	// the module already wires — writes route through UpdateTaskOutcome, never SQL.
	narrativeView := outcomematrixaction.NewNarrativeAction(&outcomematrixaction.NarrativeDeps{
		Routes:            deps.Routes,
		Labels:            deps.Labels,
		GetOutcomeMatrix:  deps.GetOutcomeMatrix,
		ResolveStaff:      deps.ResolveStaff,
		ReadTaskOutcome:   deps.ReadTaskOutcome,
		UpdateTaskOutcome: deps.UpdateTaskOutcome,
	})

	return &OutcomeMatrixModule{
		routes:    deps.Routes,
		Matrix:    matrixView,
		Record:    recordView,
		Download:  downloadView,
		Narrative: narrativeView,
		Submit:    outcomematrixaction.NewSubmitAction(transitionDeps),
		Verify:    outcomematrixaction.NewVerifyAction(transitionDeps),
		Publish:   outcomematrixaction.NewPublishAction(transitionDeps),
		Return:    outcomematrixaction.NewReturnAction(transitionDeps),
		Export:    outcomematrixlist.NewExportHandler(pageDeps),

		// Grade-sheet template settings (Wave C / P4). All four views share one
		// nil-safe Deps (the JOSDT registration parity). templateSettingsDeps maps
		// the module deps onto the view deps.
		TemplateSettings: outcomematrixtemplatesettings.NewListView(outcomeMatrixTemplateSettingsDeps(deps)),
		TemplateUpload:   outcomematrixtemplatesettings.NewUploadAction(outcomeMatrixTemplateSettingsDeps(deps)),
		TemplatePublish:  outcomematrixtemplatesettings.NewPublishAction(outcomeMatrixTemplateSettingsDeps(deps)),
		TemplateDelete:   outcomematrixtemplatesettings.NewDeleteAction(outcomeMatrixTemplateSettingsDeps(deps)),
	}
}

// outcomeMatrixTemplateSettingsDeps maps the module deps onto the
// template-settings view deps (Wave C / P4). All closures are optional/nil-safe.
func outcomeMatrixTemplateSettingsDeps(deps *OutcomeMatrixModuleDeps) *outcomematrixtemplatesettings.Deps {
	return &outcomematrixtemplatesettings.Deps{
		Routes:                 deps.Routes,
		Labels:                 deps.Labels,
		CommonLabels:           deps.CommonLabels,
		TableLabels:            deps.TableLabels,
		ListJobCategories:      deps.ListJobCategories,
		ListPriceSchedules:     deps.ListPriceSchedules,
		UploadTemplate:         deps.UploadTemplate,
		ListDocumentTemplates:  deps.ListDocumentTemplates,
		CreateDocumentTemplate: deps.CreateDocumentTemplate,
		DeleteDocumentTemplate: deps.DeleteDocumentTemplate,
		ListTemplateBindings:   deps.ListTemplateBindings,
		CreateTemplateBinding:  deps.CreateTemplateBinding,
		DeleteTemplateBinding:  deps.DeleteTemplateBinding,
		PublishTemplateBinding: deps.PublishTemplateBinding,
	}
}

// RegisterRoutes registers the outcome matrix routes.
func (m *OutcomeMatrixModule) RegisterRoutes(r view.RouteRegistrar) {
	if m.Matrix != nil {
		r.GET(m.routes.MatrixURL, m.Matrix)
	}
	if m.Record != nil {
		r.POST(m.routes.RecordURL, m.Record)
	}
	if m.Download != nil && m.routes.DownloadDrawerURL != "" {
		// GET (safe method) — registered with r.GET even though the path sits
		// under /action/*: the CSRF + action-workspace guards constrain non-safe
		// methods only, so no signed form is needed (routes.go documents this).
		r.GET(m.routes.DownloadDrawerURL, m.Download)
	}
	if m.Narrative != nil && m.routes.NarrativeURL != "" {
		// One path, two verbs (TemplateUploadURL precedent): GET renders the
		// drawer (safe method — no signed form needed), POST saves the note under
		// /action/* so it inherits the CSRF + action-workspace signature guards.
		r.GET(m.routes.NarrativeURL, m.Narrative)
		r.POST(m.routes.NarrativeURL, m.Narrative)
	}
	if m.Submit != nil {
		r.POST(m.routes.SubmitURL, m.Submit)
	}
	if m.Verify != nil {
		r.POST(m.routes.VerifyURL, m.Verify)
	}
	if m.Publish != nil {
		r.POST(m.routes.PublishURL, m.Publish)
	}
	if m.Return != nil {
		r.POST(m.routes.ReturnURL, m.Return)
	}
	if m.Export != nil && m.routes.ExportURL != "" {
		// Raw (non-view) route — the registrar's HandleFunc path wraps it with
		// the ViewAdapter's RBAC context injection (WrapHandler), so the
		// handler's view.GetUserPermissions gate observes real permissions
		// (same shape as outcome_summary's SectionExport registration).
		if rr, ok := r.(interface {
			HandleFunc(method, path string, handler http.HandlerFunc, middlewares ...string)
		}); ok {
			rr.HandleFunc("GET", m.routes.ExportURL, m.Export)
		} else {
			log.Printf("outcome matrix: RouteRegistrar does not support HandleFunc — skipping GET %s", m.routes.ExportURL)
		}
	}

	// Grade-sheet template settings (Wave C / P4): list page + upload drawer (GET
	// form / POST create) + publish (POST) + delete (POST). Gated inside each view
	// (list → :list, mutations → :create/:update/:delete). JOSDT registration
	// parity. The settings GET stays OUTSIDE /action/ (safe method); the mutations
	// live under /action/ so they inherit the CSRF + signed-workspace guards.
	if m.TemplateSettings != nil && m.routes.TemplateSettingsURL != "" {
		r.GET(m.routes.TemplateSettingsURL, m.TemplateSettings)
	}
	if m.TemplateUpload != nil && m.routes.TemplateUploadURL != "" {
		r.GET(m.routes.TemplateUploadURL, m.TemplateUpload)
		r.POST(m.routes.TemplateUploadURL, m.TemplateUpload)
	}
	if m.TemplatePublish != nil && m.routes.TemplatePublishURL != "" {
		r.POST(m.routes.TemplatePublishURL, m.TemplatePublish)
	}
	if m.TemplateDelete != nil && m.routes.TemplateDeleteURL != "" {
		r.POST(m.routes.TemplateDeleteURL, m.TemplateDelete)
	}
}
