package list

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	documentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/template"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	cardbindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary_document_template"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// captureOutcomeMatrixLogs runs fn with the package-level logger redirected
// to a buffer, so a test can assert on the single once-per-request line
// resolvePhaseDocumentScheduleID's caller emits when a phase-label input is
// missing (replacing the previous silent early return).
func captureOutcomeMatrixLogs(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	oldOut := log.Writer()
	oldFlags := log.Flags()
	oldPrefix := log.Prefix()
	log.SetOutput(&buf)
	log.SetFlags(0)
	log.SetPrefix("")
	defer func() {
		log.SetOutput(oldOut)
		log.SetFlags(oldFlags)
		log.SetPrefix(oldPrefix)
	}()
	fn()
	return buf.String()
}

// deniedSubscriptionGroupMembers fails the test if invoked — a section-scoped
// page must resolve its price_schedule_id straight off GroupScope (populated
// by ResolveGroupScope from the SAME ListJobTemplateSummaries row a STAFF
// principal can already read) and must never fall through to
// deliverygroup.ResolveOneDetail's subscription_group_member:list call, which
// AUTHZ-denies a STAFF (kind 7) principal in production.
func deniedSubscriptionGroupMembers(t *testing.T) func(context.Context, *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error) {
	return func(context.Context, *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error) {
		t.Fatal("ListSubscriptionGroupMembers must not be called for a section-scoped page — STAFF is AUTHZ-denied on subscription_group_member:list")
		return nil, errors.New("permission denied")
	}
}

func deniedSubscriptionGroups(t *testing.T) func(context.Context, *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error) {
	return func(context.Context, *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error) {
		t.Fatal("ListSubscriptionGroups must not be called for a section-scoped page — STAFF is AUTHZ-denied on subscription_group:list")
		return nil, errors.New("permission denied")
	}
}

func noopResolveBinding(context.Context, *cardbindingpb.FindApplicableJobOutcomeSummaryDocumentTemplateRequest) (*cardbindingpb.FindApplicableJobOutcomeSummaryDocumentTemplateResponse, error) {
	return &cardbindingpb.FindApplicableJobOutcomeSummaryDocumentTemplateResponse{Success: true}, nil
}

// TestResolvePhaseDocumentScheduleID_SectionScoped pins the fix: a
// section-scoped page reads price_schedule_id off GroupScope (the validated
// ListJobTemplateSummaries row) and never touches
// ListSubscriptionGroupMembers/ListSubscriptionGroups.
func TestResolvePhaseDocumentScheduleID_SectionScoped(t *testing.T) {
	deps := &PageViewDeps{
		FindApplicableReportCardBinding: noopResolveBinding,
		ListSubscriptionGroupMembers:    deniedSubscriptionGroupMembers(t),
		ListSubscriptionGroups:          deniedSubscriptionGroups(t),
	}
	section := outcome_matrix.GroupScope{GroupID: "group-1", GroupName: "Grade 10 Tantalum", PriceScheduleID: "sched-ay2627"}

	scheduleID, reason := resolvePhaseDocumentScheduleID(context.Background(), deps, section, "tmpl-1")
	if reason != "" {
		t.Fatalf("reason = %q, want empty", reason)
	}
	if scheduleID != "sched-ay2627" {
		t.Fatalf("scheduleID = %q, want %q", scheduleID, "sched-ay2627")
	}
}

