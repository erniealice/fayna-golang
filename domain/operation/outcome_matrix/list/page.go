package list

import (
	"context"
	"log"
	"slices"
	"strconv"
	"strings"

	deliverygroup "github.com/erniealice/fayna-golang/domain/operation/deliverygroup"
	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	clientattributepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client_attribute"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	outcomecriteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// clientNamePageLimit chunks the roster's client-id set into ListFilter(IN)
// batches so each ListClients call's result set stays under the adapter's
// 100-row default (the generic dbOps.List cap — see
// job/list/template_summary.go's identical pattern for
// subscription_group_member). The matrix roster is one job_template's scoped
// students (~26-30/section observed on education1), so this is one call in
// practice; chunking keeps it correct if a scope=ALL admin roster ever grows
// past 100.
const clientNamePageLimit = 100

// Scope-toggle permission (same as grade_sheet.go's superAdminScopeCode). A
// teacher-only principal never gets the widened "all clients" view.
const (
	scopeEntity = "workspace"
	scopeAction = "list"
)

// PageViewDeps holds view dependencies for the outcome matrix page.
type PageViewDeps struct {
	Routes       outcome_matrix.Routes
	Labels       outcome_matrix.Labels
	CommonLabels pyeza.CommonLabels

	// GetOutcomeMatrix — the new espyna use case (typed against the generated
	// esqyma request/response). Wired via the module Deps, never raw SQL.
	GetOutcomeMatrix func(ctx context.Context, req *matrixpb.GetOutcomeMatrixRequest) (*matrixpb.GetOutcomeMatrixResponse, error)

	// GetOutcomeSummaryRoster — the roster-scoped composite read (P2) backing
	// the CSV "Final" export (student · per-phase final · year final). Gates on
	// job_outcome_summary:list inside the espyna use case; reads stored values
	// verbatim (D8). Optional/nil-safe: nil → a period=final export 404s (no
	// composite source), never a 500.
	GetOutcomeSummaryRoster func(ctx context.Context, req *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error)

	// ResolveStaff maps the acting session user → staff_id ("" == no staff
	// identity, fail-closed). Wired via the module Deps (grade_sheet.go's
	// resolveStaff precedent), never raw SQL from the view.
	ResolveStaff func(ctx context.Context) (string, error)

	// --- Grade-sheet PDF render context (20260720 P5) ---
	//
	// ReadJobTemplate resolves the template's job_category_id (proto field 32) +
	// name for the PDF render context (the category axis keys the template
	// binding; the name headers the sheet + filenames). Sourced from the espyna
	// job_template ReadJobTemplate use case via the block seam (the ListJobs
	// precedent) — never raw SQL. Optional/nil-safe: nil → a format=pdf export
	// fails loud (no category to resolve a binding), never a 500.
	ReadJobTemplate func(ctx context.Context, req *jobtemplatepb.ReadJobTemplateRequest) (*jobtemplatepb.ReadJobTemplateResponse, error)

	// ResolveSheetTemplateBytes resolves the applicable PUBLISHED grade-sheet
	// template binding (FindApplicableJobTemplateDocumentTemplate on
	// job_category_id + price_schedule_id + document_purpose='outcome_matrix') and
	// downloads its storage bytes. Built in the app container (resolver ∘ storage
	// download); ANY miss returns (nil, nil). Asymmetric to the report-card
	// ResolveTemplateBytes BY DESIGN (Q1 / entities.html §5): the grade-sheet
	// handler treats nil bytes as FAIL-LOUD ("no template configured", 503), never
	// an embedded fallback. Nil closure → the same fail-loud 503.
	ResolveSheetTemplateBytes func(ctx context.Context, jobCategoryID, priceScheduleID string) ([]byte, error)

	// GenerateDoc / GeneratePDF wrap the injected fycha DocumentService closures
	// (template bytes + data map → .docx / → .pdf via LibreOffice). fayna does NOT
	// import fycha — these are bare-signature closures threaded from the app
	// container through the block Infra (the report-card GenerateDoc precedent).
	// The PDF export uses GeneratePDF (which renders the DOCX then converts in one
	// soffice call). Nil GeneratePDF → the format=pdf export 503s (not configured).
	GenerateDoc func(templateData []byte, data map[string]any) ([]byte, error)
	GeneratePDF func(templateData []byte, data map[string]any) ([]byte, error)

	// ListClients hydrates the roster's display names. GetOutcomeMatrix
	// deliberately returns an OPAQUE client_id as ClientLabel (espyna
	// outcome_matrix_query.go:364 "opaque id; the view resolves a display
	// name") — this closure is that resolution. Optional/nil-safe: nil ->
	// rows fall back to a truncated id (short()), same as before this fix.
	// Same closure the job drawer's client search picker already uses
	// (block.go newJobClientSearchHandler) — no new espyna surface.
	ListClients func(ctx context.Context, req *clientpb.ListClientsRequest) (*clientpb.ListClientsResponse, error)

	// Page-header delivery-group resolution (round 4 item 2): ListJobs finds
	// ONE job under the template to read its origin subscription id from;
	// ListSubscriptionGroupMembers/ListSubscriptionGroups feed the shared
	// domain/operation/deliverygroup resolver (the SAME chain job/list/
	// template_summary.go's delivery summary uses — extracted so neither
	// file duplicates it). All three are already-wired top-level closures
	// (u.Operation.Job.ListJobs, u.Subscription.SubscriptionGroupMember.*,
	// u.Subscription.SubscriptionGroup.*) — no new espyna surface. Nil-safe:
	// nil -> the header caption renders blank.
	ListJobs                     func(ctx context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error)
	ListSubscriptionGroupMembers func(ctx context.Context, req *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error)
	ListSubscriptionGroups       func(ctx context.Context, req *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error)

	// Options — app-configured row presentation (sort/description/group_by
	// through "client_attributes.<code>" references). Zero value → flat
	// roster, rendering unchanged.
	Options outcome_matrix.Options

	// Row-attribute hydration backing Options (attribute.code → attribute.id,
	// then the roster's client_attribute values). Both optional/nil-safe: nil
	// or a failed lookup disables the attribute-driven behaviors, never the
	// page (the grid falls back to the flat roster).
	ListClientAttributes     func(ctx context.Context, req *clientattributepb.ListClientAttributesRequest) (*clientattributepb.ListClientAttributesResponse, error)
	ResolveAttributeIDByCode func(ctx context.Context, code string) (string, error)

	// Header-breadcrumb back-link to the job list (the matrix's parent
	// surface): tier-resolved route pattern ("{status}" placeholder intact)
	// and active-status heading. Optional: empty values render the header's
	// title-only crumb, exactly as before.
	JobListURL   string
	JobListLabel string
}

// PageData holds the data for the outcome matrix page. The Grid field is named
// "Grid" so the render pipeline's reflection injector (parallel to TableConfig's
// "Table" branch) populates Grid.Nonce / Grid.WorkspaceID on render.
//
// WorkspaceID (promoted from the embedded types.PageData, injected by the
// ViewAdapter every render) is required by the approval bar's {{actionForm}}
// hidden-input signing.
type PageData struct {
	types.PageData
	Grid   *types.CellGridConfig
	Labels outcome_matrix.Labels

	SubjectName  string
	ScopeActive  string // "mine" | "all"
	ScopeMineURL string
	ScopeAllURL  string
	ShowScopeAll bool

	// ApprovalBar is the per-phase approval bar (one entry per template phase,
	// derived from the response roll-up over the full sheet S). Empty when the
	// template has no phase roll-up (mock build / no sheet).
	ApprovalBar     []ApprovalPhase
	ShowApprovalBar bool

	// Columns selector (plan 20260720 Q1/Q2: server-prune, ?hide= URL-canonical).
	// Cols is the FULL L1→L2→leaf tree with per-node hidden state + precomputed
	// toggle URLs — the menu is pure links, so no client JS ever owns colspan or
	// visibility state. The approval bar is intentionally NOT affected by hiding
	// (it is a per-phase status/action surface, not a data column).
	Cols            []ColsGroup
	ShowCols        bool
	ColsHiddenCount int    // effectively hidden leaf count (subtree + individual)
	ColsShowAllURL  string // current view minus ?hide=

	// ExportURL is the sheet-level CSV download carrying the SAME ?scope= +
	// ?hide= as the current view ("export what you see" — Q3). Empty hides the
	// button (no route wired or no template).
	ExportURL string

	// DownloadDrawerURL is the export-drawer GET (hx-get target of the toolbar
	// trigger), carrying the SAME ?scope= + ?hide= as ExportURL so the drawer
	// seeds its hidden inputs from the live view state. Set alongside ExportURL
	// under the same leaf-count gate.
	DownloadDrawerURL string
}

// ApprovalPhase is one phase's approval-bar entry: the derived chip state + the
// state-gated transition affordances rendered as signed HTMX POST forms.
type ApprovalPhase struct {
	PhaseID     string
	Label       string
	Status      string // enum name, e.g. PHASE_APPROVAL_STATUS_FOR_REVIEW
	StatusLabel string // lyngua status text
	ChipVariant string // "" (default) | "warning" | "info" | "success"
	Slug        string // testid-safe phase slug

	Mixed       bool
	NotStarted  bool
	HardFrozen  bool
	HasData     bool
	TargetCount int32
	TargetLabel string // "{count} students/…" caption
	BlankCount  int32

	Hint string // workflow-locked or hard-frozen hint text ("" when neither)

	// State-gated action affordances. Each *Path is the EXACT resolved POST path
	// (used identically for hx-post and {{actionForm}} so the action-workspace
	// guard's r.URL.Path signature check matches).
	CanSubmit   bool
	CanVerify   bool
	CanPublish  bool
	CanReturn   bool
	SubmitPath  string
	VerifyPath  string
	PublishPath string
	ReturnPath  string

	SubmitConfirm  string // hx-confirm text with {count} substituted (D6)
	VerifyConfirm  string
	PublishConfirm string
	ReturnConfirm  string

	// ReturnReasonRequired hints the UI (server is authoritative) that a return
	// will require a non-blank reason because a member is/was published.
	ReturnReasonRequired bool
	ReturnReasonLabel    string
}

// NewView creates the outcome matrix GET view.
func NewView(deps *PageViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("task_outcome", "read") {
			return view.Forbidden("task_outcome:read")
		}

		templateID := viewCtx.Request.PathValue("id")

		// Scope resolution. Explicit ?scope=all|mine is honored as requested. With
		// NO explicit scope, an operator holding workspace:list defaults to ALL: a
		// non-staff MINE view is zero rows by design, so an admin landing on the
		// page should see the roster. The server-side re-check below (effectiveAll)
		// still gates ALL on workspace:list, unchanged.
		canSeeAll := perms.Can(scopeEntity, scopeAction)
		scopeParam := viewCtx.Request.URL.Query().Get("scope")
		requestedAll := scopeParam == "all"
		if scopeParam == "" {
			requestedAll = canSeeAll
		}
		effectiveAll := requestedAll && canSeeAll

		l := deps.Labels

		var resp *matrixpb.GetOutcomeMatrixResponse
		if templateID != "" && deps.GetOutcomeMatrix != nil {
			scope := matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_MINE
			if effectiveAll {
				scope = matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL
			}

			var err error
			resp, err = deps.GetOutcomeMatrix(ctx, &matrixpb.GetOutcomeMatrixRequest{
				JobTemplateId: templateID,
				Scope:         scope,
			})
			if err != nil {
				log.Printf("Failed to load outcome matrix for template %s: %v", templateID, err)
				resp = nil
			}
		}

		// ?hide= — comma-separated L1 phase ids and/or leaf ColumnKeys. Resolved
		// against the response tree (unknown tokens dropped, all-hidden fails
		// safe to fully visible), then pruned server-side BEFORE render so every
		// colspan/coord derivation stays correct by construction.
		hidden := resolveHidden(viewCtx.Request.URL.Query().Get("hide"), resp)

		grid := buildGrid(ctx, deps, perms, resp, effectiveAll, templateID, hidden, viewCtx)

		subjectName := ""
		if resp != nil {
			subjectName = resp.GetJobTemplateName()
		}

		// Header title = the job_template name; header caption = its delivery
		// group (section) name (round 4 item 2 — replaces the in-body
		// "Subject: ..." line and the table's duplicate <caption>, both
		// removed below / in matrix.html). groupName resolves via the SAME
		// job -> origin subscription -> subscription_group_member ->
		// subscription_group chain job/list/template_summary.go uses,
		// through the shared domain/operation/deliverygroup package — one
		// sample job under the template is enough (every job of one
		// template shares the same section, verified
		// docs/plan/20260710-staff-class-list/s3b-view.md).
		headerTitle := l.Page.Title
		if subjectName != "" {
			headerTitle = subjectName
		}
		groupName := ""
		if originID := sampleOriginSubscription(ctx, deps, templateID); originID != "" {
			groupName, _ = deliverygroup.ResolveOne(ctx, deps.ListSubscriptionGroupMembers, deps.ListSubscriptionGroups, originID)
		}

		scopeActive := "mine"
		if effectiveAll {
			scopeActive = "all"
		}

		// Breadcrumb crumb ("Active Classes › <subject>" on education): the
		// job list's active-status page is the matrix's parent surface. Both
		// fields are optional — empty leaves the title-only crumb.
		breadcrumbURL := ""
		if deps.JobListURL != "" {
			breadcrumbURL = route.ResolveURL(deps.JobListURL, "status", "active")
		}

		// Every navigation URL on the page carries the SAME ?scope= + ?hide=
		// pair so the view state round-trips through scope toggles, selector
		// toggles, and the export link ("," and ":" are legal query characters;
		// html/template attribute-escapes on render).
		matrixBase := route.ResolveURL(deps.Routes.MatrixURL, "id", templateID)
		hideCSV := ""
		if resp != nil {
			hideCSV = hiddenCSV(hidden, resp.GetPhases())
		}
		withParams := func(base, scope, hide string) string {
			u := base + "?scope=" + scope
			if hide != "" {
				u += "&hide=" + hide
			}
			return u
		}

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:        viewCtx.CacheVersion,
				Title:               headerTitle,
				ContentTemplate:     "outcome-matrix-content",
				CurrentPath:         viewCtx.CurrentPath,
				ActiveNav:           deps.Routes.ActiveNav,
				ActiveSubNav:        deps.Routes.ActiveSubNav,
				HeaderBreadcrumb:    deps.JobListLabel,
				HeaderBreadcrumbURL: breadcrumbURL,
				HeaderTitle:         headerTitle,
				HeaderSubtitle:      groupName,
				HeaderIcon:          "icon-grid",
				CommonLabels:        deps.CommonLabels,
			},
			Grid:         grid,
			Labels:       l,
			SubjectName:  subjectName,
			ScopeActive:  scopeActive,
			ScopeMineURL: withParams(matrixBase, "mine", hideCSV),
			ScopeAllURL:  withParams(matrixBase, "all", hideCSV),
			ShowScopeAll: canSeeAll,
		}

		bar := buildApprovalBar(deps, perms, resp, templateID)
		pageData.ApprovalBar = bar
		pageData.ShowApprovalBar = len(bar) > 0

		if resp != nil {
			fullCols := buildColumns(resp.GetPhases())
			phases := resp.GetPhases()
			cols, hiddenLeaves := buildColsSelector(fullCols, hidden, func(h map[string]bool) string {
				return withParams(matrixBase, scopeActive, hiddenCSV(h, phases))
			})
			pageData.Cols = cols
			pageData.ShowCols = len(fullCols) > 0
			pageData.ColsHiddenCount = hiddenLeaves
			pageData.ColsShowAllURL = withParams(matrixBase, scopeActive, "")
			// Leaf-count gate: an unconfigured/empty matrix would render a
			// button whose GET can only 404 (the handler rejects zero-leaf
			// grids) — don't offer a download that cannot succeed.
			if deps.Routes.ExportURL != "" && grid.LeafColumnCount() > 0 {
				pageData.ExportURL = withParams(
					route.ResolveURL(deps.Routes.ExportURL, "id", templateID), scopeActive, hideCSV)
				// The drawer trigger is gated on ExportURL in matrix.html, so its
				// hx-get target is resolved under the SAME leaf-count guard, with
				// the SAME live ?scope=/?hide= carried through.
				if deps.Routes.DownloadDrawerURL != "" {
					pageData.DownloadDrawerURL = withParams(
						route.ResolveURL(deps.Routes.DownloadDrawerURL, "id", templateID), scopeActive, hideCSV)
				}
			}
		}

		return view.OK("outcome-matrix", pageData)
	})
}

