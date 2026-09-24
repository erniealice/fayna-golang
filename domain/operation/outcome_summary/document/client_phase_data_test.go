package document

import (
	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"reflect"
	"testing"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	categorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplatetaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	phaseoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	templatecriteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

const testProgressReportPhaseCode = "progress_report"

func TestBuildClientPhaseReportDataMapsAllCategoriesAndNestedAssessments(t *testing.T) {
	card := clientPhaseProjectionFixture()
	data, err := buildClientPhaseReportData(&Deps{DocumentHeaderName: "Sample School"}, card, testProgressReportPhaseCode, "Teacher A", "2026-09-23")
	if err != nil {
		t.Fatalf("buildClientPhaseReportData() error = %v", err)
	}

	for key, want := range map[string]any{
		"school_name":   "Sample School",
		"academic_year": "AY 2026-27",
		"student_name":  "Año, N",
		"section_name":  "Grade 10",
		"printed_by":    "Teacher A",
		"printed_at":    "2026-09-23",
		"phase_name":    "Progress Report",
	} {
		if got := data[key]; !reflect.DeepEqual(got, want) {
			t.Errorf("data[%q] = %#v, want %#v", key, got, want)
		}
	}

	jobs, ok := data["jobs"].([]any)
	if !ok || len(jobs) != 4 {
		t.Fatalf("jobs = %#v, want four active jobs across categories", data["jobs"])
	}
	orderedNames := []string{"Art", "Art", "Biology", "Chemistry"}
	for i, want := range orderedNames {
		row := jobs[i].(map[string]any)
		if got := row["job_name"]; got != want {
			t.Errorf("jobs[%d].job_name = %v, want %q", i, got, want)
		}
	}

	artAssessments := jobs[0].(map[string]any)["assessments"].([]any)
	if len(artAssessments) != 3 {
		t.Fatalf("Art assessments = %#v, want three nested criteria across ordered tasks", artAssessments)
	}
	if got := artAssessments[0].(map[string]any)["assessment_name"]; got != "Planning" {
		t.Errorf("first assessment = %v, want first task's Planning assessment", got)
	}
	graded := artAssessments[1].(map[string]any)
	if graded["assessment_name"] != "Technique" || graded["achievement_level"] != "Proficient" || graded["maximum"] != "4" || graded["comment"] != "Careful work" {
		t.Errorf("graded assessment = %#v, want mapped mark, maximum and comment", graded)
	}
	blank := artAssessments[2].(map[string]any)
	if blank["assessment_name"] != "Reflection" || blank["achievement_level"] != "" || blank["maximum"] != "3" || blank["comment"] != "" {
		t.Errorf("ungraded assessment = %#v, want a defined blank assessment", blank)
	}
	if got := jobs[0].(map[string]any)["phase_comment"]; got != "Strong progress" {
		t.Errorf("phase_comment = %v, want phase summary narrative", got)
	}
	sections, ok := data["outcome_sections"].([]any)
	if !ok || len(sections) != 1 {
		t.Fatalf("outcome_sections = %#v, want one evidence-backed category section", data["outcome_sections"])
	}
	section := sections[0].(map[string]any)
	if section["section_code"] != "academic" || section["section_name"] != "Academic" {
		t.Errorf("outcome section = %#v, want projected category code/name", section)
	}
	rows := section["rows"].([]any)
	if len(rows) != 3 {
		t.Fatalf("outcome rows = %#v, want configured criteria including blank outcomes", rows)
	}
	row := rows[1].(map[string]any)
	if row["row_name"] != "Technique" || row["total"] != "3" {
		t.Errorf("outcome row = %#v, want Technique and recorded numeric total 3", row)
	}
	cells := row["cells"].([]any)
	if len(cells) != 1 || cells[0].(map[string]any)["period_code"] != testProgressReportPhaseCode || cells[0].(map[string]any)["value"] != "Proficient" {
		t.Errorf("outcome cells = %#v, want projected period and scaled value", cells)
	}
}

func TestBuildProjectedOutcomeSectionsSpansProjectedPhasesAndRetainsZero(t *testing.T) {
	card := clientPhaseProjectionFixture()
	card.JobTemplatePhases = append(card.JobTemplatePhases, &jobtemplatephasepb.JobTemplatePhase{
		Id: "template-phase-art-final", JobTemplateId: "template-art", Name: "Second Term", Code: ptr("term_two"), PhaseOrder: 2, Active: true,
	})
	card.JobTemplateTasks[0].Code = ptr("m07")
	card.JobTemplateTasks = append(card.JobTemplateTasks, &jobtemplatetaskpb.JobTemplateTask{
		Id: "template-task-m08", JobTemplatePhaseId: "template-phase-art", Name: "August", Code: ptr("m08"), StepOrder: 2, Active: true,
	})
	card.TemplateTaskCriteria = append(card.TemplateTaskCriteria, &templatecriteriapb.TemplateTaskCriteria{
		Id: "link-m08", JobTemplateTaskId: "template-task-m08", OutcomeCriteriaId: "criterion-technique", SequenceOrder: 1, Active: true,
	})
	card.JobPhases = append(card.JobPhases, &jobphasepb.JobPhase{
		Id: "phase-art-final", JobId: "job-art", TemplatePhaseId: ptr("template-phase-art-final"), Active: true,
	})
	card.JobTasks = append(card.JobTasks, &jobtaskpb.JobTask{
		Id: "task-art-m08", JobPhaseId: "phase-art", TemplateTaskId: ptr("template-task-m08"), Active: true,
	})
	card.JobTasks = append(card.JobTasks, &jobtaskpb.JobTask{
		Id: "task-art-final", JobPhaseId: "phase-art-final", TemplateTaskId: ptr("template-task-art"), Active: true,
	})
	card.TaskOutcomes = append(card.TaskOutcomes, &exportpb.ClientReportCardTaskOutcome{
		JobTaskId: "task-art-m08", TemplateTaskCriteriaId: "link-m08", NumericValue: ptr(float64(2)), RecordedDate: ptr(int64(55)),
	}, &exportpb.ClientReportCardTaskOutcome{
		JobTaskId: "task-art-final", TemplateTaskCriteriaId: "link-technique", NumericValue: ptr(float64(0)), RecordedDate: ptr(int64(60)),
	})
	sections := buildProjectedOutcomeSections(card, false)
	section := sections[0].(map[string]any)
	rows := section["rows"].([]any)
	if len(rows) != 3 {
		t.Fatalf("rows = %#v, want all configured criteria across periods", rows)
	}
	row := rows[1].(map[string]any)
	cells := row["cells"].([]any)
	if len(cells) != 3 {
		t.Fatalf("cells = %#v, want phase/task pairs across projected periods", cells)
	}
	if cells[0].(map[string]any)["period_code"] != testProgressReportPhaseCode || cells[0].(map[string]any)["task_code"] != "m07" ||
		cells[1].(map[string]any)["period_code"] != testProgressReportPhaseCode || cells[1].(map[string]any)["task_code"] != "m08" ||
		cells[2].(map[string]any)["period_code"] != "term_two" || cells[2].(map[string]any)["task_code"] != "m07" {
		t.Errorf("phase/task cells = %#v, want deterministic phase/task pairs", cells)
	}
	if row["total"] != "5" {
		t.Errorf("total = %#v, want recorded values including zero to sum to 5", row["total"])
	}
}

func TestBuildProjectedOutcomeSectionsKeepsStructureWithBlankOutcomes(t *testing.T) {
	card := clientPhaseProjectionFixture()
	card.TaskOutcomes = nil
	sections := buildProjectedOutcomeSections(card, false)
	if len(sections) != 1 {
		t.Fatalf("sections = %#v, want configured category despite missing marks", sections)
	}
	rows := sections[0].(map[string]any)["rows"].([]any)
	if len(rows) != 3 {
		t.Fatalf("rows = %#v, want all configured criteria", rows)
	}
	foundTechnique := false
	for _, raw := range rows {
		row := raw.(map[string]any)
		cells := row["cells"].([]any)
		if len(cells) == 0 {
			t.Errorf("row %q has no structural cells", row["row_name"])
			continue
		}
		for _, rawCell := range cells {
			cell := rawCell.(map[string]any)
			if cell["task_name"] == "Studio" {
				foundTechnique = true
			}
			if cell["value"] != "" {
				t.Errorf("missing outcome value = %#v, want blank", cell["value"])
			}
		}
		if _, exists := row["total"]; exists {
			t.Errorf("row %q manufactured a total without numeric outcomes: %#v", row["row_name"], row["total"])
		}
	}
	if !foundTechnique {
		t.Error("configured Technique task name was not preserved")
	}
}

func TestBuildProjectedOutcomeCellIndexBindsExactCodesAndOmitsMissingCoordinates(t *testing.T) {
	card := clientPhaseProjectionFixture()
	cells, totals, _, _ := buildProjectedOutcomeCellIndex(card, false)

	path := "academic.art-template.technique.progress_report.m07.value"
	if got, ok := resolvePath(cells, path); !ok || got != "Proficient" {
		t.Fatalf("cells.%s = %#v, %v; want exact projected mark", path, got, ok)
	}
	if got, ok := resolvePath(cells, "academic.art-template.technique.progress_report.m07.numeric_value"); !ok || got != "3" {
		t.Fatalf("numeric cell = %#v, %v; want source numeric value 3", got, ok)
	}
	if got, ok := resolvePath(totals, "academic.art-template.technique.numeric_value"); !ok || got != "3" {
		t.Fatalf("total = %#v, %v; want numeric total 3", got, ok)
	}
	if got, ok := resolvePath(cells, "academic.art-template.reflection.progress_report.m07.value"); ok {
		t.Fatalf("missing coordinate resolved to %#v; want absent path", got)
	}
	if got, ok := resolvePath(cells, "other.art-template.technique.progress_report.m07.value"); ok {
		t.Fatalf("cross-category path resolved to %#v; want absent path", got)
	}
}

func TestBuildClientPhaseReportDataExposesDottedRootScalarPaths(t *testing.T) {
	card := clientPhaseProjectionFixture()
	data, err := buildClientPhaseReportData(&Deps{}, card, testProgressReportPhaseCode, "", "")
	if err != nil {
		t.Fatalf("buildClientPhaseReportData() error = %v", err)
	}
	for path, want := range map[string]string{
		"outcome_cells.academic.art-template.technique.progress_report.m07.numeric_value": "3",
		"outcome_totals.academic.art-template.technique.numeric_value":                    "3",
	} {
		got, ok := resolvePath(data, path)
		if !ok || got != want {
			t.Errorf("root scalar %s = %#v, %v; want %q", path, got, ok, want)
		}
	}
	for _, path := range []string{
		"outcome_cells.academic.art-template.technique.progress_report.m08.numeric_value",
		"outcome_cells.other.art-template.technique.progress_report.m07.numeric_value",
		"outcome_totals.other.art-template.technique.numeric_value",
	} {
		if got, ok := resolvePath(data, path); ok {
			t.Errorf("missing root scalar %s = %#v; want absent for blank-safe DOCX replacement", path, got)
		}
	}
}

func TestBuildProjectedOutcomeCellIndexSuppressesConflictingDuplicateTuple(t *testing.T) {
	card := clientPhaseProjectionFixture()
	clientID := card.GetClient().GetClientId()
	card.JobTasks = append(card.JobTasks, &jobtaskpb.JobTask{
		Id: "task-art-2", JobPhaseId: "phase-art-2", TemplateTaskId: ptr("template-task-art"), Active: true,
	})
	card.TaskOutcomes = append(card.TaskOutcomes, &exportpb.ClientReportCardTaskOutcome{
		JobTaskId: "task-art-2", TemplateTaskCriteriaId: "link-technique", NumericValue: ptr(float64(4)),
		ScaledLabel: ptr("Advanced"), RecordedDate: ptr(int64(50)),
	})
	if clientID == "" {
		t.Fatal("fixture must have a client identity")
	}
	cells, totals, _, _ := buildProjectedOutcomeCellIndex(card, false)
	if got, ok := resolvePath(cells, "academic.art-template.technique.progress_report.m07.value"); ok {
		t.Fatalf("conflicting duplicate cell resolved to %#v; want omitted", got)
	}
	if got, ok := resolvePath(totals, "academic.art-template.technique.numeric_value"); ok {
		t.Fatalf("conflicting duplicate total resolved to %#v; want omitted", got)
	}
}

func TestBuildProjectedOutcomeCellIndexDeduplicatesIdenticalCellValues(t *testing.T) {
	card := clientPhaseProjectionFixture()
	card.JobTasks = append(card.JobTasks, &jobtaskpb.JobTask{
		Id: "task-art-2", JobPhaseId: "phase-art-2", TemplateTaskId: ptr("template-task-art"), Active: true,
	})
	card.TaskOutcomes = append(card.TaskOutcomes, &exportpb.ClientReportCardTaskOutcome{
		JobTaskId: "task-art-2", TemplateTaskCriteriaId: "link-technique", NumericValue: ptr(float64(3)),
		ScaledLabel: ptr("Proficient"), RecordedDate: ptr(int64(50)),
	})
	cells, totals, _, _ := buildProjectedOutcomeCellIndex(card, false)
	path := "academic.art-template.technique.progress_report.m07"
	if got, ok := resolvePath(cells, path+".value"); !ok || got != "Proficient" {
		t.Fatalf("identical duplicate cell = %#v, %v; want deterministic deduplicated mark", got, ok)
	}
	if got, ok := resolvePath(cells, path+".numeric_value"); !ok || got != "3" {
		t.Fatalf("identical duplicate numeric cell = %#v, %v; want value 3", got, ok)
	}
	if got, ok := resolvePath(totals, "academic.art-template.technique.numeric_value"); ok {
		t.Fatalf("duplicate-job total = %#v; want omitted because summing identical records would be ambiguous", got)
	}
}

func TestBuildProjectedOutcomeCellIndexDoesNotDoubleCountRepeatedSourceTask(t *testing.T) {
	card := clientPhaseProjectionFixture()
	// A projection join can repeat the same task row. Its cell identity and
	// numeric contribution must remain one source assessment, not two.
	card.JobTasks = append(card.JobTasks, card.JobTasks[0])
	cells, totals, _, _ := buildProjectedOutcomeCellIndex(card, false)
	if got, ok := resolvePath(cells, "academic.art-template.technique.progress_report.m07.numeric_value"); !ok || got != "3" {
		t.Fatalf("repeated source task cell = %#v, %v; want 3", got, ok)
	}
	if got, ok := resolvePath(totals, "academic.art-template.technique.numeric_value"); !ok || got != "3" {
		t.Fatalf("repeated source task total = %#v, %v; want one contribution of 3", got, ok)
	}
}

func TestBuildProjectedOutcomeCellIndexEscapesDotsAndIgnoresForeignClientJobs(t *testing.T) {
	card := clientPhaseProjectionFixture()
	card.JobCategories[0].Code = ptr("other.category")
	cells, _, _, _ := buildProjectedOutcomeCellIndex(card, false)
	if got, ok := resolvePath(cells, "academic.art-template.technique.progress_report.m07.value"); !ok || got != "Proficient" {
		t.Fatalf("own-client cell = %#v, %v; want unaffected academic value", got, ok)
	}

	card.JobCategories[1].Code = ptr("academic.with.dot")
	cells, _, _, _ = buildProjectedOutcomeCellIndex(card, false)
	if got, ok := resolvePath(cells, "academic~1with~1dot.art-template.technique.progress_report.m07.value"); !ok || got != "Proficient" {
		t.Fatalf("escaped category cell = %#v, %v; want reversible dotted-code path", got, ok)
	}

	foreignClientID := "another-client"
	card.Jobs = append(card.Jobs, &jobpb.Job{
		Id: "foreign-job", Name: "Foreign", JobTemplateId: ptr("template-art"), ClientId: &foreignClientID, Active: true,
	})
	card.JobPhases = append(card.JobPhases, &jobphasepb.JobPhase{
		Id: "foreign-phase", JobId: "foreign-job", TemplatePhaseId: ptr("template-phase-art"), Active: true,
	})
	card.JobTasks = append(card.JobTasks, &jobtaskpb.JobTask{
		Id: "foreign-task", JobPhaseId: "foreign-phase", TemplateTaskId: ptr("template-task-art"), Active: true,
	})
	card.TaskOutcomes = append(card.TaskOutcomes, &exportpb.ClientReportCardTaskOutcome{
		JobTaskId: "foreign-task", TemplateTaskCriteriaId: "link-technique", NumericValue: ptr(float64(99)),
		ScaledLabel: ptr("Foreign marker"), RecordedDate: ptr(int64(99)),
	})
	cells, totals, _, _ := buildProjectedOutcomeCellIndex(card, false)
	if got, ok := resolvePath(cells, "academic~1with~1dot.art-template.technique.progress_report.m07.value"); !ok || got != "Proficient" {
		t.Fatalf("foreign job contaminated cell = %#v, %v; want own-client value", got, ok)
	}
	if got, ok := resolvePath(totals, "academic~1with~1dot.art-template.technique.numeric_value"); !ok || got != "3" {
		t.Fatalf("foreign job contaminated total = %#v, %v; want own-client total", got, ok)
	}
}

func TestBuildClientPhaseReportDataUsesJobIDAsStableTieBreak(t *testing.T) {
	card := clientPhaseProjectionFixture()
	// Two jobs with the same category and template sort by stable job identity,
	// regardless of the incoming projection order.
	card.Jobs[0], card.Jobs[1] = card.Jobs[1], card.Jobs[0]
	card.JobTemplateTasks[0], card.JobTemplateTasks[1] = card.JobTemplateTasks[1], card.JobTemplateTasks[0]
	card.TemplateTaskCriteria[0], card.TemplateTaskCriteria[2] = card.TemplateTaskCriteria[2], card.TemplateTaskCriteria[0]
	data, err := buildClientPhaseReportData(&Deps{}, card, testProgressReportPhaseCode, "", "")
	if err != nil {
		t.Fatalf("buildClientPhaseReportData() error = %v", err)
	}
	jobs := data["jobs"].([]any)
	gotNames := make([]string, 0, len(jobs))
	for _, job := range jobs {
		gotNames = append(gotNames, job.(map[string]any)["job_name"].(string))
	}
	if want := []string{"Art", "Art", "Biology", "Chemistry"}; !reflect.DeepEqual(gotNames, want) {
		t.Errorf("ordered jobs = %v, want %v", gotNames, want)
	}
	assessments := jobs[0].(map[string]any)["assessments"].([]any)
	assessmentNames := make([]string, 0, len(assessments))
	for _, assessment := range assessments {
		assessmentNames = append(assessmentNames, assessment.(map[string]any)["assessment_name"].(string))
	}
	if want := []string{"Planning", "Technique", "Reflection"}; !reflect.DeepEqual(assessmentNames, want) {
		t.Errorf("ordered assessments = %v, want %v", assessmentNames, want)
	}
}

func TestBuildClientPhaseReportDataOmitsJobsWithoutSelectedActivePhase(t *testing.T) {
	card := clientPhaseProjectionFixture()
	for _, phase := range card.JobTemplatePhases {
		if phase.GetId() == "template-phase-chem" {
			phase.Active = false
		}
	}
	data, err := buildClientPhaseReportData(&Deps{}, card, testProgressReportPhaseCode, "", "")
	if err != nil {
		t.Fatalf("buildClientPhaseReportData() error = %v", err)
	}
	jobs := data["jobs"].([]any)
	if len(jobs) != 3 {
		t.Fatalf("jobs length = %d, want chemistry omitted because its selected template phase is inactive", len(jobs))
	}
}

func TestBuildClientPhaseReportDataRejectsMissingPhaseAndMalformedProjection(t *testing.T) {
	tests := []struct {
		name string
		card *exportpb.ClientReportCardProjection
	}{
		{name: "nil projection"},
		{name: "missing context", card: &exportpb.ClientReportCardProjection{Client: &exportpb.ClientReportCardClient{ClientId: "client-1"}}},
		{name: "missing client", card: &exportpb.ClientReportCardProjection{Context: &exportpb.SubscriptionGroupOutcomeExportContext{SubscriptionGroupId: "group-1"}}},
		{name: "job belongs to another client", card: func() *exportpb.ClientReportCardProjection {
			card := clientPhaseProjectionFixture()
			card.Jobs[0].ClientId = ptr("client-other")
			return card
		}()},
		{name: "no requested phase", card: func() *exportpb.ClientReportCardProjection {
			card := clientPhaseProjectionFixture()
			for _, phase := range card.JobTemplatePhases {
				phase.Code = ptr("term_1")
			}
			return card
		}()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := buildClientPhaseReportData(&Deps{}, tt.card, testProgressReportPhaseCode, "", ""); err == nil {
				t.Fatal("buildClientPhaseReportData() error = nil, want rejection")
			}
		})
	}
}

