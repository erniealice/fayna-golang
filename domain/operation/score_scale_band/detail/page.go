package detail

import (
	"context"
	"fmt"
	"log"
	"strconv"

	score_scale_band "github.com/erniealice/fayna-golang/domain/operation/score_scale_band"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	ssbpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale_band"
)

// PageData holds the data for the score scale band detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Band            map[string]any
	Labels          score_scale_band.Labels
	ActiveTab       string
	TabItems        []pyeza.TabItem
}

// bandToMap converts a ScoreScaleBand protobuf to a map[string]any for template use.
func bandToMap(b *ssbpb.ScoreScaleBand) map[string]any {
	return map[string]any{
		"id":                    b.GetId(),
		"score_scale_id":        b.GetScoreScaleId(),
		"sequence_order":        b.GetSequenceOrder(),
		"input_min":             formatOptFloat(b.InputMin),
		"input_max":             formatOptFloat(b.InputMax),
		"input_match":           b.GetInputMatch(),
		"output_value":          formatOptFloat(b.OutputValue),
		"output_label":          b.GetOutputLabel(),
		"band_role":             b.GetBandRole(),
		"determination":         determinationString(b.GetDetermination()),
		"determination_variant": determinationVariant(b.GetDetermination()),
		"active":                b.GetActive(),
		"date_created_string":   b.GetDateCreatedString(),
		"date_modified_string":  b.GetDateModifiedString(),
	}
}

func formatOptFloat(v *float64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(*v, 'f', -1, 64)
}

func determinationString(d enums.Determination) string {
	switch d {
	case enums.Determination_DETERMINATION_PASS:
		return "Pass"
	case enums.Determination_DETERMINATION_FAIL:
		return "Fail"
	case enums.Determination_DETERMINATION_PASS_WITH_CONDITION:
		return "Pass with Condition"
	case enums.Determination_DETERMINATION_NOT_EVALUATED:
		return "Not Evaluated"
	case enums.Determination_DETERMINATION_NOT_APPLICABLE:
		return "Not Applicable"
	case enums.Determination_DETERMINATION_DEFERRED:
		return "Deferred"
	default:
		return "Unspecified"
	}
}

func determinationVariant(d enums.Determination) string {
	switch d {
	case enums.Determination_DETERMINATION_PASS:
		return "success"
	case enums.Determination_DETERMINATION_FAIL:
		return "danger"
	case enums.Determination_DETERMINATION_PASS_WITH_CONDITION:
		return "warning"
	default:
		return "default"
	}
}

// NewView creates the score scale band detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("score_scale_band", "read") {
			return view.Forbidden("score_scale_band:read")
		}
		_ = perms

		id := viewCtx.Request.PathValue("id")

		resp, err := deps.ReadScoreScaleBand(ctx, &ssbpb.ReadScoreScaleBandRequest{
			Data: &ssbpb.ScoreScaleBand{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read score scale band %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load band: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Score scale band %s not found", id)
			return view.Error(fmt.Errorf("band not found"))
		}
		band := bandToMap(data[0])

		outputLabel, _ := band["output_label"].(string)
		l := deps.Labels

		activeTab := viewCtx.QueryParams["tab"]
		if activeTab == "" {
			activeTab = "info"
		}
		tabItems := buildTabItems(l, id, deps.Routes)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          outputLabel,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				HeaderTitle:    outputLabel,
				HeaderSubtitle: l.Detail.PageTitle,
				HeaderIcon:     "icon-sliders",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "score-scale-band-detail-content",
			Band:            band,
			Labels:          l,
			ActiveTab:       activeTab,
			TabItems:        tabItems,
		}

		return view.OK("score-scale-band-detail", pageData)
	})
}

func buildTabItems(l score_scale_band.Labels, id string, routes score_scale_band.Routes) []pyeza.TabItem {
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

		resp, err := deps.ReadScoreScaleBand(ctx, &ssbpb.ReadScoreScaleBandRequest{
			Data: &ssbpb.ScoreScaleBand{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read score scale band %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load band: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			return view.Error(fmt.Errorf("band not found"))
		}
		band := bandToMap(data[0])

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion: viewCtx.CacheVersion,
				CommonLabels: deps.CommonLabels,
			},
			Band:      band,
			Labels:    l,
			ActiveTab: tab,
			TabItems:  buildTabItems(l, id, deps.Routes),
		}

		templateName := "score-scale-band-tab-" + tab
		return view.OK(templateName, pageData)
	})
}
