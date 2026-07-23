package list

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/erniealice/fayna-golang/domain/operation/job_category"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
)

// ListViewDeps holds the job category list view dependencies.
type ListViewDeps struct {
	Routes            job_category.Routes
	Labels            job_category.Labels
	CommonLabels      pyeza.CommonLabels
	TableLabels       types.TableLabels
	ListJobCategories func(ctx context.Context, req *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error)
}

// PageData holds the list page data.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the job category list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_category", "list") {
			return view.Forbidden("job_category:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}
		wantActive := status != "inactive"

		var items []*jobcategorypb.JobCategory
		if deps.ListJobCategories != nil {
			resp, err := deps.ListJobCategories(ctx, &jobcategorypb.ListJobCategoriesRequest{})
			if err != nil {
				log.Printf("Failed to list job categories: %v", err)
				return view.Error(fmt.Errorf("failed to load categories: %w", err))
			}
			items = resp.GetData()
		}

		l := deps.Labels
		rows := buildRows(items, wantActive, l, deps.Routes, perms)
		columns := columns(l)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:          "job-category-table",
			RefreshURL:  route.ResolveURL(deps.Routes.ListURL, "status", status),
			Columns:     columns,
			Rows:        rows,
			ShowSearch:  true,
			ShowActions: true,
			ShowSort:    true,
			ShowColumns: true,
			ShowDensity: true,
			ShowEntries: true,
			Labels:      deps.TableLabels,
			EmptyState: types.TableEmptyState{
				Title:   l.Empty.Title,
				Message: l.Empty.Message,
			},
			PrimaryAction: &types.PrimaryAction{
				Label:           l.Buttons.Add,
				ActionURL:       deps.Routes.AddURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("job_category", "create"),
				DisabledTooltip: l.Errors.Unauthorized,
			},
		}
		types.ApplyTableSettings(tableConfig)

		title := l.Page.ActiveTitle
		if !wantActive {
			title = l.Page.InactiveTitle
		}
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          title,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   deps.Routes.ActiveSubNav,
				HeaderTitle:    title,
				HeaderSubtitle: l.Page.Subtitle,
				HeaderIcon:     "icon-layers",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "job-category-list-content",
			Table:           tableConfig,
		}
		return view.OK("job-category-list", pageData)
	})
}

func columns(l job_category.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "name", Label: l.Columns.Name},
		{Key: "code", Label: l.Columns.Code, WidthClass: "col-3xl"},
		{Key: "sort_order", Label: l.Columns.SortOrder, WidthClass: "col-lg"},
		{Key: "status", Label: l.Columns.Status, WidthClass: "col-2xl"},
	}
}

func buildRows(
	items []*jobcategorypb.JobCategory,
	wantActive bool,
	l job_category.Labels,
	routes job_category.Routes,
	perms *types.UserPermissions,
) []types.TableRow {
	filtered := make([]*jobcategorypb.JobCategory, 0, len(items))
	for _, c := range items {
		if c.GetActive() != wantActive {
			continue
		}
		filtered = append(filtered, c)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		si, sj := filtered[i].GetSortOrder(), filtered[j].GetSortOrder()
		if si != sj {
			return si < sj
		}
		return filtered[i].GetName() < filtered[j].GetName()
	})

	rows := make([]types.TableRow, 0, len(filtered))
	for _, c := range filtered {
		id := c.GetId()
		name := c.GetName()
		code := c.GetCode()
		if code == "" {
			code = l.Detail.NoCode
		}
		order := ""
		if c.SortOrder != nil {
			order = strconv.Itoa(int(c.GetSortOrder()))
		}
		statusLabel, statusVariant := "Inactive", "warning"
		if c.GetActive() {
			statusLabel, statusVariant = "Active", "success"
		}
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: name},
				{Type: "text", Value: code},
				{Type: "text", Value: order},
				{Type: "badge", Value: statusLabel, Variant: statusVariant},
			},
			DataAttrs: map[string]string{"name": name, "code": code, "status": statusLabel},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Buttons.View, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Buttons.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Buttons.Edit, Disabled: !perms.Can("job_category", "update"), DisabledTooltip: l.Errors.Unauthorized},
				{Type: "delete", Label: l.Buttons.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: name, Disabled: !perms.Can("job_category", "delete"), DisabledTooltip: l.Errors.Unauthorized},
			},
		})
	}
	return rows
}
