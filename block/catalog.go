package block

import (
	"log"
	"net/http"

	"github.com/erniealice/espyna-golang/consumer/compose"
	fulfillmentdomain "github.com/erniealice/fayna-golang/domain/fulfillment"
	fulfillmentpkg "github.com/erniealice/fayna-golang/domain/fulfillment/fulfillment"
	operation "github.com/erniealice/fayna-golang/domain/operation"
	"github.com/erniealice/fayna-golang/domain/operation/activity_expense"
	"github.com/erniealice/fayna-golang/domain/operation/activity_labor"
	"github.com/erniealice/fayna-golang/domain/operation/activity_material"
	"github.com/erniealice/fayna-golang/domain/operation/evaluation"
	"github.com/erniealice/fayna-golang/domain/operation/evaluation_cycle"
	"github.com/erniealice/fayna-golang/domain/operation/evaluation_cycle_member"
	"github.com/erniealice/fayna-golang/domain/operation/evaluation_template"
	"github.com/erniealice/fayna-golang/domain/operation/evaluation_template_item"
	"github.com/erniealice/fayna-golang/domain/operation/job"
	"github.com/erniealice/fayna-golang/domain/operation/job_activity"
	"github.com/erniealice/fayna-golang/domain/operation/job_outcome_line"
	"github.com/erniealice/fayna-golang/domain/operation/job_phase"
	"github.com/erniealice/fayna-golang/domain/operation/job_task"
	"github.com/erniealice/fayna-golang/domain/operation/job_template"
	"github.com/erniealice/fayna-golang/domain/operation/job_template_phase"
	"github.com/erniealice/fayna-golang/domain/operation/job_template_task"
	"github.com/erniealice/fayna-golang/domain/operation/outcome_criteria"
	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"
	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/fayna-golang/domain/operation/performance"
	"github.com/erniealice/fayna-golang/domain/operation/reporting_checkpoint"
	"github.com/erniealice/fayna-golang/domain/operation/score_scale"
	"github.com/erniealice/fayna-golang/domain/operation/score_scale_band"
	"github.com/erniealice/fayna-golang/domain/operation/scoring_component"
	"github.com/erniealice/fayna-golang/domain/operation/scoring_component_criteria"
	"github.com/erniealice/fayna-golang/domain/operation/scoring_scheme"
	"github.com/erniealice/fayna-golang/domain/operation/task_outcome"
	"github.com/erniealice/fayna-golang/domain/operation/template_task_criteria"
)

func JobUnit(uc *UseCases, infra *Infra) compose.Unit {
	u := job.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*job.Routes)
		l := u.Labels.(*job.Labels)

		deps := &operation.JobModuleDeps{
			Routes:                *r,
			Labels:                *l,
			CommonLabels:          mc.Common,
			TableLabels:           mc.Table,
			UploadFile:            infra.UploadFile,
			ListAttachments:       infra.ListAttachments,
			CreateAttachment:      infra.CreateAttachment,
			DeleteAttachment:      infra.DeleteAttachment,
			NewID:                 infra.NewAttachmentID,
			SubscriptionDetailURL: infra.SubscriptionDetailURL,
			ClientSearchURL:       r.ClientSearchURL,
			LocationSearchURL:     r.LocationSearchURL,
			// BusinessType branches the List view to the template-grain
			// delivery summary for "education" (20260710 staff-class-list).
			BusinessType: mc.BusinessType,
		}
		if jaRoutes, ok := compose.RoutesOf[*job_activity.Routes](mc, "operation.job_activity"); ok {
			deps.JobActivityRoutes = *jaRoutes
		}
		if jaLabels, ok := compose.LabelsOf[*job_activity.Labels](mc, "operation.job_activity"); ok {
			deps.JobActivityLabels = *jaLabels
		}
		// outcome_matrix.matrix route — the template-grain delivery summary
		// row link ("/outcome-matrix/{id}", id=job_template_id).
		if omRoutes, ok := compose.RoutesOf[*outcome_matrix.Routes](mc, "operation.outcome_matrix"); ok {
			deps.MatrixDetailURL = omRoutes.MatrixURL
		}
		if infra.RefChecker != nil {
			deps.GetInUseIDs = infra.RefChecker.GetJobInUseIDs
		}
		wireJobDeps(deps, uc)
		wireJobDashboard(deps, uc)
		operation.NewJobModule(deps).RegisterRoutes(mc.Routes)

		clientSearchFn := newJobClientSearchHandler(uc.Entity.Client.SearchClientsByName, uc.Entity.Client.ListClients)
		locationSearchFn := newJobLocationSearchHandler(infra.DB)
		compose.HandleFunc(mc.Routes, "GET", r.ClientSearchURL, clientSearchFn)
		compose.HandleFunc(mc.Routes, "GET", r.LocationSearchURL, locationSearchFn)
		return nil
	}
	return u
}