func TestBuildClientPhaseReportDataSelectsAnyActiveClientPhaseCode(t *testing.T) {
	card := clientPhaseProjectionFixture()
	for _, phase := range card.JobTemplatePhases {
		phase.Code = ptr("s1")
		phase.Name = "Semester 1"
	}
	data, err := buildClientPhaseReportData(&Deps{}, card, "s1", "", "")
	if err != nil {
		t.Fatalf("buildClientPhaseReportData(s1) error = %v", err)
	}
	if got := data["phase_name"]; got != "Semester 1" {
		t.Fatalf("phase_name = %v, want selected phase display name", got)
	}
	if jobs, ok := data["jobs"].([]any); !ok || len(jobs) != 4 {
		t.Fatalf("jobs = %#v, want four selected client jobs", data["jobs"])
	}
}

func TestStaffNamesUsesTheSelectedJobPhase(t *testing.T) {
	card := clientPhaseProjectionFixture()
	card.TeacherAssignments = []*exportpb.ClientReportCardTeacherAssignment{
		{JobId: "job-art", JobPhaseId: "phase-art", DisplayName: "Selected Teacher"},
		{JobId: "job-art", JobPhaseId: "phase-art-2", DisplayName: "Other Phase Teacher"},
		{JobId: "job-bio", JobPhaseId: "phase-art", DisplayName: "Mismatched Job Teacher"},
	}
	if got := staffNames(card, "job-art", "phase-art"); got != "Selected Teacher" {
		t.Fatalf("staffNames() = %q, want selected phase teacher only", got)
	}
}

