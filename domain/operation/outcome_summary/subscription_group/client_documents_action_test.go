package subscription_group

import (
	"bytes"
	"context"
	"html/template"
	"strings"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	clientdocumentbatch "github.com/erniealice/fayna-golang/domain/operation/outcome_summary/client_document_batch"

	bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/subscription_group_document_template"
	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

func clientDocumentsBatchDeps(finalAvailable bool) *Deps {
	deps := reportViewDeps()
	deps.Routes.ClientDocumentsDownloadDrawerURL = "/action/report-cards/section/{id}/download-all"
	deps.Routes.ClientDocumentsManifestURL = "/report-cards/section/{id}/download-all/manifest"
	deps.Options.SubscriptionGroupExport.WholeReportProfile = bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_CLIENT_PHASE_OUTCOME_REPORT_V1
	deps.ClientDocumentMounted = true
	category := reportCategory("cat-a", "Alpha", finalAvailable)
	deps.GetSubscriptionGroupOutcomeExport = func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		if req.GetOutcomeSelector() == nil {
			return reportOptions("sg-1", category), nil
		}
		return reportMatrix("sg-1", category, []*exportpb.JobTemplateColumn{{JobTemplateId: "tmpl-a", DisplayName: "Alpha"}}, []*exportpb.SubscriptionGroupOutcomeClientRow{
			reportClientRow("client-a", "", "Ava", "Adams", reportCell("tmpl-a", "A", nil, true, true)),
		}), nil
	}
	return deps
}

func TestReportView_ClientDocumentsBatchActionIsTableToolbarSibling(t *testing.T) {
	deps := clientDocumentsBatchDeps(true)
	deps.Routes.SubscriptionGroupDownloadDrawerURL = "/action/report-cards/section/{id}/download"
	page := reportPageData(t, NewView(deps).Handle(reportContext(t, deps.Routes, "sg-1", ""), reportViewContext(t, deps.Routes, "sg-1", "")))
	if page.Table == nil || page.ClientDocumentsAction == nil {
		t.Fatalf("computed page table=%v action=%v, want both", page.Table != nil, page.ClientDocumentsAction)
	}
	if page.Table.ToolbarPrefixTemplate != clientdocumentbatch.OpenActionTemplate {
		t.Fatalf("toolbar prefix template = %q", page.Table.ToolbarPrefixTemplate)
	}
	action, ok := page.Table.ToolbarPrefixData.(*clientdocumentbatch.OpenAction)
	if !ok || action.ActionURL != "/action/report-cards/section/sg-1/download-all" {
		t.Fatalf("toolbar prefix data = %#v", page.Table.ToolbarPrefixData)
	}
	if page.Table.PrimaryAction == nil {
		t.Fatal("the category-pinned group download (PrimaryAction) must stay alongside the batch action")
	}
}

func TestReportView_ClientDocumentsBatchActionOnNotComputedPage(t *testing.T) {
	deps := clientDocumentsBatchDeps(false)
	page := reportPageData(t, NewView(deps).Handle(reportContext(t, deps.Routes, "sg-1", ""), reportViewContext(t, deps.Routes, "sg-1", "")))
	if !page.NotComputed || page.ClientDocumentsAction == nil {
		t.Fatalf("not-computed page action = %v, want the batch action on the page toolbar", page.ClientDocumentsAction)
	}
}

func TestReportView_ClientDocumentsBatchActionHiddenWithoutRoutesOrMount(t *testing.T) {
	for name, mutate := range map[string]func(*Deps){
		"no drawer route":        func(d *Deps) { d.Routes.ClientDocumentsDownloadDrawerURL = "" },
		"no manifest route":      func(d *Deps) { d.Routes.ClientDocumentsManifestURL = "" },
		"client route unmounted": func(d *Deps) { d.ClientDocumentMounted = false },
		"non-staff principal":    func(d *Deps) { d.ResolvePrincipalKind = func(context.Context) int32 { return 0 } },
	} {
		t.Run(name, func(t *testing.T) {
			deps := clientDocumentsBatchDeps(true)
			mutate(deps)
			result := NewView(deps).Handle(reportContext(t, deps.Routes, "sg-1", ""), reportViewContext(t, deps.Routes, "sg-1", ""))
			page, ok := result.Data.(*PageData)
			if !ok {
				return // forbidden page: no action can render
			}
			if page.ClientDocumentsAction != nil || (page.Table != nil && page.Table.ToolbarPrefixTemplate != "") {
				t.Fatalf("batch action offered: %+v", page.ClientDocumentsAction)
			}
		})
	}
}

func TestSubscriptionGroupTemplate_NotComputedToolbarRendersBatchAction(t *testing.T) {
	tmpl := template.New("page").Funcs(template.FuncMap{"dict": func(...any) map[string]any { return nil }})
	if _, err := tmpl.Parse(`{{define "tabs"}}{{end}}{{define "alert"}}{{end}}{{define "table-card"}}{{end}}{{define "app-shell"}}{{end}}`); err != nil {
		t.Fatal(err)
	}
	if _, err := tmpl.ParseFiles("../templates/subscription_group.html", "../templates/client-documents-download-drawer.html"); err != nil {
		t.Fatalf("parse templates: %v", err)
	}
	labels := outcome_summary.DefaultLabels().ClientDocumentsDownload
	var out bytes.Buffer
	err := tmpl.ExecuteTemplate(&out, "outcome-summary-subscription-group-content", &PageData{
		NotComputed: true,
		ClientDocumentsAction: &clientdocumentbatch.OpenAction{
			ActionURL: "/action/report-cards/section/sg-1/download-all",
			Label:     labels.OpenAction,
			Title:     labels.DrawerTitle,
		},
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	rendered := out.String()
	if !strings.Contains(rendered, `data-testid="rc-client-documents-download-open"`) || !strings.Contains(rendered, `class="page-toolbar"`) {
		t.Fatalf("not-computed toolbar missing the batch action:\n%s", rendered)
	}
	if strings.Contains(rendered, `rc-subscription-group-download-open`) {
		t.Fatalf("group download button rendered without a drawer URL:\n%s", rendered)
	}
}