// sampleOriginSubscription finds ONE job under templateID and returns its
// origin subscription id (ORIGIN_TYPE_SUBSCRIPTION jobs only). Limit 1 — any
// single job's origin is enough; every job under one job_template shares the
// same delivery group (section), so no pagination loop is needed here
// (contrast job/list/template_summary.go, which pages through ALL scoped
// jobs because it also needs the per-template item COUNT — this call only
// needs one sample). Nil-safe: deps.ListJobs == nil -> "".
func sampleOriginSubscription(ctx context.Context, deps *PageViewDeps, templateID string) string {
	if deps.ListJobs == nil || templateID == "" {
		return ""
	}
	resp, err := deps.ListJobs(ctx, &jobpb.ListJobsRequest{
		Filters: &commonpb.FilterRequest{
			Filters: []*commonpb.TypedFilter{{
				Field: "job_template_id",
				FilterType: &commonpb.TypedFilter_StringFilter{
					StringFilter: &commonpb.StringFilter{
						Value:    templateID,
						Operator: commonpb.StringOperator_STRING_EQUALS,
					},
				},
			}},
		},
		Pagination: &commonpb.PaginationRequest{
			Limit:  1,
			Method: &commonpb.PaginationRequest_Offset{Offset: &commonpb.OffsetPagination{Page: 1}},
		},
	})
	if err != nil {
		log.Printf("Failed to sample a job for template %s: %v", templateID, err)
		return ""
	}
	for _, j := range resp.GetData() {
		if j.GetOriginType() == enums.OriginType_ORIGIN_TYPE_SUBSCRIPTION {
			if oid := j.GetOriginId(); oid != "" {
				return oid
			}
		}
	}
	return ""
}

