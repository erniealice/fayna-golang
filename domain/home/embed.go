package home

import "embed"

// TemplatesFS embeds the home surface templates. Template names are
// package-prefixed ("fayna-home", "fayna-home-content", "fayna-home-ribbon")
// so they can coexist in a renderer that also parses an app-local home
// template set (the Q6-B fork-discipline contract: the app-local module and
// its templates are NOT edited).
//
//go:embed templates/*.html
var TemplatesFS embed.FS
