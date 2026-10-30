package client_document_batch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/render"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/subscription_group_document_template"
	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

const testGroupID = "group-1"

// fakeExport records every scoped export read and answers options mode
// (no category) with categories and category mode with member rows.
type fakeExport struct {
	calls      []*exportpb.GetSubscriptionGroupOutcomeExportRequest
	options    *exportpb.GetSubscriptionGroupOutcomeExportResponse
	members    *exportpb.GetSubscriptionGroupOutcomeExportResponse
	optionsErr error
	membersErr error
}

func (f *fakeExport) get(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
	f.calls = append(f.calls, req)
	if req.JobCategoryId == nil {
		return f.options, f.optionsErr
	}
	return f.members, f.membersErr
}

func testContext(groupID string) *exportpb.SubscriptionGroupOutcomeExportContext {
	schedule := "schedule-1"
	return &exportpb.SubscriptionGroupOutcomeExportContext{
		SubscriptionGroupId:   groupID,
		SubscriptionGroupName: "Grade 7 – Ñiño",
		PriceScheduleId:       &schedule,
	}
}

func optionsResponse() *exportpb.GetSubscriptionGroupOutcomeExportResponse {
	return &exportpb.GetSubscriptionGroupOutcomeExportResponse{
		Success: true,
		Context: testContext(testGroupID),
		JobCategories: []*exportpb.JobCategoryOption{
			{
				JobCategoryId: "category-a",
				JobTemplatePhases: []*exportpb.JobTemplatePhaseOption{
					{Code: "s2", Name: "Term 2", SequenceOrder: 2},
					{Code: "s1", Name: "Term 1", SequenceOrder: 1},
					{Code: "ambiguous_phase", Name: "Ambiguous", SequenceOrder: 9, Ambiguous: true},
				},
			},
			{
				JobCategoryId:         "category-b",
				FinalOutcomeAvailable: true,
				JobTemplatePhases: []*exportpb.JobTemplatePhaseOption{
					{Code: "s1", Name: "Term 1", SequenceOrder: 1},
					{Code: "progress_report", Name: "Progress Report", SequenceOrder: 3},
				},
			},
		},
	}
}

func membersResponse() *exportpb.GetSubscriptionGroupOutcomeExportResponse {
	return &exportpb.GetSubscriptionGroupOutcomeExportResponse{
		Success: true,
		Context: testContext(testGroupID),
		ClientRows: []*exportpb.SubscriptionGroupOutcomeClientRow{
			{ClientId: "client-2", ClientFirstName: "Ana", ClientLastName: "Cruz"},
			{ClientId: "client-1", ClientName: "Ben Reyes"},
			{ClientId: "client-2", ClientFirstName: "Ana", ClientLastName: "Cruz"},
			{ClientId: "  "},
			{ClientId: "client/3"},
		},
	}
}

func testRoutes() outcome_summary.Routes {
	return outcome_summary.Routes{
		ClientDocumentURL:                "/report-cards/section/{id}/student/{client_id}/document",
		ClientDocumentsDownloadDrawerURL: "/action/report-cards/section/{id}/download-all",
		ClientDocumentsManifestURL:       "/report-cards/section/{id}/download-all/manifest",
	}
}

func testOptions() outcome_summary.Options {
	return outcome_summary.Options{SubscriptionGroupExport: outcome_summary.SubscriptionGroupExportOptions{
		Enabled:            true,
		WholeReportProfile: bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_CLIENT_PHASE_OUTCOME_REPORT_V1,
	}}
}

type templateProbe struct {
	calls  int
	phases []string
	bytes  []byte
	err    error
}

func (p *templateProbe) resolve(_ context.Context, priceScheduleID, phaseCode string) ([]byte, error) {
	p.calls++
	p.phases = append(p.phases, priceScheduleID+"/"+phaseCode)
	return p.bytes, p.err
}

