package detail

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	job_category "github.com/erniealice/fayna-golang/domain/operation/job_category"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
)

// DetailViewDeps holds the job category detail view dependencies.
type DetailViewDeps struct {
	Routes          job_category.Routes
	Labels          job_category.Labels
	CommonLabels    pyeza.CommonLabels
	ReadJobCategory func(ctx context.Context, req *jobcategorypb.ReadJobCategoryRequest) (*jobcategorypb.ReadJobCategoryResponse, error)
}

// PageData holds the detail page data.
type PageData struct {
	types.PageData
	ContentTemplate string
	Labels          job_category.Labels
	EditURL         string

	Name         string
	Code         string
	SortOrder    string
	StatusLabel  string
	Active       bool
	DateCreated  string
	DateModified string
}

// NewView creates the job category detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_category", "list") {
			return view.Forbidden("job_category:list")
		}
		id := strings.TrimSpace(viewCtx.Request.PathValue("id"))
		if id == "" {
			return view.Forbidden("job_category:list")
		}
		resp, err := deps.ReadJobCategory(ctx, &jobcategorypb.ReadJobCategoryRequest{Data: &jobcategorypb.JobCategory{Id: id}})
		if err != nil || len(resp.GetData()) == 0 {
			return view.Error(errors.New(deps.Labels.Errors.NotFound))
		}
		c := resp.GetData()[0]

		l := deps.Labels
		code := c.GetCode()
		if code == "" {
			code = l.Detail.NoCode
		}
		order := l.Detail.NoCode
		if c.SortOrder != nil {
			order = strconv.Itoa(int(c.GetSortOrder()))
		}
		statusLabel := "Inactive"
		if c.GetActive() {
			statusLabel = "Active"
		}

		pd := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          c.GetName(),
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   deps.Routes.ActiveSubNav,
				HeaderTitle:    c.GetName(),
				HeaderSubtitle: l.Detail.Title,
				HeaderIcon:     "icon-layers",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "job-category-detail-content",
			Labels:          l,
			EditURL:         route.ResolveURL(deps.Routes.EditURL, "id", id),
			Name:            c.GetName(),
			Code:            code,
			SortOrder:       order,
			StatusLabel:     statusLabel,
			Active:          c.GetActive(),
			DateCreated:     millisToDate(c.DateCreated),
			DateModified:    millisToDate(c.DateModified),
		}
		return view.OK("job-category-detail", pd)
	})
}

// millisToDate formats an optional epoch-millis timestamp as YYYY-MM-DD, or "—".
func millisToDate(ms *int64) string {
	if ms == nil || *ms == 0 {
		return "—"
	}
	return time.UnixMilli(*ms).UTC().Format("2006-01-02")
}
