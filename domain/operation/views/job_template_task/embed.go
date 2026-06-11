package job_template_task

import "embed"

//go:embed templates/*.html
var TemplatesFS embed.FS
