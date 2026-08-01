package document

import (
	"context"
	"errors"
	"fmt"
	"testing"

	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
)

func strptr(s string) *string { return &s }

// testGroup is the card's own group for the group-grain tests; testSiblingGroup
// is a different group that shares the same template phase (the case the plan
// exists for: one template phase spanning every group it is delivered to).
const (
	testGroup        = "grp-palladium"
	testSiblingGroup = "grp-platinum"
)

// grainGroup is shorthand for the group-grain option value.
const grainGroup = outcome_summary.GateGrainSubscriptionGroup

// gateDeps builds a Deps with canned job_phase / job_task / task_outcome closures
// (same set returned for every query — used for singleton, null-template-phase
// cases). Zero-valued DocOptions → the template-grain (zero-value) gate.
func gateDeps(phases []*jobphasepb.JobPhase, tasks []*jobtaskpb.JobTask, outcomes []*taskoutcomepb.TaskOutcome) *Deps {
	return &Deps{
		ListJobPhases: func(_ context.Context, _ *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
			return &jobphasepb.ListJobPhasesResponse{Data: phases, Success: true}, nil
		},
		ListJobTasks: func(_ context.Context, _ *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
			return &jobtaskpb.ListJobTasksResponse{Data: tasks, Success: true}, nil
		},
		ListTaskOutcomes: func(_ context.Context, _ *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
			return &taskoutcomepb.ListTaskOutcomesResponse{Data: outcomes, Success: true}, nil
		},
	}
}

// fullSDeps distinguishes the by-job_id card read (cardPhases) from the
// by-template_phase_id full-sheet read (sheetPhases) so the full-S aggregate can be
// exercised (codex §B4). Zero-valued DocOptions → the template-grain gate.
func fullSDeps(cardPhases, sheetPhases []*jobphasepb.JobPhase, tasks []*jobtaskpb.JobTask, outcomes []*taskoutcomepb.TaskOutcome) *Deps {
	return (&gateFake{card: cardPhases, sheet: sheetPhases, tasks: tasks, outcomes: outcomes}).deps("", false)
}

// ---------------------------------------------------------------------------
// gateFake — a request-capturing Deps fake that MODELS THE GATE-ROLLUP PORT.
//
// The by-job_id card read returns `card`. Under the ZERO grain the
// by-template_phase_id sheet read returns `sheet` (the whole template sheet —
// the pre-option behavior). Under the GROUP grain the gate must never issue
// that read: it calls the rollup closure, which this fake computes HONESTLY
// from `sheet` + `group` (phase id → the group its job belongs to; the mapping
// is out of band here because it is out of band in production too — job_phase
// carries no group column, the relation lives on the group-membership rows and
// is resolved inside the provider's SQL). Per requested template phase the
// members are that group's active sheet phases; ZERO members ⇒ the rollup row
// is ABSENT (the provider contract — the consumer's coverage check turns
// absence into an error). has_data mirrors the canned tasks+outcomes.
//
// `rollup`, when set, replaces the honest model verbatim — the fail-closed /
// misbehaving-provider cases.
// ---------------------------------------------------------------------------
type gateFake struct {
	card     []*jobphasepb.JobPhase
	sheet    []*jobphasepb.JobPhase
	group    map[string]string // phase id → subscription_group_id
	tasks    []*jobtaskpb.JobTask
	outcomes []*taskoutcomepb.TaskOutcome
	sheetErr error

	rollup func(req *matrixpb.GetPhaseApprovalGateRollupRequest) (*matrixpb.GetPhaseApprovalGateRollupResponse, error)

	phaseReqs  []*jobphasepb.ListJobPhasesRequest
	taskReqs   []*jobtaskpb.ListJobTasksRequest
	outReqs    []*taskoutcomepb.ListTaskOutcomesRequest
	rollupReqs []*matrixpb.GetPhaseApprovalGateRollupRequest
}

// filterField returns the first filter's field name (the id filter is always index 0).
func filterField(req *jobphasepb.ListJobPhasesRequest) string {
	if f := req.GetFilters().GetFilters(); len(f) > 0 {
		return f[0].GetField()
	}
	return ""
}

func (g *gateFake) honestRollup(req *matrixpb.GetPhaseApprovalGateRollupRequest) (*matrixpb.GetPhaseApprovalGateRollupResponse, error) {
	hasData := len(g.tasks) > 0 && len(g.outcomes) > 0
	resp := &matrixpb.GetPhaseApprovalGateRollupResponse{Success: true}
	for _, tp := range req.GetJobTemplatePhaseIds() {
		var members []*jobphasepb.JobPhase
		for _, p := range g.sheet {
			if p.GetActive() && p.GetTemplatePhaseId() == tp && g.group[p.GetId()] == req.GetSubscriptionGroupId() {
				members = append(members, p)
			}
		}
		if len(members) == 0 {
			continue // zero-member phases are ABSENT from the response
		}
		anyEntered, allPublished := false, true
		for _, p := range members {
			if phaseWorkflowEntered(p) {
				anyEntered = true
			}
			if p.GetApprovalStatus() != jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED {
				allPublished = false
			}
		}
		resp.Rollups = append(resp.Rollups, &matrixpb.PhaseApprovalGateRollup{
			JobTemplatePhaseId:         tp,
			AppliedSubscriptionGroupId: req.GetSubscriptionGroupId(),
			TargetCount:                int32(len(members)),
			AnyWorkflowEntered:         anyEntered,
			AllPublished:               allPublished,
			HasData:                    hasData,
		})
	}
	return resp, nil
}

