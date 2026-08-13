package section_template_settings

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	documenttemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/template"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/subscription_group_document_template"
	planpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/plan"
	priceschedulepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/price_schedule"
)

func settingsPerms(codes ...string) context.Context {
	return view.WithUserPermissions(context.Background(), types.NewUserPermissions(codes))
}

func canonicalSettingsDOCX(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("../section_document/subscription-group-outcome-matrix-single-period-11-v1.docx")
	if err != nil {
		t.Fatalf("read canonical DOCX: %v", err)
	}
	return data
}

func settingsCategory() *jobcategorypb.JobCategory {
	code, order := "academic", int32(1)
	return &jobcategorypb.JobCategory{Id: "cat-academic", Name: "Academic", Active: true, Code: &code, SortOrder: &order}
}

type settingsRecorder struct {
	order                      []string
	storedContainer, storedKey string
	storedBytes                []byte
	artifact                   *documenttemplatepb.DocumentTemplate
	binding                    *bindingpb.SubscriptionGroupDocumentTemplate
	deleteContainer, deleteKey string
	deleteObjectErr            error
	deletePairIDs              []string
	deletePairCallsAtObject    int
	pairErr                    error
	deletePairErr              error
	bindings                   []*bindingpb.SubscriptionGroupDocumentTemplate
}

func (r *settingsRecorder) deps(t *testing.T) *Deps {
	t.Helper()
	labels := outcome_summary.DefaultLabels()
	return &Deps{
		Routes:  outcome_summary.Routes{SectionTemplateSettingsURL: "/section-templates", SectionTemplateUploadURL: "/section-templates/upload", SectionTemplatePublishURL: "/section-templates/publish", SectionTemplateDeleteURL: "/section-templates/delete"},
		Labels:  labels,
		Options: outcome_summary.Options{SectionExport: outcome_summary.SectionExportOptions{ProfileByCategoryCode: map[string]bindingpb.RenderProfile{"academic": bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1}}},
		ListJobCategories: func(context.Context, *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error) {
			return &jobcategorypb.ListJobCategoriesResponse{Success: true, Data: []*jobcategorypb.JobCategory{settingsCategory()}}, nil
		},
		ListPriceSchedules: func(context.Context, *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error) {
			return &priceschedulepb.ListPriceSchedulesResponse{Success: true, Data: []*priceschedulepb.PriceSchedule{{Id: "schedule-1", Name: "Term 1"}}}, nil
		},
		ListPlans: func(context.Context, *planpb.ListPlansRequest) (*planpb.ListPlansResponse, error) {
			return &planpb.ListPlansResponse{Success: true, Data: []*planpb.Plan{{Id: strptr("plan-1"), Name: "Standard"}}}, nil
		},
		StoreTemplate: func(_ context.Context, key string, content []byte, _ string) (string, error) {
			r.order = append(r.order, "store")
			r.storedContainer, r.storedKey, r.storedBytes = "physical-templates", key, append([]byte(nil), content...)
			return r.storedContainer, nil
		},
		DeleteTemplateObject: func(_ context.Context, container, key string) error {
			r.order = append(r.order, "delete-object")
			r.deletePairCallsAtObject = len(r.deletePairIDs)
			r.deleteContainer, r.deleteKey = container, key
			return r.deleteObjectErr
		},
		CreateUploadPair: func(_ context.Context, artifact *documenttemplatepb.DocumentTemplate, binding *bindingpb.SubscriptionGroupDocumentTemplate) (*documenttemplatepb.DocumentTemplate, *bindingpb.SubscriptionGroupDocumentTemplate, error) {
			r.order = append(r.order, "pair")
			r.artifact, r.binding = artifact, binding
			if r.pairErr != nil {
				return nil, nil, r.pairErr
			}
			return artifact, binding, nil
		},
		ListTemplateBindings: func(context.Context, *bindingpb.ListSubscriptionGroupDocumentTemplatesRequest) (*bindingpb.ListSubscriptionGroupDocumentTemplatesResponse, error) {
			return &bindingpb.ListSubscriptionGroupDocumentTemplatesResponse{Success: true, Data: r.bindings}, nil
		},
		DeleteDraftPair: func(_ context.Context, id string) (*documenttemplatepb.DocumentTemplate, error) {
			r.order = append(r.order, "pair-delete")
			r.deletePairIDs = append(r.deletePairIDs, id)
			if r.deletePairErr != nil {
				return nil, r.deletePairErr
			}
			container, key := "physical", "templates/subscription_group/"+id+".docx"
			return &documenttemplatepb.DocumentTemplate{Id: "doc-" + id, StorageContainer: &container, StorageKey: &key}, nil
		},
	}
}

