package testdata

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed fixtures/* layouts/* components/* pages/* partials/*
var TestFixtures embed.FS

//go:embed layouts/* components/* partials/*
var TestBaseTemplatesFS embed.FS

//go:embed pages/*
var TestPagesTemplatesFS embed.FS

func CreateTempFilesystem(t *testing.T) (string, fs.ReadDirFS, fs.ReadDirFS) {
	tempDir, err := os.MkdirTemp("", "keystone-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	baseDir := filepath.Join(tempDir, "base")
	pagesDir := filepath.Join(tempDir, "pages")

	err = os.MkdirAll(baseDir, 0o755)
	require.NoError(t, err)
	err = os.MkdirAll(pagesDir, 0o755)
	require.NoError(t, err)

	layoutContent := `{{ define "layouts/default.tmpl" }}
<!DOCTYPE html>
<html>
<head>
  <title>{{ .Title }}</title>
</head>
<body>
  {{ template "content" . }}
</body>
</html>
{{ end }}`

	err = os.WriteFile(filepath.Join(baseDir, "default.tmpl"), []byte(layoutContent), 0o644)
	require.NoError(t, err)

	productContent := `{{ define "content" }}
<div>
  <h1>{{ .Name }}</h1>
  <p>{{ .Description }}</p>
  <p>In stock: {{ .Stock }}</p>
  <p>Price: {{ .Price }}</p>
</div>
{{ end }}`

	err = os.WriteFile(filepath.Join(pagesDir, "product.tmpl"), []byte(productContent), 0o644)
	require.NoError(t, err)

	baseFS := NewTestFS(baseDir)
	pagesFS := NewTestFS(pagesDir)

	return tempDir, baseFS, pagesFS
}

func UpdateTemplate(t *testing.T, tempDir, relativePath, newContent string) {
	fullPath := filepath.Join(tempDir, relativePath)
	err := os.WriteFile(fullPath, []byte(newContent), 0o644)
	require.NoError(t, err)
}