// TestResolvePhaseDocumentScheduleID_Unscoped pins the unchanged behavior of
// the template-scoped (all-sections) page: it still samples a job's origin
// subscription and resolves the schedule through deliverygroup.ResolveOneDetail.
func TestResolvePhaseDocumentScheduleID_Unscoped(t *testing.T) {
	origin := "sub-origin-1"
	memberCalls, groupCalls := 0, 0
	deps := &PageViewDeps{
		FindApplicableReportCardBinding: noopResolveBinding,
		ListJobs: func(context.Context, *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error) {
			return &jobpb.ListJobsResponse{Data: []*jobpb.Job{
				{OriginType: enums.OriginType_ORIGIN_TYPE_SUBSCRIPTION, OriginId: &origin},
			}}, nil
		},
		ListSubscriptionGroupMembers: func(_ context.Context, req *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error) {
			memberCalls++
			return &subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse{Data: []*subscriptiongroupmemberpb.SubscriptionGroupMember{
				{SubscriptionId: origin, SubscriptionGroupId: "group-1"},
			}}, nil
		},
		ListSubscriptionGroups: func(context.Context, *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error) {
			groupCalls++
			sched := "sched-ay2627"
			return &subscriptiongrouppb.ListSubscriptionGroupsResponse{Data: []*subscriptiongrouppb.SubscriptionGroup{
				{Id: "group-1", PriceScheduleId: &sched},
			}}, nil
		},
	}

	scheduleID, reason := resolvePhaseDocumentScheduleID(context.Background(), deps, outcome_matrix.GroupScope{}, "tmpl-1")
	if reason != "" {
		t.Fatalf("reason = %q, want empty", reason)
	}
	if scheduleID != "sched-ay2627" {
		t.Fatalf("scheduleID = %q, want %q", scheduleID, "sched-ay2627")
	}
	if memberCalls == 0 || groupCalls == 0 {
		t.Fatalf("unscoped path must still sample via ListSubscriptionGroupMembers/ListSubscriptionGroups: memberCalls=%d groupCalls=%d", memberCalls, groupCalls)
	}
}

// TestResolvePhaseDocumentScheduleID_MissingInputsLog pins the three
// distinct once-per-request log reasons that replaced the previous silent
// early returns.
func TestResolvePhaseDocumentScheduleID_MissingInputsLog(t *testing.T) {
	cases := []struct {
		name       string
		deps       *PageViewDeps
		section    outcome_matrix.GroupScope
		wantReason string
	}{
		{
			name:       "resolver nil",
			deps:       &PageViewDeps{},
			section:    outcome_matrix.GroupScope{GroupID: "group-1", PriceScheduleID: "sched-ay2627"},
			wantReason: "resolver nil",
		},
		{
			name:       "section scoped but schedule unresolved",
			deps:       &PageViewDeps{FindApplicableReportCardBinding: noopResolveBinding},
			section:    outcome_matrix.GroupScope{GroupID: "group-1"},
			wantReason: "schedule unresolved",
		},
		{
			name: "unscoped with no origin job",
			deps: &PageViewDeps{
				FindApplicableReportCardBinding: noopResolveBinding,
				ListJobs: func(context.Context, *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error) {
					return &jobpb.ListJobsResponse{}, nil
				},
			},
			section:    outcome_matrix.GroupScope{},
			wantReason: "no origin",
		},
		{
			name: "unscoped origin found but group unresolved",
			deps: &PageViewDeps{
				FindApplicableReportCardBinding: noopResolveBinding,
				ListJobs: func(context.Context, *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error) {
					subID := "sub-1"
					return &jobpb.ListJobsResponse{Data: []*jobpb.Job{
						{OriginType: enums.OriginType_ORIGIN_TYPE_SUBSCRIPTION, OriginId: &subID},
					}}, nil
				},
				ListSubscriptionGroupMembers: func(context.Context, *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error) {
					return &subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse{}, nil
				},
				ListSubscriptionGroups: func(context.Context, *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error) {
					return &subscriptiongrouppb.ListSubscriptionGroupsResponse{}, nil
				},
			},
			section:    outcome_matrix.GroupScope{},
			wantReason: "schedule unresolved",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scheduleID, reason := resolvePhaseDocumentScheduleID(context.Background(), tc.deps, tc.section, "tmpl-1")
			if scheduleID != "" {
				t.Fatalf("scheduleID = %q, want empty", scheduleID)
			}
			if reason != tc.wantReason {
				t.Fatalf("reason = %q, want %q", reason, tc.wantReason)
			}

			// Pin the caller's actual log line (the once-per-request
			// replacement for the old silent early return).
			logged := captureOutcomeMatrixLogs(t, func() {
				if _, r := resolvePhaseDocumentScheduleID(context.Background(), tc.deps, tc.section, "tmpl-1"); r != "" {
					log.Printf("outcome matrix: phase document labels skipped for template %s (section scoped=%v): %s", "tmpl-1", tc.section.Scoped(), r)
				}
			})
			if !strings.Contains(logged, tc.wantReason) {
				t.Errorf("log output = %q, want it to contain %q", logged, tc.wantReason)
			}
			if !strings.Contains(logged, "tmpl-1") {
				t.Errorf("log output = %q, want it to name the template id", logged)
			}
		})
	}
}

