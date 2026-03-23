package list

import (
	"context"
	"fmt"
	"log"

	fayna "github.com/erniealice/fayna-golang"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	outcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
)

// ListViewDeps holds view dependencies.
type ListViewDeps struct {
	Routes           fayna.TaskOutcomeRoutes
	ListTaskOutcomes func(ctx context.Context, req *outcomepb.ListTaskOutcomesRequest) (*outcomepb.ListTaskOutcomesResponse, error)
	Labels           fayna.TaskOutcomeLabels
	CommonLabels     pyeza.CommonLabels
	TableLabels      types.TableLabels
}

// PageData holds the data for the task outcome list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the task outcome list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}

		resp, err := deps.ListTaskOutcomes(ctx, &outcomepb.ListTaskOutcomesRequest{})
		if err != nil {
			log.Printf("Failed to list task outcomes: %v", err)
			return view.Error(fmt.Errorf("failed to load outcomes: %w", err))
		}

		l := deps.Labels
		columns := outcomeColumns(l)
		rows := buildTableRows(resp.GetData(), l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "task-outcomes-table",
			Columns:              columns,
			Rows:                 rows,
			ShowSearch:           true,
			ShowActions:          true,
			ShowSort:             true,
			ShowColumns:          true,
			ShowDensity:          true,
			ShowEntries:          true,
			DefaultSortColumn:    "date",
			DefaultSortDirection: "desc",
			Labels:               deps.TableLabels,
			EmptyState: types.TableEmptyState{
				Title:   l.Empty.Title,
				Message: l.Empty.Message,
			},
			PrimaryAction: &types.PrimaryAction{
				Label:           l.Buttons.RecordOutcome,
				ActionURL:       deps.Routes.AddURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("task_outcome", "create"),
				DisabledTooltip: l.Errors.PermissionDenied,
			},
		}
		types.ApplyTableSettings(tableConfig)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          l.Page.Heading,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   deps.Routes.ActiveSubNav,
				HeaderTitle:    l.Page.Heading,
				HeaderSubtitle: l.Page.Caption,
				HeaderIcon:     "icon-clipboard-check",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "task-outcome-list-content",
			Table:           tableConfig,
		}

		return view.OK("task-outcome-list", pageData)
	})
}

func outcomeColumns(l fayna.TaskOutcomeLabels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "task", Label: l.Columns.Task, Sortable: true},
		{Key: "criteria", Label: l.Columns.Criteria, Sortable: true},
		{Key: "value", Label: l.Columns.Value, Sortable: false, Width: "120px"},
		{Key: "determination", Label: l.Columns.Determination, Sortable: true, Width: "150px"},
		{Key: "recorded_by", Label: l.Columns.RecordedBy, Sortable: true, Width: "150px"},
		{Key: "date", Label: l.Columns.Date, Sortable: true, Width: "150px"},
	}
}

func buildTableRows(
	items []*outcomepb.TaskOutcome,
	l fayna.TaskOutcomeLabels,
	routes fayna.TaskOutcomeRoutes,
	perms *types.UserPermissions,
) []types.TableRow {
	rows := []types.TableRow{}
	for _, o := range items {
		id := o.GetId()
		determination := determinationString(o.GetDetermination())
		recordedBy := o.GetRecordedByName()
		date := o.GetDateCreatedString()

		// Build task name
		taskName := o.GetJobTaskId()
		if jt := o.GetJobTask(); jt != nil {
			if jt.GetName() != "" {
				taskName = jt.GetName()
			}
		}

		// Build criteria name
		criteriaName := o.GetCriteriaVersionId()
		if cv := o.GetCriteriaVersion(); cv != nil {
			if cv.GetName() != "" {
				criteriaName = cv.GetName()
			}
		}

		// Build displayed value
		value := buildValueDisplay(o)

		detailURL := route.ResolveURL(routes.DetailURL, "id", id)

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: taskName},
				{Type: "text", Value: criteriaName},
				{Type: "text", Value: value},
				{Type: "badge", Value: determination, Variant: determinationVariant(o.GetDetermination())},
				{Type: "text", Value: recordedBy},
				types.DateTimeCell(date, types.DateReadable),
			},
			DataAttrs: map[string]string{
				"task":          taskName,
				"criteria":      criteriaName,
				"value":         value,
				"determination": determination,
				"recorded_by":   recordedBy,
				"date":          date,
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("task_outcome", "update"), DisabledTooltip: l.Errors.PermissionDenied},
				{Type: "delete", Label: l.Actions.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: id, Disabled: !perms.Can("task_outcome", "delete"), DisabledTooltip: l.Errors.PermissionDenied},
			},
		})
	}
	return rows
}

func buildValueDisplay(o *outcomepb.TaskOutcome) string {
	switch o.GetCriteriaType() {
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_RANGE, enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE:
		return fmt.Sprintf("%.2f", o.GetNumericValue())
	case enums.CriteriaType_CRITERIA_TYPE_TEXT:
		return o.GetTextValue()
	case enums.CriteriaType_CRITERIA_TYPE_CATEGORICAL:
		return o.GetCategoricalValue()
	case enums.CriteriaType_CRITERIA_TYPE_PASS_FAIL:
		if o.GetPassFailValue() {
			return "Pass"
		}
		return "Fail"
	default:
		if o.GetTextValue() != "" {
			return o.GetTextValue()
		}
		return fmt.Sprintf("%.2f", o.GetNumericValue())
	}
}

func determinationString(d enums.Determination) string {
	switch d {
	case enums.Determination_DETERMINATION_PASS:
		return "pass"
	case enums.Determination_DETERMINATION_FAIL:
		return "fail"
	case enums.Determination_DETERMINATION_PASS_WITH_CONDITION:
		return "conditional"
	case enums.Determination_DETERMINATION_NOT_EVALUATED:
		return "not_evaluated"
	case enums.Determination_DETERMINATION_NOT_APPLICABLE:
		return "n_a"
	case enums.Determination_DETERMINATION_DEFERRED:
		return "deferred"
	default:
		return "unspecified"
	}
}

func determinationVariant(d enums.Determination) string {
	switch d {
	case enums.Determination_DETERMINATION_PASS:
		return "success"
	case enums.Determination_DETERMINATION_FAIL:
		return "danger"
	case enums.Determination_DETERMINATION_PASS_WITH_CONDITION:
		return "warning"
	case enums.Determination_DETERMINATION_NOT_EVALUATED:
		return "default"
	case enums.Determination_DETERMINATION_NOT_APPLICABLE:
		return "default"
	case enums.Determination_DETERMINATION_DEFERRED:
		return "info"
	default:
		return "default"
	}
}