// buildGrid converts the proto response into a *types.CellGridConfig. The acting
// staff read-only gating is applied here (view layer, per view-scope.md §4).
func buildGrid(
	ctx context.Context,
	deps *PageViewDeps,
	perms *types.UserPermissions,
	resp *matrixpb.GetOutcomeMatrixResponse,
	effectiveAll bool,
	templateID string,
	hidden map[string]bool,
	viewCtx *view.ViewContext,
) *types.CellGridConfig {
	l := deps.Labels

	scopeStr := "mine"
	if effectiveAll {
		scopeStr = "all"
	}

	// AutoSave (W2 grade-sheet edit mode) turns on the per-cell micro-batch
	// client. Enable it whenever the grid is save-enabled — the record action
	// accepts an update-only OR create-only save, so either grant qualifies (the
	// inverse of the SaveDisabled create-gate, widened to include update). A
	// read-only viewer (neither grant) gets the plain grid with no editors.
	canSave := perms.Can("task_outcome", "create") || perms.Can("task_outcome", "update")
	saveMode := "batch"
	if canSave {
		saveMode = "cell"
	}

	cfg := &types.CellGridConfig{
		ID: "outcome-matrix-grid",
		// Caption intentionally left unset (round 4 item 2): the page
		// header (HeaderTitle/HeaderSubtitle, set in NewView) now carries
		// the template name + delivery group — the table's own <caption>
		// used to duplicate the header with the same static l.Page.Title
		// text. cell-grid-card.html's {{if .Caption}} guard already no-ops
		// on empty, so this is a clean removal, not a hidden feature loss.
		FreezeFirstCol:   true,
		FreezeHeaderRows: 3,
		SaveURL:          route.ResolveURL(deps.Routes.RecordURL, "id", templateID),
		SaveMode:         saveMode,
		AutoSave:         canSave,
		JobTemplateID:    templateID,
		Scope:            scopeStr,
		SaveDisabled:     !perms.Can("task_outcome", "create"),
		Labels:           l.Grid.CellGridLabels,
		CacheVersion:     viewCtx.CacheVersion,
	}

	if resp == nil {
		return cfg
	}

	// Acting staff for the read-only gate.
	var actingStaff string
	if deps.ResolveStaff != nil {
		actingStaff, _ = deps.ResolveStaff(viewCtx.Request.Context())
	}

	// Roster name hydration: GetOutcomeMatrix's ClientLabel is deliberately
	// the opaque client_id (espyna outcome_matrix_query.go:364 "the view
	// resolves a display name") — resolve it here, the same User-hydrating
	// page-data pattern used for the job list's deliverer column
	// (job/list/template_summary.go fetchAllStaff), except Client carries its
	// own `name` column directly (no join needed — see clientDisplayName).
	rosterIDs := distinctClientIDs(resp.GetRows())
	clientNames := fetchClientNames(ctx, deps, rosterIDs)

	// Row-attribute hydration for the configured Options (one fetch per
	// distinct referenced code; nothing configured → no fetch, nil map).
	attrValues := fetchClientAttributeValues(ctx, deps, rosterIDs)

	// Prune BEFORE assignAutoSaveCoords: the head/band colspans and the W2
	// keyboard coords all derive from cfg.Columns, so removing subtrees here
	// keeps every downstream computation correct by construction. `hidden` has
	// already been resolved fail-safe (resolveHidden) — never empties the grid.
	cfg.Columns = pruneColumns(buildColumns(resp.GetPhases()), hidden)
	cfg.Rows = buildRows(resp.GetRows(), actingStaff, l.Grid.ReadOnlyTooltip, clientNames, phaseEditableFunc(resp))
	applyRowOptions(cfg, deps.Options, attrValues, clientNames)
	if cfg.AutoSave {
		// Populate the W2 edit-mode per-cell fields AFTER applyRowOptions has
		// settled the final display order (RowIndex counts data rows in that
		// order; the JS keyboard grid-nav reads data-row-index/data-col-index).
		assignAutoSaveCoords(cfg)
	}
	return cfg
}