func JobTemplateUnit(uc *UseCases, infra *Infra) compose.Unit {
	u := job_template.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*job_template.Routes)
		l := u.Labels.(*job_template.Labels)

		deps := &operation.JobTemplateModuleDeps{
			Routes:           *r,
			Labels:           *l,
			CommonLabels:     mc.Common,
			TableLabels:      mc.Table,
			UploadFile:       infra.UploadFile,
			ListAttachments:  infra.ListAttachments,
			CreateAttachment: infra.CreateAttachment,
			DeleteAttachment: infra.DeleteAttachment,
			NewID:            infra.NewAttachmentID,
		}
		if jtpRoutes, ok := compose.RoutesOf[*job_template_phase.Routes](mc, "operation.job_template_phase"); ok {
			deps.PhaseRoutes = *jtpRoutes
		}
		if jttRoutes, ok := compose.RoutesOf[*job_template_task.Routes](mc, "operation.job_template_task"); ok {
			deps.TaskRoutes = *jttRoutes
		}
		if infra.RefChecker != nil {
			deps.GetInUseIDs = infra.RefChecker.GetJobTemplateInUseIDs
		}
		wireJobTemplateDeps(deps, uc)
		operation.NewJobTemplateModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

func JobTemplatePhaseUnit(uc *UseCases, _ *Infra) compose.Unit {
	u := job_template_phase.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*job_template_phase.Routes)
		l := u.Labels.(*job_template_phase.Labels)

		deps := &operation.JobTemplatePhaseModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
		}
		wireJobTemplatePhaseDeps(deps, uc)
		operation.NewJobTemplatePhaseModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

func JobTemplateTaskUnit(uc *UseCases, _ *Infra) compose.Unit {
	u := job_template_task.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*job_template_task.Routes)
		l := u.Labels.(*job_template_task.Labels)

		deps := &operation.JobTemplateTaskModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
		}
		wireJobTemplateTaskDeps(deps, uc)
		operation.NewJobTemplateTaskModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

func JobActivityUnit(uc *UseCases, infra *Infra) compose.Unit {
	u := job_activity.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*job_activity.Routes)
		l := u.Labels.(*job_activity.Labels)

		deps := &operation.JobActivityModuleDeps{
			Routes:           *r,
			Labels:           *l,
			CommonLabels:     mc.Common,
			TableLabels:      mc.Table,
			UploadFile:       infra.UploadFile,
			ListAttachments:  infra.ListAttachments,
			CreateAttachment: infra.CreateAttachment,
			DeleteAttachment: infra.DeleteAttachment,
			NewID:            infra.NewAttachmentID,
		}
		if alRoutes, ok := compose.RoutesOf[*activity_labor.Routes](mc, "operation.activity_labor"); ok {
			deps.ActivityLaborRoutes = *alRoutes
		}
		if amRoutes, ok := compose.RoutesOf[*activity_material.Routes](mc, "operation.activity_material"); ok {
			deps.ActivityMaterialRoutes = *amRoutes
		}
		if aeRoutes, ok := compose.RoutesOf[*activity_expense.Routes](mc, "operation.activity_expense"); ok {
			deps.ActivityExpenseRoutes = *aeRoutes
		}
		if infra.RefChecker != nil {
			deps.GetInUseIDs = infra.RefChecker.GetJobActivityInUseIDs
		}
		wireJobActivityDeps(deps, uc)
		operation.NewJobActivityModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

func JobPhaseUnit(uc *UseCases, infra *Infra) compose.Unit {
	u := job_phase.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*job_phase.Routes)
		l := u.Labels.(*job_phase.Labels)

		deps := &operation.JobPhaseModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
		}
		deps.AttachmentOps.UploadFile = infra.UploadFile
		deps.AttachmentOps.ListAttachments = infra.ListAttachments
		deps.AttachmentOps.CreateAttachment = infra.CreateAttachment
		deps.AttachmentOps.DeleteAttachment = infra.DeleteAttachment
		deps.AttachmentOps.NewAttachmentID = infra.NewAttachmentID
		if infra.RefChecker != nil {
			deps.GetInUseIDs = infra.RefChecker.GetJobPhaseInUseIDs
		}
		wireJobPhaseDeps(deps, uc)
		operation.NewJobPhaseModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

