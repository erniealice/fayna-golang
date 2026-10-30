package document

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	categorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	phaseoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

// omitFixture: one client with five jobs over two categories. The client is
// graded only in "Item C". Templates: A and B are parallel offerings (A graded
// for a classmate, B blank for the whole group); A-pair / B-pair are their
// per-item rows in the second category.
func omitFixture() *exportpb.ClientReportCardProjection {
	clientID := "client-1"
	templates := []struct{ id, name, category string }{
		{"t-a", "Item A", "cat-main"},
		{"t-b", "Item B", "cat-main"},
		{"t-c", "Item C", "cat-main"},
		{"t-a-pair", "Item A (Pair)", "cat-pair"},
		{"t-b-pair", "Item B (Pair)", "cat-pair"},
	}
	card := &exportpb.ClientReportCardProjection{
		Context: &exportpb.SubscriptionGroupOutcomeExportContext{SubscriptionGroupId: "group-1"},
		Client:  &exportpb.ClientReportCardClient{ClientId: clientID, Name: "Client One"},
		JobCategories: []*categorypb.JobCategory{
			{Id: "cat-main", Code: ptr("main"), Active: true},
			{Id: "cat-pair", Code: ptr("pair"), Active: true},
		},
	}
	for _, tpl := range templates {
		card.JobTemplates = append(card.JobTemplates, &jobtemplatepb.JobTemplate{Id: tpl.id, Name: tpl.name, JobCategoryId: ptr(tpl.category), Active: true})
		card.JobTemplatePhases = append(card.JobTemplatePhases, &jobtemplatephasepb.JobTemplatePhase{Id: "tp-" + tpl.id, JobTemplateId: tpl.id, Name: "Period One", Code: ptr("p1"), Active: true})
		card.Jobs = append(card.Jobs, &jobpb.Job{Id: "job-" + tpl.id, Name: tpl.name, JobTemplateId: ptr(tpl.id), ClientId: &clientID, Active: true})
		card.JobPhases = append(card.JobPhases, &jobphasepb.JobPhase{Id: "ph-" + tpl.id, JobId: "job-" + tpl.id, TemplatePhaseId: ptr("tp-" + tpl.id), Active: true})
		card.JobTasks = append(card.JobTasks, &jobtaskpb.JobTask{Id: "task-" + tpl.id, JobPhaseId: "ph-" + tpl.id, Active: true})
		// Every job carries the all-zero scaffold and a floor band.
		card.TaskOutcomes = append(card.TaskOutcomes, &exportpb.ClientReportCardTaskOutcome{JobTaskId: "task-" + tpl.id, TemplateTaskCriteriaId: "ttc-" + tpl.id, NumericValue: ptr(float64(0))})
		card.PhaseOutcomeSummaries = append(card.PhaseOutcomeSummaries, &phaseoutcomepb.PhaseOutcomeSummary{Id: "sum-" + tpl.id, JobPhaseId: "ph-" + tpl.id, ScaledLabel: ptr("1"), Active: true})
	}
	card.TaskOutcomes = append(card.TaskOutcomes, &exportpb.ClientReportCardTaskOutcome{JobTaskId: "task-t-c", TemplateTaskCriteriaId: "ttc-t-c-2", NumericValue: ptr(float64(6))})
	return card
}

func gradedCell(templateID string) *exportpb.SubscriptionGroupOutcomeCell {
	return &exportpb.SubscriptionGroupOutcomeCell{JobTemplateId: templateID, JobPresent: true, ScaledLabel: ptr("5"), EnrollmentEvidence: &exportpb.EnrollmentEvidence{HasMarks: true, HasPositiveMark: true}}
}

func blankCell(templateID string) *exportpb.SubscriptionGroupOutcomeCell {
	return &exportpb.SubscriptionGroupOutcomeCell{JobTemplateId: templateID, JobPresent: true, ScaledLabel: ptr("1"), EnrollmentEvidence: &exportpb.EnrollmentEvidence{HasMarks: true}}
}

// omitDeps answers the group read: client-2 is graded in A and C; nobody in B.
func omitDeps(t *testing.T, calls *int, mu *sync.Mutex) *Deps {
	t.Helper()
	resetClassmateCache()
	return &Deps{
		CategoryFilter: "main",
		DocOptions:     outcome_summary.DocumentOptions{PhaseJobCategoryCodes: []string{"main"}},
		GetSubscriptionGroupOutcomeExport: func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
			mu.Lock()
			*calls++
			mu.Unlock()
			if req.GetSubscriptionGroupId() != "group-1" || req.GetJobCategoryId() != "cat-main" {
				t.Errorf("export request = %+v", req)
			}
			return &exportpb.GetSubscriptionGroupOutcomeExportResponse{Success: true, ClientRows: []*exportpb.SubscriptionGroupOutcomeClientRow{
				{ClientId: "client-1", Cells: []*exportpb.SubscriptionGroupOutcomeCell{blankCell("t-a"), blankCell("t-b"), gradedCell("t-c")}},
				{ClientId: "client-2", Cells: []*exportpb.SubscriptionGroupOutcomeCell{gradedCell("t-a"), blankCell("t-b"), gradedCell("t-c")}},
			}}, nil
		},
	}
}

