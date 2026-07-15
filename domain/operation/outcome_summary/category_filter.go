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
// Returns (categoryID, ok):
//   - ok=true,  categoryID==""  — NO category filter configured (blank code, or an
//     unwired closure: service-admin). KeepJobInCategory then keeps every job. This
//     is the deliberate "no filter" mode, NOT a failure.
//   - ok=true,  categoryID!=""  — the code resolved to a category; out-of-category
//     jobs (e.g. deportment) are dropped from the academic surfaces.
//   - ok=FALSE, categoryID==""  — a filter WAS configured (non-blank code) but could
//     NOT be resolved: the list read ERRORED, or the code matched no category. The
//     caller MUST fail CLOSED and keep NO jobs. Fail-open-to-unfiltered here would
//     leak same-origin deportment jobs into the academic grade surfaces (incl. the
//     shared CSV/DOCX renderers) on any transient lookup failure (gate H2, fail-open
//     hardening — B4 codex finding #2). education1 has 0 NULL-category jobs, so this
//     is defense for a resolution ERROR, not a data change.
func ResolveCategoryID(
	ctx context.Context,
	list func(ctx context.Context, req *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error),
	code string,
) (string, bool) {
	code = strings.TrimSpace(code)
	if code == "" || list == nil {
		// No filter configured — keep every job (service-admin / unfiltered tier).
		return "", true
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
		// A configured filter whose lookup errored: fail CLOSED (keep no jobs)
		// rather than regressing a filtered surface to unfiltered data.
		log.Printf("outcome summary: resolve job_category code %q: %v (FAIL CLOSED — no jobs)", code, err)
		return "", false
	}
	for _, c := range resp.GetData() {
		if strings.EqualFold(strings.TrimSpace(c.GetCode()), code) {
			return c.GetId(), true
		}
	}
	// Configured code did not resolve to any category: fail CLOSED.
	log.Printf("outcome summary: job_category code %q did not resolve (FAIL CLOSED — no jobs)", code)
	return "", false
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