// assignAutoSaveCoords populates the W2 AutoSave per-cell fields the edit-mode
// client needs: RowIndex/ColIndex logical coordinates (keyboard grid-nav),
// SavedValue (the server baseline the client dirty-checks + reverts to), and a
// stable InputID/StatusID pair (focus target + aria-live status region). Only
// called in AutoSave mode so the legacy grid's DOM is byte-unchanged.
func assignAutoSaveCoords(cfg *types.CellGridConfig) {
	// Leaf-column index, left-to-right across the whole phase→task→criterion tree.
	colIndex := make(map[string]int)
	ci := 0
	for _, l1 := range cfg.Columns {
		for _, l2 := range l1.Level2 {
			for _, l3 := range l2.Level3 {
				colIndex[l3.ColumnKey] = ci
				ci++
			}
		}
	}
	for ri := range cfg.Rows {
		row := &cfg.Rows[ri]
		rowSlug := slug(row.ID)
		for colKey, cell := range row.Cells {
			c := cell
			c.RowIndex = ri
			if idx, ok := colIndex[colKey]; ok {
				c.ColIndex = idx
			}
			// The server baseline the client compares live input against.
			c.SavedValue = c.Value
			inputID := "om-in-" + rowSlug + "-" + slug(colKey)
			c.InputID = inputID
			c.StatusID = inputID + "-st"
			row.Cells[colKey] = c
		}
	}
}