func resetClassmateCache() {
	classmateCache.Lock()
	classmateCache.entries = map[string]*classmateEntry{}
	classmateCache.Unlock()
}

func omittedIDs(omitted []omittedJob) []string {
	ids := []string{}
	for _, o := range omitted {
		ids = append(ids, o.jobID)
	}
	return ids
}

// AC-OMIT-01/02/03/04: omitted only when classmates are graded and the client
// is not; the paired row follows; a blank group keeps its row.
func TestOmitNonEnrolledJobsClassmatesGradedRule(t *testing.T) {
	var calls int
	var mu sync.Mutex
	got := omittedIDs(omitNonEnrolledJobs(context.Background(), omitDeps(t, &calls, &mu), "group-1", omitFixture(), "p1"))
	if want := []string{"job-t-a", "job-t-a-pair"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("omitted = %v, want %v (B kept: group blank; C kept: own evidence)", got, want)
	}
	if calls != 1 {
		t.Errorf("export calls = %d, want one group read", calls)
	}
}

// AC-OMIT-04: a paired row with its own evidence is kept even when its main row
// is omitted.
func TestOmitNonEnrolledJobsKeepsPairWithOwnEvidence(t *testing.T) {
	var calls int
	var mu sync.Mutex
	card := omitFixture()
	card.TaskOutcomes = append(card.TaskOutcomes, &exportpb.ClientReportCardTaskOutcome{JobTaskId: "task-t-a-pair", TemplateTaskCriteriaId: "ttc-t-a-pair-2", NumericValue: ptr(float64(3))})
	got := omittedIDs(omitNonEnrolledJobs(context.Background(), omitDeps(t, &calls, &mu), "group-1", card, "p1"))
	if want := []string{"job-t-a"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("omitted = %v, want %v", got, want)
	}
}

// AC-OMIT-03: a real stored band (> 1) is evidence even without a positive mark.
func TestOmitNonEnrolledJobsKeepsRealStoredBand(t *testing.T) {
	var calls int
	var mu sync.Mutex
	card := omitFixture()
	for _, s := range card.PhaseOutcomeSummaries {
		if s.GetId() == "sum-t-a" {
			s.SummaryScore = ptr(float64(4))
		}
	}
	got := omittedIDs(omitNonEnrolledJobs(context.Background(), omitDeps(t, &calls, &mu), "group-1", card, "p1"))
	if len(got) != 0 {
		t.Fatalf("omitted = %v, want none (A has a stored score)", got)
	}
}

// The client's own row never counts as a classmate.
func TestOmitNonEnrolledJobsIgnoresOwnRowInGroupRead(t *testing.T) {
	resetClassmateCache()
	d := &Deps{
		DocOptions: outcome_summary.DocumentOptions{PhaseJobCategoryCodes: []string{"main"}},
		GetSubscriptionGroupOutcomeExport: func(context.Context, *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
			return &exportpb.GetSubscriptionGroupOutcomeExportResponse{Success: true, ClientRows: []*exportpb.SubscriptionGroupOutcomeClientRow{
				{ClientId: "client-1", Cells: []*exportpb.SubscriptionGroupOutcomeCell{gradedCell("t-a")}},
			}}, nil
		},
	}
	if got := omitNonEnrolledJobs(context.Background(), d, "group-1", omitFixture(), "p1"); len(got) != 0 {
		t.Fatalf("omitted = %v, want none", omittedIDs(got))
	}
}

// AC-OMIT-05: no port, a read error or an unsuccessful read omits nothing.
func TestOmitNonEnrolledJobsFailsSafe(t *testing.T) {
	for name, export := range map[string]func(context.Context, *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error){
		"nil port": nil,
		"error": func(context.Context, *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
			return nil, errors.New("unavailable")
		},
		"unsuccessful": func(context.Context, *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
			return &exportpb.GetSubscriptionGroupOutcomeExportResponse{}, nil
		},
	} {
		t.Run(name, func(t *testing.T) {
			resetClassmateCache()
			d := &Deps{DocOptions: outcome_summary.DocumentOptions{PhaseJobCategoryCodes: []string{"main"}}, GetSubscriptionGroupOutcomeExport: export}
			if got := omitNonEnrolledJobs(context.Background(), d, "group-1", omitFixture(), "p1"); len(got) != 0 {
				t.Fatalf("omitted = %v, want none", omittedIDs(got))
			}
		})
	}
}

