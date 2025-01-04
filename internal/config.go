package pkg

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

	"github.com/chigopher/pathlib"
	"github.com/go-viper/mapstructure/v2"
	"github.com/jinzhu/copier"
	koanfYAML "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/vektra/mockery/v3/internal/logging"
	"github.com/vektra/mockery/v3/internal/stackerr"
	mockeryTemplate "github.com/vektra/mockery/v3/template"
	"golang.org/x/tools/go/packages"
	"gopkg.in/yaml.v3"
)

type RootConfig struct {
	Config
	Packages map[string]PackageConfig `koanf:"packages"`
}

type PackageConfig struct {
	Config     Config                     `koanf:"config"`
	Interfaces map[string]InterfaceConfig `koanf:"interfaces"`
}

type InterfaceConfig struct {
	Config  Config   `koanf:"config"`
	Configs []Config `koanf:"configs"`
}

type Config struct {
	All             bool                   `koanf:"all"`
	Anchors         map[string]any         `koanf:"_anchors"`
	BoilerplateFile string                 `koanf:"boilerplate-file"`
	BuildTags       string                 `koanf:"tags"`
	Config          string                 `koanf:"config"`
	Dir             string                 `koanf:"dir"`
	Exclude         []string               `koanf:"exclude"`
	ExcludeRegex    string                 `koanf:"exclude-regex"`
	FileName        string                 `koanf:"filename"`
	Formatter       string                 `koanf:"formatter"`
	IncludeRegex    string                 `koanf:"include-regex"`
	LogLevel        string                 `koanf:"log-level"`
	MockBuildTags   string                 `koanf:"mock-build-tags"`
	MockName        string                 `koanf:"mockname"`
	PkgName         string                 `koanf:"pkgname"`
	Packages        map[string]interface{} `koanf:"packages"`
	Recursive       bool                   `koanf:"recursive"`
	Template        string                 `koanf:"template"`
	TemplateData    map[string]any         `koanf:"template-data"`
	UnrollVariadic  bool                   `koanf:"unroll-variadic"`
	Version         bool                   `koanf:"version"`
	// Viper throws away case-sensitivity when it marshals into this struct. This
	// destroys necessary information we need, specifically around interface names.
	// So, we re-read the config into this map outside of viper.
	// https://github.com/spf13/viper/issues/1014
	_cfgAsMap      map[string]any
	pkgConfigCache map[string]*Config
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

func NewConfig(configFile *pathlib.Path, flags *pflag.FlagSet) (*RootConfig, *koanf.Koanf, error) {
	// 2. Flags
	// 3. Config file
	var err error
	var rootConfig RootConfig
	k := koanf.New("::")
	if configFile == nil {
		configFile, err = findConfig()
		if err != nil {
			return nil, k, fmt.Errorf("discovering mockery config: %w", err)
		}
	}

	if flags != nil {
		if err := k.Load(posflag.Provider(flags, ".", k), nil); err != nil {
			return nil, k, fmt.Errorf("loading flags: %w", err)
		}
	}

	if err := k.Load(file.Provider(configFile.String()), koanfYAML.Parser()); err != nil {
		return nil, k, fmt.Errorf("loading config file: %w", err)
	}

	if err := k.Unmarshal("", &rootConfig); err != nil {
		return nil, k, fmt.Errorf("unmarshalling config: %w", err)
	}
	return &rootConfig, k, nil
}

func NewConfigFromViper(v *viper.Viper) (*Config, error) {
	c := &Config{
		Config: v.ConfigFileUsed(),
	}

	v.SetDefault("dir", "mocks/{{.SrcPackagePath}}")
	v.SetDefault("filename", "mock_{{.InterfaceName}}.go")
	v.SetDefault("formatter", "goimports")
	v.SetDefault("mockname", "Mock{{.InterfaceName}}")
	v.SetDefault("pkgname", "{{.SrcPackageName}}")
	v.SetDefault("log-level", "info")

	if err := v.UnmarshalExact(c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return c, nil
}

func (c *Config) Initialize(ctx context.Context) error {
	log := zerolog.Ctx(ctx)
	if err := c.discoverRecursivePackages(ctx); err != nil {
		return fmt.Errorf("failed to discover recursive packages: %w", err)
	}

	log.Trace().Msg("merging in config")
	if err := c.mergeInConfig(ctx); err != nil {
		return err
	}
	return nil
}

func (c *Config) FilePath(ctx context.Context) *pathlib.Path {
	return pathlib.NewPath(c.Dir).Join(c.FileName)
}

// CfgAsMap reads in the config file and returns a map representation, instead of a
// struct representation. This is mainly needed because viper throws away case-sensitivity
// in the `packages` section, which won't work when defining interface names 😞
func (c *Config) CfgAsMap(ctx context.Context) (map[string]any, error) {
	log := zerolog.Ctx(ctx)

	configPath := pathlib.NewPath(c.Config)

	if c._cfgAsMap == nil {
		log.Debug().Msgf("config map is nil, reading: %v", configPath)
		newCfg := make(map[string]any)

		fileBytes, err := os.ReadFile(configPath.String())
		if err != nil {
			if os.IsNotExist(err) {
				log.Debug().Msg("config file doesn't exist, returning empty config map")
				return map[string]any{}, nil
			}
			return nil, stackerr.NewStackErrf(err, "failed to read file: %v", configPath)
		}

		if err := yaml.Unmarshal(fileBytes, newCfg); err != nil {
			return nil, stackerr.NewStackErrf(err, "failed to unmarshal yaml")
		}
		c._cfgAsMap = newCfg
	}
	return c._cfgAsMap, nil
}

func (c *Config) getDecoder(result any) (*mapstructure.Decoder, error) {
	return mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused:          true,
		Result:               result,
		IgnoreUntaggedFields: true,
	})
}

