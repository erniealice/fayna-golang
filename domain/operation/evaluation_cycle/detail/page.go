package detail

import (
	"context"
	"fmt"
	"log"

	evaluation_cycle "github.com/erniealice/fayna-golang/domain/operation/evaluation_cycle"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	cyclepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_cycle"
	memberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_cycle_member"
	evalpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation"
)

// DetailViewDeps holds dependencies for the evaluation cycle detail view.
type DetailViewDeps struct {
	Routes       evaluation_cycle.Routes
	Labels       evaluation_cycle.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	ReadEvaluationCycle        func(ctx context.Context, req *cyclepb.ReadEvaluationCycleRequest) (*cyclepb.ReadEvaluationCycleResponse, error)
	ListEvaluationCycleMembers func(ctx context.Context, req *memberpb.ListEvaluationCycleMembersRequest) (*memberpb.ListEvaluationCycleMembersResponse, error)
	// GetCycleProgress is the preferred server-side X-of-Y read-UC. When nil the
	// banner is computed inline from ListEvaluationCycleMembers + ListEvaluations.
	GetCycleProgress func(ctx context.Context, cycleID string) (*evaluation_cycle.CycleProgress, error)
	ListEvaluations  func(ctx context.Context, req *evalpb.ListEvaluationsRequest) (*evalpb.ListEvaluationsResponse, error)
}

// PageData holds the data for the evaluation cycle detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Cycle           map[string]any
	Labels          evaluation_cycle.Labels
	ActiveTab       string
	TabItems        []pyeza.TabItem

	// Banner — the shared X-of-Y READ projection (cycle-progress-banner.html).
	Banner *BannerData

	// Members tab (count = Y, TEST-2).
	Members      []map[string]any
	MemberCount  int
}

// BannerData feeds cycle-progress-banner.html. Computed (never materialized).
type BannerData struct {
	CycleID      string
	CycleName    string
	Completed    int    // X
	Total        int    // Y (frozen member denominator)
	SignOffDue   string
	CloseDate    string
	ProgressText string // "X of Y complete"
	Labels       evaluation_cycle.BannerLabels
}

func cycleToMap(c *cyclepb.EvaluationCycle) map[string]any {
	return map[string]any{
		"id":                   c.GetId(),
		"workspace_id":         c.GetWorkspaceId(),
		"subscription_id":      c.GetSubscriptionId(),
		"name":                 c.GetName(),
		"period_start":         c.GetPeriodStart(),
		"period_end":           c.GetPeriodEnd(),
		"sign_off_due_date":    c.GetSignOffDueDate(),
		"close_date":           c.GetCloseDate(),
		"status":               cycleStatusString(c.GetStatus()),
		"status_variant":       statusVariant(c.GetStatus()),
		"active":               c.GetActive(),
		"date_created_string":  c.GetDateCreatedString(),
		"date_modified_string": c.GetDateModifiedString(),
	}
}

func cycleStatusString(s cyclepb.EvaluationCycleStatus) string {
	switch s {
	case cyclepb.EvaluationCycleStatus_EVALUATION_CYCLE_STATUS_OPEN:
		return "open"
	case cyclepb.EvaluationCycleStatus_EVALUATION_CYCLE_STATUS_SIGN_OFF:
		return "sign_off"
	case cyclepb.EvaluationCycleStatus_EVALUATION_CYCLE_STATUS_CLOSED:
		return "closed"
	default:
		return "open"
	}
}

func statusVariant(s cyclepb.EvaluationCycleStatus) string {
	switch s {
	case cyclepb.EvaluationCycleStatus_EVALUATION_CYCLE_STATUS_OPEN:
		return "info"
	case cyclepb.EvaluationCycleStatus_EVALUATION_CYCLE_STATUS_SIGN_OFF:
		return "warning"
	case cyclepb.EvaluationCycleStatus_EVALUATION_CYCLE_STATUS_CLOSED:
		return "success"
	default:
		return "default"
	}
}

// NewView creates the evaluation cycle detail view (Info + Members tabs + banner).
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation_cycle", "read") {
			return view.Forbidden("evaluation_cycle:read")
		}

		id := viewCtx.Request.PathValue("id")
		cycle, err := loadCycle(ctx, deps, id)
		if err != nil {
			return view.Error(err)
		}

		l := deps.Labels
		activeTab := viewCtx.QueryParams["tab"]
		if activeTab == "" {
			activeTab = "info"
		}

		name, _ := cycle["name"].(string)
		banner := computeBanner(ctx, deps, cycle, l)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          name,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				HeaderTitle:    name,
				HeaderSubtitle: l.Detail.PageTitle,
				HeaderIcon:     "icon-refresh-cw",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "evaluation-cycle-detail-content",
			Cycle:           cycle,
			Labels:          l,
			ActiveTab:       activeTab,
			TabItems:        buildTabItems(l, id, deps.Routes),
			Banner:          banner,
		}

		loadTabData(ctx, deps, pageData, id)

		return view.OK("evaluation-cycle-detail", pageData)
	})
}

// NewMembersTabAction renders the Members tab partial (HTMX tab-swap). Member
// count = Y (TEST-2). No standalone member list page (STR-1).
func NewMembersTabAction(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation_cycle", "read") {
			return view.Forbidden("evaluation_cycle:read")
		}

		id := viewCtx.Request.PathValue("id")
		cycle, err := loadCycle(ctx, deps, id)
		if err != nil {
			return view.Error(err)
		}

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion: viewCtx.CacheVersion,
				CommonLabels: deps.CommonLabels,
			},
			Cycle:     cycle,
			Labels:    l,
			ActiveTab: "members",
			TabItems:  buildTabItems(l, id, deps.Routes),
		}
		loadTabData(ctx, deps, pageData, id)

		return view.OK("evaluation-cycle-members-tab", pageData)
	})
}