func testDeps(export *fakeExport, templates *templateProbe) *Deps {
	return &Deps{
		Routes:                            testRoutes(),
		Labels:                            outcome_summary.DefaultLabels(),
		Options:                           testOptions(),
		ResolvePrincipalKind:              func(context.Context) int32 { return outcome_summary.PrincipalKindStaff },
		PDFAvailable:                      true,
		DOCXAvailable:                     true,
		GetSubscriptionGroupOutcomeExport: export.get,
		ResolveTemplateBytes:              templates.resolve,
	}
}

func exportCtx() context.Context {
	return view.WithUserPermissions(context.Background(), types.NewUserPermissions([]string{
		outcome_summary.ExportPermissionEntity + ":" + outcome_summary.ExportReadAction,
	}))
}

func manifestRequest(ctx context.Context, groupID, query string) *http.Request {
	request := httptest.NewRequest(http.MethodGet, "/report-cards/section/"+groupID+"/download-all/manifest?"+query, nil).WithContext(ctx)
	request.SetPathValue("id", groupID)
	return request
}

func serveManifest(t *testing.T, deps *Deps, request *http.Request) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	recorder := httptest.NewRecorder()
	NewManifestHandler(deps).ServeHTTP(recorder, request)
	if got := recorder.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
	if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q", got)
	}
	body := map[string]any{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("manifest body is not JSON: %v\n%s", err, recorder.Body.String())
	}
	return recorder, body
}

func TestManifest_DeniedCallerIsForbiddenBeforeAnyRead(t *testing.T) {
	export := &fakeExport{options: optionsResponse(), members: membersResponse()}
	templates := &templateProbe{bytes: []byte("docx")}
	recorder, body := serveManifest(t, testDeps(export, templates), manifestRequest(context.Background(), testGroupID, "period=s1&format=pdf"))
	if recorder.Code != http.StatusForbidden || body["error"] != "forbidden" {
		t.Fatalf("status %d body %v, want 403 forbidden", recorder.Code, body)
	}
	if len(export.calls) != 0 || templates.calls != 0 {
		t.Fatalf("denied caller reached reads: export=%d templates=%d", len(export.calls), templates.calls)
	}
}

func TestManifest_ClientPrincipalIsForbidden(t *testing.T) {
	export := &fakeExport{options: optionsResponse(), members: membersResponse()}
	deps := testDeps(export, &templateProbe{bytes: []byte("docx")})
	deps.ResolvePrincipalKind = func(context.Context) int32 { return outcome_summary.PrincipalKindStaff + 1 }
	recorder, _ := serveManifest(t, deps, manifestRequest(exportCtx(), testGroupID, "period=s1"))
	if recorder.Code != http.StatusForbidden || len(export.calls) != 0 {
		t.Fatalf("non-staff principal: status %d reads %d, want 403 and no reads", recorder.Code, len(export.calls))
	}
}

func TestManifest_ScopedOptionsErrorOrForeignGroupFailsClosed(t *testing.T) {
	for name, export := range map[string]*fakeExport{
		"options error":   {optionsErr: errors.New("scope denied"), members: membersResponse()},
		"foreign context": {options: &exportpb.GetSubscriptionGroupOutcomeExportResponse{Success: true, Context: testContext("group-other"), JobCategories: optionsResponse().GetJobCategories()}, members: membersResponse()},
		"unsuccessful":    {options: &exportpb.GetSubscriptionGroupOutcomeExportResponse{Success: false, Context: testContext(testGroupID)}, members: membersResponse()},
		"members foreign": {options: optionsResponse(), members: &exportpb.GetSubscriptionGroupOutcomeExportResponse{Success: true, Context: testContext("group-other"), ClientRows: membersResponse().GetClientRows()}},
	} {
		t.Run(name, func(t *testing.T) {
			recorder, body := serveManifest(t, testDeps(export, &templateProbe{bytes: []byte("docx")}), manifestRequest(exportCtx(), testGroupID, "period=s1"))
			if recorder.Code != http.StatusForbidden || body["error"] != "forbidden" {
				t.Fatalf("status %d body %v, want 403 forbidden", recorder.Code, body)
			}
			if _, listed := body["clients"]; listed {
				t.Fatalf("forbidden manifest listed clients: %v", body)
			}
		})
	}
}

