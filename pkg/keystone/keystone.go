package keystone

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"sync"
)

type Registry struct {
	Source        fs.ReadDirFS
	FuncMap       template.FuncMap
	Reload        bool
	baseTemplate  *template.Template
	templateCache map[string]*template.Template
	mu            sync.RWMutex
}

func New(templateFS fs.ReadDirFS) (*Registry, error) {
	ks := &Registry{
		Source:        templateFS,
		baseTemplate:  &template.Template{},
		templateCache: make(map[string]*template.Template),
		FuncMap:       template.FuncMap{},
		Reload:        false,
	}

	err := ks.Load()
	if err != nil {
		return ks, err
	}

	return ks, nil
}

func (ks *Registry) Load() error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	ks.templateCache = make(map[string]*template.Template)

	templateFiles, err := ks.allFilesInPath()
	if err != nil {
		return err
	}

	base, err := template.New("").Funcs(ks.FuncMap).ParseFS(ks.Source, templateFiles...)
	if err != nil {
		return err
	}

	ks.baseTemplate = base

	err = ks.insertTemplates(".")
	return err
}

func (ks *Registry) allFilesInPath() ([]string, error) {
	fileList := []string{}
	err := fs.WalkDir(ks.Source, ".", func(path string, d fs.DirEntry, err error) error {
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

func (ks *Registry) insertTemplates(path string) error {
	dirContents, err := ks.Source.ReadDir(path)
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
			if err := ks.insertTemplates(fullPath); err != nil {
				return err
			}
			continue
		}

		bc, err := ks.baseTemplate.Clone()
		if err != nil {
			return err
		}

		t, err := bc.ParseFS(ks.Source, fullPath)
		if err != nil {
			return fmt.Errorf("could not parse %v, %v", fullPath, err)
		}
		ks.templateCache[fullPath] = t
	}

	return nil
}

func (ks *Registry) Get(name string) (*template.Template, error) {
	if ks.Reload {
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

func (ks *Registry) Exists(name string) bool {
	tmpl, err := ks.Get(name)
	return err == nil && tmpl != nil
}

func (ks *Registry) Render(w io.Writer, name string, data any) error {
	tmpl, err := ks.Get(name)
	if err != nil {
		return fmt.Errorf("could not render %v, %v", name, err)
	}
	if tmpl == nil {
		return fmt.Errorf("could not render %v, template is missing, known templates: %v", name, ks.ListAll())
	}
	return tmpl.ExecuteTemplate(w, filepath.Base(name), data)
}

func (ks *Registry) ListAll() []string {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	names := []string{}

	for name := range ks.templateCache {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}
