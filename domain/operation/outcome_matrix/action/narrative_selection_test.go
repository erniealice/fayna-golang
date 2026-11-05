package action

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
	"github.com/erniealice/pyeza-golang/view"
)

const (
	selStatusResolved = enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_RESOLVED
)

// selEntry builds one exact-map band.
func selEntry(id, match, text string) *matrixpb.RatingDescription {
	return &matrixpb.RatingDescription{ScaleKind: enums.ScaleKind_SCALE_KIND_EXACT_MAP, InputMatch: &match, Description: text, RatingDescriptionSetEntryId: &id}
}

// selOption is one set: bands 3 and 4 for the cell's criterion, no band 0.
func selOption(set, label string, isDefault bool) *matrixpb.CellRatingSelectionOption {
	return &matrixpb.CellRatingSelectionOption{
		RatingDescriptionSetId: set, OptionLabel: label, IsDefault: isDefault,
		Descriptions: []*matrixpb.RatingDescription{
			selEntry(set+"-e3", "3", set+" wording for three"),
			selEntry(set+"-e4", "4", set+" wording for four"),
		},
	}
}

// selectionRec is a descriptor-capable double around a description-mode cell.
type selectionRec struct {
	narrativeRec
	numeric        *float64
	note           string
	owner          string
	status         enums.RatingDescriptionResolutionStatus
	options        []*matrixpb.CellRatingSelectionOption
	resolveErr     error
	conflict       bool
	writeErr       error
	unwiredWrite   bool
	unwiredResolve bool
	flagSeen       []bool
	writes         []*taskoutcomepb.TaskOutcome
	expected       *taskoutcomepb.TaskOutcome
	plainUpdates   int
	// optionsByCriteria answers for other criteria, as the resolver would.
	optionsByCriteria map[string][]*matrixpb.CellRatingSelectionOption
	writeAttempts     int
}

func newSelectionRec(value float64) *selectionRec {
	return &selectionRec{numeric: &value, note: "set-a wording for three", status: selStatusResolved,
		options: []*matrixpb.CellRatingSelectionOption{selOption("set-a", "1", true), selOption("set-b", "3", false), selOption("set-c", "5", false)}}
}

func descriptionMatrix() *matrixpb.GetOutcomeMatrixResponse {
	m := numericMatrix(true)
	crit := m.GetPhases()[0].GetTasks()[0].GetCriteria()[0]
	crit.RatingMode = enums.RatingMode_RATING_MODE_NUMERIC_WITH_DESCRIPTION
	return m
}

func (r *selectionRec) drawerDeps() *NarrativeDeps {
	d := r.narrativeRec.deps(descriptionMatrix())
	owner := r.owner
	if owner == "" {
		owner = staffID
	}
	d.ReadTaskOutcome = func(_ context.Context, req *taskoutcomepb.ReadTaskOutcomeRequest) (*taskoutcomepb.ReadTaskOutcomeResponse, error) {
		modified := int64(1234)
		note := r.note
		return &taskoutcomepb.ReadTaskOutcomeResponse{Data: []*taskoutcomepb.TaskOutcome{{
			Id: req.GetData().GetId(), RecordedBy: owner, NumericValue: r.numeric, DeterminationNote: &note, DateModified: &modified,
		}}}, nil
	}
	d.UpdateTaskOutcome = func(context.Context, *taskoutcomepb.UpdateTaskOutcomeRequest) (*taskoutcomepb.UpdateTaskOutcomeResponse, error) {
		r.plainUpdates++
		return &taskoutcomepb.UpdateTaskOutcomeResponse{}, nil
	}
	if !r.unwiredResolve {
		d.ResolveCellRatingDescriptions = func(_ context.Context, req *matrixpb.ResolveCellRatingDescriptionsRequest) (*matrixpb.ResolveCellRatingDescriptionsResponse, error) {
			r.flagSeen = append(r.flagSeen, req.GetIncludeSelectionOptions())
			if r.resolveErr != nil {
				return nil, r.resolveErr
			}
			var defaults []*matrixpb.RatingDescription
			for _, o := range r.options {
				if o.GetIsDefault() {
					defaults = o.GetDescriptions()
				}
			}
			return &matrixpb.ResolveCellRatingDescriptionsResponse{Success: true, Results: []*matrixpb.CellRatingResolution{{
				Cell: req.GetCells()[0], Status: r.status, Descriptions: defaults, SelectionOptions: r.optionsFor(req.GetCells()[0].GetOutcomeCriteriaId()),
			}}}, nil
		}
	}
	if !r.unwiredWrite {
		d.UpdateTaskOutcomeIfUnchanged = func(_ context.Context, req *taskoutcomepb.UpdateTaskOutcomeRequest, expected *taskoutcomepb.TaskOutcome) (*taskoutcomepb.UpdateTaskOutcomeResponse, bool, error) {
			r.writeAttempts++
			if r.writeErr != nil {
				return nil, false, r.writeErr
			}
			if r.conflict {
				return nil, true, nil
			}
			r.writes = append(r.writes, req.GetData())
			r.expected = expected
			return &taskoutcomepb.UpdateTaskOutcomeResponse{}, false, nil
		}
	}
	return d
}