func TestManifest_ListsOnlyScopedMembersWithPrefixedDocumentURLs(t *testing.T) {
	export := &fakeExport{options: optionsResponse(), members: membersResponse()}
	templates := &templateProbe{bytes: []byte("docx-template")}
	ctx := render.WithRouteMap(exportCtx(), map[string]string{
		clientDocumentRouteKey: "/w/school/report-cards/section/{id}/student/{client_id}/document",
	})
	recorder := httptest.NewRecorder()
	NewManifestHandler(testDeps(export, templates)).ServeHTTP(recorder, manifestRequest(ctx, testGroupID, "period=S2&format=docx"))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d body %s", recorder.Code, recorder.Body.String())
	}
	var manifest Manifest
	if err := json.Unmarshal(recorder.Body.Bytes(), &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}

	if len(export.calls) != 2 {
		t.Fatalf("export reads = %d, want options + one member query", len(export.calls))
	}
	memberCall := export.calls[1]
	if memberCall.GetSubscriptionGroupId() != testGroupID || memberCall.GetJobCategoryId() != "category-a" || memberCall.GetJobTemplatePhaseCode() != "s2" {
		t.Fatalf("member query = %+v, want category-a / s2", memberCall)
	}
	if templates.calls != 1 || templates.phases[0] != "schedule-1/s2" {
		t.Fatalf("template preflight = %v, want one schedule-1/s2 resolve", templates.phases)
	}

	if manifest.Expected != 3 || len(manifest.Clients) != 3 {
		t.Fatalf("expected=%d clients=%d, want 3 de-duplicated members", manifest.Expected, len(manifest.Clients))
	}
	want := []ManifestClient{
		{ClientID: "client-2", Name: "Cruz, Ana", DocumentURL: "/w/school/report-cards/section/group-1/student/client-2/document?format=docx&period=s2"},
		{ClientID: "client-1", Name: "Ben Reyes", DocumentURL: "/w/school/report-cards/section/group-1/student/client-1/document?format=docx&period=s2"},
		{ClientID: "client/3", Name: "client/3", DocumentURL: "/w/school/report-cards/section/group-1/student/client%2F3/document?format=docx&period=s2"},
	}
	for index, client := range manifest.Clients {
		if client != want[index] {
			t.Errorf("client[%d] = %+v, want %+v", index, client, want[index])
		}
	}
	if manifest.Period != "s2" || manifest.PeriodLabel != "Term 2" || manifest.Format != FormatDOCX {
		t.Fatalf("manifest period/format = %q %q %q", manifest.Period, manifest.PeriodLabel, manifest.Format)
	}
	if manifest.ZipFilename != "grade-7-nino-term-2-docx.zip" {
		t.Fatalf("zip filename = %q", manifest.ZipFilename)
	}
	if strings.Contains(recorder.Body.String(), "docx-template") {
		t.Fatal("manifest leaked template bytes")
	}
}

func TestManifest_FallsBackToBootRouteWithoutWorkspaceRouteMap(t *testing.T) {
	export := &fakeExport{options: optionsResponse(), members: membersResponse()}
	recorder := httptest.NewRecorder()
	NewManifestHandler(testDeps(export, &templateProbe{bytes: []byte("t")})).ServeHTTP(recorder, manifestRequest(exportCtx(), testGroupID, "period=s1"))
	var manifest Manifest
	if err := json.Unmarshal(recorder.Body.Bytes(), &manifest); err != nil || recorder.Code != http.StatusOK {
		t.Fatalf("status %d err %v", recorder.Code, err)
	}
	if got := manifest.Clients[0].DocumentURL; got != "/report-cards/section/group-1/student/client-2/document?format=pdf&period=s1" {
		t.Fatalf("document url = %q (format must default to pdf)", got)
	}
}

