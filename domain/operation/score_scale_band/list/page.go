package list

import (
	"context"
	"fmt"
	"log"
	"strconv"

	score_scale_band "github.com/erniealice/fayna-golang/domain/operation/score_scale_band"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	ssbpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale_band"
)

// ListViewDeps holds view dependencies.
type ListViewDeps struct {
	Routes              score_scale_band.Routes
	ListScoreScaleBands func(ctx context.Context, req *ssbpb.ListScoreScaleBandsRequest) (*ssbpb.ListScoreScaleBandsResponse, error)
	Labels              score_scale_band.Labels
	CommonLabels        pyeza.CommonLabels
	TableLabels         types.TableLabels
}

// PageData holds the data for the score scale band list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the score scale band list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("score_scale_band", "list") {
			return view.Forbidden("score_scale_band:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}

		resp, err := deps.ListScoreScaleBands(ctx, &ssbpb.ListScoreScaleBandsRequest{})
		if err != nil {
			log.Printf("Failed to list score scale bands: %v", err)
			return view.Error(fmt.Errorf("failed to load score scale bands: %w", err))
		}

		l := deps.Labels
		columns := bandColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "score-scale-band-table",
			Columns:              columns,
			Rows:                 rows,
			ShowSearch:           true,
			ShowActions:          true,
			ShowSort:             true,
			ShowColumns:          true,
			ShowDensity:          true,
			ShowEntries:          true,
			DefaultSortColumn:    "sequence_order",
			DefaultSortDirection: "asc",
			Labels:               deps.TableLabels,
			EmptyState: types.TableEmptyState{
				Title:   emptyTitle(l, status),
				Message: emptyMessage(l, status),
			},
			PrimaryAction: &types.PrimaryAction{
				Label:           l.Buttons.AddBand,
				ActionURL:       deps.Routes.AddURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("score_scale_band", "create"),
				DisabledTooltip: l.Errors.PermissionDenied,
			},
		}
		types.ApplyTableSettings(tableConfig)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          statusPageTitle(l, status),
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   deps.Routes.ActiveSubNav,
				HeaderTitle:    statusPageTitle(l, status),
				HeaderSubtitle: statusPageCaption(l, status),
				HeaderIcon:     "icon-sliders",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "score-scale-band-list-content",
			Table:           tableConfig,
		}

		return view.OK("score-scale-band-list", pageData)
	})
}

func bandColumns(l score_scale_band.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "output_label", Label: l.Columns.OutputLabel},
		{Key: "sequence_order", Label: l.Columns.SequenceOrder, WidthClass: "col-lg", SortKind: "number"},
		{Key: "input_min", Label: l.Columns.InputMin, WidthClass: "col-3xl"},
		{Key: "input_max", Label: l.Columns.InputMax, WidthClass: "col-3xl"},
		{Key: "output_value", Label: l.Columns.OutputValue, WidthClass: "col-3xl"},
		{Key: "determination", Label: l.Columns.Determination, WidthClass: "col-4xl"},
		{Key: "status", Label: l.Columns.Status, WidthClass: "col-3xl"},
	}
}

func buildTableRows(
	items []*ssbpb.ScoreScaleBand,
	status string,
	l score_scale_band.Labels,
	routes score_scale_band.Routes,
	perms *types.UserPermissions,
) []types.TableRow {
	rows := []types.TableRow{}
	for _, b := range items {
		activeStatus := "inactive"
		if b.GetActive() {
			activeStatus = "active"
		}
		if activeStatus != status {
			continue
		}

		id := b.GetId()
		outputLabel := b.GetOutputLabel()
		seqOrder := fmt.Sprintf("%d", b.GetSequenceOrder())
		inputMin := formatOptFloat(b.InputMin)
		inputMax := formatOptFloat(b.InputMax)
		outputValue := formatOptFloat(b.OutputValue)
		determination := determinationString(b.GetDetermination())
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: outputLabel},
				{Type: "number", Value: seqOrder},
				{Type: "text", Value: inputMin},
				{Type: "text", Value: inputMax},
				{Type: "text", Value: outputValue},
				{Type: "badge", Value: determination, Variant: determinationVariant(b.GetDetermination())},
				{Type: "badge", Value: activeStatus, Variant: activeVariant(b.GetActive())},
			},
			DataAttrs: map[string]string{
				"output_label":   outputLabel,
				"sequence_order": seqOrder,
				"input_min":      inputMin,
				"input_max":      inputMax,
				"output_value":   outputValue,
				"determination":  determination,
				"status":         activeStatus,
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("score_scale_band", "update"), DisabledTooltip: l.Errors.PermissionDenied},
				{Type: "delete", Label: l.Actions.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: outputLabel, Disabled: !perms.Can("score_scale_band", "delete"), DisabledTooltip: l.Errors.PermissionDenied},
			},
		})
	}
	return rows
}

func formatOptFloat(v *float64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(*v, 'f', -1, 64)
}

func determinationString(d enums.Determination) string {
	switch d {
	case enums.Determination_DETERMINATION_PASS:
		return "Pass"
	case enums.Determination_DETERMINATION_FAIL:
		return "Fail"
	case enums.Determination_DETERMINATION_PASS_WITH_CONDITION:
		return "Pass with Condition"
	case enums.Determination_DETERMINATION_NOT_EVALUATED:
		return "Not Evaluated"
	case enums.Determination_DETERMINATION_NOT_APPLICABLE:
		return "Not Applicable"
	case enums.Determination_DETERMINATION_DEFERRED:
		return "Deferred"
	default:
		return "Unspecified"
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
	case enums.Determination_DETERMINATION_NOT_EVALUATED, enums.Determination_DETERMINATION_NOT_APPLICABLE, enums.Determination_DETERMINATION_DEFERRED:
		return "default"
	default:
		return "default"
	}
}

func activeVariant(active bool) string {
	if active {
		return "success"
	}
	return "default"
}

func statusPageTitle(l score_scale_band.Labels, status string) string {
	switch status {
	case "active":
		return l.Page.HeadingActive
	case "inactive":
		return l.Page.HeadingInactive
	default:
		return l.Page.Heading
	}
}

func statusPageCaption(l score_scale_band.Labels, status string) string {
	switch status {
	case "active":
		return l.Page.CaptionActive
	case "inactive":
		return l.Page.CaptionInactive
	default:
		return l.Page.Caption
	}
}

func emptyTitle(l score_scale_band.Labels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveTitle
	case "inactive":
		return l.Empty.InactiveTitle
	default:
		return l.Empty.Title
	}
}

func emptyMessage(l score_scale_band.Labels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveMessage
	case "inactive":
		return l.Empty.InactiveMessage
	default:
		return l.Empty.Message
	}
}
