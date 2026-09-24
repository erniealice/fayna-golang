// outcome-matrix-return.js — CSP-safe reason dialog for approval-bar returns.
// The page template owns the per-phase copy and form association; this file
// moves that declarative fragment into pyeza's shared dialog shell and guards
// its host-form request lifecycle.
(function () {
    'use strict';

    window.lf = window.lf || {};
    window.lf.fayna = window.lf.fayna || {};
    if (window.lf.fayna.outcomeMatrixReturnBound) return;
    window.lf.fayna.outcomeMatrixReturnBound = true;

    function dialogAPI() {
        if (window.lf.ui && window.lf.ui.Dialog) return window.lf.ui.Dialog;
        if (window.lf.Dialog) return window.lf.Dialog;
        return null;
    }

    function dialogShell() {
        var overlay = document.querySelector('[data-dialog-overlay]');
        if (!overlay) return null;
        var container = overlay.querySelector('[data-dialog-container]');
        if (!container) return null;
        return { overlay: overlay, container: container };
    }

    function clear(node) {
        while (node.firstChild) node.removeChild(node.firstChild);
    }

    function submitForm(form, submitter) {
        if (!form) return;
        if (typeof form.requestSubmit === 'function') {
            if (submitter) {
                form.requestSubmit(submitter);
            } else {
                form.requestSubmit();
            }
            return;
        }
        form.submit();
    }

    function returnForm(trigger) {
        if (!trigger) return null;
        var formID = trigger.getAttribute('data-return-form');
        return formID ? document.getElementById(formID) : null;
    }

    function returnTrigger(form) {
        if (!form) return null;
        if (form._returnTrigger) return form._returnTrigger;

        var triggers = document.querySelectorAll('[data-return-reason-open]');
        for (var i = 0; i < triggers.length; i++) {
            if (triggers[i].getAttribute('data-return-form') === form.id) {
                return triggers[i];
            }
        }
        return null;
    }

    function beginReturnRequest(form, trigger, confirm) {
        if (!form || form.dataset.returnInflight === 'true') return false;

        var activeTrigger = trigger || returnTrigger(form);
        form._returnTrigger = activeTrigger;
        form._returnConfirm = confirm || null;
        form.dataset.returnInflight = 'true';

        if (activeTrigger) {
            activeTrigger.disabled = true;
            activeTrigger.setAttribute('aria-busy', 'true');
        }
        if (confirm) {
            confirm.disabled = true;
            confirm.dataset.returnSubmitInflight = 'true';
        }
        return true;
    }

    function finishReturnRequest(form, successful) {
        if (!form) return;

        delete form.dataset.returnInflight;

        var trigger = returnTrigger(form);
        if (trigger) {
            trigger.disabled = false;
            trigger.removeAttribute('aria-busy');
        }

        var confirm = form._returnConfirm;
        if (confirm) {
            confirm.disabled = false;
            delete confirm.dataset.returnSubmitInflight;
        }

        if (successful) {
            var api = dialogAPI();
            if (api && typeof api.close === 'function') api.close();
        }
    }

    function nativeFallback(trigger, fragment) {
        var messageNode = fragment.querySelector('[data-return-reason-message]');
        var labelNode = fragment.querySelector('label[for]');
        var message = messageNode ? messageNode.textContent : '';
        if (!window.confirm(message)) return;

        var label = labelNode ? labelNode.textContent : '';
        var reason = window.prompt(label, '');
        if (reason === null) return;

        var form = returnForm(trigger);
        if (!form) return;
        var input = document.createElement('input');
        input.type = 'hidden';
        input.name = 'reason';
        input.value = reason;
        form.appendChild(input);
        if (!beginReturnRequest(form, trigger, null)) {
            input.remove();
            return;
        }
        try {
            submitForm(form, null);
        } catch (err) {
            input.remove();
            finishReturnRequest(form, false);
            return;
        }
        window.setTimeout(function () { input.remove(); }, 0);
    }

    function openReasonDialog(trigger) {
        var form = returnForm(trigger);
        if (!form || trigger.disabled || form.dataset.returnInflight === 'true') return;

        var templateID = trigger.getAttribute('data-return-reason-dialog');
        var template = templateID ? document.getElementById(templateID) : null;
        if (!template || !template.content) return;

        form._returnTrigger = trigger;

        var shell = dialogShell();
        var api = dialogAPI();
        if (!shell || !api || typeof api.open !== 'function') {
            nativeFallback(trigger, template.content);
            return;
        }

        if (!shell.overlay.hidden && shell.overlay.classList.contains('visible')) return;
        clear(shell.container);
        shell.container.appendChild(template.content.cloneNode(true));
        api.open();
    }

    document.addEventListener('htmx:afterRequest', function (event) {
        var detail = event.detail || {};
        var form = detail.elt || event.target;
        if (form && form.tagName !== 'FORM' && form.closest) form = form.closest('form');
        if ((!form || form.tagName !== 'FORM') && event.target && event.target.closest) {
            form = event.target.closest('form');
        }
        if (!form || form.tagName !== 'FORM') return;
        if (detail.elt && detail.elt !== form) return;
        if (form.dataset.returnInflight !== 'true') return;

        // HTMX 1.9 handles an HX-Location response before it sets detail.successful,
        // so fall back to the HTTP status (review wave-3 #2).
        var xhr = detail.xhr;
        var ok = detail.successful === true ||
            (detail.successful === undefined && !!xhr && xhr.status >= 200 && xhr.status < 300);
        finishReturnRequest(form, ok);
    });

    // Mark that HTMX actually started the request, so a submission HTMX never
    // issues (validation/cancel before beforeRequest) cannot leave the host form
    // stuck in-flight (review wave-3 #3).
    document.addEventListener('htmx:beforeRequest', function (event) {
        var detail = event.detail || {};
        var form = detail.elt || event.target;
        if (form && form.tagName === 'FORM' && form.dataset.returnInflight === 'true') {
            form.dataset.returnSent = 'true';
        }
    });

    document.addEventListener('click', function (event) {
        var target = event.target;
        var trigger = target && target.closest ? target.closest('[data-return-reason-open]') : null;
        if (!trigger) return;
        event.preventDefault();
        openReasonDialog(trigger);
    });

    document.addEventListener('click', function (event) {
        var target = event.target;
        var confirm = target && target.closest ? target.closest('[data-return-reason-confirm]') : null;
        if (!confirm) return;

        var formID = confirm.getAttribute('form');
        var form = formID ? document.getElementById(formID) : null;
        if (!form) return;

        event.preventDefault();
        if (confirm.disabled || form.dataset.returnInflight === 'true') return;
        if (!beginReturnRequest(form, null, confirm)) return;

        delete form.dataset.returnSent;
        try {
            submitForm(form, confirm);
        } catch (err) {
            finishReturnRequest(form, false);
            return;
        }
        // HTMX issues the request synchronously from the submit event; if it did
        // not start (no htmx:beforeRequest), release the in-flight state now.
        if (form.dataset.returnSent !== 'true') finishReturnRequest(form, false);
    });
}());
