package keystone

import (
	"html/template"
	"io/fs"
	"sync"
)

type Keystone struct {
	source          fs.ReadDirFS
	pagesSource     fs.ReadDirFS
	baseTemplate    *template.Template
	templateCache   map[string]*template.Template
	templateFuncMap template.FuncMap
	mu              sync.RWMutex
	hotReload       bool
}

func New(baseTemplateFS fs.ReadDirFS, pagesTemplateFS fs.ReadDirFS, templateFns template.FuncMap) (*Keystone, error) {
	ks := &Keystone{
		source:          baseTemplateFS,
		pagesSource:     pagesTemplateFS,
		baseTemplate:    nil,
		templateCache:   make(map[string]*template.Template),
		templateFuncMap: templateFns,
		hotReload:       false,
	}

	err := ks.Load()
	if err != nil {
		return ks, err
	}

	return ks, nil
}

func NewWithHotReload(baseTemplateFS fs.ReadDirFS, pagesTemplateFS fs.ReadDirFS, templateFns template.FuncMap) (*Keystone, error) {
	ks, err := New(baseTemplateFS, pagesTemplateFS, templateFns)
	if err != nil {
		return ks, err
	}

	ks.hotReload = true
	return ks, nil
}

func (ks *Keystone) Load() error {
	base, err := template.New("").Funcs(ks.templateFuncMap).ParseFS(ks.source, "**/*.tmpl")
	if err != nil {
		return err
	}

	ks.baseTemplate = base
	return nil
}

// func (ks *Keystone) insertTemplates(templates *template.Template, base *template.Template) error {
// }

func (ks *Keystone) Exists(name string) bool {
	_, exists := ks.templateCache[name]
	if !exists {
		tmpl := ks.baseTemplate.Lookup(name)
		if tmpl != nil {
			return true
		}
	}
	return exists
}

// read in
// provide Render interface
