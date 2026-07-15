package block

import (
	"context"
	"log"
	"sort"
	"time"

	fulfillmentdashboardview "github.com/erniealice/fayna-golang/domain/fulfillment/fulfillment/dashboard"
	job "github.com/erniealice/fayna-golang/domain/operation/job"
	jobdashboardview "github.com/erniealice/fayna-golang/domain/operation/job/dashboard"
	outcome_matrix "github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"
	outcome_summary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"

	"github.com/erniealice/espyna-golang/consumer"
	consumerapp "github.com/erniealice/espyna-golang/consumer/app"
	espynaports "github.com/erniealice/espyna-golang/ports"
	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	documenttemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/template"
	staffpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/staff"
	fulfillmentdashpb "github.com/erniealice/esqyma/pkg/schema/v1/service/dashboard/fulfillment"
	jobdashpb "github.com/erniealice/esqyma/pkg/schema/v1/service/dashboard/job"
)

// EngineOption configures the engine block from the consuming app (the app's
// view option block). Options are forwarded verbatim to AllUnits.
type EngineOption func(*engineConfig)

// engineConfig collects the per-unit view options an app may set.
type engineConfig struct {
	outcomeMatrixOptions  outcome_matrix.Options
	outcomeSummaryOptions outcome_summary.Options
	jobListOptions        job.Options
}

// WithOutcomeMatrixOptions sets the outcome-matrix row-presentation options
// (sort / description / group_by through "client_attributes.<code>" field
// references). See outcome_matrix.Options.
func WithOutcomeMatrixOptions(o outcome_matrix.Options) EngineOption {
	return func(c *engineConfig) { c.outcomeMatrixOptions = o }
}

// WithOutcomeSummaryOptions sets the outcome-summary presentation options for
// the report-cards surfaces: the view-1 tabstrip (Tab), what view-1 lists
// (List), and view-2 row bands/sort (Row). See outcome_summary.Options. The
// zero value renders the current flat job_outcome_summary list unchanged
// (backward-compatible for consumers that do not set it).
func WithOutcomeSummaryOptions(o outcome_summary.Options) EngineOption {
	return func(c *engineConfig) { c.outcomeSummaryOptions = o }
}

// WithJobListOptions sets the job list ("/classes") presentation options: the
// job_category tabstrip (Tab). See job.Options. The zero value renders the
// current flat job list unchanged (backward-compatible for consumers — e.g.
// service-admin — that do not set it).
func WithJobListOptions(o job.Options) EngineOption {
	return func(c *engineConfig) { c.jobListOptions = o }
}