func JobTaskUnit(uc *UseCases, infra *Infra) compose.Unit {
	u := job_task.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*job_task.Routes)
		l := u.Labels.(*job_task.Labels)

		deps := &operation.JobTaskModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
		}
		deps.AttachmentOps.UploadFile = infra.UploadFile
		deps.AttachmentOps.ListAttachments = infra.ListAttachments
		deps.AttachmentOps.CreateAttachment = infra.CreateAttachment
		deps.AttachmentOps.DeleteAttachment = infra.DeleteAttachment
		deps.AttachmentOps.NewAttachmentID = infra.NewAttachmentID
		if infra.RefChecker != nil {
			deps.GetInUseIDs = infra.RefChecker.GetJobTaskInUseIDs
		}
		wireJobTaskDeps(deps, uc)

		staffSearchFn := newActivityLaborStaffSearchHandler(uc.Entity.Staff.ListStaffs)
		if staffSearchFn == nil {
			r.StaffSearchURL = ""
		}
		r.ResourceSearchURL = ""
		r.TemplateTaskSearchURL = ""
		operation.NewJobTaskModule(deps).RegisterRoutes(mc.Routes)
		compose.HandleFunc(mc.Routes, "GET", r.StaffSearchURL, staffSearchFn)
		return nil
	}
	return u
}

func ActivityLaborUnit(uc *UseCases, _ *Infra) compose.Unit {
	u := activity_labor.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*activity_labor.Routes)
		l := u.Labels.(*activity_labor.Labels)

		deps := &operation.ActivityLaborModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
		}
		wireActivityLaborDeps(deps, uc)

		staffSearchFn := newActivityLaborStaffSearchHandler(uc.Entity.Staff.ListStaffs)
		if staffSearchFn == nil {
			r.StaffSearchURL = ""
			log.Printf("compose: staff search handler not available — ActivityLabor drawer will use flat filter mode")
		}
		operation.NewActivityLaborModule(deps).RegisterRoutes(mc.Routes)
		compose.HandleFunc(mc.Routes, "GET", r.StaffSearchURL, staffSearchFn)
		return nil
	}
	return u
}

func ActivityMaterialUnit(uc *UseCases, infra *Infra) compose.Unit {
	u := activity_material.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*activity_material.Routes)
		l := u.Labels.(*activity_material.Labels)

		deps := &operation.ActivityMaterialModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
		}
		wireActivityMaterialDeps(deps, uc)

		var productSearchFn, locationSearchFn http.HandlerFunc
		if infra.DB != nil {
			productSearchFn = newActivityMaterialProductSearchHandler(infra.DB)
			locationSearchFn = newJobLocationSearchHandler(infra.DB)
		}
		if productSearchFn == nil {
			r.ProductSearchURL = ""
		}
		if locationSearchFn == nil {
			r.LocationSearchURL = ""
		}
		operation.NewActivityMaterialModule(deps).RegisterRoutes(mc.Routes)
		compose.HandleFunc(mc.Routes, "GET", r.ProductSearchURL, productSearchFn)
		compose.HandleFunc(mc.Routes, "GET", r.LocationSearchURL, locationSearchFn)
		return nil
	}
	return u
}

func ActivityExpenseUnit(uc *UseCases, infra *Infra) compose.Unit {
	u := activity_expense.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*activity_expense.Routes)
		l := u.Labels.(*activity_expense.Labels)

		deps := &operation.ActivityExpenseModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
		}
		wireActivityExpenseDeps(deps, uc)

		var expenseCategorySearchFn http.HandlerFunc
		if infra.DB != nil {
			expenseCategorySearchFn = newActivityExpenseExpenseCategorySearchHandler(infra.DB)
		}
		if expenseCategorySearchFn == nil {
			r.ExpenseCategorySearchURL = ""
		}
		operation.NewActivityExpenseModule(deps).RegisterRoutes(mc.Routes)
		compose.HandleFunc(mc.Routes, "GET", r.ExpenseCategorySearchURL, expenseCategorySearchFn)
		return nil
	}
	return u
}

