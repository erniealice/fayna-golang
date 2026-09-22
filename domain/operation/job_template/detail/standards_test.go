package detail

import (
	"context"
	"testing"

	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplateTaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	templatetaskcriteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
	templateTaskCriteria "github.com/erniealice/fayna-golang/domain/operation/template_task_criteria"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"
)

func TestLoadStandardsTabAddsContextPreservingEditAction(t *testing.T) {
	ctx := view.WithUserPermissions(context.Background(), types.NewUserPermissions([]string{
		"template_task_criteria:list",
		"template_task_criteria:update",
	}))
	deps := &DetailViewDeps{
		CriteriaRoutes: templateTaskCriteria.Routes{
			EditURL: "/action/template-task-criteria/edit/{id}",
		},
		ListPhasesByJobTemplate: func(context.Context, *jobtemplatephasepb.ListByJobTemplateRequest) (*jobtemplatephasepb.ListByJobTemplateResponse, error) {
			return &jobtemplatephasepb.ListByJobTemplateResponse{
				JobTemplatePhases: []*jobtemplatephasepb.JobTemplatePhase{{Id: "phase-1", Name: "Term 1"}},
			}, nil
		},
		ListTasksByPhase: func(context.Context, *jobtemplateTaskpb.ListJobTemplateTasksByPhaseRequest) (*jobtemplateTaskpb.ListJobTemplateTasksByPhaseResponse, error) {
			return &jobtemplateTaskpb.ListJobTemplateTasksByPhaseResponse{
				JobTemplateTasks: []*jobtemplateTaskpb.JobTemplateTask{{Id: "task-1", Name: "Activity 1", StepOrder: 1}},
			}, nil
		},
		ListCriteriaByTask: func(context.Context, *templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskRequest) (*templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskResponse, error) {
			return &templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskResponse{
				TemplateTaskCriterias: []*templatetaskcriteriapb.TemplateTaskCriteria{{
					Id: "binding-1", OutcomeCriteriaId: "criterion-1", SequenceOrder: 1,
				}},
			}, nil
		},
	}

	pageData := &PageData{}
	loadStandardsTab(ctx, deps, pageData, "template-1")

	if pageData.StandardsTable == nil {
		t.Fatal("StandardsTable is nil")
	}
	if !pageData.StandardsTable.ShowActions {
		t.Fatal("StandardsTable.ShowActions = false, want true for update permission")
	}
	if len(pageData.StandardsTable.Rows) != 1 || len(pageData.StandardsTable.Rows[0].Actions) != 1 {
		t.Fatalf("row actions = %#v, want one edit action", pageData.StandardsTable.Rows[0].Actions)
	}
	action := pageData.StandardsTable.Rows[0].Actions[0]
	if action.Type != "edit" || action.HxGet != "/action/template-task-criteria/edit/binding-1?job_template_id=template-1" {
		t.Fatalf("edit action = %#v, want context-preserving HTMX edit URL", action)
	}
	if action.HxTarget != "#sheetContent" || action.DrawerTitle != templateTaskCriteria.DefaultLabels().Actions.Edit {
		t.Fatalf("edit action drawer wiring = %#v, want sheetContent/%s", action, templateTaskCriteria.DefaultLabels().Actions.Edit)
	}
}

func TestLoadStandardsTabHidesEditActionWithoutUpdatePermission(t *testing.T) {
	ctx := view.WithUserPermissions(context.Background(), types.NewUserPermissions([]string{
		"template_task_criteria:list",
	}))
	deps := &DetailViewDeps{
		CriteriaRoutes: templateTaskCriteria.Routes{EditURL: "/action/template-task-criteria/edit/{id}"},
		ListPhasesByJobTemplate: func(context.Context, *jobtemplatephasepb.ListByJobTemplateRequest) (*jobtemplatephasepb.ListByJobTemplateResponse, error) {
			return &jobtemplatephasepb.ListByJobTemplateResponse{JobTemplatePhases: []*jobtemplatephasepb.JobTemplatePhase{{Id: "phase-1"}}}, nil
		},
		ListTasksByPhase: func(context.Context, *jobtemplateTaskpb.ListJobTemplateTasksByPhaseRequest) (*jobtemplateTaskpb.ListJobTemplateTasksByPhaseResponse, error) {
			return &jobtemplateTaskpb.ListJobTemplateTasksByPhaseResponse{JobTemplateTasks: []*jobtemplateTaskpb.JobTemplateTask{{Id: "task-1"}}}, nil
		},
		ListCriteriaByTask: func(context.Context, *templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskRequest) (*templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskResponse, error) {
			return &templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskResponse{TemplateTaskCriterias: []*templatetaskcriteriapb.TemplateTaskCriteria{{Id: "binding-1"}}}, nil
		},
	}

	pageData := &PageData{}
	loadStandardsTab(ctx, deps, pageData, "template-1")

	if pageData.StandardsTable == nil {
		t.Fatal("StandardsTable is nil")
	}
	if pageData.StandardsTable.ShowActions {
		t.Fatal("StandardsTable.ShowActions = true without update/delete permission, want false")
	}
	if len(pageData.StandardsTable.Rows) != 1 || len(pageData.StandardsTable.Rows[0].Actions) != 0 {
		t.Fatalf("row actions = %#v, want none without update/delete permission", pageData.StandardsTable.Rows[0].Actions)
	}
}