// faynaEngineBlock returns a pyeza.AppOption that registers all fayna
// operation + fulfillment domain modules via the compose engine.
func EngineBlock(opts ...EngineOption) consumerapp.AppOption {
	return func(ctx *consumerapp.AppContext) error {
		uc, err := consumerapp.RequireUseCases(ctx, "faynaEngineBlock")
		if err != nil {
			return err
		}
		adapted := buildFaynaUseCases(uc)

		// Phase 5 (engine identity bridge): wire the WorkflowAssigneeQueryService
		// from the espyna container into fayna's Service.Workflow closure. The
		// container is passed through ctx.Container (pyeza.AppContext), set by
		// service-admin's buildAppContextBase. The query service is a read-only
		// port resolved at espyna container init (container.go wireAssigneeQuery);
		// nil when the postgres adapter is not loaded (mock_db builds).
		//
		// Identity inputs (workspace_user_id, workspace_id) are sourced from
		// session context here in the adapter closure — NEVER from the wire.
		// This is the Q-EIB-BRIDGE security invariant (OAD-1/OAD-2).
		if c, ok := ctx.Container.(*consumer.Container); ok && c != nil {
			if querySvc := c.GetWorkflowAssigneeQueryService(); querySvc != nil {
				adapted.Service.Workflow.ListPendingActivitiesForAssignee = func(
					reqCtx context.Context,
					req *WorkflowAssigneeQueryRequest,
				) (*WorkflowAssigneeQueryResponse, error) {
					// Source identity from context, not from the request struct.
					// The request struct carries Limit/Offset only; the identity
					// fields are populated here from the session context per the
					// Q-EIB-BRIDGE security contract.
					wsUserID := consumer.GetWorkspaceUserIDFromContext(reqCtx)
					wsID := consumer.GetWorkspaceIDFromContext(reqCtx)
					limit, offset := 0, 0
					if req != nil {
						limit = req.Limit
						offset = req.Offset
					}

					resp, err := querySvc.ListPendingActivitiesForAssignee(reqCtx, &espynaports.ListPendingActivitiesForAssigneeRequest{
						WorkspaceUserID: wsUserID,
						WorkspaceID:     wsID,
						Limit:           limit,
						Offset:          offset,
					})
					if err != nil {
						return nil, err
					}
					if resp == nil {
						return &WorkflowAssigneeQueryResponse{}, nil
					}
					return &WorkflowAssigneeQueryResponse{
						Activities: resp.Activities,
						Total:      resp.Total,
					}, nil
				}
				log.Printf("  compose: fayna engine — WorkflowAssigneeQueryService wired (engine identity bridge)")
			}
		}

		infra := &Infra{}
		infra.UploadFile, _ = ctx.UploadFile.(func(context.Context, string, string, []byte, string) error)
		infra.ListAttachments, _ = ctx.ListAttachments.(func(context.Context, string, string) (*attachmentpb.ListAttachmentsResponse, error))
		infra.CreateAttachment, _ = ctx.CreateAttachment.(func(context.Context, *attachmentpb.CreateAttachmentRequest) (*attachmentpb.CreateAttachmentResponse, error))
		infra.DeleteAttachment, _ = ctx.DeleteAttachment.(func(context.Context, *attachmentpb.DeleteAttachmentRequest) (*attachmentpb.DeleteAttachmentResponse, error))
		infra.NewAttachmentID, _ = ctx.NewAttachmentID.(func() string)
		if ctx.RefChecker != nil {
			if rc, ok := ctx.RefChecker.(espynaports.Checker); ok {
				infra.RefChecker = rc
			}
		}
		if db, ok := ctx.DB.(ListSimpler); ok {
			infra.DB = db
		}
		infra.GenerateDoc, _ = ctx.GenerateDoc.(func([]byte, map[string]any) ([]byte, error))
		infra.ResolveTemplateBytes, _ = ctx.ResolveTemplateBytes.(func(context.Context, string) ([]byte, error))
		// TB3 template settings artifact closures (upload + document_template CRUD).
		infra.UploadTemplate, _ = ctx.UploadTemplate.(func(context.Context, string, string, []byte, string) error)
		infra.ListDocTemplates, _ = ctx.ListDocTemplates.(func(context.Context, *documenttemplatepb.ListDocumentTemplatesRequest) (*documenttemplatepb.ListDocumentTemplatesResponse, error))
		infra.CreateDocTemplate, _ = ctx.CreateDocTemplate.(func(context.Context, *documenttemplatepb.CreateDocumentTemplateRequest) (*documenttemplatepb.CreateDocumentTemplateResponse, error))

		units := AllUnits(adapted, infra, opts...)
		return consumerapp.AssembleEngineBlock("fayna", units, ctx)
	}
}

