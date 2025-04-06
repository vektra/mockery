package testremotetemplates

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"

	"github.com/chigopher/pathlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var configTemplate = `
dir: %s
filename: %s
template: %s
formatter: noop
force-file-write: true
pkgname: test_pkgname
template-data:
  foo: foo
  bar: bar
packages:
  github.com/vektra/mockery/v3/internal/fixtures/template_exercise:
    interfaces:
      Exercise:
`

func TestRemoteTemplates(t *testing.T) {
	var err error

	// the temp dir needs to reside within the mockery project because mockery
	// requires a go.mod file to function correctly. Using t.TempDir() won't work
	// because of this.
	tmpDirBase := pathlib.NewPath("./test")
	require.NoError(t, tmpDirBase.Mkdir())
	tmpDirBase, err = tmpDirBase.ResolveAll()
	require.NoError(t, err)
	defer assert.NoError(t, tmpDirBase.RemoveAll())

	type test struct {
		name             string
		schema           string
		expectMockeryErr bool
	}
	for _, tt := range []test{
		{
			name: "schema validation OK",
			schema: `{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"title": "vektra/mockery matryer mock",
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"foo": {
		"type": "string"
		},
		"bar": {
		"type": "string"
		}
	},
	"required": ["foo", "bar"]
}`,
			expectMockeryErr: false,
		},
		{
			name: "Required parameter doesn't exist",
			schema: `{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"title": "vektra/mockery matryer mock",
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"foo": {
			"type": "string"
		},
		"bar": {
			"type": "string"
		},
		"baz": {
			"type": "string"
		}
	},
	"required": ["foo", "bar", "baz"]
}`,
			expectMockeryErr: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tmpdir := tmpDirBase.Join(t.Name())
			require.NoError(t, tmpdir.MkdirAll())

			configFile := tmpdir.Join(".mockery.yml")
			outFile := tmpdir.Join("out.txt")

			templateName := "template.templ"
			mux := http.NewServeMux()
			mux.HandleFunc(fmt.Sprintf("/%s", templateName), func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "Hello, world!")
			})
			mux.HandleFunc(fmt.Sprintf("/%s.schema.json", templateName), func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, tt.schema)
			})

			ts := httptest.NewServer(mux)
			defer ts.Close()

			fullPath := fmt.Sprintf("%s/%s", ts.URL, templateName)

			configFileContents := fmt.Sprintf(
				configTemplate,
				outFile.Parent().String(),
				outFile.Name(),
				fullPath,
			)
			require.NoError(t, configFile.WriteFile([]byte(configFileContents)))

			//nolint: gosec
			out, err := exec.Command(
				"go", "run", "github.com/vektra/mockery/v3",
				"--config", configFile.String()).CombinedOutput()
			if tt.expectMockeryErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err, string(out))
				outFileBytes, err := outFile.ReadFile()
				require.NoError(t, err)
				assert.Equal(t, "Hello, world!", string(outFileBytes))
			}
		})
	}
}
