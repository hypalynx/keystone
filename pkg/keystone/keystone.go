package keystone

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
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
	ks.mu.Lock()
	defer ks.mu.Unlock()

	templateFiles, err := ks.allFilesInPath()
	if err != nil {
		return err
	}

	base, err := template.New("").Funcs(ks.templateFuncMap).ParseFS(ks.source, templateFiles...)
	if err != nil {
		return err
	}

	ks.baseTemplate = base

	err = ks.insertPageTemplates(".")
	return err
}

func (ks *Keystone) allFilesInPath() ([]string, error) {
	fileList := []string{}
	err := fs.WalkDir(ks.source, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		return fileList, err
	}

	return fileList, nil
}

func (ks *Keystone) insertPageTemplates(path string) error {
	dirContents, err := ks.pagesSource.ReadDir(path)
	if err != nil {
		return err
	}

	for _, e := range dirContents {
		fullPath := path
		if path != "." {
			fullPath = path + "/" + e.Name()
		} else {
			fullPath = e.Name()
		}

		if e.IsDir() {
			if err := ks.insertPageTemplates(fullPath); err != nil {
				return err
			}
			continue
		}

		bc, err := ks.baseTemplate.Clone()
		if err != nil {
			return err
		}

		t, err := bc.ParseFS(ks.pagesSource, fullPath)
		if err != nil {
			return fmt.Errorf("could not parse %v, %v", fullPath, err)
		}
		ks.templateCache[fullPath] = t
	}

	return nil
}

func (ks *Keystone) Get(name string) (*template.Template, error) {
	if ks.reload {
		err := ks.Load()
		if err != nil {
			return nil, err
		}
	}

	ks.mu.RLock()
	defer ks.mu.RUnlock()
	tmpl, exists := ks.templateCache[name]
	if !exists {
		return ks.baseTemplate.Lookup(name), nil
	}
	return tmpl, nil
}

func (ks *Keystone) Exists(name string) bool {
	tmpl, err := ks.Get(name)
	return err == nil && tmpl != nil
}

func (ks *Keystone) Render(w io.Writer, name string, data any) error {
	tmpl, err := ks.Get(name)
	if err != nil {
		return fmt.Errorf("could not render %v, %v", name, err)
	}
	if tmpl == nil {
		return fmt.Errorf("could not render %v, template is missing, known templates: %v", name, ks.ListAll())
	}
	return tmpl.ExecuteTemplate(w, filepath.Base(name), data)
}

func (ks *Keystone) ListAll() []string {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	names := []string{}

	for _, tmpl := range ks.baseTemplate.Templates() {
		name := tmpl.Name()
		if name != "" && name != "content" && strings.Contains(name, "/") {
			names = append(names, name)
		}
	}

	for name := range ks.templateCache {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}
