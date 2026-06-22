package detail

import (
	"context"
	"fmt"
	"log"

	scoring_scheme "github.com/erniealice/fayna-golang/domain/operation/scoring_scheme"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	schemepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/scoring_scheme"
)

// PageData holds the data for the scoring scheme detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Scheme          map[string]any
	Labels          scoring_scheme.Labels
	ActiveTab       string
	TabItems        []pyeza.TabItem

	// Tab data — nil means the tab renders a coming-soon panel.
	CriteriaTable *types.TableConfig
	VersionsTable *types.TableConfig
	// Audit history tab
	AuditEntries    []auditlog.AuditEntryView
	AuditHasNext    bool
	AuditNextCursor string
	AuditHistoryURL string
}

// schemeToMap converts a ScoringScheme protobuf to a map[string]any for template use.
func schemeToMap(s *schemepb.ScoringScheme) map[string]any {
	return map[string]any{
		"id":                      s.GetId(),
		"name":                    s.GetName(),
		"composite_method":        compositeMethodString(s.GetCompositeMethod()),
		"rounding_mode":           roundingModeString(s.GetRoundingMode()),
		"weights_must_sum_to_one": s.GetWeightsMustSumToOne(),
		"score_scale_id":          s.GetScoreScaleId(),
		"scheme_group_id":         s.GetSchemeGroupId(),
		"version":                 s.GetVersion(),
		"version_status":          versionStatusString(s.GetVersionStatus()),
		"version_status_variant":  versionStatusVariant(s.GetVersionStatus()),
		"active":                  s.GetActive(),
		"date_created_string":     s.GetDateCreatedString(),
		"date_modified_string":    s.GetDateModifiedString(),
	}
}

func compositeMethodString(m enums.ScoringMethod) string {
	switch m {
	case enums.ScoringMethod_SCORING_METHOD_EQUAL_WEIGHT:
		return "Equal Weight"
	case enums.ScoringMethod_SCORING_METHOD_WEIGHTED_AVERAGE:
		return "Weighted Average"
	case enums.ScoringMethod_SCORING_METHOD_MINIMUM_DETERMINATION:
		return "Minimum Determination"
	case enums.ScoringMethod_SCORING_METHOD_PERCENTAGE_PASS:
		return "Percentage Pass"
	case enums.ScoringMethod_SCORING_METHOD_SUM:
		return "Sum"
	default:
		return "Unspecified"
	}
}

func roundingModeString(m enums.RoundingMode) string {
	switch m {
	case enums.RoundingMode_ROUNDING_MODE_HALF_UP:
		return "Half Up"
	case enums.RoundingMode_ROUNDING_MODE_HALF_DOWN:
		return "Half Down"
	case enums.RoundingMode_ROUNDING_MODE_HALF_EVEN:
		return "Half Even"
	default:
		return "Unspecified"
	}
}

func versionStatusString(s enums.VersionStatus) string {
	switch s {
	case enums.VersionStatus_VERSION_STATUS_DRAFT:
		return "draft"
	case enums.VersionStatus_VERSION_STATUS_PUBLISHED:
		return "published"
	case enums.VersionStatus_VERSION_STATUS_DEPRECATED:
		return "deprecated"
	default:
		return "draft"
	}
}

func versionStatusVariant(s enums.VersionStatus) string {
	switch s {
	case enums.VersionStatus_VERSION_STATUS_DRAFT:
		return "default"
	case enums.VersionStatus_VERSION_STATUS_PUBLISHED:
		return "success"
	case enums.VersionStatus_VERSION_STATUS_DEPRECATED:
		return "warning"
	default:
		return "default"
	}
}

// loadTabData populates the tab-specific fields on pageData based on the active tab.
func loadTabData(ctx context.Context, deps *DetailViewDeps, pd *PageData, id string, viewCtx *view.ViewContext) {
	switch pd.ActiveTab {
	case "criteria":
		// TODO: load criteria linked to this scheme
		pd.CriteriaTable = nil
	case "versions":
		// TODO: load versions table data
		pd.VersionsTable = nil
	case "audit-history":
		if deps.ListAuditHistory != nil {
			cursor := viewCtx.QueryParams["cursor"]
			auditResp, err := deps.ListAuditHistory(ctx, &auditlog.ListAuditRequest{
				EntityType:  "scoring_scheme",
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

// NewView creates the scoring scheme detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("scoring_scheme", "read") {
			return view.Forbidden("scoring_scheme:read")
		}
		_ = perms

		id := viewCtx.Request.PathValue("id")

		resp, err := deps.ReadScoringScheme(ctx, &schemepb.ReadScoringSchemeRequest{
			Data: &schemepb.ScoringScheme{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read scoring scheme %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load scoring scheme: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Scoring scheme %s not found", id)
			return view.Error(fmt.Errorf("scoring scheme not found"))
		}
		scheme := schemeToMap(data[0])

		name, _ := scheme["name"].(string)
		l := deps.Labels

		activeTab := viewCtx.QueryParams["tab"]
		if activeTab == "" {
			activeTab = "info"
		}
		tabItems := buildTabItems(l, id, deps.Routes)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          name,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				HeaderTitle:    name,
				HeaderSubtitle: l.Detail.PageTitle,
				HeaderIcon:     "icon-sliders",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "scoring-scheme-detail-content",
			Scheme:          scheme,
			Labels:          l,
			ActiveTab:       activeTab,
			TabItems:        tabItems,
		}

		loadTabData(ctx, deps, pageData, id, viewCtx)

		return view.OK("scoring-scheme-detail", pageData)
	})
}

func buildTabItems(l scoring_scheme.Labels, id string, routes scoring_scheme.Routes) []pyeza.TabItem {
	base := route.ResolveURL(routes.DetailURL, "id", id)
	action := route.ResolveURL(routes.TabActionURL, "id", id, "tab", "")
	return []pyeza.TabItem{
		{Key: "info", Label: l.Tabs.Info, Href: base + "?tab=info", HxGet: action + "info", Icon: "icon-info"},
		{Key: "criteria", Label: l.Tabs.Criteria, Href: base + "?tab=criteria", HxGet: action + "criteria", Icon: "icon-check-square"},
		{Key: "versions", Label: l.Tabs.Versions, Href: base + "?tab=versions", HxGet: action + "versions", Icon: "icon-clock"},
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

		resp, err := deps.ReadScoringScheme(ctx, &schemepb.ReadScoringSchemeRequest{
			Data: &schemepb.ScoringScheme{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read scoring scheme %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load scoring scheme: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Scoring scheme %s not found", id)
			return view.Error(fmt.Errorf("scoring scheme not found"))
		}
		scheme := schemeToMap(data[0])

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion: viewCtx.CacheVersion,
				CommonLabels: deps.CommonLabels,
			},
			Scheme:    scheme,
			Labels:    l,
			ActiveTab: tab,
			TabItems:  buildTabItems(l, id, deps.Routes),
		}

		loadTabData(ctx, deps, pageData, id, viewCtx)

		templateName := "scoring-scheme-tab-" + tab
		if tab == "audit-history" {
			templateName = "audit-history-tab"
		}
		return view.OK(templateName, pageData)
	})
}