func (r *selectionRec) optionsFor(criteriaID string) []*matrixpb.CellRatingSelectionOption {
	if opts, ok := r.optionsByCriteria[criteriaID]; ok {
		return opts
	}
	return r.options
}

func (r *selectionRec) get(t *testing.T) *NarrativeDrawerData {
	t.Helper()
	return narrativeData(t, narrativeInvoke(t, r.drawerDeps(), http.MethodGet, "outcome_id="+existingID, "", narrPerms))
}

func (r *selectionRec) pick(t *testing.T, entryID string, extra ...string) view.ViewResult {
	t.Helper()
	form := url.Values{"outcome_id": {existingID}, "use_descriptor": {"1"}, "rating_description_set_entry_id": {entryID}}
	for i := 0; i+1 < len(extra); i += 2 {
		form.Set(extra[i], extra[i+1])
	}
	return narrativeInvoke(t, r.drawerDeps(), http.MethodPost, "", form.Encode(), narrPerms)
}

func isSuccess(res view.ViewResult) bool {
	return res.StatusCode == http.StatusOK && strings.Contains(res.Headers["HX-Trigger"], "formSuccess")
}

func TestNarrativeSelection_GET_ShowsSameRatingAcrossSets(t *testing.T) {
	r := newSelectionRec(3)
	data := r.get(t)
	if len(data.DescriptorOptions) != 3 {
		t.Fatalf("options = %d, want 3", len(data.DescriptorOptions))
	}
	wantValue := []string{"set-a-e3", "set-b-e3", "set-c-e3"}
	wantText := []string{"set-a wording for three", "set-b wording for three", "set-c wording for three"}
	wantLabel := []string{"Option 1 (default)", "Option 3", "Option 5"}
	for i, o := range data.DescriptorOptions {
		if o.Value != wantValue[i] || o.Description != wantText[i] || o.Label != wantLabel[i] {
			t.Errorf("option %d = %+v", i, o)
		}
	}
	if data.SelectedEntryID != "set-a-e3" {
		t.Errorf("selected = %q", data.SelectedEntryID)
	}
	if len(r.flagSeen) != 1 || !r.flagSeen[0] {
		t.Errorf("drawer GET must ask for options once, flags = %v", r.flagSeen)
	}

	r = newSelectionRec(4)
	r.note = "set-b wording for four"
	data = r.get(t)
	if data.DescriptorOptions[0].Value != "set-a-e4" || data.SelectedEntryID != "set-b-e4" {
		t.Errorf("rating 4: %+v selected %q", data.DescriptorOptions, data.SelectedEntryID)
	}

	// Identical text stays as separate options, and the default is preferred for pre-selection.
	r = newSelectionRec(3)
	for _, o := range r.options {
		o.GetDescriptions()[0].Description = "same text"
	}
	r.note = "same text"
	data = r.get(t)
	if len(data.DescriptorOptions) != 3 || data.SelectedEntryID != "set-a-e3" {
		t.Errorf("identical text: %d options, selected %q", len(data.DescriptorOptions), data.SelectedEntryID)
	}

	// A hand-written note selects nothing.
	r = newSelectionRec(3)
	r.note = "typed by hand"
	if data = r.get(t); data.SelectedEntryID != "" {
		t.Errorf("free text selected %q", data.SelectedEntryID)
	}
}

