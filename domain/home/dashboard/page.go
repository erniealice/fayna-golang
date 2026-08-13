// Package dashboard implements the persona-aware home dashboard views
// (docs/plan/20260801-persona-home-dashboard §4 — Q1–Q7 locked 2026-08-01,
// including the §4.3 amendment banner ride-alongs: the A2 current-period
// window is rendered WITH its period visible, returned + not-started counts
// surface on the approval series, and one operational-exceptions row feeds
// the attention widget).
//
// Persona axis (runtime, R1): the session's principal kind — resolved by a
// COMPOSITION-INJECTED closure (the getDashboardData precedent); this package
// never imports espyna/shared/identity — picks the declared variant
// fail-closed (home.Options.VariantFor).
//
// Option axis (deploy-time, R2): the app declares rosters/tabs/links through
// home.Options; zero value = the block never mounts (inert).
//
// Degrade contract (Q4/T-9): every data widget renders in exactly one of the
// designed states — full, empty (visible, labelled), or denied (visible card
// + lyngua reason + data-testid="home-widget-{ref}-denied" + a server-side
// AUTHZ_RBAC_DENY log line). A deny NEVER silently removes a widget and the
// page stays 200.
package dashboard

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	home "github.com/erniealice/fayna-golang/domain/home"

	ocpb "github.com/erniealice/esqyma/pkg/schema/v1/service/dashboard/outcome_completion"
)

// ErrSummaryDenied classifies a GetCompletionSummary failure as an
// AUTHORIZATION deny (vs a data/infrastructure error). The composition-side
// closure wraps use-case authorization errors with this sentinel
// (errors.Is); the view then renders the designed DENIED cards + emits the
// AUTHZ_RBAC_DENY log line — never the benign "nothing assigned" empty state
// (Q4/T-9: a deny must be visible, labelled, and logged).
var ErrSummaryDenied = errors.New("home dashboard: completion summary read denied")

// Deps holds the view dependencies. Every closure is nil-safe: a nil
// GetCompletionSummary degrades data widgets to their empty state; a nil
// ResolvePrincipalKind resolves kind 0 (→ the fail-closed default variant);
// a nil ResolveWorkspaceName omits the workspace line; a nil LogDeny skips
// only the log line (the visible denied card still renders).
type Deps struct {
	Routes       home.Routes
	Options      home.Options
	CommonLabels pyeza.CommonLabels

	// GetCompletionSummary is the ONE aggregate read (Q5 lock) behind the
	// completion / approval / attention widgets AND the category tabstrip.
	// Identity and row scope come from ctx inside the use case/adapter —
	// never from this package (Q-EIB-BRIDGE).
	GetCompletionSummary func(ctx context.Context) (*ocpb.GetOutcomeCompletionSummaryResponse, error)

	// ResolvePrincipalKind maps the session to its principal kind (0 =
	// unresolved). Injected by the consuming app's composition layer.
	ResolvePrincipalKind func(ctx context.Context) int32

	// ResolveWorkspaceName resolves the session workspace's display name
	// (replacing the retired hardcoded demo name).
	ResolveWorkspaceName func(ctx context.Context) string

	// LogDeny emits the AUTHZ_RBAC_DENY server-log line for a view-level
	// widget deny (the T-9 observability half). Injected by the block so
	// this package needs no espyna context accessors.
	LogDeny func(ctx context.Context, permissionCode, widgetRef string)

	// ResolveOverviewURL returns the WORKSPACE-QUALIFIED overview URL the bare
	// /home redirect targets (20260809 D2/M5). Injected by the app's
	// composition so this package never imports espyna's workspace-slug
	// helpers (dependency-direction rule). Nil/empty ⇒ the bare Routes.OverviewURL.
	ResolveOverviewURL func(ctx context.Context) string
}

// Module registers the routed home surface and its retained partial endpoints.
type Module struct{ deps *Deps }

// NewModule creates the home dashboard module.
func NewModule(deps *Deps) *Module { return &Module{deps: deps} }