// fetchClientAttributeValues resolves each Options-referenced attribute code
// to its attribute id, then loads the roster clients' values for it —
// code → (client_id → value). Chunked like fetchClientNames. Nil-safe on
// every dependency; a failed code lookup logs and skips that code.
func fetchClientAttributeValues(ctx context.Context, deps *PageViewDeps, clientIDs []string) map[string]map[string]string {
	codes := deps.Options.AttributeCodes()
	if len(codes) == 0 || deps.ListClientAttributes == nil || deps.ResolveAttributeIDByCode == nil || len(clientIDs) == 0 {
		return nil
	}
	out := make(map[string]map[string]string, len(codes))
	for _, code := range codes {
		attrID, err := deps.ResolveAttributeIDByCode(ctx, code)
		if err != nil || attrID == "" {
			log.Printf("outcome matrix: attribute code %q did not resolve (options ignored for it): %v", code, err)
			continue
		}
		vals := map[string]string{}
		for start := 0; start < len(clientIDs); start += clientNamePageLimit {
			end := start + clientNamePageLimit
			if end > len(clientIDs) {
				end = len(clientIDs)
			}
			resp, err := deps.ListClientAttributes(ctx, &clientattributepb.ListClientAttributesRequest{
				Filters: &commonpb.FilterRequest{
					Filters: []*commonpb.TypedFilter{
						{
							Field: "attribute_id",
							FilterType: &commonpb.TypedFilter_StringFilter{
								StringFilter: &commonpb.StringFilter{Value: attrID, Operator: commonpb.StringOperator_STRING_EQUALS},
							},
						},
						{
							Field: "client_id",
							FilterType: &commonpb.TypedFilter_ListFilter{
								ListFilter: &commonpb.ListFilter{Values: clientIDs[start:end], Operator: commonpb.ListOperator_LIST_IN},
							},
						},
					},
				},
			})
			if err != nil {
				log.Printf("outcome matrix: list client attributes for code %q: %v", code, err)
				continue
			}
			for _, ca := range resp.GetData() {
				if cid, v := ca.GetClientId(), strings.TrimSpace(ca.GetValue()); cid != "" && v != "" {
					vals[cid] = v
				}
			}
		}
		out[code] = vals
	}
	return out
}

// applyRowOptions applies the configured sort / description / group_by row
// presentation to the built grid rows. Row identity and cells are untouched —
// only order, Label (banded class-list form), Description, and GroupLabel
// markers change. Sorting is stable so the adapter's roster order is
// preserved among equals. With zero-valued options nothing here runs — the
// flat grid renders exactly as built (the backward-compat contract).
func applyRowOptions(cfg *types.CellGridConfig, opts outcome_matrix.Options, attrValues map[string]map[string]string, clientNames map[string]clientName) {
	valueFor := func(field, clientID string) (string, bool) {
		code, ok := outcome_matrix.ClientAttributeCode(field)
		if !ok {
			return "", false
		}
		vals, ok := attrValues[code]
		if !ok {
			return "", false
		}
		return vals[clientID], true
	}

	_, banded := valueFor(opts.RowGroupByField, "")

	// Banded presentation implies the class-list name form "{last}, {first}"
	// (prod report-card parity; the outcome_summary section grid does the
	// same). Relabel BEFORE sorting so label tie-breaks order by last name.
	if banded {
		for i := range cfg.Rows {
			if name := clientNames[cfg.Rows[i].ID].listName(); name != "" {
				cfg.Rows[i].Label = name
			}
		}
	}

	if _, ok := valueFor(opts.RowDescriptionField, ""); ok {
		for i := range cfg.Rows {
			desc, _ := valueFor(opts.RowDescriptionField, cfg.Rows[i].ID)
			cfg.Rows[i].Description = desc
		}
	}

	byAttr := func(field string) func(a, b types.CellGridRow) int {
		return func(a, b types.CellGridRow) int {
			av, _ := valueFor(field, a.ID)
			bv, _ := valueFor(field, b.ID)
			// rows without a value sort last
			if av == "" && bv != "" {
				return 1
			}
			if av != "" && bv == "" {
				return -1
			}
			if c := strings.Compare(strings.ToLower(av), strings.ToLower(bv)); c != 0 {
				return c
			}
			return strings.Compare(strings.ToLower(a.Label), strings.ToLower(b.Label))
		}
	}

	if _, ok := valueFor(opts.RowSortField, ""); ok {
		cmp := byAttr(opts.RowSortField)
		if opts.RowDirection() == "desc" {
			asc := cmp
			cmp = func(a, b types.CellGridRow) int { return -asc(a, b) }
		}
		slices.SortStableFunc(cfg.Rows, cmp)
	}

	if banded {
		// Partition into bands: stable-sort by the group value — configured
		// RowGroupValueOrder first (listed values lead, in list order), then
		// value-asc, no-value band last — and mark each band's first row with
		// its GroupLabel.
		slices.SortStableFunc(cfg.Rows, func(a, b types.CellGridRow) int {
			av, _ := valueFor(opts.RowGroupByField, a.ID)
			bv, _ := valueFor(opts.RowGroupByField, b.ID)
			if av == "" && bv != "" {
				return 1
			}
			if av != "" && bv == "" {
				return -1
			}
			ra, aListed := opts.RowGroupValueRank(av)
			rb, bListed := opts.RowGroupValueRank(bv)
			if aListed != bListed {
				if aListed {
					return -1
				}
				return 1
			}
			if aListed && bListed && ra != rb {
				return ra - rb
			}
			return strings.Compare(strings.ToLower(av), strings.ToLower(bv))
		})
		prev := ""
		started := false
		for i := range cfg.Rows {
			v, _ := valueFor(opts.RowGroupByField, cfg.Rows[i].ID)
			if !started || v != prev {
				label := v
				if label == "" {
					label = "—"
				}
				cfg.Rows[i].GroupLabel = label
				prev, started = v, true
			}
		}

		// Sequence numbers, CONTINUOUS across bands (prod's class-list
		// numbering: male 1..N, female N+1..M), applied after banding so
		// they reflect the final render order.
		for i := range cfg.Rows {
			cfg.Rows[i].Label = strconv.Itoa(i+1) + " " + cfg.Rows[i].Label
		}
	}
}

// distinctClientIDs collects the roster's DISTINCT client ids in row order.
func distinctClientIDs(rows []*matrixpb.OutcomeRow) []string {
	ids := make([]string, 0, len(rows))
	seen := make(map[string]bool, len(rows))
	for _, r := range rows {
		if cid := r.GetClientId(); cid != "" && !seen[cid] {
			seen[cid] = true
			ids = append(ids, cid)
		}
	}
	return ids
}