func TestNarrativeSelection_GET_NoOptions_NoList(t *testing.T) {
	cases := map[string]func(r *selectionRec){
		"no selectable rows": func(r *selectionRec) { r.options = []*matrixpb.CellRatingSelectionOption{} },
		"one option":         func(r *selectionRec) { r.options = r.options[:1] },
		"resolver error":     func(r *selectionRec) { r.resolveErr = errors.New("boom") },
		"resolver unwired":   func(r *selectionRec) { r.unwiredResolve = true },
		"only one matches": func(r *selectionRec) {
			r.options[1].Descriptions = nil
			r.options[2].Descriptions = nil
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			r := newSelectionRec(3)
			mutate(r)
			data := r.get(t)
			if len(data.DescriptorOptions) != 0 {
				t.Errorf("options = %d, want none", len(data.DescriptorOptions))
			}
			if !data.Editable || data.Note != r.note {
				t.Errorf("drawer must still render as today: %+v", data)
			}
		})
	}

	t.Run("column not in description mode", func(t *testing.T) {
		r := newSelectionRec(3)
		d := r.drawerDeps()
		d.GetOutcomeMatrix = func(context.Context, *matrixpb.GetOutcomeMatrixRequest) (*matrixpb.GetOutcomeMatrixResponse, error) {
			return numericMatrix(true), nil
		}
		data := narrativeData(t, narrativeInvoke(t, d, http.MethodGet, "outcome_id="+existingID, "", narrPerms))
		if len(data.DescriptorOptions) != 0 || len(r.flagSeen) != 0 {
			t.Errorf("non-description column: options=%d resolver calls=%d", len(data.DescriptorOptions), len(r.flagSeen))
		}
	})

	// The drawer's script blocks Save while a chosen wording is waiting, unless a
	// typed note overrides it. Without JS, or with a typed note, a plain save stores
	// the textarea and the radio value is deliberately ignored.
	t.Run("plain save stores the typed note and ignores the radio", func(t *testing.T) {
		r := newSelectionRec(3)
		res := narrativeInvoke(t, r.drawerDeps(), http.MethodPost, "", url.Values{"outcome_id": {existingID}, "note": {"typed"}, "rating_description_set_entry_id": {"set-b-e3"}}.Encode(), narrPerms)
		if !isSuccess(res) || r.plainUpdates != 1 || len(r.writes) != 0 {
			t.Errorf("plain save: success=%v plain=%d picks=%d", isSuccess(res), r.plainUpdates, len(r.writes))
		}
	})
}

func TestNarrativeSelection_FrozenOrForeign_Denied(t *testing.T) {
	frozen := func(t *testing.T) *selectionRec { return newSelectionRec(3) }

	t.Run("read-only cell has no list and refuses a pick", func(t *testing.T) {
		r := frozen(t)
		d := r.drawerDeps()
		d.GetOutcomeMatrix = func(context.Context, *matrixpb.GetOutcomeMatrixRequest) (*matrixpb.GetOutcomeMatrixResponse, error) {
			m := descriptionMatrix()
			for _, row := range m.GetRows() {
				for _, c := range row.GetCells() {
					c.Editable = false
				}
			}
			return m, nil
		}
		data := narrativeData(t, narrativeInvoke(t, d, http.MethodGet, "outcome_id="+existingID, "", narrPerms))
		if data.Editable || len(data.DescriptorOptions) != 0 || len(r.flagSeen) != 0 {
			t.Errorf("read-only: editable=%v options=%d resolver=%d", data.Editable, len(data.DescriptorOptions), len(r.flagSeen))
		}
		res := narrativeInvoke(t, d, http.MethodPost, "", url.Values{"outcome_id": {existingID}, "use_descriptor": {"1"}, "rating_description_set_entry_id": {"set-b-e3"}}.Encode(), narrPerms)
		if isSuccess(res) || len(r.writes) != 0 {
			t.Errorf("read-only pick succeeded")
		}
	})

	t.Run("another grader's row", func(t *testing.T) {
		r := frozen(t)
		r.owner = "someone-else"
		if res := r.pick(t, "set-b-e3"); isSuccess(res) || len(r.writes) != 0 {
			t.Errorf("foreign pick succeeded")
		}
	})

	t.Run("without update permission", func(t *testing.T) {
		r := frozen(t)
		data := narrativeData(t, narrativeInvoke(t, r.drawerDeps(), http.MethodGet, "outcome_id="+existingID, "", []string{"task_outcome:read"}))
		if data.Editable || len(data.DescriptorOptions) != 0 {
			t.Errorf("read permission only: %+v", data)
		}
		res := narrativeInvoke(t, r.drawerDeps(), http.MethodPost, "", url.Values{"outcome_id": {existingID}, "use_descriptor": {"1"}, "rating_description_set_entry_id": {"set-b-e3"}}.Encode(), []string{"task_outcome:read"})
		if isSuccess(res) || len(r.writes) != 0 {
			t.Errorf("pick without update permission succeeded")
		}
	})
}