// RegisterRoutes registers the routed-section surface (20260809):
//   - GET /home            → workspace-aware 302 redirect to /home/overview
//   - GET /home/{section}  → the four explicit section views
//   - GET /home/content    → RETAINED overview-content compat endpoint (M4)
//   - GET /action/home/ribbon → the ribbon self-refresh
//
// Section routes are EXPLICIT (no {section} wildcard): each handler is bound to
// its own SectionKey. An unrecognized /home/{x} is simply unregistered → 404;
// nothing links to it (legacy /home#pulse strips the fragment client-side to
// /home, which redirects to overview).
func (m *Module) RegisterRoutes(r pyeza.RouteRegistrar) {
	m.deps.Routes = withRouteDefaults(m.deps.Routes)
	rt := m.deps.Routes
	r.GET(rt.DashboardURL, m.redirectToOverview())
	r.GET(rt.OverviewURL, m.sectionView(home.SectionOverview))
	r.GET(rt.PulseURL, m.sectionView(home.SectionPulse))
	r.GET(rt.PerformanceURL, m.sectionView(home.SectionPerformance))
	r.GET(rt.AttentionURL, m.sectionView(home.SectionAttention))
	r.GET(rt.ContentURL, m.contentView())
	r.GET(rt.RibbonURL, m.ribbonView())
}

func withRouteDefaults(r home.Routes) home.Routes {
	d := home.DefaultRoutes()
	if r.DashboardURL == "" {
		r.DashboardURL = d.DashboardURL
	}
	if r.ContentURL == "" {
		r.ContentURL = d.ContentURL
	}
	if r.RibbonURL == "" {
		r.RibbonURL = d.RibbonURL
	}
	if r.OverviewURL == "" {
		r.OverviewURL = d.OverviewURL
	}
	if r.PulseURL == "" {
		r.PulseURL = d.PulseURL
	}
	if r.PerformanceURL == "" {
		r.PerformanceURL = d.PerformanceURL
	}
	if r.AttentionURL == "" {
		r.AttentionURL = d.AttentionURL
	}
	return r
}

// PageData is the template payload for all three templates.
type PageData struct {
	types.PageData

	// WorkspaceName is the resolved workspace display name ("" omits it).
	WorkspaceName string
	// Today is the server-rendered current date line.
	Today string
	// Slot is the resolved persona slot (data-testid="home-variant-{slot}").
	Slot int32

	// Section is the resolved routed section key (data-testid="home-section-{key}").
	Section string
	// SectionTitle / SectionSubtitle are resolved in Go from the dynamic
	// "home.section.{key}.title/subtitle" lyngua keys (20260809 L2 — a Go
	// template cannot interpolate {key} inside a quoted .T argument).
	SectionTitle    string
	SectionSubtitle string
	// RefreshRouteKey is the explicit route-map key for the resolved section.
	// It keeps refreshes on that section instead of calling the retained
	// overview-only /home/content compatibility endpoint.
	RefreshRouteKey string

	// TabsEnabled / Tabs / ActiveTab drive the Q2 category tabstrip.
	TabsEnabled bool
	Tabs        []TabItemData
	ActiveTab   string // "all" or the selected category id

	// Widgets is the resolved, ordered roster with per-widget state.
	Widgets []WidgetData
}

// TabItemData is one tabstrip entry.
type TabItemData struct {
	// ID doubles as the element id and data-testid ("home-tab-all" /
	// "home-tab-{category_id}" — the §A-6 selector contract).
	ID     string
	Label  string
	Active bool
	// Query is the content-URL suffix ("" for All, "?t={category_id}").
	Query string
}

// redirectToOverview serves the bare /home route as a workspace-aware redirect
// to /home/overview (20260809 D2). The target is resolved by the app-injected
// ResolveOverviewURL closure (workspace-qualified — M5); it falls back to the
// bare Routes.OverviewURL when unwired. The ViewAdapter emits HX-Redirect for
// HTMX requests and a real 3xx otherwise.
func (m *Module) redirectToOverview() view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		target := ""
		if m.deps.ResolveOverviewURL != nil {
			target = m.deps.ResolveOverviewURL(ctx)
		}
		if target == "" {
			target = m.deps.Routes.OverviewURL
		}
		return view.ViewResult{Redirect: target, StatusCode: http.StatusFound}
	})
}