func loadCycle(ctx context.Context, deps *DetailViewDeps, id string) (map[string]any, error) {
	resp, err := deps.ReadEvaluationCycle(ctx, &cyclepb.ReadEvaluationCycleRequest{
		Data: &cyclepb.EvaluationCycle{Id: id},
	})
	if err != nil {
		log.Printf("Failed to read evaluation cycle %s: %v", id, err)
		return nil, fmt.Errorf("failed to load evaluation cycle: %w", err)
	}
	data := resp.GetData()
	if len(data) == 0 {
		log.Printf("Evaluation cycle %s not found", id)
		return nil, fmt.Errorf("evaluation cycle not found")
	}
	return cycleToMap(data[0]), nil
}

// loadTabData populates the Members tab fields (count = Y).
func loadTabData(ctx context.Context, deps *DetailViewDeps, pd *PageData, id string) {
	if pd.ActiveTab != "members" {
		return
	}
	if deps.ListEvaluationCycleMembers == nil {
		return
	}
	resp, err := deps.ListEvaluationCycleMembers(ctx, &memberpb.ListEvaluationCycleMembersRequest{})
	if err != nil {
		log.Printf("Failed to list cycle members for %s: %v", id, err)
		return
	}
	members := []map[string]any{}
	for _, m := range resp.GetData() {
		if m.GetEvaluationCycleId() != id {
			continue
		}
		members = append(members, map[string]any{
			"id":               m.GetId(),
			"subject_staff_id": m.GetSubjectStaffId(),
			"client_id":        m.GetClientId(),
			"is_probation":     m.GetIsProbation(),
		})
	}
	pd.Members = members
	pd.MemberCount = len(members) // = Y
}

// computeBanner builds the X-of-Y banner. Prefers GetCycleProgress; otherwise
// computes inline: Y = COUNT(members for cycle), X = COUNT(members joined to a
// SUBMITTED/SIGNED_OFF evaluation on (subject_staff_id, client_id)). NEVER
// materializes DRAFT evaluations (SR-1/STR-2).
func computeBanner(ctx context.Context, deps *DetailViewDeps, cycle map[string]any, l evaluation_cycle.Labels) *BannerData {
	id, _ := cycle["id"].(string)
	name, _ := cycle["name"].(string)
	signOff, _ := cycle["sign_off_due_date"].(string)
	closeDate, _ := cycle["close_date"].(string)

	x, y := 0, 0
	if deps.GetCycleProgress != nil {
		if p, err := deps.GetCycleProgress(ctx, id); err == nil && p != nil {
			x, y = p.Completed, p.Total
		}
	} else {
		x, y = computeProgressInline(ctx, deps, id)
	}

	return &BannerData{
		CycleID:      id,
		CycleName:    name,
		Completed:    x,
		Total:        y,
		SignOffDue:   signOff,
		CloseDate:    closeDate,
		ProgressText: fmt.Sprintf(l.Banner.ProgressFmt, x, y),
		Labels:       l.Banner,
	}
}

// computeProgressInline is the fallback when no server-side GetCycleProgress is
// wired. Y is the frozen member count; X counts members whose (subject_staff_id,
// client_id) has a SUBMITTED or SIGNED_OFF evaluation. This is a READ over
// existing rows — it does NOT insert anything.
func computeProgressInline(ctx context.Context, deps *DetailViewDeps, cycleID string) (x, y int) {
	if deps.ListEvaluationCycleMembers == nil {
		return 0, 0
	}
	memResp, err := deps.ListEvaluationCycleMembers(ctx, &memberpb.ListEvaluationCycleMembersRequest{})
	if err != nil {
		log.Printf("Failed to list cycle members for progress %s: %v", cycleID, err)
		return 0, 0
	}
	members := []*memberpb.EvaluationCycleMember{}
	for _, m := range memResp.GetData() {
		if m.GetEvaluationCycleId() == cycleID {
			members = append(members, m)
		}
	}
	y = len(members) // frozen denominator

	if deps.ListEvaluations == nil || y == 0 {
		return 0, y
	}
	evalResp, err := deps.ListEvaluations(ctx, &evalpb.ListEvaluationsRequest{})
	if err != nil {
		log.Printf("Failed to list evaluations for progress %s: %v", cycleID, err)
		return 0, y
	}
	// Build a set of (subject_staff_id|client_id) with a completed evaluation.
	completed := map[string]bool{}
	for _, e := range evalResp.GetData() {
		st := e.GetStatus()
		if st != evalpb.EvaluationStatus_EVALUATION_STATUS_SUBMITTED &&
			st != evalpb.EvaluationStatus_EVALUATION_STATUS_SIGNED_OFF {
			continue
		}
		completed[e.GetSubjectStaffId()+"|"+e.GetClientId()] = true
	}
	for _, m := range members {
		if completed[m.GetSubjectStaffId()+"|"+m.GetClientId()] {
			x++
		}
	}
	return x, y
}

func buildTabItems(l evaluation_cycle.Labels, id string, routes evaluation_cycle.Routes) []pyeza.TabItem {
	base := route.ResolveURL(routes.DetailURL, "id", id)
	membersAction := route.ResolveURL(routes.MembersTabURL, "id", id)
	return []pyeza.TabItem{
		{Key: "info", Label: l.Tabs.Info, Href: base + "?tab=info", HxGet: base + "?tab=info", Icon: "icon-info"},
		{Key: "members", Label: l.Tabs.Members, Href: base + "?tab=members", HxGet: membersAction, Icon: "icon-users"},
	}
}
