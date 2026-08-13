package section

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	clientattributepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client_attribute"
	bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/subscription_group_document_template"
	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

func exportCell(id string, label *string, score *float64, hasMarks, positive bool) *exportpb.SubscriptionGroupOutcomeCell {
	return &exportpb.SubscriptionGroupOutcomeCell{
		JobTemplateId: id, ScaledLabel: label, ScaledScore: score,
		EnrollmentEvidence: &exportpb.EnrollmentEvidence{HasMarks: hasMarks, HasPositiveMark: positive},
	}
}

func exportString(s string) *string { return &s }

func exportFixture() *exportpb.GetSubscriptionGroupOutcomeExportResponse {
	return &exportpb.GetSubscriptionGroupOutcomeExportResponse{
		Success: true,
		Context: &exportpb.SubscriptionGroupOutcomeExportContext{SubscriptionGroupId: "group-1", SubscriptionGroupName: "Term 1"},
		JobCategories: []*exportpb.JobCategoryOption{{
			JobCategoryId: "cat-a", Code: "academic", Name: "Academic", SortOrder: 1,
			FinalOutcomeAvailable: true,
			JobTemplatePhases:     []*exportpb.JobTemplatePhaseOption{{Code: "q1", Name: "Quarter 1", SequenceOrder: 1}, {Code: "final", Name: "Phase named final", SequenceOrder: 2}},
		}},
		// Deliberately use a stable column order while each row's cells are permuted.
		JobTemplateColumns: []*exportpb.JobTemplateColumn{{JobTemplateId: "job-b", DisplayName: "Math"}, {JobTemplateId: "job-a", DisplayName: "English"}},
		ClientRows: []*exportpb.SubscriptionGroupOutcomeClientRow{{
			ClientId: "client-1", ClientName: "Alpha", ClientFirstName: "Alpha", ClientLastName: "One",
			Cells: []*exportpb.SubscriptionGroupOutcomeCell{
				exportCell("job-a", exportString("0"), nil, true, true),
				exportCell("job-b", exportString("1"), nil, true, false),
			},
		}},
	}
}

func exportDeps(resp *exportpb.GetSubscriptionGroupOutcomeExportResponse) (*Deps, *int) {
	calls := 0
	labels := outcome_summary.DefaultLabels()
	labels.Section.Title = "Report Cards"
	labels.Section.ClientColumn = "Client"
	return &Deps{
		Labels:               labels,
		Options:              outcome_summary.Options{SectionExport: outcome_summary.SectionExportOptions{Enabled: true}},
		ResolvePrincipalKind: func(context.Context) int32 { return outcome_summary.PrincipalKindOperatorOwner },
		GetSubscriptionGroupOutcomeExport: func(context.Context, *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
			calls++
			return resp, nil
		},
	}, &calls
}

// runExport keeps the HTTP contract visible in tests without involving a live mux.
func runExport(deps *Deps, url string, permissions ...string) *httptest.ResponseRecorder {
	for i, permission := range permissions {
		if permission == "job_outcome_summary:list" {
			permissions[i] = "subscription_group_outcome_export:read"
		}
	}
	return runExportExact(deps, url, permissions...)
}

func runExportExact(deps *Deps, url string, permissions ...string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", url, nil)
	r.SetPathValue("id", "group-1")
	ctx := view.WithUserPermissions(r.Context(), types.NewUserPermissions(permissions))
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()
	NewExportHandler(deps)(w, r)
	return w
}

func captureSectionExportLogs(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	oldOut := log.Writer()
	oldFlags := log.Flags()
	oldPrefix := log.Prefix()
	log.SetOutput(&buf)
	log.SetFlags(0)
	log.SetPrefix("")
	defer func() {
		log.SetOutput(oldOut)
		log.SetFlags(oldFlags)
		log.SetPrefix(oldPrefix)
	}()
	fn()
	return buf.String()
}

func TestSectionExport_ExplicitListOnlyDeniedBeforeCompositeRead(t *testing.T) {
	deps, calls := exportDeps(exportFixture())
	w := runExportExact(deps, "/export?format=csv&job_category_id=cat-a&period=final", "job_outcome_summary:list")
	if w.Code != 403 || *calls != 0 {
		t.Fatalf("list-only explicit export status/calls=%d/%d, want 403/0", w.Code, *calls)
	}
}

