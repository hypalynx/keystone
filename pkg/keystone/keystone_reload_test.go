package keystone_test

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"text/template"
	"time"

	"github.com/hypalynx/keystone/pkg/keystone"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestFS struct {
	rootDir string
}

func NewTestFS(rootDir string) *TestFS {
	return &TestFS{rootDir: rootDir}
}

func (tfs *TestFS) Open(name string) (fs.File, error) {
	fullPath := filepath.Join(tfs.rootDir, name)
	fmt.Printf("Opening file: %s\n", fullPath)
	f, err := os.Open(fullPath)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", fullPath, err)
	}
	return f, err
}

func (tfs *TestFS) ReadDir(name string) ([]fs.DirEntry, error) {
	fullPath := filepath.Join(tfs.rootDir, name)
	fmt.Printf("Reading dir: %s\n", fullPath)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		fmt.Printf("Error reading dir %s: %v\n", fullPath, err)
	} else {
		fmt.Printf("Found %d entries in %s\n", len(entries), fullPath)
		for i, entry := range entries {
			fmt.Printf("  %d: %s (isDir: %v)\n", i, entry.Name(), entry.IsDir())
		}
	}
	return entries, err
}

type KeystoneReloadTestSuite struct {
	suite.Suite
	tempDir  string
	baseDir  string
	pagesDir string
}

func (s *KeystoneReloadTestSuite) SetupTest() {
	tempDir, err := os.MkdirTemp("", "keystone-reload-test")
	require.NoError(s.T(), err)
	s.tempDir = tempDir
	fmt.Printf("Created temp dir: %s\n", tempDir)

	s.baseDir = filepath.Join(tempDir, "base")
	s.pagesDir = filepath.Join(tempDir, "pages")

	layoutsDir := filepath.Join(s.baseDir, "layouts")
	err = os.MkdirAll(layoutsDir, 0o755)
	require.NoError(s.T(), err)
	fmt.Printf("Created layouts dir: %s\n", layoutsDir)

	err = os.MkdirAll(s.pagesDir, 0o755)
	require.NoError(s.T(), err)
	fmt.Printf("Created pages dir: %s\n", s.pagesDir)

	layoutContent := `{{ define "layouts/default.tmpl" }}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .Title }}</title>
</head>
<body>
    {{ block "content" . }}
    This appears when no "content" is provided.
    {{ end }}
</body>
</html>
{{ end }}`

	err = os.WriteFile(filepath.Join(layoutsDir, "default.tmpl"), []byte(layoutContent), 0o644)
	require.NoError(s.T(), err)
	fmt.Printf("Created default.tmpl\n")

	productContent := `{{ template "layouts/default.tmpl" . }}
{{ define "content" }}
<div>
  <h1>{{ .Name }}</h1>
  <p>{{ .Description }}</p>
  <p>In stock: {{ .Stock }}</p>
  <p>Price: {{ .Price }}</p>
</div>
{{ end }}`

	err = os.WriteFile(filepath.Join(s.pagesDir, "product.tmpl"), []byte(productContent), 0o644)
	require.NoError(s.T(), err)
	fmt.Printf("Created product.tmpl\n")
}

func (s *KeystoneReloadTestSuite) TearDownTest() {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
		fmt.Printf("Removed temp dir: %s\n", s.tempDir)
	}
}

