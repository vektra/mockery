package template

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/brunoga/deep"
	"github.com/chigopher/pathlib"
	koanfYAML "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"github.com/vektra/mockery/v3/internal/logging"
	"github.com/vektra/mockery/v3/internal/stackerr"
	"golang.org/x/tools/go/packages"
)

type Interface struct {
	Name     string // Name of the type to be mocked.
	FileName string
	File     *ast.File
	Pkg      *packages.Package
	Config   *Config
}

func NewInterface(name string, filename string, file *ast.File, pkg *packages.Package, config *Config) *Interface {
	return &Interface{
		Name:     name,
		FileName: filename,
		File:     file,
		Pkg:      pkg,
		Config:   config,
	}
}

type RootConfig struct {
	*Config    `koanf:",squash"`
	Packages   map[string]*PackageConfig `koanf:"packages"`
	koanf      *koanf.Koanf
	configFile *pathlib.Path
}

func NewRootConfig(
	ctx context.Context,
	flags *pflag.FlagSet,
) (*RootConfig, *koanf.Koanf, error) {
	var configFile *pathlib.Path

	log := zerolog.Ctx(ctx)
	var err error

	conf := &Config{}
	// Set all parameters to their respective zero-values. Need to use
	// reflection for this sadly.
	v := reflect.ValueOf(conf).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() != reflect.Pointer {
			continue
		}
		if !field.IsNil() {
			continue
		}
		field.Set(reflect.New(field.Type().Elem()))
	}

	var rootConfig RootConfig = RootConfig{
		Config: conf,
	}
	k := koanf.New("|")
	rootConfig.koanf = k
	k.Set("dir", "{{.InterfaceDir}}")
	k.Set("filename", "mocks_test.go")
	k.Set("formatter", "goimports")
	k.Set("mockname", "Mock{{.InterfaceName}}")
	k.Set("pkgname", "{{.SrcPackageName}}")
	k.Set("log-level", "info")

	configFileFromEnv := os.Getenv("MOCKERY_CONFIG")
	if configFileFromEnv != "" {
		configFile = pathlib.NewPath(configFileFromEnv)
	}
	if configFile == nil {
		configFileFromFlags, err := flags.GetString("config")
		if err != nil {
			return nil, nil, fmt.Errorf("getting --config from flags: %w", err)
		}
		if configFileFromFlags != "" {
			configFile = pathlib.NewPath(configFileFromFlags)
		}
	}
	if configFile == nil {
		log.Debug().Msg("config file not specified, searching")
		configFile, err = findConfig()
		if err != nil {
			return nil, k, fmt.Errorf("discovering mockery config: %w", err)
		}
		log.Debug().Str("config-file", configFile.String()).Msg("config file found")
	}
	rootConfig.configFile = configFile
	k.Load(
		env.Provider(
			"MOCKERY_",
			".",
			func(s string) string {
				return strings.Replace(strings.ToLower(strings.TrimPrefix(s, "MOCKERY_")), "_", "-", -1)
			}),
		nil,
	)

	if err := k.Load(file.Provider(configFile.String()), koanfYAML.Parser()); err != nil {
		return nil, k, fmt.Errorf("loading config file: %w", err)
	}

	if flags != nil {
		if err := k.Load(posflag.Provider(flags, ".", k), nil); err != nil {
			return nil, k, fmt.Errorf("loading flags: %w", err)
		}
	}

	if err := k.Unmarshal("", &rootConfig); err != nil {
		return nil, k, fmt.Errorf("unmarshalling config: %w", err)
	}
	if err := rootConfig.Initialize(ctx); err != nil {
		return nil, k, fmt.Errorf("initializing root config: %w", err)
	}
	return &rootConfig, k, nil
}

func (c *RootConfig) ConfigFileUsed() *pathlib.Path {
	return c.configFile
}

