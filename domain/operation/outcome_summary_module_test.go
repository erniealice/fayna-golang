package operation

import (
	"net/http"
	"testing"

	outcomesummarypkg "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/view"
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
