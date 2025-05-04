package keystone

import (
	"fmt"
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
	reload          bool
}

func New(baseTemplateFS fs.ReadDirFS, pagesTemplateFS fs.ReadDirFS, templateFns template.FuncMap) (*Keystone, error) {
	ks := &Keystone{
		source:          baseTemplateFS,
		pagesSource:     pagesTemplateFS,
		baseTemplate:    nil,
		templateCache:   make(map[string]*template.Template),
		templateFuncMap: templateFns,
		reload:          false,
	}

	err := ks.Load()
	if err != nil {
		return ks, err
	}

	return ks, nil
}

func NewWithReload(baseTemplateFS fs.ReadDirFS, pagesTemplateFS fs.ReadDirFS, templateFns template.FuncMap) (*Keystone, error) {
	ks, err := New(baseTemplateFS, pagesTemplateFS, templateFns)
	if err != nil {
		return ks, err
	}

	ks.reload = true
	return ks, nil
}

func (ks *Keystone) Load() error {
	base, err := template.New("").Funcs(ks.templateFuncMap).ParseFS(ks.source, "**/*.tmpl")
	if err != nil {
		return err
	}

	ks.baseTemplate = base

	err = ks.insertPageTemplates("pages")
	fmt.Println(ks.templateCache)
	return err
}

func (ks *Keystone) insertPageTemplates(path string) error {
	dirContents, err := ks.pagesSource.ReadDir(path)
	if err != nil {
		return err
	}

	for _, e := range dirContents {
		if e.IsDir() {
			continue // recurse here for subdir support
		}

		bc, err := ks.baseTemplate.Clone()
		if err != nil {
			return err
		}

		tn := e.Name()
		t, err := bc.ParseFS(ks.pagesSource, path+"/"+tn) // recursed dirpath here
		if err != nil {
			return fmt.Errorf("could not parse %v, %v", tn, err)
		}
		ks.templateCache[path+"/"+e.Name()] = t
	}

	return nil
}

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