func TestFormatClientReportMaximumPreservesFractionalValues(t *testing.T) {
	if got := formatClientReportMaximum(2.5); got != "2.5" {
		t.Fatalf("formatClientReportMaximum(2.5) = %q, want 2.5", got)
	}
}

func clientPhaseProjectionFixture() *exportpb.ClientReportCardProjection {
	clientID := "client-1"
	categoryAcademic, categoryOther := &categorypb.JobCategory{Id: "cat-academic", Name: "Academic", Code: ptr("academic"), SortOrder: ptr(int32(1))}, &categorypb.JobCategory{Id: "cat-other", Name: "Other", Code: ptr("other"), SortOrder: ptr(int32(2))}
	return &exportpb.ClientReportCardProjection{
		Context:       &exportpb.SubscriptionGroupOutcomeExportContext{SubscriptionGroupId: "group-1", SubscriptionGroupName: "Grade 10", PriceScheduleName: "AY 2026-27"},
		Client:        &exportpb.ClientReportCardClient{ClientId: clientID, Name: "Año, N"},
		JobCategories: []*categorypb.JobCategory{categoryOther, categoryAcademic},
		JobTemplates: []*jobtemplatepb.JobTemplate{
			{Id: "template-bio", Name: "Biology", JobCategoryId: ptr("cat-academic"), Active: true},
			{Id: "template-chem", Name: "Chemistry", JobCategoryId: ptr("cat-other"), Active: true},
			{Id: "template-art", Name: "Art", TemplateCode: ptr("art-template"), JobCategoryId: ptr("cat-academic"), Active: true},
		},
		JobTemplatePhases: []*jobtemplatephasepb.JobTemplatePhase{
			{Id: "template-phase-bio", JobTemplateId: "template-bio", Name: "Progress Report", Code: ptr("progress_report"), Active: true},
			{Id: "template-phase-chem", JobTemplateId: "template-chem", Name: "Progress Report", Code: ptr("progress_report"), Active: true},
			{Id: "template-phase-art", JobTemplateId: "template-art", Name: "Progress Report", Code: ptr("progress_report"), PhaseOrder: 1, Active: true},
		},
		Jobs: []*jobpb.Job{
			{Id: "job-bio", Name: "Biology", JobTemplateId: ptr("template-bio"), ClientId: &clientID, Active: true},
			{Id: "job-chem", Name: "Chemistry", JobTemplateId: ptr("template-chem"), ClientId: &clientID, Active: true},
			{Id: "job-art-2", Name: "Art", JobTemplateId: ptr("template-art"), ClientId: &clientID, Active: true},
			{Id: "job-art", Name: "Art", JobTemplateId: ptr("template-art"), ClientId: &clientID, Active: true},
		},
		JobPhases: []*jobphasepb.JobPhase{
			{Id: "phase-bio", JobId: "job-bio", TemplatePhaseId: ptr("template-phase-bio"), Active: true},
			{Id: "phase-chem", JobId: "job-chem", TemplatePhaseId: ptr("template-phase-chem"), Active: true},
			{Id: "phase-art-2", JobId: "job-art-2", TemplatePhaseId: ptr("template-phase-art"), Active: true},
			{Id: "phase-art", JobId: "job-art", TemplatePhaseId: ptr("template-phase-art"), Active: true},
		},
		JobTemplateTasks: []*jobtemplatetaskpb.JobTemplateTask{
			{Id: "template-task-art", JobTemplatePhaseId: "template-phase-art", Name: "Studio", Code: ptr("m07"), StepOrder: 1, Active: true},
			{Id: "template-task-planning", JobTemplatePhaseId: "template-phase-art", Name: "Planning", StepOrder: 0, Active: true},
		},
		JobTasks: []*jobtaskpb.JobTask{
			{Id: "task-art", JobPhaseId: "phase-art", TemplateTaskId: ptr("template-task-art"), Active: true},
			{Id: "task-planning", JobPhaseId: "phase-art", TemplateTaskId: ptr("template-task-planning"), Active: true},
		},
		OutcomeCriteria: []*criteriapb.OutcomeCriteria{
			{Id: "criterion-technique", Name: "Technique", Code: ptr("technique"), MaxScore: ptr(int32(4)), Active: true},
			{Id: "criterion-reflection", Name: "Reflection", MaxScore: ptr(int32(3)), Active: true},
			{Id: "criterion-planning", Name: "Planning", MaxScore: ptr(int32(2)), Active: true},
		},
		TemplateTaskCriteria: []*templatecriteriapb.TemplateTaskCriteria{
			{Id: "link-technique", JobTemplateTaskId: "template-task-art", OutcomeCriteriaId: "criterion-technique", SequenceOrder: 1, Active: true},
			{Id: "link-reflection", JobTemplateTaskId: "template-task-art", OutcomeCriteriaId: "criterion-reflection", SequenceOrder: 2, Active: true},
			{Id: "link-planning", JobTemplateTaskId: "template-task-planning", OutcomeCriteriaId: "criterion-planning", SequenceOrder: 1, Active: true},
		},
		TaskOutcomes: []*exportpb.ClientReportCardTaskOutcome{
			{JobTaskId: "task-art", TemplateTaskCriteriaId: "link-technique", NumericValue: ptr(float64(3)), ScaledLabel: ptr("Proficient"), DeterminationNote: ptr("Careful work"), RecordedDate: ptr(int64(50))},
		},
		PhaseOutcomeSummaries: []*phaseoutcomepb.PhaseOutcomeSummary{
			{Id: "summary-art", JobPhaseId: "phase-art", ScaledLabel: ptr("Meeting"), Narrative: ptr("Strong progress"), Active: true},
		},
	}
}