func settingsUploadPost(t *testing.T, content []byte, fields map[string]string) *view.ViewContext {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("name", "Canonical Template")
	for key, value := range fields {
		_ = mw.WriteField(key, value)
	}
	part, err := mw.CreateFormFile("template_file", "template.docx")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/section-templates/upload", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return &view.ViewContext{Request: req}
}

func TestSectionTemplateSettings_PermissionAndLifecycleActions(t *testing.T) {
	for _, tc := range []struct {
		name   string
		perms  []string
		wantOK bool
	}{
		{"none", nil, false},
		{"artifact only", []string{"document_template:create"}, false},
		{"binding only", []string{"subscription_group_document_template:create"}, false},
		{"both", []string{"document_template:create", "subscription_group_document_template:create"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := &settingsRecorder{}
			res := NewUploadAction(rec.deps(t)).Handle(settingsPerms(tc.perms...), settingsUploadPost(t, canonicalSettingsDOCX(t), map[string]string{"job_category_id": "cat-academic", "profile": "forged", "storage_key": "forged", "storage_container": "forged"}))
			if (res.StatusCode == http.StatusOK) != tc.wantOK {
				t.Fatalf("status=%d wantOK=%v", res.StatusCode, tc.wantOK)
			}
			if !tc.wantOK && len(rec.order) != 0 {
				t.Fatalf("denied upload made calls: %v", rec.order)
			}
			if tc.wantOK {
				if fmt.Sprint(rec.order) != "[store pair]" {
					t.Fatalf("order=%v, want store then pair", rec.order)
				}
				if rec.artifact.GetStorageContainer() != "physical-templates" || rec.artifact.GetStorageKey() != rec.storedKey || !strings.HasPrefix(rec.storedKey, storagePrefix+"/") {
					t.Fatalf("artifact locator=%q/%q", rec.artifact.GetStorageContainer(), rec.artifact.GetStorageKey())
				}
				if rec.binding.GetJobCategoryId() != "cat-academic" || rec.binding.GetRenderProfile() != bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1 {
					t.Fatalf("binding category/profile=%q/%v", rec.binding.GetJobCategoryId(), rec.binding.GetRenderProfile())
				}
			}
		})
	}

	rec := &settingsRecorder{pairErr: errors.New("pair failed")}
	res := NewUploadAction(rec.deps(t)).Handle(settingsPerms("document_template:create", "subscription_group_document_template:create"), settingsUploadPost(t, canonicalSettingsDOCX(t), map[string]string{"job_category_id": "cat-academic"}))
	if res.StatusCode == http.StatusOK || fmt.Sprint(rec.order) != "[store pair delete-object]" || rec.deleteContainer != "physical-templates" || rec.deleteKey != rec.storedKey {
		t.Fatalf("pair failure status/order/locator=%d/%v/%q/%q", res.StatusCode, rec.order, rec.deleteContainer, rec.deleteKey)
	}

	rec = &settingsRecorder{pairErr: errors.New("pair failed"), deleteObjectErr: errors.New("cleanup failed")}
	res = NewUploadAction(rec.deps(t)).Handle(settingsPerms("document_template:create", "subscription_group_document_template:create"), settingsUploadPost(t, canonicalSettingsDOCX(t), map[string]string{"job_category_id": "cat-academic"}))
	if got := res.Headers["HX-Error-Message"]; got != rec.deps(t).Labels.SectionTemplateSettings.CleanupFailed {
		t.Fatalf("cleanup failure message=%q", got)
	}
}

