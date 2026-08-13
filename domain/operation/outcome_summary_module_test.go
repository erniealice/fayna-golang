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

func TestOutcomeSummaryModule_SectionTemplateRoutesRequireTypedEnablement(t *testing.T) {
	routes := outcomesummarypkg.DefaultRoutes()
	routes.SectionDownloadDrawerURL = "/outcomes/section/{id}/download"
	routes.SectionTemplateSettingsURL = "/outcomes/section-templates"
	routes.SectionTemplateUploadURL = "/outcomes/section-templates/upload"
	routes.SectionTemplatePublishURL = "/outcomes/section-templates/{id}/publish"
	routes.SectionTemplateDeleteURL = "/outcomes/section-templates/{id}/delete"

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
					List:          outcomesummarypkg.ListOptions{Entity: tt.entity},
					SectionExport: outcomesummarypkg.SectionExportOptions{Enabled: tt.exportEnabled},
				},
			})
			recorder := newOutcomeSummaryRouteRecorder()
			module.RegisterRoutes(recorder)

			if !recorder.raw["GET "+routes.SectionExportURL] {
				t.Fatal("legacy section CSV route must remain mounted")
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
			assertRouteState("GET", routes.SectionDownloadDrawerURL, tt.exportEnabled)
			assertRouteState("GET", routes.SectionTemplateSettingsURL, tt.exportEnabled)
			assertRouteState("GET", routes.SectionTemplateUploadURL, tt.exportEnabled)
			assertRouteState("POST", routes.SectionTemplateUploadURL, tt.exportEnabled)
			assertRouteState("POST", routes.SectionTemplatePublishURL, tt.exportEnabled)
			assertRouteState("POST", routes.SectionTemplateDeleteURL, tt.exportEnabled)
		})
	}
}
