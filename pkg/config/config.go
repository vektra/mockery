package config

import "runtime/debug"

// SemVer is the version of mockery at build time.
var SemVer = "0.0.0-dev"

// GetSemverInfo attempts to read and return the go build info of Mockery
// If a version is not found, the fallback versioning is returned
func GetSemverInfo() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return SemVer
	}
	return info.Main.Version
}

type Config struct {
	All                  bool
	BuildTags            string `mapstructure:"tags"`
	Case                 string
	Config               string
	Cpuprofile           string
	Dir                  string
	DisableVersionString bool `mapstructure:"disable-version-string"`
	DryRun               bool `mapstructure:"dry-run"`
	Exported             bool `mapstructure:"exported"`
	FileName             string
	InPackage            bool
	KeepTree             bool
	LogLevel             string `mapstructure:"log-level"`
	Name                 string
	Note                 string
	Outpkg               string
	Packageprefix        string
	Output               string
	Print                bool
	Profile              string
	Quiet                bool
	Recursive            bool
	SrcPkg               string
	BoilerplateFile      string `mapstructure:"boilerplate-file"`
	// StructName overrides the name given to the mock struct and should only be nonempty
	// when generating for an exact match (non regex expression in -name).
	StructName     string
	Tags           string
	TestOnly       bool
	UnrollVariadic bool `mapstructure:"unroll-variadic"`
	Version        bool
}