func TestComposePhaseDocumentLabels(t *testing.T) {
	phases := []*matrixpb.PhaseColumn{
		{Code: "q1", Label: "Term 1 (Music)", PhaseName: "Term 1", VariantLabel: "Music"},
		{Code: "q1", Label: "Term 1 (Arts)", PhaseName: "Term 1", VariantLabel: "Arts"},
		{Code: "q2", Label: "Term 2", PhaseName: "Term 2"},
		{Code: "q3", Label: "Term 3", PhaseName: "Term 3"},
		{Code: "q4", Label: "Term 4", PhaseName: "Term 4"},
		{Code: "q4", Label: "Term 4 (Arts)", PhaseName: "Term 4", VariantLabel: "Arts"},
	}
	calls := map[string]int{}
	resolve := func(_ context.Context, req *cardbindingpb.FindApplicableJobOutcomeSummaryDocumentTemplateRequest) (*cardbindingpb.FindApplicableJobOutcomeSummaryDocumentTemplateResponse, error) {
		if req.GetPriceScheduleId() != "ay-1" {
			t.Fatalf("schedule = %q", req.GetPriceScheduleId())
		}
		code := req.GetJobTemplatePhaseCode()
		calls[code]++
		switch code {
		case "q1":
			return &cardbindingpb.FindApplicableJobOutcomeSummaryDocumentTemplateResponse{Success: true, Found: true, Binding: &cardbindingpb.JobOutcomeSummaryDocumentTemplate{JobTemplatePhaseCode: &code, DocumentTemplate: &documentpb.DocumentTemplate{Name: "Progress Report"}}}, nil
		case "q2":
			return &cardbindingpb.FindApplicableJobOutcomeSummaryDocumentTemplateResponse{Success: true}, nil
		case "q4":
			return &cardbindingpb.FindApplicableJobOutcomeSummaryDocumentTemplateResponse{Success: true, Found: true, Binding: &cardbindingpb.JobOutcomeSummaryDocumentTemplate{JobTemplatePhaseCode: &code, DocumentTemplate: &documentpb.DocumentTemplate{Name: "Final Report"}}}, nil
		default:
			return nil, errors.New("lookup failed")
		}
	}
	composePhaseDocumentLabels(context.Background(), resolve, "ay-1", phases)
	want := []string{"Term 1 - Progress Report (Music)", "Term 1 - Progress Report (Arts)", "Term 2", "Term 3", "Term 4 - Final Report", "Term 4 - Final Report (Arts)"}
	for i, phase := range phases {
		if phase.GetLabel() != want[i] {
			t.Errorf("phase %d label = %q, want %q", i, phase.GetLabel(), want[i])
		}
	}
	if len(calls) != 4 || calls["q1"] != 1 || calls["q2"] != 1 || calls["q3"] != 1 || calls["q4"] != 1 {
		t.Errorf("resolver calls = %#v", calls)
	}
}

func ip(s jobphasepb.PhaseApprovalStatus) jobphasepb.PhaseApprovalStatus { return s }

func rollup(phaseID string, status jobphasepb.PhaseApprovalStatus, mixed, hasData, frozen bool, target, blank int32) *matrixpb.PhaseApprovalRollup {
	return &matrixpb.PhaseApprovalRollup{
		JobTemplatePhaseId: phaseID,
		Status:             status,
		Mixed:              mixed,
		HasData:            hasData,
		HardFrozen:         frozen,
		TargetCount:        target,
		BlankRequiredCount: blank,
	}
}