func TestSectionExport_LegacyNoSelectorRequiresListAndRead(t *testing.T) {
	deps, calls := exportDeps(exportFixture())
	w := runExportExact(deps, "/export", "job_outcome_summary:list")
	if w.Code != 403 || *calls != 0 {
		t.Fatalf("list-only legacy export status/calls=%d/%d, want 403/0", w.Code, *calls)
	}
}

func readCSV(t *testing.T, body string) [][]string {
	t.Helper()
	rows, err := csv.NewReader(strings.NewReader(body)).ReadAll()
	if err != nil {
		t.Fatalf("read csv: %v; body=%q", err, body)
	}
	return rows
}

func TestSectionExport_RejectsTamperedSelectors(t *testing.T) {
	cases := []struct {
		name string
		url  string
		want int
		call bool
	}{
		{"phase final is namespaced", "/export?format=csv&job_category_id=cat-a&period=phase:final", 200, true},
		{"legacy and explicit conflict", "/export?format=csv&job_category_id=cat-a&period=final&jc=legacy", 400, false},
		{"invalid format", "/export?format=x&job_category_id=cat-a&period=final", 400, false},
		{"raw phase", "/export?format=csv&job_category_id=cat-a&period=q1", 400, false},
		{"empty phase", "/export?format=csv&job_category_id=cat-a&period=phase:", 400, false},
		{"invalid phase character", "/export?format=csv&job_category_id=cat-a&period=phase:Q1", 400, false},
		{"unknown category", "/export?format=csv&job_category_id=missing&period=final", 400, true},
		{"phase belongs to another category", "/export?format=csv&job_category_id=cat-a&period=phase:q2", 400, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deps, calls := exportDeps(exportFixture())
			w := runExport(deps, tc.url, "job_outcome_summary:list")
			if w.Code != tc.want {
				t.Fatalf("status=%d want %d body=%q", w.Code, tc.want, w.Body.String())
			}
			if got := *calls; (got > 0) != tc.call {
				t.Fatalf("composite calls=%d, call=%v", got, tc.call)
			}
			if tc.want != 200 && strings.Contains(w.Header().Get("Content-Type"), "text/csv") {
				t.Fatalf("error wrote CSV download headers: %v", w.Header())
			}
		})
	}
}

func TestSectionExport_SelectionFailureLogsStructuredReason(t *testing.T) {
	resp := exportFixture()
	deps, _ := exportDeps(resp)
	logs := captureSectionExportLogs(t, func() {
		w := runExport(deps, "/export?format=pdf&job_category_id=cat-a&period=q1", "job_outcome_summary:list")
		if w.Code != 400 {
			t.Fatalf("status=%d", w.Code)
		}
	})
	if !strings.Contains(logs, "stage=selection") {
		t.Fatalf("selection failure log missing: %q", logs)
	}
	if !strings.Contains(logs, "reason=selection_unoffered") {
		t.Fatalf("selection failure reason missing: %q", logs)
	}
	if strings.Contains(logs, "group-1") || strings.Contains(logs, "Alpha") {
		t.Fatalf("log exposed row identity: %q", logs)
	}
}

func TestSectionExport_ExplicitDisableStopsBeforeCompositeReadRegardlessOfListEntity(t *testing.T) {
	for _, entity := range []string{"", "client", "subscription", outcome_summary.ListEntitySubscriptionGroup} {
		t.Run("entity="+entity, func(t *testing.T) {
			deps, calls := exportDeps(exportFixture())
			deps.Options.List.Entity = entity
			deps.Options.SectionExport.Enabled = false
			w := runExport(deps, "/export?format=csv&job_category_id=cat-a&period=final", "job_outcome_summary:list")
			if w.Code != 404 || *calls != 0 || strings.Contains(w.Header().Get("Content-Type"), "text/csv") {
				t.Fatalf("disabled explicit export status/calls/header=%d/%d/%v, want 404/0/non-CSV", w.Code, *calls, w.Header())
			}
		})
	}
}

