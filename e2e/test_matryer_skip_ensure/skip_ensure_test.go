package test_matryer_skip_ensure

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInterfaceSkipEnsure(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "mockery")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, "github.com/vektra/mockery/v3")
	build.Env = append(os.Environ(), "GOWORK=off")
	output, err := build.CombinedOutput()
	require.NoError(t, err, string(output))

	for _, tt := range []struct {
		name         string
		fileSkip     bool
		firstSkip    bool
		secondSkip   bool
		mixed        bool
		methodImport bool
	}{
		{name: "interface disables ensure", firstSkip: true},
		{name: "interface enables ensure", fileSkip: true},
		{name: "mixed shared output", fileSkip: true, mixed: true, secondSkip: true},
		{name: "method retains source import", firstSkip: true, methodImport: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.Mkdir(filepath.Join(dir, "api"), 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/skipensure\n\ngo 1.25.5\n"), 0o600))
			source := "package api\ntype First interface { Ping() string }; type Second interface { Pong() string }; type Value struct { N int }; type Typed interface { Get() Value }\n"
			require.NoError(t, os.WriteFile(filepath.Join(dir, "api", "api.go"), []byte(source), 0o600))
			first := "First"
			if tt.methodImport {
				first = "Typed"
			}
			interfaces := map[string]any{
				first: map[string]any{"config": map[string]any{"template-data": map[string]any{"skip-ensure": tt.firstSkip}}},
			}
			if tt.mixed {
				interfaces["Second"] = map[string]any{"config": map[string]any{"template-data": map[string]any{"skip-ensure": tt.secondSkip}}}
			}
			config := map[string]any{
				"template": "matryer", "all": false, "force-file-write": true,
				"dir": "./mocks", "filename": "mocks.go", "pkgname": "mocks", "structname": "{{.InterfaceName}}Mock",
				"template-data": map[string]any{"skip-ensure": tt.fileSkip},
				"packages":      map[string]any{"example.com/skipensure/api": map[string]any{"interfaces": interfaces}},
			}
			data, err := json.Marshal(config)
			require.NoError(t, err)
			configPath := filepath.Join(dir, "mockery.yml")
			require.NoError(t, os.WriteFile(configPath, data, 0o600))
			generate := exec.CommandContext(t.Context(), binary, "--config", configPath)
			generate.Dir = dir
			generate.Env = append(os.Environ(), "GOWORK=off")
			output, err := generate.CombinedOutput()
			require.NoError(t, err, string(output))

			consumer := "package mocks\nimport \"testing\"\n"
			if tt.methodImport {
				consumer += "import \"example.com/skipensure/api\"\nfunc TestGenerated(t *testing.T) { m := &TypedMock{GetFunc:func() api.Value {return api.Value{N:7}}}; if m.Get().N!=7 || len(m.GetCalls())!=1 {t.Fatal(\"mock behavior\")} }\n"
			} else {
				consumer += "func TestGenerated(t *testing.T) { m := &FirstMock{PingFunc:func()string{return \"ok\"}}; if m.Ping()!=\"ok\" || len(m.PingCalls())!=1 {t.Fatal(\"mock behavior\")}"
				if tt.mixed {
					consumer += "; second := &SecondMock{PongFunc:func()string{return \"pong\"}}; if second.Pong()!=\"pong\" || len(second.PongCalls())!=1 {t.Fatal(\"second mock behavior\")}"
				}
				consumer += "}\n"
			}
			require.NoError(t, os.WriteFile(filepath.Join(dir, "mocks", "consumer_test.go"), []byte(consumer), 0o600))
			test := exec.CommandContext(t.Context(), "go", "test", "-count=1", "./...")
			test.Dir = dir
			test.Env = append(os.Environ(), "GOWORK=off")
			output, err = test.CombinedOutput()
			require.NoError(t, err, string(output))
		})
	}
}
