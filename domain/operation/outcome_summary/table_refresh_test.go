package outcome_summary

import (
	"net/http/httptest"
	"testing"
)

func TestIsTableRefresh(t *testing.T) {
	cases := []struct {
		name, hxRequest, hxTarget string
		want                      bool
	}{
		{"card target", "true", "outcome-summary-templates-table-card", true},
		{"table target", "true", "outcome-summary-templates-table", true},
		{"hash-prefixed target", "true", "#outcome-summary-templates-table-card", true},
		{"other target", "true", "sheetContent", false},
		{"full navigation", "", "", false},
		{"htmx without target", "true", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/report-cards/templates", nil)
			if tc.hxRequest != "" {
				r.Header.Set("HX-Request", tc.hxRequest)
			}
			if tc.hxTarget != "" {
				r.Header.Set("HX-Target", tc.hxTarget)
			}
			if got := IsTableRefresh(r, "outcome-summary-templates-table"); got != tc.want {
				t.Fatalf("IsTableRefresh = %v, want %v", got, tc.want)
			}
		})
	}
	if IsTableRefresh(nil, "x") {
		t.Fatal("nil request must not be a table refresh")
	}
}
