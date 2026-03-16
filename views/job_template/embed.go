package job_template

import "embed"

//go:embed templates/*.html
var TemplatesFS embed.FS
