package keystone_test

import (
	"testing"
	"text/template"

	"github.com/hypalynx/keystone/pkg/keystone"
	"github.com/hypalynx/keystone/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateIngestion(t *testing.T) {
	ks, err := keystone.New(testdata.TestBaseTemplateFS, testdata.TestPagesTemplateFS, template.FuncMap{})

	require.NoError(t, err)
	assert.Equal(
		t,
		[]string{
			"components/card.tmpl",
			"components/catalog/description.tmpl",
			"components/catalog/non-tmpl-ext-template.html",
			"layouts/default.tmpl",
			"pages/catalog/product.tmpl",
			"pages/test.tmpl",
			"partials/results.tmpl",
		},
		ks.ListAll(),
	)
}

func TestEmbeddedPartial(t *testing.T) {
	assert.NoError(t, nil)
}
