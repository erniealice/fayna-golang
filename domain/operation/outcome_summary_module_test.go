package operation

import (
	"archive/zip"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	outcomesummarypkg "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	subscriptiongroupdocumenttemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/subscription_group_document_template"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	productplanpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/product/product_plan"
	productplanstaffpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/product/product_plan_staff"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
	sgppspb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_product_plan_staff"
)

type outcomeSummaryRouteRecorder struct {
	get  map[string]bool
	post map[string]bool
	raw  map[string]bool
}

func newOutcomeSummaryRouteRecorder() *outcomeSummaryRouteRecorder {
	return &outcomeSummaryRouteRecorder{
		get:  make(map[string]bool),
		post: make(map[string]bool),
		raw:  make(map[string]bool),
	}
}

func (r *outcomeSummaryRouteRecorder) GET(path string, _ view.View, _ ...string) {
	r.get[path] = true
}

func (r *outcomeSummaryRouteRecorder) POST(path string, _ view.View, _ ...string) {
	r.post[path] = true
}

func (r *outcomeSummaryRouteRecorder) HandleFunc(method, path string, _ http.HandlerFunc, _ ...string) {
	r.raw[method+" "+path] = true
}

func TestOutcomeSummaryModule_SubscriptionGroupDocumentTemplateRoutesRequireTypedEnablement(t *testing.T) {
	routes := outcomesummarypkg.DefaultRoutes()
	routes.SubscriptionGroupDownloadDrawerURL = "/outcomes/subscription-group/{id}/download"
	routes.SubscriptionGroupDocumentTemplateSettingsURL = "/outcomes/subscription-group-document-templates"
	routes.SubscriptionGroupDocumentTemplateUploadURL = "/outcomes/subscription-group-document-templates/upload"
	routes.SubscriptionGroupDocumentTemplatePublishURL = "/outcomes/subscription-group-document-templates/{id}/publish"
	routes.SubscriptionGroupDocumentTemplateDeleteURL = "/outcomes/subscription-group-document-templates/{id}/delete"

	tests := []struct {
		name          string
		entity        string
		exportEnabled bool
	}{
		{name: "zero"},
		{name: "grouped only", entity: outcomesummarypkg.ListEntitySubscriptionGroup},
		{name: "export only", exportEnabled: true},
		{name: "grouped and export", entity: outcomesummarypkg.ListEntitySubscriptionGroup, exportEnabled: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module := NewOutcomeSummaryModule(&OutcomeSummaryModuleDeps{
				Routes: routes,
				Labels: outcomesummarypkg.DefaultLabels(),
				Options: outcomesummarypkg.Options{
					List:                    outcomesummarypkg.ListOptions{Entity: tt.entity},
					SubscriptionGroupExport: outcomesummarypkg.SubscriptionGroupExportOptions{Enabled: tt.exportEnabled},
				},
			})
			recorder := newOutcomeSummaryRouteRecorder()
			module.RegisterRoutes(recorder)

			if !recorder.raw["GET "+routes.SubscriptionGroupExportURL] {
				t.Fatal("legacy subscription-group CSV route must remain mounted")
			}
			assertRouteState := func(method, path string, mounted bool) {
				t.Helper()
				var got bool
				switch method {
				case "GET":
					got = recorder.get[path]
				case "POST":
					got = recorder.post[path]
				default:
					t.Fatalf("unsupported method %q", method)
				}
				if got != mounted {
					t.Fatalf("%s %s mounted=%v, want %v", method, path, got, mounted)
				}
			}
			assertRouteState("GET", routes.SubscriptionGroupDownloadDrawerURL, tt.exportEnabled)
			assertRouteState("GET", routes.SubscriptionGroupDocumentTemplateSettingsURL, tt.exportEnabled)
			assertRouteState("GET", routes.SubscriptionGroupDocumentTemplateUploadURL, tt.exportEnabled)
			assertRouteState("POST", routes.SubscriptionGroupDocumentTemplateUploadURL, tt.exportEnabled)
			assertRouteState("POST", routes.SubscriptionGroupDocumentTemplatePublishURL, tt.exportEnabled)
			assertRouteState("POST", routes.SubscriptionGroupDocumentTemplateDeleteURL, tt.exportEnabled)
		})
	}
}

