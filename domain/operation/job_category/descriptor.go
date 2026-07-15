package job_category

import "github.com/erniealice/espyna-golang/consumer/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.job_category",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "job_category"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "job_category.json", Key: "job_category"},
		LabelName: "JobCategoryLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "job_category:list",
			Items: []compose.NavItem{
				{Key: "job-categories-active", Route: "job_category.list", Params: map[string]string{"status": "active"}, Label: "Active", Icon: "icon-layers", Permission: "job_category:list", LabelKey: "job_category_active_label", IconKey: "job_category_icon"},
				{Key: "job-categories-inactive", Route: "job_category.list", Params: map[string]string{"status": "inactive"}, Label: "Inactive", Icon: "icon-circle", Permission: "job_category:list", LabelKey: "job_category_inactive_label", IconKey: "job_category_inactive_icon"},
			},
		},
	}
}