// mergreStringMaps merges two (possibly nested) maps.
func mergeStringMaps(src, dest map[string]any) {
	for srcKey, srcValue := range src {
		if destValue, exists := dest[srcKey]; exists {
			// If the source value is a map, merge recursively
			if destMap, ok := destValue.(map[string]any); ok {
				if srcMap, ok := srcValue.(map[string]any); ok {
					mergeStringMaps(srcMap, destMap)
					continue
				}
			}
			continue
		}
		// Otherwise, set the value directly
		dest[srcKey] = srcValue
	}
}

// mergeConfigs merges the values from c1 into c2.
func mergeConfigs(ctx context.Context, src Config, dest *Config) {
	log := zerolog.Ctx(ctx)
	// Merge root config with package config
	srcValue := reflect.ValueOf(src)
	destValue := reflect.ValueOf(dest)

	for i := 0; i < srcValue.NumField(); i++ {
		fieldLog := log.With().
			Int("index", i).
			Str("name", srcValue.Type().Field(i).Name).
			Logger()
		fieldLog.Debug().Msg("Iterating over field for merging")
		srcFieldValue := srcValue.Field(i)
		destFieldValue := destValue.Elem().Field(i)

		if srcFieldValue.Kind() == reflect.Map {
			srcMap, ok := srcFieldValue.Interface().(map[string]any)
			if !ok {
				log.Debug().Msg("field value is not `any`, skipping merge")
				continue
			}
			destMap, ok := destFieldValue.Interface().(map[string]any)
			if !ok {
				log.Debug().Msg("dest map value is not `any`, skipping")
				continue
			}
			if destMap == nil {
				destFieldValue.Set(reflect.ValueOf(make(map[string]any)))
			}
			destMap = destFieldValue.Interface().(map[string]any)
			mergeStringMaps(srcMap, destMap)
		} else if srcFieldValue.Kind() == reflect.Pointer && destFieldValue.IsNil() {
			// Attribute is a pointer. We need to allocate a new value of the
			// same type as the type being pointed to.
			newValue := reflect.New(srcFieldValue.Elem().Type())
			// Then, set this new value to the same value as the src.
			newValue.Elem().Set(srcFieldValue.Elem())
			// newValue is already an address, so we can set destFieldValue
			// to it as-is.
			destFieldValue.Set(newValue)
		} else if destFieldValue.CanSet() && destFieldValue.IsZero() {
			destFieldValue.Set(srcFieldValue)
		} else {
			fieldLog.Debug().
				Bool("can-set", destFieldValue.CanSet()).
				Bool("is-zero", destFieldValue.IsZero()).
				Msg("field not addressable, not merging.")
		}
	}
}

func (c *RootConfig) Initialize(ctx context.Context) error {
	log := zerolog.Ctx(ctx)
	recursivePackages := []string{}
	for pkgName, pkgConfig := range c.Packages {
		if pkgConfig == nil {
			pkgConfig = NewPackageConfig()
			c.Packages[pkgName] = pkgConfig
		}
		if pkgConfig.Config == nil {
			pkgConfig.Config = &Config{}
		}
		if pkgConfig.Interfaces == nil {
			pkgConfig.Interfaces = map[string]*InterfaceConfig{}
		}
		pkgLog := log.With().Str("package-path", pkgName).Logger()
		pkgCtx := pkgLog.WithContext(ctx)

		mergeConfigs(pkgCtx, *c.Config, pkgConfig.Config)
		if err := pkgConfig.Initialize(pkgCtx); err != nil {
			return fmt.Errorf("initializing root config: %w", err)
		}
		if *pkgConfig.Config.Recursive {
			recursivePackages = append(recursivePackages, pkgName)
		}
	}

	for _, recursivePackageName := range recursivePackages {
		pkgLog := log.With().Str(logging.LogKeyPackagePath, recursivePackageName).Logger()
		pkgCtx := pkgLog.WithContext(ctx)
		pkgLog.Debug().Msg("package marked as recursive")

		subpkgs, err := c.subPackages(recursivePackageName)
		if err != nil {
			return fmt.Errorf("discovering sub packages of %s: %w", recursivePackageName, err)
		}
		parentPkgConfig := c.Packages[recursivePackageName]
		for _, subpkg := range subpkgs {
			if c.ShouldExcludeSubpkg(subpkg) {
				pkgLog.Debug().Msg("package was marked for exclusion")
				continue
			}
			var subPkgConfig *PackageConfig
			if existingSubPkg, exists := c.Packages[subpkg]; exists {
				subPkgConfig = existingSubPkg
			} else {
				subPkgConfig = NewPackageConfig()
			}
			mergeConfigs(pkgCtx, *parentPkgConfig.Config, subPkgConfig.Config)
			c.Packages[subpkg] = subPkgConfig
		}
	}
	return nil
}