func TestOutcomeSummaryModule_ClientDrawerRequiresMountedDocumentHandler(t *testing.T) {
	routes := outcomesummarypkg.DefaultRoutes()
	routes.ClientDownloadDrawerURL = "/report-cards/section/{id}/student/{client_id}/download"

	withoutRenderer := NewOutcomeSummaryModule(&OutcomeSummaryModuleDeps{
		Routes:  routes,
		Labels:  outcomesummarypkg.DefaultLabels(),
		Options: outcomesummarypkg.Options{SubscriptionGroupExport: outcomesummarypkg.SubscriptionGroupExportOptions{Enabled: true}},
	})
	missingRendererRoutes := newOutcomeSummaryRouteRecorder()
	withoutRenderer.RegisterRoutes(missingRendererRoutes)
	if missingRendererRoutes.get[routes.ClientDownloadDrawerURL] {
		t.Fatal("client drawer route must not mount without its document handler")
	}

	withRenderer := NewOutcomeSummaryModule(&OutcomeSummaryModuleDeps{
		Routes:  routes,
		Labels:  outcomesummarypkg.DefaultLabels(),
		Options: outcomesummarypkg.Options{SubscriptionGroupExport: outcomesummarypkg.SubscriptionGroupExportOptions{Enabled: true}},
		GenerateDoc: func([]byte, map[string]any) ([]byte, error) {
			return []byte("docx"), nil
		},
	})
	mountedRoutes := newOutcomeSummaryRouteRecorder()
	withRenderer.RegisterRoutes(mountedRoutes)
	if !mountedRoutes.get[routes.ClientDownloadDrawerURL] {
		t.Fatal("client drawer route must mount when the document handler is mounted")
	}
}

