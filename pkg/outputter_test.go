package pkg

import (
	"context"
	"reflect"
	"testing"

	"github.com/chigopher/pathlib"
	"github.com/stretchr/testify/assert"
	pkgMocks "github.com/vektra/mockery/v2/mocks/github.com/vektra/mockery/v2/pkg"
	"github.com/vektra/mockery/v2/pkg/config"
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

func configPath(t *testing.T) *pathlib.Path {
	return pathlib.NewPath(t.TempDir()).Join("config.yaml")
}

func configString() string {
	return `
packages:
	`
}
func newConfig(t *testing.T) *config.Config {
	return &config.Config{}
}

func Test_parseConfigTemplates(t *testing.T) {
	type args struct {
		c     *config.Config
		iface *Interface
	}
	tests := []struct {
		name string
		args args

		// pkg is used to generate a mock types.Package object.
		// It has to take in the testing.T object so we can
		// assert expectations.
		pkg     func(t *testing.T) *pkgMocks.TypesPackage
		want    *config.Config
		wantErr bool
	}{
		{
			name: "standards",
			args: args{
				c: &config.Config{
					Dir:      "{{.InterfaceDir}}/{{.PackagePath}}",
					FileName: "{{.InterfaceName}}_{{.InterfaceNameCamel}}_{{.InterfaceNameSnake}}.go",
					MockName: "{{.InterfaceNameLowerCamel}}",
					Outpkg:   "{{.PackageName}}",
				},

				iface: &Interface{
					Name:     "FooBar",
					FileName: "path/to/foobar.go",
				},
			},
			pkg: func(t *testing.T) *pkgMocks.TypesPackage {
				m := pkgMocks.NewTypesPackage(t)
				m.EXPECT().Path().Return("github.com/user/project/package")
				m.EXPECT().Name().Return("packageName")
				return m
			},
			want: &config.Config{
				Dir:      "path/to/github.com/user/project/package",
				FileName: "FooBar_FooBar_foo_bar.go",
				MockName: "fooBar",
				Outpkg:   "packageName",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.iface.Pkg = tt.pkg(t)

			if err := parseConfigTemplates(context.Background(), tt.args.c, tt.args.iface); (err != nil) != tt.wantErr {
				t.Errorf("parseConfigTemplates() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.args.c, tt.want) {
				t.Errorf("*config.Config = %v, want %v", tt.args.c, tt.want)
			}
		})
	}
}