// deps wires the fake into a *Deps. grain selects DocOptions.GateGrain;
// rollupWired=false leaves the port nil (the unpublished-implementation case).
func (g *gateFake) deps(grain string, rollupWired bool) *Deps {
	d := &Deps{
		DocOptions: outcome_summary.DocumentOptions{GateGrain: grain},
		ListJobPhases: func(_ context.Context, req *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
			g.phaseReqs = append(g.phaseReqs, req)
			if filterField(req) != "template_phase_id" {
				return &jobphasepb.ListJobPhasesResponse{Data: g.card, Success: true}, nil
			}
			if g.sheetErr != nil {
				return nil, g.sheetErr
			}
			return &jobphasepb.ListJobPhasesResponse{Data: g.sheet, Success: true}, nil
		},
		ListJobTasks: func(_ context.Context, req *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
			g.taskReqs = append(g.taskReqs, req)
			return &jobtaskpb.ListJobTasksResponse{Data: g.tasks, Success: true}, nil
		},
		ListTaskOutcomes: func(_ context.Context, req *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
			g.outReqs = append(g.outReqs, req)
			return &taskoutcomepb.ListTaskOutcomesResponse{Data: g.outcomes, Success: true}, nil
		},
	}
	if rollupWired {
		d.GetPhaseApprovalGateRollup = func(_ context.Context, req *matrixpb.GetPhaseApprovalGateRollupRequest) (*matrixpb.GetPhaseApprovalGateRollupResponse, error) {
			g.rollupReqs = append(g.rollupReqs, req)
			if g.rollup != nil {
				return g.rollup(req)
			}
			return g.honestRollup(req)
		}
	}
	return d
}

// sheetReadCount counts the by-template_phase_id (template-sheet widening) reads.
func (g *gateFake) sheetReadCount() int {
	n := 0
	for _, r := range g.phaseReqs {
		if filterField(r) == "template_phase_id" {
			n++
		}
	}
	return n
}

func phase(id string, status jobphasepb.PhaseApprovalStatus, returnedBy string) *jobphasepb.JobPhase {
	p := &jobphasepb.JobPhase{Id: id, Active: true, ApprovalStatus: status}
	if returnedBy != "" {
		p.ReturnedBy = strptr(returnedBy)
	}
	return p
}

func sheetPhase(id, tp string, status jobphasepb.PhaseApprovalStatus, returnedBy string) *jobphasepb.JobPhase {
	p := phase(id, status, returnedBy)
	p.TemplatePhaseId = strptr(tp)
	return p
}

// TestReportRenderStatus pins the D5 render-gate predicate (singleton / null
// template-phase cases — evaluated as their own sheet). Q2: a NULL-template-
// phase singleton behaves EXACTLY the same under both grains, and the group
// grain issues NO rollup call for it (the singleton branch is reached only by
// the no-template-relation partition, never by a group step).
func TestReportRenderStatus(t *testing.T) {
	const (
		ip  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_IN_PROGRESS
		fr  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW
		pub = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED
	)
	task := &jobtaskpb.JobTask{Id: "jt-1", Active: true}
	outcome := &taskoutcomepb.TaskOutcome{Id: "to-1", Active: true}

	cases := []struct {
		name     string
		phases   []*jobphasepb.JobPhase
		tasks    []*jobtaskpb.JobTask
		outcomes []*taskoutcomepb.TaskOutcome
		want     bool
	}{
		{"never_workflowed_backfill_renders", []*jobphasepb.JobPhase{phase("jp-1", ip, "")}, []*jobtaskpb.JobTask{task}, []*taskoutcomepb.TaskOutcome{outcome}, false},
		{"for_review_with_data_blocks", []*jobphasepb.JobPhase{phase("jp-1", fr, "")}, []*jobtaskpb.JobTask{task}, []*taskoutcomepb.TaskOutcome{outcome}, true},
		{"published_renders", []*jobphasepb.JobPhase{phase("jp-1", pub, "")}, []*jobtaskpb.JobTask{task}, []*taskoutcomepb.TaskOutcome{outcome}, false},
		{"returned_with_data_blocks", []*jobphasepb.JobPhase{phase("jp-1", ip, "user-x")}, []*jobtaskpb.JobTask{task}, []*taskoutcomepb.TaskOutcome{outcome}, true},
		{"for_review_no_data_renders", []*jobphasepb.JobPhase{phase("jp-1", fr, "")}, []*jobtaskpb.JobTask{task}, []*taskoutcomepb.TaskOutcome{}, false},
		{"for_review_no_tasks_renders", []*jobphasepb.JobPhase{phase("jp-1", fr, "")}, []*jobtaskpb.JobTask{}, []*taskoutcomepb.TaskOutcome{outcome}, false},
	}
	for _, c := range cases {
		t.Run("zero_grain/"+c.name, func(t *testing.T) {
			d := gateDeps(c.phases, c.tasks, c.outcomes)
			got, err := reportRenderStatus(context.Background(), d, []string{"job-1"}, testGroup)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Errorf("reportRenderStatus = %v, want %v", got, c.want)
			}
		})
		t.Run("group_grain/"+c.name, func(t *testing.T) {
			g := &gateFake{card: c.phases, tasks: c.tasks, outcomes: c.outcomes}
			got, err := reportRenderStatus(context.Background(), g.deps(grainGroup, true), []string{"job-1"}, testGroup)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Errorf("reportRenderStatus = %v, want %v (singletons must behave identically under both grains)", got, c.want)
			}
			if n := len(g.rollupReqs); n != 0 {
				t.Errorf("a card of NULL-template-phase singletons must issue NO rollup call, got %d", n)
			}
		})
	}
}

