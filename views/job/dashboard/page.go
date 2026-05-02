// Package dashboard implements the read-only Job live dashboard view
// (Phase 3 — Pyeza dashboard block + per-app live dashboards plan).
package dashboard

import (
	"context"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	fayna "github.com/erniealice/fayna-golang"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
)

// JobRiskRow mirrors espyna.dashboard.JobRisk (top-completion-risk widget).
type JobRiskRow struct {
	JobID         string
	Code          string
	Name          string
	CompletionPct float64
	DateEnd       time.Time
}

// Request mirrors the use-case request shape — kept narrow so the view
// package does not import espyna directly.
type Request struct {
	WorkspaceID string
	Now         time.Time
}

// Response mirrors the use-case response shape.
type Response struct {
	ActiveJobs        int64
	DoneThisMonth     int64
	OverdueJobs       int64
	HoursThisWeek     float64
	TrendLabels       []string
	TrendValues       []float64
	UpcomingDeadlines []*jobpb.Job
	RiskTopRows       []JobRiskRow
	RecentActivity    []*jobactivitypb.JobActivity
}

// Deps holds view dependencies.
type Deps struct {
	Routes               fayna.JobRoutes
	JobTemplateRoutes    fayna.JobTemplateRoutes
	JobActivityRoutes    fayna.JobActivityRoutes
	Labels               fayna.JobLabels
	CommonLabels         pyeza.CommonLabels
	GetDashboardPageData func(ctx context.Context, req *Request) (*Response, error)
}

// PageData is the dashboard template payload.
type PageData struct {
	types.PageData
	ContentTemplate string
	Dashboard       types.DashboardData
}

// NewView creates the job dashboard view.
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

		// 8-week hours-per-week chart.
		labels := resp.TrendLabels
		values := resp.TrendValues
		if len(labels) == 0 {
			labels = []string{"-7w", "-6w", "-5w", "-4w", "-3w", "-2w", "-1w", "Now"}
			values = []float64{0, 0, 0, 0, 0, 0, 0, 0}
		}
		trend := &types.ChartData{
			Labels: labels,
			Series: []types.ChartSeries{{
				Name:   l.HoursPerWeek,
				Values: values,
				Color:  "teal",
			}},
			YAxis: l.AxisHours,
		}
		trend.AutoScale()

		// Recent activity list.
		recent := buildRecentActivityList(resp.RecentActivity, l)

		dash := types.DashboardData{
			Title:    l.Title,
			Icon:     "icon-briefcase",
			Subtitle: l.Subtitle,
			QuickActions: []types.QuickAction{
				{Icon: "icon-plus", Label: l.QuickNewJob, Href: deps.Routes.AddURL, Variant: "primary", TestID: "job-action-new"},
				{Icon: "icon-file-plus", Label: l.QuickNewTemplate, Href: deps.JobTemplateRoutes.AddURL, TestID: "job-action-new-template"},
				{Icon: "icon-clock", Label: l.QuickLogHours, Href: deps.JobActivityRoutes.AddURL, TestID: "job-action-log-hours"},
				{Icon: "icon-calendar", Label: l.QuickJobCalendar, Href: deps.Routes.ListURL, TestID: "job-action-calendar"},
			},
			Stats: []types.StatCardData{
				{Icon: "icon-briefcase", Value: strconv.FormatInt(resp.ActiveJobs, 10), Label: l.StatActive, Color: "terracotta", TestID: "job-stat-active"},
				{Icon: "icon-check-circle", Value: strconv.FormatInt(resp.DoneThisMonth, 10), Label: l.StatDoneMonth, Color: "sage", TestID: "job-stat-done-month"},
				{Icon: "icon-alert-triangle", Value: strconv.FormatInt(resp.OverdueJobs, 10), Label: l.StatOverdue, Color: "amber", TestID: "job-stat-overdue"},
				{Icon: "icon-clock", Value: formatHours(resp.HoursThisWeek), Label: l.StatHoursWeek, Color: "navy", TestID: "job-stat-hours-week"},
			},
			Widgets: []types.DashboardWidget{
				{
					ID: "hours-per-week", Title: l.HoursPerWeek,
					Type: "chart", ChartKind: "bar",
					ChartData: trend, Span: 2,
				},
				{
					ID: "upcoming-deadlines", Title: l.UpcomingDeadlines, Type: "custom", Span: 2,
					Custom: buildUpcomingDeadlinesHTML(resp.UpcomingDeadlines, l, deps.Routes),
					EmptyState: &types.EmptyStateData{
						Icon: "icon-calendar", Title: l.UpcomingDeadlines, Desc: l.NoDeadlines,
					},
				},
				{
					ID: "recent-activity", Title: l.RecentActivity, Type: "list", Span: 1,
					HeaderActions: []types.QuickAction{
						{Label: l.ViewAll, Href: deps.JobActivityRoutes.ListURL},
					},
					ListItems: recent,
					EmptyState: &types.EmptyStateData{
						Icon: "icon-activity", Title: l.RecentActivity, Desc: l.NoActivity,
					},
				},
			},
		}

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          l.Title,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      "job",
				ActiveSubNav:   "dashboard",
				HeaderTitle:    l.Title,
				HeaderSubtitle: l.Subtitle,
				HeaderIcon:     "icon-briefcase",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "job-dashboard-content",
			Dashboard:       dash,
		}
		return view.OK("job-dashboard", pageData)
	})
}