func TestManifest_YearFinalUsesFinalSelectorWithoutTemplatePreflight(t *testing.T) {
	export := &fakeExport{options: optionsResponse(), members: membersResponse()}
	templates := &templateProbe{}
	recorder := httptest.NewRecorder()
	NewManifestHandler(testDeps(export, templates)).ServeHTTP(recorder, manifestRequest(exportCtx(), testGroupID, "period=year_final&format=pdf"))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d body %s", recorder.Code, recorder.Body.String())
	}
	if templates.calls != 0 {
		t.Fatalf("year final ran a template preflight")
	}
	memberCall := export.calls[1]
	if memberCall.GetJobCategoryId() != "category-b" || !memberCall.GetFinalOutcome() {
		t.Fatalf("year final member query = %+v, want category-b final", memberCall)
	}
}

func TestManifest_MissingPublishedTemplateIsUnprocessable(t *testing.T) {
	for name, templates := range map[string]*templateProbe{
		"nil bytes":  {},
		"error":      {err: errors.New("no binding")},
		"empty body": {bytes: []byte{}},
	} {
		t.Run(name, func(t *testing.T) {
			export := &fakeExport{options: optionsResponse(), members: membersResponse()}
			recorder, body := serveManifest(t, testDeps(export, templates), manifestRequest(exportCtx(), testGroupID, "period=progress_report"))
			if recorder.Code != http.StatusUnprocessableEntity || body["error"] != "template_unavailable" {
				t.Fatalf("status %d body %v, want 422 template_unavailable", recorder.Code, body)
			}
			if len(export.calls) != 1 {
				t.Fatalf("member list read despite missing template (reads=%d)", len(export.calls))
			}
		})
	}
}

func TestManifest_RejectsUnavailablePeriodAndFormat(t *testing.T) {
	cases := []struct {
		name, query string
		pdf, docx   bool
		status      int
		code        string
	}{
		{"missing period", "format=pdf", true, true, http.StatusBadRequest, "invalid_period"},
		{"unknown period", "period=s9", true, true, http.StatusNotFound, "period_unavailable"},
		{"ambiguous everywhere", "period=ambiguous_phase", true, true, http.StatusNotFound, "period_unavailable"},
		{"unknown format", "period=s1&format=xlsx", true, true, http.StatusBadRequest, "invalid_format"},
		{"pdf not wired", "period=s1&format=pdf", false, true, http.StatusUnprocessableEntity, "format_unavailable"},
		{"docx not wired", "period=s1&format=docx", true, false, http.StatusUnprocessableEntity, "format_unavailable"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			export := &fakeExport{options: optionsResponse(), members: membersResponse()}
			deps := testDeps(export, &templateProbe{bytes: []byte("t")})
			deps.PDFAvailable, deps.DOCXAvailable = tc.pdf, tc.docx
			recorder, body := serveManifest(t, deps, manifestRequest(exportCtx(), testGroupID, tc.query))
			if recorder.Code != tc.status || body["error"] != tc.code {
				t.Fatalf("status %d body %v, want %d %s", recorder.Code, body, tc.status, tc.code)
			}
		})
	}
}

func TestManifest_MemberReadFailureAndEmptyGroup(t *testing.T) {
	export := &fakeExport{options: optionsResponse(), membersErr: errors.New("db down")}
	recorder, body := serveManifest(t, testDeps(export, &templateProbe{bytes: []byte("t")}), manifestRequest(exportCtx(), testGroupID, "period=s1"))
	if recorder.Code != http.StatusServiceUnavailable || body["error"] != "members_unavailable" {
		t.Fatalf("member error: status %d body %v", recorder.Code, body)
	}

	empty := &fakeExport{options: optionsResponse(), members: &exportpb.GetSubscriptionGroupOutcomeExportResponse{Success: true, Context: testContext(testGroupID)}}
	recorder, body = serveManifest(t, testDeps(empty, &templateProbe{bytes: []byte("t")}), manifestRequest(exportCtx(), testGroupID, "period=s1"))
	if recorder.Code != http.StatusNotFound || body["error"] != "empty" {
		t.Fatalf("empty group: status %d body %v", recorder.Code, body)
	}
}