// TestReportRenderStatus_FullSheet exercises the zero-grain full-S,
// template-phase-grain aggregate (codex §B4): the card's own phase alone would
// render, but a sibling member of the same approval sheet forces a block.
func TestReportRenderStatus_FullSheet(t *testing.T) {
	const (
		ip  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_IN_PROGRESS
		fr  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW
		pub = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED
	)
	task := &jobtaskpb.JobTask{Id: "jt-1", Active: true}
	outcome := &taskoutcomepb.TaskOutcome{Id: "to-1", Active: true}

	// (2) published member with data + a late pristine IN_PROGRESS member → block.
	// The card only holds the PUBLISHED phase (would render alone), but the sheet has
	// a pristine sibling, so BOOL_AND(published) is false → block.
	t.Run("published_plus_pristine_member_blocks", func(t *testing.T) {
		card := []*jobphasepb.JobPhase{sheetPhase("jp-pub", "sheet-1", pub, "")}
		sheet := []*jobphasepb.JobPhase{
			sheetPhase("jp-pub", "sheet-1", pub, ""),
			sheetPhase("jp-pristine", "sheet-1", ip, ""),
		}
		d := fullSDeps(card, sheet, []*jobtaskpb.JobTask{task}, []*taskoutcomepb.TaskOutcome{outcome})
		got, err := reportRenderStatus(context.Background(), d, []string{"job-1"}, testGroup)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Errorf("full-S with a pristine sibling member must block, got render")
		}
	})

	// (3) workflow-entered member and data-bearing member differ.
	t.Run("entered_and_data_on_different_members_blocks", func(t *testing.T) {
		card := []*jobphasepb.JobPhase{sheetPhase("jp-a", "sheet-1", ip, "")} // card's own is pristine
		sheet := []*jobphasepb.JobPhase{
			sheetPhase("jp-a", "sheet-1", ip, ""), // data-bearing (below), not entered
			sheetPhase("jp-b", "sheet-1", fr, ""), // entered, no data of its own
		}
		d := fullSDeps(card, sheet, []*jobtaskpb.JobTask{task}, []*taskoutcomepb.TaskOutcome{outcome})
		got, err := reportRenderStatus(context.Background(), d, []string{"job-1"}, testGroup)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Errorf("entered + data on different members must block, got render")
		}
	})

	// fully published sheet renders.
	t.Run("all_published_renders", func(t *testing.T) {
		card := []*jobphasepb.JobPhase{sheetPhase("jp-a", "sheet-1", pub, "")}
		sheet := []*jobphasepb.JobPhase{
			sheetPhase("jp-a", "sheet-1", pub, ""),
			sheetPhase("jp-b", "sheet-1", pub, ""),
		}
		d := fullSDeps(card, sheet, []*jobtaskpb.JobTask{task}, []*taskoutcomepb.TaskOutcome{outcome})
		got, err := reportRenderStatus(context.Background(), d, []string{"job-1"}, testGroup)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Errorf("fully published sheet must render, got blocked")
		}
	})
}

// twoGroupSheet reproduces the shape the plan exists for: ONE template phase
// delivered to two groups. testGroup was driven through the workflow
// (VERIFIED); testSiblingGroup was never touched (pristine IN_PROGRESS, every
// approval audit stamp null).
func twoGroupSheet() (sheet []*jobphasepb.JobPhase, group map[string]string) {
	const (
		ip  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_IN_PROGRESS
		ver = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_VERIFIED
	)
	a1 := sheetPhase("jp-pal-1", "tp-1", ver, "")
	a2 := sheetPhase("jp-pal-2", "tp-1", ver, "")
	b1 := sheetPhase("jp-plat-1", "tp-1", ip, "")
	b2 := sheetPhase("jp-plat-2", "tp-1", ip, "")
	return []*jobphasepb.JobPhase{a1, a2, b1, b2}, map[string]string{
		"jp-pal-1":  testGroup,
		"jp-pal-2":  testGroup,
		"jp-plat-1": testSiblingGroup,
		"jp-plat-2": testSiblingGroup,
	}
}

// TestReportRenderStatus_GroupGrain is THE regression this plan exists for:
// one template phase, two groups, only one of them mid-workflow. Under the
// zero-value template grain BOTH cards 409. Under the group grain the
// workflow-entered group blocks and the untouched sibling renders — and the
// SAME fixture through the zero-grain path proves the OPTION is what flips it.
func TestReportRenderStatus_GroupGrain(t *testing.T) {
	task := &jobtaskpb.JobTask{Id: "jt-1", Active: true}
	outcome := &taskoutcomepb.TaskOutcome{Id: "to-1", Active: true}
	sheet, group := twoGroupSheet()

	newFake := func(card []*jobphasepb.JobPhase) *gateFake {
		return &gateFake{
			card: card, sheet: sheet, group: group,
			tasks: []*jobtaskpb.JobTask{task}, outcomes: []*taskoutcomepb.TaskOutcome{outcome},
		}
	}

	// A: the VERIFIED group. Its OWN rows are mid-workflow → must still 409.
	t.Run("own_group_verified_blocks", func(t *testing.T) {
		g := newFake([]*jobphasepb.JobPhase{sheet[0]})
		got, err := reportRenderStatus(context.Background(), g.deps(grainGroup, true), []string{"job-pal"}, testGroup)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Errorf("a card whose OWN group is VERIFIED (workflow-entered, not published) must block, got render")
		}
	})

	// B: the untouched sibling group. Pristine IN_PROGRESS, no audit stamps →
	// never-workflowed carve-out applies to ITS sheet → must render.
	t.Run("sibling_group_mid_workflow_does_not_block", func(t *testing.T) {
		g := newFake([]*jobphasepb.JobPhase{sheet[2]})
		got, err := reportRenderStatus(context.Background(), g.deps(grainGroup, true), []string{"job-plat"}, testSiblingGroup)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Errorf("a pristine group must render even though a SIBLING group is mid-workflow, got blocked")
		}
		if n := g.sheetReadCount(); n != 0 {
			t.Errorf("group grain must not issue the template-sheet widening read, got %d", n)
		}
	})

	// CONTROL — the SAME fixture through the ZERO grain: the sibling's card
	// still blocks (the stricter template-phase sheet). This is what makes the
	// test above a genuine regression test rather than a tautology: the OPTION
	// is the only thing that changed.
	t.Run("control_zero_grain_blocks_both", func(t *testing.T) {
		g := newFake([]*jobphasepb.JobPhase{sheet[2]})
		got, err := reportRenderStatus(context.Background(), g.deps("", true), []string{"job-plat"}, testSiblingGroup)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Errorf("zero grain: the sibling group's card must still block (template-phase sheet), got render")
		}
		if n := len(g.rollupReqs); n != 0 {
			t.Errorf("zero grain must never call the rollup port, got %d calls", n)
		}
	})

	// Same group, mixed states → still blocks. Narrowing must not weaken the
	// gate WITHIN a group.
	t.Run("same_group_mixed_states_still_block", func(t *testing.T) {
		const (
			ip  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_IN_PROGRESS
			fr  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW
			pub = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED
		)
		mixed := []*jobphasepb.JobPhase{
			sheetPhase("jp-x-1", "tp-1", pub, ""), // published, has data
			sheetPhase("jp-x-2", "tp-1", fr, ""),  // group-mate still for-review
			sheetPhase("jp-x-3", "tp-1", ip, ""),  // group-mate pristine
			sheetPhase("jp-y-1", "tp-1", pub, ""), // OTHER group, fully published
		}
		g := &gateFake{
			card:  []*jobphasepb.JobPhase{mixed[0]},
			sheet: mixed,
			group: map[string]string{
				"jp-x-1": testGroup, "jp-x-2": testGroup, "jp-x-3": testGroup,
				"jp-y-1": testSiblingGroup,
			},
			tasks:    []*jobtaskpb.JobTask{task},
			outcomes: []*taskoutcomepb.TaskOutcome{outcome},
		}
		got, err := reportRenderStatus(context.Background(), g.deps(grainGroup, true), []string{"job-x"}, testGroup)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Errorf("mixed states WITHIN the card's own group must still block, got render")
		}
	})
}

