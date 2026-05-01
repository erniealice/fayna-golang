package detail

import (
	"context"
	"fmt"
	"log"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/hybra-golang/views/attachment"
	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
)

// PageData holds the data for the job template detail page.
type PageData struct {
	types.PageData
	ContentTemplate     string
	JobTemplate         map[string]any
	Labels              fayna.JobTemplateLabels
	ActiveTab           string
	TabItems            []pyeza.TabItem
	PhasesTable         *types.TableConfig
	AttachmentTable     *types.TableConfig
	AttachmentUploadURL string
	// Audit history tab
	AuditEntries    []auditlog.AuditEntryView
	AuditHasNext    bool
	AuditNextCursor string
	AuditHistoryURL string
}

// jobTemplateToMap converts a JobTemplate protobuf to a map[string]any for template use.
func jobTemplateToMap(t *jobtemplatepb.JobTemplate) map[string]any {
	recordStatus := "active"
	if !t.GetActive() {
		recordStatus = "inactive"
	}
	statusVariant := "success"
	if recordStatus == "inactive" {
		statusVariant = "warning"
	}

	return map[string]any{
		"id":                   t.GetId(),
		"name":                 t.GetName(),
		"description":          t.GetDescription(),
		"active":               t.GetActive(),
		"status":               recordStatus,
		"status_variant":       statusVariant,
		"date_created_string":  t.GetDateCreatedString(),
		"date_modified_string": t.GetDateModifiedString(),
	}
}

// NewView creates the job template detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		id := viewCtx.Request.PathValue("id")

		resp, err := deps.ReadJobTemplate(ctx, &jobtemplatepb.ReadJobTemplateRequest{
			Data: &jobtemplatepb.JobTemplate{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read job template %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load job template: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Job template %s not found", id)
			return view.Error(fmt.Errorf("job template not found"))
		}
		tmpl := jobTemplateToMap(data[0])

		name, _ := tmpl["name"].(string)

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
				HeaderIcon:     "icon-clipboard",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "job-template-detail-content",
			JobTemplate:     tmpl,
			Labels:          l,
			ActiveTab:       activeTab,
			TabItems:        tabItems,
		}

		loadTabData(ctx, deps, pageData, id, activeTab, viewCtx)

		return view.OK("job-template-detail", pageData)
	})
}

func buildTabItems(l fayna.JobTemplateLabels, id string, routes fayna.JobTemplateRoutes) []pyeza.TabItem {
	base := route.ResolveURL(routes.DetailURL, "id", id)
	action := route.ResolveURL(routes.TabActionURL, "id", id, "tab", "")
	return []pyeza.TabItem{
		{Key: "info", Label: l.Tabs.Info, Href: base + "?tab=info", HxGet: action + "info", Icon: "icon-info"},
		{Key: "phases", Label: l.Tabs.Phases, Href: base + "?tab=phases", HxGet: action + "phases", Icon: "icon-list"},
		{Key: "attachments", Label: l.Tabs.Attachments, Href: base + "?tab=attachments", HxGet: action + "attachments", Icon: "icon-paperclip"},
		{Key: "audit-history", Label: "History", Href: base + "?tab=audit-history", HxGet: action + "audit-history", Icon: "icon-clock"},
	}
}

