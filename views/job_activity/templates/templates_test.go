// Package templates holds drift-prevention tests for the proto-enum select
// option lists rendered directly in the job-activity drawer template.
//
// Background — 2026-04-30 enum-select-canonicalize plan §2 (Wave 2):
// Hand-rolled per-handler option builders for proto enums repeatedly drifted
// behind the proto. Wave 1 (centymo BillingKind/AmountBasis/BillingTreatment,
// commit ccf8d57) moved the option lists into the drawer templates and added
// per-template drift tests. Wave 2 extends the same pattern to fayna's
// job_activity drawer for BillableStatus, the only Wave 2 enum that surfaces
// as a <select> widget. (BillingEventStatus, BillingEventTrigger, and
// JobPhaseStatus are display/badge or button-action only — there is no
// <select> to canonicalize, so they are out of scope per plan §2.)
//
// Strategy:
// The drawer form template is now the source of truth — the option values
// are hardcoded as <option value="..."> tags in the HTML. The drift test
// below:
//   1. Reads the template HTML via os.ReadFile.
//   2. Extracts the <option value="..."> values for the named <select>.
//   3. Asserts the set equals the proto enum's _name map (minus UNSPECIFIED).
//
// When someone adds a proto enum value (e.g. BILLABLE_STATUS_WRITE_OFF was
// missing from the template before this commit), the matching test fails
// until the template is updated. No html parser dependency — the templates
// are controlled and a stdlib regex suffices.
package templates_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
)

// TestBillableStatusOptionsMatchProto guards the BillableStatus <option> list
// in drawer-form.html against drift from jobactivitypb.BillableStatus_name.
func TestBillableStatusOptionsMatchProto(t *testing.T) {
	t.Parallel()
	body := readTemplate(t, "drawer-form.html")
	got := extractEnumOptionValues(t, body, "billable_status")
	want := protoEnumNames(jobactivitypb.BillableStatus_name)
	assertSameSet(t, "billable_status", want, got)
}

// readTemplate loads a template file from the same directory as this test.
// The test binary's CWD is the package directory at test time, so a relative
// path is sufficient and avoids the embed-FS dance.
func readTemplate(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(".", name)
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(body)
}

// selectBlock extracts the <select ... name="X" ...> ... </select> block.
// We slice on the first matching select so the test is robust to other
// surrounding form-group contents (info popovers, hint spans, etc.).
func selectBlock(body, name string) string {
	openRe := regexp.MustCompile(`(?s)<select[^>]*\bname="` + regexp.QuoteMeta(name) + `"[^>]*>`)
	open := openRe.FindStringIndex(body)
	if open == nil {
		return ""
	}
	rest := body[open[1]:]
	closeIdx := strings.Index(rest, "</select>")
	if closeIdx < 0 {
		return ""
	}
	return rest[:closeIdx]
}

// extractEnumOptionValues pulls the <option value="..."> values out of the
// named <select>. Empty values (placeholder rows) are skipped.
func extractEnumOptionValues(t *testing.T, body, name string) []string {
	t.Helper()
	block := selectBlock(body, name)
	if block == "" {
		t.Fatalf("could not locate <select name=%q> in template", name)
	}
	optRe := regexp.MustCompile(`<option\s+value="([^"]*)"`)
	matches := optRe.FindAllStringSubmatch(block, -1)
	values := make([]string, 0, len(matches))
	for _, m := range matches {
		v := m[1]
		if v == "" {
			continue // placeholder row, not a real enum value
		}
		values = append(values, v)
	}
	sort.Strings(values)
	return values
}

// protoEnumNames returns the proto enum string names from a generated
// _name map (e.g. BillableStatus_name), filtering out the zero-valued
// *_UNSPECIFIED sentinel that should never reach the UI.
func protoEnumNames(m map[int32]string) []string {
	out := make([]string, 0, len(m))
	for _, name := range m {
		if strings.HasSuffix(name, "_UNSPECIFIED") {
			continue
		}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// assertSameSet fails the test with a clear diff when the template's option
// values don't match the proto's enum names.
func assertSameSet(t *testing.T, field string, want, got []string) {
	t.Helper()
	if equal(want, got) {
		return
	}
	missing := diff(want, got)
	extra := diff(got, want)
	t.Fatalf("\n%s option drift detected (template vs. proto):\n"+
		"  proto _name: %v\n"+
		"  template:    %v\n"+
		"  missing in template: %v\n"+
		"  extra in template:   %v\n"+
		"  fix: update the matching template (or proto), then re-run.",
		field, want, got, missing, extra)
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// diff returns elements in a that are not in b.
func diff(a, b []string) []string {
	bset := make(map[string]struct{}, len(b))
	for _, v := range b {
		bset[v] = struct{}{}
	}
	out := []string{}
	for _, v := range a {
		if _, ok := bset[v]; !ok {
			out = append(out, v)
		}
	}
	return out
}