func TestNarrativeSelection_POST_WritesRenderedEntry_IgnoresNote(t *testing.T) {
	r := newSelectionRec(3)
	res := r.pick(t, "set-c-e3", "note", "forged text", "clear", "1")
	if !isSuccess(res) {
		t.Fatalf("pick failed: %+v", res)
	}
	if len(r.writes) != 1 || r.writes[0].GetDeterminationNote() != "set-c wording for three" {
		t.Fatalf("writes = %+v", r.writes)
	}
	if r.plainUpdates != 0 {
		t.Errorf("the plain update path must not run")
	}
}

func TestNarrativeSelection_POST_ForgedEntry_Denied(t *testing.T) {
	for name, id := range map[string]string{
		"other rating": "set-b-e4",
		"junk":         "' OR 1=1 --",
	} {
		t.Run(name, func(t *testing.T) {
			r := newSelectionRec(3)
			res := r.pick(t, id)
			if isSuccess(res) || len(r.writes) != 0 {
				t.Errorf("forged id %q accepted", id)
			}
		})
	}

	// A real entry of the same set and rating, on a different criterion.
	t.Run("other criterion", func(t *testing.T) {
		r := newSelectionRec(3)
		other := []*matrixpb.CellRatingSelectionOption{selOption("set-b-other-criterion", "3", false)}
		r.optionsByCriteria = map[string][]*matrixpb.CellRatingSelectionOption{"another-criterion": other}
		id := other[0].GetDescriptions()[0].GetRatingDescriptionSetEntryId()
		if got := r.optionsFor("another-criterion")[0].GetDescriptions()[0].GetRatingDescriptionSetEntryId(); got != id || id == "" {
			t.Fatalf("precondition: the id must exist on the other criterion, got %q", got)
		}
		if res := r.pick(t, id); isSuccess(res) || len(r.writes) != 0 || r.writeAttempts != 0 {
			t.Errorf("other-criterion entry %q accepted", id)
		}
	})

	// A real entry for this criterion and rating, in a set with no selectable row.
	t.Run("non-selectable set", func(t *testing.T) {
		r := newSelectionRec(3)
		notListed := selOption("set-z", "9", false)
		id := notListed.GetDescriptions()[0].GetRatingDescriptionSetEntryId()
		if id == "" || notListed.GetDescriptions()[0].GetInputMatch() != "3" {
			t.Fatalf("precondition: the id must be a rating-3 entry, got %q", id)
		}
		if res := r.pick(t, id); isSuccess(res) || len(r.writes) != 0 || r.writeAttempts != 0 {
			t.Errorf("non-selectable entry %q accepted", id)
		}
	})
}

func TestNarrativeSelection_POST_NoRadioChosen_AsksForAWording(t *testing.T) {
	r := newSelectionRec(3)
	res := r.pick(t, "")
	if isSuccess(res) || r.writeAttempts != 0 {
		t.Fatalf("empty pick must write nothing: attempts=%d", r.writeAttempts)
	}
	want := r.drawerDeps().Labels.Errors.DescriptorChooseWording
	if want == "" || res.Headers["HX-Error-Message"] != want {
		t.Errorf("message = %q, want %q", res.Headers["HX-Error-Message"], want)
	}
	if res.Headers["HX-Error-Message"] == r.drawerDeps().Labels.Errors.DescriptorNotAvailable {
		t.Errorf("must not reuse the not-available message")
	}
}

func TestNarrativeSelection_POST_NoRadioChosen_ForeignRowStillDenied(t *testing.T) {
	r := newSelectionRec(3)
	r.owner = "someone-else"
	res := r.pick(t, "")
	if isSuccess(res) || res.Headers["HX-Error-Message"] == r.drawerDeps().Labels.Errors.DescriptorChooseWording {
		t.Errorf("a foreign row must get the denial, got %q", res.Headers["HX-Error-Message"])
	}
}