// GetPackages returns a list of the packages that are defined in
// the `packages` config section.
func (c *Config) GetPackages(ctx context.Context) ([]string, error) {
	// NOTE: The reason why we can't rely on viper to get the
	// values in the `packages` section is because viper throws
	// away maps with no values. Our config allows empty maps,
	// so this breaks our logic. We need to manually parse this section
	// instead. See: https://github.com/spf13/viper/issues/819
	log := zerolog.Ctx(ctx)
	cfgMap, err := c.CfgAsMap(ctx)
	if err != nil {
		return nil, err
	}
	packagesSection, ok := cfgMap["packages"]
	if !ok {
		log.Debug().Msg("packages section is not defined")
		return []string{}, nil
	}
	packageSection, ok := packagesSection.(map[string]any)
	if !ok {
		msg := "packages section is of the wrong type"
		log.Error().Msg(msg)
		return []string{}, errors.New(msg)
	}
	packageList := []string{}
	for key := range packageSection {
		packageList = append(packageList, key)
	}
	return packageList, nil
}

// getPackageConfigMap returns the map for the particular package, which includes
// (but is not limited to) both the `configs` section and the `interfaces` section.
// Note this does NOT return the `configs` section for the package. It returns the
// entire mapping for the package.
func (c *Config) getPackageConfigMap(ctx context.Context, packageName string) (map[string]any, error) {
	log := zerolog.Ctx(ctx)
	log.Trace().Msg("getting package config map")

	cfgMap, err := c.CfgAsMap(ctx)
	if err != nil {
		return nil, err
	}
	packageSection := cfgMap["packages"].(map[string]any)
	configUnmerged, ok := packageSection[packageName]
	if !ok {
		return nil, ErrPkgNotFound
	}
	configAsMap, isMap := configUnmerged.(map[string]any)
	if isMap {
		log.Trace().Msg("package's value is a map, returning")
		return configAsMap, nil
	}
	log.Trace().Msg("package's value is not a map")

	// Package is something other than map, so set its value to an
	// empty map.
	emptyMap := map[string]any{}
	packageSection[packageName] = emptyMap
	return emptyMap, nil
}

