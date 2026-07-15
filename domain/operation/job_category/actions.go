package job_category

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	jobcategoryform "github.com/erniealice/fayna-golang/domain/operation/job_category/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
)

// NewAddAction — GET = drawer form, POST = create.
func NewAddAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_category", "create") {
			return view.HTMXError(deps.Labels.Errors.Unauthorized)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("job-category-drawer-form", &jobcategoryform.Data{
				FormAction:   deps.Routes.AddURL,
				Active:       true,
				Labels:       deps.Labels,
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.CreateFailed)
		}
		r := viewCtx.Request
		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			return view.HTMXError(deps.Labels.Errors.CreateFailed)
		}
		status := "active"
		data := &jobcategorypb.JobCategory{
			Name:      name,
			Code:      strPtrIfNotEmpty(r.FormValue("code")),
			SortOrder: int32PtrIfSet(r.FormValue("sort_order")),
			Status:    &status,
			Active:    true,
		}
		if _, err := deps.CreateJobCategory(ctx, &jobcategorypb.CreateJobCategoryRequest{Data: data}); err != nil {
			log.Printf("Failed to create job category: %v", err)
			return view.HTMXError(err.Error())
		}
		return view.HTMXSuccess("job-category-table")
	})
}

// NewEditAction — GET = pre-filled drawer, POST = update.
func NewEditAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_category", "update") {
			return view.HTMXError(deps.Labels.Errors.Unauthorized)
		}
		id := viewCtx.Request.PathValue("id")
		if id == "" {
			id = viewCtx.Request.URL.Query().Get("id")
		}

		if viewCtx.Request.Method == http.MethodGet {
			if id == "" {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			resp, err := deps.ReadJobCategory(ctx, &jobcategorypb.ReadJobCategoryRequest{Data: &jobcategorypb.JobCategory{Id: id}})
			if err != nil || len(resp.GetData()) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			c := resp.GetData()[0]
			return view.OK("job-category-drawer-form", &jobcategoryform.Data{
				FormAction:   route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:       true,
				ID:           id,
				Name:         c.GetName(),
				Code:         c.GetCode(),
				SortOrder:    c.GetSortOrder(),
				Active:       c.GetActive(),
				Labels:       deps.Labels,
				CommonLabels: nil,
			})
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.UpdateFailed)
		}
		r := viewCtx.Request
		if id == "" {
			id = r.FormValue("id")
		}
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.NotFound)
		}
		data := &jobcategorypb.JobCategory{
			Id:        id,
			Name:      strings.TrimSpace(r.FormValue("name")),
			Code:      strPtrIfNotEmpty(r.FormValue("code")),
			SortOrder: int32PtrIfSet(r.FormValue("sort_order")),
			Active:    r.FormValue("active") == "true" || r.FormValue("active") == "on",
		}
		if _, err := deps.UpdateJobCategory(ctx, &jobcategorypb.UpdateJobCategoryRequest{Data: data}); err != nil {
			log.Printf("Failed to update job category %s: %v", id, err)
			return view.HTMXError(err.Error())
		}
		return view.ViewResult{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"HX-Trigger":  `{"formSuccess":true}`,
				"HX-Redirect": route.ResolveURL(deps.Routes.DetailURL, "id", id),
			},
		}
	})
}

// NewDeleteAction — POST only (id via ?id= or form).
func NewDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_category", "delete") {
			return view.HTMXError(deps.Labels.Errors.Unauthorized)
		}
		id := viewCtx.Request.URL.Query().Get("id")
		if id == "" {
			_ = viewCtx.Request.ParseForm()
			id = viewCtx.Request.FormValue("id")
		}
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.NotFound)
		}
		if _, err := deps.DeleteJobCategory(ctx, &jobcategorypb.DeleteJobCategoryRequest{Data: &jobcategorypb.JobCategory{Id: id}}); err != nil {
			log.Printf("Failed to delete job category %s: %v", id, err)
			return view.HTMXError(err.Error())
		}
		return view.HTMXSuccess("job-category-table")
	})
}

// NewBulkDeleteAction — POST only.
func NewBulkDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_category", "delete") {
			return view.HTMXError(deps.Labels.Errors.Unauthorized)
		}
		_ = viewCtx.Request.ParseMultipartForm(32 << 20)
		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return view.HTMXError("No IDs provided")
		}
		for _, id := range ids {
			if _, err := deps.DeleteJobCategory(ctx, &jobcategorypb.DeleteJobCategoryRequest{Data: &jobcategorypb.JobCategory{Id: id}}); err != nil {
				log.Printf("Failed to delete job category %s: %v", id, err)
			}
		}
		return view.HTMXSuccess("job-category-table")
	})
}

func strPtrIfNotEmpty(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func int32PtrIfSet(s string) *int32 {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	n := int32(v)
	return &n
}