// fetchClientNames resolves client_id -> display name via ListClients,
// chunked into batches of clientNamePageLimit ids per ListFilter(IN) call
// (bounded output regardless of adapter Pagination support, mirroring
// job/list/template_summary.go's fetchSubscriptionGroupIDs). Optional/
// nil-safe: deps.ListClients == nil -> empty map -> every row falls back to
// short(clientID) in rowLabel, i.e. today's (bugged) behavior.
func fetchClientNames(ctx context.Context, deps *PageViewDeps, clientIDs []string) map[string]clientName {
	out := map[string]clientName{}
	if deps.ListClients == nil || len(clientIDs) == 0 {
		return out
	}
	for start := 0; start < len(clientIDs); start += clientNamePageLimit {
		end := start + clientNamePageLimit
		if end > len(clientIDs) {
			end = len(clientIDs)
		}
		chunk := clientIDs[start:end]
		resp, err := deps.ListClients(ctx, &clientpb.ListClientsRequest{
			Filters: &commonpb.FilterRequest{
				Filters: []*commonpb.TypedFilter{{
					Field: "id",
					FilterType: &commonpb.TypedFilter_ListFilter{
						ListFilter: &commonpb.ListFilter{Values: chunk, Operator: commonpb.ListOperator_LIST_IN},
					},
				}},
			},
		})
		if err != nil {
			log.Printf("Failed to list clients for outcome matrix roster: %v", err)
			continue
		}
		for _, c := range resp.GetData() {
			id := c.GetId()
			if id == "" {
				continue
			}
			last := strings.TrimSpace(c.GetLastName())
			first := strings.TrimSpace(c.GetFirstName())
			if u := c.GetUser(); u != nil {
				if last == "" {
					last = strings.TrimSpace(u.GetLastName())
				}
				if first == "" {
					first = strings.TrimSpace(u.GetFirstName())
				}
			}
			out[id] = clientName{display: clientDisplayName(c), last: last, first: first}
		}
	}
	return out
}

// clientName carries the roster-name parts: display is the flat fallback;
// last/first feed the banded "Last, First" class-list form (either may be
// empty — e.g. organization clients — in which case display is used).
type clientName struct {
	display, last, first string
}

// listName renders "{last}, {first}" when both parts exist, else display.
func (n clientName) listName() string {
	if n.last != "" && n.first != "" {
		return n.last + ", " + n.first
	}
	return n.display
}

// clientDisplayName mirrors the established fayna client-name pattern
// (block.go newJobClientSearchHandler): the client's own `name` column
// first (populated directly on `client`, no join — this is what makes
// Client different from Staff, which has no name column of its own and
// needs the User-hydrating GetStaffListPageData read instead), falling back
// to the embedded User's first+last name, then the raw id.
func clientDisplayName(c *clientpb.Client) string {
	if name := strings.TrimSpace(c.GetName()); name != "" {
		return name
	}
	if u := c.GetUser(); u != nil {
		if name := strings.TrimSpace(u.GetFirstName() + " " + u.GetLastName()); name != "" {
			return name
		}
	}
	return c.GetId()
}

// buildApprovalBar maps the response roll-up (derived over the full sheet S) into
// the per-phase approval-bar entries. Each entry carries the derived chip state +
// the state-gated transition affordances (permission-gated view mirror; the
// espyna use-case ActionGatekeeper + strict authorizer are authoritative). The
// action buttons render as real signed HTMX POST forms in the template.
func buildApprovalBar(deps *PageViewDeps, perms *types.UserPermissions, resp *matrixpb.GetOutcomeMatrixResponse, templateID string) []ApprovalPhase {
	if resp == nil || templateID == "" {
		return nil
	}
	rollups := resp.GetApprovalRollups()
	if len(rollups) == 0 {
		return nil
	}
	l := deps.Labels.Approval

	phaseLabel := map[string]string{}
	for _, ph := range resp.GetPhases() {
		phaseLabel[ph.GetJobTemplatePhaseId()] = ph.GetLabel()
	}

	// Layer-2 view gates cite the SAME job_phase:<verb> codes the use-case
	// ActionGatekeeper.Check + strict authorizer verify (copya.md gate discipline).
	canSubmit := perms.Can("job_phase", "submit")
	canVerify := perms.Can("job_phase", "verify")
	canPublish := perms.Can("job_phase", "publish")
	canReturn := perms.Can("job_phase", "return")

	submitPath := route.ResolveURL(deps.Routes.SubmitURL, "id", templateID)
	verifyPath := route.ResolveURL(deps.Routes.VerifyURL, "id", templateID)
	publishPath := route.ResolveURL(deps.Routes.PublishURL, "id", templateID)
	returnPath := route.ResolveURL(deps.Routes.ReturnURL, "id", templateID)

	out := make([]ApprovalPhase, 0, len(rollups))
	for _, ru := range rollups {
		status := ru.GetStatus()
		inProgress := status == jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_IN_PROGRESS
		forReview := status == jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW
		verified := status == jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_VERIFIED
		published := status == jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED
		frozen := ru.GetHardFrozen()
		mixed := ru.GetMixed()

		ap := ApprovalPhase{
			PhaseID:        ru.GetJobTemplatePhaseId(),
			Label:          phaseLabel[ru.GetJobTemplatePhaseId()],
			Status:         status.String(),
			StatusLabel:    approvalStatusLabel(l, status),
			ChipVariant:    approvalChipVariant(status),
			Slug:           slug(ru.GetJobTemplatePhaseId()),
			Mixed:          mixed,
			NotStarted:     inProgress && !ru.GetHasData(),
			HardFrozen:     frozen,
			HasData:        ru.GetHasData(),
			TargetCount:    ru.GetTargetCount(),
			TargetLabel:    subCount(l.Chip.PublishedCount, ru.GetTargetCount()),
			BlankCount:     ru.GetBlankRequiredCount(),
			SubmitPath:     submitPath,
			VerifyPath:     verifyPath,
			PublishPath:    publishPath,
			ReturnPath:     returnPath,
			SubmitConfirm:  subCount(l.Confirm.Submit, ru.GetBlankRequiredCount()),
			VerifyConfirm:  l.Confirm.Verify,
			PublishConfirm: l.Confirm.Publish,
			ReturnConfirm:  l.Confirm.Return,
		}

		// One hint line: hard-frozen dominates; else any workflow lock.
		switch {
		case frozen:
			ap.Hint = l.HardFrozenHint
		case !inProgress || mixed:
			ap.Hint = l.LockedHint
		}

		// State-gated action affordances (a hard-frozen sheet exposes none).
		ap.CanSubmit = canSubmit && inProgress && !mixed && !frozen
		ap.CanVerify = canVerify && forReview && !mixed && !frozen
		ap.CanPublish = canPublish && verified && !mixed && !frozen
		ap.CanReturn = canReturn && !frozen && (mixed || !inProgress)

		// A published sheet's return needs a reason (server enforces; UI marks the
		// input required as a best-effort hint — a mixed/was-published case the
		// roll-up cannot see is still enforced server-side, surfacing as a 422).
		ap.ReturnReasonRequired = published
		if ap.ReturnReasonRequired {
			ap.ReturnReasonLabel = l.Actions.ReturnReasonRequired
		} else {
			ap.ReturnReasonLabel = l.Actions.ReturnReason
		}

		out = append(out, ap)
	}
	return out
}