func TestSectionExport_MissingForeignAndZeroMatrixStatuses(t *testing.T) {
	cases := []struct {
		name string
		resp *exportpb.GetSubscriptionGroupOutcomeExportResponse
		want int
	}{
		{"missing or foreign context", func() *exportpb.GetSubscriptionGroupOutcomeExportResponse {
			r := exportFixture()
			r.Context = nil
			return r
		}(), 404},
		{"zero columns", func() *exportpb.GetSubscriptionGroupOutcomeExportResponse {
			r := exportFixture()
			r.JobTemplateColumns = nil
			return r
		}(), 404},
		{"zero rows", func() *exportpb.GetSubscriptionGroupOutcomeExportResponse {
			r := exportFixture()
			r.ClientRows = nil
			return r
		}(), 404},
		{"service error", exportFixture(), 503},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deps, calls := exportDeps(tc.resp)
			if tc.name == "service error" {
				deps.GetSubscriptionGroupOutcomeExport = func(context.Context, *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
					return tc.resp, context.Canceled
				}
			}
			w := runExport(deps, "/export?format=csv&job_category_id=cat-a&period=final", "job_outcome_summary:list")
			if w.Code != tc.want {
				t.Fatalf("status=%d want %d body=%q", w.Code, tc.want, w.Body.String())
			}
			if tc.want != 200 && strings.Contains(w.Header().Get("Content-Type"), "text/csv") {
				t.Fatalf("error wrote CSV download headers: %v", w.Header())
			}
			_ = calls
		})
	}
}

func TestSectionExport_PermutedIDsAndEnrollmentRules(t *testing.T) {
	deps, _ := exportDeps(exportFixture())
	w := runExport(deps, "/export?format=csv&job_category_id=cat-a&period=final", "job_outcome_summary:list")
	if w.Code != 200 {
		t.Fatalf("status=%d body=%q", w.Code, w.Body.String())
	}
	rows := readCSV(t, w.Body.String())
	if len(rows) != 3 || rows[0][0] != "Client" || rows[0][1] != "Math" || rows[0][2] != "English" {
		t.Fatalf("header/row count = %#v", rows)
	}
	// job-b is first despite its cell being second in the source row. The
	// positive-evidence zero remains a real grade; the all-zero placeholder is blank.
	if got, want := rows[1], []string{"[1] Alpha", "", "0"}; strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("ID-keyed/evidence row=%#v want %#v", got, want)
	}
}

func TestSectionExport_CorruptMatrixFailsBeforeBytes(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*exportpb.GetSubscriptionGroupOutcomeExportResponse)
	}{
		{"duplicate columns", func(r *exportpb.GetSubscriptionGroupOutcomeExportResponse) {
			r.JobTemplateColumns[1].JobTemplateId = "job-b"
		}},
		{"missing cell", func(r *exportpb.GetSubscriptionGroupOutcomeExportResponse) {
			r.ClientRows[0].Cells = r.ClientRows[0].Cells[:1]
		}},
		{"unknown cell", func(r *exportpb.GetSubscriptionGroupOutcomeExportResponse) {
			r.ClientRows[0].Cells[0].JobTemplateId = "unknown"
		}},
		{"missing evidence", func(r *exportpb.GetSubscriptionGroupOutcomeExportResponse) {
			r.ClientRows[0].Cells[0].EnrollmentEvidence = nil
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := exportFixture()
			tc.mutate(resp)
			deps, _ := exportDeps(resp)
			w := runExport(deps, "/export?format=csv&job_category_id=cat-a&period=final", "job_outcome_summary:list")
			if w.Code != 400 || w.Body.Len() == 0 {
				t.Fatalf("status/body=%d/%q, want 400 with error and no CSV bytes", w.Code, w.Body.String())
			}
			if strings.Contains(w.Header().Get("Content-Type"), "text/csv") {
				t.Fatalf("CSV download content type set on corrupt matrix: %v", w.Header())
			}
		})
	}
}

func TestSectionExport_NormalizeFailureLogsStructuredReason(t *testing.T) {
	resp := exportFixture()
	resp.ClientRows[0].Cells = resp.ClientRows[0].Cells[:1]
	deps, _ := exportDeps(resp)
	logs := captureSectionExportLogs(t, func() {
		w := runExport(deps, "/export?format=csv&job_category_id=cat-a&period=final", "job_outcome_summary:list")
		if w.Code != 400 || w.Body.Len() == 0 {
			t.Fatalf("status/body=%d/%q", w.Code, w.Body.String())
		}
	})
	if !strings.Contains(logs, "stage=normalize") {
		t.Fatalf("normalize failure log missing: %q", logs)
	}
	if !strings.Contains(logs, "reason=invalid_client_rectangle") {
		t.Fatalf("normalize reason missing: %q", logs)
	}
	if !strings.Contains(logs, "expected_slots=2") || !strings.Contains(logs, "actual_slots=1") {
		t.Fatalf("normalize slot diagnostics missing: %q", logs)
	}
	if !strings.Contains(logs, "canonical_order=true") {
		t.Fatalf("canonical order flag missing: %q", logs)
	}
}

