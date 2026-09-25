package action

import (
	"testing"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"
)

// Q26 (schema-proposal §9.1; codex-review-impl1 #1/#7/#8) — conditional
// grade-sheet writes, duplicate-resolver fail-closed, translated rejections.

func resolvedLevels() *matrixpb.ResolveCellRatingDescriptionsResponse {
	return ratingResp(enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_RESOLVED, levelDescs())
}

// The update is written through UpdateTaskOutcomeIfUnchanged with the
// snapshot this request read (value, note, date_modified) as the expectation.
func TestQ26_Update_UsesConditionalWriteWithReadSnapshot(t *testing.T) {
	mod := int64(1727250000123)
	r := &recorder{readNote: "staff note", readNumeric: f64(3), readModified: &mod, wireRating: true, ratingResp: resolvedLevels()}
	items := cells(t, invoke(t, r.deps(descMatrix(true)), "save_mode=cell&cells."+existingID+"=5", allPerms))
	if len(items) != 1 || !items[0].OK {
		t.Fatalf("update failed: %+v", items)
	}
	if r.condUpdateCalls != 1 || r.plainUpdateCalls != 0 {
		t.Fatalf("want the conditional write only, got cond=%d plain=%d", r.condUpdateCalls, r.plainUpdateCalls)
	}
	e := r.lastExpected
	if e == nil || e.NumericValue == nil || *e.NumericValue != 3 ||
		e.DeterminationNote == nil || *e.DeterminationNote != "staff note" ||
		e.DateModified == nil || *e.DateModified != mod {
		t.Fatalf("expected snapshot = %+v, want value 3 / note %q / modified %d", e, "staff note", mod)
	}
}

// CONFLICT: another save landed between this request's read and its write.
// The cell is rejected with the translated cell_changed_retry message and
// nothing is written (value + note stay as the other writer left them).
func TestQ26_Update_ConflictRejectsAndLeavesValueAndNote(t *testing.T) {
	r := &recorder{readNote: "level-7 wording from the other save", readNumeric: f64(7), wireRating: true, ratingResp: resolvedLevels(), updateConflict: true}
	items := cells(t, invoke(t, r.deps(descMatrix(true)), "save_mode=cell&cells."+existingID+"=4", allPerms))
	got := items[0]
	if got.OK {
		t.Fatalf("a conflicting update must be rejected: %+v", got)
	}
	if want := outcome_matrix.DefaultLabels().Errors.CellChangedRetry; got.Error != want {
		t.Fatalf("error = %q, want the translated cell_changed_retry %q", got.Error, want)
	}
	if r.updateCalls != 0 || r.lastUpdate != nil || r.plainUpdateCalls != 0 {
		t.Fatalf("a conflict must write nothing (updateCalls=%d plain=%d last=%v)", r.updateCalls, r.plainUpdateCalls, r.lastUpdate)
	}
	if got.Value != "" {
		t.Fatalf("a rejected cell must not report a new saved value, got %q", got.Value)
	}
}

