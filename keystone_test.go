package keystone_test

import (
	"bytes"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/hypalynx/keystone"
	"github.com/hypalynx/keystone/testdata"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type KeystoneTestSuite struct {
	suite.Suite
	keystone *keystone.Registry
}

func (s *KeystoneTestSuite) SetupTest() {
	ks, err := keystone.New(testdata.TestTemplatesFS)
	require.NoError(s.T(), err)
	s.keystone = ks
}

func (s *KeystoneTestSuite) TestTemplateIngestion() {
	s.Equal(
		[]string{
			"components/card.tmpl",
			"components/catalog/description.tmpl",
			"components/catalog/non-tmpl-ext-template.html",
			"layouts/default.tmpl",
			"pages/catalog/product.tmpl",
			"pages/test.tmpl",
			"partials/results.tmpl",
		},
		s.keystone.ListAll(),
	)

	s.True(s.keystone.Exists("pages/test.tmpl"))
}

func (s *KeystoneTestSuite) TestTemplateRetrival() {
	tmpl, err := s.keystone.Get("pages/catalog/product.tmpl")
	s.NoError(err)
	var dst bytes.Buffer
	require.NoError(s.T(), tmpl.ExecuteTemplate(&dst, filepath.Base("pages/catalog/product.tmpl"), map[string]any{
		"Name":        "Pen",
		"Description": "This is a pen, you can write with it!",
		"Stock":       7,
		"Price":       "£8.99",
	}))
	expected, err := testdata.TestFixtures.ReadFile("fixtures/pen.html")
	require.NoError(s.T(), err)
	s.Equal(string(expected), dst.String())
}

func (s *KeystoneTestSuite) TestTemplateRender() {
	var dst bytes.Buffer
	require.NoError(s.T(), s.keystone.Render(&dst, "pages/catalog/product.tmpl", map[string]any{
		"Name":        "Pen",
		"Description": "This is a pen, you can write with it!",
		"Stock":       7,
		"Price":       "£8.99",
	}))
	expected, err := testdata.TestFixtures.ReadFile("fixtures/pen.html")
	require.NoError(s.T(), err)
	s.Equal(string(expected), dst.String())
}

// This test recreates an issue I ran into where I could use a partial inside a page template but not directly.
func TestFullPathTemplateRendering(t *testing.T) {
	// Create an in-memory filesystem for testing
	fsys := fstest.MapFS{
		"partials/test-partial.tmpl": &fstest.MapFile{
			Data: []byte(`{{ define "partials/test-partial.tmpl" }}<div class="partial">Hello, {{ .Name }}!</div>{{ end }}`),
		},
		"pages/test-parent.tmpl": &fstest.MapFile{
			Data: []byte(`
<!DOCTYPE html>
<html>
<body>
    <div id="container">
        {{ template "partials/test-partial.tmpl" . }}
    </div>
</body>
</html>
`),
		},
	}

	registry, err := keystone.New(fsys)
	require.NoError(t, err)

	registry.Debug = true

	// Render parent template (initial page load scenario)
	var parentBuf bytes.Buffer
	err = registry.Render(&parentBuf, "pages/test-parent.tmpl", map[string]any{
		"Name": "World",
	})
	require.NoError(t, err, "Should render parent template successfully")
	require.Contains(t, parentBuf.String(), "Hello, World!", "Parent template should contain rendered partial")

	// Render just the partial (HTMX partial load scenario)
	var partialBuf bytes.Buffer
	err = registry.Render(&partialBuf, "partials/test-partial.tmpl", map[string]any{
		"Name": "HTMX",
	})
	require.NoError(t, err, "Should render partial template successfully")
	require.Contains(t, partialBuf.String(), "Hello, HTMX!", "Partial should render correctly")
}

func TestKeystoneSuite(t *testing.T) {
	suite.Run(t, new(KeystoneTestSuite))
}
