package keystone

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type Registry struct {
	Debug            bool
	Extensions       []string
	FuncMap          template.FuncMap
	Reload           bool
	Source           fs.FS
	baseTemplate     *template.Template
	extensionsLookup map[string]bool
	mu               sync.RWMutex
	templateCache    map[string]*template.Template
}

var defaultExtensions = []string{"tmpl", "html", "gohtml", "gotmpl", "tpl"}

func New(templateFS fs.FS) (*Registry, error) {
	ks := &Registry{
		Debug:            false,
		Extensions:       defaultExtensions,
		FuncMap:          template.FuncMap{},
		Reload:           false,
		Source:           templateFS,
		baseTemplate:     &template.Template{},
		extensionsLookup: make(map[string]bool),
		templateCache:    make(map[string]*template.Template),
	}

	err := ks.Load()
	if err != nil {
		return ks, err
	}

	return ks, nil
}

func (ks *Registry) ensureDefaults() error {
	if ks.Source == nil {
		return fmt.Errorf("no keystone.Source provided to source templates from")
	}

	if ks.Extensions == nil {
		ks.Extensions = []string{"tmpl", "html", "gohtml", "gotmpl", "tpl"}
	}

	extMap := make(map[string]bool)
	for _, ext := range ks.Extensions {
		extMap[ext] = true
	}
	ks.extensionsLookup = extMap

	ks.templateCache = make(map[string]*template.Template)

	return nil
}

func (ks *Registry) Load() error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	err := ks.ensureDefaults()
	if err != nil {
		return err
	}

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

func (ks *Registry) isTemplate(path string) bool {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	return len(ks.Extensions) == 0 || ks.extensionsLookup[ext]
}

func (ks *Registry) allFilesInPath() ([]string, error) {
	fileList := []string{}

	err := fs.WalkDir(ks.Source, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			if ks.isTemplate(path) {
				fileList = append(fileList, path)
			}
		}
		return nil
	})
	if err != nil {
		return fileList, err
	}

	return fileList, nil
}

func (ks *Registry) insertTemplates(path string) error {
	dirContents, err := fs.ReadDir(ks.Source, path)
	if err != nil {
		return err
	}
	for _, e := range dirContents {
		var fullPath string
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
		if !ks.isTemplate(fullPath) {
			continue
		}

		bc, err := ks.baseTemplate.Clone()
		if err != nil {
			return err
		}

		namedTemplate := bc.New(fullPath)

		content, err := fs.ReadFile(ks.Source, fullPath)
		if err != nil {
			return fmt.Errorf("could not read template file %v: %v", fullPath, err)
		}

		_, err = namedTemplate.Parse(string(content))
		if err != nil {
			return fmt.Errorf("could not parse %v: %v", fullPath, err)
		}

		ks.templateCache[fullPath] = bc
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
	if ks.Debug {
		keys := []string{}
		for _, t := range tmpl.Templates() {
			keys = append(keys, t.Name())
		}
		slog.Info("Rendering template", "name", name, "keys", keys)
	}

	return tmpl.ExecuteTemplate(w, name, data)
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