func TestSectionTemplateSettings_ProfileAwareWildcardScopes(t *testing.T) {
	rec := &settingsRecorder{}
	res := NewUploadAction(rec.deps(t)).Handle(settingsPerms("document_template:create", "subscription_group_document_template:create"), settingsUploadPost(t, canonicalSettingsDOCX(t), map[string]string{"job_category_id": "cat-academic"}))
	if res.StatusCode != http.StatusOK {
		t.Fatalf("wildcard upload status=%d", res.StatusCode)
	}
	if rec.binding.GetPriceScheduleId() != "" || rec.binding.GetPlanId() != "" {
		t.Fatalf("blank wildcard axes should remain nil: schedule=%q plan=%q", rec.binding.GetPriceScheduleId(), rec.binding.GetPlanId())
	}

	rec = &settingsRecorder{}
	deps := rec.deps(t)
	deps.ListPriceSchedules = func(context.Context, *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error) {
		return &priceschedulepb.ListPriceSchedulesResponse{Success: true, Data: []*priceschedulepb.PriceSchedule{{Id: "schedule-1", Name: "Term 1"}}}, nil
	}
	deps.ListPlans = func(context.Context, *planpb.ListPlansRequest) (*planpb.ListPlansResponse, error) {
		return &planpb.ListPlansResponse{Success: true, Data: []*planpb.Plan{{Id: strptr("plan-1"), Name: "Standard"}}}, nil
	}
	res = NewUploadAction(deps).Handle(settingsPerms("document_template:create", "subscription_group_document_template:create"), settingsUploadPost(t, canonicalSettingsDOCX(t), map[string]string{"job_category_id": "cat-academic", "price_schedule_id": "schedule-1", "plan_id": "plan-1"}))
	if res.StatusCode != http.StatusOK || rec.binding.GetPriceScheduleId() != "schedule-1" || rec.binding.GetPlanId() != "plan-1" {
		t.Fatalf("specific axes status/binding=%d/%q/%q", res.StatusCode, rec.binding.GetPriceScheduleId(), rec.binding.GetPlanId())
	}

	rec = &settingsRecorder{}
	res = NewUploadAction(rec.deps(t)).Handle(settingsPerms("document_template:create", "subscription_group_document_template:create"), settingsUploadPost(t, canonicalSettingsDOCX(t), map[string]string{}))
	if res.StatusCode == http.StatusOK || len(rec.order) != 0 {
		t.Fatalf("category-null accepted or called storage: status=%d order=%v", res.StatusCode, rec.order)
	}
}

func TestSectionTemplateSettings_RejectsManifestMismatch(t *testing.T) {
	rec := &settingsRecorder{}
	res := NewUploadAction(rec.deps(t)).Handle(settingsPerms("document_template:create", "subscription_group_document_template:create"), settingsUploadPost(t, []byte("not a DOCX"), map[string]string{"job_category_id": "cat-academic"}))
	if res.StatusCode == http.StatusOK || len(rec.order) != 0 {
		t.Fatalf("invalid manifest reached persistence: status=%d order=%v", res.StatusCode, rec.order)
	}
}

