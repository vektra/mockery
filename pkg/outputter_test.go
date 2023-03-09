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

func Test_foobar(t *testing.T) {
	mockPackage := pkgMocks.NewTypesPackage(t)
	mockPackage.EXPECT().Name().Return("pkg")
	mockPackage.EXPECT().Path().Return("github.com/vektra/mockery")

	iface := &Interface{
		Name: "interfaceName",
		Pkg:  mockPackage,
	}
	path, err := outputFilePath(context.Background(), iface, &config.Config{
		FileName: "filename.go",
		Dir:      "dirname",
	}, "")
	assert.NoError(t, err)
	assert.Equal(t, pathlib.NewPath("dirname/filename.go"), path)
}

func Test_outputFilePath(t *testing.T) {
	type parameters struct {
		packageName      string
		packagePath      string
		interfaceName    string
		fileName         string
		fileNameTemplate string
		dirTemplate      string
		mockName         string
	}
	tests := []struct {
		name    string
		params  parameters
		want    *pathlib.Path
		wantErr bool
	}{
		{
			name: "defaults",
			params: parameters{
				packageName:      "pkg",
				packagePath:      "github.com/vektra/mockery",
				interfaceName:    "Foo",
				fileNameTemplate: "mock_{{.InterfaceName}}.go",
				dirTemplate:      "mocks/{{.PackagePath}}",
			},
			want: pathlib.NewPath("mocks/github.com/vektra/mockery/mock_Foo.go"),
		},
		{
			name: "dir and filename templates",
			params: parameters{
				packageName:      "pkg",
				packagePath:      "github.com/vektra/mockery",
				interfaceName:    "Foo",
				fileNameTemplate: "{{.MockName}}_{{.PackageName}}_{{.InterfaceName}}.go",
				dirTemplate:      "{{.PackagePath}}",
				mockName:         "MockFoo",
			},
			want: pathlib.NewPath("github.com/vektra/mockery/MockFoo_pkg_Foo.go"),
		},
		{
			name: "mock next to original interface",
			params: parameters{
				packageName:      "pkg",
				packagePath:      "github.com/vektra/mockery/pkg/internal",
				interfaceName:    "Foo",
				fileName:         "pkg/internal/foo.go",
				dirTemplate:      "{{.InterfaceDir}}",
				fileNameTemplate: "mock_{{.InterfaceName}}.go",
				mockName:         "MockFoo",
			},
			want: pathlib.NewPath("pkg/internal/mock_Foo.go"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPackage := pkgMocks.NewTypesPackage(t)
			mockPackage.EXPECT().Name().Return(tt.params.packageName)
			mockPackage.EXPECT().Path().Return(tt.params.packagePath)

			iface := &Interface{
				Name:     tt.params.interfaceName,
				Pkg:      mockPackage,
				FileName: tt.params.fileName,
			}

			got, err := outputFilePath(
				context.Background(),
				iface,
				&config.Config{
					FileName: tt.params.fileNameTemplate,
					Dir:      tt.params.dirTemplate,
				},
				tt.params.mockName,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("outputFilePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("outputFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