func TestSectionExport_RowBandResolutionAndPermissions(t *testing.T) {
	resp := exportFixture()
	deps, _ := exportDeps(resp)
	deps.Options.Row.GroupByField = "client_attributes.gender"
	deps.Options.SectionExport.GroupByAttributeModule = "client"
	definitionCalls, valueCalls := 0, 0
	deps.ListAttributes = func(context.Context, *commonpb.ListAttributesRequest) (*commonpb.ListAttributesResponse, error) {
		definitionCalls++
		return &commonpb.ListAttributesResponse{Success: true, Data: []*commonpb.Attribute{{Id: "attr-gender", Code: "gender", Module: "client", Active: true}}}, nil
	}
	deps.ListClientAttributes = func(context.Context, *clientattributepb.ListClientAttributesRequest) (*clientattributepb.ListClientAttributesResponse, error) {
		valueCalls++
		return &clientattributepb.ListClientAttributesResponse{Success: true, Data: []*clientattributepb.ClientAttribute{{ClientId: "client-1", AttributeId: "attr-gender", Value: "A", Active: true}}}, nil
	}
	w := runExport(deps, "/export?format=csv&job_category_id=cat-a&period=final", "job_outcome_summary:list", "attribute:list", "client_attribute:list")
	if w.Code != 200 || definitionCalls != 1 || valueCalls != 1 {
		t.Fatalf("band status/calls=%d/%d/%d body=%q", w.Code, definitionCalls, valueCalls, w.Body.String())
	}

	// The band permission conjunction is checked before the composite read.
	deps, calls := exportDeps(resp)
	deps.Options.Row.GroupByField = "client_attributes.gender"
	deps.Options.SectionExport.GroupByAttributeModule = "client"
	deps.ListAttributes = func(context.Context, *commonpb.ListAttributesRequest) (*commonpb.ListAttributesResponse, error) {
		t.Fatal("attribute list called while denied")
		return nil, nil
	}
	w = runExport(deps, "/export?format=csv&job_category_id=cat-a&period=final", "job_outcome_summary:list")
	if w.Code != 403 || *calls != 0 {
		t.Fatalf("missing band permission status/composite=%d/%d", w.Code, *calls)
	}
}

func TestSectionExport_BandRosterChunkingAndScopeCorruption(t *testing.T) {
	resp := exportFixture()
	resp.ClientRows = make([]*exportpb.SubscriptionGroupOutcomeClientRow, 101)
	for i := range resp.ClientRows {
		id := "client-" + strings.Repeat("x", i%3) + string(rune('a'+i%26)) + string(rune('0'+i%10))
		resp.ClientRows[i] = &exportpb.SubscriptionGroupOutcomeClientRow{ClientId: id, ClientName: id, Cells: []*exportpb.SubscriptionGroupOutcomeCell{exportCell("job-b", exportString("5"), nil, false, false), exportCell("job-a", exportString("6"), nil, false, false)}}
	}
	deps, _ := exportDeps(resp)
	deps.Options.Row.GroupByField = "client_attributes.gender"
	deps.Options.SectionExport.GroupByAttributeModule = "client"
	definitionCalls, valueCalls, chunkSizes := 0, 0, []int{}
	deps.ListAttributes = func(context.Context, *commonpb.ListAttributesRequest) (*commonpb.ListAttributesResponse, error) {
		definitionCalls++
		return &commonpb.ListAttributesResponse{Success: true, Data: []*commonpb.Attribute{{Id: "attr", Code: "gender", Module: "client", Active: true}}}, nil
	}
	deps.ListClientAttributes = func(_ context.Context, req *clientattributepb.ListClientAttributesRequest) (*clientattributepb.ListClientAttributesResponse, error) {
		valueCalls++
		for _, f := range req.GetFilters().GetFilters() {
			if in := f.GetListFilter(); in != nil {
				chunkSizes = append(chunkSizes, len(in.GetValues()))
			}
		}
		return &clientattributepb.ListClientAttributesResponse{Success: true}, nil
	}
	w := runExport(deps, "/export?format=csv&job_category_id=cat-a&period=final", "job_outcome_summary:list", "attribute:list", "client_attribute:list")
	if w.Code != 200 || definitionCalls != 1 || valueCalls != 2 || len(chunkSizes) != 2 || chunkSizes[0] != 100 || chunkSizes[1] != 1 {
		t.Fatalf("chunk status/calls/chunks=%d/%d/%d/%v", w.Code, definitionCalls, valueCalls, chunkSizes)
	}

	for name, bad := range map[string]*clientattributepb.ClientAttribute{
		"out of scope":    {ClientId: "foreign", AttributeId: "attr", Value: "X", Active: true},
		"wrong attribute": {ClientId: resp.ClientRows[0].ClientId, AttributeId: "other", Value: "X", Active: true},
	} {
		t.Run(name, func(t *testing.T) {
			deps, _ := exportDeps(resp)
			deps.Options.Row.GroupByField = "client_attributes.gender"
			deps.Options.SectionExport.GroupByAttributeModule = "client"
			deps.ListAttributes = func(context.Context, *commonpb.ListAttributesRequest) (*commonpb.ListAttributesResponse, error) {
				return &commonpb.ListAttributesResponse{Success: true, Data: []*commonpb.Attribute{{Id: "attr", Code: "gender", Module: "client", Active: true}}}, nil
			}
			deps.ListClientAttributes = func(context.Context, *clientattributepb.ListClientAttributesRequest) (*clientattributepb.ListClientAttributesResponse, error) {
				return &clientattributepb.ListClientAttributesResponse{Success: true, Data: []*clientattributepb.ClientAttribute{bad}}, nil
			}
			w := runExport(deps, "/export?format=csv&job_category_id=cat-a&period=final", "job_outcome_summary:list", "attribute:list", "client_attribute:list")
			if w.Code != 400 || strings.Contains(w.Header().Get("Content-Type"), "text/csv") {
				t.Fatalf("status/header=%d/%v", w.Code, w.Header())
			}
		})
	}
}

