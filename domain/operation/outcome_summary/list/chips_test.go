package list

// chips_test.go — R9 Phase-B approval-status DISTRIBUTION chips (W-B1/W-B2;
// Q-R9-1, Q-R9-4). Proves: multi-status distribution + ladder order, the
// staff-fold dedup, no-data exclusion (still counted in the total), the mixed
// attention overlay, the education1 all-IN_PROGRESS shape, and the degrade to
// the Phase-A count-only cell.

import (
	"context"
	"reflect"
	"testing"

	outcome_summary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/types"

	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	summarypb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/job_template_summary"
)

// Approval-ladder enum shorthands.
var (
	stIP  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_IN_PROGRESS
	stFR  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW
	stVER = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_VERIFIED
	stPUB = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED
)

// phaseBDeps is dynamicDeps (page_test.go) with the Phase-B approval-chip labels
// set — dynamicDeps only sets the Phase-A landing labels.
func phaseBDeps() *ListViewDeps {
	deps := dynamicDeps()
	deps.Labels.Landing.ApprovalStatus = outcome_summary.ApprovalStatusChipLabels{
		InProgress: "In Progress",
		ForReview:  "For Review",
		Verified:   "Verified",
		Published:  "Published",
		Mixed:      "Attention — mixed",
	}
	return deps
}

// TestLandingStatusChipsDistribution pins a multi-status distribution in ladder
// order AND the staff-fold dedup: t1 (VERIFIED) arrives on TWO staff rows but is
// counted ONCE; the cat-a cell shows in_progress → verified → published (no
// for_review bucket), total count 3.
func TestLandingStatusChipsDistribution(t *testing.T) {
	deps := phaseBDeps()
	deps.ListJobTemplateSummaries = func(context.Context, *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error) {
		return &summarypb.ListJobTemplateSummariesResponse{Summaries: []*summarypb.JobTemplateSummary{
			{JobTemplateId: "t1", SubscriptionGroupId: "g-1", JobCount: 28, JobCategoryId: "cat-a", GroupPhaseCount: 2, GroupLowestStatus: stVER},
			{JobTemplateId: "t1", SubscriptionGroupId: "g-1", JobCount: 28, JobCategoryId: "cat-a", GroupPhaseCount: 2, GroupLowestStatus: stVER}, // second staff row — dedup
			{JobTemplateId: "t2", SubscriptionGroupId: "g-1", JobCount: 28, JobCategoryId: "cat-a", GroupPhaseCount: 2, GroupLowestStatus: stPUB},
			{JobTemplateId: "t3", SubscriptionGroupId: "g-1", JobCount: 28, JobCategoryId: "cat-a", GroupPhaseCount: 1, GroupLowestStatus: stIP},
		}}, nil
	}
	ctx, vc := landingReq("")
	pd := mustPageData(t, NewView(deps).Handle(ctx, vc))

	cat := pd.Table.Rows[0].Cells[2].Composite
	if cat == nil {
		t.Fatal("cat-a Composite nil")
	}
	if cat.Count != 3 {
		t.Fatalf("cat-a count = %d, want 3 (t1 deduped across staff rows)", cat.Count)
	}
	want := []types.CompositeStatusChip{
		{Label: "In Progress", Count: 1, Variant: "default"},
		{Label: "Verified", Count: 1, Variant: "info"},
		{Label: "Published", Count: 1, Variant: "success"},
	}
	if !reflect.DeepEqual(cat.Chips, want) {
		t.Fatalf("cat-a chips = %+v, want %+v (ladder order, staff-fold deduped)", cat.Chips, want)
	}
}