// buildFaynaUseCases maps espyna's *consumer.UseCases to fayna block's typed
// shape (packages/fayna-golang/block/usecases.go). All sub-group wiring is
// nil-safe — when espyna has not wired a sub-domain the corresponding closure
// stays nil and fayna's RequireFor()/module nil-checks degrade gracefully.
//
// Phase 2 Round 2 (Q-WIRE-1): the two service-driven dashboard slots
// (Service.Dashboard.{Job,Fulfillment}) return the fayna VIEW types directly;
// the proto→view translation that the prior fayna/block/wiring.go performed via
// reflection now lives here, where both the proto Response and the fayna view
// Response are importable without a dependency cycle.
func buildFaynaUseCases(uc *consumer.UseCases) *UseCases {
	result := &UseCases{
		GetWorkspaceIDFromCtx: consumer.GetWorkspaceIDFromContext,
	}

	// -- Operation domain --------------------------------------------------------
	if uc.Operation != nil {
		op := uc.Operation

		if op.Job != nil {
			result.Operation.Job.CreateJob = op.Job.CreateJob.Execute
			result.Operation.Job.ReadJob = op.Job.ReadJob.Execute
			result.Operation.Job.UpdateJob = op.Job.UpdateJob.Execute
			result.Operation.Job.DeleteJob = op.Job.DeleteJob.Execute
			result.Operation.Job.ListJobs = op.Job.ListJobs.Execute
		}

		if op.JobPhase != nil {
			result.Operation.JobPhase.CreateJobPhase = op.JobPhase.CreateJobPhase.Execute
			result.Operation.JobPhase.ReadJobPhase = op.JobPhase.ReadJobPhase.Execute
			result.Operation.JobPhase.UpdateJobPhase = op.JobPhase.UpdateJobPhase.Execute
			result.Operation.JobPhase.DeleteJobPhase = op.JobPhase.DeleteJobPhase.Execute
			result.Operation.JobPhase.ListJobPhases = op.JobPhase.ListJobPhases.Execute
		}

		if op.JobTask != nil {
			result.Operation.JobTask.CreateJobTask = op.JobTask.CreateJobTask.Execute
			result.Operation.JobTask.ReadJobTask = op.JobTask.ReadJobTask.Execute
			result.Operation.JobTask.UpdateJobTask = op.JobTask.UpdateJobTask.Execute
			result.Operation.JobTask.DeleteJobTask = op.JobTask.DeleteJobTask.Execute
			result.Operation.JobTask.ListJobTasks = op.JobTask.ListJobTasks.Execute
			// espyna field is ListByPhase (Execute takes ListJobTasksByPhaseRequest).
			result.Operation.JobTask.ListJobTasksByPhase = op.JobTask.ListByPhase.Execute
		}

		if op.JobActivity != nil {
			result.Operation.JobActivity.GetJobActivityListPageData = op.JobActivity.GetJobActivityListPageData.Execute
			result.Operation.JobActivity.ReadJobActivity = op.JobActivity.ReadJobActivity.Execute
			result.Operation.JobActivity.CreateJobActivity = op.JobActivity.CreateJobActivity.Execute
			result.Operation.JobActivity.UpdateJobActivity = op.JobActivity.UpdateJobActivity.Execute
			result.Operation.JobActivity.DeleteJobActivity = op.JobActivity.DeleteJobActivity.Execute
			result.Operation.JobActivity.ListJobActivities = op.JobActivity.ListJobActivities.Execute
			result.Operation.JobActivity.GetJobActivityRollup = op.JobActivity.GetJobActivityRollup.Execute
			result.Operation.JobActivity.SubmitForApproval = op.JobActivity.SubmitForApproval.Execute
			result.Operation.JobActivity.ApproveActivity = op.JobActivity.ApproveActivity.Execute
			result.Operation.JobActivity.RejectActivity = op.JobActivity.RejectActivity.Execute
		}

		if op.JobTemplate != nil {
			result.Operation.JobTemplate.CreateJobTemplate = op.JobTemplate.CreateJobTemplate.Execute
			result.Operation.JobTemplate.ReadJobTemplate = op.JobTemplate.ReadJobTemplate.Execute
			result.Operation.JobTemplate.UpdateJobTemplate = op.JobTemplate.UpdateJobTemplate.Execute
			result.Operation.JobTemplate.DeleteJobTemplate = op.JobTemplate.DeleteJobTemplate.Execute
			result.Operation.JobTemplate.GetJobTemplateListPageData = op.JobTemplate.GetJobTemplateListPageData.Execute
			result.Operation.JobTemplate.ListJobTemplates = op.JobTemplate.ListJobTemplates.Execute
		}

		// JobCategory — the "/classes" tab-split reads ListJobCategories (one tab
		// per category). Optional/nil-safe: a nil aggregate leaves the closures
		// nil → the list renders flat.
		if op.JobCategory != nil {
			result.Operation.JobCategory.CreateJobCategory = op.JobCategory.CreateJobCategory.Execute
			result.Operation.JobCategory.ReadJobCategory = op.JobCategory.ReadJobCategory.Execute
			result.Operation.JobCategory.UpdateJobCategory = op.JobCategory.UpdateJobCategory.Execute
			result.Operation.JobCategory.DeleteJobCategory = op.JobCategory.DeleteJobCategory.Execute
			result.Operation.JobCategory.ListJobCategories = op.JobCategory.ListJobCategories.Execute
		}

		// JobOutcomeSummaryDocumentTemplate — report-card template binding (TB3).
		// Backs the template settings page (list/upload/publish/delete). Optional/
		// nil-safe: a nil aggregate leaves the closures nil → "not configured".
		if op.JobOutcomeSummaryDocumentTemplate != nil {
			b := op.JobOutcomeSummaryDocumentTemplate
			result.Operation.JobOutcomeSummaryDocumentTemplate.ListJobOutcomeSummaryDocumentTemplates = b.ListJobOutcomeSummaryDocumentTemplates.Execute
			result.Operation.JobOutcomeSummaryDocumentTemplate.CreateJobOutcomeSummaryDocumentTemplate = b.CreateJobOutcomeSummaryDocumentTemplate.Execute
			result.Operation.JobOutcomeSummaryDocumentTemplate.DeleteJobOutcomeSummaryDocumentTemplate = b.DeleteJobOutcomeSummaryDocumentTemplate.Execute
			result.Operation.JobOutcomeSummaryDocumentTemplate.PublishJobOutcomeSummaryDocumentTemplate = b.PublishJobOutcomeSummaryDocumentTemplate.Execute
		}

		if op.JobTemplatePhase != nil {
			result.Operation.JobTemplatePhase.CreateJobTemplatePhase = op.JobTemplatePhase.CreateJobTemplatePhase.Execute
			result.Operation.JobTemplatePhase.ReadJobTemplatePhase = op.JobTemplatePhase.ReadJobTemplatePhase.Execute
			result.Operation.JobTemplatePhase.UpdateJobTemplatePhase = op.JobTemplatePhase.UpdateJobTemplatePhase.Execute
			result.Operation.JobTemplatePhase.DeleteJobTemplatePhase = op.JobTemplatePhase.DeleteJobTemplatePhase.Execute
			result.Operation.JobTemplatePhase.ListByJobTemplate = op.JobTemplatePhase.ListByJobTemplate.Execute
		}

		if op.JobTemplateTask != nil {
			result.Operation.JobTemplateTask.CreateJobTemplateTask = op.JobTemplateTask.CreateJobTemplateTask.Execute
			result.Operation.JobTemplateTask.ReadJobTemplateTask = op.JobTemplateTask.ReadJobTemplateTask.Execute
			result.Operation.JobTemplateTask.UpdateJobTemplateTask = op.JobTemplateTask.UpdateJobTemplateTask.Execute
			result.Operation.JobTemplateTask.DeleteJobTemplateTask = op.JobTemplateTask.DeleteJobTemplateTask.Execute
			result.Operation.JobTemplateTask.ListByPhase = op.JobTemplateTask.ListByPhase.Execute
		}

		if op.TemplateTaskCriteria != nil {
			result.Operation.TemplateTaskCriteria.CreateTemplateTaskCriteria = op.TemplateTaskCriteria.CreateTemplateTaskCriteria.Execute
			result.Operation.TemplateTaskCriteria.ReadTemplateTaskCriteria = op.TemplateTaskCriteria.ReadTemplateTaskCriteria.Execute
			result.Operation.TemplateTaskCriteria.UpdateTemplateTaskCriteria = op.TemplateTaskCriteria.UpdateTemplateTaskCriteria.Execute
			result.Operation.TemplateTaskCriteria.DeleteTemplateTaskCriteria = op.TemplateTaskCriteria.DeleteTemplateTaskCriteria.Execute
			result.Operation.TemplateTaskCriteria.ListTemplateTaskCriterias = op.TemplateTaskCriteria.ListTemplateTaskCriteria.Execute
			result.Operation.TemplateTaskCriteria.ListByTemplateTask = op.TemplateTaskCriteria.ListByTemplateTask.Execute
		}

		if op.OutcomeCriteria != nil {
			result.Operation.OutcomeCriteria.CreateOutcomeCriteria = op.OutcomeCriteria.CreateOutcomeCriteria.Execute
			result.Operation.OutcomeCriteria.ReadOutcomeCriteria = op.OutcomeCriteria.ReadOutcomeCriteria.Execute
			result.Operation.OutcomeCriteria.UpdateOutcomeCriteria = op.OutcomeCriteria.UpdateOutcomeCriteria.Execute
			result.Operation.OutcomeCriteria.DeleteOutcomeCriteria = op.OutcomeCriteria.DeleteOutcomeCriteria.Execute
			// espyna field is ListOutcomeCriteria (Execute returns ListOutcomeCriteriasResponse).
			result.Operation.OutcomeCriteria.ListOutcomeCriterias = op.OutcomeCriteria.ListOutcomeCriteria.Execute
		}

		if op.TaskOutcome != nil {
			result.Operation.TaskOutcome.CreateTaskOutcome = op.TaskOutcome.CreateTaskOutcome.Execute
			result.Operation.TaskOutcome.ReadTaskOutcome = op.TaskOutcome.ReadTaskOutcome.Execute
			result.Operation.TaskOutcome.UpdateTaskOutcome = op.TaskOutcome.UpdateTaskOutcome.Execute
			result.Operation.TaskOutcome.DeleteTaskOutcome = op.TaskOutcome.DeleteTaskOutcome.Execute
			result.Operation.TaskOutcome.ListTaskOutcomes = op.TaskOutcome.ListTaskOutcomes.Execute
		}

		if op.JobOutcomeSummary != nil {
			result.Operation.JobOutcomeSummary.GetByJob = op.JobOutcomeSummary.GetByJob.Execute
			result.Operation.JobOutcomeSummary.ListJobOutcomeSummaries = op.JobOutcomeSummary.ListJobOutcomeSummaries.Execute
		}

		if op.PhaseOutcomeSummary != nil {
			result.Operation.PhaseOutcomeSummary.GetByJobPhase = op.PhaseOutcomeSummary.GetByJobPhase.Execute
			result.Operation.PhaseOutcomeSummary.ListByJob = op.PhaseOutcomeSummary.ListByJob.Execute
		}

		// -- Education grading (20260616 v1) — single-repo CRUD entities -----------
		if op.ScoringScheme != nil {
			result.Operation.ScoringScheme.CreateScoringScheme = op.ScoringScheme.CreateScoringScheme.Execute
			result.Operation.ScoringScheme.ReadScoringScheme = op.ScoringScheme.ReadScoringScheme.Execute
			result.Operation.ScoringScheme.UpdateScoringScheme = op.ScoringScheme.UpdateScoringScheme.Execute
			result.Operation.ScoringScheme.DeleteScoringScheme = op.ScoringScheme.DeleteScoringScheme.Execute
			result.Operation.ScoringScheme.ListScoringSchemes = op.ScoringScheme.ListScoringSchemes.Execute
		}

		if op.ScoringComponent != nil {
			result.Operation.ScoringComponent.CreateScoringComponent = op.ScoringComponent.CreateScoringComponent.Execute
			result.Operation.ScoringComponent.ReadScoringComponent = op.ScoringComponent.ReadScoringComponent.Execute
			result.Operation.ScoringComponent.UpdateScoringComponent = op.ScoringComponent.UpdateScoringComponent.Execute
			result.Operation.ScoringComponent.DeleteScoringComponent = op.ScoringComponent.DeleteScoringComponent.Execute
			result.Operation.ScoringComponent.ListScoringComponents = op.ScoringComponent.ListScoringComponents.Execute
		}

		if op.ScoringComponentCriteria != nil {
			result.Operation.ScoringComponentCriteria.CreateScoringComponentCriteria = op.ScoringComponentCriteria.CreateScoringComponentCriteria.Execute
			result.Operation.ScoringComponentCriteria.ReadScoringComponentCriteria = op.ScoringComponentCriteria.ReadScoringComponentCriteria.Execute
			result.Operation.ScoringComponentCriteria.UpdateScoringComponentCriteria = op.ScoringComponentCriteria.UpdateScoringComponentCriteria.Execute
			result.Operation.ScoringComponentCriteria.DeleteScoringComponentCriteria = op.ScoringComponentCriteria.DeleteScoringComponentCriteria.Execute
			result.Operation.ScoringComponentCriteria.ListScoringComponentCriterias = op.ScoringComponentCriteria.ListScoringComponentCriterias.Execute
		}

		if op.ScoreScale != nil {
			result.Operation.ScoreScale.CreateScoreScale = op.ScoreScale.CreateScoreScale.Execute
			result.Operation.ScoreScale.ReadScoreScale = op.ScoreScale.ReadScoreScale.Execute
			result.Operation.ScoreScale.UpdateScoreScale = op.ScoreScale.UpdateScoreScale.Execute
			result.Operation.ScoreScale.DeleteScoreScale = op.ScoreScale.DeleteScoreScale.Execute
			result.Operation.ScoreScale.ListScoreScales = op.ScoreScale.ListScoreScales.Execute
		}

		if op.ScoreScaleBand != nil {
			result.Operation.ScoreScaleBand.CreateScoreScaleBand = op.ScoreScaleBand.CreateScoreScaleBand.Execute
			result.Operation.ScoreScaleBand.ReadScoreScaleBand = op.ScoreScaleBand.ReadScoreScaleBand.Execute
			result.Operation.ScoreScaleBand.UpdateScoreScaleBand = op.ScoreScaleBand.UpdateScoreScaleBand.Execute
			result.Operation.ScoreScaleBand.DeleteScoreScaleBand = op.ScoreScaleBand.DeleteScoreScaleBand.Execute
			result.Operation.ScoreScaleBand.ListScoreScaleBands = op.ScoreScaleBand.ListScoreScaleBands.Execute
		}

		if op.JobOutcomeLine != nil {
			result.Operation.JobOutcomeLine.CreateJobOutcomeLine = op.JobOutcomeLine.CreateJobOutcomeLine.Execute
			result.Operation.JobOutcomeLine.ReadJobOutcomeLine = op.JobOutcomeLine.ReadJobOutcomeLine.Execute
			result.Operation.JobOutcomeLine.UpdateJobOutcomeLine = op.JobOutcomeLine.UpdateJobOutcomeLine.Execute
			result.Operation.JobOutcomeLine.DeleteJobOutcomeLine = op.JobOutcomeLine.DeleteJobOutcomeLine.Execute
			result.Operation.JobOutcomeLine.ListJobOutcomeLines = op.JobOutcomeLine.ListJobOutcomeLines.Execute
		}

		if op.ReportingCheckpoint != nil {
			result.Operation.ReportingCheckpoint.CreateReportingCheckpoint = op.ReportingCheckpoint.CreateReportingCheckpoint.Execute
			result.Operation.ReportingCheckpoint.ReadReportingCheckpoint = op.ReportingCheckpoint.ReadReportingCheckpoint.Execute
			result.Operation.ReportingCheckpoint.UpdateReportingCheckpoint = op.ReportingCheckpoint.UpdateReportingCheckpoint.Execute
			result.Operation.ReportingCheckpoint.DeleteReportingCheckpoint = op.ReportingCheckpoint.DeleteReportingCheckpoint.Execute
			result.Operation.ReportingCheckpoint.ListReportingCheckpoints = op.ReportingCheckpoint.ListReportingCheckpoints.Execute
		}
	}

	// -- Fulfillment domain ------------------------------------------------------
	if uc.Fulfillment != nil {
		ff := uc.Fulfillment
		result.Fulfillment.GetFulfillmentListPageData = ff.GetFulfillmentListPageData.Execute
		result.Fulfillment.GetFulfillmentItemPageData = ff.GetFulfillmentItemPageData.Execute
		result.Fulfillment.CreateFulfillment = ff.CreateFulfillment.Execute
		result.Fulfillment.UpdateFulfillment = ff.UpdateFulfillment.Execute
		result.Fulfillment.DeleteFulfillment = ff.DeleteFulfillment.Execute
		result.Fulfillment.TransitionStatus = ff.TransitionStatus.Execute
	}

	// -- Subscription (cross-domain Job origin breadcrumb; optional) -------------
	if uc.Subscription != nil && uc.Subscription.Subscription != nil {
		result.Subscription.Subscription.ReadSubscription = uc.Subscription.Subscription.ReadSubscription.Execute
	}

	// -- Subscription (template-grain delivery summary deps; education tier's
	// job list — optional; nil → the group/deliverer/schedule columns render
	// blank) -----------------------------------------------------------------
	if uc.Subscription != nil {
		if uc.Subscription.SubscriptionSeat != nil {
			result.Subscription.SubscriptionSeat.GetSubscriptionSeatListPageData = uc.Subscription.SubscriptionSeat.GetSubscriptionSeatListPageData.Execute
		}
		if uc.Subscription.SubscriptionGroup != nil {
			result.Subscription.SubscriptionGroup.ListSubscriptionGroups = uc.Subscription.SubscriptionGroup.ListSubscriptionGroups.Execute
		}
		if uc.Subscription.SubscriptionGroupMember != nil {
			result.Subscription.SubscriptionGroupMember.ListSubscriptionGroupMembers = uc.Subscription.SubscriptionGroupMember.ListSubscriptionGroupMembers.Execute
		}
		if uc.Subscription.SubscriptionGroupWorkspaceUser != nil {
			result.Subscription.SubscriptionGroupWorkspaceUser.ListSubscriptionGroupWorkspaceUsers = uc.Subscription.SubscriptionGroupWorkspaceUser.ListSubscriptionGroupWorkspaceUsers.Execute
		}
		// PriceSchedule list backs the report-cards view-1 tabstrip (one tab per
		// price_schedule row, incl. inactive — Q-TAB-1). Optional/nil-safe.
		if uc.Subscription.PriceSchedule != nil {
			result.Subscription.PriceSchedule.ListPriceSchedules = uc.Subscription.PriceSchedule.ListPriceSchedules.Execute
		}
	}

	// -- Product (cross-domain; template-grain delivery summary's deliverer
	// resolution — optional; nil → deliverer column renders blank) -----------
	if uc.Product != nil && uc.Product.ProductPlan != nil {
		result.Product.ProductPlan.ListProductPlans = uc.Product.ProductPlan.ListProductPlans.Execute
	}

	// -- Entity (drawer search pickers; optional → flat-filter fallback) ---------
	if uc.Entity != nil {
		if uc.Entity.Client != nil {
			result.Entity.Client.SearchClientsByName = uc.Entity.Client.SearchClientsByName.Execute
			result.Entity.Client.ListClients = uc.Entity.Client.ListClients.Execute
		}
		if uc.Entity.ClientAttribute != nil && uc.Entity.ClientAttribute.ListClientAttributes != nil {
			result.Entity.ClientAttribute.ListClientAttributes = uc.Entity.ClientAttribute.ListClientAttributes.Execute
		}
		if uc.Entity.Staff != nil {
			result.Entity.Staff.ListStaffs = uc.Entity.Staff.ListStaffs.Execute
			result.Entity.Staff.GetStaffListPageData = uc.Entity.Staff.GetStaffListPageData.Execute
		}
		if uc.Entity.WorkspaceUser != nil {
			result.Entity.WorkspaceUser.ListWorkspaceUsers = uc.Entity.WorkspaceUser.ListWorkspaceUsers.Execute
		}
	}

	// The attribute code→id resolver behind the outcome-matrix row options
	// ("client_attributes.<code>"). Lives on espyna's Common aggregate.
	if uc.Common != nil && uc.Common.Attribute != nil {
		result.Entity.ClientAttribute.ResolveAttributeIDByCode = uc.Common.Attribute.ReadAttributeByCode
	}

	// -- Service/operation OutcomeMatrix (generic grading grid) ------------------
	//
	// The read use case lives on espyna's SERVICE aggregate (service/operation/
	// outcome_matrix), NOT the Operation aggregate — hence it is sourced from
	// uc.Service.OutcomeMatrix here while surfacing on fayna's Operation group
	// (the view module lives in fayna domain/operation/outcome_matrix). Nil-safe:
	// a nil Service / OutcomeMatrix leaves the closure nil → the grid renders empty.
	if uc.Service != nil && uc.Service.OutcomeMatrix != nil &&
		uc.Service.OutcomeMatrix.GetOutcomeMatrix != nil {
		result.Operation.OutcomeMatrix.GetOutcomeMatrix = uc.Service.OutcomeMatrix.GetOutcomeMatrix.Execute
	}
	// ResolveStaff maps the session user → active staff_id through the typed staff
	// list use case (the read-only gate + record-action IDOR guard authority).
	// Wired only when the staff list closure is present; "" fails those gates closed.
	if result.Entity.Staff.ListStaffs != nil {
		result.Operation.OutcomeMatrix.ResolveStaff = newStaffResolver(result.Entity.Staff.ListStaffs)
	}

	// -- Service/operation JobTemplateSummary (generic template-grain summary) ----
	//
	// Like OutcomeMatrix, the read use case lives on espyna's SERVICE aggregate
	// (service/operation/job_template_summary), sourced here from
	// uc.Service.JobTemplateSummary while surfacing on fayna's Operation group
	// (the consuming view is the Job list module). Nil-safe: a nil Service /
	// JobTemplateSummary leaves the closure nil → the education-tier list renders
	// empty.
	if uc.Service != nil && uc.Service.JobTemplateSummary != nil &&
		uc.Service.JobTemplateSummary.ListJobTemplateSummaries != nil {
		result.Operation.JobTemplateSummary.ListJobTemplateSummaries = uc.Service.JobTemplateSummary.ListJobTemplateSummaries.Execute
	}

	// -- Service.Dashboard.Job — proto→view translation --------------------------
	//
	// Proto Response carries *JobStats (pointer-to-struct), TrendLabels/Values,
	// UpcomingDeadlines ([]*jobpb.Job), RecentActivity ([]*jobactivitypb.JobActivity),
	// and RiskTopRows []*JobRiskRow (DateEndMillis int64). The view-layer Response
	// wants value-type JobRiskRow with DateEnd time.Time. Translation here, where
	// both proto + view types are importable (the prior reflection in
	// fayna/block/wiring.go is gone).
	if uc.Service != nil && uc.Service.Dashboard != nil &&
		uc.Service.Dashboard.Job != nil && uc.Service.Dashboard.Job.GetJobDashboard != nil {
		jobDash := uc.Service.Dashboard.Job.GetJobDashboard
		result.Service.Dashboard.Job = func(ctx context.Context, req *jobdashboardview.Request) (*jobdashboardview.Response, error) {
			workspaceID := ""
			var now time.Time
			if req != nil {
				workspaceID = req.WorkspaceID
				now = req.Now
			}
			// Empty-workspace fallback: postgres dashboard queries treat empty
			// workspace as "no filter" and would otherwise render cross-workspace
			// aggregates.
			if workspaceID == "" {
				workspaceID = consumer.GetWorkspaceIDFromContext(ctx)
			}
			if now.IsZero() {
				now = time.Now()
			}
			nowMillis := now.UnixMilli()
			resp, err := jobDash.Execute(ctx, &jobdashpb.GetJobDashboardRequest{
				WorkspaceId: workspaceID,
				NowMillis:   &nowMillis,
			})
			if err != nil {
				return nil, err
			}
			if resp == nil {
				return nil, nil
			}
			out := &jobdashboardview.Response{
				ActiveJobs:        resp.GetStats().GetActiveJobs(),
				DoneThisMonth:     resp.GetStats().GetDoneThisMonth(),
				OverdueJobs:       resp.GetStats().GetOverdueJobs(),
				HoursThisWeek:     resp.GetStats().GetHoursThisWeek(),
				TrendLabels:       resp.GetTrendLabels(),
				TrendValues:       resp.GetTrendValues(),
				UpcomingDeadlines: resp.GetUpcomingDeadlines(),
				RecentActivity:    resp.GetRecentActivity(),
			}
			for _, r := range resp.GetRiskTopRows() {
				if r == nil {
					continue
				}
				row := jobdashboardview.JobRiskRow{
					JobID:         r.GetJobId(),
					Code:          r.GetCode(),
					Name:          r.GetName(),
					CompletionPct: r.GetCompletionPct(),
				}
				if ms := r.GetDateEndMillis(); ms != 0 {
					row.DateEnd = time.UnixMilli(ms).UTC()
				}
				out.RiskTopRows = append(out.RiskTopRows, row)
			}
			return out, nil
		}
	}

	// -- Service.Dashboard.Fulfillment — proto→view translation ------------------
	//
	// Proto Response carries *FulfillmentStats (pointer-to-struct) plus the slice
	// fields. Same empty-workspace fallback as the Job dashboard.
	if uc.Service != nil && uc.Service.Dashboard != nil &&
		uc.Service.Dashboard.Fulfillment != nil && uc.Service.Dashboard.Fulfillment.GetFulfillmentDashboard != nil {
		ffDash := uc.Service.Dashboard.Fulfillment.GetFulfillmentDashboard
		result.Service.Dashboard.Fulfillment = func(ctx context.Context, req *fulfillmentdashboardview.Request) (*fulfillmentdashboardview.Response, error) {
			workspaceID := ""
			var now time.Time
			if req != nil {
				workspaceID = req.WorkspaceID
				now = req.Now
			}
			if workspaceID == "" {
				workspaceID = consumer.GetWorkspaceIDFromContext(ctx)
			}
			if now.IsZero() {
				now = time.Now()
			}
			nowMillis := now.UnixMilli()
			resp, err := ffDash.Execute(ctx, &fulfillmentdashpb.GetFulfillmentDashboardRequest{
				WorkspaceId: workspaceID,
				NowMillis:   &nowMillis,
			})
			if err != nil {
				return nil, err
			}
			if resp == nil {
				return nil, nil
			}
			return &fulfillmentdashboardview.Response{
				Pending:          resp.GetStats().GetPending(),
				InTransit:        resp.GetStats().GetInTransit(),
				DeliveredToday:   resp.GetStats().GetDeliveredToday(),
				Exceptions:       resp.GetStats().GetExceptions(),
				AvgFulfillDays:   resp.GetStats().GetAvgFulfillDays(),
				StatusMixLabels:  resp.GetStatusMixLabels(),
				StatusMixValues:  resp.GetStatusMixValues(),
				TrendLabels:      resp.GetTrendLabels(),
				TrendValues:      resp.GetTrendValues(),
				RecentExceptions: resp.GetRecentExceptions(),
			}, nil
		}
	}

	return result
}

