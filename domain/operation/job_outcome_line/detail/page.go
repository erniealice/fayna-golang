package detail

import (
	"context"
	"fmt"
	"log"

	"github.com/erniealice/fayna-golang/domain/operation/job_outcome_line"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	joboutcomelinepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_line"
)

// PageData holds the data for the job outcome line detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Line            map[string]any
	Labels          job_outcome_line.Labels
	ActiveTab       string
	TabItems        []pyeza.TabItem

	// Audit history tab
	AuditEntries    []auditlog.AuditEntryView
	AuditHasNext    bool
	AuditNextCursor string
	AuditHistoryURL string
}

// lineToMap converts a JobOutcomeLine protobuf to a map[string]any for template use.
func lineToMap(line *joboutcomelinepb.JobOutcomeLine) map[string]any {
	return map[string]any{
		"id":                     line.GetId(),
		"job_outcome_summary_id": line.GetJobOutcomeSummaryId(),
		"label":                  line.GetLabel(),
		"weight_or_credits":      line.GetWeightOrCredits(),
		"output_value":           line.GetOutputValue(),
		"output_label":           line.GetOutputLabel(),
		"score_scale_band_id":    line.GetScoreScaleBandId(),
		"reporting_role":         reportingRoleString(line.GetReportingRole()),
		"reporting_role_variant": reportingRoleVariant(line.GetReportingRole()),
		"active":                 line.GetActive(),
		"date_created_string":    line.GetDateCreatedString(),
		"date_modified_string":   line.GetDateModifiedString(),
	}
}

func reportingRoleString(r enums.ReportingRole) string {
	switch r {
	case enums.ReportingRole_REPORTING_ROLE_PRIMARY:
		return "Primary"
	case enums.ReportingRole_REPORTING_ROLE_ALTERNATE:
		return "Alternate"
	case enums.ReportingRole_REPORTING_ROLE_TRANSCRIPT:
		return "Transcript"
	case enums.ReportingRole_REPORTING_ROLE_PERCENTILE:
		return "Percentile"
	default:
		return "Unspecified"
	}
}

func reportingRoleVariant(r enums.ReportingRole) string {
	switch r {
	case enums.ReportingRole_REPORTING_ROLE_PRIMARY:
		return "success"
	case enums.ReportingRole_REPORTING_ROLE_TRANSCRIPT:
		return "info"
	case enums.ReportingRole_REPORTING_ROLE_PERCENTILE:
		return "warning"
	default:
		return "default"
	}
}

// loadTabData populates tab-specific fields on pageData based on the active tab.
func loadTabData(ctx context.Context, deps *DetailViewDeps, pd *PageData, id string, viewCtx *view.ViewContext) {
	switch pd.ActiveTab {
	case "audit-history":
		if deps.ListAuditHistory != nil {
			cursor := viewCtx.QueryParams["cursor"]
			auditResp, err := deps.ListAuditHistory(ctx, &auditlog.ListAuditRequest{
				EntityType:  "job_outcome_line",
				EntityID:    id,
				Limit:       20,
				CursorToken: cursor,
			})
			if err != nil {
				log.Printf("Failed to load audit history: %v", err)
			}
			if auditResp != nil {
				pd.AuditEntries = auditResp.Entries
				pd.AuditHasNext = auditResp.HasNext
				pd.AuditNextCursor = auditResp.NextCursor
			}
		}
		pd.AuditHistoryURL = route.ResolveURL(deps.Routes.TabActionURL, "id", id, "tab", "") + "audit-history"
	}
}

// NewView creates the job outcome line detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_outcome_line", "read") {
			return view.Forbidden("job_outcome_line:read")
		}
		_ = perms

		id := viewCtx.Request.PathValue("id")

		resp, err := deps.ReadJobOutcomeLine(ctx, &joboutcomelinepb.ReadJobOutcomeLineRequest{
			Data: &joboutcomelinepb.JobOutcomeLine{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read job outcome line %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load outcome line: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Job outcome line %s not found", id)
			return view.Error(fmt.Errorf("outcome line not found"))
		}
		lineMap := lineToMap(data[0])

		label, _ := lineMap["label"].(string)
		l := deps.Labels

		activeTab := viewCtx.QueryParams["tab"]
		if activeTab == "" {
			activeTab = "info"
		}
		tabItems := buildTabItems(l, id, deps.Routes)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          label,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				HeaderTitle:    label,
				HeaderSubtitle: l.Detail.PageTitle,
				HeaderIcon:     "icon-list",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "job-outcome-line-detail-content",
			Line:            lineMap,
			Labels:          l,
			ActiveTab:       activeTab,
			TabItems:        tabItems,
		}

		loadTabData(ctx, deps, pageData, id, viewCtx)

		return view.OK("job-outcome-line-detail", pageData)
	})
}

func buildTabItems(l job_outcome_line.Labels, id string, routes job_outcome_line.Routes) []pyeza.TabItem {
	base := route.ResolveURL(routes.DetailURL, "id", id)
	action := route.ResolveURL(routes.TabActionURL, "id", id, "tab", "")
	return []pyeza.TabItem{
		{Key: "info", Label: l.Tabs.Info, Href: base + "?tab=info", HxGet: action + "info", Icon: "icon-info"},
		{Key: "audit-history", Label: l.Tabs.History, Href: base + "?tab=audit-history", HxGet: action + "audit-history", Icon: "icon-clock"},
	}
}

// NewTabAction creates the tab action view (partial — returns only the tab content).
func NewTabAction(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		id := viewCtx.Request.PathValue("id")
		tab := viewCtx.Request.PathValue("tab")
		if tab == "" {
			tab = "info"
		}

		resp, err := deps.ReadJobOutcomeLine(ctx, &joboutcomelinepb.ReadJobOutcomeLineRequest{
			Data: &joboutcomelinepb.JobOutcomeLine{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read job outcome line %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load outcome line: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Job outcome line %s not found", id)
			return view.Error(fmt.Errorf("outcome line not found"))
		}
		lineMap := lineToMap(data[0])

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion: viewCtx.CacheVersion,
				CommonLabels: deps.CommonLabels,
			},
			Line:      lineMap,
			Labels:    l,
			ActiveTab: tab,
			TabItems:  buildTabItems(l, id, deps.Routes),
		}

		loadTabData(ctx, deps, pageData, id, viewCtx)

		templateName := "job-outcome-line-tab-" + tab
		if tab == "audit-history" {
			templateName = "audit-history-tab"
		}
		return view.OK(templateName, pageData)
	})
}