// TestLandingStatusChipsNoDataExcluded: a subject with no data-bearing phase
// (group_phase_count == 0) is EXCLUDED from the chips but STILL counted in the
// subject total (the Phase-A count is unchanged).
func TestLandingStatusChipsNoDataExcluded(t *testing.T) {
	deps := phaseBDeps()
	deps.ListJobTemplateSummaries = func(context.Context, *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error) {
		return &summarypb.ListJobTemplateSummariesResponse{Summaries: []*summarypb.JobTemplateSummary{
			{JobTemplateId: "t1", SubscriptionGroupId: "g-1", JobCount: 28, JobCategoryId: "cat-a", GroupPhaseCount: 2, GroupLowestStatus: stPUB},
			{JobTemplateId: "t2", SubscriptionGroupId: "g-1", JobCount: 28, JobCategoryId: "cat-a", GroupPhaseCount: 0, GroupLowestStatus: stIP}, // no data-bearing phase
		}}, nil
	}
	ctx, vc := landingReq("")
	pd := mustPageData(t, NewView(deps).Handle(ctx, vc))

	cat := pd.Table.Rows[0].Cells[2].Composite
	if cat.Count != 2 {
		t.Fatalf("cat-a total = %d, want 2 (the no-data subject STILL counts in the total)", cat.Count)
	}
	want := []types.CompositeStatusChip{{Label: "Published", Count: 1, Variant: "success"}}
	if !reflect.DeepEqual(cat.Chips, want) {
		t.Fatalf("chips = %+v, want only the data-bearing PUBLISHED subject (no-data excluded)", cat.Chips)
	}
}

// TestLandingStatusChipsMixedMarker: any subject carrying group_mixed_attention
// appends a TRAILING warning attention chip (the R7-P3 "Attention — mixed"
// idiom), Count = subjects internally mixed. The overlay does NOT partition the
// ladder buckets (both subjects still count VERIFIED).
func TestLandingStatusChipsMixedMarker(t *testing.T) {
	deps := phaseBDeps()
	deps.ListJobTemplateSummaries = func(context.Context, *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error) {
		return &summarypb.ListJobTemplateSummariesResponse{Summaries: []*summarypb.JobTemplateSummary{
			{JobTemplateId: "t1", SubscriptionGroupId: "g-1", JobCount: 28, JobCategoryId: "cat-a", GroupPhaseCount: 2, GroupLowestStatus: stVER, GroupMixedAttention: true},
			{JobTemplateId: "t2", SubscriptionGroupId: "g-1", JobCount: 28, JobCategoryId: "cat-a", GroupPhaseCount: 2, GroupLowestStatus: stVER},
		}}, nil
	}
	ctx, vc := landingReq("")
	pd := mustPageData(t, NewView(deps).Handle(ctx, vc))

	chips := pd.Table.Rows[0].Cells[2].Composite.Chips
	want := []types.CompositeStatusChip{
		{Label: "Verified", Count: 2, Variant: "info"},
		{Label: "Attention — mixed", Count: 1, Variant: "warning"},
	}
	if !reflect.DeepEqual(chips, want) {
		t.Fatalf("chips = %+v, want a VERIFIED bucket + a TRAILING warning mixed overlay", chips)
	}
}

// TestLandingStatusChipsAllInProgress pins the education1 LIVE shape: every
// subject at IN_PROGRESS → exactly ONE neutral chip carrying the full count.
func TestLandingStatusChipsAllInProgress(t *testing.T) {
	deps := phaseBDeps()
	deps.ListJobTemplateSummaries = func(context.Context, *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error) {
		return &summarypb.ListJobTemplateSummariesResponse{Summaries: []*summarypb.JobTemplateSummary{
			{JobTemplateId: "t1", SubscriptionGroupId: "g-1", JobCount: 28, JobCategoryId: "cat-a", GroupPhaseCount: 2, GroupLowestStatus: stIP},
			{JobTemplateId: "t2", SubscriptionGroupId: "g-1", JobCount: 28, JobCategoryId: "cat-a", GroupPhaseCount: 2, GroupLowestStatus: stIP},
			{JobTemplateId: "t3", SubscriptionGroupId: "g-1", JobCount: 28, JobCategoryId: "cat-a", GroupPhaseCount: 1, GroupLowestStatus: stIP},
		}}, nil
	}
	ctx, vc := landingReq("")
	pd := mustPageData(t, NewView(deps).Handle(ctx, vc))

	chips := pd.Table.Rows[0].Cells[2].Composite.Chips
	want := []types.CompositeStatusChip{{Label: "In Progress", Count: 3, Variant: "default"}}
	if !reflect.DeepEqual(chips, want) {
		t.Fatalf("chips = %+v, want a single neutral IN_PROGRESS chip with the full count (education1 shape)", chips)
	}
}

