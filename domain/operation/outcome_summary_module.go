package operation

import (
	"context"

	outcomesummarypkg "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	jobsummary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary/job_summary"
	summarylist "github.com/erniealice/fayna-golang/domain/operation/outcome_summary/list"
	phasesummary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary/phase_summary"
	sectionview "github.com/erniealice/fayna-golang/domain/operation/outcome_summary/section"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	clientattributepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client_attribute"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	phasesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	priceschedulepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/price_schedule"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
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

	// Phase outcome summary operations
	GetPhaseOutcomeSummaryByJobPhase func(ctx context.Context, req *phasesumpb.GetPhaseOutcomeSummaryByJobPhaseRequest) (*phasesumpb.GetPhaseOutcomeSummaryByJobPhaseResponse, error)
	ListPhaseOutcomeSummarysByJob    func(ctx context.Context, req *phasesumpb.ListPhaseOutcomeSummarysByJobRequest) (*phasesumpb.ListPhaseOutcomeSummarysByJobResponse, error)

	// Report-cards navigation deps (view-1 landing + view-2 section grid). All
	// optional/nil-safe: a nil closure degrades the affected surface to its
	// empty/flat state, never a panic.
	ListPriceSchedules           func(ctx context.Context, req *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error)
	ListSubscriptionGroups       func(ctx context.Context, req *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error)
	ListSubscriptionGroupMembers func(ctx context.Context, req *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error)
	ListJobs                     func(ctx context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error)
	ListJobTemplates             func(ctx context.Context, req *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error)
	ListClients                  func(ctx context.Context, req *clientpb.ListClientsRequest) (*clientpb.ListClientsResponse, error)
	ListClientAttributes         func(ctx context.Context, req *clientattributepb.ListClientAttributesRequest) (*clientattributepb.ListClientAttributesResponse, error)
	ResolveAttributeIDByCode     func(ctx context.Context, code string) (string, error)
	ListJobTemplateSummaries     func(ctx context.Context, req *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error)
}

// OutcomeSummaryModule holds all constructed outcome summary views.
type OutcomeSummaryModule struct {
	routes       outcomesummarypkg.Routes
	List         view.View
	Section      view.View
	JobSummary   view.View
	PhaseSummary view.View
}

// NewOutcomeSummaryModule creates a new outcome summary module with all views wired.
func NewOutcomeSummaryModule(deps *OutcomeSummaryModuleDeps) *OutcomeSummaryModule {
	return &OutcomeSummaryModule{
		routes: deps.Routes,
		List: summarylist.NewView(&summarylist.ListViewDeps{
			Routes:                   deps.Routes,
			Labels:                   deps.Labels,
			CommonLabels:             deps.CommonLabels,
			TableLabels:              deps.TableLabels,
			Options:                  deps.Options,
			ListJobOutcomeSummarys:   deps.ListJobOutcomeSummarys,
			ListPriceSchedules:       deps.ListPriceSchedules,
			ListSubscriptionGroups:   deps.ListSubscriptionGroups,
			ListJobTemplateSummaries: deps.ListJobTemplateSummaries,
		}),
		Section: sectionview.NewView(&sectionview.Deps{
			Routes:                       deps.Routes,
			Labels:                       deps.Labels,
			CommonLabels:                 deps.CommonLabels,
			TableLabels:                  deps.TableLabels,
			Options:                      deps.Options,
			ListSubscriptionGroups:       deps.ListSubscriptionGroups,
			ListSubscriptionGroupMembers: deps.ListSubscriptionGroupMembers,
			ListJobs:                     deps.ListJobs,
			ListJobTemplates:             deps.ListJobTemplates,
			ListClients:                  deps.ListClients,
			ListJobOutcomeSummarys:       deps.ListJobOutcomeSummarys,
			ListClientAttributes:         deps.ListClientAttributes,
			ResolveAttributeIDByCode:     deps.ResolveAttributeIDByCode,
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

// RegisterRoutes registers all outcome summary routes.
func (m *OutcomeSummaryModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	if m.Section != nil && m.routes.SectionURL != "" {
		r.GET(m.routes.SectionURL, m.Section)
	}
	r.GET(m.routes.JobSummaryURL, m.JobSummary)
	r.GET(m.routes.PhaseSummaryURL, m.PhaseSummary)
}
