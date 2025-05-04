package testdata

import "embed"

//go:embed layouts/*.tmpl components/*.tmpl partials/*.tmpl
var TestBaseTemplateFS embed.FS

//go:embed pages/*.tmpl
var TestPagesTemplateFS embed.FS