func TestSectionExport_DeterministicBandsSortAndFormulaNeutralization(t *testing.T) {
	resp := exportFixture()
	resp.ClientRows = []*exportpb.SubscriptionGroupOutcomeClientRow{
		{ClientId: "c2", ClientName: "Zed", ClientFirstName: "Zed", ClientLastName: "Z", Cells: []*exportpb.SubscriptionGroupOutcomeCell{exportCell("job-b", exportString("2"), nil, false, false), exportCell("job-a", exportString("=SUM(A1)"), nil, false, false)}},
		{ClientId: "c1", ClientName: "Amy", ClientFirstName: "Amy", ClientLastName: "A", Cells: []*exportpb.SubscriptionGroupOutcomeCell{exportCell("job-b", exportString("3"), nil, false, false), exportCell("job-a", exportString("+1"), nil, false, false)}},
	}
	deps, _ := exportDeps(resp)
	deps.Options.Row.GroupByField = "client_attributes.gender"
	deps.Options.Row.GroupValueOrder = []string{"A", "B"}
	deps.Options.SectionExport.GroupByAttributeModule = "client"
	deps.ListAttributes = func(context.Context, *commonpb.ListAttributesRequest) (*commonpb.ListAttributesResponse, error) {
		return &commonpb.ListAttributesResponse{Success: true, Data: []*commonpb.Attribute{{Id: "attr", Code: "gender", Module: "client", Active: true}}}, nil
	}
	deps.ListClientAttributes = func(_ context.Context, req *clientattributepb.ListClientAttributesRequest) (*clientattributepb.ListClientAttributesResponse, error) {
		var out []*clientattributepb.ClientAttribute
		for _, f := range req.GetFilters().GetFilters() {
			if in := f.GetListFilter(); in != nil {
				for _, id := range in.GetValues() {
					group := "A"
					if id == "c2" {
						group = "B"
					}
					out = append(out, &clientattributepb.ClientAttribute{ClientId: id, AttributeId: "attr", Value: group, Active: true})
				}
			}
		}
		return &clientattributepb.ListClientAttributesResponse{Success: true, Data: out}, nil
	}
	w := runExport(deps, "/export?format=csv&job_category_id=cat-a&period=final", "job_outcome_summary:list", "attribute:list", "client_attribute:list")
	if w.Code != 200 {
		t.Fatalf("status=%d body=%q", w.Code, w.Body.String())
	}
	rows := readCSV(t, w.Body.String())
	if len(rows) != 7 || rows[1][0] != "A" || rows[2][0] != "[1] Amy" || rows[3][0] != "" || rows[4][0] != "B" || rows[5][0] != "[2] Zed" || rows[6][0] != "" {
		t.Fatalf("deterministic rows=%#v", rows)
	}
	if !strings.HasPrefix(rows[2][2], "\t") || !strings.HasPrefix(rows[5][2], "\t") {
		t.Fatalf("formula values not neutralized: %#v", rows)
	}
}