func (c *RootConfig) subPackages(pkgPath string) ([]string, error) {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedFiles,
	}, pkgPath+"/...")
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	convertPkgPath := func(pkgs []*packages.Package) []string {
		paths := make([]string, 0, len(pkgs))
		for _, pkg := range pkgs {
			if len(pkg.GoFiles) == 0 {
				continue
			}
			paths = append(paths, pkg.PkgPath)
		}
		return paths
	}

	return convertPkgPath(pkgs), nil
}

func (c *RootConfig) GetPackageConfig(ctx context.Context, pkgPath string) (*PackageConfig, error) {
	pkgConfig, ok := c.Packages[pkgPath]
	if !ok {
		return nil, stackerr.NewStackErr(fmt.Errorf("package %s does not exist in the config", pkgPath))
	}
	return pkgConfig, nil
}

// GetPackages returns a list of the packages that are defined in
// the `packages` config section.
func (c *RootConfig) GetPackages(ctx context.Context) ([]string, error) {
	packages := []string{}
	for key := range c.Packages {
		packages = append(packages, key)
	}
	return packages, nil
}

type PackageConfig struct {
	Config     *Config                     `koanf:"config"`
	Interfaces map[string]*InterfaceConfig `koanf:"interfaces"`
}

func NewPackageConfig() *PackageConfig {
	return &PackageConfig{
		Config:     &Config{},
		Interfaces: map[string]*InterfaceConfig{},
	}
}

func (c *PackageConfig) Initialize(ctx context.Context) error {
	for idx, ifaceConfig := range c.Interfaces {
		if ifaceConfig == nil {
			ifaceConfig = NewInterfaceConfig()
			c.Interfaces[idx] = ifaceConfig
		}
		if ifaceConfig.Config == nil {
			ifaceConfig.Config = &Config{}
		}
		mergeConfigs(ctx, *c.Config, ifaceConfig.Config)
		if err := ifaceConfig.Initialize(ctx); err != nil {
			return fmt.Errorf("initializing package config: %w", err)
		}
	}
	return nil
}

func (c PackageConfig) GetInterfaceConfig(ctx context.Context, interfaceName string) *InterfaceConfig {
	log := zerolog.Ctx(ctx)
	if ifaceConfig, ok := c.Interfaces[interfaceName]; ok {
		return ifaceConfig
	}
	ifaceConfig := NewInterfaceConfig()

	newConfig, err := deep.Copy(c.Config)
	if err != nil {
		log.Err(err).Msg("issue when deep-copying package config to interface config")
		panic(err)
	}

	ifaceConfig.Config = newConfig
	ifaceConfig.Configs = []*Config{newConfig}
	return ifaceConfig
}