// GetPackageConfig returns a struct representation of the package's config
// as provided in yaml. If the package did not specify a config section,
// this method will inject the top-level config into the package's config.
// This is especially useful as it allows us to lazily evaluate a package's
// config. If the package does specify config, this method takes care to merge
// the top-level config with the values specified for this package.
func (c *Config) GetPackageConfig(ctx context.Context, packageName string) (*Config, error) {
	log := zerolog.Ctx(ctx).With().Str("package-path", packageName).Logger()

	if c.pkgConfigCache == nil {
		log.Debug().Msg("package cache is nil")
		c.pkgConfigCache = make(map[string]*Config)
	} else if pkgConf, ok := c.pkgConfigCache[packageName]; ok {
		log.Debug().Msgf("package cache is not nil, returning cached result")
		return pkgConf, nil
	}

	pkgConfig := &Config{}
	if err := copier.Copy(pkgConfig, c); err != nil {
		return nil, fmt.Errorf("failed to copy config: %w", err)
	}

	configMap, err := c.getPackageConfigMap(ctx, packageName)
	if err != nil {
		return nil, stackerr.NewStackErrf(err, "unable to get map config for package")
	}

	configSection, ok := configMap["config"]
	if !ok {
		log.Debug().Msg("config section not provided for package")
		configMap["config"] = map[string]any{}
		c.pkgConfigCache[packageName] = pkgConfig
		return pkgConfig, nil
	}

	// We know that the package specified config that is overriding the top-level
	// config. We use a mapstructure decoder to decode the values in the yaml
	// into the pkgConfig struct. This has the effect of merging top-level
	// config with package-level config.
	decoder, err := c.getDecoder(pkgConfig)
	if err != nil {
		return nil, stackerr.NewStackErrf(err, "failed to get decoder")
	}
	if err := decoder.Decode(configSection); err != nil {
		return nil, err
	}
	c.pkgConfigCache[packageName] = pkgConfig
	return pkgConfig, nil
}

func (c *Config) ExcludePath(path string) bool {
	for _, ex := range c.Exclude {
		if strings.HasPrefix(path, ex) {
			return true
		}
	}
	return false
}

func (c *Config) ShouldGenerateInterface(ctx context.Context, packageName, interfaceName string) (bool, error) {
	pkgConfig, err := c.GetPackageConfig(ctx, packageName)
	if err != nil {
		return false, fmt.Errorf("getting package config: %w", err)
	}

	log := zerolog.Ctx(ctx)
	if pkgConfig.All {
		if pkgConfig.IncludeRegex != "" {
			log.Warn().Msg("interface config has both `all` and `include-regex` set: `include-regex` will be ignored")
		}
		if pkgConfig.ExcludeRegex != "" {
			log.Warn().Msg("interface config has both `all` and `exclude-regex` set: `exclude-regex` will be ignored")
		}
		return true, nil
	}

	interfacesSection, err := c.getInterfacesSection(ctx, packageName)
	if err != nil {
		return false, fmt.Errorf("getting interfaces section: %w", err)
	}
	_, interfaceExists := interfacesSection[interfaceName]
	if interfaceExists {
		return true, nil
	}

	includeRegex := pkgConfig.IncludeRegex
	excludeRegex := pkgConfig.ExcludeRegex
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
		return false, nil
	}
	if excludeRegex == "" {
		return true, nil
	}
	excludedByRegex, err := regexp.MatchString(excludeRegex, interfaceName)
	if err != nil {
		return false, fmt.Errorf("evaluating `exclude-regex`: %w", err)
	}
	return !excludedByRegex, nil
}

func (c *Config) getInterfacesSection(ctx context.Context, packageName string) (map[string]any, error) {
	pkgMap, err := c.getPackageConfigMap(ctx, packageName)
	if err != nil {
		return nil, err
	}
	interfaceSection, exists := pkgMap["interfaces"]
	if !exists {
		return make(map[string]any), nil
	}
	mapConfig, ok := interfaceSection.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("interfaces section has type %T, expected map[string]any", interfaceSection)
	}
	return mapConfig, nil
}

