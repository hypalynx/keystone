package keystone_test

import (
	"testing"

	"github.com/hypalynx/keystone/pkg/keystone"
	"github.com/hypalynx/keystone/testdata"
	"github.com/stretchr/testify/assert"
)

func TestIngestingTemplates(t *testing.T) {
	ks := keystone.New(testdata.TestBaseTemplateFS, testdata.TestPagesTemplateFS)

	assert.True(t, ks.Exists("test.tmpl"))
}

func TestEmbeddedPartial(t *testing.T) {
	assert.NoError(t, nil)
}