func TestSectionTemplateSettings_PublishDeleteAndListActionState(t *testing.T) {
	rec := &settingsRecorder{}
	published := ""
	deps := rec.deps(t)
	deps.PublishTemplateBinding = func(_ context.Context, req *bindingpb.PublishSubscriptionGroupDocumentTemplateRequest) (*bindingpb.PublishSubscriptionGroupDocumentTemplateResponse, error) {
		published = req.GetId()
		return &bindingpb.PublishSubscriptionGroupDocumentTemplateResponse{Success: true}, nil
	}
	vc := &view.ViewContext{Request: httptest.NewRequest(http.MethodPost, "/publish?id=b-1", nil)}
	if res := NewPublishAction(deps).Handle(settingsPerms("document_template:update"), vc); res.StatusCode == http.StatusOK || published != "" {
		t.Fatalf("artifact update must not publish: %d/%q", res.StatusCode, published)
	}
	if res := NewPublishAction(deps).Handle(settingsPerms("subscription_group_document_template:update"), vc); res.StatusCode != http.StatusOK || published != "b-1" {
		t.Fatalf("binding update publish=%d/%q", res.StatusCode, published)
	}

	artifact := &documenttemplatepb.DocumentTemplate{Id: "doc-1", Name: "Academic Template"}
	binding := &bindingpb.SubscriptionGroupDocumentTemplate{Id: "b-1", DocumentTemplateId: "doc-1", DocumentTemplate: artifact, RenderProfile: bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1, Version: 3, VersionStatus: enums.VersionStatus_VERSION_STATUS_DRAFT, PriceScheduleId: strptr("schedule-1"), PlanId: strptr("plan-1"), JobCategoryId: strptr("cat-academic"), JobCategory: settingsCategory()}
	deps.ListTemplateBindings = func(context.Context, *bindingpb.ListSubscriptionGroupDocumentTemplatesRequest) (*bindingpb.ListSubscriptionGroupDocumentTemplatesResponse, error) {
		return &bindingpb.ListSubscriptionGroupDocumentTemplatesResponse{Success: true, Data: []*bindingpb.SubscriptionGroupDocumentTemplate{binding}}, nil
	}
	page := NewListView(deps).Handle(settingsPerms("subscription_group_document_template:list"), &view.ViewContext{Request: httptest.NewRequest(http.MethodGet, "/section-templates", nil)})
	if page.StatusCode != http.StatusOK {
		t.Fatalf("list status=%d", page.StatusCode)
	}
	data := page.Data.(*PageData)
	if len(data.Table.Rows) != 1 || len(data.Table.Rows[0].Actions) != 2 || data.Table.Rows[0].Cells[4].Value == "—" || !strings.Contains(data.Table.Rows[0].Cells[0].Value, "Academic") {
		t.Fatalf("list row=%+v", data.Table.Rows[0])
	}
	if !data.Table.Rows[0].Actions[0].Disabled || !data.Table.Rows[0].Actions[1].Disabled {
		t.Fatalf("missing action permissions should disable both: %+v", data.Table.Rows[0].Actions)
	}
	attrs := data.Table.Rows[0].DataAttrs
	if attrs["renderprofile"] != "subscription_group_outcome_matrix_single_period_11_v1" || attrs["category"] != "academic" || attrs["categoryscope"] != "exact" || attrs["schedulescope"] != "exact" || attrs["planscope"] != "exact" || attrs["status"] != "draft" {
		t.Fatalf("list row contract attributes=%v", attrs)
	}
	var renderedAttrs bytes.Buffer
	attrTemplate := template.Must(template.New("row-attrs").Parse(`<tr{{range $key, $value := .}} data-{{$key}}="{{$value}}"{{end}}></tr>`))
	if err := attrTemplate.Execute(&renderedAttrs, attrs); err != nil {
		t.Fatalf("render list row contract attributes: %v", err)
	}
	if got := renderedAttrs.String(); strings.Contains(got, "ZgotmplZ") || !strings.Contains(got, `data-renderprofile="subscription_group_outcome_matrix_single_period_11_v1"`) {
		t.Fatalf("unsafe list row contract attributes=%s", got)
	}
	if data.Table.Rows[0].Actions[0].TestID != "section-template-publish-b-1" || data.Table.Rows[0].Actions[1].TestID != "section-template-delete-b-1" {
		t.Fatalf("list row action test IDs=%+v", data.Table.Rows[0].Actions)
	}

	vc = &view.ViewContext{Request: httptest.NewRequest(http.MethodPost, "/delete?id=b-1", nil)}
	rec = &settingsRecorder{bindings: []*bindingpb.SubscriptionGroupDocumentTemplate{{Id: "b-1", DocumentTemplateId: "doc-1", DocumentTemplate: &documenttemplatepb.DocumentTemplate{Id: "doc-1", StorageContainer: strptr("physical"), StorageKey: strptr("templates/subscription_group/doc-1.docx")}, VersionStatus: enums.VersionStatus_VERSION_STATUS_DRAFT}}}
	deps = rec.deps(t)
	if res := NewDeleteAction(deps).Handle(settingsPerms("subscription_group_document_template:delete"), vc); res.StatusCode == http.StatusOK || len(rec.order) != 0 {
		t.Fatalf("binding-only delete bypassed conjunction: %d/%v", res.StatusCode, rec.order)
	}
	if res := NewDeleteAction(deps).Handle(settingsPerms("subscription_group_document_template:delete", "document_template:delete"), vc); res.StatusCode != http.StatusOK || fmt.Sprint(rec.order) != "[pair-delete delete-object]" {
		t.Fatalf("delete status/order=%d/%v", res.StatusCode, rec.order)
	}
	if len(rec.deletePairIDs) != 1 || rec.deletePairIDs[0] != "b-1" || rec.deletePairCallsAtObject != 1 || rec.deleteContainer != "physical" || rec.deleteKey != "templates/subscription_group/b-1.docx" {
		t.Fatalf("delete call/locator=%v/%d/%q/%q", rec.deletePairIDs, rec.deletePairCallsAtObject, rec.deleteContainer, rec.deleteKey)
	}

	rec = &settingsRecorder{deletePairErr: errors.New("pair failed")}
	deps = rec.deps(t)
	res := NewDeleteAction(deps).Handle(settingsPerms("subscription_group_document_template:delete", "document_template:delete"), vc)
	if res.StatusCode == http.StatusOK || fmt.Sprint(rec.order) != "[pair-delete]" {
		t.Fatalf("pair failure status/order=%d/%v", res.StatusCode, rec.order)
	}

	rec = &settingsRecorder{deleteObjectErr: errors.New("cleanup failed")}
	deps = rec.deps(t)
	res = NewDeleteAction(deps).Handle(settingsPerms("subscription_group_document_template:delete", "document_template:delete"), vc)
	if got := res.Headers["HX-Error-Message"]; got != deps.Labels.SectionTemplateSettings.CleanupFailed {
		t.Fatalf("post-commit cleanup failure message=%q", got)
	}
	if fmt.Sprint(rec.order) != "[pair-delete delete-object]" {
		t.Fatalf("cleanup failure order=%v", rec.order)
	}
	if len(rec.deletePairIDs) != 1 || rec.deletePairIDs[0] != "b-1" {
		t.Fatalf("cleanup failure repeated DB call: %v", rec.deletePairIDs)
	}
}