func (c *Config) GetInterfaceConfig(ctx context.Context, packageName string, interfaceName string) ([]*Config, error) {
	log := zerolog.
		Ctx(ctx).
		With().
		Str(logging.LogKeyQualifiedName, packageName).
		Str(logging.LogKeyInterface, interfaceName).
		Logger()
	ctx = log.WithContext(ctx)
	configs := []*Config{}

	pkgConfig, err := c.GetPackageConfig(ctx, packageName)
	if err != nil {
		return nil, stackerr.NewStackErrf(err, "failed to get config for package when iterating over interface")
	}
	interfacesSection, err := c.getInterfacesSection(ctx, packageName)
	if err != nil {
		return nil, err
	}

	// Copy the package-level config to our interface-level config
	pkgConfigCopy := &Config{}
	if err := copier.Copy(pkgConfigCopy, pkgConfig); err != nil {
		return nil, stackerr.NewStackErrf(err, "failed to create a copy of package config")
	}

	interfaceSection, ok := interfacesSection[interfaceName]
	if !ok {
		log.Debug().Msg("interface not defined in package configuration")
		return []*Config{pkgConfigCopy}, nil
	}

	interfaceSectionTyped, ok := interfaceSection.(map[string]any)
	if !ok {
		// check if it's an empty map... sometimes we just want to "enable"
		// the interface but not provide any additional config beyond what
		// is provided at the package level
		if reflect.ValueOf(&interfaceSection).Elem().IsZero() {
			return []*Config{pkgConfigCopy}, nil
		}
		msgString := "bad type provided for interface config"
		log.Error().Msg(msgString)
		return nil, stackerr.NewStackErr(errors.New(msgString))
	}

	configSection, ok := interfaceSectionTyped["config"]
	if ok {
		log.Debug().Msg("config section exists for interface")
		// if `config` is provided, we'll overwrite the values in our
		// pkgConfigCopy struct to act as the "new" base config.
		// This will allow us to set the default values for the interface
		// but override them further for each mock defined in the
		// `configs` section.
		decoder, err := c.getDecoder(pkgConfigCopy)
		if err != nil {
			return nil, stackerr.NewStackErrf(err, "unable to create mapstructure decoder")
		}
		if err := decoder.Decode(configSection); err != nil {
			return nil, stackerr.NewStackErrf(err, "unable to decode interface config")
		}
	} else {
		log.Debug().Msg("config section for interface doesn't exist")
	}

	configsSection, ok := interfaceSectionTyped["configs"]
	if ok {
		log.Debug().Msg("configs section exists for interface")
		configsSectionTyped := configsSection.([]any)
		for _, configMap := range configsSectionTyped {
			// Create a copy of the package-level config
			currentInterfaceConfig := reflect.New(reflect.ValueOf(pkgConfigCopy).Elem().Type()).Interface()
			if err := copier.Copy(currentInterfaceConfig, pkgConfigCopy); err != nil {
				return nil, stackerr.NewStackErrf(err, "failed to copy package config")
			}

			// decode the new values into the struct
			decoder, err := c.getDecoder(currentInterfaceConfig)
			if err != nil {
				return nil, stackerr.NewStackErrf(err, "unable to create mapstructure decoder")
			}
			if err := decoder.Decode(configMap); err != nil {
				return nil, stackerr.NewStackErrf(err, "unable to decode interface config")
			}

			configs = append(configs, currentInterfaceConfig.(*Config))
		}
		return configs, nil
	}
	log.Debug().Msg("configs section doesn't exist for interface")

	if len(configs) == 0 {
		configs = append(configs, pkgConfigCopy)
	}
	return configs, nil
}