func buildRecentActivityList(items []*jobactivitypb.JobActivity, l fayna.JobDashboardLabels) []types.ActivityItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]types.ActivityItem, 0, len(items))
	for i, a := range items {
		desc := ""
		if d := a.GetDescription(); d != "" {
			desc = d
		}
		entryDate := ""
		if s := a.GetEntryDateString(); s != "" {
			entryDate = s
		}
		out = append(out, types.ActivityItem{
			IconName:    "icon-clock",
			IconVariant: "client",
			Title:       fmt.Sprintf("%.2f h", a.GetQuantity()),
			Description: desc,
			Time:        entryDate,
			TestID:      fmt.Sprintf("job-list-item-%d", i),
		})
	}
	_ = l
	return out
}

// buildUpcomingDeadlinesHTML renders a small custom-HTML table for the
// upcoming deadlines widget. Per the plan: table widget with custom HTML
// (mirrors fycha loan-dashboard's TopLoans pattern).
func buildUpcomingDeadlinesHTML(jobs []*jobpb.Job, l fayna.JobDashboardLabels, routes fayna.JobRoutes) template.HTML {
	if len(jobs) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<table class="dashboard-mini-table"><thead><tr>`)
	b.WriteString(`<th>` + escapeHTML(l.UpcomingDeadlines) + `</th>`)
	b.WriteString(`<th class="numeric">` + escapeHTML(l.AxisHours) + `</th>`)
	b.WriteString(`</tr></thead><tbody>`)
	for _, j := range jobs {
		b.WriteString(`<tr data-testid="job-table-row-` + escapeHTML(j.GetId()) + `">`)
		b.WriteString(`<td>` + escapeHTML(j.GetName()) + `</td>`)
		b.WriteString(`<td class="numeric">`)
		if s := j.GetPlannedEndString(); s != "" {
			b.WriteString(escapeHTML(s))
		} else if j.GetPlannedEnd() > 0 {
			t := time.UnixMilli(j.GetPlannedEnd()).UTC()
			b.WriteString(escapeHTML(t.Format("2006-01-02")))
		}
		b.WriteString(`</td></tr>`)
	}
	b.WriteString(`</tbody></table>`)
	_ = routes
	return template.HTML(b.String()) //nolint:gosec
}

// escapeHTML is a small dependency-free HTML-escape helper.
func escapeHTML(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return r.Replace(s)
}

// formatHours formats a fractional hour count for stat-card display.
// 12.5 → "12.5h". 0 → "0h". 1.0 → "1h". Centavos rule does not apply (these
// are already hours, not centi-hours, by the time the view sees them — the
// ÷100 conversion happens in the use case).
func formatHours(h float64) string {
	if h == 0 {
		return "0h"
	}
	if h == float64(int64(h)) {
		return fmt.Sprintf("%dh", int64(h))
	}
	return fmt.Sprintf("%.1fh", h)
}
