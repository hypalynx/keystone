package keystone_test

import (
	"html/template"
	"os"
	"strings"
	"testing"

	"github.com/hypalynx/keystone"
	"github.com/hypalynx/keystone/testdata"
	"github.com/stretchr/testify/require"
)

func TestKeystoneAsPackage(t *testing.T) {
	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
	}

	diskFS := os.DirFS("./testdata/")
	ksFromDisk := &keystone.Registry{
		Source:  diskFS,
		Reload:  true,
		FuncMap: funcMap,
	}
	require.NoError(t, ksFromDisk.Load())

	ksFromEmbed := &keystone.Registry{
		Source:  testdata.TestTemplatesFS,
		Reload:  false,
		FuncMap: funcMap,
	}
	require.NoError(t, ksFromEmbed.Load())
}

func TestDefaultValues(t *testing.T) {
	ks := &keystone.Registry{
		Source: testdata.TestTemplatesFS,
	}

	require.Equal(t, []string(nil), ks.Extensions)
	require.NoError(t, ks.Load())
	require.Equal(t, []string{"tmpl", "html", "gohtml", "gotmpl", "tpl"}, ks.Extensions)
}
