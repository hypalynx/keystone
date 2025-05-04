package keystone

import (
	"io/fs"
	"sync"
	"text/template"
)

type Keystone struct {
	source          fs.ReadDirFS
	pagesSource     fs.ReadDirFS
	templateCache   map[string]*template.Template
	mu              sync.RWMutex
	developmentMode bool
}

func New(baseTemplateFS fs.ReadDirFS, pagesTemplateFS fs.ReadDirFS) *Keystone {
	return &Keystone{
		source:          baseTemplateFS,
		pagesSource:     pagesTemplateFS,
		templateCache:   make(map[string]*template.Template),
		developmentMode: false, // TODO do we need this?
	}
}

func (ks *Keystone) Exists(name string) bool {
	_, exists := ks.templateCache[name]
	return exists
}

// read in
// provide Render interface
