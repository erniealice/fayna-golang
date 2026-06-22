package reporting_checkpoint

import "embed"

//go:embed templates/*.html
var TemplatesFS embed.FS