func TestManifest_NotConfiguredIsUnavailable(t *testing.T) {
	export := &fakeExport{options: optionsResponse(), members: membersResponse()}
	deps := testDeps(export, &templateProbe{bytes: []byte("t")})
	deps.Routes.ClientDocumentsManifestURL = ""
	recorder, body := serveManifest(t, deps, manifestRequest(exportCtx(), testGroupID, "period=s1"))
	if recorder.Code != http.StatusServiceUnavailable || body["error"] != "not_configured" || len(export.calls) != 0 {
		t.Fatalf("status %d body %v reads %d", recorder.Code, body, len(export.calls))
	}
}

func TestGroupPeriods_UnionInCatalogOrderPlusYearFinal(t *testing.T) {
	periods := groupPeriods(outcome_summary.DefaultLabels(), optionsResponse().GetJobCategories())
	var codes []string
	for _, period := range periods {
		codes = append(codes, period.Code)
	}
	if got := strings.Join(codes, ","); got != "s1,s2,progress_report,year_final" {
		t.Fatalf("periods = %s", got)
	}
	if len(groupPeriods(outcome_summary.DefaultLabels(), nil)) != 0 {
		t.Fatal("a group without categories must offer no period")
	}
}

func TestDrawer_GatesThenOffersScopedPeriodsAndFormats(t *testing.T) {
	export := &fakeExport{options: optionsResponse(), members: membersResponse()}
	deps := testDeps(export, &templateProbe{})
	request := httptest.NewRequest(http.MethodGet, "/action/report-cards/section/group-1/download-all", nil)
	request.SetPathValue("id", testGroupID)

	denied := NewDrawer(deps).Handle(context.Background(), &view.ViewContext{Request: request})
	if denied.StatusCode != http.StatusForbidden || len(export.calls) != 0 {
		t.Fatalf("denied drawer: status %d reads %d", denied.StatusCode, len(export.calls))
	}

	result := NewDrawer(deps).Handle(exportCtx(), &view.ViewContext{Request: request, CacheVersion: "v42"})
	if result.Error != nil || result.Template != DrawerTemplate {
		t.Fatalf("drawer result = %q status %d err %v", result.Template, result.StatusCode, result.Error)
	}
	data, ok := result.Data.(*DrawerData)
	if !ok {
		t.Fatalf("drawer data = %T", result.Data)
	}
	if data.ManifestURL != "/report-cards/section/group-1/download-all/manifest" || data.CacheVersion != "v42" || data.Concurrency != DefaultConcurrency {
		t.Fatalf("drawer data = %+v", data)
	}
	if len(data.Periods) != 4 || !data.Periods[0].Selected || data.Periods[3].Value != PeriodYearFinal {
		t.Fatalf("periods = %+v", data.Periods)
	}
	if len(data.Formats) != 2 || data.Formats[0].Value != FormatPDF || !data.Formats[0].Selected {
		t.Fatalf("formats = %+v", data.Formats)
	}
	if len(export.calls) != 1 || export.calls[0].JobCategoryId != nil {
		t.Fatalf("drawer must read options only, reads = %d", len(export.calls))
	}
	deps.Options.Document.BatchConcurrency = 4
	configured := NewDrawer(deps).Handle(exportCtx(), &view.ViewContext{Request: request, CacheVersion: "v42"})
	if got, ok := configured.Data.(*DrawerData); !ok || got.Concurrency != 4 {
		t.Fatalf("configured drawer data = %+v", configured.Data)
	}
	deps.Options.Document.BatchConcurrency = 0

	foreign := &fakeExport{options: &exportpb.GetSubscriptionGroupOutcomeExportResponse{Success: true, Context: testContext("group-other")}}
	if got := NewDrawer(testDeps(foreign, &templateProbe{})).Handle(exportCtx(), &view.ViewContext{Request: request}); got.StatusCode != http.StatusForbidden {
		t.Fatalf("foreign group drawer status = %d, want 403", got.StatusCode)
	}
}

