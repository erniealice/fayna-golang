// Package form — options.go holds pure-function option builders for the
// job_template_relation drawer form. Each builder takes a narrow list-
// closure signature (not the full ModuleDeps).
package form

import (
	"context"
	"sort"

	"github.com/erniealice/pyeza-golang/types"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	relationpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_relation"
)

// BuildTemplateOptions calls the narrow ListJobTemplates closure and returns
// select options for the Parent/Child Template pickers. Limit is set to the
// adapter's documented pagination cap (100) — a workspace with more active
// templates than that will silently truncate this picker (see
// ui-detail-tabs.md "page through the child list"); acceptable for a v1
// picker, flagged in the W4 report. A nil closure or a failed call yields an
// empty slice.
func BuildTemplateOptions(ctx context.Context, listFn func(context.Context, *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error), selected string) []types.SelectOption {
	if listFn == nil {
		return nil
	}
	resp, err := listFn(ctx, &jobtemplatepb.ListJobTemplatesRequest{
		Pagination: &commonpb.PaginationRequest{Limit: 100},
	})
	if err != nil || resp == nil {
		return nil
	}
	items := resp.GetData()
	opts := make([]types.SelectOption, 0, len(items))
	for _, t := range items {
		opts = append(opts, types.SelectOption{
			Value:    t.GetId(),
			Label:    t.GetName(),
			Selected: t.GetId() == selected,
		})
	}
	sort.Slice(opts, func(i, j int) bool { return opts[i].Label < opts[j].Label })
	return opts
}

// BuildRelationTypeOptions returns select options for the relation_type
// field. Values match the proto enum string names (JOB_TEMPLATE_RELATION_
// TYPE_*); UNSPECIFIED is skipped (see job_template_relation.proto).
func BuildRelationTypeOptions(current string) []types.SelectOption {
	rows := []struct {
		value string
		enum  relationpb.JobTemplateRelationType
		label string
	}{
		{"JOB_TEMPLATE_RELATION_TYPE_SUB_TEMPLATE", relationpb.JobTemplateRelationType_JOB_TEMPLATE_RELATION_TYPE_SUB_TEMPLATE, "Sub-Template"},
		{"JOB_TEMPLATE_RELATION_TYPE_ONCE_AT_ENGAGEMENT_START", relationpb.JobTemplateRelationType_JOB_TEMPLATE_RELATION_TYPE_ONCE_AT_ENGAGEMENT_START, "Once at Engagement Start"},
	}
	opts := make([]types.SelectOption, 0, len(rows))
	for _, r := range rows {
		_ = r.enum // import-guard: ensures enum const is referenced
		opts = append(opts, types.SelectOption{
			Value:    r.value,
			Label:    r.label,
			Selected: r.value == current,
		})
	}
	return opts
}
