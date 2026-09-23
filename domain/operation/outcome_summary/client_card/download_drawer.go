package client_card

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	subscriptiongroupdocument "github.com/erniealice/fayna-golang/domain/operation/outcome_summary/subscription_group_document"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

const (
	clientReportYearFinalPeriod = "year_final"
)

// DrawerDeps is the narrow composition seam for the single-client download
// drawer. The typed report-card query is the only source of membership and
// available phase evidence; no adapter, storage, or template resolver is
// reachable from this view.
type DrawerDeps struct {
	Routes               outcome_summary.Routes
	Labels               outcome_summary.Labels
	CommonLabels         any
	Options              outcome_summary.Options
	ResolvePrincipalKind func(context.Context) int32
	ClientAttributeCodes []string

	GetSubscriptionGroupClientReportCard func(context.Context, *exportpb.GetSubscriptionGroupClientReportCardRequest) (*exportpb.GetSubscriptionGroupClientReportCardResponse, error)
}

// DownloadDrawerData is the template-facing state for the HTMX-loaded partial.
type DownloadDrawerData struct {
	FormURL      string
	Periods      []types.SelectOption
	Formats      []types.SelectOption
	Labels       outcome_summary.ClientDocumentDownloadLabels
	CommonLabels any
	Nonce        string
}

// NewDownloadDrawer creates a per-client Period/Format selector. It verifies
// the explicit export capability before invoking the typed use case, whose
// scoped query validates the exact group/client membership and tenant.
func NewDownloadDrawer(deps *DrawerDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		if deps == nil || !outcome_summary.CanExplicitExport(view.GetUserPermissions(ctx), ctx, deps.ResolvePrincipalKind) {
			return view.Forbidden(outcome_summary.ExportPermissionEntity + ":" + outcome_summary.ExportReadAction)
		}
		profile, profileRegistered := subscriptiongroupdocument.LookupProfile(deps.Options.SubscriptionGroupExport.WholeReportProfile)
		if !deps.Options.SubscriptionGroupExportEnabled() || !profileRegistered ||
			profile.CategoryBindingScope != subscriptiongroupdocument.CategoryBindingScopeAllCategories ||
			strings.TrimSpace(deps.Routes.ClientDownloadDrawerURL) == "" ||
			strings.TrimSpace(deps.Routes.ClientDocumentURL) == "" ||
			deps.GetSubscriptionGroupClientReportCard == nil {
			return view.Error(fmt.Errorf("client report download is not configured"))
		}
		if viewCtx == nil || viewCtx.Request == nil || viewCtx.Request.URL == nil {
			return view.Error(fmt.Errorf("client report download request is missing"))
		}
		groupID := strings.TrimSpace(viewCtx.Request.PathValue("id"))
		clientID := strings.TrimSpace(viewCtx.Request.PathValue("client_id"))
		if groupID == "" || clientID == "" {
			return view.ViewResult{Error: fmt.Errorf("subscription group and client ids are required"), StatusCode: http.StatusBadRequest}
		}

		response, err := deps.GetSubscriptionGroupClientReportCard(ctx, &exportpb.GetSubscriptionGroupClientReportCardRequest{
			SubscriptionGroupId:  groupID,
			ClientId:             clientID,
			ClientAttributeCodes: append([]string(nil), deps.ClientAttributeCodes...),
		})
		if err != nil {
			return view.Error(fmt.Errorf("load client report download options: %w", err))
		}
		if response == nil || !response.GetSuccess() || response.GetReportCard() == nil {
			return view.ViewResult{Error: fmt.Errorf("client report card not found"), StatusCode: http.StatusNotFound}
		}
		projection := response.GetReportCard()
		if projection.GetContext() == nil || projection.GetContext().GetSubscriptionGroupId() != groupID ||
			projection.GetClient() == nil || projection.GetClient().GetClientId() != clientID ||
			len(projection.GetClientSubscriptionIds()) == 0 {
			return view.ViewResult{Error: fmt.Errorf("client report card not found"), StatusCode: http.StatusNotFound}
		}

		periods := clientReportPeriodOptions(deps.Labels, projection)
		if len(periods) == 0 {
			return view.ViewResult{Error: fmt.Errorf("client report card has no downloadable outcomes"), StatusCode: http.StatusNotFound}
		}
		return view.OK("outcome-summary-client-download-drawer-form", &DownloadDrawerData{
			FormURL:      route.ResolveURL(deps.Routes.ClientDocumentURL, "id", groupID, "client_id", clientID),
			Periods:      periods,
			Formats:      clientReportFormatOptions(deps.Labels),
			Labels:       deps.Labels.ClientDocumentDownload,
			CommonLabels: deps.CommonLabels,
		})
	})
}