// sectionView renders one routed section as a full page (or, on an HTMX
// main-content swap, its content partial — handled by the ViewAdapter).
func (m *Module) sectionView(key home.SectionKey) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		return view.OK("fayna-home", m.buildSectionData(ctx, viewCtx, key))
	})
}

// contentView serves the retained /home/content compat endpoint as the
// overview section's content (M4).
func (m *Module) contentView() view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		return view.OK("fayna-home-content", m.buildSectionData(ctx, viewCtx, home.SectionOverview))
	})
}

func (m *Module) ribbonView() view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		// The ribbon reads NO data (no per-60s aggregate re-query); it only
		// needs the date + workspace line.
		d := m.baseData(ctx, viewCtx)
		return view.OK("fayna-home-ribbon", d)
	})
}

// baseData assembles the data-free page chrome shared by all three views.
func (m *Module) baseData(ctx context.Context, viewCtx *view.ViewContext) *PageData {
	deps := m.deps
	wsName := ""
	if deps.ResolveWorkspaceName != nil {
		wsName = deps.ResolveWorkspaceName(ctx)
	}
	kind := int32(0)
	if deps.ResolvePrincipalKind != nil {
		kind = deps.ResolvePrincipalKind(ctx)
	}
	_, slot, _ := deps.Options.VariantFor(kind)

	return &PageData{
		PageData: types.PageData{
			CacheVersion:    viewCtx.CacheVersion,
			Title:           viewCtx.T("home.title"),
			CurrentPath:     viewCtx.CurrentPath,
			ActiveNav:       "home",
			ContentTemplate: "fayna-home-content",
			HeaderTitle:     viewCtx.T("home.title"),
			HeaderIcon:      "icon-layout-dashboard",
			CommonLabels:    deps.CommonLabels,
			Messages:        viewCtx.Messages,
		},
		WorkspaceName: wsName,
		Today:         time.Now().Format("Monday, 2 January 2006"),
		Slot:          slot,
	}
}

// buildData assembles the OVERVIEW section payload — the back-compat entry
// point (contentView + older callers). New code uses buildSectionData.
func (m *Module) buildData(ctx context.Context, viewCtx *view.ViewContext) *PageData {
	return m.buildSectionData(ctx, viewCtx, home.SectionOverview)
}

// buildSectionData assembles one routed section's payload: variant→section
// resolution, the single aggregate read, tab selection (section-gated), and
// per-widget states.
func (m *Module) buildSectionData(ctx context.Context, viewCtx *view.ViewContext, sectionKey home.SectionKey) *PageData {
	deps := m.deps
	data := m.baseData(ctx, viewCtx)

	kind := int32(0)
	if deps.ResolvePrincipalKind != nil {
		kind = deps.ResolvePrincipalKind(ctx)
	}
	section, slot, ok := deps.Options.SectionFor(kind, sectionKey)
	data.Slot = slot
	if !ok {
		return data
	}
	data.Section = string(section.Key)
	data.SectionTitle = viewCtx.T("home.section." + string(section.Key) + ".title")
	data.SectionSubtitle = viewCtx.T("home.section." + string(section.Key) + ".subtitle")
	data.RefreshRouteKey = sectionRouteKey(section.Key)
	roster := section.KnownWidgets()

	perms := view.GetUserPermissions(ctx)

	// The single aggregate read (Q5) — issued once, only when at least one
	// data widget's ENTIRE bundle resolves (Q4 all-or-nothing). A denied
	// bundle never issues the read for that widget. A read FAILURE is
	// classified (skeptic F2, T-9): an authorization deny (ErrSummaryDenied,
	// wrapped by the composition closure) renders the designed DENIED cards +
	// AUTHZ_RBAC_DENY log; any other error renders the labelled UNAVAILABLE
	// state — never the benign "nothing assigned" empty state, and the page
	// stays 200 either way.
	var resp *ocpb.GetOutcomeCompletionSummaryResponse
	var summaryDenied, summaryFailed bool
	dataNeeded := false
	for _, ref := range roster {
		if len(home.WidgetBundle(ref)) > 0 && bundleResolves(perms, home.WidgetBundle(ref)) {
			dataNeeded = true
			break
		}
	}
	if dataNeeded && deps.GetCompletionSummary != nil {
		r, err := deps.GetCompletionSummary(ctx)
		switch {
		case err == nil:
			resp = r
		case errors.Is(err, ErrSummaryDenied):
			summaryDenied = true
			log.Printf("home dashboard: completion summary read DENIED (rendering denied cards): %v", err)
		default:
			summaryFailed = true
			log.Printf("home dashboard: completion summary read failed (rendering unavailable states): %v", err)
		}
	}

	// Tab axis (Q2): category tabs + All, from the aggregate response (the
	// tab render shares the widget bundle — copya.md §C-1). The `t` query
	// param is validated against the live category ids; unknown ⇒ All
	// (fail-safe, mirroring the TabOptions grammar).
	selected := ""
	if viewCtx.Request != nil {
		selected = strings.TrimSpace(viewCtx.Request.URL.Query().Get("t"))
	}
	if resp == nil || !categoryExists(resp.GetCategoryRows(), selected) {
		selected = ""
	}
	data.ActiveTab = "all"
	if selected != "" {
		data.ActiveTab = selected
	}
	if section.HasTabs && deps.Options.Tab.Enabled() && resp != nil {
		data.TabsEnabled = true
		data.Tabs = buildTabs(viewCtx, resp.GetCategoryRows(), selected)
	}

	// Selected row: the All rollup or the matching category row.
	row := selectRow(resp, selected)

	for _, ref := range roster {
		data.Widgets = append(data.Widgets, m.buildWidget(ctx, viewCtx, ref, perms, resp, row, slot, summaryDenied, summaryFailed))
	}
	return data
}