// approvalStatusLabel maps the ladder enum to its lyngua status text.
func approvalStatusLabel(l outcome_matrix.ApprovalLabels, s jobphasepb.PhaseApprovalStatus) string {
	switch s {
	case jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW:
		return l.Status.ForReview
	case jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_VERIFIED:
		return l.Status.Verified
	case jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED:
		return l.Status.Published
	default:
		return l.Status.InProgress
	}
}

// approvalChipVariant is the Go badge-variant switch (NOT lyngua) → the pyeza
// badge modifier suffix: in_progress → neutral, for_review → warning, verified →
// info, published → success. The template renders `badge badge-{{.ChipVariant}}`.
func approvalChipVariant(s jobphasepb.PhaseApprovalStatus) string {
	switch s {
	case jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW:
		return "warning"
	case jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_VERIFIED:
		return "info"
	case jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED:
		return "success"
	default:
		return "neutral"
	}
}

// subCount substitutes the "{count}" placeholder with n (once).
func subCount(tmpl string, n int32) string {
	return strings.Replace(tmpl, "{count}", strconv.FormatInt(int64(n), 10), 1)
}

// buildColumns maps the proto phase→task→criterion tree into CellGridLevel1/2/3.
func buildColumns(phases []*matrixpb.PhaseColumn) []types.CellGridLevel1 {
	columns := make([]types.CellGridLevel1, 0, len(phases))
	for _, ph := range phases {
		l1 := types.CellGridLevel1{
			Key:   ph.GetJobTemplatePhaseId(),
			Label: ph.GetLabel(),
		}
		for _, tk := range ph.GetTasks() {
			l2 := types.CellGridLevel2{
				Key:   tk.GetJobTemplateTaskId(),
				Label: tk.GetLabel(),
			}
			for _, cr := range tk.GetCriteria() {
				l2.Level3 = append(l2.Level3, types.CellGridLevel3{
					ColumnKey: cr.GetColumnKey(),
					Label:     criterionLabel(cr),
					CellInput: buildCellInput(cr.GetCriteria()),
				})
			}
			l1.Level2 = append(l1.Level2, l2)
		}
		columns = append(columns, l1)
	}
	return columns
}

// phaseEditableFunc returns a colKey → bool predicate mirroring the server
// cell_editable gate at PHASE grain (plan §4.4: cell_editable = IN_PROGRESS &&
// !hard_frozen && ownership). The ownership half is the espyna OutcomeCell.
// Editable flag (already recorder/assignee-scoped); THIS adds the phase-status
// half from the roll-up over the full sheet S: a cell is render-editable only
// when its phase's roll-up is cleanly IN_PROGRESS (not mixed) and not
// hard-frozen. The server GuardCellWrite stays authoritative per job_phase row;
// this render mirror never widens editability, only narrows it (a mixed sheet
// locks in the UI and awaits a return). Phases with no roll-up entry default to
// permissive (the server still guards) so an unrelated read path never over-locks.
func phaseEditableFunc(resp *matrixpb.GetOutcomeMatrixResponse) func(colKey string) bool {
	taskToPhase := map[string]string{}
	for _, ph := range resp.GetPhases() {
		for _, tk := range ph.GetTasks() {
			taskToPhase[tk.GetJobTemplateTaskId()] = ph.GetJobTemplatePhaseId()
		}
	}
	phaseEditable := map[string]bool{}
	for _, ru := range resp.GetApprovalRollups() {
		phaseEditable[ru.GetJobTemplatePhaseId()] =
			ru.GetStatus() == jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_IN_PROGRESS &&
				!ru.GetMixed() && !ru.GetHardFrozen()
	}
	return func(colKey string) bool {
		taskID := colKey
		if i := strings.Index(colKey, ":"); i >= 0 {
			taskID = colKey[:i]
		}
		ph, ok := taskToPhase[taskID]
		if !ok {
			return true
		}
		ed, ok := phaseEditable[ph]
		if !ok {
			return true
		}
		return ed
	}
}

