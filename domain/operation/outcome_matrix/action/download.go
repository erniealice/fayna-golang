package action

// download.go — the export DRAWER GET view (20260720 Q3). It renders the
// download form (Period × Format selects + hidden scope/hide carrying the live
// view state) into #sheetContent. The form itself is a NATIVE GET form posting
// to ExportURL (see templates/download-drawer.html) — this view only builds its
// options and seeds its current state; the actual file download is served by
// list/export.go.
//
// TWO TRIGGERS, ONE DRAWER (20260726). The toolbar button opens it with no
// ?period=, so it lands on "All periods". Each L1 column header ALSO opens it
// (list/ratings.go), passing that column's period token — a phase CODE, or the
// reserved "final" — which is PRE-SELECTED here. The header triggers used to
// link straight at the CSV, which was the only download path left once
// .om-toolbar went display:none and so made the PDF export unreachable.
//
// The token is matched, never echoed: an unknown ?period= simply selects
// nothing extra and the browser falls back to the first option ("All"), the
// same fail-safe shape as export.go's periodKnown 400 guard.
//
// Route: GET DownloadDrawerURL (/action/outcome-matrix/{id}/download). It lives
// under /action/* for slug tidiness but is a GET (safe method): the CSRF hook +
// action-workspace signature guard constrain non-safe methods only, so no signed
// form is needed (routes.go documents this). Same task_outcome:read gate as the
// grid view and the CSV handler — fail-closed first statement.

import (
	"context"
	"net/http"
	"sort"
	"strings"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"

	"github.com/erniealice/pyeza-golang/route"
	pyezatypes "github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// Scope-toggle capability — same code the grid view + CSV handler gate the
// widened "all clients" scope on (list/page.go scopeEntity/scopeAction).
const (
	drawerScopeEntity = "workspace"
	drawerScopeAction = "list"
)

// DrawerDeps holds the dependencies for the download-drawer GET view. It reuses
// the SAME GetOutcomeMatrix closure the grid + CSV handler use, so the drawer's
// period options are derived from the exact response the export will prune.
type DrawerDeps struct {
	Routes outcome_matrix.Routes
	Labels outcome_matrix.Labels

	GetOutcomeMatrix func(ctx context.Context, req *matrixpb.GetOutcomeMatrixRequest) (*matrixpb.GetOutcomeMatrixResponse, error)

	// ListJobTemplateSummaries backs the (template, section) pair guard on the
	// Group* routes. nil ⇒ those routes 404 (fail closed).
	ListJobTemplateSummaries outcome_matrix.SummaryLister
}

// DrawerData is the template-facing shape for outcome-matrix-download-drawer-form.
// Nonce / CommonLabels / WorkspaceID are injected by the ViewAdapter via
// reflection (the CSP nonce backs the inline close/format-lock script).
type DrawerData struct {
	ExportAction  string                    // native GET form action (Group/ExportURL, resolved)
	Scope         string                    // live "mine"/"all" — hidden input
	Hide          string                    // live ?hide= tokens — hidden input
	PeriodOptions []pyezatypes.SelectOption // All + per-phase (code!="") + Final; one may be Selected
	FormatOptions []pyezatypes.SelectOption // csv / pdf
	Labels        outcome_matrix.Labels
	CommonLabels  any    // injected by ViewAdapter
	Nonce         string // injected by ViewAdapter (CSP nonce)
	WorkspaceID   string // injected by ViewAdapter
}

// NewDownloadDrawer creates the export-drawer GET view.
func NewDownloadDrawer(deps *DrawerDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		// Same Layer-3 gate as the grid view (NewView) + CSV handler — fail-closed.
		if !perms.Can("task_outcome", "read") {
			return view.Forbidden("task_outcome:read")
		}

		templateID := viewCtx.Request.PathValue("id")

		// Section narrowing ({group_id}) — validated before any read, same guard
		// as the grid view. Empty on the template-scoped route.
		section, ok := outcome_matrix.ResolveGroupScope(ctx, viewCtx.Request, templateID, deps.ListJobTemplateSummaries)
		if !ok {
			return view.ViewResult{Error: outcome_matrix.ErrGroupNotInTemplate, StatusCode: http.StatusNotFound}
		}

		// Scope resolution — byte-identical to list/export.go (widened admin
		// default + server-side workspace:list re-check).
		canSeeAll := perms.Can(drawerScopeEntity, drawerScopeAction)
		scopeParam := viewCtx.Request.URL.Query().Get("scope")
		requestedAll := scopeParam == "all"
		if scopeParam == "" {
			requestedAll = canSeeAll
		}
		effectiveAll := requestedAll && canSeeAll
		scopeStr := "mine"
		if effectiveAll {
			scopeStr = "all"
		}

		var resp *matrixpb.GetOutcomeMatrixResponse
		if templateID != "" && deps.GetOutcomeMatrix != nil {
			scope := matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_MINE
			if effectiveAll {
				scope = matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL
			}
			req := &matrixpb.GetOutcomeMatrixRequest{
				JobTemplateId: templateID,
				Scope:         scope,
			}
			if section.Scoped() {
				req.SubscriptionGroupId = &section.GroupID
			}
			resp, _ = deps.GetOutcomeMatrix(ctx, req)
		}

		// A section-scoped drawer must submit to the section-scoped export, or
		// the file would silently widen back to every student under the template
		// (87 vs the 29 on screen for the AY2026-27 grade-scoped generation).
		// The group route already reached this view; only its form action was
		// still template-scoped.
		exportAction := route.ResolveURL(deps.Routes.ExportURL, "id", templateID)
		if section.Scoped() && deps.Routes.GroupExportURL != "" {
			exportAction = route.ResolveURL(deps.Routes.GroupExportURL,
				"id", templateID, "group_id", section.GroupID)
		}

		data := &DrawerData{
			ExportAction:  exportAction,
			Scope:         scopeStr,
			Hide:          strings.TrimSpace(viewCtx.Request.URL.Query().Get("hide")),
			PeriodOptions: buildPeriodOptions(deps.Labels, resp, strings.TrimSpace(viewCtx.Request.URL.Query().Get("period"))),
			FormatOptions: buildFormatOptions(deps.Labels),
			Labels:        deps.Labels,
		}
		return view.OK("outcome-matrix-download-drawer-form", data)
	})
}