func sectionRouteKey(key home.SectionKey) string {
	switch key {
	case home.SectionOverview:
		return "home.overview_url"
	case home.SectionPulse:
		return "home.pulse_url"
	case home.SectionPerformance:
		return "home.performance_url"
	case home.SectionAttention:
		return "home.attention_url"
	default:
		return ""
	}
}

// bundleResolves reports whether EVERY code in the bundle is held (nil/empty
// permission set fails closed — pyeza HasCode nil-receiver semantics).
func bundleResolves(perms *types.UserPermissions, bundle []string) bool {
	for _, code := range bundle {
		if !perms.HasCode(code) {
			return false
		}
	}
	return true
}

// buildWidget assembles one widget's designed state.
func (m *Module) buildWidget(
	ctx context.Context,
	viewCtx *view.ViewContext,
	ref home.WidgetRef,
	perms *types.UserPermissions,
	resp *ocpb.GetOutcomeCompletionSummaryResponse,
	row *ocpb.OutcomeCompletionCategoryRow,
	slot int32,
	summaryDenied bool,
	summaryFailed bool,
) WidgetData {
	w := WidgetData{Ref: string(ref)}

	// Q4/T-9: bundle deny → visible, labelled, logged denied card.
	if bundle := home.WidgetBundle(ref); len(bundle) > 0 && !bundleResolves(perms, bundle) {
		w.State = StateDenied
		w.Title = widgetTitle(viewCtx, ref)
		w.DeniedText = viewCtx.T("home.widget.denied")
		w.DeniedHint = viewCtx.T("home.widget.denied_hint")
		if m.deps.LogDeny != nil {
			for _, code := range bundle {
				if !perms.HasCode(code) {
					m.deps.LogDeny(ctx, code, string(ref))
				}
			}
		}
		return w
	}

	// The aggregate read itself was refused by the use-case gate (view
	// permission set and use-case verdict can disagree — e.g. a legacy-path
	// permission union vs the kind allowlist). Same designed DENIED card +
	// AUTHZ_RBAC_DENY log line as a bundle deny (skeptic F2: an authorization
	// deny must never be mislabelled "you have no work").
	if bundle := home.WidgetBundle(ref); len(bundle) > 0 && summaryDenied {
		w.State = StateDenied
		w.Title = widgetTitle(viewCtx, ref)
		w.DeniedText = viewCtx.T("home.widget.denied")
		w.DeniedHint = viewCtx.T("home.widget.denied_hint")
		if m.deps.LogDeny != nil {
			for _, code := range bundle {
				m.deps.LogDeny(ctx, code, string(ref))
			}
		}
		return w
	}

	// Data/infrastructure failure → the labelled UNAVAILABLE state (distinct
	// from both the deny and the legitimate empty state — T-9's "never
	// silently blank" rule for data widgets).
	if len(home.WidgetBundle(ref)) > 0 && summaryFailed {
		w.State = StateUnavailable
		w.Title = widgetTitle(viewCtx, ref)
		w.UnavailableText = viewCtx.T("home.widget.unavailable")
		w.UnavailableHint = viewCtx.T("home.widget.unavailable_hint")
		return w
	}

	switch ref {
	case home.WidgetCompletionSummary:
		return buildCompletionWidget(viewCtx, row, slot)
	case home.WidgetApprovalProgress:
		return buildApprovalWidget(viewCtx, row)
	case home.WidgetAttention:
		return buildAttentionWidget(viewCtx, resp)
	case home.WidgetReviewPipeline:
		return buildReviewPipelineWidget(viewCtx, row)
	case home.WidgetPerformanceTable:
		return buildPerformanceWidget(viewCtx, resp, m.deps.Options.PerformanceTargetBasisPoints)
	case home.WidgetQuickLinks:
		return buildQuickLinksWidget(viewCtx, m.deps.Options.QuickLinks, perms)
	}
	// Unreachable for KnownWidgets rosters; keep fail-safe anyway.
	w.State = StateEmpty
	return w
}