// TestReportRenderStatus_GroupGrainMixedCard pins the Q2 partition on a card
// that carries BOTH kinds of phase at once: template-backed sheets (proven
// through the rollup) AND a NULL-template-phase singleton (evaluated locally).
// blocked = OR(group sheets) OR OR(singletons) — a clean rollup verdict must
// NOT swallow a blocked singleton (the group step returning "clean" is not the
// end of the evaluation), and a clean singleton must leave the rollup verdict
// in charge. This is the mutation-killer for a groupGrainRenderStatus rewrite
// that returns the group verdict directly whenever template sheets exist.
func TestReportRenderStatus_GroupGrainMixedCard(t *testing.T) {
	const (
		ip  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_IN_PROGRESS
		fr  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW
		ver = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_VERIFIED
		pub = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED
	)
	task := &jobtaskpb.JobTask{Id: "jt-1", Active: true}
	outcome := &taskoutcomepb.TaskOutcome{Id: "to-1", Active: true}

	// Fully published group sheet (clean) + a FOR_REVIEW singleton with data
	// (blocked) → the card must still 409: the singleton branch runs AFTER a
	// clean group step, it is never skipped because template sheets exist.
	t.Run("clean_group_sheet_plus_blocked_singleton_blocks", func(t *testing.T) {
		g := &gateFake{
			card: []*jobphasepb.JobPhase{
				sheetPhase("jp-tpl", "tp-1", pub, ""),
				phase("jp-solo", fr, ""), // NULL template_phase_id
			},
			sheet: []*jobphasepb.JobPhase{
				sheetPhase("jp-tpl", "tp-1", pub, ""),
				sheetPhase("jp-tpl-2", "tp-1", pub, ""),
			},
			group: map[string]string{"jp-tpl": testGroup, "jp-tpl-2": testGroup},
			tasks: []*jobtaskpb.JobTask{task}, outcomes: []*taskoutcomepb.TaskOutcome{outcome},
		}
		got, err := reportRenderStatus(context.Background(), g.deps(grainGroup, true), []string{"job-1"}, testGroup)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Errorf("a blocked singleton on a mixed card must block even when every group sheet is clean, got render")
		}
		if n := len(g.rollupReqs); n != 1 {
			t.Errorf("the mixed card's group step must still run (exactly one rollup call), got %d", n)
		}
	})

	// The inverse: pristine singleton (clean) + a VERIFIED group sheet
	// (blocked) → the rollup verdict decides and the card 409s.
	t.Run("clean_singleton_defers_to_blocking_rollup_verdict", func(t *testing.T) {
		g := &gateFake{
			card: []*jobphasepb.JobPhase{
				sheetPhase("jp-tpl", "tp-1", ver, ""),
				phase("jp-solo", ip, ""), // pristine: never workflowed
			},
			sheet: []*jobphasepb.JobPhase{sheetPhase("jp-tpl", "tp-1", ver, "")},
			group: map[string]string{"jp-tpl": testGroup},
			tasks: []*jobtaskpb.JobTask{task}, outcomes: []*taskoutcomepb.TaskOutcome{outcome},
		}
		got, err := reportRenderStatus(context.Background(), g.deps(grainGroup, true), []string{"job-1"}, testGroup)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Errorf("a mid-workflow group sheet must block the mixed card regardless of its clean singleton, got render")
		}
	})

	// Both partitions clean → render, with the group step consulted exactly
	// once and no template-sheet widening read.
	t.Run("clean_group_sheet_and_clean_singleton_render", func(t *testing.T) {
		g := &gateFake{
			card: []*jobphasepb.JobPhase{
				sheetPhase("jp-tpl", "tp-1", pub, ""),
				phase("jp-solo", ip, ""), // pristine: never workflowed
			},
			sheet: []*jobphasepb.JobPhase{
				sheetPhase("jp-tpl", "tp-1", pub, ""),
				sheetPhase("jp-tpl-2", "tp-1", pub, ""),
			},
			group: map[string]string{"jp-tpl": testGroup, "jp-tpl-2": testGroup},
			tasks: []*jobtaskpb.JobTask{task}, outcomes: []*taskoutcomepb.TaskOutcome{outcome},
		}
		got, err := reportRenderStatus(context.Background(), g.deps(grainGroup, true), []string{"job-1"}, testGroup)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Errorf("a clean mixed card must render, got blocked")
		}
		if n := len(g.rollupReqs); n != 1 {
			t.Errorf("the mixed card's group step must run exactly once, got %d rollup calls", n)
		}
		if n := g.sheetReadCount(); n != 0 {
			t.Errorf("the mixed card must never issue the template-sheet widening read, got %d", n)
		}
	})
}