func TestSectionExport_PDFFailLoud(t *testing.T) {
	deps, calls := exportDeps(exportFixture())
	w := runExport(deps, "/export?format=pdf&job_category_id=cat-a&period=final", "job_outcome_summary:list")
	if w.Code != 503 || strings.Contains(w.Header().Get("Content-Type"), "text/csv") || *calls != 1 {
		t.Fatalf("pdf status/header/calls=%d/%v/%d", w.Code, w.Header(), *calls)
	}
}

func TestSectionExport_NoBandRequiresNoAttributeCalls(t *testing.T) {
	deps, _ := exportDeps(exportFixture())
	deps.ListAttributes = func(context.Context, *commonpb.ListAttributesRequest) (*commonpb.ListAttributesResponse, error) {
		t.Fatal("unexpected attribute definition call")
		return nil, nil
	}
	deps.ListClientAttributes = func(context.Context, *clientattributepb.ListClientAttributesRequest) (*clientattributepb.ListClientAttributesResponse, error) {
		t.Fatal("unexpected client attribute call")
		return nil, nil
	}
	w := runExport(deps, "/export?format=csv&job_category_id=cat-a&period=final", "job_outcome_summary:list")
	if w.Code != 200 {
		t.Fatalf("status=%d body=%q", w.Code, w.Body.String())
	}
}

func fullPDFFixture() *exportpb.GetSubscriptionGroupOutcomeExportResponse {
	resp := exportFixture()
	resp.Context.PriceScheduleId = exportString("schedule-1")
	resp.Context.PriceScheduleName = "Term 1"
	resp.Context.PlanId = exportString("plan-1")
	resp.Context.PlanName = "Standard"
	resp.JobTemplateColumns = make([]*exportpb.JobTemplateColumn, 11)
	cells := make([]*exportpb.SubscriptionGroupOutcomeCell, 11)
	for i := 0; i < 11; i++ {
		id := fmt.Sprintf("job-%02d", i+1)
		resp.JobTemplateColumns[i] = &exportpb.JobTemplateColumn{JobTemplateId: id, DisplayName: fmt.Sprintf("Subject %02d", i+1)}
		label := fmt.Sprintf("%d", (i%7)+1)
		cells[10-i] = exportCell(id, &label, nil, false, false)
	}
	resp.ClientRows = []*exportpb.SubscriptionGroupOutcomeClientRow{{ClientId: "client-1", ClientName: "Alpha", Cells: cells}}
	return resp
}

func canonicalPDFDocx(t *testing.T) []byte {
	t.Helper()
	docx, err := os.ReadFile("../section_document/subscription-group-outcome-matrix-single-period-11-v1.docx")
	if err != nil {
		t.Fatalf("read generated DOCX: %v", err)
	}
	return docx
}

func stalePDFDocx(t *testing.T, docx []byte) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(docx), int64(len(docx)))
	if err != nil {
		t.Fatalf("open generated DOCX: %v", err)
	}
	var out bytes.Buffer
	writer := zip.NewWriter(&out)
	for _, file := range reader.File {
		body, readErr := io.ReadAll(func() io.Reader {
			r, e := file.Open()
			if e != nil {
				t.Fatalf("open DOCX part: %v", e)
			}
			return r
		}())
		if readErr != nil {
			t.Fatalf("read DOCX part: %v", readErr)
		}
		if file.Name == "word/document.xml" {
			body = bytes.Replace(body, []byte("{{sheet_title}}"), []byte("stale"), 1)
		}
		header := file.FileHeader
		entry, createErr := writer.CreateHeader(&header)
		if createErr != nil {
			t.Fatalf("create DOCX part: %v", createErr)
		}
		if _, writeErr := entry.Write(body); writeErr != nil {
			t.Fatalf("write DOCX part: %v", writeErr)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close stale DOCX: %v", err)
	}
	return out.Bytes()
}

