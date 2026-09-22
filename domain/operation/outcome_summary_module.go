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
	subscriptiongroupview "github.com/erniealice/fayna-golang/domain/operation/outcome_summary/subscription_group"
	subscriptiongroupdocumenttemplate "github.com/erniealice/fayna-golang/domain/operation/outcome_summary/subscription_group_document_template_settings"
	templatesettings "github.com/erniealice/fayna-golang/domain/operation/outcome_summary/template_settings"

	espynaports "github.com/erniealice/espyna-golang/ports"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	documenttemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/template"
	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	clientattributepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client_attribute"
	staffpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/staff"
	workspaceuserpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/workspace_user"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	joblinepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_line"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary_document_template"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	phasesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	subscriptiongroupdocumenttemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/subscription_group_document_template"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	ttcpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
	planpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/plan"
	priceschedulepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/price_schedule"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
	subscriptiongroupworkspaceuserpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_workspace_user"
	summarypb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/job_template_summary"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
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
	// ResolvePrincipalKind is owned by app composition and returns the active
	// session principal kind. Nil is fail-closed for explicit group exports.
	ResolvePrincipalKind func(context.Context) int32

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
	ListJobTasks     func(ctx context.Context, req *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error)
	ListTaskOutcomes func(ctx context.Context, req *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error)
	// Ownership-joined latest-cell read carrying phase/task/criterion codes
	// (job → template ancestry). Optional/nil-safe.
	ListCodedTaskOutcomeValuesByJob func(ctx context.Context, req *taskoutcomepb.ListCodedTaskOutcomeValuesByJobRequest) (*taskoutcomepb.ListCodedTaskOutcomeValuesByJobResponse, error)
	// Past-AY sibling of the above: admits inactive historical ancestry so past
	// report cards resolve their coded/attendance cells. Optional/nil-safe.
	ListCodedTaskOutcomeValuesByJobHistorical func(ctx context.Context, req *taskoutcomepb.ListCodedTaskOutcomeValuesByJobRequest) (*taskoutcomepb.ListCodedTaskOutcomeValuesByJobResponse, error)
	ListTemplateTaskCriterias                 func(ctx context.Context, req *ttcpb.ListTemplateTaskCriteriasRequest) (*ttcpb.ListTemplateTaskCriteriasResponse, error)
	// v2 block-layout document enrichments (optional/nil-safe): criterion
	// display names + the User-hydrating staff read for the per-subject staff
	// line and the group-lead ("Adviser") resolution.
	ListOutcomeCriterias func(ctx context.Context, req *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error)
	GetStaffListPageData func(ctx context.Context, req *staffpb.GetStaffListPageDataRequest) (*staffpb.GetStaffListPageDataResponse, error)
	GenerateDoc          func(templateData []byte, data map[string]any) ([]byte, error)
	// GeneratePDF is the injected fycha ProcessBytesToPDF closure (DOCX → PDF via
	// LibreOffice) — a SECOND closure mirroring GenerateDoc. Nil-safe: the
	// download route still registers on GenerateDoc alone (DOCX baseline); a
	// ?format=pdf request with GeneratePDF nil is a narrower per-format 503.
	GeneratePDF func(templateData []byte, data map[string]any) ([]byte, error)
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

	// Report-cards navigation deps (view-1 landing + view-2 group grid). All
	// optional/nil-safe: a nil closure degrades the affected surface to its
	// empty/flat state, never a panic.
	ListPriceSchedules                  func(ctx context.Context, req *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error)
	ListPlans                           func(ctx context.Context, req *planpb.ListPlansRequest) (*planpb.ListPlansResponse, error)
	ListSubscriptionGroups              func(ctx context.Context, req *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error)
	ListSubscriptionGroupMembers        func(ctx context.Context, req *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error)
	ListSubscriptionGroupWorkspaceUsers func(ctx context.Context, req *subscriptiongroupworkspaceuserpb.ListSubscriptionGroupWorkspaceUsersRequest) (*subscriptiongroupworkspaceuserpb.ListSubscriptionGroupWorkspaceUsersResponse, error)
	ListWorkspaceUsers                  func(ctx context.Context, req *workspaceuserpb.ListWorkspaceUsersRequest) (*workspaceuserpb.ListWorkspaceUsersResponse, error)
	ListJobs                            func(ctx context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error)
	ListJobPhases                       func(ctx context.Context, req *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error)
	// GetPhaseApprovalGateRollup — the group-grain render-gate input port
	// (per-template-phase group rollup with the applied-group echo). Consumed
	// ONLY when Options.Document.GateGrain selects the subscription-group
	// grain; nil with that grain configured fails the document gate CLOSED
	// (503) — never a fallback to the template-grain walk. Unset grain never
	// calls it.
	GetPhaseApprovalGateRollup func(ctx context.Context, req *matrixpb.GetPhaseApprovalGateRollupRequest) (*matrixpb.GetPhaseApprovalGateRollupResponse, error)
	// ListJobTemplatePhasesByTemplate resolves a job_template's phases (with their
	// stable `code`) so the report-card block tree can key per-phase leaves by
	// phase code. Optional/nil-safe.
	ListJobTemplatePhasesByTemplate          func(ctx context.Context, req *jobtemplatephasepb.ListByJobTemplateRequest) (*jobtemplatephasepb.ListByJobTemplateResponse, error)
	ListJobTemplates                         func(ctx context.Context, req *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error)
	ListClients                              func(ctx context.Context, req *clientpb.ListClientsRequest) (*clientpb.ListClientsResponse, error)
	ListClientAttributes                     func(ctx context.Context, req *clientattributepb.ListClientAttributesRequest) (*clientattributepb.ListClientAttributesResponse, error)
	ListAttributes                           func(ctx context.Context, req *commonpb.ListAttributesRequest) (*commonpb.ListAttributesResponse, error)
	ResolveAttributeIDByCode                 func(ctx context.Context, code string) (string, error)
	GetSubscriptionGroupOutcomeExport        func(ctx context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error)
	ListSubscriptionGroupOutcomeLanding      func(ctx context.Context, req *espynaports.SubscriptionGroupOutcomeLandingRequest) (*espynaports.SubscriptionGroupOutcomeLandingResponse, error)
	ResolveSubscriptionGroupDocumentTemplate func(ctx context.Context, req *exportpb.ResolveSubscriptionGroupOutcomeDocumentForRenderRequest) (*outcomesummarypkg.ResolvedSubscriptionGroupDocumentTemplate, error)
	ListJobTemplateSummaries                 func(ctx context.Context, req *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error)

	// ListJobListTabSupport (R9 W-A2) — the ONE no-argument single-statement
	// UNION tab-support read (all job_category rows + ACTIVE job_template
	// stubs, per-kind permission-intersected in the espyna use case). The
	// landing consumes it for its dynamic category count columns (headers +
	// template→category map); it is the SAME closure the job list's "/classes"
	// tabstrip consumes. Optional/nil-safe — nil degrades the landing to its
	// static column set.
	ListJobListTabSupport func(ctx context.Context) ([]*jobcategorypb.JobCategory, []*jobtemplatepb.JobTemplate, error)

	// ListJobCategories resolves Options.CategoryFilter (a job_category code, e.g.
	// "academic") to its id so the group grid, client card, and report-card
	// document drop same-origin deportment jobs (gate H2). Optional/nil-safe —
	// a nil closure (or empty CategoryFilter) applies no filter.
	ListJobCategories func(ctx context.Context, req *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error)

	// Report-card template settings (TB3). The document_template artifact +
	// storage closures come from the app AppContext; the binding lifecycle
	// closures come from the espyna binding use cases via the block seam. All
	// optional/nil-safe — a nil write closure degrades the settings surface to a
	// "not configured" error (never a panic). The list page still renders.
	UploadTemplate         func(ctx context.Context, bucket, key string, content []byte, contentType string) error
	ListDocumentTemplates  func(ctx context.Context, req *documenttemplatepb.ListDocumentTemplatesRequest) (*documenttemplatepb.ListDocumentTemplatesResponse, error)
	CreateDocumentTemplate func(ctx context.Context, req *documenttemplatepb.CreateDocumentTemplateRequest) (*documenttemplatepb.CreateDocumentTemplateResponse, error)
	// DeleteDocumentTemplate backs the Q4 upload-orphan cleanup (upload
	// compensation + post-delete artifact reap). Optional/nil-safe — when
	// unwired the cleanup degrades to a logged no-op.
	DeleteDocumentTemplate func(ctx context.Context, req *documenttemplatepb.DeleteDocumentTemplateRequest) (*documenttemplatepb.DeleteDocumentTemplateResponse, error)
	ListTemplateBindings   func(ctx context.Context, req *bindingpb.ListJobOutcomeSummaryDocumentTemplatesRequest) (*bindingpb.ListJobOutcomeSummaryDocumentTemplatesResponse, error)
	CreateTemplateBinding  func(ctx context.Context, req *bindingpb.CreateJobOutcomeSummaryDocumentTemplateRequest) (*bindingpb.CreateJobOutcomeSummaryDocumentTemplateResponse, error)
	DeleteTemplateBinding  func(ctx context.Context, req *bindingpb.DeleteJobOutcomeSummaryDocumentTemplateRequest) (*bindingpb.DeleteJobOutcomeSummaryDocumentTemplateResponse, error)
	PublishTemplateBinding func(ctx context.Context, req *bindingpb.PublishJobOutcomeSummaryDocumentTemplateRequest) (*bindingpb.PublishJobOutcomeSummaryDocumentTemplateResponse, error)

	// Subscription-group document template settings use a distinct binding family and a bytes-first,
	// atomic artifact+DRAFT pair. Storage closures accept only the generated
	// exact locator and are never exposed as routes.
	StoreSubscriptionGroupDocumentTemplate            func(context.Context, string, []byte, string) (string, error)
	DeleteSubscriptionGroupDocumentTemplateObject     func(context.Context, string, string) error
	CreateSubscriptionGroupDocumentTemplateUploadPair func(context.Context, *documenttemplatepb.DocumentTemplate, *subscriptiongroupdocumenttemplatepb.SubscriptionGroupDocumentTemplate) (*documenttemplatepb.DocumentTemplate, *subscriptiongroupdocumenttemplatepb.SubscriptionGroupDocumentTemplate, error)
	ListSubscriptionGroupDocumentTemplateBindings     func(context.Context, *subscriptiongroupdocumenttemplatepb.ListSubscriptionGroupDocumentTemplatesRequest) (*subscriptiongroupdocumenttemplatepb.ListSubscriptionGroupDocumentTemplatesResponse, error)
	DeleteSubscriptionGroupDocumentTemplateDraftPair  func(context.Context, string) (*documenttemplatepb.DocumentTemplate, error)
	DeleteSubscriptionGroupDocumentTemplateBinding    func(context.Context, *subscriptiongroupdocumenttemplatepb.DeleteSubscriptionGroupDocumentTemplateRequest) (*subscriptiongroupdocumenttemplatepb.DeleteSubscriptionGroupDocumentTemplateResponse, error)
	PublishSubscriptionGroupDocumentTemplateBinding   func(context.Context, *subscriptiongroupdocumenttemplatepb.PublishSubscriptionGroupDocumentTemplateRequest) (*subscriptiongroupdocumenttemplatepb.PublishSubscriptionGroupDocumentTemplateResponse, error)
}