func OutcomeCriteriaUnit(uc *UseCases, infra *Infra) compose.Unit {
	u := outcome_criteria.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*outcome_criteria.Routes)
		l := u.Labels.(*outcome_criteria.Labels)

		deps := &operation.OutcomeCriteriaModuleDeps{
			Routes:           *r,
			Labels:           *l,
			CommonLabels:     mc.Common,
			TableLabels:      mc.Table,
			UploadFile:       infra.UploadFile,
			ListAttachments:  infra.ListAttachments,
			CreateAttachment: infra.CreateAttachment,
			DeleteAttachment: infra.DeleteAttachment,
			NewID:            infra.NewAttachmentID,
		}
		wireOutcomeCriteriaDeps(deps, uc)
		operation.NewOutcomeCriteriaModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

// ---------------------------------------------------------------------------
// Education grading units (20260616 v1) — fayna/operation domain.
//
// All grading CRUD closures are OPTIONAL / nil-able (NOT in RequireFor): a
// missing closure degrades the view to empty-state rather than refusing boot.
// buildFaynaUseCases (engineblock.go) is the espyna→block adapter that populates
// uc.Operation.{Entity}. Audit-history ops are nil-able (history tab empty).
// ---------------------------------------------------------------------------

func ScoringSchemeUnit(uc *UseCases, _ *Infra) compose.Unit {
	u := scoring_scheme.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*scoring_scheme.Routes)
		l := u.Labels.(*scoring_scheme.Labels)

		deps := &operation.ScoringSchemeModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
		}
		wireScoringSchemeDeps(deps, uc)
		operation.NewScoringSchemeModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

func ScoringComponentUnit(uc *UseCases, _ *Infra) compose.Unit {
	u := scoring_component.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*scoring_component.Routes)
		l := u.Labels.(*scoring_component.Labels)

		deps := &operation.ScoringComponentModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
		}
		wireScoringComponentDeps(deps, uc)
		operation.NewScoringComponentModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

func ScoringComponentCriteriaUnit(uc *UseCases, _ *Infra) compose.Unit {
	u := scoring_component_criteria.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*scoring_component_criteria.Routes)
		l := u.Labels.(*scoring_component_criteria.Labels)

		deps := &operation.ScoringComponentCriteriaModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
		}
		wireScoringComponentCriteriaDeps(deps, uc)
		operation.NewScoringComponentCriteriaModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

func TemplateTaskCriteriaUnit(uc *UseCases, _ *Infra) compose.Unit {
	u := template_task_criteria.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*template_task_criteria.Routes)
		l := u.Labels.(*template_task_criteria.Labels)

		deps := &operation.TemplateTaskCriteriaModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
		}
		wireTemplateTaskCriteriaDeps(deps, uc)
		operation.NewTemplateTaskCriteriaModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

func ScoreScaleUnit(uc *UseCases, _ *Infra) compose.Unit {
	u := score_scale.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*score_scale.Routes)
		l := u.Labels.(*score_scale.Labels)

		deps := &operation.ScoreScaleModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
		}
		wireScoreScaleDeps(deps, uc)
		operation.NewScoreScaleModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

func ScoreScaleBandUnit(uc *UseCases, _ *Infra) compose.Unit {
	u := score_scale_band.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*score_scale_band.Routes)
		l := u.Labels.(*score_scale_band.Labels)

		deps := &operation.ScoreScaleBandModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
		}
		wireScoreScaleBandDeps(deps, uc)
		operation.NewScoreScaleBandModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

func JobOutcomeLineUnit(uc *UseCases, _ *Infra) compose.Unit {
	u := job_outcome_line.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*job_outcome_line.Routes)
		l := u.Labels.(*job_outcome_line.Labels)

		deps := &operation.JobOutcomeLineModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
		}
		wireJobOutcomeLineDeps(deps, uc)
		operation.NewJobOutcomeLineModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

func ReportingCheckpointUnit(uc *UseCases, _ *Infra) compose.Unit {
	u := reporting_checkpoint.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*reporting_checkpoint.Routes)
		l := u.Labels.(*reporting_checkpoint.Labels)

		deps := &operation.ReportingCheckpointModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
		}
		wireReportingCheckpointDeps(deps, uc)
		operation.NewReportingCheckpointModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

