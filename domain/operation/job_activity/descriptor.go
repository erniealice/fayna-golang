package job_activity

import "github.com/erniealice/pyeza-golang/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.job_activity",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "job_activity"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "job_activity.json", Key: "job_activity"},
		LabelName: "JobActivityLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "job_activity:list",
			Items: []compose.NavItem{
				{Key: "te-drafts", Route: "job_activity.list", Query: "?status=draft", Label: "Drafts", Icon: "icon-edit", Permission: "job_activity:list"},
				{Key: "te-pending", Route: "job_activity.list", Query: "?status=submitted", Label: "Awaiting Approval", Icon: "icon-clock", Permission: "job_activity:list"},
				{Key: "te-approved", Route: "job_activity.list", Query: "?status=approved", Label: "Approved", Icon: "icon-check", Permission: "job_activity:list"},
				{Key: "te-posted", Route: "job_activity.list", Query: "?status=posted", Label: "Posted", Icon: "icon-check-circle", Permission: "job_activity:list"},
			},
		},
	}
}