// TestLandingStatusChipsDegradeToPhaseA: the default dynamicDeps summaries carry
// NO group-grain status (all group_phase_count 0), so every composite cell is
// count-only (nil chips) — byte-identical to the Phase-A render.
func TestLandingStatusChipsDegradeToPhaseA(t *testing.T) {
	deps := phaseBDeps()
	ctx, vc := landingReq("")
	pd := mustPageData(t, NewView(deps).Handle(ctx, vc))
	for i := 2; i <= 4; i++ {
		if c := pd.Table.Rows[0].Cells[i].Composite; c != nil && c.Chips != nil {
			t.Fatalf("cell[%d] chips = %+v, want nil (no data-bearing status → Phase-A count-only)", i, c.Chips)
		}
	}
}

// TestBuildStatusChipsLadderOrderAndVariants unit-pins the render helper: chips
// emit in ascending ladder rank with the R7 variant tokens, mixed appended last,
// and a nil distribution degrades to nil chips.
func TestBuildStatusChipsLadderOrderAndVariants(t *testing.T) {
	var l outcome_summary.Labels
	l.Landing.ApprovalStatus = outcome_summary.ApprovalStatusChipLabels{
		InProgress: "In Progress", ForReview: "For Review", Verified: "Verified", Published: "Published", Mixed: "Attention — mixed",
	}
	// byStatus deliberately inserted out of order to prove the emit order comes
	// from the ladder, not the map.
	d := &cellStatusDist{
		byStatus: map[jobphasepb.PhaseApprovalStatus]int{stPUB: 4, stIP: 1, stVER: 2, stFR: 3},
		mixed:    2,
	}
	got := buildStatusChips(d, l)
	want := []types.CompositeStatusChip{
		{Label: "In Progress", Count: 1, Variant: "default"},
		{Label: "For Review", Count: 3, Variant: "warning"},
		{Label: "Verified", Count: 2, Variant: "info"},
		{Label: "Published", Count: 4, Variant: "success"},
		{Label: "Attention — mixed", Count: 2, Variant: "warning"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildStatusChips = %+v, want ladder order + trailing mixed", got)
	}
	if buildStatusChips(nil, l) != nil {
		t.Fatal("nil dist must yield nil chips (Phase-A degrade)")
	}
}

// TestRecordSubjectStatusExcludesNoDataAndOverlaysMixed unit-pins the collation:
// a data-bearing subject records its status + mixed overlay; a no-data subject
// (group_phase_count 0) is skipped entirely and never appears in byStatus.
func TestRecordSubjectStatusExcludesNoDataAndOverlaysMixed(t *testing.T) {
	dist := map[string]map[string]*cellStatusDist{}
	recordSubjectStatus(dist, "g", "c", &summarypb.JobTemplateSummary{GroupPhaseCount: 2, GroupLowestStatus: stVER, GroupMixedAttention: true})
	recordSubjectStatus(dist, "g", "c", &summarypb.JobTemplateSummary{GroupPhaseCount: 0, GroupLowestStatus: stPUB}) // no data-bearing phase

	d := dist["g"]["c"]
	if d == nil {
		t.Fatal("expected a distribution for (g,c)")
	}
	if d.byStatus[stVER] != 1 || d.mixed != 1 {
		t.Fatalf("dist = %+v, want verified=1 mixed=1", d)
	}
	if len(d.byStatus) != 1 {
		t.Fatalf("byStatus has %d entries, want 1 (the no-data PUBLISHED subject must not appear)", len(d.byStatus))
	}
}