func TaskOutcomeUnit(uc *UseCases, infra *Infra) compose.Unit {
	u := task_outcome.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*task_outcome.Routes)
		l := u.Labels.(*task_outcome.Labels)

		deps := &operation.TaskOutcomeModuleDeps{
			Routes:           *r,
			Labels:           *l,
			CommonLabels:     mc.Common,
			TableLabels:      mc.Table,
			UploadFile:       infra.UploadFile,
			ListAttachments:  infra.ListAttachments,
			CreateAttachment: infra.CreateAttachment,
			DeleteAttachment: infra.DeleteAttachment,
			NewID:            infra.NewAttachmentID,
		}
		wireTaskOutcomeDeps(deps, uc)
		operation.NewTaskOutcomeModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

// OutcomeMatrixUnit registers the generic principal-scoped grading grid
// (rows = client × job_template, columns = phase→task→criterion, cells =
// task_outcome) — the cross-vertical replacement for the education-specific
// grade_sheet. Read + batch-save only (no CRUD sub-entity). All auth/scope/IDOR
// gates live in the espyna read use case + the record action; the closures here
// are nil-safe (empty-state / fail-closed) on non-postgres builds. options is
// the app's row-presentation config (EngineBlock's view option block); the
// zero value renders the flat roster unchanged.
func OutcomeMatrixUnit(uc *UseCases, _ *Infra, options outcome_matrix.Options) compose.Unit {
	u := outcome_matrix.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*outcome_matrix.Routes)
		l := u.Labels.(*outcome_matrix.Labels)

		deps := &operation.OutcomeMatrixModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			Options:      options,
		}
		wireOutcomeMatrixDeps(deps, uc)
		operation.NewOutcomeMatrixModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

func OutcomeSummaryUnit(uc *UseCases, _ *Infra) compose.Unit {
	u := outcome_summary.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*outcome_summary.Routes)
		l := u.Labels.(*outcome_summary.Labels)

		deps := &operation.OutcomeSummaryModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
		}
		wireOutcomeSummaryDeps(deps, uc)
		operation.NewOutcomeSummaryModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

func FulfillmentUnit(uc *UseCases, infra *Infra) compose.Unit {
	u := fulfillmentpkg.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*fulfillmentpkg.Routes)
		l := u.Labels.(*fulfillmentpkg.Labels)

		deps := &fulfillmentdomain.FulfillmentModuleDeps{
			Routes:           *r,
			Labels:           *l,
			CommonLabels:     mc.Common,
			TableLabels:      mc.Table,
			UploadFile:       infra.UploadFile,
			ListAttachments:  infra.ListAttachments,
			CreateAttachment: infra.CreateAttachment,
			DeleteAttachment: infra.DeleteAttachment,
			NewID:            infra.NewAttachmentID,
		}
		wireFulfillmentDeps(deps, uc)
		wireFulfillmentDashboard(deps, uc)
		fulfillmentdomain.NewFulfillmentModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

// ---------------------------------------------------------------------------
// Performance-Evaluation units (20260604) — fayna/operation domain.
//
// All eval use-case closures are OPTIONAL / nil-able (NOT in RequireFor): a
// missing closure degrades the view to empty-state rather than refusing boot.
// service-admin's buildFaynaUseCases (adapters_fayna.go) is the espyna→block
// adapter that populates uc.Operation.Evaluation* / uc.Service.Performance.
// All IDOR / CR-5 servicing gates live INSIDE the closures (espyna QUERY
// PREDICATE) — the view supplies no client_id.
// ---------------------------------------------------------------------------

// EvaluationUnit registers the evaluation entity (staff Reviews list/detail +
// the polymorphic drawer-form + the client-portal "Rate My Team").
func EvaluationUnit(uc *UseCases, infra *Infra) compose.Unit {
	u := evaluation.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*evaluation.Routes)
		l := u.Labels.(*evaluation.Labels)

		deps := &operation.EvaluationModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
			NewID:        infra.NewAttachmentID,
		}
		wireEvaluationDeps(deps, uc)
		operation.NewEvaluationModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

// EvaluationTemplateUnit registers the staff-only evaluation_template authoring
// surface. It wires the rubric-item drawer routes from the item unit so the
// detail Items tab can mount Add Question / edit / remove endpoints.
func EvaluationTemplateUnit(uc *UseCases, infra *Infra) compose.Unit {
	u := evaluation_template.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*evaluation_template.Routes)
		l := u.Labels.(*evaluation_template.Labels)

		deps := &operation.EvaluationTemplateModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
			NewID:        infra.NewAttachmentID,
		}
		if itemRoutes, ok := compose.RoutesOf[*evaluation_template_item.Routes](mc, "operation.evaluation_template_item"); ok {
			deps.ItemRoutes = *itemRoutes
		}
		wireEvaluationTemplateDeps(deps, uc)
		operation.NewEvaluationTemplateModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

