package pkg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/chigopher/pathlib"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pkgMocks "github.com/vektra/mockery/v2/mocks/github.com/vektra/mockery/v2/pkg"
	"github.com/vektra/mockery/v2/pkg/config"
	"github.com/vektra/mockery/v2/pkg/logging"
)

func TestFilenameBare(t *testing.T) {
	out := FileOutputStreamProvider{InPackage: false, TestOnly: false}
	assert.Equal(t, "name.go", out.filename("name"))
}

func TestFilenameMockOnly(t *testing.T) {
	out := FileOutputStreamProvider{InPackage: true, TestOnly: false}
	assert.Equal(t, "mock_name.go", out.filename("name"))
}

func TestFilenameMockOnlyWithSuffix(t *testing.T) {
	out := FileOutputStreamProvider{InPackage: true, InPackageSuffix: true, TestOnly: false}
	assert.Equal(t, "name_mock.go", out.filename("name"))
}

func TestFilenameMockTest(t *testing.T) {
	out := FileOutputStreamProvider{InPackage: true, TestOnly: true}
	assert.Equal(t, "mock_name_test.go", out.filename("name"))
}

func TestFilenameMockTestWithSuffix(t *testing.T) {
	out := FileOutputStreamProvider{InPackage: true, InPackageSuffix: true, TestOnly: true}
	assert.Equal(t, "name_mock_test.go", out.filename("name"))
}

func TestFilenameKeepTreeInPackage(t *testing.T) {
	out := FileOutputStreamProvider{KeepTree: true, InPackage: true}
	assert.Equal(t, "name.go", out.filename("name"))
}

func TestFilenameTest(t *testing.T) {
	out := FileOutputStreamProvider{InPackage: false, TestOnly: true}
	assert.Equal(t, "name_test.go", out.filename("name"))
}

func TestFilenameOverride(t *testing.T) {
	out := FileOutputStreamProvider{InPackage: false, TestOnly: true, FileName: "override.go"}
	assert.Equal(t, "override.go", out.filename("anynamehere"))
}

func TestUnderscoreCaseName(t *testing.T) {
	assert.Equal(t, "notify_event", (&FileOutputStreamProvider{}).underscoreCaseName("NotifyEvent"))
	assert.Equal(t, "repository", (&FileOutputStreamProvider{}).underscoreCaseName("Repository"))
	assert.Equal(t, "http_server", (&FileOutputStreamProvider{}).underscoreCaseName("HTTPServer"))
	assert.Equal(t, "awesome_http_server", (&FileOutputStreamProvider{}).underscoreCaseName("AwesomeHTTPServer"))
	assert.Equal(t, "csv", (&FileOutputStreamProvider{}).underscoreCaseName("CSV"))
	assert.Equal(t, "position0_size", (&FileOutputStreamProvider{}).underscoreCaseName("Position0Size"))
}

