// Package form — options.go holds pure-function option builders for the
// job template drawer form. Each builder takes a narrow list-closure
// signature (not the full action.Deps) and returns select options.
package form

import (
	"context"
	"sort"

	"github.com/erniealice/pyeza-golang/types"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	productpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/product/product"
)

// BuildInitialStatusOptions returns select options for the initial_status
// field — the lifecycle status a spawned Job takes when this template
// materializes (job_template.proto field 50; carries a JobStatus enum NAME
// string, e.g. "JOB_STATUS_ACTIVE"). A NULL/empty value falls back to
// JOB_STATUS_PLANNED at spawn time (see espyna's resolveInitialJobStatus) —
// this builder does NOT pre-select a default; the placeholder option is left
// selected until the operator picks one, matching that NULL fallback.
func BuildInitialStatusOptions(current string) []types.SelectOption {
	rows := []struct {
		value string
		enum  enums.JobStatus
		label string
	}{
		{"JOB_STATUS_DRAFT", enums.JobStatus_JOB_STATUS_DRAFT, "Draft"},
		{"JOB_STATUS_PENDING", enums.JobStatus_JOB_STATUS_PENDING, "Pending"},
		{"JOB_STATUS_PLANNED", enums.JobStatus_JOB_STATUS_PLANNED, "Planned"},
		{"JOB_STATUS_RELEASED", enums.JobStatus_JOB_STATUS_RELEASED, "Released"},
		{"JOB_STATUS_ACTIVE", enums.JobStatus_JOB_STATUS_ACTIVE, "Active"},
		{"JOB_STATUS_PAUSED", enums.JobStatus_JOB_STATUS_PAUSED, "Paused"},
		{"JOB_STATUS_COMPLETED", enums.JobStatus_JOB_STATUS_COMPLETED, "Completed"},
		{"JOB_STATUS_CLOSED", enums.JobStatus_JOB_STATUS_CLOSED, "Closed"},
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

// BuildVersionStatusOptions returns select options for the version_status field
// — the template's publication state (VersionStatus enum NAME string). Only
// Draft and Published are offered (UNSPECIFIED/DEPRECATED are not operator
// choices). The create default is Draft: an empty/unspecified current pre-selects
// Draft so a new template is never left VERSION_STATUS_UNSPECIFIED. No
// approval-ladder mechanics — Published is a plain operator choice.
func BuildVersionStatusOptions(current string) []types.SelectOption {
	if current == "" || current == "VERSION_STATUS_UNSPECIFIED" {
		current = "VERSION_STATUS_DRAFT"
	}
	rows := []struct {
		value string
		enum  enums.VersionStatus
		label string
	}{
		{"VERSION_STATUS_DRAFT", enums.VersionStatus_VERSION_STATUS_DRAFT, "Draft"},
		{"VERSION_STATUS_PUBLISHED", enums.VersionStatus_VERSION_STATUS_PUBLISHED, "Published"},
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

// BuildCategoryOptions calls the narrow ListJobCategories closure and returns
// select options for the Category picker. A nil closure or a failed call
// yields an empty slice — the field stays optional and the drawer still
// renders (fail-soft, not fail-closed: this is a picker, not a permission
// gate).
func BuildCategoryOptions(ctx context.Context, listFn func(context.Context, *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error), selected string) []types.SelectOption {
	if listFn == nil {
		return nil
	}
	resp, err := listFn(ctx, &jobcategorypb.ListJobCategoriesRequest{})
	if err != nil || resp == nil {
		return nil
	}
	items := resp.GetData()
	opts := make([]types.SelectOption, 0, len(items))
	for _, c := range items {
		opts = append(opts, types.SelectOption{
			Value:    c.GetId(),
			Label:    c.GetName(),
			Selected: c.GetId() == selected,
		})
	}
	sort.Slice(opts, func(i, j int) bool { return opts[i].Label < opts[j].Label })
	return opts
}

// BuildOutputProductOptions calls the narrow ListProducts closure and returns
// select options for the Output Product picker. No product-search endpoint
// is reachable from fayna's wiring today (see action.Deps.ListProducts doc),
// so this backs a plain select rather than an action-mode auto-complete.
// A nil closure or a failed call yields an empty slice.
func BuildOutputProductOptions(ctx context.Context, listFn func(context.Context, *productpb.ListProductsRequest) (*productpb.ListProductsResponse, error), selected string) []types.SelectOption {
	if listFn == nil {
		return nil
	}
	resp, err := listFn(ctx, &productpb.ListProductsRequest{})
	if err != nil || resp == nil {
		return nil
	}
	items := resp.GetData()
	opts := make([]types.SelectOption, 0, len(items))
	for _, p := range items {
		opts = append(opts, types.SelectOption{
			Value:    p.GetId(),
			Label:    p.GetName(),
			Selected: p.GetId() == selected,
		})
	}
	sort.Slice(opts, func(i, j int) bool { return opts[i].Label < opts[j].Label })
	return opts
}