func clientReportPeriodOptions(labels outcome_summary.Labels, projection *exportpb.ClientReportCardProjection) []types.SelectOption {
	if projection == nil {
		return nil
	}
	projectedJobs := make(map[string]struct{}, len(projection.GetJobs()))
	for _, job := range projection.GetJobs() {
		if job != nil && strings.TrimSpace(job.GetId()) != "" {
			projectedJobs[job.GetId()] = struct{}{}
		}
	}
	referencedTemplatePhaseIDs := make(map[string]struct{})
	for _, phase := range projection.GetJobPhases() {
		if phase == nil || !phase.GetActive() {
			continue
		}
		if _, belongsToClient := projectedJobs[phase.GetJobId()]; !belongsToClient {
			continue
		}
		if templatePhaseID := strings.TrimSpace(phase.GetTemplatePhaseId()); templatePhaseID != "" {
			referencedTemplatePhaseIDs[templatePhaseID] = struct{}{}
		}
	}
	entries := make([]outcome_summary.ClientReportPhaseEntry, 0, len(referencedTemplatePhaseIDs))
	for _, phase := range projection.GetJobTemplatePhases() {
		if phase == nil || !phase.GetActive() {
			continue
		}
		phaseID := strings.TrimSpace(phase.GetId())
		if _, referenced := referencedTemplatePhaseIDs[phaseID]; !referenced || phaseID == "" {
			continue
		}
		entries = append(entries, outcome_summary.ClientReportPhaseEntry{
			Code:   phase.GetCode(),
			Name:   phase.GetName(),
			Order:  phase.GetPhaseOrder(),
			Active: phase.GetActive(),
		})
	}
	catalog := outcome_summary.BuildClientReportPhaseCatalog(entries)
	options := make([]types.SelectOption, 0, len(catalog)+1)
	for index, phase := range catalog {
		options = append(options, types.SelectOption{Value: phase.Code, Label: phase.Name, Selected: index == 0})
	}
	if clientProjectionHasYearFinal(projection) {
		label := strings.TrimSpace(labels.ClientDocumentDownload.PeriodYearFinal)
		if label == "" {
			label = "Year Final"
		}
		options = append(options, types.SelectOption{Value: clientReportYearFinalPeriod, Label: label, Selected: len(catalog) == 0})
	}
	return options
}

func clientProjectionHasYearFinal(projection *exportpb.ClientReportCardProjection) bool {
	if projection == nil {
		return false
	}
	jobIDs := make(map[string]bool, len(projection.GetJobs()))
	for _, job := range projection.GetJobs() {
		if job != nil && strings.TrimSpace(job.GetId()) != "" {
			jobIDs[job.GetId()] = true
		}
	}
	// The typed document handler can render Year Final from the projected job
	// even before its outcome summary is published; values remain blank then.
	return len(jobIDs) > 0
}

func clientReportFormatOptions(labels outcome_summary.Labels) []types.SelectOption {
	formatLabels := labels.ClientDocumentDownload
	docx := strings.TrimSpace(formatLabels.FormatDocx)
	if docx == "" {
		docx = "DOCX"
	}
	pdf := strings.TrimSpace(formatLabels.FormatPdf)
	if pdf == "" {
		pdf = "PDF"
	}
	return []types.SelectOption{
		{Value: "docx", Label: docx, Selected: true},
		{Value: "pdf", Label: pdf},
	}
}
