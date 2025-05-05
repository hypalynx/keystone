package keystone_test

import (
	"bytes"
	"path/filepath"
	"testing"

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

// Test Render on disk change.. good luck writing that test!

func TestKeystoneSuite(t *testing.T) {
	suite.Run(t, new(KeystoneTestSuite))
}
