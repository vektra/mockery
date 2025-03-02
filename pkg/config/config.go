package config

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/chigopher/pathlib"
	"github.com/jinzhu/copier"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/vektra/mockery/v2/pkg/logging"
	"github.com/vektra/mockery/v2/pkg/stackerr"
	"golang.org/x/tools/go/packages"
	"gopkg.in/yaml.v3"
)

var (
	ErrNoConfigFile         = fmt.Errorf("no config file exists")
	ErrNoGoFilesFoundInRoot = fmt.Errorf("no go files found in root search path")
	ErrPkgNotFound          = fmt.Errorf("package not found in config")
)

type Interface struct {
	Config Config `mapstructure:"config"`
}

type Config struct {
	All                         bool                   `mapstructure:"all"`
	Anchors                     map[string]any         `mapstructure:"_anchors"`
	BoilerplateFile             string                 `mapstructure:"boilerplate-file"`
	BuildTags                   string                 `mapstructure:"tags"`
	Case                        string                 `mapstructure:"case"`
	Config                      string                 `mapstructure:"config"`
	Cpuprofile                  string                 `mapstructure:"cpuprofile"`
	Dir                         string                 `mapstructure:"dir"`
	DisableConfigSearch         bool                   `mapstructure:"disable-config-search"`
	DisableDeprecationWarnings  bool                   `mapstructure:"disable-deprecation-warnings"`
	DisabledDeprecationWarnings []string               `mapstructure:"disabled-deprecation-warnings"`
	DisableFuncMocks            bool                   `mapstructure:"disable-func-mocks"`
	DisableVersionString        bool                   `mapstructure:"disable-version-string"`
	DryRun                      bool                   `mapstructure:"dry-run"`
	Exclude                     []string               `mapstructure:"exclude"`
	ExcludeRegex                string                 `mapstructure:"exclude-regex"`
	Exported                    bool                   `mapstructure:"exported"`
	FailOnMissing               bool                   `mapstructure:"fail-on-missing"`
	FileName                    string                 `mapstructure:"filename"`
	InPackage                   bool                   `mapstructure:"inpackage"`
	InPackageSuffix             bool                   `mapstructure:"inpackage-suffix"`
	IncludeAutoGenerated        bool                   `mapstructure:"include-auto-generated"`
	IncludeRegex                string                 `mapstructure:"include-regex"`
	Issue845Fix                 bool                   `mapstructure:"issue-845-fix"`
	KeepTree                    bool                   `mapstructure:"keeptree"`
	LogLevel                    string                 `mapstructure:"log-level"`
	MockBuildTags               string                 `mapstructure:"mock-build-tags"`
	MockName                    string                 `mapstructure:"mockname"`
	Name                        string                 `mapstructure:"name"`
	Note                        string                 `mapstructure:"note"`
	Outpkg                      string                 `mapstructure:"outpkg"`
	Output                      string                 `mapstructure:"output"`
	Packageprefix               string                 `mapstructure:"packageprefix"`
	Packages                    map[string]interface{} `mapstructure:"packages"`
	Print                       bool                   `mapstructure:"print"`
	Profile                     string                 `mapstructure:"profile"`
	Quiet                       bool                   `mapstructure:"quiet"`
	Recursive                   bool                   `mapstructure:"recursive"`
	ReplaceType                 []string               `mapstructure:"replace-type"`
	ResolveTypeAlias            bool                   `mapstructure:"resolve-type-alias"`
	SrcPkg                      string                 `mapstructure:"srcpkg"`
	// StructName overrides the name given to the mock struct and should only be nonempty
	// when generating for an exact match (non regex expression in -name).
	StructName     string `mapstructure:"structname"`
	TestOnly       bool   `mapstructure:"testonly"`
	UnrollVariadic bool   `mapstructure:"unroll-variadic"`
	Version        bool   `mapstructure:"version"`
	WithExpecter   bool   `mapstructure:"with-expecter"`
	// Viper throws away case-sensitivity when it marshals into this struct. This
	// destroys necessary information we need, specifically around interface names.
	// So, we re-read the config into this map outside of viper.
	// https://github.com/spf13/viper/issues/1014
	_cfgAsMap      map[string]any
	pkgConfigCache map[string]*Config
}

