package block

import (
	"log"
	"net/http"

	fulfillmentdomain "github.com/erniealice/fayna-golang/domain/fulfillment"
	fulfillmentpkg "github.com/erniealice/fayna-golang/domain/fulfillment/fulfillment"
	operation "github.com/erniealice/fayna-golang/domain/operation"
	"github.com/erniealice/fayna-golang/domain/operation/activity_expense"
	"github.com/erniealice/fayna-golang/domain/operation/activity_labor"
	"github.com/erniealice/fayna-golang/domain/operation/activity_material"
	"github.com/erniealice/fayna-golang/domain/operation/job"
	"github.com/erniealice/fayna-golang/domain/operation/job_activity"
	"github.com/erniealice/fayna-golang/domain/operation/job_phase"
	"github.com/erniealice/fayna-golang/domain/operation/job_task"
	"github.com/erniealice/fayna-golang/domain/operation/job_template"
	"github.com/erniealice/fayna-golang/domain/operation/job_template_phase"
	"github.com/erniealice/fayna-golang/domain/operation/job_template_task"
	"github.com/erniealice/fayna-golang/domain/operation/outcome_criteria"
	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/fayna-golang/domain/operation/task_outcome"
	"github.com/erniealice/pyeza-golang/compose"
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
		}
		if jaRoutes, ok := compose.RoutesOf[*job_activity.Routes](mc, "operation.job_activity"); ok {
			deps.JobActivityRoutes = *jaRoutes
		}
		if jaLabels, ok := compose.LabelsOf[*job_activity.Labels](mc, "operation.job_activity"); ok {
			deps.JobActivityLabels = *jaLabels
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

// AllUnits returns the complete curated unit list for the fayna/operation +
// fulfillment domains, in the same registration order as Block().
func AllUnits(uc *UseCases, infra *Infra) []compose.Unit {
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
		TaskOutcomeUnit(uc, infra),
		OutcomeSummaryUnit(uc, infra),
		FulfillmentUnit(uc, infra),
	}
}
