package client_document_batch

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

// DrawerTemplate is the HTMX-loaded Period/Format form.
const DrawerTemplate = "outcome-summary-client-documents-download-drawer-form"

// OpenActionTemplate renders OpenAction (table toolbar prefix / page toolbar).
const OpenActionTemplate = "outcome-summary-client-documents-download-open"

// DrawerData is the template-facing state. ManifestURL ends in URL so the view
// adapter's RewriteNavURLs pass prefixes /w/{slug}; the browser module fetches
// it with the chosen period and format.
type DrawerData struct {
	ManifestURL  string
	Periods      []types.SelectOption
	Formats      []types.SelectOption
	Labels       outcome_summary.ClientDocumentsDownloadLabels
	OptionLabels outcome_summary.ClientDocumentDownloadLabels
	CommonLabels any
	CacheVersion string
	Concurrency  int
	Nonce        string
}

// NewDrawer creates the subscription-group batch Period/Format drawer. It gates
// on the per-client document capability before any read, then derives the
// period choices from the scoped export options (the same reportScope as the
// per-client document projection).
func NewDrawer(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		if deps == nil || !outcome_summary.CanExplicitExport(view.GetUserPermissions(ctx), ctx, deps.ResolvePrincipalKind) {
			return view.Forbidden(outcome_summary.ExportPermissionEntity + ":" + outcome_summary.ExportReadAction)
		}
		if !Configured(deps.Options, deps.Routes) || deps.GetSubscriptionGroupOutcomeExport == nil {
			return view.Error(fmt.Errorf("client document batch download is not configured"))
		}
		if viewCtx == nil || viewCtx.Request == nil || viewCtx.Request.URL == nil {
			return view.Error(fmt.Errorf("client document batch download request is missing"))
		}
		groupID := strings.TrimSpace(viewCtx.Request.PathValue("id"))
		if groupID == "" {
			return view.ViewResult{Error: fmt.Errorf("subscription group id is required"), StatusCode: http.StatusBadRequest}
		}

		options, err := deps.GetSubscriptionGroupOutcomeExport(ctx, &exportpb.GetSubscriptionGroupOutcomeExportRequest{
			SubscriptionGroupId: groupID,
		})
		if err != nil || !contextMatches(options, groupID) {
			if err != nil {
				log.Printf("client document batch drawer: options: %v", err)
			}
			// Fail closed exactly like the hosting report view (report_view.go).
			return view.Forbidden(outcome_summary.ExportPermissionEntity + ":" + outcome_summary.ExportReadAction)
		}
		periods := groupPeriods(deps.Labels, options.GetJobCategories())
		if len(periods) == 0 {
			return view.ViewResult{Error: fmt.Errorf("subscription group has no downloadable outcomes"), StatusCode: http.StatusNotFound}
		}
		formats := formatOptions(deps)
		if len(formats) == 0 {
			return view.Error(fmt.Errorf("client document batch download has no available format"))
		}
		cacheVersion := ""
		if viewCtx != nil {
			cacheVersion = viewCtx.CacheVersion
		}
		return view.OK(DrawerTemplate, &DrawerData{
			ManifestURL:  route.ResolveURL(deps.Routes.ClientDocumentsManifestURL, "id", groupID),
			Periods:      periodSelectOptions(periods),
			Formats:      formats,
			Labels:       deps.Labels.ClientDocumentsDownload,
			OptionLabels: deps.Labels.ClientDocumentDownload,
			CommonLabels: deps.CommonLabels,
			CacheVersion: cacheVersion,
			Concurrency:  concurrencyFor(deps.Options.Document.BatchConcurrency),
		})
	})
}

func periodSelectOptions(periods []periodOption) []types.SelectOption {
	options := make([]types.SelectOption, 0, len(periods))
	for index, period := range periods {
		options = append(options, types.SelectOption{Value: period.Code, Label: period.Label, Selected: index == 0})
	}
	return options
}

// formatOptions mirrors client_card.clientReportFormatOptions (PDF first and
// selected) and drops a format whose render closure is not wired.
func formatOptions(deps *Deps) []types.SelectOption {
	l := deps.Labels.ClientDocumentDownload
	options := make([]types.SelectOption, 0, 2)
	if deps.PDFAvailable {
		options = append(options, types.SelectOption{Value: FormatPDF, Label: firstNonEmpty(l.FormatPdf, "PDF")})
	}
	if deps.DOCXAvailable {
		options = append(options, types.SelectOption{Value: FormatDOCX, Label: firstNonEmpty(l.FormatDocx, "DOCX")})
	}
	if len(options) > 0 {
		options[0].Selected = true
	}
	return options
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
