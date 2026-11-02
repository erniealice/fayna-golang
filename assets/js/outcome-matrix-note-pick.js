// outcome-matrix-note-pick.js — keeps a chosen wording from being lost on Save.
// In the narrative drawer, a radio only takes effect through "Use this wording".
// While a radio differs from how the drawer opened, the primary Save is disabled
// and a hint explains why. Going back to the original radio, editing the textarea
// (an explicit typed note wins), or "Keep my note" in the hint enables Save again; the
// last is the only way back from a hand-edited note, which opens with no radio chosen,
// because a radio cannot be unchosen by clicking it. The POST is untouched.
// Without this script Save stays enabled and the hint stays hidden.
(function () {
    'use strict';

    window.lf = window.lf || {};
    window.lf.fayna = window.lf.fayna || {};
    if (window.lf.fayna.outcomeMatrixNotePickBound) return;
    window.lf.fayna.outcomeMatrixNotePickBound = true;

    var RADIO = 'input[type="radio"][name="rating_description_set_entry_id"]';

    function pickForm(el) {
        return el && el.closest ? el.closest('form[data-om-note-pick]') : null;
    }

    function setBlocked(form, blocked) {
        var save = form.querySelector('[data-testid="om-note-save"]');
        var hint = form.querySelector('[data-testid="om-note-pick-hint"]');
        if (save) save.disabled = blocked;
        if (hint) hint.hidden = !blocked;
    }

    document.addEventListener('change', function (event) {
        var target = event.target;
        if (!target || !target.matches || !target.matches(RADIO)) return;
        var form = pickForm(target);
        if (!form) return;
        // defaultChecked is the state the server rendered, so the drawer's opening state.
        setBlocked(form, !target.defaultChecked);
    });

    document.addEventListener('click', function (event) {
        var target = event.target;
        var keep = target && target.closest ? target.closest('[data-om-note-keep]') : null;
        var form = pickForm(keep);
        if (!form) return;
        form.querySelectorAll(RADIO).forEach(function (radio) {
            radio.checked = radio.defaultChecked;
        });
        setBlocked(form, false);
        // The button just hid with its hint, so focus goes where the note is edited.
        var note = form.querySelector('textarea[name="note"]');
        if (note) note.focus();
    });

    document.addEventListener('input', function (event) {
        var target = event.target;
        if (!target || !target.matches || !target.matches('textarea[name="note"]')) return;
        var form = pickForm(target);
        if (form) setBlocked(form, false);
    });
}());
