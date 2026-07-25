package outcome_matrix

import (
	"context"
	"errors"
	"net/http"
	"strings"

	summarypb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/job_template_summary"
)

// ErrGroupNotInTemplate is returned to the client as a 404 when {group_id}
// does not pair with {id}. Deliberately terse — it must not disclose whether the
// section exists, belongs to another workspace, or merely delivers a different
// template.
var ErrGroupNotInTemplate = errors.New("not found")

// section_scope.go — the (template, section) pair guard for the Group* routes.
//
// WHY A GUARD AND NOT JUST A FILTER. The adapter narrows a section by
//
//	j.client_id IN (SELECT client_id FROM subscription_group_member
//	                WHERE subscription_group_id = $n AND ...)
//
// with NO academic-year term, and a student holds active membership in a section
// in EVERY academic year they were enrolled. So passing a section id that belongs
// to a DIFFERENT year than the template does not fail and does not return zero
// rows — it returns the intersection, a plausible non-empty PARTIAL roster.
// Measured on education1 (2026-07-25) against "Arts — AY 2025-2026", whose true
// roster is Grade 8 Mercury / 30 students: six foreign-AY section ids returned
// 14, 9, 8, 7, 7 and 4 students respectively. Each would render as a normal,
// believable grade sheet.
//
// Applying the filter alone would therefore turn a mistyped or guessed id into
// silent cross-cohort data exposure inside the workspace. The pair must be
// VALIDATED, and it must fail CLOSED.

// GroupScope carries the validated section narrowing for one request.
// GroupID is empty on the template-scoped routes, which keeps every existing
// call site behaving exactly as before.
type GroupScope struct {
	GroupID   string
	GroupName string
}

// Scoped reports whether a section narrowing is in effect.
func (s GroupScope) Scoped() bool { return s.GroupID != "" }

// SummaryLister is the narrow read the guard needs. It matches the
// ListJobTemplateSummaries closure threaded through the module deps.
type SummaryLister func(ctx context.Context, req *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error)

// ResolveGroupScope reads {group_id} off the request path and validates that
// the section actually delivers the given template.
//
// Returns (zero, true) when the route carries no {group_id} — the template-scoped
// case, always allowed. Returns (zero, false) when a {group_id} IS present but
// does not pair with the template, or cannot be checked; callers MUST 404 on
// false rather than falling through to an unnarrowed render.
//
// Fail-closed cases, all deliberate:
//   - lister == nil        — no way to verify, so refuse rather than trust
//   - lister returns error — same
//   - no matching row      — the pair does not exist
func ResolveGroupScope(ctx context.Context, r *http.Request, templateID string, lister SummaryLister) (GroupScope, bool) {
	if r == nil {
		return GroupScope{}, true
	}
	groupID := strings.TrimSpace(r.PathValue("group_id"))
	if groupID == "" {
		return GroupScope{}, true
	}
	if templateID == "" || lister == nil {
		return GroupScope{}, false
	}

	// Narrow the aggregate to this section and look for the template among its
	// deliveries. The aggregate is already at (template x group) grain and is
	// workspace-bound by the use case, so a section from another workspace
	// simply yields no rows — the same false this returns for a wrong pair.
	resp, err := lister(ctx, &summarypb.ListJobTemplateSummariesRequest{
		SubscriptionGroupId: &groupID,
	})
	if err != nil || resp == nil {
		return GroupScope{}, false
	}
	for _, s := range resp.GetSummaries() {
		if s.GetJobTemplateId() == templateID {
			return GroupScope{GroupID: groupID, GroupName: s.GetSubscriptionGroupName()}, true
		}
	}
	return GroupScope{}, false
}
