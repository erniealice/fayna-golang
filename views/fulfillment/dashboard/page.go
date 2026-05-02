// Package dashboard implements the read-only Fulfillment live dashboard view
// (Phase 3 — Pyeza dashboard block + per-app live dashboards plan).
package dashboard

import (
	"context"
	"fmt"
	"strconv"
	"time"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	fayna "github.com/erniealice/fayna-golang"

	fulfillmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/fulfillment"
)

// Request mirrors the use-case request shape — kept narrow so the view
// package does not import espyna directly.
type Request struct {
	WorkspaceID string
	Now         time.Time
}

// Response mirrors the use-case response shape.
type Response struct {
	Pending          int64
	InTransit        int64
	DeliveredToday   int64
	Exceptions       int64
	AvgFulfillDays   float64
	StatusMixLabels  []string
	StatusMixValues  []float64
	TrendLabels      []string
	TrendValues      []float64
	RecentExceptions []*fulfillmentpb.Fulfillment
}

// Deps holds view dependencies.
type Deps struct {
	Routes               fayna.FulfillmentRoutes
	Labels               fayna.FulfillmentLabels
	CommonLabels         pyeza.CommonLabels
	GetDashboardPageData func(ctx context.Context, req *Request) (*Response, error)
}

// PageData is the dashboard template payload.
type PageData struct {
	types.PageData
	ContentTemplate string
	Dashboard       types.DashboardData
}

// NewView creates the fulfillment dashboard view.
func NewView(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		l := deps.Labels.Dashboard

		var resp *Response
		if deps.GetDashboardPageData != nil {
			r, err := deps.GetDashboardPageData(ctx, &Request{WorkspaceID: "", Now: time.Now()})
			if err == nil && r != nil {
				resp = r
			}
		}
		if resp == nil {
			resp = &Response{}
		}

		// 30-day daily-delivered chart.
		dailyLabels := resp.TrendLabels
		dailyValues := resp.TrendValues
		if len(dailyLabels) == 0 {
			dailyLabels = make([]string, 30)
			dailyValues = make([]float64, 30)
		}
		dailyChart := &types.ChartData{
			Labels: dailyLabels,
			Series: []types.ChartSeries{{
				Name:   l.DailyDelivered,
				Values: dailyValues,
				Color:  "sage",
			}},
			YAxis: l.AxisCount,
		}
		dailyChart.AutoScale()

		// Donut: status mix.
		statusLabels := resp.StatusMixLabels
		statusValues := resp.StatusMixValues
		if len(statusLabels) == 0 {
			// donut empty — fall through to empty state via widget body
			statusLabels = []string{l.StatPending}
			statusValues = []float64{0}
		}
		statusChart := &types.ChartData{
			Labels: statusLabels,
			Series: []types.ChartSeries{{
				Name:   l.StatusMix,
				Values: statusValues,
				Color:  "navy",
			}},
			Donut: true,
		}
		statusChart.AutoScale()

		// Recent exceptions list.
		recents := buildExceptionsList(resp.RecentExceptions, l)

		dash := types.DashboardData{
			Title:    l.Title,
			Icon:     "icon-truck",
			Subtitle: l.Subtitle,
			QuickActions: []types.QuickAction{
				{Icon: "icon-plus", Label: l.QuickNewFulfillment, Href: deps.Routes.AddURL, Variant: "primary", TestID: "fulfillment-action-new"},
				{Icon: "icon-package", Label: l.QuickPickPack, Href: deps.Routes.ListURL, TestID: "fulfillment-action-pick-pack"},
				{Icon: "icon-corner-up-left", Label: l.QuickProcessReturn, Href: deps.Routes.ListURL, TestID: "fulfillment-action-return"},
				{Icon: "icon-check", Label: l.QuickMarkDelivered, Href: deps.Routes.ListURL, TestID: "fulfillment-action-mark-delivered"},
			},
			Stats: []types.StatCardData{
				{Icon: "icon-clock", Value: strconv.FormatInt(resp.Pending, 10), Label: l.StatPending, Color: "amber", TestID: "fulfillment-stat-pending"},
				{Icon: "icon-truck", Value: strconv.FormatInt(resp.InTransit, 10), Label: l.StatInTransit, Color: "navy", TestID: "fulfillment-stat-in-transit"},
				{Icon: "icon-check-circle", Value: strconv.FormatInt(resp.DeliveredToday, 10), Label: l.StatDeliveredToday, Color: "sage", TestID: "fulfillment-stat-delivered-today"},
				{Icon: "icon-alert-triangle", Value: strconv.FormatInt(resp.Exceptions, 10), Label: l.StatExceptions, Color: "terracotta", TestID: "fulfillment-stat-exceptions"},
			},
			Widgets: []types.DashboardWidget{
				{
					ID: "daily-delivered", Title: l.DailyDelivered,
					Type: "chart", ChartKind: "line",
					ChartData: dailyChart, Span: 2,
				},
				{
					ID: "status-mix", Title: l.StatusMix,
					Type: "chart", ChartKind: "donut",
					ChartData: statusChart, Span: 1,
				},
				{
					ID: "recent-exceptions", Title: l.RecentExceptions, Type: "list", Span: 2,
					HeaderActions: []types.QuickAction{
						{Label: l.ViewAll, Href: deps.Routes.ListURL},
					},
					ListItems: recents,
					EmptyState: &types.EmptyStateData{
						Icon: "icon-alert-triangle", Title: l.RecentExceptions, Desc: l.NoExceptions,
					},
				},
			},
		}

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          l.Title,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      "fulfillment",
				ActiveSubNav:   "dashboard",
				HeaderTitle:    l.Title,
				HeaderSubtitle: l.Subtitle,
				HeaderIcon:     "icon-truck",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "fulfillment-dashboard-content",
			Dashboard:       dash,
		}
		return view.OK("fulfillment-dashboard", pageData)
	})
}

func buildExceptionsList(items []*fulfillmentpb.Fulfillment, l fayna.FulfillmentDashboardLabels) []types.ActivityItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]types.ActivityItem, 0, len(items))
	for i, f := range items {
		title := f.GetId()
		desc := f.GetStatus()
		if ps := f.GetProviderStatus(); ps != "" {
			desc = fmt.Sprintf("%s · %s", desc, ps)
		}
		when := ""
		if ms := f.GetDateModified(); ms > 0 {
			when = time.UnixMilli(ms).UTC().Format("2006-01-02 15:04")
		}
		out = append(out, types.ActivityItem{
			IconName:    "icon-alert-triangle",
			IconVariant: "client",
			Title:       title,
			Description: desc,
			Time:        when,
			TestID:      fmt.Sprintf("fulfillment-list-item-%d", i),
		})
	}
	_ = l
	return out
}
