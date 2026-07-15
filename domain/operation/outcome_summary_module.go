package operation

import (
	"context"
	"log"
	"net/http"

	outcomesummarypkg "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	clientcard "github.com/erniealice/fayna-golang/domain/operation/outcome_summary/client_card"
	documentview "github.com/erniealice/fayna-golang/domain/operation/outcome_summary/document"
	jobsummary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary/job_summary"
	summarylist "github.com/erniealice/fayna-golang/domain/operation/outcome_summary/list"
	phasesummary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary/phase_summary"
	sectionview "github.com/erniealice/fayna-golang/domain/operation/outcome_summary/section"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	clientattributepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client_attribute"
	workspaceuserpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/workspace_user"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	joblinepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_line"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	phasesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	ttcpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
	priceschedulepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/price_schedule"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
	subscriptiongroupworkspaceuserpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_workspace_user"
	summarypb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/job_template_summary"
)

// OutcomeSummaryModuleDeps holds all dependencies for the outcome summary module.
type OutcomeSummaryModuleDeps struct {
	Routes       outcomesummarypkg.Routes
	Labels       outcomesummarypkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Options — app-configured presentation for the report-cards surfaces
	// (view-1 tabstrip + what to list, view-2 row bands/sort). Zero value →
	// view-1 renders the flat job_outcome_summary list unchanged.
	Options outcomesummarypkg.Options

	// Job outcome summary operations
	GetJobOutcomeSummaryByJob func(ctx context.Context, req *jobsumpb.GetJobOutcomeSummaryByJobRequest) (*jobsumpb.GetJobOutcomeSummaryByJobResponse, error)
	ListJobOutcomeSummarys    func(ctx context.Context, req *jobsumpb.ListJobOutcomeSummarysRequest) (*jobsumpb.ListJobOutcomeSummarysResponse, error)

	// Report-card document (.docx) download deps. ListJobOutcomeLines backs the
	// per-criterion transcript fetch (G2); GenerateDoc is the injected fycha
	// doctemplate closure (nil-safe — the download route fails closed with 503).
	ListJobOutcomeLines func(ctx context.Context, req *joblinepb.ListJobOutcomeLinesRequest) (*joblinepb.ListJobOutcomeLinesResponse, error)
	// Per-criterion (crit_a..crit_d + criteria_total) transcript path: task_outcome
	// reached through job_task, A/B/C/D ordered via template_task_criteria. All
	// optional/nil-safe.
	ListJobTasks              func(ctx context.Context, req *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error)
	ListTaskOutcomes          func(ctx context.Context, req *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error)
	ListTemplateTaskCriterias func(ctx context.Context, req *ttcpb.ListTemplateTaskCriteriasRequest) (*ttcpb.ListTemplateTaskCriteriasResponse, error)
	GenerateDoc               func(templateData []byte, data map[string]any) ([]byte, error)
	// ResolveTemplateBytes resolves the operator-uploaded, AY-scoped report-card
	// template binding for a card's price_schedule (binding resolver ∘ storage
	// download). Returns (nil, nil) → the document handler falls back to the
	// embedded template. Optional/nil-safe (no download regression).
	ResolveTemplateBytes func(ctx context.Context, priceScheduleID string) ([]byte, error)
	// DocumentHeaderName is the generic report-card document header (lyngua-sourced;
	// blank falls back to the landing title). Generic — no vertical vocabulary in
	// code (the rendered "school name" wording lives in a lyngua value).
	DocumentHeaderName string

	// Phase outcome summary operations
	GetPhaseOutcomeSummaryByJobPhase func(ctx context.Context, req *phasesumpb.GetPhaseOutcomeSummaryByJobPhaseRequest) (*phasesumpb.GetPhaseOutcomeSummaryByJobPhaseResponse, error)
	ListPhaseOutcomeSummarysByJob    func(ctx context.Context, req *phasesumpb.ListPhaseOutcomeSummarysByJobRequest) (*phasesumpb.ListPhaseOutcomeSummarysByJobResponse, error)

	// Report-cards navigation deps (view-1 landing + view-2 section grid). All
	// optional/nil-safe: a nil closure degrades the affected surface to its
	// empty/flat state, never a panic.
	ListPriceSchedules                  func(ctx context.Context, req *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error)
	ListSubscriptionGroups              func(ctx context.Context, req *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error)
	ListSubscriptionGroupMembers        func(ctx context.Context, req *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error)
	ListSubscriptionGroupWorkspaceUsers func(ctx context.Context, req *subscriptiongroupworkspaceuserpb.ListSubscriptionGroupWorkspaceUsersRequest) (*subscriptiongroupworkspaceuserpb.ListSubscriptionGroupWorkspaceUsersResponse, error)
	ListWorkspaceUsers                  func(ctx context.Context, req *workspaceuserpb.ListWorkspaceUsersRequest) (*workspaceuserpb.ListWorkspaceUsersResponse, error)
	ListJobs                            func(ctx context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error)
	ListJobPhases                       func(ctx context.Context, req *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error)
	ListJobTemplates                    func(ctx context.Context, req *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error)
	ListClients                         func(ctx context.Context, req *clientpb.ListClientsRequest) (*clientpb.ListClientsResponse, error)
	ListClientAttributes                func(ctx context.Context, req *clientattributepb.ListClientAttributesRequest) (*clientattributepb.ListClientAttributesResponse, error)
	ResolveAttributeIDByCode            func(ctx context.Context, code string) (string, error)
	ListJobTemplateSummaries            func(ctx context.Context, req *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error)

	// ListJobCategories resolves Options.CategoryFilter (a job_category code, e.g.
	// "academic") to its id so the section grid, client card, and report-card
	// document drop same-origin deportment jobs (gate H2). Optional/nil-safe —
	// a nil closure (or empty CategoryFilter) applies no filter.
	ListJobCategories func(ctx context.Context, req *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error)
}

