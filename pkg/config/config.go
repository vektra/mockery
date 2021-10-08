package config

import "runtime/debug"

const (
	_defaultSemVer = "v0.0.0-dev"
)

// SemVer is the version of mockery at build time.
var SemVer = ""

func GetSemverInfo() string {
	if SemVer != "" {
		return SemVer
	}
	version, ok := debug.ReadBuildInfo()
	if ok && version.Main.Version != "(devel)" {
		return version.Main.Version
	}
	return _defaultSemVer
}

// Config contains the app configuration.
type Config struct {
	All           bool   //KILLME
	Dir           string //KILLME
	FileName      string //KILLME
	InPackage     bool   //KILLME
	Name          string //KILLME
	Outpkg        string //KILLME
	Packageprefix string //KILLME
	Output        string //KILLME
	Recursive     bool   //KILLME
	SrcPkg        string //KILLME
	TestOnly      bool   //KILLME
	KeepTree      bool

	Mocks []MockDef

	BuildTags            string `mapstructure:"tags"`
	Case                 string
	Config               string
	Cpuprofile           string
	DisableVersionString bool   `mapstructure:"disable-version-string"`
	DryRun               bool   `mapstructure:"dry-run"`
	Exported             bool   `mapstructure:"exported"`
	LogLevel             string `mapstructure:"log-level"`
	Note                 string
	Print                bool
	Profile              string
	Quiet                bool
	BoilerplateFile      string `mapstructure:"boilerplate-file"`
	Tags                 string
	UnrollVariadic       bool `mapstructure:"unroll-variadic"`
	Version              bool
}

// MockDef contains a single mock definition.
type MockDef struct {
	// Package is the relative path of the package to generate mocks for.
	Package string
	// Interface is the interface name.
	Interface string
	// FileName is the name of generated file.
	FileName string
	// InPackage is true when the mock is to go goes inside the original package.
	InPackage bool
	// Outpkg contains the name of the generated package.
	Outpkg string
	// Output contains the directory to write mocks to
	Output string
	// StructName overrides the name given to the mock struct.
	StructName string
	// TestOnly, if true, makes mockery generate a mock in a _test.go file.
	TestOnly bool
}

func (m *MockDef) GeneratedPackageName() string {
	if m.Outpkg == "" {
		return "mocks"
	}
	return m.Outpkg
}