// OutcomeSummaryModule holds all constructed outcome summary views.
type OutcomeSummaryModule struct {
	routes                         outcomesummarypkg.Routes
	subscriptionGroupExportEnabled bool
	List                           view.View
	SubscriptionGroup              view.View
	// SubscriptionGroupDownload is the HTMX drawer fragment for category, one period,
	// and output format selection. It remains report-read gated and separate
	// from subscription-group document template management permissions.
	SubscriptionGroupDownload view.View
	ClientCard                view.View
	JobSummary                view.View
	PhaseSummary              view.View
	// SubscriptionGroupExport is the group-grid CSV download (a raw handler — the
	// registrar wraps it with the same RBAC context injection as views).
	SubscriptionGroupExport http.HandlerFunc
	// ClientDocument is the per-client report-card .docx download (a raw
	// handler wrapped like SubscriptionGroupExport). Nil when GenerateDoc is not wired.
	ClientDocument http.HandlerFunc

	// Report-card template settings surface (TB3).
	TemplateSettings view.View
	TemplateUpload   view.View
	TemplatePublish  view.View
	TemplateDelete   view.View

	SubscriptionGroupDocumentTemplateSettings view.View
	SubscriptionGroupDocumentTemplateUpload   view.View
	SubscriptionGroupDocumentTemplatePublish  view.View
	SubscriptionGroupDocumentTemplateDelete   view.View
}