func rollupWithReason(phaseID string, status jobphasepb.PhaseApprovalStatus, mixed, hasData, frozen bool, target, blank int32, reason string) *matrixpb.PhaseApprovalRollup {
	r := rollup(phaseID, status, mixed, hasData, frozen, target, blank)
	r.LastReturnReason = reason
	return r
}

func phaseCol(phaseID, label, taskID string) *matrixpb.PhaseColumn {
	return &matrixpb.PhaseColumn{
		JobTemplatePhaseId: phaseID,
		Label:              label,
		Tasks:              []*matrixpb.TaskColumn{{JobTemplateTaskId: taskID}},
	}
}

const (
	sIP  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_IN_PROGRESS
	sFR  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW
	sVER = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_VERIFIED
	sPUB = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED
)

// TestBuildApprovalBar pins the state-gated action affordances + chip variants +
// the D6 blank-count substitution over the full-sheet roll-up.
func TestBuildApprovalBar(t *testing.T) {
	deps := &PageViewDeps{
		Labels: outcome_matrix.DefaultLabels(),
		Routes: outcome_matrix.DefaultRoutes(),
	}
	perms := types.NewUserPermissions([]string{
		"job_phase:submit", "job_phase:verify", "job_phase:publish", "job_phase:return",
	})
	resp := &matrixpb.GetOutcomeMatrixResponse{
		Phases: []*matrixpb.PhaseColumn{
			phaseCol("pA", "Sem 1", "tA"),
			phaseCol("pB", "Sem 2", "tB"),
			phaseCol("pC", "Sem 3", "tC"),
			phaseCol("pD", "Sem 4", "tD"),
			phaseCol("pE", "Sem 5", "tE"),
			phaseCol("pF", "Sem 6", "tF"),
		},
		ApprovalRollups: []*matrixpb.PhaseApprovalRollup{
			rollup("pA", ip(sIP), false, true, false, 30, 3),  // in_progress → submit only
			rollup("pB", ip(sFR), false, true, false, 30, 0),  // for_review → verify + return
			rollup("pC", ip(sVER), false, true, false, 30, 0), // verified → publish + return
			rollup("pD", ip(sPUB), false, true, false, 30, 0), // published → return (reason required)
			rollup("pE", ip(sIP), true, true, false, 30, 5),   // mixed → return only
			rollup("pF", ip(sVER), false, true, true, 30, 0),  // hard-frozen → no actions
		},
	}

	bar := buildApprovalBar(deps, perms, resp, "tmpl-1", "")
	if len(bar) != 6 {
		t.Fatalf("expected 6 bar phases, got %d", len(bar))
	}
	by := map[string]ApprovalPhase{}
	for _, p := range bar {
		by[p.PhaseID] = p
	}

	// pA: in_progress
	a := by["pA"]
	if a.ChipVariant != "neutral" {
		t.Errorf("pA chip variant = %q, want neutral", a.ChipVariant)
	}
	if !a.CanSubmit || a.CanVerify || a.CanPublish || a.CanReturn {
		t.Errorf("pA actions submit/verify/publish/return = %v/%v/%v/%v, want true/false/false/false", a.CanSubmit, a.CanVerify, a.CanPublish, a.CanReturn)
	}
	if !strings.Contains(a.SubmitConfirm, "3") {
		t.Errorf("pA submit confirm should carry blank count 3: %q", a.SubmitConfirm)
	}
	if a.SubmitPath != "/action/outcome-matrix/tmpl-1/submit" {
		t.Errorf("pA submit path = %q", a.SubmitPath)
	}

	// pB: for_review
	b := by["pB"]
	if b.ChipVariant != "warning" || !b.CanVerify || b.CanSubmit || b.CanPublish || !b.CanReturn {
		t.Errorf("pB unexpected: variant=%q verify=%v submit=%v publish=%v return=%v", b.ChipVariant, b.CanVerify, b.CanSubmit, b.CanPublish, b.CanReturn)
	}

	// pC: verified
	c := by["pC"]
	if c.ChipVariant != "info" || !c.CanPublish || c.CanVerify || !c.CanReturn {
		t.Errorf("pC unexpected: variant=%q publish=%v verify=%v return=%v", c.ChipVariant, c.CanPublish, c.CanVerify, c.CanReturn)
	}

	// pD: published — return only (reason retired 2026-08-01, never required)
	d := by["pD"]
	if d.ChipVariant != "success" || !d.CanReturn || d.CanPublish {
		t.Errorf("pD unexpected: variant=%q return=%v publish=%v", d.ChipVariant, d.CanReturn, d.CanPublish)
	}

	// pE: mixed (lowest in_progress) — submit blocked by mixed, return available
	e := by["pE"]
	if !e.Mixed || e.CanSubmit || !e.CanReturn {
		t.Errorf("pE unexpected: mixed=%v submit=%v return=%v", e.Mixed, e.CanSubmit, e.CanReturn)
	}

	// pF: hard-frozen — no actions at all
	f := by["pF"]
	if !f.HardFrozen || f.CanSubmit || f.CanVerify || f.CanPublish || f.CanReturn {
		t.Errorf("pF hard-frozen exposed an action: submit=%v verify=%v publish=%v return=%v", f.CanSubmit, f.CanVerify, f.CanPublish, f.CanReturn)
	}
	if f.Hint == "" {
		t.Errorf("pF hard-frozen should carry a hint")
	}
}