// TestTemplateSettingsDeps_WiresPhaseManifestValidator locks the R2/C3
// (DEC-1a) wiring: template_settings' Deps.ValidatePhaseTemplate must be
// non-nil once the module builds it, and it must actually invoke the SAME
// subscription_group_document.ValidateTemplate contract the Section
// Templates upload path applies to the client-phase render profile (rather
// than, say, an always-nil or always-error stub).
func TestTemplateSettingsDeps_WiresPhaseManifestValidator(t *testing.T) {
	deps := templateSettingsDeps(&OutcomeSummaryModuleDeps{})
	if deps.ValidatePhaseTemplate == nil {
		t.Fatal("templateSettingsDeps must wire a non-nil ValidatePhaseTemplate closure")
	}
	// A well-formed but manifest-empty DOCX must be rejected by the wired
	// closure (proving it reaches the real manifest-token contract, not a
	// closure that always returns nil).
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, part := range []struct{ name, body string }{
		{"[Content_Types].xml", `<?xml version="1.0"?><Types/>`},
		{"word/document.xml", `<?xml version="1.0"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"/>`},
	} {
		w, err := zw.Create(part.name)
		if err != nil {
			t.Fatalf("zip create %q: %v", part.name, err)
		}
		if _, err := w.Write([]byte(part.body)); err != nil {
			t.Fatalf("zip write %q: %v", part.name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	if err := deps.ValidatePhaseTemplate(buf.Bytes()); err == nil {
		t.Fatal("the wired validator must reject a manifest-empty DOCX, not silently accept it")
	}
}

// TestSubscriptionGroupDocumentTemplateSettingsDeps_DropsWholeReportProfile
// locks the OTHER half of R2/C3 (DEC-1a): the client-phase render profile is
// no longer offered as an upload choice in the subscription_group_document_template settings. Building the
// settings Deps from module deps carrying a trusted WholeReportProfile must
// zero it in the copy handed to the settings view, while leaving the
// module-level deps.Options completely untouched (other consumers, e.g. the
// ClientDownloadDrawer's own gate, still read the original trusted value).
func TestSubscriptionGroupDocumentTemplateSettingsDeps_DropsWholeReportProfile(t *testing.T) {
	moduleDeps := &OutcomeSummaryModuleDeps{
		Options: outcomesummarypkg.Options{
			SubscriptionGroupExport: outcomesummarypkg.SubscriptionGroupExportOptions{
				Enabled:            true,
				WholeReportProfile: subscriptiongroupdocumenttemplatepb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_CLIENT_PHASE_OUTCOME_REPORT_V1,
			},
		},
	}
	settingsDeps := subscriptionGroupDocumentTemplateSettingsDeps(moduleDeps)
	if settingsDeps.Options.SubscriptionGroupExport.WholeReportProfile != subscriptiongroupdocumenttemplatepb.RenderProfile_RENDER_PROFILE_UNSPECIFIED {
		t.Fatalf("the subscription_group_document_template settings Deps must not offer the client-phase profile, got %v", settingsDeps.Options.SubscriptionGroupExport.WholeReportProfile)
	}
	if moduleDeps.Options.SubscriptionGroupExport.WholeReportProfile != subscriptiongroupdocumenttemplatepb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_CLIENT_PHASE_OUTCOME_REPORT_V1 {
		t.Fatalf("building the settings Deps must not mutate the module-level Options other consumers (e.g. the ClientDownloadDrawer gate) read, got %v", moduleDeps.Options.SubscriptionGroupExport.WholeReportProfile)
	}
	if !settingsDeps.Options.SubscriptionGroupExport.Enabled {
		t.Fatal("unrelated Options fields must still pass through unchanged")
	}
}

// TestOutcomeSummaryModule_WiresClassEdgeAndEligibilityIntoDocumentDeps
// (T-A6) proves the module actually threads ListSubscriptionGroupProductPlanStaffs,
// ListProductPlans, and ListProductPlanStaffs from OutcomeSummaryModuleDeps
// into documentview.Deps: before this wiring these three closures were
// declared but nothing ever assigned them, so fetchClassEdgeTeachers always
// saw a nil ListSubscriptionGroupProductPlanStaffs/ListProductPlans and the
// eligibility gate never ran. Driving a real ClientDocument download and
// observing all three stubs fire is the only way to prove the FULL chain
// (module -> newClientDocumentHandler -> document.Deps -> fetchClassEdgeTeachers)
// is connected, not just that the struct fields exist.
func TestOutcomeSummaryModule_WiresClassEdgeAndEligibilityIntoDocumentDeps(t *testing.T) {
	sp := func(s string) *string { return &s }
	var sgppsCalled, productPlansCalled, productPlanStaffsCalled bool

	moduleDeps := &OutcomeSummaryModuleDeps{
		Labels:      outcomesummarypkg.DefaultLabels(),
		GenerateDoc: func([]byte, map[string]any) ([]byte, error) { return []byte("DOCX"), nil },
		ListSubscriptionGroups: func(context.Context, *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error) {
			return &subscriptiongrouppb.ListSubscriptionGroupsResponse{Data: []*subscriptiongrouppb.SubscriptionGroup{
				{Id: "sec-1", Active: true, Name: "Grade 10 Gold (AY 2025-2026)"},
			}}, nil
		},
		ListSubscriptionGroupMembers: func(context.Context, *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error) {
			return &subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse{Data: []*subscriptiongroupmemberpb.SubscriptionGroupMember{
				{ClientId: "stu-1", SubscriptionId: "sub-1", Active: true},
			}}, nil
		},
		ListJobs: func(context.Context, *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error) {
			return &jobpb.ListJobsResponse{Data: []*jobpb.Job{{
				Id: "job-1", JobTemplateId: sp("tmpl-1"), OriginId: sp("sub-1"), Active: true,
				OriginType: enums.OriginType_ORIGIN_TYPE_SUBSCRIPTION, OutputProductId: sp("prod-1"),
			}}}, nil
		},
		ListJobTemplates: func(context.Context, *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error) {
			return &jobtemplatepb.ListJobTemplatesResponse{Data: []*jobtemplatepb.JobTemplate{{Id: "tmpl-1", Name: "Mathematics"}}}, nil
		},
		ListJobOutcomeSummarys: func(context.Context, *jobsumpb.ListJobOutcomeSummarysRequest) (*jobsumpb.ListJobOutcomeSummarysResponse, error) {
			return &jobsumpb.ListJobOutcomeSummarysResponse{Data: []*jobsumpb.JobOutcomeSummary{{JobId: "job-1", Active: true, ScaledLabel: sp("7")}}}, nil
		},
		ListClients: func(context.Context, *clientpb.ListClientsRequest) (*clientpb.ListClientsResponse, error) {
			return &clientpb.ListClientsResponse{Data: []*clientpb.Client{{Id: "stu-1", LastName: sp("Dela Cruz"), FirstName: sp("Juan")}}}, nil
		},
		// D5 render-gate reads (required, wired — an unwired gate fails closed).
		ListJobPhases: func(context.Context, *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
			return &jobphasepb.ListJobPhasesResponse{Success: true}, nil
		},
		ListJobTasks: func(context.Context, *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
			return &jobtaskpb.ListJobTasksResponse{Success: true}, nil
		},
		ListTaskOutcomes: func(context.Context, *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
			return &taskoutcomepb.ListTaskOutcomesResponse{Success: true}, nil
		},
		// The three closures under test. The edge links a product_plan_staff row
		// so the eligibility gate actually has something to fetch — a nil-link
		// edge would let fetchClassEdgeTeachers skip ListProductPlanStaffs
		// entirely and this test would pass for the wrong reason.
		ListSubscriptionGroupProductPlanStaffs: func(context.Context, *sgppspb.ListSubscriptionGroupProductPlanStaffsRequest) (*sgppspb.ListSubscriptionGroupProductPlanStaffsResponse, error) {
			sgppsCalled = true
			return &sgppspb.ListSubscriptionGroupProductPlanStaffsResponse{Data: []*sgppspb.SubscriptionGroupProductPlanStaff{
				{Id: "e-1", StaffId: "staff-1", Role: "primary", ProductPlanId: "pp-1", SubscriptionGroupId: "sec-1", Active: true, ProductPlanStaffId: sp("pps-1")},
			}}, nil
		},
		ListProductPlans: func(context.Context, *productplanpb.ListProductPlansRequest) (*productplanpb.ListProductPlansResponse, error) {
			productPlansCalled = true
			return &productplanpb.ListProductPlansResponse{Data: []*productplanpb.ProductPlan{{Id: "pp-1", ProductId: "prod-1"}}}, nil
		},
		ListProductPlanStaffs: func(context.Context, *productplanstaffpb.ListProductPlanStaffsRequest) (*productplanstaffpb.ListProductPlanStaffsResponse, error) {
			productPlanStaffsCalled = true
			return &productplanstaffpb.ListProductPlanStaffsResponse{Data: []*productplanstaffpb.ProductPlanStaff{
				{Id: "pps-1", Active: true},
			}}, nil
		},
	}

	module := NewOutcomeSummaryModule(moduleDeps)
	if module.ClientDocument == nil {
		t.Fatal("ClientDocument handler must be mounted when GenerateDoc is wired")
	}

	r := httptest.NewRequest(http.MethodGet, "/doc", nil)
	r.SetPathValue("id", "sec-1")
	r.SetPathValue("client_id", "stu-1")
	perms := types.NewUserPermissions([]string{"job_outcome_summary:list", "job_outcome_summary:read"})
	r = r.WithContext(view.WithUserPermissions(r.Context(), perms))

	w := httptest.NewRecorder()
	module.ClientDocument(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("download must succeed to prove the class-edge fetch actually ran end to end, got %d: %s", w.Code, w.Body.String())
	}
	if !sgppsCalled {
		t.Fatal("module must wire ListSubscriptionGroupProductPlanStaffs into document.Deps — fetchClassEdgeTeachers never called it")
	}
	if !productPlansCalled {
		t.Fatal("module must wire ListProductPlans into document.Deps — fetchClassEdgeTeachers never called it")
	}
	if !productPlanStaffsCalled {
		t.Fatal("module must wire ListProductPlanStaffs into document.Deps — the eligibility gate never called it")
	}
}
