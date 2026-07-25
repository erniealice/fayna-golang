package list

import (
	"bytes"
	"html/template"
	"strconv"
	"strings"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"
	"github.com/erniealice/pyeza-golang/types"

	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// head_summary.go — builds the frozen corner header's caption stack.
//
// pyeza's CellGridConfig exposes RowHeadHTML as an opaque slot: the component
// renders the fragment and knows nothing about students, sections or gender.
// All of that vocabulary lives HERE, in the consumer, which is what keeps the
// grid component generic.
//
// ESCAPING. RowHeadHTML is template.HTML, so it bypasses Go's contextual
// escaping. The counts are integers, but the BREAKDOWN LABELS are workspace
// data — an attribute value's own label, which a workspace can rename to
// anything. They are therefore rendered through html/template (which escapes
// each {{.}}) rather than concatenated into a string. Never build this fragment
// with fmt.Sprintf or string +.

var headStackTmpl = template.Must(template.New("cg-headstack").Parse(
	`<div class="cell-grid-headstack">` +
		`<span class="cell-grid-headstack-caption">{{.Caption}}</span>` +
		`{{if .Total}}<span class="cell-grid-headline">` +
		`<span class="badge badge-neutral badge-sm">{{.Total}}</span>` +
		`</span>{{end}}` +
		`{{if .Breakdown}}<span class="cell-grid-headline cell-grid-headstack-breakdown">{{.Breakdown}}</span>{{end}}` +
		`</div>`))

type headStackData struct {
	Caption   string
	Total     string
	Breakdown string
}

// One L1 (phase) header cell: its title, then its state badge on the same line.
var phaseLabelTmpl = template.Must(template.New("cg-phaselabel").Parse(
	`<span class="cell-grid-headinline">` +
		`<span>{{.Label}}</span>` +
		`<span class="badge badge-{{.Variant}} badge-sm">{{.Badge}}</span>` +
		`</span>`))

type phaseLabelData struct {
	Label   string
	Badge   string
	Variant string
}

// phaseChip is one phase's derived approval state, as the approval band already
// renders it. Kept here so the header cell and the band cannot disagree.
type phaseChip struct {
	label   string
	variant string
}

// phaseChips derives each phase's badge from the response roll-ups, reusing the
// SAME two functions the approval band uses (approvalStatusLabel /
// approvalChipVariant) so the header cell and the band can never drift apart or
// mint a second colour vocabulary.
//
// Note the roll-ups are derived over the whole template — section_id does not
// narrow them. That is fine for a STATE badge (a phase is in progress or not,
// and that is a property of the phase), and it is exactly why the band's
// numeric count is suppressed on a section page while this badge is not.
func phaseChips(l outcome_matrix.ApprovalLabels, resp *matrixpb.GetOutcomeMatrixResponse) map[string]phaseChip {
	rollups := resp.GetApprovalRollups()
	if len(rollups) == 0 {
		return nil
	}
	out := make(map[string]phaseChip, len(rollups))
	for _, ru := range rollups {
		id := ru.GetJobTemplatePhaseId()
		if id == "" {
			continue
		}
		out[id] = phaseChip{
			label:   approvalStatusLabel(l, ru.GetStatus()),
			variant: approvalChipVariant(ru.GetStatus()),
		}
	}
	return out
}

// PhaseActions is the payload for the L1 header cell's action slot. It is
// EXPORTED because the slot template dereferences its fields; pyeza hands it
// through as an opaque `any` and never reads it.
//
// It carries the whole ApprovalPhase rather than a trimmed copy so the header
// controls and the band render from one source — a divergence here would mean a
// button appearing in one place and not the other for the same state.
type PhaseActions struct {
	Phase  ApprovalPhase
	Labels outcome_matrix.ApprovalLabels

	// Download* wire the per-column CSV export affordance (20260725): an
	// icon-only GET link to the sheet export carrying this column's period
	// token (a phase's code, or the reserved "final" on the trailing composite
	// column). Stamped by the view AUGMENTATION (ratings.go), not by
	// phaseActions below — the export base depends on the route scope (group
	// vs template), which only NewView resolves. A GET link needs no form, so
	// the nested-form trap the approval buttons dodge does not apply here.
	DownloadURL    string
	DownloadAria   string // pre-composed accessible name (icon-only control)
	DownloadTestID string
}

// Any reports whether this phase has at least one available control. Templates
// use it to skip the wrapper entirely: a rendered-but-empty slot would still
// consume the header cell's space-between gap.
func (p PhaseActions) Any() bool {
	return p.Phase.CanSubmit || p.Phase.CanVerify || p.Phase.CanPublish || p.Phase.CanReturn ||
		p.DownloadURL != ""
}

// phaseActions indexes the already-derived approval bar by phase id, so each L1
// header cell can carry its OWN controls. Returns nil when nothing is
// actionable, which leaves every header cell's Actions nil and the slot unused.
func phaseActions(bar []ApprovalPhase, l outcome_matrix.Labels, _ string) map[string]any {
	if len(bar) == 0 {
		return nil
	}
	out := make(map[string]any, len(bar))
	for _, ph := range bar {
		if ph.PhaseID == "" {
			continue
		}
		pa := PhaseActions{Phase: ph, Labels: l.Approval}
		if !pa.Any() {
			continue // no controls ⇒ no slot for this column
		}
		out[ph.PhaseID] = pa
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// buildPhaseLabelHTML renders "Semester 1 (Music) [IN PROGRESS]" for an L1
// header cell. Returns "" when the phase has no derived state, which makes
// pyeza fall back to the plain Label — so a consumer with no state, or a phase
// missing from the roll-up, renders exactly as before.
func buildPhaseLabelHTML(label string, chip phaseChip) template.HTML {
	if label == "" || chip.label == "" || chip.variant == "" {
		return ""
	}
	var buf bytes.Buffer
	if err := phaseLabelTmpl.Execute(&buf, phaseLabelData{
		Label:   label,
		Badge:   chip.label,
		Variant: chip.variant,
	}); err != nil {
		return ""
	}
	return template.HTML(buf.String())
}

// headBucket is one grouping value and how many rows carry it.
type headBucket struct {
	label string // the attribute value's own label, or the "unassigned" wording
	count int
}

// buildRowHeadHTML renders the caption stack for the frozen corner cell:
//
//	Student
//	[ 29 students ]
//	11 male, 15 female, 3 not assigned
//
// Returns "" when there is nothing to add beyond the plain caption, which makes
// pyeza fall back to rendering Labels.ClientColumn exactly as before.
func buildRowHeadHTML(l types.CellGridLabels, total int, buckets []headBucket) template.HTML {
	if total <= 0 {
		return ""
	}

	data := headStackData{Caption: l.ClientColumn}
	if l.ClientTotal != "" {
		data.Total = strings.Replace(l.ClientTotal, "{count}", strconv.Itoa(total), 1)
	}

	// Only render a breakdown that says something the total does not. One
	// bucket holding every row is just the total again.
	if len(buckets) > 1 {
		parts := make([]string, 0, len(buckets))
		for _, b := range buckets {
			if b.count == 0 || b.label == "" {
				continue
			}
			parts = append(parts, strconv.Itoa(b.count)+" "+b.label)
		}
		if len(parts) > 1 {
			data.Breakdown = strings.Join(parts, ", ")
		}
	}

	var buf bytes.Buffer
	if err := headStackTmpl.Execute(&buf, data); err != nil {
		// Fail to the plain caption rather than a half-written fragment.
		return ""
	}
	return template.HTML(buf.String())
}

// groupBuckets counts rows per grouping value, preserving the order the rows are
// already banded in, so the breakdown reads in the same order as the bands the
// operator sees. Rows with no value collapse into one bucket labelled with the
// "unassigned" wording, and it always sorts last (matching the band order).
func groupBuckets(rows []types.CellGridRow, valueOf func(rowID string) string, unassignedLabel string) []headBucket {
	order := make([]string, 0, 4)
	counts := map[string]int{}
	unassigned := 0

	for _, r := range rows {
		v := strings.TrimSpace(valueOf(r.ID))
		if v == "" {
			unassigned++
			continue
		}
		if _, seen := counts[v]; !seen {
			order = append(order, v)
		}
		counts[v]++
	}

	out := make([]headBucket, 0, len(order)+1)
	for _, v := range order {
		// Rendered VERBATIM. The value is DATA (attribute_value.label) — no
		// lyngua key per value (a workspace may add or rename one at any time)
		// and no Go-side casing either: the band rows below present the same
		// values, so any casing decision belongs in CSS where both can share
		// it, not duplicated here.
		out = append(out, headBucket{label: v, count: counts[v]})
	}
	if unassigned > 0 && unassignedLabel != "" {
		out = append(out, headBucket{label: unassignedLabel, count: unassigned})
	}
	return out
}