// buildRows maps the proto rows into CellGridRows with the read-only gate applied.
// allowEdit is the phase-status render mirror (phaseEditableFunc): a locked or
// hard-frozen phase forces every cell read-only regardless of ownership.
func buildRows(rows []*matrixpb.OutcomeRow, actingStaff, readOnlyTooltip string, clientNames map[string]clientName, allowEdit func(colKey string) bool) []types.CellGridRow {
	out := make([]types.CellGridRow, 0, len(rows))
	for _, r := range rows {
		clientID := r.GetClientId()
		gr := types.CellGridRow{
			ID:     clientID,
			Label:  rowLabel(r, clientNames),
			Cells:  make(map[string]types.CellGridCell, len(r.GetCells())),
			TestID: "om-row-" + short(clientID),
		}
		for colKey, cell := range r.GetCells() {
			// Read-only when recorded by a different staff member. Honor the
			// backend editable flag too (job not spawned → not editable).
			readOnly := cell.GetRecordedBy() != "" && cell.GetRecordedBy() != actingStaff
			editable := cell.GetEditable()
			// Phase-status render mirror: a workflow-locked / hard-frozen phase
			// forces the cell read-only (server GuardCellWrite is authoritative).
			if allowEdit != nil && !allowEdit(colKey) {
				editable = false
				readOnly = true
			}
			gr.Cells[colKey] = types.CellGridCell{
				OutcomeID:       cell.GetOutcomeId(),
				JobTaskID:       cell.GetJobTaskId(),
				CriteriaID:      criteriaIDFromColumnKey(colKey),
				Value:           cellValue(cell),
				TextValue:       cellSecondaryText(cell),
				Editable:        editable,
				ReadOnly:        readOnly,
				ReadOnlyTooltip: readOnlyTooltip,
				TestID:          cellTestID(clientID, colKey, readOnly),
			}
		}
		out = append(out, gr)
	}
	return out
}

// buildCellInput derives a CellInputDescriptor from the criterion's enforcement
// contract (the embedded outcome_criteria entity).
func buildCellInput(oc *outcomecriteriapb.OutcomeCriteria) types.CellInputDescriptor {
	d := types.CellInputDescriptor{Type: "text"}
	if oc == nil {
		return d
	}
	d.Type = criteriaTypeString(oc.GetCriteriaType())

	if oc.MinScore != nil {
		v := float64(oc.GetMinScore())
		d.Min = &v
	}
	if oc.MaxScore != nil {
		v := float64(oc.GetMaxScore())
		d.Max = &v
	}
	if oc.ScoreIncrement != nil {
		v := oc.GetScoreIncrement()
		d.Step = &v
	}
	d.Decimals = int(oc.GetDecimalPlaces())
	d.Unit = oc.GetUnit()

	d.PassLabel = oc.GetPassLabel()
	d.FailLabel = oc.GetFailLabel()

	for _, opt := range oc.GetAllowedDeterminations() {
		d.Options = append(d.Options, types.SelectOption{Value: opt, Label: opt})
	}

	if oc.MaxTextLength != nil {
		v := int(oc.GetMaxTextLength())
		d.MaxLength = &v
	}
	d.Prompt = oc.GetTextPrompt()
	return d
}

// criteriaTypeString maps the criteria_type enum onto the pyeza descriptor's
// string vocabulary. MULTI_CHECK maps through as "multi_check"; the component
// renders "—" for it (OQ-6: multi_check deferred to V2).
func criteriaTypeString(t enums.CriteriaType) string {
	switch t {
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_RANGE, enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE:
		return "numeric"
	case enums.CriteriaType_CRITERIA_TYPE_PASS_FAIL:
		return "pass_fail"
	case enums.CriteriaType_CRITERIA_TYPE_CATEGORICAL:
		return "categorical"
	case enums.CriteriaType_CRITERIA_TYPE_TEXT:
		return "text"
	case enums.CriteriaType_CRITERIA_TYPE_MULTI_CHECK:
		return "multi_check"
	default:
		return "text"
	}
}

func criterionLabel(cr *matrixpb.CriterionColumn) string {
	if oc := cr.GetCriteria(); oc != nil && oc.GetName() != "" {
		return oc.GetName()
	}
	return cr.GetColumnKey()
}

// rowLabel prefers the hydrated display name (fetchClientNames). GetOutcomeMatrix's
// ClientLabel is deliberately the opaque client_id when no hydration ran
// (espyna outcome_matrix_query.go:364) — it is NEVER a real name, so it is
// no longer consulted here (the pre-fix bug: this function trusted it and
// always got the id back). Falls back to a truncated id when the roster
// fetch found no match (deps.ListClients nil, or the client row missing).
func rowLabel(r *matrixpb.OutcomeRow, clientNames map[string]clientName) string {
	if name := clientNames[r.GetClientId()].display; name != "" {
		return name
	}
	return short(r.GetClientId())
}

// cellValue serialises the cell's typed value into a display string.
func cellValue(c *matrixpb.OutcomeCell) string {
	switch {
	case c.NumericValue != nil:
		return strconv.FormatFloat(c.GetNumericValue(), 'f', -1, 64)
	case c.TextValue != nil:
		return c.GetTextValue()
	case c.CategoricalValue != nil:
		return c.GetCategoricalValue()
	case c.PassFailValue != nil:
		if c.GetPassFailValue() {
			return "true"
		}
		return "false"
	}
	return ""
}

// cellSecondaryText returns the recorded text_value ONLY when it did not
// already win cellValue()'s priority order (s7pre gap 5: "text ratings/
// text_value invisible" — every criterion seeded in education1 is
// NUMERIC_SCORE, so cellValue() always picks NumericValue and the
// coexisting descriptor text was silently dropped, even though the espyna
// adapter already forwards it on OutcomeCell.TextValue). For a genuinely
// text-typed criterion, TextValue IS Value already — returning "" here
// avoids rendering the same string twice. Read-only display only; write
// semantics (record.go) are unaffected — this is a pure display concern.
func cellSecondaryText(c *matrixpb.OutcomeCell) string {
	if c.NumericValue == nil || c.TextValue == nil {
		return ""
	}
	return c.GetTextValue()
}

// criteriaIDFromColumnKey extracts the outcome_criteria_id from a
// "{job_template_task_id}:{outcome_criteria_id}" column key.
func criteriaIDFromColumnKey(k string) string {
	if i := strings.LastIndex(k, ":"); i >= 0 {
		return k[i+1:]
	}
	return k
}

func cellTestID(clientID, colKey string, readOnly bool) string {
	base := "om-cell-"
	if readOnly {
		base = "om-cell-ro-"
	}
	return base + short(clientID) + "-" + slug(colKey)
}

// short truncates an opaque id for a testid suffix.
func short(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// slug lowercases and keeps [a-z0-9]; space/'-'/':' collapse to '-'. Mirrors
// grade_sheet.go's slug() helper (extended with ':' for the composite column key).
func slug(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == ':':
			b.WriteByte('-')
		}
	}
	return b.String()
}