// TestReportRenderStatus_GroupGrainRequestShape pins the group-grain wire
// contract: the card read stays the UNNARROWED by-job_id read (no group field —
// the additive request field stays unconsumed), the template-sheet widening
// read is NEVER issued, and exactly ONE rollup call carries the exact group id
// plus every distinct card template phase.
func TestReportRenderStatus_GroupGrainRequestShape(t *testing.T) {
	sheet, group := twoGroupSheet()
	g := &gateFake{
		card: []*jobphasepb.JobPhase{sheet[0]}, sheet: sheet, group: group,
		tasks:    []*jobtaskpb.JobTask{{Id: "jt-1", Active: true}},
		outcomes: []*taskoutcomepb.TaskOutcome{{Id: "to-1", Active: true}},
	}
	if _, err := reportRenderStatus(context.Background(), g.deps(grainGroup, true), []string{"job-pal"}, testGroup); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(g.phaseReqs) != 1 {
		t.Fatalf("group grain must issue exactly the by-job_id card read, got %d job_phase reads", len(g.phaseReqs))
	}
	cardReq := g.phaseReqs[0]
	if filterField(cardReq) != "job_id" {
		t.Errorf("card read filter field = %q, want job_id", filterField(cardReq))
	}
	if n := len(cardReq.GetFilters().GetFilters()); n != 1 {
		t.Errorf("card read must carry exactly the job_id filter (byte-identical to the pre-option gate), got %d filters", n)
	}
	if cardReq.GetSubscriptionGroupId() != "" || cardReq.SubscriptionGroupId != nil {
		t.Errorf("card read must NOT set the request group field — the additive field stays unconsumed")
	}
	if n := g.sheetReadCount(); n != 0 {
		t.Errorf("group grain must issue ZERO template_phase_id widening reads, got %d", n)
	}
	if len(g.rollupReqs) != 1 {
		t.Fatalf("group grain must issue exactly ONE rollup call, got %d", len(g.rollupReqs))
	}
	req := g.rollupReqs[0]
	if req.GetSubscriptionGroupId() != testGroup {
		t.Errorf("rollup request group = %q, want %q", req.GetSubscriptionGroupId(), testGroup)
	}
	if len(req.GetJobTemplatePhaseIds()) != 1 || req.GetJobTemplatePhaseIds()[0] != "tp-1" {
		t.Errorf("rollup request template phases = %v, want [tp-1]", req.GetJobTemplatePhaseIds())
	}
}

// TestReportRenderStatus_ZeroGrainByteIdentity pins the zero-value regression
// proof: with GateGrain unset the gate issues EXACTLY the pre-option request
// sequence — the by-job_id card read then the by-template_phase_id sheet read,
// each with the single id filter, the pageLimit offset pagination, the id-asc
// sort, and NO group field on the request — and NEVER calls the rollup port,
// even when one is wired.
func TestReportRenderStatus_ZeroGrainByteIdentity(t *testing.T) {
	fr := jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW
	card := []*jobphasepb.JobPhase{sheetPhase("jp-1", "tp-1", fr, "")}
	g := &gateFake{
		card: card, sheet: card,
		tasks:    []*jobtaskpb.JobTask{{Id: "jt-1", Active: true}},
		outcomes: []*taskoutcomepb.TaskOutcome{{Id: "to-1", Active: true}},
	}
	got, err := reportRenderStatus(context.Background(), g.deps("", true), []string{"job-1"}, testGroup)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Errorf("for-review sheet with data must block under the zero grain, got render")
	}

	if len(g.rollupReqs) != 0 {
		t.Fatalf("zero grain must NEVER call the rollup port, got %d calls", len(g.rollupReqs))
	}
	if len(g.phaseReqs) != 2 {
		t.Fatalf("zero grain must issue the card read then the sheet read, got %d job_phase reads", len(g.phaseReqs))
	}
	wantFields := []string{"job_id", "template_phase_id"}
	for i, req := range g.phaseReqs {
		f := req.GetFilters().GetFilters()
		if len(f) != 1 {
			t.Errorf("read %d must carry exactly one filter, got %d", i, len(f))
			continue
		}
		if f[0].GetField() != wantFields[i] {
			t.Errorf("read %d filter field = %q, want %q", i, f[0].GetField(), wantFields[i])
		}
		if req.GetSubscriptionGroupId() != "" || req.SubscriptionGroupId != nil {
			t.Errorf("read %d must NOT set the request group field (byte-identical pre-option request)", i)
		}
		if p := req.GetPagination(); p.GetLimit() != pageLimit || p.GetOffset().GetPage() != 1 {
			t.Errorf("read %d pagination = limit %d page %d, want limit %d page 1", i, p.GetLimit(), p.GetOffset().GetPage(), pageLimit)
		}
		s := req.GetSort().GetFields()
		if len(s) != 1 || s[0].GetField() != "id" {
			t.Errorf("read %d must keep the deterministic id sort", i)
		}
	}
	if len(g.taskReqs) != 1 || len(g.outReqs) != 1 {
		t.Errorf("zero grain probe = %d task reads / %d outcome reads, want 1/1", len(g.taskReqs), len(g.outReqs))
	}
}

