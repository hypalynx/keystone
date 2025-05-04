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
	assert.True(t, ks.Exists("components/card.tmpl"))
	assert.True(t, ks.Exists("pages/test.tmpl"))
}

func TestEmbeddedPartial(t *testing.T) {
	assert.NoError(t, nil)
}
