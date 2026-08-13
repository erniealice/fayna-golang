package section

// download.go renders the report-only section outcome export drawer.  The
// drawer asks the composite export query for options once, then derives all
// category/period choices locally from that trusted response.  It deliberately
// has no document-template resolver: the native export form owns the actual
// CSV/PDF request and its server-side render gate.

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

const uncategorizedExportValue = "uncategorized"

// DrawerDeps is the deliberately narrow composition seam for the section
// download drawer.  GetSubscriptionGroupOutcomeExport is the report query
// use-case closure; no storage or template resolver belongs in this view.
type DrawerDeps struct {
	Routes               outcome_summary.Routes
	Labels               outcome_summary.Labels
	Options              outcome_summary.Options
	ResolvePrincipalKind func(context.Context) int32

	GetSubscriptionGroupOutcomeExport func(context.Context, *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error)
}

// DrawerData is the template-facing state for the HTMX-loaded drawer partial.
// CommonLabels and Nonce are intentionally generic/injected by the app view
// adapter, matching the other Fayna drawer contracts.
type DrawerData struct {
	FormAction    string
	RefreshAction string
	Categories    []types.SelectOption
	Periods       []types.SelectOption
	Formats       []types.SelectOption
	Labels        outcome_summary.SectionExportLabels
	CommonLabels  any
	Nonce         string
}

// NewDownloadDrawer creates the section export options drawer. The export
// capability and allowed principal kind are checked before any dependency or
// composite query.
func NewDownloadDrawer(deps *DrawerDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if deps == nil || !outcome_summary.CanExplicitExport(perms, ctx, deps.ResolvePrincipalKind) {
			return view.Forbidden(outcome_summary.ExportPermissionEntity + ":" + outcome_summary.ExportReadAction)
		}
		if deps == nil || !deps.Options.SectionExportEnabled() ||
			strings.TrimSpace(deps.Routes.SectionDownloadDrawerURL) == "" ||
			strings.TrimSpace(deps.Routes.SectionExportURL) == "" ||
			deps.GetSubscriptionGroupOutcomeExport == nil {
			return view.Error(fmt.Errorf("section outcome export is not configured"))
		}
		if viewCtx == nil || viewCtx.Request == nil {
			return view.Error(fmt.Errorf("section outcome export request is missing"))
		}

		groupID := strings.TrimSpace(viewCtx.Request.PathValue("id"))
		if groupID == "" {
			return view.ViewResult{Error: fmt.Errorf("subscription group id is required"), StatusCode: http.StatusBadRequest}
		}

		// Options mode intentionally sends only the trusted path group id.  The
		// browser's category value is a local selection, never a persistence
		// selector in this first (and only) composite call.
		resp, err := deps.GetSubscriptionGroupOutcomeExport(ctx, &exportpb.GetSubscriptionGroupOutcomeExportRequest{
			SubscriptionGroupId: groupID,
		})
		if err != nil {
			return view.Error(fmt.Errorf("load section outcome export options: %w", err))
		}
		if resp == nil || !resp.GetSuccess() {
			return view.ViewResult{Error: fmt.Errorf("load section outcome export options: unavailable response"), StatusCode: http.StatusServiceUnavailable}
		}
		if resp.GetContext() == nil || resp.GetContext().GetSubscriptionGroupId() != groupID {
			// Missing, foreign, and unreachable groups deliberately share one shape.
			return view.ViewResult{Error: fmt.Errorf("section outcome export not found"), StatusCode: http.StatusNotFound}
		}
		if len(resp.GetJobCategories()) == 0 {
			return view.ViewResult{Error: fmt.Errorf("section outcome export is not computed"), StatusCode: http.StatusNotFound}
		}

		requestedCategory := strings.TrimSpace(viewCtx.Request.URL.Query().Get("job_category_id"))
		category := chooseCategory(resp.GetJobCategories(), requestedCategory, deps.Options.SectionExport.DefaultCategoryCode)
		data := &DrawerData{
			FormAction:    route.ResolveURL(deps.Routes.SectionExportURL, "id", groupID),
			RefreshAction: route.ResolveURL(deps.Routes.SectionDownloadDrawerURL, "id", groupID),
			Categories:    buildCategoryOptions(deps.Labels, resp.GetJobCategories(), category),
			Periods:       buildPeriodOptions(deps.Labels, category),
			Formats:       buildFormatOptions(deps.Options, category, deps.Labels),
			Labels:        deps.Labels.SectionExport,
		}
		return view.OK("outcome-summary-section-download-drawer-form", data)
	})
}