func configurePDFDeps(t *testing.T, resp *exportpb.GetSubscriptionGroupOutcomeExportResponse) (*Deps, *int, *int, **exportpb.ResolveSubscriptionGroupOutcomeDocumentForRenderRequest) {
	t.Helper()
	deps, _ := exportDeps(resp)
	deps.Options.SectionExport.ProfileByCategoryCode = map[string]bindingpb.RenderProfile{
		"academic": bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1,
	}
	resolverCalls, engineCalls := 0, 0
	seen := new(*exportpb.ResolveSubscriptionGroupOutcomeDocumentForRenderRequest)
	deps.ResolveSectionTemplate = func(_ context.Context, req *exportpb.ResolveSubscriptionGroupOutcomeDocumentForRenderRequest) (*outcome_summary.ResolvedSectionTemplate, error) {
		resolverCalls++
		copy := *req
		*seen = &copy
		return &outcome_summary.ResolvedSectionTemplate{Bytes: canonicalPDFDocx(t), RenderProfile: req.GetRenderProfile(), JobCategoryID: req.GetJobCategoryId()}, nil
	}
	deps.GeneratePDF = func(template []byte, data map[string]any) ([]byte, error) {
		engineCalls++
		if len(template) == 0 || len(data) == 0 {
			t.Fatalf("PDF engine received empty template/data")
		}
		return []byte("%PDF-1.7\nsynthetic"), nil
	}
	return deps, &resolverCalls, &engineCalls, seen
}

func TestSectionExport_PDFContentAndPeriod(t *testing.T) {
	resp := fullPDFFixture()
	deps, resolverCalls, engineCalls, seen := configurePDFDeps(t, resp)
	w := runExport(deps, "/export?format=pdf&job_category_id=cat-a&period=final", "job_outcome_summary:list")
	if w.Code != 200 || !bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF-")) {
		t.Fatalf("status/body=%d/%q", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Content-Type"); got != "application/pdf" {
		t.Fatalf("content type=%q", got)
	}
	if !strings.Contains(w.Header().Get("Content-Disposition"), "report-cards-term-1-final.pdf") {
		t.Fatalf("content disposition=%q", w.Header().Get("Content-Disposition"))
	}
	if *resolverCalls != 1 || *engineCalls != 1 || seen == nil || *seen == nil {
		t.Fatalf("resolver/engine/request=%d/%d/%v", *resolverCalls, *engineCalls, seen)
	}
	profile := bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1
	if (*seen).SubscriptionGroupId != "group-1" || (*seen).JobCategoryId != "cat-a" || (*seen).RenderProfile != profile || (*seen).GetExpectedPlanId() != "plan-1" || (*seen).GetExpectedPriceScheduleId() != "schedule-1" {
		t.Fatalf("resolver request=%+v", **seen)
	}
}

func TestSectionExport_PDFResolverFailureStopsBeforeEngine(t *testing.T) {
	deps, resolverCalls, engineCalls, _ := configurePDFDeps(t, fullPDFFixture())
	wantErr := errors.New("bounded storage read failed")
	deps.ResolveSectionTemplate = func(context.Context, *exportpb.ResolveSubscriptionGroupOutcomeDocumentForRenderRequest) (*outcome_summary.ResolvedSectionTemplate, error) {
		*resolverCalls++
		return nil, wantErr
	}

	w := runExport(deps, "/export?format=pdf&job_category_id=cat-a&period=final", "job_outcome_summary:list")
	if w.Code != 503 || strings.Contains(w.Header().Get("Content-Type"), "application/pdf") || w.Body.Len() == 0 {
		t.Fatalf("status/header/body=%d/%v/%q", w.Code, w.Header(), w.Body.String())
	}
	if *resolverCalls != 1 || *engineCalls != 0 {
		t.Fatalf("resolver/engine calls=%d/%d, want 1/0", *resolverCalls, *engineCalls)
	}
}

