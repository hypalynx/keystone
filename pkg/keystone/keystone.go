package keystone

import (
	"fmt"
	"html/template"
	"io/fs"
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
	templateFiles, err := ks.allFilesInPath()
	if err != nil {
		return err
	}

	base, err := template.New("").Funcs(ks.templateFuncMap).ParseFS(ks.source, templateFiles...)
	if err != nil {
		return err
	}

	ks.baseTemplate = base

	err = ks.insertPageTemplates("pages")
	fmt.Println(ks.templateCache)
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
		fmt.Printf("Error walking directory: %v\n", err)
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
		if e.IsDir() {
			if err := ks.insertPageTemplates(path + "/" + e.Name()); err != nil {
				return err
			}
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

func (ks *Keystone) ListAll() []string {
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

// read in
// provide Render interface