// AC-OMIT-06: concurrent documents of one group share a single group read.
func TestOmitNonEnrolledJobsSharesGroupReadAcrossRequests(t *testing.T) {
	var calls int
	var mu sync.Mutex
	d := omitDeps(t, &calls, &mu)
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			omitNonEnrolledJobs(context.Background(), d, "group-1", omitFixture(), "p1")
		}()
	}
	wg.Wait()
	if calls != 1 {
		t.Fatalf("export calls = %d, want 1 shared read", calls)
	}
}

// AC-OMIT-07: the year-final read uses the final selector and the omitted job
// leaves the category loop the block template renders.
func TestOmitNonEnrolledJobsYearFinalDropsCategoryJob(t *testing.T) {
	var calls int
	var mu sync.Mutex
	d := omitDeps(t, &calls, &mu)
	export := d.GetSubscriptionGroupOutcomeExport
	d.GetSubscriptionGroupOutcomeExport = func(ctx context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		if !req.GetFinalOutcome() {
			t.Errorf("year-final read selector = %+v, want final outcome", req.GetOutcomeSelector())
		}
		return export(ctx, req)
	}
	card := omitFixture()
	omitted := omitNonEnrolledJobs(context.Background(), d, "group-1", card, clientDocumentPeriodYearFinal)
	if want := []string{"job-t-a", "job-t-a-pair"}; !reflect.DeepEqual(omittedIDs(omitted), want) {
		t.Fatalf("omitted = %v, want %v", omittedIDs(omitted), want)
	}
	filtered := withoutJobs(card, omitted)
	if len(card.Jobs) != 5 {
		t.Fatalf("source card mutated: %d jobs", len(card.Jobs))
	}
	data := buildProjectedYearFinalData(d, filtered, "", "", testTime())
	categories, _ := data["job_categories"].(map[string]any)
	for code, wantN := range map[string]int{"main": 2, "pair": 1} {
		category, _ := categories[code].(map[string]any)
		if jobs, _ := category["jobs"].([]any); len(jobs) != wantN {
			t.Errorf("job_categories.%s.jobs = %d, want %d", code, len(jobs), wantN)
		}
	}
}

// AC-OMIT-08: the phase document drops the omitted job from jobs and its pair
// from summary_jobs.
func TestOmitNonEnrolledJobsPhaseDocumentRows(t *testing.T) {
	var calls int
	var mu sync.Mutex
	d := omitDeps(t, &calls, &mu)
	d.DocOptions.PeriodSummaryJobCategoryCode = "pair"
	card := omitFixture()
	data, err := buildClientPhaseReportData(d, withoutJobs(card, omitNonEnrolledJobs(context.Background(), d, "group-1", card, "p1")), "p1", "", "")
	if err != nil {
		t.Fatalf("buildClientPhaseReportData() error = %v", err)
	}
	if jobs, _ := data["jobs"].([]any); len(jobs) != 2 {
		t.Errorf("jobs = %d, want 2 (B and C)", len(jobs))
	}
	if rows, _ := data["summary_jobs"].([]any); len(rows) != 1 || rows[0].(map[string]any)["summary_job_name"] != "Item B" {
		t.Errorf("summary_jobs = %#v, want only the Item B pair", data["summary_jobs"])
	}
}

func TestOmittedItemsHeaderValueEncodesNames(t *testing.T) {
	got := omittedItemsHeaderValue([]omittedJob{{name: "Item B"}, {name: "Item A; (Pair)"}})
	if want := "Item%20B; Item%20A%3B%20%28Pair%29"; got != want {
		t.Fatalf("header = %q, want %q", got, want)
	}
}

// AC-OMIT-09: a client with no evidence anywhere is ungraded, not choosing:
// nothing is omitted, so the document never empties.
func TestOmitNonEnrolledJobsKeepsFullyUngradedClient(t *testing.T) {
	var calls int
	var mu sync.Mutex
	card := omitFixture()
	card.TaskOutcomes = card.TaskOutcomes[:len(card.TaskOutcomes)-1]
	if got := omitNonEnrolledJobs(context.Background(), omitDeps(t, &calls, &mu), "group-1", card, "p1"); len(got) != 0 {
		t.Fatalf("omitted = %v, want none", omittedIDs(got))
	}
}