func sortedCategories(categories []*exportpb.JobCategoryOption) []*exportpb.JobCategoryOption {
	result := append([]*exportpb.JobCategoryOption(nil), categories...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i] == nil {
			return false
		}
		if result[j] == nil {
			return true
		}
		if result[i].GetSortOrder() != result[j].GetSortOrder() {
			return result[i].GetSortOrder() < result[j].GetSortOrder()
		}
		leftName := strings.ToLower(strings.TrimSpace(result[i].GetName()))
		rightName := strings.ToLower(strings.TrimSpace(result[j].GetName()))
		if leftName != rightName {
			return leftName < rightName
		}
		return result[i].GetJobCategoryId() < result[j].GetJobCategoryId()
	})
	return result
}

func categoryValue(category *exportpb.JobCategoryOption) string {
	if category == nil || strings.TrimSpace(category.GetJobCategoryId()) == "" {
		return uncategorizedExportValue
	}
	return category.GetJobCategoryId()
}

func categoryLabel(labels outcome_summary.Labels, category *exportpb.JobCategoryOption) string {
	if category == nil || strings.TrimSpace(category.GetName()) == "" {
		return labels.Landing.UncategorizedColumn
	}
	return category.GetName()
}

// chooseCategory validates the request against the response, then applies the
// trusted configured code and finally the response's deterministic sort order.
func chooseCategory(categories []*exportpb.JobCategoryOption, requestedID, configuredCode string) *exportpb.JobCategoryOption {
	sorted := sortedCategories(categories)
	for _, category := range sorted {
		if category != nil && requestedID != "" && categoryValue(category) == requestedID {
			return category
		}
	}
	for _, category := range sorted {
		if category != nil && strings.TrimSpace(configuredCode) != "" && category.GetCode() == strings.TrimSpace(configuredCode) {
			return category
		}
	}
	for _, category := range sorted {
		if category != nil {
			return category
		}
	}
	return nil
}

func buildCategoryOptions(labels outcome_summary.Labels, categories []*exportpb.JobCategoryOption, selected *exportpb.JobCategoryOption) []types.SelectOption {
	selectedValue := categoryValue(selected)
	options := make([]types.SelectOption, 0, len(categories))
	for _, category := range sortedCategories(categories) {
		if category == nil {
			continue
		}
		value := categoryValue(category)
		options = append(options, types.SelectOption{
			Value:    value,
			Label:    categoryLabel(labels, category),
			Selected: value == selectedValue,
		})
	}
	return options
}

func buildPeriodOptions(labels outcome_summary.Labels, category *exportpb.JobCategoryOption) []types.SelectOption {
	if category == nil {
		return nil
	}
	phases := make([]*exportpb.JobTemplatePhaseOption, 0, len(category.GetJobTemplatePhases()))
	for _, phase := range category.GetJobTemplatePhases() {
		if phase == nil || strings.TrimSpace(phase.GetCode()) == "" || phase.GetAmbiguous() {
			continue
		}
		phases = append(phases, phase)
	}
	sort.SliceStable(phases, func(i, j int) bool {
		if phases[i].GetSequenceOrder() != phases[j].GetSequenceOrder() {
			return phases[i].GetSequenceOrder() < phases[j].GetSequenceOrder()
		}
		return phases[i].GetCode() < phases[j].GetCode()
	})
	options := make([]types.SelectOption, 0, len(phases)+1)
	for _, phase := range phases {
		label := strings.TrimSpace(phase.GetName())
		if label == "" {
			label = phase.GetCode()
		}
		options = append(options, types.SelectOption{Value: "phase:" + phase.GetCode(), Label: label})
	}
	if category.GetFinalOutcomeAvailable() {
		options = append(options, types.SelectOption{Value: "final", Label: labels.SectionExport.PeriodFinal})
	}
	return options
}

func buildFormatOptions(options outcome_summary.Options, category *exportpb.JobCategoryOption, labels outcome_summary.Labels) []types.SelectOption {
	categoryCode := ""
	if category != nil {
		categoryCode = category.GetCode()
	}
	_, pdfAvailable := options.SectionExport.ProfileForCategoryCode(categoryCode)
	return []types.SelectOption{
		{Value: "csv", Label: labels.SectionExport.FormatCSV, Selected: true},
		{Value: "pdf", Label: labels.SectionExport.FormatPDF, Disabled: !pdfAvailable},
	}
}
