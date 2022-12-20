package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/mockery/v3/pkg/config"
)

func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()
	assert.Equal(t, "mockery", cmd.Name())
}

func TestConfigEnvFlags(t *testing.T) {
	expected := config.Config{
		Config:               "my_file.yaml",
		Name:                 "SomeInterface",
		Print:                true,
		Output:               "/some/dir",
		Outpkg:               "some/package",
		Packageprefix:        "prefix_",
		Dir:                  "dir/to/search",
		Recursive:            true,
		All:                  true,
		InPackage:            true,
		TestOnly:             true,
		Case:                 "underscore",
		Note:                 "// this is a test",
		Cpuprofile:           "test.pprof",
		Version:              true,
		KeepTree:             true,
		BuildTags:            "test mock",
		FileName:             "my-file.go",
		StructName:           "Interface1",
		LogLevel:             "warn",
		SrcPkg:               "some/other/package",
		DryRun:               true,
		DisableVersionString: true,
		BoilerplateFile:      "some/file",
		UnrollVariadic:       false,
		Exported:             true,
		WithExpecter:         true,
	}

	env(t, "CONFIG", expected.Config)
	env(t, "NAME", expected.Name)
	env(t, "PRINT", fmt.Sprint(expected.Print))
	env(t, "OUTPUT", expected.Output)
	env(t, "OUTPKG", expected.Outpkg)
	env(t, "PACKAGEPREFIX", expected.Packageprefix)
	env(t, "DIR", expected.Dir)
	env(t, "RECURSIVE", fmt.Sprint(expected.Recursive))
	env(t, "ALL", fmt.Sprint(expected.All))
	env(t, "INPACKAGE", fmt.Sprint(expected.InPackage))
	env(t, "TESTONLY", fmt.Sprint(expected.TestOnly))
	env(t, "CASE", expected.Case)
	env(t, "NOTE", expected.Note)
	env(t, "CPUPROFILE", expected.Cpuprofile)
	env(t, "VERSION", fmt.Sprint(expected.Version))
	env(t, "QUIET", fmt.Sprint(expected.Quiet))
	env(t, "KEEPTREE", fmt.Sprint(expected.KeepTree))
	env(t, "TAGS", expected.BuildTags)
	env(t, "FILENAME", expected.FileName)
	env(t, "STRUCTNAME", expected.StructName)
	env(t, "LOG_LEVEL", expected.LogLevel)
	env(t, "SRCPKG", expected.SrcPkg)
	env(t, "DRY_RUN", fmt.Sprint(expected.DryRun))
	env(t, "DISABLE_VERSION_STRING", fmt.Sprint(expected.DisableVersionString))
	env(t, "BOILERPLATE_FILE", expected.BoilerplateFile)
	env(t, "UNROLL_VARIADIC", fmt.Sprint(expected.UnrollVariadic))
	env(t, "EXPORTED", fmt.Sprint(expected.Exported))
	env(t, "WITH_EXPECTER", fmt.Sprint(expected.WithExpecter))

	initConfig()

	app, err := GetRootAppFromViper(viper.GetViper())
	require.NoError(t, err)

	assert.Equal(t, expected, app.Config)
}

func env(t *testing.T, key, value string) {
	key = "MOCKERY_" + key
	t.Cleanup(func() { os.Unsetenv(key) })
	os.Setenv(key, value)
}