// buildPeriodOptions builds the period select: All + one option per phase that
// carries a non-empty code (ordered by sequence_order, label from the phase's
// display name) + the reserved Final. A zero-phase / codeless template yields
// only All + Final (guard-rail: no bogus semester option).
//
// preset pre-selects the option whose Value matches it — the column-header
// trigger's ?period= token. It is matched against the options this function
// just built from the LIVE response, so an unknown, stale or hand-typed token
// selects nothing and the browser lands on the first option ("All periods").
// preset == "" is the toolbar trigger's no-token case, which selects All
// explicitly rather than by browser default (they agree; being explicit keeps
// the rendered <select> self-describing).
func buildPeriodOptions(l outcome_matrix.Labels, resp *matrixpb.GetOutcomeMatrixResponse, preset string) []pyezatypes.SelectOption {
	opts := []pyezatypes.SelectOption{{Value: "", Label: l.Export.PeriodAll}}

	type phase struct {
		code, label string
		seq         int32
	}
	var phases []phase
	if resp != nil {
		for _, p := range resp.GetPhases() {
			if c := p.GetCode(); c != "" {
				phases = append(phases, phase{code: c, label: p.GetLabel(), seq: p.GetSequenceOrder()})
			}
		}
	}
	sort.SliceStable(phases, func(i, j int) bool { return phases[i].seq < phases[j].seq })
	for _, p := range phases {
		opts = append(opts, pyezatypes.SelectOption{Value: p.code, Label: p.label})
	}

	opts = append(opts, pyezatypes.SelectOption{Value: "final", Label: l.Export.PeriodFinal})

	for i := range opts {
		if opts[i].Value == preset {
			opts[i].Selected = true
			break
		}
	}
	return opts
}

// buildFormatOptions builds the format select (csv default first, pdf second).
func buildFormatOptions(l outcome_matrix.Labels) []pyezatypes.SelectOption {
	return []pyezatypes.SelectOption{
		{Value: "csv", Label: l.Export.FormatCSV},
		{Value: "pdf", Label: l.Export.FormatPDF},
	}
}