// TestReportRenderStatus_GroupGrainFailClosed covers every unprovable rollup
// state. EACH case must return an ERROR (→ handler 503) — never a render,
// never a template-grain widen, never a singleton degrade.
func TestReportRenderStatus_GroupGrainFailClosed(t *testing.T) {
	fr := jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW
	card := []*jobphasepb.JobPhase{sheetPhase("jp-own", "tp-1", fr, "")}
	task := &jobtaskpb.JobTask{Id: "jt-1", Active: true}
	outcome := &taskoutcomepb.TaskOutcome{Id: "to-1", Active: true}

	newFake := func() *gateFake {
		return &gateFake{
			card: card, sheet: card, group: map[string]string{"jp-own": testGroup},
			tasks: []*jobtaskpb.JobTask{task}, outcomes: []*taskoutcomepb.TaskOutcome{outcome},
		}
	}
	mustErr := func(t *testing.T, d *Deps, groupID, why string) {
		t.Helper()
		blocked, err := reportRenderStatus(context.Background(), d, []string{"job-own"}, groupID)
		if err == nil {
			t.Fatalf("%s must fail closed (error), got blocked=%v err=nil", why, blocked)
		}
		if blocked {
			t.Errorf("%s must not report blocked alongside the error", why)
		}
	}

	// Grain configured, port unpublished (nil closure) → 503, never a
	// template-grain fallback.
	t.Run("nil_rollup_closure_errors", func(t *testing.T) {
		g := newFake()
		mustErr(t, g.deps(grainGroup, false), testGroup, "group grain with a nil rollup port")
		if n := g.sheetReadCount(); n != 0 {
			t.Errorf("the failed group path must NOT widen to the template-sheet read, got %d reads", n)
		}
	})

	// Empty group id on a group-grain gate → error (unreachable via the
	// handler, which 404s first — the gate re-checks; it must NOT degrade to
	// a singleton evaluation).
	t.Run("empty_group_id_errors", func(t *testing.T) {
		mustErr(t, newFake().deps(grainGroup, true), "", "group grain with an empty group id")
	})

	t.Run("rollup_transport_error_fails_closed", func(t *testing.T) {
		g := newFake()
		g.rollup = func(_ *matrixpb.GetPhaseApprovalGateRollupRequest) (*matrixpb.GetPhaseApprovalGateRollupResponse, error) {
			return nil, errors.New("permission denied: job_phase:list")
		}
		mustErr(t, g.deps(grainGroup, true), testGroup, "a rollup transport error")
	})

	t.Run("nil_response_fails_closed", func(t *testing.T) {
		g := newFake()
		g.rollup = func(_ *matrixpb.GetPhaseApprovalGateRollupRequest) (*matrixpb.GetPhaseApprovalGateRollupResponse, error) {
			return nil, nil
		}
		mustErr(t, g.deps(grainGroup, true), testGroup, "a nil rollup response")
	})

	t.Run("success_false_fails_closed", func(t *testing.T) {
		g := newFake()
		g.rollup = func(_ *matrixpb.GetPhaseApprovalGateRollupRequest) (*matrixpb.GetPhaseApprovalGateRollupResponse, error) {
			return &matrixpb.GetPhaseApprovalGateRollupResponse{Success: false}, nil
		}
		mustErr(t, g.deps(grainGroup, true), testGroup, "a success=false rollup response")
	})

	// Coverage gap: a requested template phase with no rollup row. This is
	// also the R-1/R-3 pin — an empty/under-resolved narrow is UNPROVABLE,
	// never "no sheet, render away" and never a singleton degrade.
	t.Run("missing_coverage_fails_closed", func(t *testing.T) {
		g := newFake()
		g.rollup = func(_ *matrixpb.GetPhaseApprovalGateRollupRequest) (*matrixpb.GetPhaseApprovalGateRollupResponse, error) {
			return &matrixpb.GetPhaseApprovalGateRollupResponse{Success: true}, nil // no rollups at all
		}
		mustErr(t, g.deps(grainGroup, true), testGroup, "a coverage gap (requested phase absent)")
	})

	// The ignored-provider model: an implementation that ignored the narrow
	// and aggregated the WHOLE template sheet cannot echo the requested group.
	t.Run("echo_mismatch_sibling_fails_closed", func(t *testing.T) {
		g := newFake()
		g.rollup = func(req *matrixpb.GetPhaseApprovalGateRollupRequest) (*matrixpb.GetPhaseApprovalGateRollupResponse, error) {
			return &matrixpb.GetPhaseApprovalGateRollupResponse{Success: true, Rollups: []*matrixpb.PhaseApprovalGateRollup{{
				JobTemplatePhaseId:         "tp-1",
				AppliedSubscriptionGroupId: testSiblingGroup, // applied/echoed the WRONG group
				TargetCount:                4,
				AnyWorkflowEntered:         true,
			}}}, nil
		}
		mustErr(t, g.deps(grainGroup, true), testGroup, "a wrong-group echo")
	})

	t.Run("echo_empty_fails_closed", func(t *testing.T) {
		g := newFake()
		g.rollup = func(req *matrixpb.GetPhaseApprovalGateRollupRequest) (*matrixpb.GetPhaseApprovalGateRollupResponse, error) {
			return &matrixpb.GetPhaseApprovalGateRollupResponse{Success: true, Rollups: []*matrixpb.PhaseApprovalGateRollup{{
				JobTemplatePhaseId: "tp-1", // zero-value echo: an old/ignoring implementation
				TargetCount:        4,
			}}}, nil
		}
		mustErr(t, g.deps(grainGroup, true), testGroup, "an empty applied-group echo")
	})

	// Card anchor: collectCard proved this card's own membership, so a
	// correct narrow can never resolve an EMPTY sheet for a phase the card
	// draws from. target_count=0 is an under-resolved read.
	t.Run("missing_card_anchor_fails_closed", func(t *testing.T) {
		g := newFake()
		g.rollup = func(req *matrixpb.GetPhaseApprovalGateRollupRequest) (*matrixpb.GetPhaseApprovalGateRollupResponse, error) {
			return &matrixpb.GetPhaseApprovalGateRollupResponse{Success: true, Rollups: []*matrixpb.PhaseApprovalGateRollup{{
				JobTemplatePhaseId:         "tp-1",
				AppliedSubscriptionGroupId: testGroup,
				TargetCount:                0,
			}}}, nil
		}
		mustErr(t, g.deps(grainGroup, true), testGroup, "a zero-member rollup (missing card anchor)")
	})

	// Dedupe pin: one evaluation per template phase — a duplicate row is
	// ambiguous coverage (a hypothetical two-group job can never double-block
	// or double-count through this seam).
	t.Run("duplicate_rollup_rows_fail_closed", func(t *testing.T) {
		g := newFake()
		g.rollup = func(req *matrixpb.GetPhaseApprovalGateRollupRequest) (*matrixpb.GetPhaseApprovalGateRollupResponse, error) {
			row := &matrixpb.PhaseApprovalGateRollup{
				JobTemplatePhaseId:         "tp-1",
				AppliedSubscriptionGroupId: testGroup,
				TargetCount:                1,
			}
			return &matrixpb.GetPhaseApprovalGateRollupResponse{Success: true, Rollups: []*matrixpb.PhaseApprovalGateRollup{row, row}}, nil
		}
		mustErr(t, g.deps(grainGroup, true), testGroup, "duplicate rollup rows for one template phase")
	})

	t.Run("unrequested_phase_fails_closed", func(t *testing.T) {
		g := newFake()
		g.rollup = func(req *matrixpb.GetPhaseApprovalGateRollupRequest) (*matrixpb.GetPhaseApprovalGateRollupResponse, error) {
			return &matrixpb.GetPhaseApprovalGateRollupResponse{Success: true, Rollups: []*matrixpb.PhaseApprovalGateRollup{
				{JobTemplatePhaseId: "tp-1", AppliedSubscriptionGroupId: testGroup, TargetCount: 1},
				{JobTemplatePhaseId: "tp-foreign", AppliedSubscriptionGroupId: testGroup, TargetCount: 1},
			}}, nil
		}
		mustErr(t, g.deps(grainGroup, true), testGroup, "a rollup for an unrequested template phase")
	})
}

