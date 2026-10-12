package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShouldGenerateInterface(t *testing.T) {
	pkg := func(all bool, interfaces ...string) PackageConfig {
		pc := PackageConfig{
			Config: &Config{
				All:                   addr(all),
				IncludeInterfaceRegex: addr(""),
				ExcludeInterfaceRegex: addr(""),
			},
			Interfaces: map[string]*InterfaceConfig{},
		}
		for _, name := range interfaces {
			pc.Interfaces[name] = &InterfaceConfig{}
		}
		return pc
	}

	for name, tc := range map[string]struct {
		pkg       PackageConfig
		generate  *bool
		directive bool
		want      bool
	}{
		"all, generate unset":                  {pkg(true), nil, false, true},
		"all, generate true":                   {pkg(true), addr(true), false, true},
		"all, generate false wins":             {pkg(true), addr(false), false, false},
		"listed interface, generate false":     {pkg(false, "Foo"), addr(false), false, false},
		"listed interface":                     {pkg(false, "Foo"), nil, false, true},
		"not listed, not all":                  {pkg(false), nil, false, false},
		"directive comment":                    {pkg(false), nil, true, true},
		"directive comment, generate false":    {pkg(false), addr(false), true, false},
		"directive comment, all, generate off": {pkg(true), addr(false), true, false},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := tc.pkg.ShouldGenerateInterface(context.Background(), "Foo", Config{Generate: tc.generate}, tc.directive)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