func TestNarrativeSelection_POST_OptionsGone_Denied(t *testing.T) {
	r := newSelectionRec(3)
	if len(r.get(t).DescriptorOptions) != 3 {
		t.Fatalf("precondition")
	}
	r.options = nil
	if res := r.pick(t, "set-b-e3"); isSuccess(res) || len(r.writes) != 0 {
		t.Errorf("pick accepted after rows removed")
	}
}

func TestNarrativeSelection_POST_NoEntryValue_Denied(t *testing.T) {
	r := newSelectionRec(0)
	if data := r.get(t); len(data.DescriptorOptions) != 0 {
		t.Errorf("value 0 has a list: %d", len(data.DescriptorOptions))
	}
	if res := r.pick(t, "set-b-e3"); isSuccess(res) || len(r.writes) != 0 {
		t.Errorf("value 0 pick accepted")
	}

	r = newSelectionRec(3)
	r.numeric = nil
	if data := r.get(t); len(data.DescriptorOptions) != 0 {
		t.Errorf("null value has a list")
	}
	if res := r.pick(t, "set-b-e3"); isSuccess(res) || len(r.writes) != 0 {
		t.Errorf("null value pick accepted")
	}
}

func TestNarrativeSelection_POST_StaleOrConflict_NoOverwrite(t *testing.T) {
	t.Run("rating changed since the drawer opened", func(t *testing.T) {
		r := newSelectionRec(3)
		if len(r.get(t).DescriptorOptions) != 3 {
			t.Fatalf("precondition")
		}
		four := float64(4)
		r.numeric = &four
		if res := r.pick(t, "set-b-e3"); isSuccess(res) || len(r.writes) != 0 {
			t.Errorf("stale entry accepted")
		}
	})
	t.Run("conflict", func(t *testing.T) {
		r := newSelectionRec(3)
		r.conflict = true
		res := r.pick(t, "set-b-e3")
		if isSuccess(res) {
			t.Errorf("conflict reported success")
		}
		// The conditional write was reached and answered conflict; nothing fell back.
		if r.writeAttempts != 1 || r.plainUpdates != 0 {
			t.Errorf("conflict: attempts=%d plain=%d, want 1 and 0", r.writeAttempts, r.plainUpdates)
		}
		if want := r.drawerDeps().Labels.Errors.CellChangedRetry; res.Headers["HX-Error-Message"] != want {
			t.Errorf("conflict message = %q, want %q", res.Headers["HX-Error-Message"], want)
		}
	})
	t.Run("write error", func(t *testing.T) {
		r := newSelectionRec(3)
		r.writeErr = errors.New("boom")
		if res := r.pick(t, "set-b-e3"); isSuccess(res) {
			t.Errorf("write error reported success")
		}
	})
	t.Run("conditional write unwired", func(t *testing.T) {
		r := newSelectionRec(3)
		r.unwiredWrite = true
		if res := r.pick(t, "set-b-e3"); isSuccess(res) || r.plainUpdates != 0 {
			t.Errorf("unwired write must deny and not fall back: plain=%d", r.plainUpdates)
		}
	})
}

func TestNarrativeSelection_ResolverReject_FailsClosed(t *testing.T) {
	for _, status := range []enums.RatingDescriptionResolutionStatus{
		enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_UNRESOLVED_IDENTITY,
		enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_AMBIGUOUS,
		enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_INVALID_CONFIG,
		enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_PLACEHOLDER_UNRESOLVED,
		enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_NO_LINK,
	} {
		t.Run(status.String(), func(t *testing.T) {
			r := newSelectionRec(3)
			r.status = status
			if data := r.get(t); len(data.DescriptorOptions) != 0 {
				t.Errorf("list shown for %v", status)
			}
			if res := r.pick(t, "set-b-e3"); isSuccess(res) || len(r.writes) != 0 {
				t.Errorf("pick accepted for %v", status)
			}
		})
	}
	t.Run("resolver error", func(t *testing.T) {
		r := newSelectionRec(3)
		r.resolveErr = errors.New("boom")
		if res := r.pick(t, "set-b-e3"); isSuccess(res) || len(r.writes) != 0 {
			t.Errorf("pick accepted on resolver error")
		}
	})
}