// OutcomeSummaryModule holds all constructed outcome summary views.
type OutcomeSummaryModule struct {
	routes       outcomesummarypkg.Routes
	List         view.View
	Section      view.View
	ClientCard   view.View
	JobSummary   view.View
	PhaseSummary view.View
	// SectionExport is the section-grid CSV download (a raw handler — the
	// registrar wraps it with the same RBAC context injection as views).
	SectionExport http.HandlerFunc
	// StudentDocument is the per-student report-card .docx download (a raw
	// handler wrapped like SectionExport). Nil when GenerateDoc is not wired.
	StudentDocument http.HandlerFunc
}

// NewOutcomeSummaryModule creates a new outcome summary module with all views wired.
func NewOutcomeSummaryModule(deps *OutcomeSummaryModuleDeps) *OutcomeSummaryModule {
	sectionDeps := &sectionview.Deps{
		Routes:                              deps.Routes,
		Labels:                              deps.Labels,
		CommonLabels:                        deps.CommonLabels,
		TableLabels:                         deps.TableLabels,
		Options:                             deps.Options,
		ListSubscriptionGroups:              deps.ListSubscriptionGroups,
		ListSubscriptionGroupMembers:        deps.ListSubscriptionGroupMembers,
		ListJobs:                            deps.ListJobs,
		ListJobTemplates:                    deps.ListJobTemplates,
		ListClients:                         deps.ListClients,
		ListJobOutcomeSummarys:              deps.ListJobOutcomeSummarys,
		ListClientAttributes:                deps.ListClientAttributes,
		ResolveAttributeIDByCode:            deps.ResolveAttributeIDByCode,
		ListSubscriptionGroupWorkspaceUsers: deps.ListSubscriptionGroupWorkspaceUsers,
		ListWorkspaceUsers:                  deps.ListWorkspaceUsers,
		ListJobCategories:                   deps.ListJobCategories,
	}
	return &OutcomeSummaryModule{
		routes: deps.Routes,
		List: summarylist.NewView(&summarylist.ListViewDeps{
			Routes:                       deps.Routes,
			Labels:                       deps.Labels,
			CommonLabels:                 deps.CommonLabels,
			TableLabels:                  deps.TableLabels,
			Options:                      deps.Options,
			ListJobOutcomeSummarys:       deps.ListJobOutcomeSummarys,
			ListPriceSchedules:           deps.ListPriceSchedules,
			ListSubscriptionGroups:       deps.ListSubscriptionGroups,
			ListJobTemplateSummaries:     deps.ListJobTemplateSummaries,
			ListSubscriptionGroupMembers: deps.ListSubscriptionGroupMembers,
			ListJobs:                     deps.ListJobs,
		}),
		Section:         sectionview.NewView(sectionDeps),
		SectionExport:   sectionview.NewExportHandler(sectionDeps),
		StudentDocument: newStudentDocumentHandler(deps),
		ClientCard: clientcard.NewView(&clientcard.Deps{
			Routes:                        deps.Routes,
			Labels:                        deps.Labels,
			CommonLabels:                  deps.CommonLabels,
			TableLabels:                   deps.TableLabels,
			CategoryFilter:                deps.Options.CategoryFilter,
			ListJobCategories:             deps.ListJobCategories,
			ListSubscriptionGroups:        deps.ListSubscriptionGroups,
			ListSubscriptionGroupMembers:  deps.ListSubscriptionGroupMembers,
			ListJobs:                      deps.ListJobs,
			ListJobTemplates:              deps.ListJobTemplates,
			ListClients:                   deps.ListClients,
			ListJobOutcomeSummarys:        deps.ListJobOutcomeSummarys,
			ListPhaseOutcomeSummarysByJob: deps.ListPhaseOutcomeSummarysByJob,
			ListJobPhases:                 deps.ListJobPhases,
		}),
		JobSummary: jobsummary.NewView(&jobsummary.Deps{
			Routes:                    deps.Routes,
			Labels:                    deps.Labels,
			CommonLabels:              deps.CommonLabels,
			GetJobOutcomeSummaryByJob: deps.GetJobOutcomeSummaryByJob,
		}),
		PhaseSummary: phasesummary.NewView(&phasesummary.Deps{
			Routes:                           deps.Routes,
			Labels:                           deps.Labels,
			CommonLabels:                     deps.CommonLabels,
			GetPhaseOutcomeSummaryByJobPhase: deps.GetPhaseOutcomeSummaryByJobPhase,
		}),
	}
}

