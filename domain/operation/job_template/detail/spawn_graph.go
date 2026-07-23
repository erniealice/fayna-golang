package detail

import (
	"context"
	"fmt"
	"log"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobtemplaterelationpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_relation"
)

// loadSpawnGraphTab populates PageData.SpawnGraphTable with the
// JobTemplateRelation edges where this template is the PARENT — the
// child-list roster tab flavor from ui-detail-tabs.md, filtered by
// ListRelationsByParent (a chunked "by parent" list, not a generic filtered
// List call).
//
// The "+ Add Relation" CTA and per-row remove actions are wired to the
// job_template_relation module's real routes (RelationRoutes), permission-
// gated on job_template_relation:create / job_template_relation:delete —
// fail-closed inside the tab body (an empty/no-CTA table, never
// view.Forbidden — a tab-swap target must stay a partial). Also fail-closed
// on job_template_relation:list for the table itself: an operator without
// list rights sees an empty roster rather than a leaked partial list.
func loadSpawnGraphTab(ctx context.Context, deps *DetailViewDeps, pageData *PageData, templateID string) {
	perms := view.GetUserPermissions(ctx)

	if deps.RelationRoutes.AddURL != "" && perms.Can("job_template_relation", "create") {
		pageData.SpawnGraphAddURL = deps.RelationRoutes.AddURL + "?parent_template_id=" + templateID + "&return_table=jt-spawn-graph-table"
	}

	if !perms.Can("job_template_relation", "list") || deps.ListRelationsByParent == nil {
		// Unpermitted or unwired — render empty state (table stays nil).
		return
	}

	resp, err := deps.ListRelationsByParent(ctx, &jobtemplaterelationpb.ListJobTemplateRelationsByParentRequest{
		ParentTemplateId: templateID,
	})
	if err != nil {
		log.Printf("loadSpawnGraphTab: failed to list relations for template %s: %v", templateID, err)
		return
	}
	relations := resp.GetJobTemplateRelations()

	canDelete := deps.RelationRoutes.DeleteURL != "" && perms.Can("job_template_relation", "delete")
	rows := make([]types.TableRow, 0, len(relations))
	for _, rel := range relations {
		childName := ""
		if ct := rel.GetChildTemplate(); ct != nil {
			childName = ct.GetName()
		}
		if childName == "" {
			childName = rel.GetChildTemplateId()
		}

		var actions []types.TableAction
		if canDelete {
			actions = append(actions, types.TableAction{
				Type:     "delete",
				Action:   "delete",
				Label:    "Remove Relation",
				URL:      deps.RelationRoutes.DeleteURL + "?return_table=jt-spawn-graph-table",
				ItemName: childName,
			})
		}

		rows = append(rows, types.TableRow{
			ID: rel.GetId(),
			Cells: []types.TableCell{
				{Type: "text", Value: fmt.Sprintf("%d", rel.GetSequenceOrder())},
				{Type: "text", Value: childName},
				{Type: "text", Value: rel.GetRelationType().String()},
			},
			Actions: actions,
		})
	}

	pageData.SpawnGraphTable = &types.TableConfig{
		ID: "jt-spawn-graph-table",
		Columns: []types.TableColumn{
			{Key: "seq", Label: "Seq", WidthClass: "col-sm"},
			{Key: "child", Label: "Child Template"},
			{Key: "relation_type", Label: "Relation Type"},
		},
		Rows:        rows,
		Labels:      deps.TableLabels,
		ShowSearch:  false,
		ShowActions: canDelete && len(rows) > 0,
		ShowSort:    false,
		ShowColumns: false,
		ShowDensity: false,
		ShowEntries: false,
		RefreshURL:  route.ResolveURL(deps.Routes.TabActionURL, "id", templateID, "tab", "spawn-graph"),
	}
	types.ApplyColumnStyles(pageData.SpawnGraphTable.Columns, pageData.SpawnGraphTable.Rows)
}