// TestBuildApprovalBar_NoPermissions confirms the view-layer gate hides every
// action when the principal lacks the verbs (server is still authoritative).
func TestBuildApprovalBar_NoPermissions(t *testing.T) {
	deps := &PageViewDeps{Labels: outcome_matrix.DefaultLabels(), Routes: outcome_matrix.DefaultRoutes()}
	perms := types.NewEmptyUserPermissions()
	resp := &matrixpb.GetOutcomeMatrixResponse{
		Phases:          []*matrixpb.PhaseColumn{phaseCol("pA", "Sem 1", "tA")},
		ApprovalRollups: []*matrixpb.PhaseApprovalRollup{rollup("pA", sIP, false, true, false, 30, 0)},
	}
	bar := buildApprovalBar(deps, perms, resp, "tmpl-1", "")
	if len(bar) != 1 {
		t.Fatalf("want 1 phase, got %d", len(bar))
	}
	if bar[0].CanSubmit || bar[0].CanVerify || bar[0].CanPublish || bar[0].CanReturn {
		t.Errorf("no-permission principal should see no actions: %+v", bar[0])
	}
}

func TestBuildApprovalBar_ReturnReasonOnlyOnEditableInProgress(t *testing.T) {
	labels := outcome_matrix.DefaultLabels()
	deps := &PageViewDeps{Labels: labels, Routes: outcome_matrix.DefaultRoutes()}
	perms := types.NewEmptyUserPermissions()
	const reason = "Please add evidence"
	resp := &matrixpb.GetOutcomeMatrixResponse{
		Phases: []*matrixpb.PhaseColumn{
			phaseCol("pA", "Sem 1", "tA"),
			phaseCol("pB", "Sem 2", "tB"),
			phaseCol("pC", "Sem 3", "tC"),
			phaseCol("pD", "Sem 4", "tD"),
			phaseCol("pE", "Sem 5", "tE"),
		},
		ApprovalRollups: []*matrixpb.PhaseApprovalRollup{
			rollupWithReason("pA", sIP, false, true, false, 1, 0, reason),
			rollupWithReason("pB", sFR, false, true, false, 1, 0, reason),
			rollupWithReason("pC", sIP, true, true, false, 1, 0, reason),
			rollupWithReason("pD", sIP, false, true, true, 1, 0, reason),
			rollup("pE", sIP, false, true, false, 1, 0),
		},
	}

	bar := buildApprovalBar(deps, perms, resp, "tmpl-1", "")
	by := map[string]ApprovalPhase{}
	for _, phase := range bar {
		by[phase.PhaseID] = phase
	}

	want := subReason(labels.Approval.ReturnedReasonHint, reason)
	if by["pA"].ReturnReason != want {
		t.Errorf("in-progress return note = %q, want %q", by["pA"].ReturnReason, want)
	}
	indexed := phaseActions(bar, labels, "")
	phase, ok := indexed["pA"].(PhaseActions)
	if !ok {
		t.Fatalf("in-progress return note did not keep the visible header slot: %#v", indexed["pA"])
	}
	if phase.Phase.ReturnReason != want {
		t.Errorf("header action return note = %q, want %q", phase.Phase.ReturnReason, want)
	}
	for _, phaseID := range []string{"pB", "pC", "pD", "pE"} {
		if got := by[phaseID].ReturnReason; got != "" {
			t.Errorf("phase %s return note = %q, want empty", phaseID, got)
		}
	}
}

