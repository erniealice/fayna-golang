package dashboard

import (
	"testing"

	home "github.com/erniealice/fayna-golang/domain/home"
)

func TestSectionRefreshUsesExplicitSectionRoute(t *testing.T) {
	m := sectionModule(testResponse(), nil, "/home/overview")
	cases := []struct {
		section home.SectionKey
		wantKey string
	}{
		{home.SectionOverview, "home.overview_url"},
		{home.SectionPulse, "home.pulse_url"},
		{home.SectionPerformance, "home.performance_url"},
		{home.SectionAttention, "home.attention_url"},
	}
	for _, tc := range cases {
		t.Run(string(tc.section), func(t *testing.T) {
			data := m.buildSectionData(withPerms(ocRead), secViewCtx("/home/"+string(tc.section)), tc.section)
			if data.RefreshRouteKey != tc.wantKey {
				t.Fatalf("refresh route key = %q, want %q", data.RefreshRouteKey, tc.wantKey)
			}
		})
	}
}
