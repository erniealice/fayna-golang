package job

import "github.com/erniealice/pyeza-golang/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.job",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "job"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "job.json", Key: "job"},
		LabelName: "JobLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "job:list",
			AppEntry: &compose.AppEntry{
				Key:        "job",
				Route:      "job.list",
				Label:      "Operations",
				Icon:       "icon-briefcase",
				Permission: "job:list",
			},
			Items: []compose.NavItem{
				{Key: "jobs-draft", Route: "job.list", Params: map[string]string{"status": "draft"}, Label: "Draft", Icon: "icon-edit"},
				{Key: "jobs-active", Route: "job.list", Params: map[string]string{"status": "active"}, Label: "Active", Icon: "icon-briefcase"},
				{Key: "jobs-on-hold", Route: "job.list", Params: map[string]string{"status": "paused"}, Label: "On Hold", Icon: "icon-pause-circle"},
				{Key: "jobs-completed", Route: "job.list", Params: map[string]string{"status": "completed"}, Label: "Complete", Icon: "icon-check-circle"},
			},
		},
	}
}