func TestSectionTemplateSettings_PaginatesAndRefusesTypedDeleteFromSharedReference(t *testing.T) {
	rec := &settingsRecorder{}
	pageOne := make([]*bindingpb.SubscriptionGroupDocumentTemplate, 0, 100)
	for i := 0; i < 99; i++ {
		id := fmt.Sprintf("filler-%03d", i)
		pageOne = append(pageOne, &bindingpb.SubscriptionGroupDocumentTemplate{
			Id: id, DocumentTemplateId: "doc-" + id, Active: true,
			DocumentTemplate: &documenttemplatepb.DocumentTemplate{Id: "doc-" + id, Name: "Filler " + id},
			VersionStatus:    enums.VersionStatus_VERSION_STATUS_PUBLISHED,
		})
	}
	target := &bindingpb.SubscriptionGroupDocumentTemplate{
		Id: "target-draft", DocumentTemplateId: "doc-shared", Active: true,
		DocumentTemplate: &documenttemplatepb.DocumentTemplate{
			Id: "doc-shared", Name: "Target draft", StorageContainer: strptr("physical"), StorageKey: strptr("templates/target.docx"),
		},
		RenderProfile: bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1,
		VersionStatus: enums.VersionStatus_VERSION_STATUS_DRAFT,
	}
	pageOne = append(pageOne, target)
	pageTwo := []*bindingpb.SubscriptionGroupDocumentTemplate{{
		Id: "page-2-shared", DocumentTemplateId: "doc-shared", Active: true,
		DocumentTemplate: &documenttemplatepb.DocumentTemplate{Id: "doc-shared", Name: "Shared active"},
		VersionStatus:    enums.VersionStatus_VERSION_STATUS_PUBLISHED,
	}}

	var requests []*bindingpb.ListSubscriptionGroupDocumentTemplatesRequest
	deps := rec.deps(t)
	deps.ListTemplateBindings = func(_ context.Context, req *bindingpb.ListSubscriptionGroupDocumentTemplatesRequest) (*bindingpb.ListSubscriptionGroupDocumentTemplatesResponse, error) {
		requests = append(requests, req)
		switch req.GetPagination().GetOffset().GetPage() {
		case 1:
			return &bindingpb.ListSubscriptionGroupDocumentTemplatesResponse{Success: true, Data: pageOne}, nil
		case 2:
			return &bindingpb.ListSubscriptionGroupDocumentTemplatesResponse{Success: true, Data: pageTwo}, nil
		default:
			return &bindingpb.ListSubscriptionGroupDocumentTemplatesResponse{Success: true}, nil
		}
	}

	page := NewListView(deps).Handle(settingsPerms("subscription_group_document_template:list"), &view.ViewContext{Request: httptest.NewRequest(http.MethodGet, "/section-templates", nil)})
	if page.StatusCode != http.StatusOK {
		t.Fatalf("list status=%d", page.StatusCode)
	}
	if len(requests) != 2 {
		t.Fatalf("list requests=%d, want 2 pages", len(requests))
	}
	for i, req := range requests {
		if got := req.GetPagination().GetLimit(); got != 100 {
			t.Errorf("request %d limit=%d, want 100", i+1, got)
		}
		if got := req.GetPagination().GetOffset().GetPage(); got != int32(i+1) {
			t.Errorf("request %d page=%d, want %d", i+1, got, i+1)
		}
	}
	data := page.Data.(*PageData)
	if len(data.Table.Rows) != 101 {
		t.Fatalf("rows=%d, want 101 including page 2", len(data.Table.Rows))
	}
	var pageTwoRow *types.TableRow
	for i := range data.Table.Rows {
		if data.Table.Rows[i].ID == "page-2-shared" {
			pageTwoRow = &data.Table.Rows[i]
			break
		}
	}
	if pageTwoRow == nil || pageTwoRow.DataAttrs["testid"] != "section-template-row-2-shared" {
		t.Fatalf("page 2 row=%+v", pageTwoRow)
	}

	requests = nil
	rec.order = nil
	rec.deletePairErr = errors.New("shared template reference")
	res := NewDeleteAction(deps).Handle(settingsPerms("subscription_group_document_template:delete", "document_template:delete"), &view.ViewContext{Request: httptest.NewRequest(http.MethodPost, "/delete?id=target-draft", nil)})
	if res.StatusCode == http.StatusOK {
		t.Fatalf("shared-reference delete unexpectedly succeeded: %d", res.StatusCode)
	}
	if fmt.Sprint(rec.order) != "[pair-delete]" || len(rec.deletePairIDs) != 1 || rec.deletePairIDs[0] != "target-draft" {
		t.Fatalf("typed delete order/call=%v/%v", rec.order, rec.deletePairIDs)
	}
	if rec.deleteContainer != "" || rec.deleteKey != "" {
		t.Fatalf("shared-reference delete touched object: %q/%q", rec.deleteContainer, rec.deleteKey)
	}
	if len(requests) != 0 {
		t.Fatalf("delete used table listing as authority: %v", requests)
	}
}

