package document

import (
	"testing"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	categorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplatetaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	phaseoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

// periodSummaryFixture: two per-subject jobs of category "rows_by_job" and one
// job of category "rows_by_task" with two activities, over periods 1 and 2.
func periodSummaryFixture() *exportpb.ClientReportCardProjection {
	client := "client-1"
	return &exportpb.ClientReportCardProjection{
		Client: &exportpb.ClientReportCardClient{ClientId: client},
		JobCategories: []*categorypb.JobCategory{
			{Id: "cat-job", Code: ptr("rows_by_job")}, {Id: "cat-task", Code: ptr("rows_by_task")},
		},
		JobTemplates: []*jobtemplatepb.JobTemplate{
			{Id: "t-a", JobCategoryId: ptr("cat-job"), Active: true},
			{Id: "t-h", JobCategoryId: ptr("cat-task"), Active: true},
		},
		JobTemplatePhases: []*jobtemplatephasepb.JobTemplatePhase{
			{Id: "tp-a1", JobTemplateId: "t-a", PhaseOrder: 1, Active: true}, {Id: "tp-a2", JobTemplateId: "t-a", PhaseOrder: 2, Active: true},
			{Id: "tp-h1", JobTemplateId: "t-h", PhaseOrder: 1, Active: true}, {Id: "tp-h2", JobTemplateId: "t-h", PhaseOrder: 2, Active: true},
		},
		JobTemplateTasks: []*jobtemplatetaskpb.JobTemplateTask{
			{Id: "tt-h1-d", JobTemplatePhaseId: "tp-h1", Code: ptr("discipline"), Name: "Discipline", StepOrder: 1, Active: true},
			{Id: "tt-h1-e", JobTemplatePhaseId: "tp-h1", Code: ptr("empathy"), Name: "Empathy", StepOrder: 2, Active: true},
			{Id: "tt-h2-d", JobTemplatePhaseId: "tp-h2", Code: ptr("discipline"), Name: "Discipline", StepOrder: 1, Active: true},
		},
		Jobs: []*jobpb.Job{
			{Id: "j-sci", Name: "Sciences (Deportment)", JobTemplateId: ptr("t-a"), ClientId: &client, Active: true},
			{Id: "j-art", Name: "Arts (Deportment)", JobTemplateId: ptr("t-a"), ClientId: &client, Active: true},
			{Id: "j-home", Name: "Homeroom", JobTemplateId: ptr("t-h"), ClientId: &client, Active: true},
		},
		JobPhases: []*jobphasepb.JobPhase{
			{Id: "p-art-1", JobId: "j-art", TemplatePhaseId: ptr("tp-a1"), Active: true},
			{Id: "p-art-2", JobId: "j-art", TemplatePhaseId: ptr("tp-a2"), Active: true},
			{Id: "p-sci-1", JobId: "j-sci", TemplatePhaseId: ptr("tp-a1"), Active: true},
			{Id: "p-home-1", JobId: "j-home", TemplatePhaseId: ptr("tp-h1"), Active: true},
			{Id: "p-home-2", JobId: "j-home", TemplatePhaseId: ptr("tp-h2"), Active: true},
		},
		JobTasks: []*jobtaskpb.JobTask{
			{Id: "task-h1-d", JobPhaseId: "p-home-1", TemplateTaskId: ptr("tt-h1-d"), Active: true},
			{Id: "task-h1-e", JobPhaseId: "p-home-1", TemplateTaskId: ptr("tt-h1-e"), Active: true},
			{Id: "task-h2-d", JobPhaseId: "p-home-2", TemplateTaskId: ptr("tt-h2-d"), Active: true},
		},
		TaskOutcomes: []*exportpb.ClientReportCardTaskOutcome{
			{JobTaskId: "task-h1-d", TemplateTaskCriteriaId: "l1", NumericValue: ptr(float64(92))},
			{JobTaskId: "task-h1-e", TemplateTaskCriteriaId: "l2", NumericValue: ptr(float64(85))},
		},
		PhaseOutcomeSummaries: []*phaseoutcomepb.PhaseOutcomeSummary{
			{Id: "s1", JobPhaseId: "p-art-1", SummaryScore: ptr(float64(88)), ScaledLabel: ptr("VS"), Active: true},
			{Id: "s2", JobPhaseId: "p-art-2", SummaryScore: ptr(float64(91)), ScaledLabel: ptr("O"), Active: true},
			{Id: "s3", JobPhaseId: "p-home-1", SummaryScore: ptr(float64(89)), ScaledLabel: ptr("VS"), Active: true},
		},
	}
}

func TestBuildPeriodSummaries(t *testing.T) {
	out := buildPeriodSummaries(periodSummaryFixture(), "rows_by_job", "rows_by_task", false)
	jobs := out["summary_jobs"].([]any)
	if len(jobs) != 2 {
		t.Fatalf("summary_jobs = %#v, want 2 rows", jobs)
	}
	arts := jobs[0].(map[string]any) // sorted by name: Arts before Sciences
	if arts["summary_job_name"] != "Arts" || arts["summary_period_1"] != "88" || arts["summary_period_2"] != "91" || arts["summary_period_3"] != "" {
		t.Fatalf("arts row = %#v", arts)
	}
	if sci := jobs[1].(map[string]any); sci["summary_job_name"] != "Sciences" || sci["summary_period_1"] != "" {
		t.Fatalf("sciences row (no summary) = %#v", sci)
	}
	tasks := out["summary_tasks"].([]any)
	if len(tasks) != 2 {
		t.Fatalf("summary_tasks = %#v, want Discipline + Empathy", tasks)
	}
	d := tasks[0].(map[string]any)
	if d["summary_task_name"] != "Discipline" || d["summary_task_period_1"] != "92" || d["summary_task_period_2"] != "" {
		t.Fatalf("discipline row = %#v", d)
	}
	if out["summary_average_period_1"] != "89" || out["summary_transmuted_period_1"] != "VS" || out["summary_average_period_2"] != "" {
		t.Fatalf("homeroom average/transmuted = %v / %v / %v", out["summary_average_period_1"], out["summary_transmuted_period_1"], out["summary_average_period_2"])
	}
	// Nameless job rows (the live projection) fall back to the template name.
	nameless := periodSummaryFixture()
	nameless.JobTemplates[0].Name = "Arts (Deportment)"
	for _, j := range nameless.Jobs {
		j.Name = ""
	}
	if got := buildPeriodSummaries(nameless, "rows_by_job", "", false)["summary_jobs"].([]any)[0].(map[string]any)["summary_job_name"]; got != "Arts" {
		t.Fatalf("nameless job row name = %v, want template-derived Arts", got)
	}
	// Template category wins over a stale job-row category (same-origin jobs).
	stale := periodSummaryFixture()
	stale.Jobs[0].JobCategoryId = ptr("cat-task")
	if got := buildPeriodSummaries(stale, "rows_by_job", "", false)["summary_jobs"].([]any); len(got) != 2 {
		t.Fatalf("template category must decide membership, got %d rows", len(got))
	}
	// Two jobs of the activity category are ambiguous: no activity rows, no fabrication.
	card := periodSummaryFixture()
	card.Jobs = append(card.Jobs, &jobpb.Job{Id: "j-home-2", Name: "Homeroom B", JobTemplateId: ptr("t-h"), ClientId: ptr("client-1"), Active: true})
	if got := buildPeriodSummaries(card, "", "rows_by_task", false)["summary_tasks"].([]any); len(got) != 0 {
		t.Fatalf("ambiguous activity job must yield no rows, got %#v", got)
	}
}

func TestBuildProjectedOutcomeCellIndexCategoryFamily(t *testing.T) {
	card := clientPhaseProjectionFixture()
	_, _, categoryCells, categoryTotals := buildProjectedOutcomeCellIndex(card, false)
	cell := categoryCells["academic"].(map[string]any)["technique"].(map[string]any)["m07"].(map[string]any)
	if cell["numeric_value"] != "3" {
		t.Fatalf("category cell = %#v, want numeric_value 3", cell)
	}
	if total := categoryTotals["academic"].(map[string]any)["technique"].(map[string]any)["numeric_value"]; total != "3" {
		t.Fatalf("category total = %v, want 3", total)
	}
}
