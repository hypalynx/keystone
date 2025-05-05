package testdata

import "embed"

//go:embed layouts/*.tmpl components/*.tmpl components/**/* partials/*.tmpl
var TestBaseTemplateFS embed.FS

//go:embed pages/*.tmpl pages/**/*.tmpl
var TestPagesTemplateFS embed.FS