func ptr[T any](value T) *T { return &value }

// Owner 2026-09-24: the phase document's repeated subject pages carry only the
// configured job categories; each job page also carries the identity copies.
func TestBuildClientPhaseReportDataLimitsJobLoopToConfiguredCategories(t *testing.T) {
	card := clientPhaseProjectionFixture()
	deps := &Deps{DocOptions: outcome_summary.DocumentOptions{PhaseJobCategoryCodes: []string{"ACADEMIC"}}}
	data, err := buildClientPhaseReportData(deps, card, testProgressReportPhaseCode, "", "")
	if err != nil {
		t.Fatalf("buildClientPhaseReportData() error = %v", err)
	}
	jobs := data["jobs"].([]any)
	if len(jobs) == 0 {
		t.Fatal("no jobs")
	}
	for _, raw := range jobs {
		job := raw.(map[string]any)
		if job["job_category_name"] != "Academic" {
			t.Fatalf("job outside the configured categories: %#v", job)
		}
		if job["page_student_name"] != data["student_name"] || job["page_section_name"] != data["section_name"] {
			t.Fatalf("job page identity missing: %#v", job)
		}
	}
	all, err := buildClientPhaseReportData(&Deps{}, card, testProgressReportPhaseCode, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(all["jobs"].([]any)) <= len(jobs) {
		t.Fatalf("unfiltered build has %d jobs, filtered %d; want the filter to drop the non-academic job", len(all["jobs"].([]any)), len(jobs))
	}
}

func TestTrimTrailingQualifier(t *testing.T) {
	for in, want := range map[string]string{
		"Palladium (AY 2026-27)": "Palladium",
		"Palladium":              "Palladium",
		"(AY 2026-27)":           "(AY 2026-27)",
		" Nickel (A) ":           "Nickel",
	} {
		if got := trimTrailingQualifier(in); got != want {
			t.Errorf("trimTrailingQualifier(%q) = %q, want %q", in, got, want)
		}
	}
}

// Owner 2026-09-24: the program-year label comes from the group's plan
// attribute (configured code), copied onto every job page.
func TestBuildClientPhaseReportDataPlanLabelFromPlanAttribute(t *testing.T) {
	card := clientPhaseProjectionFixture()
	card.PlanAttributes = []*exportpb.ClientReportCardAttribute{{Code: "program_year", Value: "Year 5"}, {Code: "other", Value: "x"}}
	data, err := buildClientPhaseReportData(&Deps{DocOptions: outcome_summary.DocumentOptions{PlanLabelAttributeCode: "program_year"}}, card, testProgressReportPhaseCode, "", "")
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range data["jobs"].([]any) {
		if got := raw.(map[string]any)["page_plan_label"]; got != "Year 5" {
			t.Fatalf("page_plan_label = %v, want Year 5", got)
		}
	}
	blank, _ := buildClientPhaseReportData(&Deps{}, card, testProgressReportPhaseCode, "", "")
	if got := blank["jobs"].([]any)[0].(map[string]any)["page_plan_label"]; got != "" {
		t.Fatalf("unconfigured plan label = %v, want blank", got)
	}
}