func TestSectionTemplateSettings_ListMaxPageGuard(t *testing.T) {
	fullPage := make([]*bindingpb.SubscriptionGroupDocumentTemplate, 100)
	entry := &bindingpb.SubscriptionGroupDocumentTemplate{Id: "reused"}
	for i := range fullPage {
		fullPage[i] = entry
	}
	calls := 0
	deps := (&settingsRecorder{}).deps(t)
	deps.ListTemplateBindings = func(context.Context, *bindingpb.ListSubscriptionGroupDocumentTemplatesRequest) (*bindingpb.ListSubscriptionGroupDocumentTemplatesResponse, error) {
		calls++
		return &bindingpb.ListSubscriptionGroupDocumentTemplatesResponse{Success: true, Data: fullPage}, nil
	}
	page := NewListView(deps).Handle(settingsPerms("subscription_group_document_template:list"), &view.ViewContext{Request: httptest.NewRequest(http.MethodGet, "/section-templates", nil)})
	if page.StatusCode != http.StatusOK || calls != 100 {
		t.Fatalf("max-page guard status/calls=%d/%d, want 200/100", page.StatusCode, calls)
	}
	if data := page.Data.(*PageData); len(data.Table.Rows) != 0 {
		t.Fatalf("guard failure rendered rows=%d", len(data.Table.Rows))
	}
}

func strptr(value string) *string { return &value }