// newStaffResolver returns a ResolveStaff closure mapping the acting session
// user to their active staff_id via the typed staff list use case (never raw
// SQL). Returns "" when there is no authenticated user or no matching active
// staff — a fail-closed identity the outcome-matrix read-only gate and the
// record-action IDOR guard both treat as "cannot edit". The user_id filter
// narrows the query server-side; the result is re-verified client-side because
// the resolved staff_id is a security identity (the write-ownership axis) and
// must never be a row the filter did not actually match to this user.
func newStaffResolver(
	listStaffs func(context.Context, *staffpb.ListStaffsRequest) (*staffpb.ListStaffsResponse, error),
) func(context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		userID := consumer.GetUserIDFromContext(ctx)
		if userID == "" || listStaffs == nil {
			return "", nil
		}
		resp, err := listStaffs(ctx, &staffpb.ListStaffsRequest{
			Filters: &commonpb.FilterRequest{
				Filters: []*commonpb.TypedFilter{{
					Field: "user_id",
					FilterType: &commonpb.TypedFilter_StringFilter{
						StringFilter: &commonpb.StringFilter{
							Value:    userID,
							Operator: commonpb.StringOperator_STRING_EQUALS,
						},
					},
				}},
			},
		})
		if err != nil {
			return "", err
		}
		// Collect every active staff row that re-verifies to this user, then
		// select deterministically (stable sort by staff id ascending). A user may
		// hold multiple staff rows (1:1 today, but not guaranteed); the resolved
		// staff_id is a security identity (the write-ownership axis), so
		// multi-staff-per-user selection MUST NOT depend on the adapter's row
		// return order.
		var staffIDs []string
		for _, s := range resp.GetData() {
			if s.GetUserId() == userID && s.GetActive() {
				staffIDs = append(staffIDs, s.GetId())
			}
		}
		if len(staffIDs) == 0 {
			return "", nil
		}
		sort.Strings(staffIDs)
		return staffIDs[0], nil
	}
}