// TestReportRenderStatus_GroupGrainCarveOuts pins the render-permitting group
// sheets and the edges around the rollup call.
func TestReportRenderStatus_GroupGrainCarveOuts(t *testing.T) {
	const (
		ip  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_IN_PROGRESS
		pub = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED
	)
	task := &jobtaskpb.JobTask{Id: "jt-1", Active: true}
	outcome := &taskoutcomepb.TaskOutcome{Id: "to-1", Active: true}

	// Never-workflowed group sheet (any_workflow_entered=false, has_data=true)
	// → the backfill carve-out: renders.
	t.Run("never_workflowed_group_sheet_renders", func(t *testing.T) {
		members := []*jobphasepb.JobPhase{
			sheetPhase("jp-1", "tp-1", ip, ""),
			sheetPhase("jp-2", "tp-1", ip, ""),
		}
		g := &gateFake{
			card: members[:1], sheet: members,
			group: map[string]string{"jp-1": testGroup, "jp-2": testGroup},
			tasks: []*jobtaskpb.JobTask{task}, outcomes: []*taskoutcomepb.TaskOutcome{outcome},
		}
		got, err := reportRenderStatus(context.Background(), g.deps(grainGroup, true), []string{"job-1"}, testGroup)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Errorf("a never-workflowed group sheet must render (backfill carve-out), got blocked")
		}
	})

	// Fully published group sheet → renders (republish clears the gate).
	t.Run("fully_published_group_sheet_renders", func(t *testing.T) {
		members := []*jobphasepb.JobPhase{
			sheetPhase("jp-1", "tp-1", pub, ""),
			sheetPhase("jp-2", "tp-1", pub, ""),
		}
		g := &gateFake{
			card: members[:1], sheet: members,
			group: map[string]string{"jp-1": testGroup, "jp-2": testGroup},
			tasks: []*jobtaskpb.JobTask{task}, outcomes: []*taskoutcomepb.TaskOutcome{outcome},
		}
		got, err := reportRenderStatus(context.Background(), g.deps(grainGroup, true), []string{"job-1"}, testGroup)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Errorf("a fully published group sheet must render, got blocked")
		}
	})

	// Data-less entered sheet → renders. Pins fayna's CONSUMPTION of the
	// rollup's has_data field (locked predicate: entered AND NOT all_published
	// AND data-present — h2-synthesis §B4): a sheet mid-workflow with no data
	// anywhere is the zero-grain "for_review_no_data_renders" twin. Dropping
	// the HasData term from the block predicate would silently 409 this case
	// (over-blocking regression) — this is the mutation-killer for that drift.
	t.Run("entered_not_published_no_data_group_sheet_renders", func(t *testing.T) {
		fr := jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW
		g := &gateFake{
			card:  []*jobphasepb.JobPhase{sheetPhase("jp-own", "tp-1", fr, "")},
			tasks: []*jobtaskpb.JobTask{task}, outcomes: []*taskoutcomepb.TaskOutcome{outcome},
		}
		g.rollup = func(_ *matrixpb.GetPhaseApprovalGateRollupRequest) (*matrixpb.GetPhaseApprovalGateRollupResponse, error) {
			return &matrixpb.GetPhaseApprovalGateRollupResponse{Success: true, Rollups: []*matrixpb.PhaseApprovalGateRollup{{
				JobTemplatePhaseId:         "tp-1",
				AppliedSubscriptionGroupId: testGroup,
				TargetCount:                2,
				AnyWorkflowEntered:         true,
				AllPublished:               false,
				HasData:                    false, // entered, unpublished, but data-less
			}}}, nil
		}
		got, err := reportRenderStatus(context.Background(), g.deps(grainGroup, true), []string{"job-1"}, testGroup)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Errorf("an entered-but-data-less group sheet must render (has_data=false), got blocked")
		}
	})

	// Historical card: every card phase INACTIVE → the active filter empties
	// the template-phase set → NO rollup call, renders as today (the
	// documented over-refusal edge only arises if live data ever presents
	// ACTIVE phases under inactive membership — and then the gate 503s, the
	// fail-SAFE direction, covered by missing_coverage/missing_card_anchor).
	t.Run("historical_all_inactive_card_makes_no_rollup_call", func(t *testing.T) {
		p := sheetPhase("jp-old", "tp-old", pub, "")
		p.Active = false
		g := &gateFake{card: []*jobphasepb.JobPhase{p},
			tasks: []*jobtaskpb.JobTask{task}, outcomes: []*taskoutcomepb.TaskOutcome{outcome}}
		got, err := reportRenderStatus(context.Background(), g.deps(grainGroup, true), []string{"job-old"}, testGroup)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Errorf("an all-inactive historical card must render, got blocked")
		}
		if n := len(g.rollupReqs); n != 0 {
			t.Errorf("an all-inactive card must issue NO rollup call, got %d", n)
		}
	})

	// The singleton data-probe deps still FAIL CLOSED when nil with
	// candidates present — under BOTH grains.
	t.Run("singleton_probe_deps_nil_fail_closed_both_grains", func(t *testing.T) {
		fr := jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW
		for _, grain := range []string{"", grainGroup} {
			d := &Deps{
				DocOptions: outcome_summary.DocumentOptions{GateGrain: grain},
				ListJobPhases: func(_ context.Context, _ *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
					return &jobphasepb.ListJobPhasesResponse{Data: []*jobphasepb.JobPhase{phase("jp-1", fr, "")}, Success: true}, nil
				},
				GetPhaseApprovalGateRollup: func(_ context.Context, _ *matrixpb.GetPhaseApprovalGateRollupRequest) (*matrixpb.GetPhaseApprovalGateRollupResponse, error) {
					return &matrixpb.GetPhaseApprovalGateRollupResponse{Success: true}, nil
				},
			}
			if _, err := reportRenderStatus(context.Background(), d, []string{"job-1"}, testGroup); err == nil {
				t.Errorf("grain %q: nil ListJobTasks/ListTaskOutcomes with singleton candidates must fail closed (error), got nil", grain)
			}
		}
	})
}