func TestNarrativeSelection_POST_NoteOnlyWrite(t *testing.T) {
	r := newSelectionRec(3)
	if res := r.pick(t, "set-b-e3"); !isSuccess(res) {
		t.Fatalf("pick failed")
	}
	got := r.writes[0]
	if got.GetId() != existingID || got.GetDeterminationNote() == "" {
		t.Errorf("id or note missing: %+v", got)
	}
	if got.NumericValue != nil || got.TextValue != nil || got.CategoricalValue != nil || got.PassFailValue != nil ||
		got.GetCriteriaType() != enums.CriteriaType_CRITERIA_TYPE_UNSPECIFIED || got.GetCriteriaVersionId() != "" {
		t.Errorf("request carries more than id and note: %+v", got)
	}
	if r.expected == nil || r.expected.GetNumericValue() != 3 || r.expected.GetDateModified() != 1234 {
		t.Errorf("snapshot must be the row just read: %+v", r.expected)
	}
}

// The save path must never ask for options.
func TestNarrativeSelection_SavePathNeverSetsFlag(t *testing.T) {
	var flag *bool
	fn := func(_ context.Context, req *matrixpb.ResolveCellRatingDescriptionsRequest) (*matrixpb.ResolveCellRatingDescriptionsResponse, error) {
		v := req.GetIncludeSelectionOptions()
		flag = &v
		return &matrixpb.ResolveCellRatingDescriptionsResponse{Success: true, Results: []*matrixpb.CellRatingResolution{{Cell: req.GetCells()[0], Status: selStatusResolved}}}, nil
	}
	resolveRatingDescriptions(context.Background(), fn, []srvCell{{descriptionMode: true, jobID: "j", jobTaskID: "t", criteriaID: "c"}})
	if flag == nil || *flag {
		t.Fatalf("flag = %v, want set false", flag)
	}
}

// ── review edit on the descriptor pick (plan 20261105-reviewer-edit-grades, AC-7) ──

func (r *selectionRec) pickWithPerms(t *testing.T, perms []string, entryID string) view.ViewResult {
	t.Helper()
	form := url.Values{"outcome_id": {existingID}, "use_descriptor": {"1"}, "rating_description_set_entry_id": {entryID}}
	return narrativeInvoke(t, r.drawerDeps(), http.MethodPost, "", form.Encode(), perms)
}

// AC-7: non-recorder + permission ⇒ the pick is written, note column only,
// ownership columns untouched.
func TestNarrativeSelection_ReviewEdit_NonRecorderAllowed(t *testing.T) {
	r := newSelectionRec(3)
	r.owner = "someone-else"
	res := r.pickWithPerms(t, narrReviewEditPerms, "set-c-e3")
	if !isSuccess(res) {
		t.Fatalf("review edit pick must succeed: %+v", res)
	}
	if len(r.writes) != 1 {
		t.Fatalf("want 1 conditional write, got %d", len(r.writes))
	}
	w := r.writes[0]
	if w.GetDeterminationNote() != "set-c wording for three" {
		t.Errorf("note = %q", w.GetDeterminationNote())
	}
	if w.GetRecordedBy() != "" || w.ReviewedBy != nil || w.ReviewedDate != nil {
		t.Errorf("review edit must not set ownership/review columns: %+v", w)
	}
}

// AC-7: non-recorder without the permission ⇒ blocked.
func TestNarrativeSelection_ReviewEdit_NonRecorderWithoutPermBlocked(t *testing.T) {
	r := newSelectionRec(3)
	r.owner = "someone-else"
	res := r.pickWithPerms(t, narrPerms, "set-c-e3")
	if isSuccess(res) {
		t.Fatalf("non-recorder without review edit must be blocked")
	}
	if r.writeAttempts != 0 {
		t.Errorf("blocked pick must not attempt a write, got %d", r.writeAttempts)
	}
}

// AC-7: the recorder is unaffected either way.
func TestNarrativeSelection_ReviewEdit_RecorderUnaffected(t *testing.T) {
	for name, perms := range map[string][]string{"without": narrPerms, "with": narrReviewEditPerms} {
		t.Run(name, func(t *testing.T) {
			r := newSelectionRec(3)
			if res := r.pickWithPerms(t, perms, "set-c-e3"); !isSuccess(res) {
				t.Fatalf("recorder pick must succeed: %+v", res)
			}
			if len(r.writes) != 1 {
				t.Errorf("want 1 write, got %d", len(r.writes))
			}
		})
	}
}
