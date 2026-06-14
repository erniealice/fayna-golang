package evaluation_cycle

import (
	"context"
	"log"
	"net/http"

	cycleform "github.com/erniealice/fayna-golang/domain/operation/evaluation_cycle/form"

	"github.com/erniealice/pyeza-golang/view"

	cyclepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_cycle"
)

// NewAddAction creates the cycle create action (GET = drawer form, POST = create).
func NewAddAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation_cycle", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("evaluation_cycle-drawer-form", &cycleform.Data{
				FormAction:   deps.Routes.AddURL,
				Labels:       deps.Labels,
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		// POST — create cycle (status defaults to OPEN server-side; NO member
		// materialize here — enrolment happens via OpenEvaluationCycle).
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}
		r := viewCtx.Request

		_, err := deps.CreateEvaluationCycle(ctx, &cyclepb.CreateEvaluationCycleRequest{
			Data: &cyclepb.EvaluationCycle{
				Name:           r.FormValue("name"),
				SubscriptionId: r.FormValue("subscription_id"),
				PeriodStart:    r.FormValue("period_start"),
				PeriodEnd:      r.FormValue("period_end"),
				SignOffDueDate: strPtrIfNotEmpty(r.FormValue("sign_off_due_date")),
				CloseDate:      strPtrIfNotEmpty(r.FormValue("close_date")),
				Active:         true,
			},
		})
		if err != nil {
			log.Printf("Failed to create evaluation cycle: %v", err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("evaluation_cycle-list-table")
	})
}

// strPtrIfNotEmpty returns a pointer to s if non-empty, otherwise nil (for the
// SignOffDueDate / CloseDate proto oneof string fields).
func strPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// NewOpenAction opens a cycle (POST only). Maps to the espyna orchestration
// OpenUseCase — idempotent member enrolment over ACTIVE seats (INSERT … ON
// CONFLICT DO NOTHING), NOT DRAFT materialize (STR-2). Gates
// evaluation_cycle:update (+ CR-5 at the use-case layer).
func NewOpenAction(deps *ModuleDeps) view.View {
	return lifecycleAction(deps, "open")
}

// NewCloseAction closes a cycle (POST only). Maps to the espyna orchestration
// CloseUseCase. Gates evaluation_cycle:update (+ CR-5 at the use-case layer).
func NewCloseAction(deps *ModuleDeps) view.View {
	return lifecycleAction(deps, "close")
}

func lifecycleAction(deps *ModuleDeps, verb string) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation_cycle", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		req := &cyclepb.UpdateEvaluationCycleRequest{
			Data: &cyclepb.EvaluationCycle{Id: id},
		}

		var err error
		switch verb {
		case "open":
			_, err = deps.OpenEvaluationCycle(ctx, req)
		case "close":
			_, err = deps.CloseEvaluationCycle(ctx, req)
		}
		if err != nil {
			log.Printf("Failed to %s evaluation cycle %s: %v", verb, id, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("evaluation_cycle-list-table")
	})
}