func (c PackageConfig) ShouldGenerateInterface(ctx context.Context, interfaceName string) (bool, error) {
	log := zerolog.Ctx(ctx)
	if *c.Config.All {
		if *c.Config.IncludeRegex != "" {
			log.Warn().Msg("interface config has both `all` and `include-regex` set: `include-regex` will be ignored")
		}
		if *c.Config.ExcludeRegex != "" {
			log.Warn().Msg("interface config has both `all` and `exclude-regex` set: `exclude-regex` will be ignored")
		}
		log.Debug().Msg("`all: true` is set, interface should be generated")
		return true, nil
	}

	if _, exists := c.Interfaces[interfaceName]; exists {
		return true, nil
	}

	includeRegex := *c.Config.IncludeRegex
	excludeRegex := *c.Config.ExcludeRegex
	if includeRegex == "" {
		if excludeRegex != "" {
			log.Warn().Msg("interface config has `exclude-regex` set but not `include-regex`: `exclude-regex` will be ignored")
		}
		return false, nil
	}
	includedByRegex, err := regexp.MatchString(includeRegex, interfaceName)
	if err != nil {
		return false, fmt.Errorf("evaluating `include-regex`: %w", err)
	}
	if !includedByRegex {
		log.Debug().Msg("interface does not match include-regex")
		return false, nil
	}
	log.Debug().Msg("interface matches include-regex")
	if excludeRegex == "" {
		return true, nil
	}
	excludedByRegex, err := regexp.MatchString(excludeRegex, interfaceName)
	if err != nil {
		return false, fmt.Errorf("evaluating `exclude-regex`: %w", err)
	}
	if excludedByRegex {
		log.Debug().Msg("interface matches exclude-regex")
		return false, nil
	}
	log.Debug().Msg("interface does not match exclude-regex")
	return true, nil
}

type InterfaceConfig struct {
	Config  *Config   `koanf:"config"`
	Configs []*Config `koanf:"configs"`
}

func NewInterfaceConfig() *InterfaceConfig {
	return &InterfaceConfig{
		Config:  &Config{},
		Configs: []*Config{},
	}
}

func (c *InterfaceConfig) Initialize(ctx context.Context) error {
	if len(c.Configs) == 0 {
		c.Configs = []*Config{c.Config}
	} else {
		for _, subCfg := range c.Configs {
			mergeConfigs(ctx, *c.Config, subCfg)
		}
	}

	return nil
}

type ReplaceType struct {
	PkgPath  string `koanf:"pkg-path"`
	TypeName string `koanf:"type-name"`
}

type Config struct {
	All                *bool          `koanf:"all"`
	Anchors            map[string]any `koanf:"_anchors"`
	BuildTags          *string        `koanf:"tags"`
	ConfigFile         *string        `koanf:"config"`
	Dir                *string        `koanf:"dir"`
	ExcludeSubpkgRegex []string       `koanf:"exclude-subpkg-regex"`
	ExcludeRegex       *string        `koanf:"exclude-regex"`
	FileName           *string        `koanf:"filename"`
	Formatter          *string        `koanf:"formatter"`
	IncludeRegex       *string        `koanf:"include-regex"`
	LogLevel           *string        `koanf:"log-level"`
	MockName           *string        `koanf:"mockname"`
	PkgName            *string        `koanf:"pkgname"`
	Recursive          *bool          `koanf:"recursive"`
	// ReplaceType is a nested map of format map["package path"]["type name"]*ReplaceType
	ReplaceType    map[string]map[string]*ReplaceType `koanf:"replace-type"`
	Template       *string                            `koanf:"template"`
	TemplateData   map[string]any                     `koanf:"template-data"`
	UnrollVariadic *bool                              `koanf:"unroll-variadic"`
	Version        *bool                              `koanf:"version"`
}

func findConfig() (*pathlib.Path, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting current working directory: %w", err)
	}
	currentPath := pathlib.NewPath(cwd)
	for len(currentPath.Parts()) != 1 {
		for _, confName := range []string{".mockery.yaml", ".mockery.yml"} {
			configPath := currentPath.Join(confName)
			isFile, err := configPath.Exists()
			if err != nil {
				return nil, fmt.Errorf("checking if %s is file: %w", configPath.String(), err)
			}
			if isFile {
				return configPath, nil
			}
		}
		currentPath = currentPath.Parent()
	}
	return nil, errors.New("mockery config file not found")
}

func (c *Config) FilePath() *pathlib.Path {
	return pathlib.NewPath(*c.Dir).Join(*c.FileName).Clean()
}

func (c *Config) ShouldExcludeSubpkg(pkgPath string) bool {
	for _, regex := range c.ExcludeSubpkgRegex {
		matched, err := regexp.MatchString(regex, pkgPath)
		if err != nil {
			panic(err)
		}
		if matched {
			return true
		}
	}
	return false
}

