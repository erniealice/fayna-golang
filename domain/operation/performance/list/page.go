package list

import (
	"context"
	"fmt"
	"log"
	"strconv"

	performance "github.com/erniealice/fayna-golang/domain/operation/performance"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"
)

// ListViewDeps is the page-package alias of the entity-package deps (mirrors the
// other operation list packages so the block wires one shape).
type ListViewDeps = performance.ListViewDeps

// PageData holds the data for the performance admin panel page (Surface 6).
type PageData struct {
	types.PageData
	ContentTemplate string
	Labels          performance.Labels
	Groups          []GroupSection
	Banner          *BannerView
	HasRows         bool
}

// GroupSection is one matrix bucket (Overdue / Due / Up-to-date) rendered as a
// <section aria-labelledby> per HTML-1.
type GroupSection struct {
	Key      string // testid suffix: perf-group-{key}
	TestID   string // perf-group-{key}
	HeadID   string // aria-labelledby target (heading id)
	Heading  string
	Rows     []RowView
	HasRows  bool
}

// RowView is one associate row inside a group section.
type RowView struct {
	SeatID            string // perf-row-{sr_id}
	RowTestID         string // perf-row-{sr_id}
	StaffID           string
	AssociateName     string
	ClientName        string
	RatingText        string // formatted score or "Not yet rated"
	RatingTestID      string // associate-rating-{staff_id}
	HasRating         bool
	StartReviewURL    string // /action/evaluation/add?seat_id={sr_id}
	StartReviewTestID string // perf-start-review-{sr_id}
	ViewReviewURL     string // /app/evaluations/detail/{id}  ("" → hide)
	ViewReviewTestID  string // perf-view-review-{sr_id}
}

// BannerView is the rendered "X of Y" cycle-progress banner (§F.2).
type BannerView struct {
	CycleID       string // evaluation-cycle-banner-{cycle_id}
	BannerTestID  string // evaluation-cycle-banner-{cycle_id}
	ProgressID    string // cycle-progress-{cycle_id}
	Name          string
	ProgressText  string // "X of Y complete"
	SignOffText   string // "sign-offs due ..."  ("" → hide)
	CloseText     string // "closes ..."          ("" → hide)
}

// groupOrder is the fixed presentation order of the matrix sections.
var groupOrder = []performance.GroupKey{
	performance.GroupOverdue,
	performance.GroupDue,
	performance.GroupUpToDate,
}

// NewView creates the performance admin panel view (composition surface).
//
// Auth: L3 view.Forbidden(evaluation:dashboard) at the top (pages.md §G.1). The
// CR-5 servicing gate is enforced inside the block-supplied GetPanelData closure
// (espyna layer) — the view never sees out-of-scope seats and supplies no
// client_id/subscription_id.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation", "dashboard") {
			return view.Forbidden("evaluation:dashboard")
		}

		l := deps.Labels

		var data *performance.PanelData
		if deps.GetPanelData != nil {
			d, err := deps.GetPanelData(ctx)
			if err != nil {
				log.Printf("Failed to load performance panel: %v", err)
				return view.Error(fmt.Errorf("%s: %w", l.Errors.LoadFailed, err))
			}
			data = d
		}
		if data == nil {
			data = &performance.PanelData{}
		}

		groups, hasRows := buildGroups(data.Rows, l, deps.Routes)
		banner := buildBanner(data.Banner, l)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          l.Page.Heading,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   deps.Routes.ActiveSubNav,
				HeaderTitle:    l.Page.Heading,
				HeaderSubtitle: l.Page.Caption,
				HeaderIcon:     "icon-bar-chart",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "performance-panel-content",
			Labels:          l,
			Groups:          groups,
			Banner:          banner,
			HasRows:         hasRows,
		}

		return view.OK("performance-panel", pageData)
	})
}

func buildGroups(rows []performance.PanelRow, l performance.Labels, routes performance.Routes) ([]GroupSection, bool) {
	byGroup := map[performance.GroupKey][]RowView{}
	hasRows := false
	for _, r := range rows {
		g := r.Group
		if g == "" {
			g = performance.GroupUpToDate
		}
		byGroup[g] = append(byGroup[g], buildRow(r, l, routes))
		hasRows = true
	}

	sections := make([]GroupSection, 0, len(groupOrder))
	for _, g := range groupOrder {
		key := string(g)
		gr := byGroup[g]
		sections = append(sections, GroupSection{
			Key:     key,
			TestID:  "perf-group-" + key,
			HeadID:  "perf-group-head-" + key,
			Heading: groupHeading(l, g),
			Rows:    gr,
			HasRows: len(gr) > 0,
		})
	}
	return sections, hasRows
}

func buildRow(r performance.PanelRow, l performance.Labels, routes performance.Routes) RowView {
	ratingText := l.Rating.None
	hasRating := false
	if r.LatestRating != nil {
		ratingText = strconv.FormatFloat(*r.LatestRating, 'f', 1, 64)
		hasRating = true
	}

	startURL := routes.EvaluationAddURL
	if r.SeatID != "" {
		startURL = routes.EvaluationAddURL + "?seat_id=" + r.SeatID
	}

	viewURL := ""
	if r.LatestEvalID != "" {
		viewURL = route.ResolveURL(routes.EvaluationDetailURL, "id", r.LatestEvalID)
	}

	return RowView{
		SeatID:            r.SeatID,
		RowTestID:         "perf-row-" + r.SeatID,
		StaffID:           r.StaffID,
		AssociateName:     r.AssociateName,
		ClientName:        r.ClientName,
		RatingText:        ratingText,
		RatingTestID:      "associate-rating-" + r.StaffID,
		HasRating:         hasRating,
		StartReviewURL:    startURL,
		StartReviewTestID: "perf-start-review-" + r.SeatID,
		ViewReviewURL:     viewURL,
		ViewReviewTestID:  "perf-view-review-" + r.SeatID,
	}
}

func buildBanner(b *performance.CycleBanner, l performance.Labels) *BannerView {
	if b == nil {
		return nil
	}
	bv := &BannerView{
		CycleID:      b.CycleID,
		BannerTestID: "evaluation-cycle-banner-" + b.CycleID,
		ProgressID:   "cycle-progress-" + b.CycleID,
		Name:         b.Name,
		ProgressText: fmt.Sprintf(l.Banner.Progress, b.Completed, b.Total),
	}
	if b.SignOffDueLabel != "" {
		bv.SignOffText = fmt.Sprintf(l.Banner.SignOffsDue, b.SignOffDueLabel)
	}
	if b.CloseLabel != "" {
		bv.CloseText = fmt.Sprintf(l.Banner.Closes, b.CloseLabel)
	}
	return bv
}

func groupHeading(l performance.Labels, g performance.GroupKey) string {
	switch g {
	case performance.GroupOverdue:
		return l.Groups.Overdue
	case performance.GroupDue:
		return l.Groups.Due
	default:
		return l.Groups.UpToDate
	}
}