func TestSectionExport_PDFRejectsProfileCapacityCategoryAndManifestBeforeEngine(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*exportpb.GetSubscriptionGroupOutcomeExportResponse, *Deps)
	}{
		{"capacity mismatch", func(resp *exportpb.GetSubscriptionGroupOutcomeExportResponse, _ *Deps) {
			resp.JobTemplateColumns = resp.JobTemplateColumns[:10]
		}},
		{"noncanonical column order", func(resp *exportpb.GetSubscriptionGroupOutcomeExportResponse, _ *Deps) {
			resp.JobTemplateColumns[0].DisplayName, resp.JobTemplateColumns[10].DisplayName = resp.JobTemplateColumns[10].DisplayName, resp.JobTemplateColumns[0].DisplayName
		}},
		{"unmapped category", func(resp *exportpb.GetSubscriptionGroupOutcomeExportResponse, _ *Deps) {
			resp.JobCategories[0].Code = "other"
		}},
		{"stale manifest", func(*exportpb.GetSubscriptionGroupOutcomeExportResponse, *Deps) {}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := fullPDFFixture()
			deps, resolverCalls, engineCalls, _ := configurePDFDeps(t, resp)
			tc.mutate(resp, deps)
			if tc.name == "stale manifest" {
				deps.ResolveSectionTemplate = func(context.Context, *exportpb.ResolveSubscriptionGroupOutcomeDocumentForRenderRequest) (*outcome_summary.ResolvedSectionTemplate, error) {
					*resolverCalls++
					return &outcome_summary.ResolvedSectionTemplate{Bytes: stalePDFDocx(t, canonicalPDFDocx(t)), RenderProfile: bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1, JobCategoryID: "cat-a"}, nil
				}
			}
			w := runExport(deps, "/export?format=pdf&job_category_id=cat-a&period=final", "job_outcome_summary:list")
			want := 400
			if tc.name == "stale manifest" {
				want = 503
			}
			if w.Code != want || strings.Contains(w.Header().Get("Content-Type"), "application/pdf") || w.Body.Len() == 0 {
				t.Fatalf("status/header/body=%d/%v/%q", w.Code, w.Header(), w.Body.String())
			}
			if tc.name == "capacity mismatch" || tc.name == "noncanonical column order" || tc.name == "unmapped category" {
				if *resolverCalls != 0 || *engineCalls != 0 {
					t.Fatalf("calls=%d/%d, want zero", *resolverCalls, *engineCalls)
				}
			} else if *resolverCalls != 1 || *engineCalls != 0 {
				t.Fatalf("stale manifest calls=%d/%d, want 1/0", *resolverCalls, *engineCalls)
			}
		})
	}
}

func TestSectionExport_ProfileFailureLogsStructuredReason(t *testing.T) {
	resp := exportFixture()
	resp.JobCategories[0].Code = "other"
	deps, _ := exportDeps(resp)
	logs := captureSectionExportLogs(t, func() {
		w := runExport(deps, "/export?format=pdf&job_category_id=cat-a&period=final", "job_outcome_summary:list")
		if w.Code != 400 || strings.Contains(w.Header().Get("Content-Type"), "text/csv") {
			t.Fatalf("status/header=%d/%v", w.Code, w.Header())
		}
	})
	if !strings.Contains(logs, "stage=profile") {
		t.Fatalf("profile failure log missing: %q", logs)
	}
	if !strings.Contains(logs, "reason=missing_profile") {
		t.Fatalf("profile reason missing: %q", logs)
	}
}

type sectionStyleMarker struct{}

func (sectionStyleMarker) Error() string            { return "style contract" }
func (sectionStyleMarker) StyleContractError() bool { return true }

type sectionLibreMarker struct{}

func (sectionLibreMarker) Error() string                { return "libreoffice unavailable" }
func (sectionLibreMarker) LibreOfficeUnavailable() bool { return true }

func TestSectionExport_PDFGeneratorErrorsFailLoud(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"style contract", sectionStyleMarker{}, 503},
		{"libreoffice unavailable", sectionLibreMarker{}, 503},
		{"unexpected engine error", errors.New("engine exploded"), 500},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := fullPDFFixture()
			deps, resolverCalls, engineCalls, _ := configurePDFDeps(t, resp)
			deps.GeneratePDF = func([]byte, map[string]any) ([]byte, error) { *engineCalls++; return nil, tc.err }
			w := runExport(deps, "/export?format=pdf&job_category_id=cat-a&period=final", "job_outcome_summary:list")
			if w.Code != tc.want || strings.Contains(w.Header().Get("Content-Type"), "application/pdf") || *resolverCalls != 1 || w.Body.Len() == 0 {
				t.Fatalf("status/header/resolver/body=%d/%v/%d/%q", w.Code, w.Header(), *resolverCalls, w.Body.String())
			}
			_ = engineCalls
		})
	}

	deps, resolverCalls, engineCalls, _ := configurePDFDeps(t, fullPDFFixture())
	deps.GeneratePDF = func([]byte, map[string]any) ([]byte, error) { *engineCalls++; return []byte("not a PDF"), nil }
	w := runExport(deps, "/export?format=pdf&job_category_id=cat-a&period=final", "job_outcome_summary:list")
	if w.Code != 500 || strings.Contains(w.Header().Get("Content-Type"), "application/pdf") || *resolverCalls != 1 || *engineCalls != 1 {
		t.Fatalf("malformed PDF status/header/calls=%d/%v/%d/%d", w.Code, w.Header(), *resolverCalls, *engineCalls)
	}
}