func IsAutoGenerated(path *pathlib.Path) (bool, error) {
	file, err := path.OpenFile(os.O_RDONLY)
	if err != nil {
		return false, stackerr.NewStackErr(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, "DO NOT EDIT") {
			return true, nil
		} else if strings.HasPrefix(text, "package ") {
			break
		}
	}
	return false, nil
}

var ErrInfiniteLoop = fmt.Errorf("infinite loop in template variables detected")

// ParseTemplates parses various templated strings
// in the config struct into their fully defined values. This mutates
// the config object passed. An *Interface object can be supplied to satisfy
// template variables that need information about the original
// interface being mocked. If this argument is nil, interface-specific template
// variables will be set to the empty string. The srcPkg is also needed to
// satisfy template variables regarding the source package.
func (c *Config) ParseTemplates(ctx context.Context, iface *Interface, srcPkg *packages.Package) error {
	log := zerolog.Ctx(ctx)

	mock := ""
	if iface != nil {
		mock = "mock"
		if ast.IsExported(iface.Name) {
			mock = "Mock"
		}
	}

	var (
		interfaceDir         string
		interfaceDirRelative string
		interfaceFile        string
		interfaceName        string
	)
	if iface != nil {
		interfaceFile = iface.FileName
		interfaceName = iface.Name

		workingDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		interfaceDirPath := pathlib.NewPath(iface.FileName).Parent()
		interfaceDir = interfaceDirPath.String()
		interfaceDirRelativePath, err := interfaceDirPath.RelativeToStr(workingDir)
		if err != nil {
			log.Debug().Err(err).Msg("can't make path relative to working dir, setting to './'")
			interfaceDirRelative = "."
		} else {
			interfaceDirRelative = interfaceDirRelativePath.String()
		}
	}
	// data is the struct sent to the template parser
	data := ConfigData{
		ConfigDir:            filepath.Dir(*c.ConfigFile),
		InterfaceDir:         interfaceDir,
		InterfaceDirRelative: interfaceDirRelative,
		InterfaceFile:        interfaceFile,
		InterfaceName:        interfaceName,
		Mock:                 mock,
		MockName:             *c.MockName,
		SrcPackageName:       srcPkg.Types.Name(),
		SrcPackagePath:       srcPkg.Types.Path(),
	}
	// These are the config options that we allow
	// to be parsed by the templater. The keys are
	// just labels we're using for logs/errors
	templateMap := map[string]*string{
		"filename": c.FileName,
		"dir":      c.Dir,
		"mockname": c.MockName,
		"pkgname":  c.PkgName,
	}

	changesMade := true
	for i := 0; changesMade; i++ {
		if i >= 20 {
			log.Error().Msg("infinite loop in template variables detected")
			for key, val := range templateMap {
				l := log.With().Str("variable-name", key).Str("variable-value", *val).Logger()
				l.Error().Msg("config variable value")
			}
			return ErrInfiniteLoop
		}
		// Templated variables can refer to other templated variables,
		// so we need to continue parsing the templates until it can't
		// be parsed anymore.
		changesMade = false

		for name, attributePointer := range templateMap {
			oldVal := *attributePointer

			attributeTempl, err := template.New("config-template").Funcs(StringManipulationFuncs).Parse(*attributePointer)
			if err != nil {
				return fmt.Errorf("failed to parse %s template: %w", name, err)
			}
			var parsedBuffer bytes.Buffer

			if err := attributeTempl.Execute(&parsedBuffer, data); err != nil {
				return fmt.Errorf("failed to execute %s template: %w", name, err)
			}
			*attributePointer = parsedBuffer.String()
			if *attributePointer != oldVal {
				changesMade = true
			}
		}
	}

	return nil
}

func (c *Config) GetReplacement(pkgPath string, typeName string) *ReplaceType {
	pkgMap := c.ReplaceType[pkgPath]
	if pkgMap == nil {
		return nil
	}
	return pkgMap[typeName]
}