func loadTabData(ctx context.Context, deps *DetailViewDeps, pageData *PageData, id string, activeTab string, viewCtx *view.ViewContext) {
	switch activeTab {
	case "info":
		// all info fields available from jobTemplateToMap
	case "phases":
		loadPhasesTab(ctx, deps, pageData, id)
	case "attachments":
		if deps.ListAttachments != nil {
			cfg := attachmentConfig(deps)
			resp, err := deps.ListAttachments(ctx, cfg.EntityType, id)
			if err != nil {
				log.Printf("Failed to list attachments for job template %s: %v", id, err)
			}
			var items []*attachmentpb.Attachment
			if resp != nil {
				items = resp.GetData()
			}
			pageData.AttachmentTable = attachment.BuildTable(items, cfg, id)
		}
		pageData.AttachmentUploadURL = route.ResolveURL(deps.Routes.AttachmentUploadURL, "id", id)
	case "audit-history":
		if deps.ListAuditHistory != nil {
			cursor := viewCtx.QueryParams["cursor"]
			auditResp, err := deps.ListAuditHistory(ctx, &auditlog.ListAuditRequest{
				EntityType:  "job_template",
				EntityID:    id,
				Limit:       20,
				CursorToken: cursor,
			})
			if err != nil {
				log.Printf("Failed to load audit history: %v", err)
			}
			if auditResp != nil {
				pageData.AuditEntries = auditResp.Entries
				pageData.AuditHasNext = auditResp.HasNext
				pageData.AuditNextCursor = auditResp.NextCursor
			}
		}
		pageData.AuditHistoryURL = route.ResolveURL(deps.Routes.TabActionURL, "id", id, "tab", "") + "audit-history"
	}
}

func loadPhasesTab(ctx context.Context, deps *DetailViewDeps, pageData *PageData, id string) {
	if deps.ListPhasesByJobTemplate == nil {
		return
	}
	resp, err := deps.ListPhasesByJobTemplate(ctx, &jobtemplatephasepb.ListByJobTemplateRequest{
		JobTemplateId: id,
	})
	if err != nil {
		log.Printf("Failed to list phases for job template %s: %v", id, err)
		return
	}
	phases := resp.GetJobTemplatePhases()
	rows := []types.TableRow{}
	for _, p := range phases {
		rows = append(rows, types.TableRow{
			ID: p.GetId(),
			Cells: []types.TableCell{
				{Type: "text", Value: fmt.Sprintf("%d", p.GetPhaseOrder())},
				{Type: "text", Value: p.GetName()},
			},
			DataAttrs: map[string]string{
				"name":  p.GetName(),
				"order": fmt.Sprintf("%d", p.GetPhaseOrder()),
			},
		})
	}
	pageData.PhasesTable = &types.TableConfig{
		ID: "job-template-phases-table",
		Columns: []types.TableColumn{
			{Key: "order", Label: "#", WidthClass: "col-sm"},
			{Key: "name", Label: deps.Labels.Columns.Name},
		},
		Rows:        rows,
		Labels:      deps.TableLabels,
		ShowSearch:  false,
		ShowActions: false,
		ShowSort:    false,
		ShowColumns: false,
		ShowDensity: false,
		ShowEntries: false,
	}
	types.ApplyColumnStyles(pageData.PhasesTable.Columns, pageData.PhasesTable.Rows)
}

// NewTabAction creates the tab action view (partial — returns only the tab content).
func NewTabAction(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		id := viewCtx.Request.PathValue("id")
		tab := viewCtx.Request.PathValue("tab")
		if tab == "" {
			tab = "info"
		}

		resp, err := deps.ReadJobTemplate(ctx, &jobtemplatepb.ReadJobTemplateRequest{
			Data: &jobtemplatepb.JobTemplate{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read job template %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load job template: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Job template %s not found", id)
			return view.Error(fmt.Errorf("job template not found"))
		}
		tmpl := jobTemplateToMap(data[0])

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion: viewCtx.CacheVersion,
				CommonLabels: deps.CommonLabels,
			},
			JobTemplate: tmpl,
			Labels:      l,
			ActiveTab:   tab,
			TabItems:    buildTabItems(l, id, deps.Routes),
		}

		loadTabData(ctx, deps, pageData, id, tab, viewCtx)

		templateName := "jt-tab-" + tab
		if tab == "attachments" {
			templateName = "attachment-tab"
		}
		if tab == "audit-history" {
			templateName = "audit-history-tab"
		}
		return view.OK(templateName, pageData)
	})
}