// buildTabs renders the All tab plus one tab per category row (the rows
// arrive pre-ordered by job_category.sort_order asc from the adapter).
func buildTabs(viewCtx *view.ViewContext, rows []*ocpb.OutcomeCompletionCategoryRow, selected string) []TabItemData {
	tabs := make([]TabItemData, 0, len(rows)+1)
	tabs = append(tabs, TabItemData{
		ID:     "home-tab-all",
		Label:  viewCtx.T("home.tab.all"),
		Active: selected == "",
	})
	for _, r := range rows {
		if r.GetCategoryId() == "" {
			continue
		}
		tabs = append(tabs, TabItemData{
			ID:     "home-tab-" + r.GetCategoryId(),
			Label:  r.GetCategoryName(),
			Active: selected == r.GetCategoryId(),
			Query:  "?t=" + r.GetCategoryId(),
		})
	}
	return tabs
}

// categoryExists guards the ?t= param against stale/foreign/tampered ids.
func categoryExists(rows []*ocpb.OutcomeCompletionCategoryRow, id string) bool {
	if id == "" {
		return false
	}
	for _, r := range rows {
		if r.GetCategoryId() == id {
			return true
		}
	}
	return false
}

// selectRow returns the All rollup ("" selection) or the matching category
// row; nil when the response is absent.
func selectRow(resp *ocpb.GetOutcomeCompletionSummaryResponse, selected string) *ocpb.OutcomeCompletionCategoryRow {
	if resp == nil {
		return nil
	}
	if selected == "" {
		return resp.GetAllRollup()
	}
	for _, r := range resp.GetCategoryRows() {
		if r.GetCategoryId() == selected {
			return r
		}
	}
	return nil
}

// widgetTitle resolves a widget's lyngua title (used on denied cards, where
// the full builder never runs).
func widgetTitle(viewCtx *view.ViewContext, ref home.WidgetRef) string {
	switch ref {
	case home.WidgetCompletionSummary:
		return viewCtx.T("home.widget.completion.title")
	case home.WidgetApprovalProgress:
		return viewCtx.T("home.widget.approval.title")
	case home.WidgetAttention:
		return viewCtx.T("home.widget.attention.title")
	case home.WidgetReviewPipeline:
		return viewCtx.T("home.widget.pipeline.title")
	case home.WidgetPerformanceTable:
		return viewCtx.T("home.widget.performance.title")
	case home.WidgetQuickLinks:
		return viewCtx.T("home.widget.quick.title")
	}
	return string(ref)
}

// formatCount renders an int64 count for display.
func formatCount(n int64) string { return strconv.FormatInt(n, 10) }
