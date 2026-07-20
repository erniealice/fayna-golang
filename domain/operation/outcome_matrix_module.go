package operation

import (
	"context"
	"log"
	"net/http"

	outcomematrixpkg "github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"
	outcomematrixaction "github.com/erniealice/fayna-golang/domain/operation/outcome_matrix/action"
	outcomematrixlist "github.com/erniealice/fayna-golang/domain/operation/outcome_matrix/list"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/view"

	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	clientattributepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client_attribute"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
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

	GetOutcomeMatrix func(ctx context.Context, req *matrixpb.GetOutcomeMatrixRequest) (*matrixpb.GetOutcomeMatrixResponse, error)

	// GetOutcomeSummaryRoster — the roster-scoped composite read (20260720 P2)
	// backing the CSV "Final" export. Sourced from espyna's Service aggregate
	// (same seam as GetOutcomeMatrix). Optional/nil-safe: a nil closure 404s a
	// period=final export (no composite source), never a 500.
	GetOutcomeSummaryRoster func(ctx context.Context, req *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error)

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
}

// OutcomeMatrixModule holds the constructed outcome matrix views.
type OutcomeMatrixModule struct {
	routes   outcomematrixpkg.Routes
	Matrix   view.View // GET — the grid page
	Record   view.View // POST — batch save
	Download view.View // GET — export drawer form (safe method under /action/*)
	Submit   view.View // POST — approval: IN_PROGRESS → FOR_REVIEW
	Verify   view.View // POST — approval: FOR_REVIEW → VERIFIED
	Publish  view.View // POST — approval: VERIFIED → PUBLISHED
	Return   view.View // POST — approval: mixed/advanced → IN_PROGRESS

	// Export is the sheet-level CSV download — a raw GET handler (not a
	// view.View): it streams a file, so it bypasses the render pipeline but
	// is still wrapped with the ViewAdapter's RBAC context at registration
	// (same shape as outcome_summary's SectionExport).
	Export http.HandlerFunc
}

// NewOutcomeMatrixModule creates the outcome matrix module with all views wired.
func NewOutcomeMatrixModule(deps *OutcomeMatrixModuleDeps) *OutcomeMatrixModule {
	pageDeps := &outcomematrixlist.PageViewDeps{
		Routes:                       deps.Routes,
		Labels:                       deps.Labels,
		CommonLabels:                 deps.CommonLabels,
		GetOutcomeMatrix:             deps.GetOutcomeMatrix,
		GetOutcomeSummaryRoster:      deps.GetOutcomeSummaryRoster,
		ResolveStaff:                 deps.ResolveStaff,
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

	return &OutcomeMatrixModule{
		routes:   deps.Routes,
		Matrix:   matrixView,
		Record:   recordView,
		Download: downloadView,
		Submit:   outcomematrixaction.NewSubmitAction(transitionDeps),
		Verify:   outcomematrixaction.NewVerifyAction(transitionDeps),
		Publish:  outcomematrixaction.NewPublishAction(transitionDeps),
		Return:   outcomematrixaction.NewReturnAction(transitionDeps),
		Export:   outcomematrixlist.NewExportHandler(pageDeps),
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
}