func Test_parseConfigTemplates(t *testing.T) {
	mockPkg := func(t *testing.T) *pkgMocks.TypesPackage {
		m := pkgMocks.NewTypesPackage(t)
		m.EXPECT().Path().Return("github.com/user/project/package")
		m.EXPECT().Name().Return("packageName")
		return m
	}
	cwd, err := os.Getwd()
	require.NoError(t, err)

	type args struct {
		c     *config.Config
		iface *Interface
	}
	tests := []struct {
		name             string
		args             args
		disableWantCheck bool

		// pkg is used to generate a mock types.Package object.
		// It has to take in the testing.T object so we can
		// assert expectations.
		pkg     func(t *testing.T) *pkgMocks.TypesPackage
		want    *config.Config
		wantErr error
	}{
		{
			name: "standards",
			args: args{
				c: &config.Config{
					Dir:      "{{.InterfaceDir}}/{{.PackagePath}}",
					FileName: "{{.InterfaceName}}_{{.InterfaceNameCamel}}_{{.InterfaceNameSnake}}_{{.InterfaceNameLower}}_{{base .InterfaceFile}}",
					MockName: "{{.InterfaceNameLowerCamel}}",
					Outpkg:   "{{.PackageName}}",
				},

				iface: &Interface{
					Name:     "FooBar",
					FileName: "path/to/foobar.go",
				},
			},
			pkg: mockPkg,
			want: &config.Config{
				Dir:      "path/to/github.com/user/project/package",
				FileName: "FooBar_FooBar_foo_bar_foobar_foobar.go",
				MockName: "fooBar",
				Outpkg:   "packageName",
			},
		},
		{
			name: "template funcs cases",
			args: args{
				c: &config.Config{
					Dir:      "{{.InterfaceDir}}/{{.PackagePath}}",
					FileName: "{{.InterfaceName | kebabcase }}.go",
					MockName: "{{.InterfaceName | camelcase }}",
					Outpkg:   "{{.PackageName | snakecase }}",
				},

				iface: &Interface{
					Name:     "FooBar",
					FileName: "path/to/foobar.go",
				},
			},
			pkg: mockPkg,
			want: &config.Config{
				Dir:      "path/to/github.com/user/project/package",
				FileName: "foo-bar.go",
				MockName: "FooBar",
				Outpkg:   "package_name",
			},
		},
		{
			name: "InterfaceDirRelative in current working directory",
			args: args{
				c: &config.Config{
					Dir: "{{.InterfaceDirRelative}}",
				},

				iface: &Interface{
					Name:     "FooBar",
					FileName: cwd + "/path/to/foobar.go",
				},
			},
			pkg: mockPkg,
			want: &config.Config{
				Dir: "path/to",
			},
		},
		{
			name: "InterfaceDirRelative not in current working directory",
			args: args{
				c: &config.Config{
					Dir: "mocks/{{.InterfaceDirRelative}}",
				},

				iface: &Interface{
					Name:     "FooBar",
					FileName: "/path/to/foobar.go",
				},
			},
			pkg: mockPkg,
			want: &config.Config{
				Dir: "mocks/github.com/user/project/package",
			},
		},
		{
			name: "infinite loop in template variables",
			args: args{
				c: &config.Config{
					Dir:      "{{.InterfaceDir}}/{{.PackagePath}}",
					FileName: "{{.MockName}}.go",
					MockName: "Mock{{.MockName}}",
					Outpkg:   "{{.PackageName}}",
				},

				iface: &Interface{
					Name:     "FooBar",
					FileName: "path/to/foobar.go",
				},
			},
			pkg:              mockPkg,
			disableWantCheck: true,
			wantErr:          ErrInfiniteLoop,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.iface.Pkg = tt.pkg(t)

			err := parseConfigTemplates(context.Background(), tt.args.c, tt.args.iface)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("parseConfigTemplates() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.disableWantCheck && !reflect.DeepEqual(tt.args.c, tt.want) {
				t.Errorf("*config.Config = %s\n, want %+v", spew.Sdump(tt.args.c), spew.Sdump(tt.want))
			}
		})
	}
}

func TestOutputter_Generate(t *testing.T) {
	type fields struct {
		boilerplate string
		config      *config.Config
	}

	tests := []struct {
		name        string
		packagePath string
		fields      fields
		dryRun      bool
	}{
		{
			name:        "generate normal",
			packagePath: "github.com/vektra/mockery/v2/pkg/fixtures/example_project",
			dryRun:      false,
		},
		{
			name:        "generate normal",
			packagePath: "github.com/vektra/mockery/v2/pkg/fixtures/example_project",
			dryRun:      true,
		},
	}
	for _, tt := range tests {
		if tt.fields.config == nil {
			tt.fields.config = &config.Config{}
		}
		tt.fields.config.Dir = t.TempDir()
		tt.fields.config.MockName = "Mock{{.InterfaceName}}"
		tt.fields.config.FileName = "mock_{{.InterfaceName}}.go"
		tt.fields.config.Outpkg = "{{.PackageName}}"

		t.Run(tt.name, func(t *testing.T) {
			m := &Outputter{
				boilerplate: tt.fields.boilerplate,
				config:      tt.fields.config,
				dryRun:      tt.dryRun,
			}
			parser := NewParser([]string{})

			log, err := logging.GetLogger("INFO")
			require.NoError(t, err)
			ctx := log.WithContext(context.Background())

			confPath := pathlib.NewPath(t.TempDir()).Join("config.yaml")
			ymlContents := fmt.Sprintf(`
packages:
  %s:
    config:
      all: True
`, tt.packagePath)
			require.NoError(t, confPath.WriteFile([]byte(ymlContents)))
			m.config.Config = confPath.String()

			require.NoError(t, parser.ParsePackages(ctx, []string{tt.packagePath}))
			require.NoError(t, parser.Load())
			for _, intf := range parser.Interfaces() {
				t.Logf("generating interface: %s %s", intf.QualifiedName, intf.Name)
				require.NoError(t, m.Generate(ctx, intf))
				mockPath := pathlib.NewPath(tt.fields.config.Dir).Join("mock_" + intf.Name + ".go")

				t.Logf("checking if path exists: %v", mockPath)
				exists, err := mockPath.Exists()
				require.NoError(t, err)
				assert.Equal(t, !tt.dryRun, exists)
			}
		})
	}
}