func (s *KeystoneReloadTestSuite) TestReload() {
	baseFS := NewTestFS(s.baseDir)
	pagesFS := NewTestFS(s.pagesDir)

	ks, err := keystone.NewWithReload(baseFS, pagesFS, template.FuncMap{})
	require.NoError(s.T(), err)

	templates := ks.ListAll()
	fmt.Printf("Available templates after initialization: %v\n", templates)

	var initialOutput bytes.Buffer
	err = ks.Render(&initialOutput, "product.tmpl", map[string]any{
		"Title":       "Product Details",
		"Name":        "Pen",
		"Description": "This is a pen, you can write with it!",
		"Stock":       7,
		"Price":       "£8.99",
	})
	require.NoError(s.T(), err)

	initialOutputStr := initialOutput.String()
	fmt.Printf("Initial output:\n%s\n", initialOutputStr)
	require.NotEmpty(s.T(), initialOutputStr, "Initial template output should not be empty")

	modifiedContent := `{{ template "layouts/default.tmpl" . }}
{{ define "content" }}
<div class="product">
  <h2>{{ .Name }}</h2>
  <p class="description">{{ .Description }}</p>
  <p class="stock">Available: {{ .Stock }} units</p>
  <p class="price">Price: {{ .Price }}</p>
  <button>Add to Cart</button>
</div>
{{ end }}`

	err = os.WriteFile(filepath.Join(s.pagesDir, "product.tmpl"), []byte(modifiedContent), 0o644)
	require.NoError(s.T(), err)
	fmt.Printf("Modified product.tmpl\n")

	time.Sleep(100 * time.Millisecond)

	templates = ks.ListAll()
	fmt.Printf("Available templates after modification: %v\n", templates)

	var modifiedOutput bytes.Buffer
	err = ks.Render(&modifiedOutput, "product.tmpl", map[string]any{
		"Title":       "Product Details",
		"Name":        "Pen",
		"Description": "This is a pen, you can write with it!",
		"Stock":       7,
		"Price":       "£8.99",
	})
	require.NoError(s.T(), err)

	modifiedOutputStr := modifiedOutput.String()
	fmt.Printf("Modified output:\n%s\n", modifiedOutputStr)
	require.NotEmpty(s.T(), modifiedOutputStr, "Modified template output should not be empty")

	s.NotEqual(initialOutputStr, modifiedOutputStr)

	s.Contains(modifiedOutputStr, `<div class="product">`)
	s.Contains(modifiedOutputStr, `<button>Add to Cart</button>`)
}

func (s *KeystoneReloadTestSuite) TestNoReload() {
	baseFS := NewTestFS(s.baseDir)
	pagesFS := NewTestFS(s.pagesDir)

	ks, err := keystone.New(baseFS, pagesFS, template.FuncMap{})
	require.NoError(s.T(), err)

	templates := ks.ListAll()
	fmt.Printf("Available templates after initialization: %v\n", templates)

	var initialOutput bytes.Buffer
	err = ks.Render(&initialOutput, "product.tmpl", map[string]any{
		"Title":       "Product Details",
		"Name":        "Pen",
		"Description": "This is a pen, you can write with it!",
		"Stock":       7,
		"Price":       "£8.99",
	})
	require.NoError(s.T(), err)

	initialOutputStr := initialOutput.String()
	fmt.Printf("Initial output:\n%s\n", initialOutputStr)
	require.NotEmpty(s.T(), initialOutputStr, "Initial template output should not be empty")

	modifiedContent := `{{ template "layouts/default.tmpl" . }}
{{ define "content" }}
<div class="product-modified">
  <h2>MODIFIED TEMPLATE</h2>
  <p>{{ .Name }} - {{ .Description }}</p>
  <p>Stock: {{ .Stock }} | Price: {{ .Price }}</p>
</div>
{{ end }}`

	err = os.WriteFile(filepath.Join(s.pagesDir, "product.tmpl"), []byte(modifiedContent), 0o644)
	require.NoError(s.T(), err)
	fmt.Printf("Modified product.tmpl\n")

	time.Sleep(100 * time.Millisecond)

	templates = ks.ListAll()
	fmt.Printf("Available templates after modification: %v\n", templates)

	var secondOutput bytes.Buffer
	err = ks.Render(&secondOutput, "product.tmpl", map[string]any{
		"Title":       "Product Details",
		"Name":        "Pen",
		"Description": "This is a pen, you can write with it!",
		"Stock":       7,
		"Price":       "£8.99",
	})
	require.NoError(s.T(), err)

	secondOutputStr := secondOutput.String()
	fmt.Printf("Second output:\n%s\n", secondOutputStr)
	require.NotEmpty(s.T(), secondOutputStr, "Second template output should not be empty")

	s.Equal(initialOutputStr, secondOutputStr)

	s.NotContains(secondOutputStr, "MODIFIED TEMPLATE")
}

func TestKeystoneReloadSuite(t *testing.T) {
	suite.Run(t, new(KeystoneReloadTestSuite))
}
