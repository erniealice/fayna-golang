package document

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

func TestDownload_ShowUnpublishedValuesPrintsRecordedLevels(t *testing.T) {
	card := clientPhaseProjectionFixture()
	card.ClientSubscriptionIds = []string{"subscription-1"}
	allowClientProjectionRender(card, "group-1")
	card.RenderGateSheets[2].AllPublished = false
	card.RenderGateSheets[2].AnyWorkflowEntered = true
	card.RenderGateSheets[2].HasData = true
	var rendered map[string]any
	deps := &Deps{
		DocOptions:           outcome_summary.DocumentOptions{ShowUnpublishedValues: true},
		ResolvePrincipalKind: func(context.Context) int32 { return outcome_summary.PrincipalKindStaff },
		GetSubscriptionGroupClientReportCard: func(context.Context, *exportpb.GetSubscriptionGroupClientReportCardRequest) (*exportpb.GetSubscriptionGroupClientReportCardResponse, error) {
			return &exportpb.GetSubscriptionGroupClientReportCardResponse{Success: true, ReportCard: card}, nil
		},
		ResolveTemplateBytes: func(context.Context, string, string) ([]byte, error) { return []byte("template"), nil },
		GeneratePDF: func(_ []byte, data map[string]any) ([]byte, error) {
			rendered = data
			return stubPDFBytes, nil
		},
	}
	r := httptest.NewRequest(http.MethodGet, "/document?period=progress_report&format=pdf", nil)
	r.SetPathValue("id", "group-1")
	r.SetPathValue("client_id", "client-1")
	r = r.WithContext(view.WithUserPermissions(r.Context(), types.NewUserPermissions([]string{"subscription_group_outcome_export:read"})))
	w := httptest.NewRecorder()
	NewDownloadHandler(deps)(w, r)
	if w.Code != http.StatusOK || rendered == nil {
		t.Fatalf("download status=%d body=%s", w.Code, w.Body.String())
	}
	sawLevel := false
	for _, raw := range rendered["jobs"].([]any) {
		for _, a := range raw.(map[string]any)["assessments"].([]any) {
			if a.(map[string]any)["achievement_level"] != "" {
				sawLevel = true
			}
		}
	}
	if !sawLevel {
		t.Fatal("ShowUnpublishedValues: recorded achievement levels must print on an unpublished sheet")
	}
}

func TestCapNames(t *testing.T) {
	for _, tc := range []struct {
		in   string
		max  int
		want string
	}{
		{"Ana A, Ben B, Cy C", 2, "Ana A, Ben B"},
		{"Ana A, Ben B", 2, "Ana A, Ben B"},
		{"Ana A, Ben B, Cy C", 0, "Ana A, Ben B, Cy C"},
		{"", 2, ""},
	} {
		if got := capNames(tc.in, tc.max); got != tc.want {
			t.Errorf("capNames(%q, %d) = %q, want %q", tc.in, tc.max, got, tc.want)
		}
	}
}