// TestReportRenderStatus_GroupGrainLargeGroup pins H-1 absence on the group
// path: a >100-member group sheet arrives as ONE aggregate row — the gate
// issues no member-item read and no pagination whatsoever, so the generic-list
// default 100-row page cap structurally cannot exist here. (The SQL side of
// the pin — the aggregate applying the predicate before GROUP BY with no
// LIMIT/OFFSET — is pinned by the provider's own tests.)
func TestReportRenderStatus_GroupGrainLargeGroup(t *testing.T) {
	const (
		ip  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_IN_PROGRESS
		ver = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_VERIFIED
	)
	// 120 members in ONE group on one template phase — one VERIFIED, 119
	// pristine. Well past the generic-list 100-row default cap.
	var members []*jobphasepb.JobPhase
	group := map[string]string{}
	for i := 0; i < 120; i++ {
		id := fmt.Sprintf("jp-%03d", i)
		st := ip
		if i == 0 {
			st = ver
		}
		members = append(members, sheetPhase(id, "tp-1", st, ""))
		group[id] = testGroup
	}
	g := &gateFake{
		card: members[1:2], sheet: members, group: group,
		tasks:    []*jobtaskpb.JobTask{{Id: "jt-1", Active: true}},
		outcomes: []*taskoutcomepb.TaskOutcome{{Id: "to-1", Active: true}},
	}
	got, err := reportRenderStatus(context.Background(), g.deps(grainGroup, true), []string{"job-1"}, testGroup)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Errorf("a 120-member sheet with one VERIFIED member must block — the 101st+ member must not be lost to a page cap")
	}
	if len(g.rollupReqs) != 1 {
		t.Errorf("the large sheet must arrive via exactly ONE rollup call, got %d", len(g.rollupReqs))
	}
	if n := g.sheetReadCount(); n != 0 {
		t.Errorf("the group path must issue no member-item reads at all, got %d", n)
	}
	if want := int32(120); g.rollupReqs != nil {
		// sanity: the honest model really did aggregate every member.
		resp, _ := g.honestRollup(g.rollupReqs[0])
		if len(resp.GetRollups()) != 1 || resp.GetRollups()[0].GetTargetCount() != want {
			t.Errorf("fixture sanity: honest rollup target_count = %v, want %d", resp.GetRollups(), want)
		}
	}
}

// TestReportRenderStatus_FailClosed confirms the gate FAILS CLOSED (returns an
// error → 503) on nil deps or any list error / permission denial (codex §B4).
func TestReportRenderStatus_FailClosed(t *testing.T) {
	t.Run("nil_deps_errors", func(t *testing.T) {
		if _, err := reportRenderStatus(context.Background(), nil, []string{"job-1"}, testGroup); err == nil {
			t.Errorf("nil Deps must fail closed (error), got nil")
		}
	})

	t.Run("nil_job_phase_dep_errors", func(t *testing.T) {
		if _, err := reportRenderStatus(context.Background(), &Deps{}, []string{"job-1"}, testGroup); err == nil {
			t.Errorf("nil ListJobPhases must fail closed (error), got nil")
		}
	})

	t.Run("empty_job_set_renders", func(t *testing.T) {
		blocked, err := reportRenderStatus(context.Background(), gateDeps(nil, nil, nil), nil, testGroup)
		if err != nil || blocked {
			t.Errorf("empty job set must render with no error, got blocked=%v err=%v", blocked, err)
		}
	})

	t.Run("job_phase_list_error_fails_closed", func(t *testing.T) {
		d := &Deps{
			ListJobPhases: func(_ context.Context, _ *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
				return nil, errors.New("permission denied: job_phase:list")
			},
		}
		if _, err := reportRenderStatus(context.Background(), d, []string{"job-1"}, testGroup); err == nil {
			t.Errorf("a list error / permission denial must fail closed (error), got nil")
		}
	})

	t.Run("sheet_list_error_fails_closed", func(t *testing.T) {
		fr := jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW
		g := &gateFake{
			card:     []*jobphasepb.JobPhase{sheetPhase("jp-1", "tp-1", fr, "")},
			sheetErr: errors.New("permission denied: job_phase:list"),
			tasks:    []*jobtaskpb.JobTask{{Id: "jt-1", Active: true}},
			outcomes: []*taskoutcomepb.TaskOutcome{{Id: "to-1", Active: true}},
		}
		if _, err := reportRenderStatus(context.Background(), g.deps("", false), []string{"job-1"}, testGroup); err == nil {
			t.Errorf("a sheet list error must fail closed (error), got nil")
		}
	})

	t.Run("task_outcome_list_error_fails_closed", func(t *testing.T) {
		fr := jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW
		d := &Deps{
			ListJobPhases: func(_ context.Context, _ *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
				return &jobphasepb.ListJobPhasesResponse{Data: []*jobphasepb.JobPhase{phase("jp-1", fr, "")}, Success: true}, nil
			},
			ListJobTasks: func(_ context.Context, _ *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
				return &jobtaskpb.ListJobTasksResponse{Data: []*jobtaskpb.JobTask{{Id: "jt-1", Active: true}}, Success: true}, nil
			},
			ListTaskOutcomes: func(_ context.Context, _ *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
				return nil, errors.New("permission denied: task_outcome:list")
			},
		}
		if _, err := reportRenderStatus(context.Background(), d, []string{"job-1"}, testGroup); err == nil {
			t.Errorf("a task_outcome list error must fail closed (error), got nil")
		}
	})
}

// TestPhaseWorkflowEntered pins the workflow-entered predicate.
func TestPhaseWorkflowEntered(t *testing.T) {
	const (
		unspec = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_UNSPECIFIED
		ip     = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_IN_PROGRESS
		fr     = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW
		pub    = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED
	)
	cases := []struct {
		name string
		p    *jobphasepb.JobPhase
		want bool
	}{
		{"pristine_in_progress", &jobphasepb.JobPhase{ApprovalStatus: ip}, false},
		{"unspecified_null_audit", &jobphasepb.JobPhase{ApprovalStatus: unspec}, false},
		{"for_review", &jobphasepb.JobPhase{ApprovalStatus: fr}, true},
		{"published", &jobphasepb.JobPhase{ApprovalStatus: pub}, true},
		{"in_progress_but_submitted", &jobphasepb.JobPhase{ApprovalStatus: ip, SubmittedBy: strptr("u")}, true},
		{"in_progress_but_returned", &jobphasepb.JobPhase{ApprovalStatus: ip, ReturnedBy: strptr("u")}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := phaseWorkflowEntered(c.p); got != c.want {
				t.Errorf("phaseWorkflowEntered = %v, want %v", got, c.want)
			}
		})
	}
}
