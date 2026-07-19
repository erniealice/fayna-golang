package list

// chips.go — the landing's Phase-B approval-status DISTRIBUTION chips (R9
// W-B1/W-B2; Q-R9-1, Q-R9-4). Each (section × category) composite cell gains
// one pyeza status-badge chip per NONZERO approval state, ladder-ordered, plus
// an attention (mixed) overlay chip. The chips ride the SAME single
// ListJobTemplateSummaries read as the Phase-A counts (statement budget +0,
// plan §3.6): the summary rows carry R7 P4's AMENDED group+template preaggregate
// (proto fields 18–21 — group_published_count / group_phase_count /
// group_lowest_status / group_mixed_attention), so NO second collapse is rolled
// here (the parallel-collapse escape hatch is DELETED — plan §3.5 / progress
// wave 6).
//
// GENERIC naming (code-generic-lyngua-vertical): no education nouns. The status
// VOCABULARY ("In Progress" …) is lyngua DATA (Labels.Landing.ApprovalStatus,
// reusing R7's approval keys); the badge VARIANT is a Go switch (never lyngua),
// identical to R7's outcome_matrix approvalChipVariant so the SAME status shows
// the SAME theme-aware token on every surface.

import (
	outcome_summary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/types"

	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	summarypb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/job_template_summary"
)

// ladder is the approval status ORDER the chips render in (Q-R9-4): the enum's
// ascending rank (in_progress → for_review → verified → published).
var ladder = []jobphasepb.PhaseApprovalStatus{
	jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_IN_PROGRESS,
	jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW,
	jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_VERIFIED,
	jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED,
}

// cellStatusDist is the Phase-B approval-status distribution for ONE
// (section, category) cell: how many SUBJECTS (distinct templates, staff-folded)
// sit at each conservative group_lowest_status, plus how many carry the
// group_mixed_attention overlay. Subjects with NO data-bearing phase
// (group_phase_count == 0) are EXCLUDED here — they still count in the Phase-A
// subject TOTAL (the bare count) so an all-no-data cell degrades to exactly the
// Phase-A render (Q-R9-1, acceptance #13). The mixed count OVERLAYS the ladder
// buckets (a subject can be both VERIFIED and mixed), mirroring R7's grid bar
// where the status badge and the mixed badge coexist — so the chip counts are
// deliberately NOT a partition of the total.
type cellStatusDist struct {
	byStatus map[jobphasepb.PhaseApprovalStatus]int
	mixed    int
}

// recordSubjectStatus folds ONE distinct (group, template) subject's group-grain
// approval state into the (gid, cat) cell distribution. It is called EXACTLY
// ONCE per distinct subject — from inside sectionCounts' existing seen[gid][tid]
// dedup — so staff-folded summary rows (one row per deliverer, all carrying the
// SAME group-grain fields 18–21) never double-count: THIS is the staff-fold
// dedupe proof. A subject with zero data-bearing phases (group_phase_count == 0,
// or a defensive UNSPECIFIED status) is skipped: no chip, but the caller has
// already counted it in the subject total.
func recordSubjectStatus(dist map[string]map[string]*cellStatusDist, gid, cat string, s *summarypb.JobTemplateSummary) {
	if s.GetGroupPhaseCount() <= 0 {
		return
	}
	st := s.GetGroupLowestStatus()
	if st == jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_UNSPECIFIED {
		return
	}
	if dist[gid] == nil {
		dist[gid] = map[string]*cellStatusDist{}
	}
	d := dist[gid][cat]
	if d == nil {
		d = &cellStatusDist{byStatus: map[jobphasepb.PhaseApprovalStatus]int{}}
		dist[gid][cat] = d
	}
	d.byStatus[st]++
	if s.GetGroupMixedAttention() {
		d.mixed++
	}
}

// buildStatusChips renders a cell's distribution into ladder-ordered pyeza
// status chips (Q-R9-4): one chip per NONZERO status bucket (in_progress →
// for_review → verified → published), each carrying the subject Count, the
// reused R7 status Label, and the theme-aware Variant. When any subject carries
// the mixed overlay, a trailing WARNING attention chip is appended — the R7-P3
// outcome_matrix "Attention — mixed" idiom, adapted to the count-bearing
// composite chip (Count = subjects internally mixed). A nil/empty distribution
// (all subjects no-data, or a historical section the active-bound summary
// aggregate never saw) yields NO chips → the Phase-A count-only cell (the
// degrade contract, acceptance #13).
func buildStatusChips(d *cellStatusDist, l outcome_summary.Labels) []types.CompositeStatusChip {
	if d == nil {
		return nil
	}
	var chips []types.CompositeStatusChip
	for _, st := range ladder {
		if n := d.byStatus[st]; n > 0 {
			chips = append(chips, types.CompositeStatusChip{
				Label:   approvalStatusLabel(l, st),
				Count:   n,
				Variant: approvalChipVariant(st),
			})
		}
	}
	if d.mixed > 0 {
		chips = append(chips, types.CompositeStatusChip{
			Label:   l.Landing.ApprovalStatus.Mixed,
			Count:   d.mixed,
			Variant: "warning",
		})
	}
	return chips
}

// approvalStatusLabel maps the ladder enum to its lyngua status text — the SAME
// vocabulary R7's outcome_matrix uses (reused per lyngua.md; the strings live in
// the outcome_summary namespace because this package loads its own labels).
func approvalStatusLabel(l outcome_summary.Labels, s jobphasepb.PhaseApprovalStatus) string {
	switch s {
	case jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW:
		return l.Landing.ApprovalStatus.ForReview
	case jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_VERIFIED:
		return l.Landing.ApprovalStatus.Verified
	case jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED:
		return l.Landing.ApprovalStatus.Published
	default:
		return l.Landing.ApprovalStatus.InProgress
	}
}

// approvalChipVariant is the Go badge-variant switch (NOT lyngua) → the pyeza
// status-badge modifier. Semantically mirrors outcome_matrix.approvalChipVariant
// (same status → same COLOR meaning) but the two surfaces render DIFFERENT CSS
// families: the grid bar uses `badge badge-<v>` (which has `.badge-neutral`)
// while the composite cell renders `status-badge <v>` — whose muted token is
// named `default` (badge.css has no `.status-badge.neutral`). Hence
// in_progress → "default" here vs "neutral" there; for_review → warning,
// verified → info, published → success (Q-R9-4). A pyeza `.status-badge.neutral`
// alias would let both surfaces share the literal token (follow-up).
func approvalChipVariant(s jobphasepb.PhaseApprovalStatus) string {
	switch s {
	case jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW:
		return "warning"
	case jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_VERIFIED:
		return "info"
	case jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED:
		return "success"
	default:
		return "default"
	}
}
