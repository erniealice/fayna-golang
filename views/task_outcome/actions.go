package task_outcome

import (
	"context"
	"log"
	"net/http"
	"strconv"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	outcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
)

// recordingFormOption is a key/label option for the multi-check and categorical select fields.
type recordingFormOption struct {
	Value    string
	Label    string
	Selected bool
}

// recordingFormData is the template data for the task outcome recording form.
type recordingFormData struct {
	FormAction      string
	IsEdit          bool
	ID              string
	TaskID          string
	CriteriaID      string
	CriteriaName    string
	CriteriaType    string
	CriteriaOptions []recordingFormOption
	NumericValue    float64
	TextValue       string
	Notes           string
	PassFailValue   bool
	SelectedOption  string
	ScoreMin        float64
	ScoreMax        float64
	Labels          fayna.TaskOutcomeLabels
	CommonLabels    any
}

// newAddAction creates the task outcome add action (GET = recording form, POST = create).
func newAddAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("task_outcome", "create") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			q := viewCtx.Request.URL.Query()
			return view.OK("task-outcome-recording-form", &recordingFormData{
				FormAction:   deps.Routes.AddURL,
				TaskID:       q.Get("task_id"),
				CriteriaID:   q.Get("criteria_id"),
				CriteriaName: q.Get("criteria_name"),
				CriteriaType: q.Get("criteria_type"),
				Labels:       deps.Labels,
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		// POST — create task outcome
		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		r := viewCtx.Request
		numericValue, _ := strconv.ParseFloat(r.FormValue("numeric_value"), 64)
		passFailRaw := r.FormValue("pass_fail")
		passFailValue := passFailRaw == "true" || passFailRaw == "on"
		textValue := r.FormValue("text_value")

		req := &outcomepb.CreateTaskOutcomeRequest{
			Data: &outcomepb.TaskOutcome{
				JobTaskId:         r.FormValue("task_id"),
				CriteriaVersionId: r.FormValue("criteria_id"),
				Active:            true,
			},
		}

		if numericValue != 0 {
			req.Data.NumericValue = &numericValue
		}
		if textValue != "" {
			req.Data.TextValue = &textValue
		}
		req.Data.PassFailValue = &passFailValue

		if note := r.FormValue("notes"); note != "" {
			req.Data.DeterminationNote = &note
		}

		_, err := deps.CreateTaskOutcome(ctx, req)
		if err != nil {
			log.Printf("Failed to create task outcome: %v", err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("outcomes-table")
	})
}

// newEditAction creates the task outcome edit action (GET = pre-filled recording form, POST = update).
func newEditAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("task_outcome", "update") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")
		if id == "" {
			id = viewCtx.Request.URL.Query().Get("id")
		}

		if viewCtx.Request.Method == http.MethodGet {
			if id == "" {
				return fayna.HTMXError(deps.Labels.Errors.IDRequired)
			}

			readResp, err := deps.ReadTaskOutcome(ctx, &outcomepb.ReadTaskOutcomeRequest{
				Data: &outcomepb.TaskOutcome{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read task outcome %s: %v", id, err)
				return fayna.HTMXError(deps.Labels.Errors.NotFound)
			}
			readData := readResp.GetData()
			if len(readData) == 0 {
				return fayna.HTMXError(deps.Labels.Errors.NotFound)
			}
			record := readData[0]

			criteriaName := ""
			if cv := record.GetCriteriaVersion(); cv != nil {
				criteriaName = cv.GetName()
			}

			return view.OK("task-outcome-recording-form", &recordingFormData{
				FormAction:   route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:       true,
				ID:           id,
				TaskID:       record.GetJobTaskId(),
				CriteriaID:   record.GetCriteriaVersionId(),
				CriteriaName: criteriaName,
				CriteriaType: record.GetCriteriaType().String(),
				NumericValue: record.GetNumericValue(),
				TextValue:    record.GetTextValue(),
				PassFailValue: record.GetPassFailValue(),
				Notes:         record.GetDeterminationNote(),
				Labels:        deps.Labels,
				CommonLabels:  nil, // injected by ViewAdapter
			})
		}

		// POST — update task outcome
		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		r := viewCtx.Request
		if id == "" {
			id = r.FormValue("id")
		}
		if id == "" {
			return fayna.HTMXError(deps.Labels.Errors.IDRequired)
		}

		numericValue, _ := strconv.ParseFloat(r.FormValue("numeric_value"), 64)
		passFailRaw := r.FormValue("pass_fail")
		passFailValue := passFailRaw == "true" || passFailRaw == "on"
		textValue := r.FormValue("text_value")

		req := &outcomepb.UpdateTaskOutcomeRequest{
			Data: &outcomepb.TaskOutcome{
				Id:                id,
				CriteriaVersionId: r.FormValue("criteria_id"),
			},
		}

		if numericValue != 0 {
			req.Data.NumericValue = &numericValue
		}
		if textValue != "" {
			req.Data.TextValue = &textValue
		}
		req.Data.PassFailValue = &passFailValue

		if note := r.FormValue("notes"); note != "" {
			req.Data.DeterminationNote = &note
		}

		_, err := deps.UpdateTaskOutcome(ctx, req)
		if err != nil {
			log.Printf("Failed to update task outcome %s: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		return view.ViewResult{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"HX-Trigger":  `{"formSuccess":true}`,
				"HX-Redirect": route.ResolveURL(deps.Routes.DetailURL, "id", id),
			},
		}
	})
}

// newDeleteAction creates the task outcome delete action (POST only).
func newDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("task_outcome", "delete") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.URL.Query().Get("id")
		if id == "" {
			_ = viewCtx.Request.ParseForm()
			id = viewCtx.Request.FormValue("id")
		}
		if id == "" {
			return fayna.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.DeleteTaskOutcome(ctx, &outcomepb.DeleteTaskOutcomeRequest{
			Data: &outcomepb.TaskOutcome{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete task outcome %s: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("outcomes-table")
	})
}