func NewConfigFromViper(v *viper.Viper) (*Config, error) {
	c := &Config{
		Config: v.ConfigFileUsed(),
	}

	packageList, err := c.GetPackages(context.Background())
	if err != nil {
		return c, fmt.Errorf("failed to get packages: %w", err)
	}

	// Set defaults
	v.SetDefault("resolve-type-alias", true)
	if len(packageList) == 0 {
		v.SetDefault("case", "camel")
		v.SetDefault("dir", ".")
		v.SetDefault("output", "./mocks")
	} else {
		v.SetDefault("dir", "mocks/{{.PackagePath}}")
		v.SetDefault("filename", "mock_{{.InterfaceName}}.go")
		v.SetDefault("include-auto-generated", true)
		v.SetDefault("mockname", "Mock{{.InterfaceName}}")
		v.SetDefault("outpkg", "{{.PackageName}}")
		v.SetDefault("with-expecter", true)
		v.SetDefault("dry-run", false)
		v.SetDefault("log-level", "info")
	}

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
		parentConfigSection, ok := parentPkgConfig["config"].(map[string]any)
		if !ok {
			parentConfigSection = map[string]any{}
		}
		subPkgConfigSection, ok := subPkgConfig["config"].(map[string]any)
		if !ok {
			subPkgConfigSection = map[string]any{}
		}
		for key, val := range parentConfigSection {
			if _, keyInSubPkg := subPkgConfigSection[key]; !keyInSubPkg {
				subPkgConfigSection[key] = val
			}
		}
	}

	return nil
}

func isAutoGenerated(path *pathlib.Path) (bool, error) {
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
		interfaces, err := c.GetInterfacesForPackage(pkgCtx, pkgPath)
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

func (c *Config) GetInterfacesForPackage(ctx context.Context, pkgPath string) ([]string, error) {
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

// LogUnsupportedPackagesConfig is a method that will help aid migrations to the
// packages config feature. This is intended to be a temporary measure until v3
// when we can remove all legacy config options.
func (c *Config) LogUnsupportedPackagesConfig(ctx context.Context) {
	log := zerolog.Ctx(ctx)
	unsupportedOptions := make(map[string]any)
	for _, name := range []string{"Name", "KeepTree", "Case", "Output", "TestOnly"} {
		value := reflect.ValueOf(c).Elem().FieldByName(name)
		var valueAsString string
		if value.Kind().String() == "bool" {
			valueAsString = fmt.Sprintf("%v", value.Bool())
		}
		if value.Kind().String() == "string" {
			valueAsString = value.String()
		}

		if !value.IsZero() {
			unsupportedOptions[c.TagName(name)] = valueAsString
		}
	}
	if len(unsupportedOptions) == 0 {
		return
	}

	l := log.With().
		Dict("unsupported-fields", zerolog.Dict().Fields(unsupportedOptions)).
		Str("url", logging.DocsURL("/configuration/#parameter-descriptions")).
		Logger()
	l.Error().Msg("use of unsupported options detected. mockery behavior is undefined.")
}

func (c *Config) LogDeprecatedConfig(ctx context.Context) {
	if !c.WithExpecter {
		logging.WarnDeprecated(
			"with-expecter",
			"with-expecter will be permanently set to True in v3",
			nil,
		)
	}
	if c.Quiet {
		logging.WarnDeprecated(
			"quiet",
			"The --quiet parameter will be removed in v3. Use --log-level=\"\" instead",
			nil,
		)
	}
	if c.ResolveTypeAlias {
		logging.WarnDeprecated(
			"resolve-type-alias",
			"resolve-type-alias will be permanently set to False in v3. Please modify your config to set the parameter to False.",
			nil,
		)
	}
	if c.DisableVersionString {
		logging.WarnDeprecated(
			"disable-version-string",
			"disable-version-string will be permanently set to True in v3",
			nil,
		)
	}
	if c.StructName != "" {
		logging.WarnDeprecated(
			"structname",
			"structname will be removed as a parameter in v3",
			nil,
		)
	}
}