// NewOutcomeSummaryModule creates a new outcome summary module with all views wired.
func NewOutcomeSummaryModule(deps *OutcomeSummaryModuleDeps) *OutcomeSummaryModule {
	subscriptionGroupDeps := &subscriptiongroupview.Deps{
		Routes:                                   deps.Routes,
		Labels:                                   deps.Labels,
		CommonLabels:                             deps.CommonLabels,
		TableLabels:                              deps.TableLabels,
		Options:                                  deps.Options,
		ResolvePrincipalKind:                     deps.ResolvePrincipalKind,
		ListSubscriptionGroups:                   deps.ListSubscriptionGroups,
		ListSubscriptionGroupMembers:             deps.ListSubscriptionGroupMembers,
		ListJobs:                                 deps.ListJobs,
		ListJobTemplates:                         deps.ListJobTemplates,
		ListClients:                              deps.ListClients,
		ListJobOutcomeSummarys:                   deps.ListJobOutcomeSummarys,
		ListClientAttributes:                     deps.ListClientAttributes,
		ListAttributes:                           deps.ListAttributes,
		ResolveAttributeIDByCode:                 deps.ResolveAttributeIDByCode,
		GetSubscriptionGroupOutcomeExport:        deps.GetSubscriptionGroupOutcomeExport,
		ResolveSubscriptionGroupDocumentTemplate: deps.ResolveSubscriptionGroupDocumentTemplate,
		GeneratePDF:                              deps.GeneratePDF,
		ListSubscriptionGroupWorkspaceUsers:      deps.ListSubscriptionGroupWorkspaceUsers,
		ListWorkspaceUsers:                       deps.ListWorkspaceUsers,
		ListJobCategories:                        deps.ListJobCategories,
		// Non-enrolled-placeholder evidence walk (blanks untaken-elective floor
		// cells on the grid + CSV). Already injected for the DOCX handler.
		ListJobPhases:    deps.ListJobPhases,
		ListJobTasks:     deps.ListJobTasks,
		ListTaskOutcomes: deps.ListTaskOutcomes,
	}
	return &OutcomeSummaryModule{
		routes:                         deps.Routes,
		subscriptionGroupExportEnabled: deps.Options.SubscriptionGroupExportEnabled(),
		List: summarylist.NewView(&summarylist.ListViewDeps{
			Routes:                              deps.Routes,
			Labels:                              deps.Labels,
			CommonLabels:                        deps.CommonLabels,
			TableLabels:                         deps.TableLabels,
			Options:                             deps.Options,
			ResolvePrincipalKind:                deps.ResolvePrincipalKind,
			ListJobOutcomeSummarys:              deps.ListJobOutcomeSummarys,
			ListPriceSchedules:                  deps.ListPriceSchedules,
			ListSubscriptionGroups:              deps.ListSubscriptionGroups,
			ListJobTemplateSummaries:            deps.ListJobTemplateSummaries,
			ListSubscriptionGroupOutcomeLanding: deps.ListSubscriptionGroupOutcomeLanding,
			GetSubscriptionGroupOutcomeExport:   deps.GetSubscriptionGroupOutcomeExport,
			ListJobListTabSupport:               deps.ListJobListTabSupport,
			ListSubscriptionGroupMembers:        deps.ListSubscriptionGroupMembers,
			ListJobs:                            deps.ListJobs,
			// Group-visibility scoping (Options.List.ScopeByServicingGrant).
			ListWorkspaceUsers:                  deps.ListWorkspaceUsers,
			ListSubscriptionGroupWorkspaceUsers: deps.ListSubscriptionGroupWorkspaceUsers,
		}),
		SubscriptionGroup: subscriptiongroupview.NewView(subscriptionGroupDeps),
		SubscriptionGroupDownload: subscriptiongroupview.NewDownloadDrawer(&subscriptiongroupview.DrawerDeps{
			Routes:                            deps.Routes,
			Labels:                            deps.Labels,
			Options:                           deps.Options,
			ResolvePrincipalKind:              deps.ResolvePrincipalKind,
			GetSubscriptionGroupOutcomeExport: deps.GetSubscriptionGroupOutcomeExport,
		}),
		SubscriptionGroupExport: subscriptiongroupview.NewExportHandler(subscriptionGroupDeps),
		ClientDocument:          newClientDocumentHandler(deps),
		ClientCard: clientcard.NewView(&clientcard.Deps{
			Routes:               deps.Routes,
			Labels:               deps.Labels,
			CommonLabels:         deps.CommonLabels,
			TableLabels:          deps.TableLabels,
			Options:              deps.Options,
			ResolvePrincipalKind: deps.ResolvePrincipalKind,
			CategoryFilter:       deps.Options.CategoryFilter,
			// Client-card job-category banding (R9 W-A6, dedicated Options.ClientCard
			// — NOT Options.Row). BandByCategory groups the subject rows into native
			// job-category TableRowGroup bands; IncludeAllCategories lifts the card's
			// H2 academic-only filter FOR BANDING so deportment subjects appear under
			// their own band (the document/group paths keep H2 — separate fetches).
			// Both zero (service-admin / unset) → today's flat card, byte-identical.
			BandByCategory:                    deps.Options.ClientCard.BandByCategory(),
			IncludeAllCategories:              deps.Options.ClientCard.IncludeAllCategories,
			ListJobCategories:                 deps.ListJobCategories,
			GetSubscriptionGroupOutcomeExport: deps.GetSubscriptionGroupOutcomeExport,
			ListSubscriptionGroups:            deps.ListSubscriptionGroups,
			ListSubscriptionGroupMembers:      deps.ListSubscriptionGroupMembers,
			ListJobs:                          deps.ListJobs,
			ListJobTemplates:                  deps.ListJobTemplates,
			ListClients:                       deps.ListClients,
			ListJobOutcomeSummarys:            deps.ListJobOutcomeSummarys,
			ListPhaseOutcomeSummarysByJob:     deps.ListPhaseOutcomeSummarysByJob,
			ListJobPhases:                     deps.ListJobPhases,
			// Non-enrolled-placeholder evidence walk (blanks untaken-elective
			// floor grade cells). Already injected for the DOCX handler.
			ListJobTasks:     deps.ListJobTasks,
			ListTaskOutcomes: deps.ListTaskOutcomes,
		}),
		TemplateSettings: templatesettings.NewListView(templateSettingsDeps(deps)),
		TemplateUpload:   templatesettings.NewUploadAction(templateSettingsDeps(deps)),
		TemplatePublish:  templatesettings.NewPublishAction(templateSettingsDeps(deps)),
		TemplateDelete:   templatesettings.NewDeleteAction(templateSettingsDeps(deps)),
		SubscriptionGroupDocumentTemplateSettings: subscriptiongroupdocumenttemplate.NewListView(subscriptionGroupDocumentTemplateSettingsDeps(deps)),
		SubscriptionGroupDocumentTemplateUpload:   subscriptiongroupdocumenttemplate.NewUploadAction(subscriptionGroupDocumentTemplateSettingsDeps(deps)),
		SubscriptionGroupDocumentTemplatePublish:  subscriptiongroupdocumenttemplate.NewPublishAction(subscriptionGroupDocumentTemplateSettingsDeps(deps)),
		SubscriptionGroupDocumentTemplateDelete:   subscriptiongroupdocumenttemplate.NewDeleteAction(subscriptionGroupDocumentTemplateSettingsDeps(deps)),
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

func subscriptionGroupDocumentTemplateSettingsDeps(deps *OutcomeSummaryModuleDeps) *subscriptiongroupdocumenttemplate.Deps {
	return &subscriptiongroupdocumenttemplate.Deps{
		Routes:                 deps.Routes,
		Labels:                 deps.Labels,
		CommonLabels:           deps.CommonLabels,
		TableLabels:            deps.TableLabels,
		Options:                deps.Options,
		ListPriceSchedules:     deps.ListPriceSchedules,
		ListPlans:              deps.ListPlans,
		ListJobCategories:      deps.ListJobCategories,
		StoreTemplate:          deps.StoreSubscriptionGroupDocumentTemplate,
		DeleteTemplateObject:   deps.DeleteSubscriptionGroupDocumentTemplateObject,
		CreateUploadPair:       deps.CreateSubscriptionGroupDocumentTemplateUploadPair,
		ListTemplateBindings:   deps.ListSubscriptionGroupDocumentTemplateBindings,
		DeleteDraftPair:        deps.DeleteSubscriptionGroupDocumentTemplateDraftPair,
		PublishTemplateBinding: deps.PublishSubscriptionGroupDocumentTemplateBinding,
	}
}

// templateSettingsDeps maps the module deps onto the template-settings view
// deps (TB3). All closures are optional/nil-safe.
func templateSettingsDeps(deps *OutcomeSummaryModuleDeps) *templatesettings.Deps {
	return &templatesettings.Deps{
		Routes:                 deps.Routes,
		Labels:                 deps.Labels,
		CommonLabels:           deps.CommonLabels,
		TableLabels:            deps.TableLabels,
		ListPriceSchedules:     deps.ListPriceSchedules,
		UploadTemplate:         deps.UploadTemplate,
		ListDocumentTemplates:  deps.ListDocumentTemplates,
		CreateDocumentTemplate: deps.CreateDocumentTemplate,
		DeleteDocumentTemplate: deps.DeleteDocumentTemplate,
		ListTemplateBindings:   deps.ListTemplateBindings,
		CreateTemplateBinding:  deps.CreateTemplateBinding,
		DeleteTemplateBinding:  deps.DeleteTemplateBinding,
		PublishTemplateBinding: deps.PublishTemplateBinding,
	}
}

// newClientDocumentHandler builds the per-client report-card .docx download
// handler from the module deps. Returns nil when GenerateDoc is not wired (the
// app did not inject the fycha doctemplate closure) — RegisterRoutes then skips
// the route rather than registering a handler that would always 503.
func newClientDocumentHandler(deps *OutcomeSummaryModuleDeps) http.HandlerFunc {
	if deps.GenerateDoc == nil {
		return nil
	}
	return documentview.NewDownloadHandler(&documentview.Deps{
		Labels:                          deps.Labels,
		ResolvePrincipalKind:            deps.ResolvePrincipalKind,
		CommonLabels:                    deps.CommonLabels,
		DocumentHeaderName:              deps.DocumentHeaderName,
		CategoryFilter:                  deps.Options.CategoryFilter,
		DocOptions:                      deps.Options.Document,
		ListJobCategories:               deps.ListJobCategories,
		ListOutcomeCriterias:            deps.ListOutcomeCriterias,
		GetStaffListPageData:            deps.GetStaffListPageData,
		ListPriceSchedules:              deps.ListPriceSchedules,
		ListClientAttributes:            deps.ListClientAttributes,
		ResolveAttributeIDByCode:        deps.ResolveAttributeIDByCode,
		ListWorkspaceUsers:              deps.ListWorkspaceUsers,
		GenerateDoc:                     deps.GenerateDoc,
		GeneratePDF:                     deps.GeneratePDF,
		ResolveTemplateBytes:            deps.ResolveTemplateBytes,
		ListSubscriptionGroups:          deps.ListSubscriptionGroups,
		ListSubscriptionGroupMembers:    deps.ListSubscriptionGroupMembers,
		ListJobs:                        deps.ListJobs,
		ListJobTemplates:                deps.ListJobTemplates,
		ListClients:                     deps.ListClients,
		ListJobOutcomeSummarys:          deps.ListJobOutcomeSummarys,
		ListPhaseOutcomeSummarysByJob:   deps.ListPhaseOutcomeSummarysByJob,
		ListJobPhases:                   deps.ListJobPhases,
		GetPhaseApprovalGateRollup:      deps.GetPhaseApprovalGateRollup,
		ListJobTemplatePhasesByTemplate: deps.ListJobTemplatePhasesByTemplate,
		ListJobOutcomeLines:             deps.ListJobOutcomeLines,
		ListJobTasks:                    deps.ListJobTasks,
		ListTaskOutcomes:                deps.ListTaskOutcomes,
		ListCodedTaskOutcomeValuesByJob: deps.ListCodedTaskOutcomeValuesByJob,
		ListCodedTaskOutcomeValuesByJobHistorical: deps.ListCodedTaskOutcomeValuesByJobHistorical,
		ListTemplateTaskCriterias:                 deps.ListTemplateTaskCriterias,
	})
}

// RegisterRoutes registers all outcome summary routes.
func (m *OutcomeSummaryModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	// Activeness-scoped landing (/list/{scope}): the SAME list view, which reads
	// {scope} from the path and filters the price_schedule tabs by activeness
	// (current/past). Carries the same job_outcome_summary:list gate as ListURL
	// (the view checks it before any read). Distinct path depth from ListURL, so
	// no ServeMux collision.
	if m.routes.ListScopeURL != "" && m.routes.ListScopeURL != m.routes.ListURL {
		r.GET(m.routes.ListScopeURL, m.List)
	}
	if m.SubscriptionGroup != nil && m.routes.SubscriptionGroupURL != "" {
		r.GET(m.routes.SubscriptionGroupURL, m.SubscriptionGroup)
	}
	if m.subscriptionGroupExportEnabled && m.SubscriptionGroupDownload != nil && m.routes.SubscriptionGroupDownloadDrawerURL != "" {
		r.GET(m.routes.SubscriptionGroupDownloadDrawerURL, m.SubscriptionGroupDownload)
	}
	if m.ClientCard != nil && m.routes.ClientCardURL != "" {
		r.GET(m.routes.ClientCardURL, m.ClientCard)
	}
	if m.SubscriptionGroupExport != nil && m.routes.SubscriptionGroupExportURL != "" {
		// Raw (non-view) route — the registrar's HandleFunc path wraps it with
		// the ViewAdapter's RBAC context injection (WrapHandler), so the
		// handler's view.GetUserPermissions gate observes real permissions.
		if rr, ok := r.(interface {
			HandleFunc(method, path string, handler http.HandlerFunc, middlewares ...string)
		}); ok {
			rr.HandleFunc("GET", m.routes.SubscriptionGroupExportURL, m.SubscriptionGroupExport)
		} else {
			log.Printf("outcome summary: RouteRegistrar does not support HandleFunc — skipping GET %s", m.routes.SubscriptionGroupExportURL)
		}
	}
	if m.ClientDocument != nil && m.routes.ClientDocumentURL != "" {
		// Raw (non-view) route — the registrar's HandleFunc path wraps it with
		// the ViewAdapter's RBAC context injection (WrapHandler), exactly like
		// SubscriptionGroupExport, so the handler's view.GetUserPermissions gate observes
		// real permissions.
		if rr, ok := r.(interface {
			HandleFunc(method, path string, handler http.HandlerFunc, middlewares ...string)
		}); ok {
			rr.HandleFunc("GET", m.routes.ClientDocumentURL, m.ClientDocument)
		} else {
			log.Printf("outcome summary: RouteRegistrar does not support HandleFunc — skipping GET %s", m.routes.ClientDocumentURL)
		}
	}
	r.GET(m.routes.JobSummaryURL, m.JobSummary)
	r.GET(m.routes.PhaseSummaryURL, m.PhaseSummary)

	// Report-card template settings (TB3): list page + upload drawer (GET form /
	// POST create) + publish (POST) + delete (POST). Gated inside each view
	// (list → :list, mutations → :update).
	if m.TemplateSettings != nil && m.routes.TemplateSettingsURL != "" {
		r.GET(m.routes.TemplateSettingsURL, m.TemplateSettings)
	}
	if m.TemplateUpload != nil && m.routes.TemplateUploadURL != "" {
		r.GET(m.routes.TemplateUploadURL, m.TemplateUpload)
		r.POST(m.routes.TemplateUploadURL, m.TemplateUpload)
	}
	if m.TemplatePublish != nil && m.routes.TemplatePublishURL != "" {
		r.POST(m.routes.TemplatePublishURL, m.TemplatePublish)
	}
	if m.TemplateDelete != nil && m.routes.TemplateDeleteURL != "" {
		r.POST(m.routes.TemplateDeleteURL, m.TemplateDelete)
	}

	if m.subscriptionGroupExportEnabled {
		if m.SubscriptionGroupDocumentTemplateSettings != nil && m.routes.SubscriptionGroupDocumentTemplateSettingsURL != "" {
			r.GET(m.routes.SubscriptionGroupDocumentTemplateSettingsURL, m.SubscriptionGroupDocumentTemplateSettings)
		}
		if m.SubscriptionGroupDocumentTemplateUpload != nil && m.routes.SubscriptionGroupDocumentTemplateUploadURL != "" {
			r.GET(m.routes.SubscriptionGroupDocumentTemplateUploadURL, m.SubscriptionGroupDocumentTemplateUpload)
			r.POST(m.routes.SubscriptionGroupDocumentTemplateUploadURL, m.SubscriptionGroupDocumentTemplateUpload)
		}
		if m.SubscriptionGroupDocumentTemplatePublish != nil && m.routes.SubscriptionGroupDocumentTemplatePublishURL != "" {
			r.POST(m.routes.SubscriptionGroupDocumentTemplatePublishURL, m.SubscriptionGroupDocumentTemplatePublish)
		}
		if m.SubscriptionGroupDocumentTemplateDelete != nil && m.routes.SubscriptionGroupDocumentTemplateDeleteURL != "" {
			r.POST(m.routes.SubscriptionGroupDocumentTemplateDeleteURL, m.SubscriptionGroupDocumentTemplateDelete)
		}
	}
}