// newStudentDocumentHandler builds the per-student report-card .docx download
// handler from the module deps. Returns nil when GenerateDoc is not wired (the
// app did not inject the fycha doctemplate closure) — RegisterRoutes then skips
// the route rather than registering a handler that would always 503.
func newStudentDocumentHandler(deps *OutcomeSummaryModuleDeps) http.HandlerFunc {
	if deps.GenerateDoc == nil {
		return nil
	}
	return documentview.NewDownloadHandler(&documentview.Deps{
		Labels:                        deps.Labels,
		CommonLabels:                  deps.CommonLabels,
		DocumentHeaderName:            deps.DocumentHeaderName,
		CategoryFilter:                deps.Options.CategoryFilter,
		ListJobCategories:             deps.ListJobCategories,
		GenerateDoc:                   deps.GenerateDoc,
		ResolveTemplateBytes:          deps.ResolveTemplateBytes,
		ListSubscriptionGroups:        deps.ListSubscriptionGroups,
		ListSubscriptionGroupMembers:  deps.ListSubscriptionGroupMembers,
		ListJobs:                      deps.ListJobs,
		ListJobTemplates:              deps.ListJobTemplates,
		ListClients:                   deps.ListClients,
		ListJobOutcomeSummarys:        deps.ListJobOutcomeSummarys,
		ListPhaseOutcomeSummarysByJob: deps.ListPhaseOutcomeSummarysByJob,
		ListJobPhases:                 deps.ListJobPhases,
		ListJobOutcomeLines:           deps.ListJobOutcomeLines,
		ListJobTasks:                  deps.ListJobTasks,
		ListTaskOutcomes:              deps.ListTaskOutcomes,
		ListTemplateTaskCriterias:     deps.ListTemplateTaskCriterias,
	})
}

// RegisterRoutes registers all outcome summary routes.
func (m *OutcomeSummaryModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	if m.Section != nil && m.routes.SectionURL != "" {
		r.GET(m.routes.SectionURL, m.Section)
	}
	if m.ClientCard != nil && m.routes.ClientCardURL != "" {
		r.GET(m.routes.ClientCardURL, m.ClientCard)
	}
	if m.SectionExport != nil && m.routes.SectionExportURL != "" {
		// Raw (non-view) route — the registrar's HandleFunc path wraps it with
		// the ViewAdapter's RBAC context injection (WrapHandler), so the
		// handler's view.GetUserPermissions gate observes real permissions.
		if rr, ok := r.(interface {
			HandleFunc(method, path string, handler http.HandlerFunc, middlewares ...string)
		}); ok {
			rr.HandleFunc("GET", m.routes.SectionExportURL, m.SectionExport)
		} else {
			log.Printf("outcome summary: RouteRegistrar does not support HandleFunc — skipping GET %s", m.routes.SectionExportURL)
		}
	}
	if m.StudentDocument != nil && m.routes.ClientDocumentURL != "" {
		// Raw (non-view) route — the registrar's HandleFunc path wraps it with
		// the ViewAdapter's RBAC context injection (WrapHandler), exactly like
		// SectionExport, so the handler's view.GetUserPermissions gate observes
		// real permissions.
		if rr, ok := r.(interface {
			HandleFunc(method, path string, handler http.HandlerFunc, middlewares ...string)
		}); ok {
			rr.HandleFunc("GET", m.routes.ClientDocumentURL, m.StudentDocument)
		} else {
			log.Printf("outcome summary: RouteRegistrar does not support HandleFunc — skipping GET %s", m.routes.ClientDocumentURL)
		}
	}
	r.GET(m.routes.JobSummaryURL, m.JobSummary)
	r.GET(m.routes.PhaseSummaryURL, m.PhaseSummary)
}