func TestPhaseActionsTemplate_ReturnReasonIsVisibleAndEscaped(t *testing.T) {
	renderer := pyeza.NewHTMLRendererFromFS(pyeza.SharedFS, outcome_matrix.TemplatesFS)
	if err := renderer.Init(); err != nil {
		t.Fatalf("init outcome-matrix templates: %v", err)
	}

	labels := outcome_matrix.DefaultLabels()
	const rawReason = `<script>alert("x")</script>`
	data := types.CellGridSlot{Actions: PhaseActions{Phase: ApprovalPhase{
		Slug:         "pA",
		ReturnReason: subReason(labels.Approval.ReturnedReasonHint, rawReason),
	}}}
	var rendered bytes.Buffer
	if err := renderer.GetTemplates().ExecuteTemplate(&rendered, "outcome-matrix-phase-actions", data); err != nil {
		t.Fatalf("render outcome-matrix phase actions: %v", err)
	}

	html := rendered.String()
	if !strings.Contains(html, `data-testid="pa-pA-return-note"`) {
		t.Fatalf("return note testid missing from visible header actions: %s", html)
	}
	if !strings.Contains(html, "Returned with a note:") || !strings.Contains(html, "&lt;script&gt;") {
		t.Fatalf("return reason was not rendered and escaped: %s", html)
	}
	if strings.Contains(html, "<script>alert") {
		t.Fatalf("return reason rendered raw HTML: %s", html)
	}
}

// TestPhaseEditableFunc pins the cell-editability render mirror: a cell is
// render-editable only when its phase's roll-up is cleanly IN_PROGRESS and not
// hard-frozen; a locked / mixed / frozen phase forces read-only.
func TestPhaseEditableFunc(t *testing.T) {
	resp := &matrixpb.GetOutcomeMatrixResponse{
		Phases: []*matrixpb.PhaseColumn{
			phaseCol("pA", "Sem 1", "tA"), // in_progress → editable
			phaseCol("pB", "Sem 2", "tB"), // for_review → locked
			phaseCol("pC", "Sem 3", "tC"), // in_progress but hard-frozen → locked
			phaseCol("pE", "Sem 5", "tE"), // mixed → locked
		},
		ApprovalRollups: []*matrixpb.PhaseApprovalRollup{
			rollup("pA", sIP, false, true, false, 1, 0),
			rollup("pB", sFR, false, true, false, 1, 0),
			rollup("pC", sIP, false, true, true, 1, 0),
			rollup("pE", sIP, true, true, false, 1, 0),
		},
	}
	allow := phaseEditableFunc(resp)
	cases := []struct {
		colKey string
		want   bool
	}{
		{"tA:crit1", true},
		{"tB:crit1", false},
		{"tC:crit1", false},
		{"tE:crit1", false},
		{"tUnknown:crit1", true}, // no mapping → permissive (server still guards)
	}
	for _, c := range cases {
		if got := allow(c.colKey); got != c.want {
			t.Errorf("phaseEditable(%q) = %v, want %v", c.colKey, got, c.want)
		}
	}
}
