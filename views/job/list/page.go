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
	GetInUseIDs  func(ctx context.Context, ids []string) (map[string]bool, error)
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
		// 2026-05-14 permission-gates P2a: reject direct-URL access.
		if !perms.Can("job", "list") {
			return view.Forbidden("job:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}

		resp, err := deps.ListJobs(ctx, &jobpb.ListJobsRequest{})
		if err != nil {
			log.Printf("Failed to list jobs: %v", err)
			return view.Error(fmt.Errorf("failed to load jobs: %w", err))
		}

		// Collect IDs and check which are in use (referenced by dependent tables).
		var inUseIDs map[string]bool
		if deps.GetInUseIDs != nil {
			var itemIDs []string
			for _, j := range resp.GetData() {
				itemIDs = append(itemIDs, j.GetId())
			}
			inUseIDs, _ = deps.GetInUseIDs(ctx, itemIDs)
		}

		l := deps.Labels
		columns := jobColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, inUseIDs, perms)
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
			BulkActions: &types.BulkActionsConfig{
				Enabled:        true,
				SelectAllLabel: l.BulkActions.SelectAll,
				SelectedLabel:  l.BulkActions.SelectedCount,
				CancelLabel:    l.BulkActions.Cancel,
				Actions: []types.BulkAction{
					{
						Key:              "delete",
						Label:            l.BulkActions.Delete,
						Icon:             "icon-trash-2",
						Variant:          "danger",
						Endpoint:         deps.Routes.BulkDeleteURL,
						ConfirmTitle:     l.BulkActions.BulkDeleteConfirmTitle,
						ConfirmMessage:   l.BulkActions.BulkDeleteConfirmMsg,
						RequiresDataAttr: "deletable",
					},
				},
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
				if kb, _ := provider.LoadKBIfExists(viewCtx.Lang, viewCtx.BusinessType, "job"); kb != nil {
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
		{Key: "name", Label: l.Columns.Name},
		{Key: "client", Label: l.Columns.Client},
		{Key: "status", Label: l.Columns.Status, WidthClass: "col-3xl"},
		{Key: "created", Label: l.Columns.Created, WidthClass: "col-4xl"},
	}
}

func buildTableRows(jobs []*jobpb.Job, status string, l fayna.JobLabels, routes fayna.JobRoutes, inUseIDs map[string]bool, perms *types.UserPermissions) []types.TableRow {
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
				clientName = c.GetName()
			}
		}

		created := j.GetDateCreatedString()
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)

		inUse := inUseIDs[id]
		deleteDisabled := inUse || !perms.Can("job", "delete")
		deleteTooltip := l.Errors.PermissionDenied
		if inUse {
			deleteTooltip = l.Errors.InUse
		}

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
				"name":      name,
				"client":    clientName,
				"status":    jobStatus,
				"created":   created,
				"deletable": boolAttr(!inUse),
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("job", "update"), DisabledTooltip: l.Errors.PermissionDenied},
				{Type: "delete", Label: l.Actions.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: name, Disabled: deleteDisabled, DisabledTooltip: deleteTooltip},
			},
		})
	}
	return rows
}

func boolAttr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func jobStatusString(s enums.JobStatus) string {
	switch s {
	case enums.JobStatus_JOB_STATUS_DRAFT:
		return "draft"
	case enums.JobStatus_JOB_STATUS_PENDING:
		return "pending"
	case enums.JobStatus_JOB_STATUS_PLANNED:
		return "planned"
	case enums.JobStatus_JOB_STATUS_RELEASED:
		return "released"
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
	case "planned":
		return "secondary"
	case "released":
		return "success"
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
	case "planned":
		return l.Page.HeadingPlanned
	case "released":
		return l.Page.HeadingReleased
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
	case "planned":
		return l.Page.CaptionPlanned
	case "released":
		return l.Page.CaptionReleased
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