// addSubPkgConfig injects the given pkgPath into the `packages` config section.
// You specify a parentPkgPath to inherit the config from.
func (c *Config) addSubPkgConfig(ctx context.Context, subPkgPath string, parentPkgPath string) error {
	log := zerolog.Ctx(ctx).With().
		Str("parent-package", parentPkgPath).
		Str("sub-package", subPkgPath).Logger()
	ctx = log.WithContext(ctx)

	log.Debug().Msg("adding sub-package to config map")
	parentPkgConfig, err := c.getPackageConfigMap(ctx, parentPkgPath)
	if err != nil {
		log.Err(err).
			Msg("failed to get package config for parent package")
		return fmt.Errorf("failed to get package config: %w", err)
	}

	log.Debug().Msg("getting config")
	topLevelConfig, err := c.CfgAsMap(ctx)
	if err != nil {
		return fmt.Errorf("failed to get configuration map: %w", err)
	}

	log.Debug().Msg("getting packages section")
	packagesSection := topLevelConfig["packages"].(map[string]any)

	_, pkgExists := packagesSection[subPkgPath]
	if !pkgExists {
		log.Trace().Msg("sub-package doesn't exist in config")

		// Copy the parent package directly into the subpackage config section
		packagesSection[subPkgPath] = map[string]any{}
		newPkgSection := packagesSection[subPkgPath].(map[string]any)
		newPkgSection["config"] = deepCopyConfigMap(parentPkgConfig["config"].(map[string]any))
	} else {
		log.Trace().Msg("sub-package exists in config")
		// The sub-package exists in config. Check if it has its
		// own `config` section and merge with the parent package
		// if so.
		subPkgConfig, err := c.getPackageConfigMap(ctx, subPkgPath)
		if err != nil {
			log.Err(err).Msg("could not get child package config")
			return fmt.Errorf("failed to get sub-package config: %w", err)
		}
		log.Trace().Msgf("sub-package config: %v", subPkgConfig)
		log.Trace().Msgf("parent-package config: %v", parentPkgConfig)

		// Merge the parent config with the sub-package config.
		parentConfigSection := parentPkgConfig["config"].(map[string]any)
		subPkgConfigSection := subPkgConfig["config"].(map[string]any)
		for key, val := range parentConfigSection {
			if _, keyInSubPkg := subPkgConfigSection[key]; !keyInSubPkg {
				subPkgConfigSection[key] = val
			}
		}
	}

	return nil
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

func shouldExcludeModule(ctx context.Context, root *pathlib.Path, goModPath *pathlib.Path) (bool, error) {
	log := zerolog.Ctx(ctx)
	relative, err := goModPath.RelativeTo(root)
	if err != nil {
		return false, stackerr.NewStackErrf(err, "determining distance from search root")
	}

	if len(relative.Parts()) != 1 {
		log.Debug().Msg("skipping sub-module")
		return true, nil
	}
	log.Debug().Int("parts_len", len(relative.Parts())).Str("parts", fmt.Sprintf("%v", relative.Parts())).Msg("not skipping module as this is the root path")
	return false, nil
}

func (c *Config) subPackages(pkgPath string) ([]string, error) {
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

// discoverRecursivePackages parses the provided config for packages marked as
// recursive and recurses the file tree to find all sub-packages.
func (c *Config) discoverRecursivePackages(ctx context.Context) error {
	log := zerolog.Ctx(ctx)
	log.Trace().Msg("discovering recursive packages")
	recursivePackages := map[string]*Config{}
	packageList, err := c.GetPackages(ctx)
	if err != nil {
		return fmt.Errorf("failed to get packages: %w", err)
	}
	for _, pkg := range packageList {
		pkgConfig, err := c.GetPackageConfig(ctx, pkg)
		pkgLog := log.With().Str("package", pkg).Logger()
		pkgLog.Trace().Msg("iterating over package")
		if err != nil {
			return fmt.Errorf("failed to get package config: %w", err)
		}
		if pkgConfig.Recursive {
			pkgLog.Trace().Msg("package marked as recursive")
			recursivePackages[pkg] = pkgConfig
		} else {
			pkgLog.Trace().Msg("package not marked as recursive")
		}
	}
	if len(recursivePackages) == 0 {
		return nil
	}
	for pkgPath, conf := range recursivePackages {
		pkgLog := log.With().Str("package-path", pkgPath).Logger()
		pkgCtx := pkgLog.WithContext(ctx)
		pkgLog.Debug().Msg("discovering sub-packages")
		subPkgs, err := c.subPackages(pkgPath)
		if err != nil {
			return fmt.Errorf("failed to get subpackages: %w", err)
		}
		for _, subPkg := range subPkgs {
			subPkgLog := pkgLog.With().Str("sub-package", subPkg).Logger()
			subPkgCtx := subPkgLog.WithContext(pkgCtx)

			if len(conf.Exclude) > 0 {
				// pass in the forward-slash as this is a package and the os.PathSeparator
				// cannot be used here as it fails on windows.
				p := pathlib.NewPath(subPkg, pathlib.PathWithSeperator("/"))
				relativePath, err := p.RelativeTo(
					pathlib.NewPath(
						pkgPath, pathlib.PathWithAfero(p.Fs()),
						pathlib.PathWithSeperator("/"),
					),
				)
				if err != nil {
					return stackerr.NewStackErrf(err, "failed to get path for %s relative to %s", subPkg, pkgPath)
				}
				if conf.ExcludePath(relativePath.String()) {
					subPkgLog.Info().Msg("subpackage is excluded")
					continue
				}
			}

			subPkgLog.Debug().Msg("adding sub-package config")
			if err := c.addSubPkgConfig(subPkgCtx, subPkg, pkgPath); err != nil {
				subPkgLog.Err(err).Msg("failed to add sub-package config")
				return fmt.Errorf("failed to add sub-package config: %w", err)
			}
		}
	}
	log.Trace().Msg("done discovering recursive packages")

	return nil
}

func contains[T comparable](slice []T, elem T) bool {
	for _, element := range slice {
		if elem == element {
			return true
		}
	}
	return false
}

func deepCopyConfigMap(src map[string]any) map[string]any {
	newMap := map[string]any{}
	for key, val := range src {
		if contains([]string{"packages", "config", "interfaces"}, key) {
			continue
		}
		newMap[key] = val
	}
	return newMap
}

// mergeInConfig takes care of merging inheritable configuration
// in the config map. For example, it merges default config, then
// package-level config, then interface-level config.
func (c *Config) mergeInConfig(ctx context.Context) error {
	log := zerolog.Ctx(ctx)

	log.Trace().Msg("getting packages")
	pkgs, err := c.GetPackages(ctx)
	if err != nil {
		return err
	}

	log.Trace().Msg("getting default config")
	defaultCfg, err := c.CfgAsMap(ctx)
	if err != nil {
		return err
	}
	for _, pkgPath := range pkgs {
		pkgLog := log.With().Str("package-path", pkgPath).Logger()
		pkgCtx := pkgLog.WithContext(ctx)

		pkgLog.Trace().Msg("merging for package")
		packageConfig, err := c.getPackageConfigMap(pkgCtx, pkgPath)
		if err != nil {
			pkgLog.Err(err).Msg("failed to get package config")
			return fmt.Errorf("failed to get package config: %w", err)
		}
		pkgLog.Trace().Msgf("got package config map: %v", packageConfig)

		configSectionUntyped, configExists := packageConfig["config"]
		if !configExists {
			// The reason why this should never happen is because getPackageConfigMap
			// should be populating the config section with the top-level config if it
			// wasn't defined in the yaml.
			msg := "config section does not exist for package, this should never happen"
			pkgLog.Error().Msg(msg)
			return errors.New(msg)
		}

		pkgLog.Trace().Msg("got config section for package")
		// Sometimes the config section may be provided, but it's nil.
		// We need to account for this fact.
		if configSectionUntyped == nil {
			pkgLog.Trace().Msg("config section is nil, converting to empty map")
			emptyMap := map[string]any{}

			// We need to add this to the "global" config mapping so the change
			// gets persisted, and also into configSectionUntyped for the logic
			// further down.
			packageConfig["config"] = emptyMap
			configSectionUntyped = emptyMap
		} else {
			pkgLog.Trace().Msg("config section is not nil")
		}

		configSectionTyped := configSectionUntyped.(map[string]any)

		for key, value := range defaultCfg {
			if contains([]string{"packages", "config"}, key) {
				continue
			}
			keyValLog := pkgLog.With().Str("key", key).Str("value", fmt.Sprintf("%v", value)).Logger()

			_, keyExists := configSectionTyped[key]
			if !keyExists {
				keyValLog.Trace().Msg("setting key to value")
				configSectionTyped[key] = value
			}
		}
		interfaces, err := c.getInterfacesForPackage(pkgCtx, pkgPath)
		if err != nil {
			return fmt.Errorf("failed to get interfaces for package: %w", err)
		}
		for _, interfaceName := range interfaces {
			interfacesSection, err := c.getInterfacesSection(pkgCtx, pkgPath)
			if err != nil {
				return err
			}
			interfaceSectionUntyped, exists := interfacesSection[interfaceName]
			if !exists {
				continue
			}
			interfaceSection, ok := interfaceSectionUntyped.(map[string]any)
			if !ok {
				// assume interfaceSection value is nil
				continue
			}

			interfaceConfigSectionUntyped, exists := interfaceSection["config"]
			if !exists {
				interfaceSection["config"] = map[string]any{}
			}

			interfaceConfigSection, ok := interfaceConfigSectionUntyped.(map[string]any)
			if !ok {
				// Assume this interface's value in the map is nil. Just skip it.
				continue
			}
			for key, value := range configSectionTyped {
				if key == "packages" {
					continue
				}
				if _, keyExists := interfaceConfigSection[key]; !keyExists {
					interfaceConfigSection[key] = value
				}
			}
		}
	}

	return nil
}

func (c *Config) getInterfacesForPackage(ctx context.Context, pkgPath string) ([]string, error) {
	interfaces := []string{}
	packageMap, err := c.getPackageConfigMap(ctx, pkgPath)
	if err != nil {
		return nil, err
	}
	interfacesUntyped, exists := packageMap["interfaces"]
	if !exists {
		return interfaces, nil
	}

	interfacesMap := interfacesUntyped.(map[string]any)
	for key := range interfacesMap {
		interfaces = append(interfaces, key)
	}
	return interfaces, nil
}

func (c *Config) TagName(name string) string {
	field, ok := reflect.TypeOf(c).Elem().FieldByName(name)
	if !ok {
		panic(fmt.Sprintf("unknown config field: %s", name))
	}
	return string(field.Tag.Get("mapstructure"))
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
			return stackerr.NewStackErr(err)
		}
		interfaceDirRelative = interfaceDirRelativePath.String()
	}
	// data is the struct sent to the template parser
	data := mockeryTemplate.ConfigData{
		ConfigDir:            filepath.Dir(c.Config),
		InterfaceDir:         interfaceDir,
		InterfaceDirRelative: interfaceDirRelative,
		InterfaceFile:        interfaceFile,
		InterfaceName:        interfaceName,
		Mock:                 mock,
		MockName:             c.MockName,
		SrcPackageName:       srcPkg.Types.Name(),
		SrcPackagePath:       srcPkg.Types.Path(),
	}
	// These are the config options that we allow
	// to be parsed by the templater. The keys are
	// just labels we're using for logs/errors
	templateMap := map[string]*string{
		"filename": &c.FileName,
		"dir":      &c.Dir,
		"mockname": &c.MockName,
		"pkgname":  &c.PkgName,
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

			attributeTempl, err := template.New("config-template").Funcs(mockeryTemplate.StringManipulationFuncs).Parse(*attributePointer)
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

func (c *Config) ClearCfgAsMap() {
	c._cfgAsMap = nil
}