// Non-description cells also use the conditional write when wired.
func TestQ26_Update_NonDescriptionCellAlsoConditional(t *testing.T) {
	r := &recorder{readNumeric: f64(80), updateConflict: true}
	items := cells(t, invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells."+existingID+"=86", allPerms))
	if items[0].OK || r.plainUpdateCalls != 0 || r.condUpdateCalls != 1 {
		t.Fatalf("non-description conflict must reject via the conditional path: %+v cond=%d plain=%d", items[0], r.condUpdateCalls, r.plainUpdateCalls)
	}
}

// Two concurrent new-cell requests: the first creates, the second — which
// also saw an empty cell in its matrix — gets CONFLICT instead of inserting a
// duplicate outcome.
func TestQ26_Create_RaceSecondGetsConflict(t *testing.T) {
	r := &recorder{wireRating: true, ratingResp: resolvedLevels()}
	form := "save_mode=cell&new." + jobTaskID + ":" + criteriaID + "=4"
	first := cells(t, invoke(t, r.deps(descMatrix(false)), form, allPerms))
	second := cells(t, invoke(t, r.deps(descMatrix(false)), form, allPerms))
	if !first[0].OK || first[0].OutcomeID == "" {
		t.Fatalf("first create must succeed with an id: %+v", first[0])
	}
	if second[0].OK {
		t.Fatalf("second concurrent create must be rejected: %+v", second[0])
	}
	if want := outcome_matrix.DefaultLabels().Errors.CellChangedRetry; second[0].Error != want {
		t.Fatalf("error = %q, want %q", second[0].Error, want)
	}
	if r.createCalls != 1 || r.condCreateCalls != 2 || r.plainCreateCalls != 0 {
		t.Fatalf("want exactly one insert via the conditional path, got creates=%d cond=%d plain=%d", r.createCalls, r.condCreateCalls, r.plainCreateCalls)
	}
}

// Unwired conditional writes: a description-mode cell fails closed (its note
// decision depends on the snapshot), while a plain cell keeps the pre-Q26
// unconditional path.
func TestQ26_UnwiredConditional_DescriptionFailsClosed_PlainFallsBack(t *testing.T) {
	r := &recorder{plainOnly: true, readNumeric: f64(3), wireRating: true, ratingResp: resolvedLevels()}
	items := cells(t, invoke(t, r.deps(descMatrix(true)), "save_mode=cell&cells."+existingID+"=5", allPerms))
	if items[0].OK || r.updateCalls != 0 {
		t.Fatalf("description-mode update without the conditional write must reject: %+v (updates=%d)", items[0], r.updateCalls)
	}
	items = cells(t, invoke(t, r.deps(descMatrix(false)), "save_mode=cell&new."+jobTaskID+":"+criteriaID+"=5", allPerms))
	if items[0].OK || r.createCalls != 0 {
		t.Fatalf("description-mode create without the conditional write must reject: %+v (creates=%d)", items[0], r.createCalls)
	}

	r = &recorder{plainOnly: true}
	items = cells(t, invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells."+existingID+"=86", allPerms))
	if !items[0].OK || r.plainUpdateCalls != 1 {
		t.Fatalf("plain cell must keep the unconditional update: %+v plain=%d", items[0], r.plainUpdateCalls)
	}
}

// codex-review-impl1 #7: the resolver returning the same requested address
// twice rejects that cell (fail closed) — never "first result wins".
func TestRatingResolution_DuplicateResultRejectsCell(t *testing.T) {
	cell := &matrixpb.CellRatingRef{JobId: jobID, JobTaskId: jobTaskID, OutcomeCriteriaId: criteriaID}
	for _, tc := range []struct {
		name   string
		second enums.RatingDescriptionResolutionStatus
	}{
		{"resolved then invalid_config", enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_INVALID_CONFIG},
		{"resolved twice", enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_RESOLVED},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &matrixpb.ResolveCellRatingDescriptionsResponse{Success: true, Results: []*matrixpb.CellRatingResolution{
				{Cell: cell, Status: enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_RESOLVED, Descriptions: levelDescs()},
				{Cell: cell, Status: tc.second, Descriptions: levelDescs()},
			}}
			r := &recorder{readNumeric: f64(3), wireRating: true, ratingResp: resp}
			items := cells(t, invoke(t, r.deps(descMatrix(true)), "save_mode=cell&cells."+existingID+"=4", allPerms))
			if items[0].OK || r.updateCalls != 0 || r.condUpdateCalls != 0 {
				t.Fatalf("duplicate resolver result must reject before any write: %+v (updates=%d cond=%d)", items[0], r.updateCalls, r.condUpdateCalls)
			}
			if want := outcome_matrix.DefaultLabels().Errors.RatingDescriptionResolutionFailed; items[0].Error != want {
				t.Fatalf("error = %q, want %q", items[0].Error, want)
			}
		})
	}
}

// codex-review-impl1 #8: rejection codes are shown as the translated lyngua
// message; with no labels loaded the bounded code is kept (never blank).
func TestRejectionCodes_AreTranslated(t *testing.T) {
	labels := outcome_matrix.DefaultLabels().Errors
	for _, tc := range []struct {
		status enums.RatingDescriptionResolutionStatus
		want   string
		code   string
	}{
		{enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_UNRESOLVED_IDENTITY, labels.RatingDescriptionUnresolvedIdentity, "rating_description_unresolved_identity"},
		{enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_AMBIGUOUS, labels.RatingDescriptionAmbiguous, "rating_description_ambiguous"},
		{enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_INVALID_CONFIG, labels.RatingDescriptionInvalidConfig, "rating_description_invalid_config"},
	} {
		r := &recorder{readNumeric: f64(3), wireRating: true, ratingResp: ratingResp(tc.status, nil)}
		items := cells(t, invoke(t, r.deps(descMatrix(true)), "save_mode=cell&cells."+existingID+"=5", allPerms))
		if items[0].OK || items[0].Error != tc.want || tc.want == "" {
			t.Fatalf("%v: error = %q, want translated %q", tc.status, items[0].Error, tc.want)
		}

		d := r.deps(descMatrix(true))
		d.Labels = outcome_matrix.Labels{}
		items = cells(t, invoke(t, d, "save_mode=cell&cells."+existingID+"=5", allPerms))
		if items[0].Error != tc.code {
			t.Fatalf("%v with no labels: error = %q, want bounded code %q", tc.status, items[0].Error, tc.code)
		}
	}
}
