package outcome_summary

import (
	"context"
	"log"
	"strings"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
)

// ResolveCategoryID resolves a configured job_category CODE (Options.CategoryFilter,
// e.g. "academic") to its id, using the per-request ListJobCategories closure. It
// is the shared resolver for the three grade surfaces (view-2 section grid, view-3
// client card, report-card document); each surface calls it ONCE per request and
// then filters its job set with KeepJobInCategory.
//
// Returns "" — meaning "apply no category filter" — when: no code is configured,
// the closure is unwired (service-admin), the list read errors, or the code does
// not resolve to a category. Fail-open-to-unfiltered is deliberate: a blank or
// misconfigured taxonomy must never HIDE real grade rows, and a workspace with no
// seeded categories carries only NULL job_category_id jobs (which KeepJobInCategory
// keeps regardless). When the code resolves, its id is returned and out-of-category
// jobs (e.g. deportment) are dropped from the academic surfaces.
func ResolveCategoryID(
	ctx context.Context,
	list func(ctx context.Context, req *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error),
	code string,
) string {
	code = strings.TrimSpace(code)
	if code == "" || list == nil {
		return ""
	}
	resp, err := list(ctx, &jobcategorypb.ListJobCategoriesRequest{
		Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{{
			Field: "code",
			FilterType: &commonpb.TypedFilter_StringFilter{
				StringFilter: &commonpb.StringFilter{Value: code, Operator: commonpb.StringOperator_STRING_EQUALS},
			},
		}}},
	})
	if err != nil {
		log.Printf("outcome summary: resolve job_category code %q: %v (no category filter applied)", code, err)
		return ""
	}
	for _, c := range resp.GetData() {
		if strings.EqualFold(strings.TrimSpace(c.GetCode()), code) {
			return c.GetId()
		}
	}
	log.Printf("outcome summary: job_category code %q did not resolve (no category filter applied)", code)
	return ""
}

// KeepJobInCategory reports whether a job with jobCategoryID belongs to the
// resolved category filter. categoryID == "" means "no filter" (keep every job —
// today's behavior / service-admin). Otherwise keep the job when its category id
// matches OR is empty — the legacy/denorm-not-yet-populated NULL fallback (M7:
// normal materialization paths do not yet copy job_category_id, so a NULL row must
// never be dropped from a grade surface).
func KeepJobInCategory(categoryID, jobCategoryID string) bool {
	if categoryID == "" {
		return true
	}
	return jobCategoryID == "" || jobCategoryID == categoryID
}