// EvaluationTemplateItemUnit registers the rubric-item drawer (Add/Edit/Remove).
// No standalone page — surfaces via the evaluation_template detail Items tab.
func EvaluationTemplateItemUnit(uc *UseCases, infra *Infra) compose.Unit {
	u := evaluation_template_item.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*evaluation_template_item.Routes)
		l := u.Labels.(*evaluation_template_item.Labels)

		deps := &operation.EvaluationTemplateItemModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
			NewID:        infra.NewAttachmentID,
		}
		wireEvaluationTemplateItemDeps(deps, uc)
		operation.NewEvaluationTemplateItemModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

// EvaluationCycleUnit registers the evaluation_cycle module (list/detail +
// Open/Close lifecycle + Members tab + the X-of-Y progress banner).
func EvaluationCycleUnit(uc *UseCases, infra *Infra) compose.Unit {
	u := evaluation_cycle.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*evaluation_cycle.Routes)
		l := u.Labels.(*evaluation_cycle.Labels)

		deps := &operation.EvaluationCycleModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
			NewID:        infra.NewAttachmentID,
		}
		wireEvaluationCycleDeps(deps, uc)
		operation.NewEvaluationCycleModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

// EvaluationCycleMemberUnit (STR-1) is a data/templates-only Unit: it contributes
// the members-tab.html templates FS + label JSON consumed by the cycle detail
// view. No Routes, no Nav, no Mount.
func EvaluationCycleMemberUnit(_ *UseCases, _ *Infra) compose.Unit {
	return evaluation_cycle_member.Describe()
}

// PerformanceUnit registers the performance admin panel (Surface 6). The single
// page view is servicing-gated (CR-5) inside the block-supplied GetPanelData
// closure; the X-of-Y banner is supplied within PanelData by the adapter.
func PerformanceUnit(uc *UseCases, _ *Infra) compose.Unit {
	u := performance.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*performance.Routes)
		l := u.Labels.(*performance.Labels)

		deps := &operation.PerformanceModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
			GetPanelData: uc.Service.Performance.GetPanelData,
		}
		operation.NewPerformanceModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

// AllUnits returns the complete curated unit list for the fayna/operation +
// fulfillment domains, in the same registration order as Block(). Optional
// EngineOptions (the consuming app's view option block, forwarded verbatim
// by EngineBlock) configure individual units — today the outcome-matrix row
// presentation.
func AllUnits(uc *UseCases, infra *Infra, opts ...EngineOption) []compose.Unit {
	cfg := engineConfig{}
	for _, o := range opts {
		o(&cfg)
	}
	return []compose.Unit{
		JobUnit(uc, infra),
		JobTemplateUnit(uc, infra),
		JobTemplatePhaseUnit(uc, infra),
		JobTemplateTaskUnit(uc, infra),
		JobActivityUnit(uc, infra),
		JobPhaseUnit(uc, infra),
		JobTaskUnit(uc, infra),
		ActivityLaborUnit(uc, infra),
		ActivityMaterialUnit(uc, infra),
		ActivityExpenseUnit(uc, infra),
		OutcomeCriteriaUnit(uc, infra),
		// Education grading (20260616 v1) — single-repo CRUD entities.
		ScoringSchemeUnit(uc, infra),
		ScoringComponentUnit(uc, infra),
		ScoringComponentCriteriaUnit(uc, infra),
		TemplateTaskCriteriaUnit(uc, infra),
		ScoreScaleUnit(uc, infra),
		ScoreScaleBandUnit(uc, infra),
		JobOutcomeLineUnit(uc, infra),
		ReportingCheckpointUnit(uc, infra),
		TaskOutcomeUnit(uc, infra),
		OutcomeMatrixUnit(uc, infra, cfg.outcomeMatrixOptions),
		OutcomeSummaryUnit(uc, infra),
		FulfillmentUnit(uc, infra),
		// Performance-Evaluation (20260604). evaluation_template_item must be
		// registered (it has no Nav) so the template unit's RoutesOf lookup
		// resolves its Add/Edit/Remove drawer URLs.
		EvaluationUnit(uc, infra),
		EvaluationTemplateUnit(uc, infra),
		EvaluationTemplateItemUnit(uc, infra),
		EvaluationCycleUnit(uc, infra),
		EvaluationCycleMemberUnit(uc, infra),
		PerformanceUnit(uc, infra),
	}
}
