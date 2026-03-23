package list

import (
	"context"
	"fmt"
	"log"

	fayna "github.com/erniealice/fayna-golang"
	lynguaV1 "github.com/erniealice/lyngua/golang/v1"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
)

// ListViewDeps holds view dependencies.
type ListViewDeps struct {
	Routes       fayna.JobRoutes
	ListJobs     func(ctx context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error)
	Labels       fayna.JobLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels
}

// PageData holds the data for the job list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the job list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}

		resp, err := deps.ListJobs(ctx, &jobpb.ListJobsRequest{})
		if err != nil {
			log.Printf("Failed to list jobs: %v", err)
			return view.Error(fmt.Errorf("failed to load jobs: %w", err))
		}

		l := deps.Labels
		columns := jobColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "jobs-table",
			Columns:              columns,
			Rows:                 rows,
			ShowSearch:           true,
			ShowActions:          true,
			ShowSort:             true,
			ShowColumns:          true,
			ShowDensity:          true,
			ShowEntries:          true,
			DefaultSortColumn:    "name",
			DefaultSortDirection: "asc",
			Labels:               deps.TableLabels,
			EmptyState: types.TableEmptyState{
				Title:   l.Empty.Title,
				Message: l.Empty.Message,
			},
			PrimaryAction: &types.PrimaryAction{
				Label:           l.Buttons.AddJob,
				ActionURL:       deps.Routes.AddURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("job", "create"),
				DisabledTooltip: l.Errors.PermissionDenied,
			},
		}
		types.ApplyTableSettings(tableConfig)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          statusPageTitle(l, status),
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   deps.Routes.ActiveSubNav,
				HeaderTitle:    statusPageTitle(l, status),
				HeaderSubtitle: statusPageCaption(l, status),
				HeaderIcon:     "icon-briefcase",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "job-list-content",
			Table:           tableConfig,
		}

		// KB help content
		if viewCtx.Translations != nil {
			if provider, ok := viewCtx.Translations.(*lynguaV1.TranslationProvider); ok {
				if kb, _ := provider.LoadKBIfExists(viewCtx.Lang, viewCtx.BusinessType, "jobs"); kb != nil {
					pageData.HasHelp = true
					pageData.HelpContent = kb.Body
				}
			}
		}

		return view.OK("job-list", pageData)
	})
}

func jobColumns(l fayna.JobLabels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "name", Label: l.Columns.Name, Sortable: true},
		{Key: "client", Label: l.Columns.Client, Sortable: true},
		{Key: "status", Label: l.Columns.Status, Sortable: true, Width: "130px"},
		{Key: "created", Label: l.Columns.Created, Sortable: true, Width: "150px"},
	}
}

func buildTableRows(jobs []*jobpb.Job, status string, l fayna.JobLabels, routes fayna.JobRoutes, perms *types.UserPermissions) []types.TableRow {
	rows := []types.TableRow{}
	for _, j := range jobs {
		jobStatus := jobStatusString(j.GetStatus())
		if jobStatus != status {
			continue
		}

		id := j.GetId()
		name := j.GetName()

		// Build client display name
		clientName := ""
		if c := j.GetClient(); c != nil {
			if u := c.GetUser(); u != nil {
				first := u.GetFirstName()
				last := u.GetLastName()
				if first != "" || last != "" {
					clientName = first + " " + last
				}
			}
			if clientName == "" {
				clientName = c.GetCompanyName()
			}
		}

		created := j.GetDateCreatedString()
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: name},
				{Type: "text", Value: clientName},
				{Type: "badge", Value: jobStatus, Variant: jobStatusVariant(jobStatus)},
				types.DateTimeCell(created, types.DateReadable),
			},
			DataAttrs: map[string]string{
				"name":    name,
				"client":  clientName,
				"status":  jobStatus,
				"created": created,
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("job", "update"), DisabledTooltip: l.Errors.PermissionDenied},
				{Type: "delete", Label: l.Actions.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: name, Disabled: !perms.Can("job", "delete"), DisabledTooltip: l.Errors.PermissionDenied},
			},
		})
	}
	return rows
}

func jobStatusString(s enums.JobStatus) string {
	switch s {
	case enums.JobStatus_JOB_STATUS_DRAFT:
		return "draft"
	case enums.JobStatus_JOB_STATUS_PENDING:
		return "pending"
	case enums.JobStatus_JOB_STATUS_ACTIVE:
		return "active"
	case enums.JobStatus_JOB_STATUS_PAUSED:
		return "paused"
	case enums.JobStatus_JOB_STATUS_COMPLETED:
		return "completed"
	case enums.JobStatus_JOB_STATUS_CLOSED:
		return "closed"
	default:
		return "draft"
	}
}

func jobStatusVariant(status string) string {
	switch status {
	case "draft":
		return "default"
	case "pending":
		return "warning"
	case "active":
		return "success"
	case "paused":
		return "warning"
	case "completed":
		return "info"
	case "closed":
		return "default"
	default:
		return "default"
	}
}

func statusPageTitle(l fayna.JobLabels, status string) string {
	switch status {
	case "draft":
		return l.Page.HeadingDraft
	case "active":
		return l.Page.HeadingActive
	case "completed":
		return l.Page.HeadingCompleted
	case "closed":
		return l.Page.HeadingClosed
	default:
		return l.Page.Heading
	}
}

func statusPageCaption(l fayna.JobLabels, status string) string {
	switch status {
	case "draft":
		return l.Page.CaptionDraft
	case "active":
		return l.Page.CaptionActive
	case "completed":
		return l.Page.CaptionCompleted
	case "closed":
		return l.Page.CaptionClosed
	default:
		return l.Page.Caption
	}
}
