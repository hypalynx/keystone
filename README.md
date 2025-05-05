# :gem: Keystone [![Go Version](https://img.shields.io/github/go-mod/go-version/hypalynx/keystone)](https://github.com/hypalynx/keystone) [![Go Report Card](https://goreportcard.com/badge/github.com/hypalynx/keystone)](https://goreportcard.com/report/github.com/hypalynx/keystone)

A small Golang library that makes working with the stdlib `html/template` easy and safe.

## Installation

```bash
# For the latest
go get github.com/hypalynx/keystone@latest

# Or pin by version (also latest released)
go get github.com/hypalynx/keystone@v0.1.0
```

## Features

- **Hot-reloading** for development, embedded templates for production
- **Organized templates** in layouts, pages, components, and partials
- **Template inheritance** with content blocks
- **Thread-safe** for concurrent HTTP requests

## Quick Start

```go
package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/hypalynx/keystone"
)

//go:embed templates/*
var templateFS embed.FS

func main() {
	// Simple initialization
	ks, err := keystone.New(templateFS)
	if err != nil {
		log.Fatalf("Couldn't initialize templates: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"Title":   "Hello, Keystone!",
			"Message": "A simple template system for Go",
		}
		
		if err := ks.Render(w, "pages/home.tmpl", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
```

## Advanced Configuration

For more control, use the `Registry` directly:

```go
package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/hypalynx/keystone"
)

//go:embed templates/*
var templateFS embed.FS

func main() {
	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
		// Add other helper functions here
	}

	// Check if we're in development mode
	isDev := os.Getenv("GO_ENV") == "development"
	
	var ks *keystone.Registry
	var err error
	
	if isDev {
		// Development: Use disk-based FS with hot-reloading
		log.Println("Running in development mode with hot-reloading")
		diskFS := os.DirFS("./templates")
		
		ks = &keystone.Registry{
			Source:  diskFS,
			Reload:  true,  // Enable hot-reloading
			FuncMap: funcMap,
		}
	} else {
		// Production: Use embedded FS without reloading
		log.Println("Running in production mode with embedded templates")
		templatesFS, _ := fs.Sub(embeddedFS, "templates")
		
		ks = &keystone.Registry{
			Source:  templatesFS,
			Reload:  false,  // Disable reloading for performance
			FuncMap: funcMap,
		}
	}
	
	if err := ks.Load(); err != nil {
		log.Fatalf("Couldn't initialize template registry: %v", err)
	}

	// Rest of your HTTP server code...
}
```

## Template Structure

```
templates/
├── layouts/
│   └── default.tmpl          # Base layout with content blocks
├── pages/
│   ├── home.tmpl             # Pages that extend layouts
│   └── about.tmpl
├── components/
│   ├── header.tmpl           # Reusable UI components
│   └── product/
│       └── card.tmpl
└── partials/
    └── meta.tmpl             # Small reusable parts
```

### Example Layout

```html
{{ define "layouts/default.tmpl" }}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>{{ .Title }}</title>
</head>
<body>
    <main>
        {{ block "content" . }}
        <p>Default content</p>
        {{ end }}
    </main>
</body>
</html>
{{ end }}
```

### Example Page

```html
{{ template "layouts/default.tmpl" . }}

{{ define "content" }}
<div class="home-page">
    <h2>{{ .Title }}</h2>
    <p>{{ .Message }}</p>
</div>
{{ end }}
```

## API Reference

- `New(fs fs.ReadDirFS) (*Registry, error)` - Create new template registry
- `Render(w io.Writer, name string, data any) error` - Render template with data
- `Get(name string) (*template.Template, error)` - Get a template by name
- `ListAll() []string` - List all available templates
- `Exists(name string) bool` - Check if a template exists
- `Load() error` - Load or reload all templates

## Framework Integration

```go
// With chi
r := chi.NewRouter()
r.Get("/", func(w http.ResponseWriter, r *http.Request) {
    ks.Render(w, "pages/home.tmpl", data)
})

// With gin
g := gin.Default()
g.GET("/", func(c *gin.Context) {
    ks.Render(c.Writer, "pages/home.tmpl", data)
})

// With echo
type EchoRenderer struct {
    Registry *keystone.Registry
}

func (r *EchoRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
    return r.Registry.Render(w, name, data)
}

// Usage:
e := echo.New()
e.Renderer = &EchoRenderer{Registry: ks}
```

## Editor Setup

For a better experience with Go templates in Neovim:

- Use neovim with `gotmpl` and `html` treesitter plugins configured with this injection, which allows both to work together in the same file:
```bash
mkdir -p "$NEOVIM_PATH/queries/gotmpl" # probably ~/.config/nvim
echo '((text) @injection.content
 (#set! injection.language "html")
 (#set! injection.combined))' > "$NEOVIM_PATH/queries/gotmpl/injections.scm"
```
- Treesitter also supports formatting the file which you can configure to do on save.
- Use neovim with the `html`, `htmx` & `tailwindcss` LSPs too.

## License

MIT