func TestNewOpenAction_VisibleOnlyWhenPerClientDownloadIsUsable(t *testing.T) {
	cfg := ActionConfig{
		Routes:                testRoutes(),
		Labels:                outcome_summary.DefaultLabels(),
		Options:               testOptions(),
		ResolvePrincipalKind:  func(context.Context) int32 { return outcome_summary.PrincipalKindStaff },
		ClientDocumentMounted: true,
		ExportQueryWired:      true,
	}
	action := NewOpenAction(exportCtx(), cfg, testGroupID)
	if action == nil || action.ActionURL != "/action/report-cards/section/group-1/download-all" || action.Label == "" || action.Title == "" {
		t.Fatalf("action = %+v", action)
	}
	if NewOpenAction(context.Background(), cfg, testGroupID) != nil {
		t.Fatal("action offered without export permission")
	}
	unmounted := cfg
	unmounted.ClientDocumentMounted = false
	if NewOpenAction(exportCtx(), unmounted, testGroupID) != nil {
		t.Fatal("action offered without the per-client document route")
	}
	unrouted := cfg
	unrouted.Routes.ClientDocumentsDownloadDrawerURL = ""
	if NewOpenAction(exportCtx(), unrouted, testGroupID) != nil {
		t.Fatal("action offered without its drawer route")
	}
	if NewOpenAction(exportCtx(), cfg, " ") != nil {
		t.Fatal("action offered without a group id")
	}
}

func TestDrawerTemplate_RendersBulkDownloadContract(t *testing.T) {
	labels := outcome_summary.DefaultLabels()
	tmpl, err := template.ParseFiles("../templates/client-documents-download-drawer.html")
	if err != nil {
		t.Fatalf("parse batch drawer template: %v", err)
	}
	var drawer bytes.Buffer
	if err := tmpl.ExecuteTemplate(&drawer, DrawerTemplate, &DrawerData{
		ManifestURL:  "/w/school/report-cards/section/group-1/download-all/manifest",
		Periods:      []types.SelectOption{{Value: "s1", Label: "Term 1", Selected: true}},
		Formats:      []types.SelectOption{{Value: FormatPDF, Label: "PDF", Selected: true}},
		Labels:       labels.ClientDocumentsDownload,
		OptionLabels: labels.ClientDocumentDownload,
		CommonLabels: map[string]any{"Buttons": map[string]string{"Cancel": "Cancel"}},
		CacheVersion: "v42",
		Concurrency:  DefaultConcurrency,
	}); err != nil {
		t.Fatalf("render batch drawer: %v", err)
	}
	rendered := drawer.String()
	for _, want := range []string{
		`hx-boost="false" data-lf-bulk-download`,
		`data-manifest-url="/w/school/report-cards/section/group-1/download-all/manifest"`,
		`data-concurrency="6"`,
		`id="rc-client-documents-period"`,
		`id="rc-client-documents-format"`,
		`data-testid="rc-client-documents-progress"`,
		`data-testid="rc-client-documents-failures"`,
		`data-lf-bulk-download-action="retry"`,
		`data-lf-bulk-download-action="stop"`,
		`data-lf-bulk-download-action="save"`,
		`data-testid="rc-client-documents-start"`,
		`disabled data-lf-bulk-download-pending`,
		`<script src="/assets/js/pyeza/bulk-download.js?v=v42"></script>`,
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("batch drawer missing %q", want)
		}
	}
	for _, forbidden := range []string{`data-lf-download-form`, `data-download-form`, `<script>`} {
		if strings.Contains(rendered, forbidden) {
			t.Errorf("batch drawer must not contain %q (download-indicator / CSP)", forbidden)
		}
	}

	var button bytes.Buffer
	if err := tmpl.ExecuteTemplate(&button, OpenActionTemplate, &OpenAction{ActionURL: "/action/report-cards/section/group-1/download-all", Label: "Download all", Title: "Download all"}); err != nil {
		t.Fatalf("render open action: %v", err)
	}
	for _, want := range []string{`hx-get="/action/report-cards/section/group-1/download-all"`, `hx-target="#sheetContent"`, `data-lf-action="sheet-open"`, `data-testid="rc-client-documents-download-open"`} {
		if !strings.Contains(button.String(), want) {
			t.Errorf("open action missing %q:\n%s", want, button.String())
		}
	}
}
