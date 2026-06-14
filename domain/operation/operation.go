package operation

// operation.go — facade: re-exports every entity-local type with the entity
// prefix so that consumers (block/, service-admin) keep writing
// operation.JobLabels, operation.DefaultJobLabels(), etc.
//
// ROUTE types are type aliases already defined (or will be) in routes.go.
// LABEL types are re-exported here.
//
// Rule D: this file may NOT import the domain facade from inside an entity
// package (that would create a cycle). Entity packages import nothing from
// domain/operation/; only the domain/operation/ packages import entity packages.

import (
	activityexpensepkg "github.com/erniealice/fayna-golang/domain/operation/activity_expense"
	activitylaborpkg "github.com/erniealice/fayna-golang/domain/operation/activity_labor"
	activitymaterialpkg "github.com/erniealice/fayna-golang/domain/operation/activity_material"
	jobpkg "github.com/erniealice/fayna-golang/domain/operation/job"
	jobactivitypkg "github.com/erniealice/fayna-golang/domain/operation/job_activity"
	jobphasepkg "github.com/erniealice/fayna-golang/domain/operation/job_phase"
	jobtaskpkg "github.com/erniealice/fayna-golang/domain/operation/job_task"
	jobtemplatepkg "github.com/erniealice/fayna-golang/domain/operation/job_template"
	jobtemplatephasepkg "github.com/erniealice/fayna-golang/domain/operation/job_template_phase"
	jobtemplateTaskpkg "github.com/erniealice/fayna-golang/domain/operation/job_template_task"
	outcomecritetriapkg "github.com/erniealice/fayna-golang/domain/operation/outcome_criteria"
	outcomesummarypkg "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	taskoutcomepkg "github.com/erniealice/fayna-golang/domain/operation/task_outcome"
	workrequestpkg "github.com/erniealice/fayna-golang/domain/operation/work_request"
	workrequesttypepkg "github.com/erniealice/fayna-golang/domain/operation/work_request_type"
)

// ---------------------------------------------------------------------------
// Job label types (operation/job)
// ---------------------------------------------------------------------------

type JobLabels = jobpkg.Labels

func DefaultJobLabels() JobLabels { return jobpkg.DefaultLabels() }

// ---------------------------------------------------------------------------
// JobActivity label types (operation/job_activity)
// ---------------------------------------------------------------------------

type JobActivityLabels = jobactivitypkg.Labels

func DefaultJobActivityLabels() JobActivityLabels { return jobactivitypkg.DefaultLabels() }

// ---------------------------------------------------------------------------
// JobPhase label types (operation/job_phase)
// ---------------------------------------------------------------------------

type JobPhaseLabels = jobphasepkg.Labels

func DefaultJobPhaseLabels() JobPhaseLabels { return jobphasepkg.DefaultLabels() }

// ---------------------------------------------------------------------------
// JobTask label types (operation/job_task)
// ---------------------------------------------------------------------------

type JobTaskLabels = jobtaskpkg.Labels

func DefaultJobTaskLabels() JobTaskLabels { return jobtaskpkg.DefaultLabels() }

// ---------------------------------------------------------------------------
// JobTemplate label types (operation/job_template)
// ---------------------------------------------------------------------------

type JobTemplateLabels = jobtemplatepkg.Labels

func DefaultJobTemplateLabels() JobTemplateLabels { return jobtemplatepkg.DefaultLabels() }

// ---------------------------------------------------------------------------
// JobTemplatePhase label types (operation/job_template_phase)
// ---------------------------------------------------------------------------

type JobTemplatePhaseLabels = jobtemplatephasepkg.Labels

func DefaultJobTemplatePhaseLabels() JobTemplatePhaseLabels {
	return jobtemplatephasepkg.DefaultLabels()
}

// ---------------------------------------------------------------------------
// JobTemplateTask label types (operation/job_template_task)
// ---------------------------------------------------------------------------

type JobTemplateTaskLabels = jobtemplateTaskpkg.Labels

func DefaultJobTemplateTaskLabels() JobTemplateTaskLabels {
	return jobtemplateTaskpkg.DefaultLabels()
}

// ---------------------------------------------------------------------------
// ActivityLabor label types (operation/activity_labor)
// ---------------------------------------------------------------------------

type ActivityLaborLabels = activitylaborpkg.Labels

func DefaultActivityLaborLabels() ActivityLaborLabels { return activitylaborpkg.DefaultLabels() }

// ---------------------------------------------------------------------------
// ActivityMaterial label types (operation/activity_material)
// ---------------------------------------------------------------------------

type ActivityMaterialLabels = activitymaterialpkg.Labels

func DefaultActivityMaterialLabels() ActivityMaterialLabels {
	return activitymaterialpkg.DefaultLabels()
}

// ---------------------------------------------------------------------------
// ActivityExpense label types (operation/activity_expense)
// ---------------------------------------------------------------------------

type ActivityExpenseLabels = activityexpensepkg.Labels

func DefaultActivityExpenseLabels() ActivityExpenseLabels {
	return activityexpensepkg.DefaultLabels()
}

// ---------------------------------------------------------------------------
// OutcomeCriteria label types (operation/outcome_criteria)
// ---------------------------------------------------------------------------

type OutcomeCriteriaLabels = outcomecritetriapkg.Labels

func DefaultOutcomeCriteriaLabels() OutcomeCriteriaLabels {
	return outcomecritetriapkg.DefaultLabels()
}

// ---------------------------------------------------------------------------
// TaskOutcome label types (operation/task_outcome)
// ---------------------------------------------------------------------------

type TaskOutcomeLabels = taskoutcomepkg.Labels

func DefaultTaskOutcomeLabels() TaskOutcomeLabels { return taskoutcomepkg.DefaultLabels() }

// ---------------------------------------------------------------------------
// OutcomeSummary label types (operation/outcome_summary)
// ---------------------------------------------------------------------------

type OutcomeSummaryLabels = outcomesummarypkg.Labels

func DefaultOutcomeSummaryLabels() OutcomeSummaryLabels { return outcomesummarypkg.DefaultLabels() }

// ---------------------------------------------------------------------------
// WorkRequest label types (operation/work_request)
// ---------------------------------------------------------------------------

type WorkRequestLabels = workrequestpkg.Labels

func DefaultWorkRequestLabels() WorkRequestLabels { return workrequestpkg.DefaultLabels() }

// ---------------------------------------------------------------------------
// WorkRequestType label types (operation/work_request_type)
// ---------------------------------------------------------------------------

type WorkRequestTypeLabels = workrequesttypepkg.Labels

func DefaultWorkRequestTypeLabels() WorkRequestTypeLabels {
	return workrequesttypepkg.DefaultLabels()
}
