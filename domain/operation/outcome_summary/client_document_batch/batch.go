// Package client_document_batch serves the subscription-group batch download of
// every member's client document for ONE period. The server side is read-only:
// a Period/Format drawer and a JSON manifest. The browser module
// (pyeza bulk-download.js) fetches each document through the existing
// per-client document route, so authorization, render gating, redaction and
// template resolution stay in exactly one place (document.handleExplicitClientDocument).
//
// Both surfaces gate on the same capability as that per-client route
// (outcome_summary.CanExplicitExport) and read through the same scoped espyna
// export use case (GetSubscriptionGroupOutcomeExport shares reportScope with
// GetSubscriptionGroupClientOutcomeSummary), so a manifest can never list a
// member whose document route would deny the caller.
package client_document_batch

import (
	"context"
	"strings"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	subscriptiongroupdocument "github.com/erniealice/fayna-golang/domain/operation/outcome_summary/subscription_group_document"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

// PeriodYearFinal is the per-client document route's whole-year period value
// (document.clientDocumentPeriodYearFinal).
const PeriodYearFinal = "year_final"

// Formats accepted by the per-client document route.
const (
	FormatPDF  = "pdf"
	FormatDOCX = "docx"
)

// DefaultConcurrency is the browser's in-flight document fetch count when the
// app sets no Options.Document.BatchConcurrency: three fully used converter
// instances at containerConcurrency 2 (see the plan discovery).
const DefaultConcurrency = 6

// concurrencyFor returns the configured in-flight count, or DefaultConcurrency
// when unset.
func concurrencyFor(configured int) int {
	if configured > 0 {
		return configured
	}
	return DefaultConcurrency
}

// Deps is the narrow composition seam. Only the typed scoped export query and
// the published-template resolver are reachable; no document bytes are built.
type Deps struct {
	Routes               outcome_summary.Routes
	Labels               outcome_summary.Labels
	CommonLabels         any
	Options              outcome_summary.Options
	ResolvePrincipalKind func(context.Context) int32

	// PDFAvailable / DOCXAvailable mirror whether the per-client route has its
	// GeneratePDF / GenerateDoc closures (a missing closure 503s every file).
	PDFAvailable  bool
	DOCXAvailable bool

	GetSubscriptionGroupOutcomeExport func(context.Context, *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error)
	// ResolveTemplateBytes is the same published report-card binding resolver
	// the per-client route uses; an explicit phase requires a bound template.
	ResolveTemplateBytes func(ctx context.Context, priceScheduleID, phaseCode string) ([]byte, error)
}

// Configured reports whether the batch surface may be offered and mounted. It
// mirrors the single-client drawer's configuration gate
// (client_card.NewDownloadDrawer) plus the two batch routes.
func Configured(options outcome_summary.Options, routes outcome_summary.Routes) bool {
	profile, registered := subscriptiongroupdocument.LookupProfile(options.SubscriptionGroupExport.WholeReportProfile)
	return options.SubscriptionGroupExportEnabled() && registered &&
		profile.CategoryBindingScope == subscriptiongroupdocument.CategoryBindingScopeAllCategories &&
		strings.TrimSpace(routes.ClientDocumentURL) != "" &&
		strings.TrimSpace(routes.ClientDocumentsDownloadDrawerURL) != "" &&
		strings.TrimSpace(routes.ClientDocumentsManifestURL) != ""
}

// OpenAction is the toolbar button that opens the batch drawer. ActionURL is
// an /action/ route (workspace-rewrite pass-through).
type OpenAction struct {
	ActionURL string
	Label     string
	Title     string
}

// ActionConfig carries what the subscription-group page already holds.
type ActionConfig struct {
	Routes               outcome_summary.Routes
	Labels               outcome_summary.Labels
	Options              outcome_summary.Options
	ResolvePrincipalKind func(context.Context) int32
	// ClientDocumentMounted mirrors the module's per-client document route mount;
	// ExportQueryWired mirrors a non-nil GetSubscriptionGroupOutcomeExport.
	ClientDocumentMounted bool
	ExportQueryWired      bool
}

// NewOpenAction returns the toolbar action, or nil when the caller could not use
// it. Visibility follows the per-row explicit client download (page.go
// canDownloadClientDocument) so the batch never appears where a single
// document would be denied.
func NewOpenAction(ctx context.Context, cfg ActionConfig, subscriptionGroupID string) *OpenAction {
	subscriptionGroupID = strings.TrimSpace(subscriptionGroupID)
	if subscriptionGroupID == "" || !cfg.ClientDocumentMounted || !cfg.ExportQueryWired ||
		!Configured(cfg.Options, cfg.Routes) ||
		!outcome_summary.CanExplicitExport(view.GetUserPermissions(ctx), ctx, cfg.ResolvePrincipalKind) {
		return nil
	}
	l := cfg.Labels.ClientDocumentsDownload
	return &OpenAction{
		ActionURL: route.ResolveURL(cfg.Routes.ClientDocumentsDownloadDrawerURL, "id", subscriptionGroupID),
		Label:     l.OpenAction,
		Title:     l.DrawerTitle,
	}
}

// periodOption is one downloadable period for the group.
type periodOption struct {
	Code  string
	Label string
}

// groupPeriods returns the union of non-ambiguous phase options across the
// group's categories in catalog order, then Year Final when any category
// exists. A phase that is ambiguous in every category is not offered: its
// member list could not be proven through the scoped export validation.
func groupPeriods(labels outcome_summary.Labels, categories []*exportpb.JobCategoryOption) []periodOption {
	entries := make([]outcome_summary.ClientReportPhaseEntry, 0)
	present := false
	for _, category := range categories {
		if category == nil || strings.TrimSpace(category.GetJobCategoryId()) == "" {
			continue
		}
		present = true
		for _, phase := range category.GetJobTemplatePhases() {
			if phase == nil || phase.GetAmbiguous() {
				continue
			}
			entries = append(entries, outcome_summary.ClientReportPhaseEntry{
				Code:   phase.GetCode(),
				Name:   phase.GetName(),
				Order:  phase.GetSequenceOrder(),
				Active: true,
			})
		}
	}
	catalog := outcome_summary.BuildClientReportPhaseCatalog(entries)
	periods := make([]periodOption, 0, len(catalog)+1)
	for _, phase := range catalog {
		periods = append(periods, periodOption{Code: phase.Code, Label: phase.Name})
	}
	if present {
		label := strings.TrimSpace(labels.ClientDocumentDownload.PeriodYearFinal)
		if label == "" {
			label = "Year Final"
		}
		periods = append(periods, periodOption{Code: PeriodYearFinal, Label: label})
	}
	return periods
}

func findPeriod(periods []periodOption, code string) (periodOption, bool) {
	for _, period := range periods {
		if period.Code == code {
			return period, true
		}
	}
	return periodOption{}, false
}

// memberQuery builds the ONE scoped export request that lists the group's
// members. Client rows are category- and selector-independent under the
// current whole-assigned-subscription-group policy (postgres matrix_members CTE), so the
// requested period's own selector is preferred only to keep the read aligned
// with the period; any validating selector proves the same member set.
func memberQuery(subscriptionGroupID, period string, categories []*exportpb.JobCategoryOption) *exportpb.GetSubscriptionGroupOutcomeExportRequest {
	request := func(categoryID string) *exportpb.GetSubscriptionGroupOutcomeExportRequest {
		id := categoryID
		return &exportpb.GetSubscriptionGroupOutcomeExportRequest{SubscriptionGroupId: subscriptionGroupID, JobCategoryId: &id}
	}
	if period == PeriodYearFinal {
		for _, category := range categories {
			if category != nil && strings.TrimSpace(category.GetJobCategoryId()) != "" && category.GetFinalOutcomeAvailable() {
				req := request(category.GetJobCategoryId())
				req.OutcomeSelector = &exportpb.GetSubscriptionGroupOutcomeExportRequest_FinalOutcome{FinalOutcome: true}
				return req
			}
		}
	} else {
		for _, category := range categories {
			if code, ok := nonAmbiguousPhase(category, period); ok {
				req := request(category.GetJobCategoryId())
				req.OutcomeSelector = &exportpb.GetSubscriptionGroupOutcomeExportRequest_JobTemplatePhaseCode{JobTemplatePhaseCode: code}
				return req
			}
		}
	}
	for _, category := range categories {
		if code, ok := nonAmbiguousPhase(category, ""); ok {
			req := request(category.GetJobCategoryId())
			req.OutcomeSelector = &exportpb.GetSubscriptionGroupOutcomeExportRequest_JobTemplatePhaseCode{JobTemplatePhaseCode: code}
			return req
		}
	}
	return nil
}

// nonAmbiguousPhase returns want (or, when want is empty, the first phase) if
// the category offers it unambiguously.
func nonAmbiguousPhase(category *exportpb.JobCategoryOption, want string) (string, bool) {
	if category == nil || strings.TrimSpace(category.GetJobCategoryId()) == "" {
		return "", false
	}
	for _, phase := range category.GetJobTemplatePhases() {
		if phase == nil || phase.GetAmbiguous() || strings.TrimSpace(phase.GetCode()) == "" {
			continue
		}
		if want == "" || phase.GetCode() == want {
			return phase.GetCode(), true
		}
	}
	return "", false
}

func contextMatches(response *exportpb.GetSubscriptionGroupOutcomeExportResponse, subscriptionGroupID string) bool {
	return response != nil && response.GetSuccess() && response.GetContext() != nil &&
		response.GetContext().GetSubscriptionGroupId() == subscriptionGroupID
}

// clientDisplayName mirrors the group report grid's "Last, First" display
// (subscription_group.reportClientName) without the exporter's row prefix.
func clientDisplayName(row *exportpb.SubscriptionGroupOutcomeClientRow) string {
	last := strings.TrimSpace(row.GetClientLastName())
	first := strings.TrimSpace(row.GetClientFirstName())
	if last != "" && first != "" {
		return last + ", " + first
	}
	if name := strings.TrimSpace(row.GetClientName()); name != "" {
		return name
	}
	if name := strings.TrimSpace(strings.Join([]string{first, last}, " ")); name != "" {
		return name
	}
	return strings.TrimSpace(row.GetClientId())
}
