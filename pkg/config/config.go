package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/chigopher/pathlib"
	"github.com/jinzhu/copier"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/vektra/mockery/v2/pkg/logging"
	"gopkg.in/yaml.v3"
)

type Interface struct {
	Config Config `mapstructure:"config"`
}

type Package struct {
	Config     Config               `mapstructure:"config"`
	Interfaces map[string]Interface `mapstructure:"interfaces"`
}

type Packages map[string]Package

type Config struct {
	All                  bool                   `mapstructure:"all"`
	BuildTags            string                 `mapstructure:"tags"`
	Case                 string                 `mapstructure:"case"`
	Config               string                 `mapstructure:"config"`
	Cpuprofile           string                 `mapstructure:"cpuprofile"`
	Dir                  string                 `mapstructure:"dir"`
	DisableConfigSearch  bool                   `mapstructure:"disable-config-search"`
	DisableVersionString bool                   `mapstructure:"disable-version-string"`
	DryRun               bool                   `mapstructure:"dry-run"`
	Exported             bool                   `mapstructure:"exported"`
	FileName             string                 `mapstructure:"filename"`
	InPackage            bool                   `mapstructure:"inpackage"`
	InPackageSuffix      bool                   `mapstructure:"inpackage-suffix"`
	KeepTree             bool                   `mapstructure:"keeptree"`
	LogLevel             string                 `mapstructure:"log-level"`
	Name                 string                 `mapstructure:"name"`
	Note                 string                 `mapstructure:"note"`
	Outpkg               string                 `mapstructure:"outpkg"`
	Output               string                 `mapstructure:"output"`
	Packages             map[string]interface{} `mapstructure:"packages"`
	Packageprefix        string                 `mapstructure:"packageprefix"`
	Print                bool                   `mapstructure:"print"`
	Profile              string                 `mapstructure:"profile"`
	Quiet                bool                   `mapstructure:"quiet"`
	Recursive            bool                   `mapstructure:"recursive"`
	SrcPkg               string                 `mapstructure:"srcpkg"`
	BoilerplateFile      string                 `mapstructure:"boilerplate-file"`
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

// cfgAsMap reads in the config file and returns a map representation, instead of a
// struct representation. This is mainly needed because viper throws away case-sensitivity
// in the `packages` section, which won't work when defining interface names 😞
func (c *Config) cfgAsMap(ctx context.Context) (map[string]any, error) {
	log := zerolog.Ctx(ctx)

	configPath := pathlib.NewPath(c.Config)

	if c._cfgAsMap == nil {
		log.Debug().Msgf("config map is nil, reading: %v", configPath)
		newCfg := make(map[string]any)

		fileBytes, err := ioutil.ReadFile(configPath.String())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read file: %v", configPath)
		}

		if err := yaml.Unmarshal(fileBytes, newCfg); err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal yaml")
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
	cfgMap, err := c.cfgAsMap(ctx)
	if err != nil {
		return nil, err
	}
	packageSection, ok := cfgMap["packages"].(map[string]any)
	if !ok {
		return []string{}, nil
	}
	packages := []string{}
	for key := range packageSection {
		packages = append(packages, key)
	}
	return packages, nil
}

func (c *Config) getPackageConfigMap(ctx context.Context, packageName string) (map[string]any, error) {
	cfgMap, err := c.cfgAsMap(ctx)
	if err != nil {
		return nil, err
	}
	packageSection := cfgMap["packages"].(map[string]any)
	configUnmerged, ok := packageSection[packageName]
	if !ok {
		return nil, fmt.Errorf("package %s is not found in config", packageName)
	}
	return configUnmerged.(map[string]any), nil

}
func (c *Config) GetPackageConfig(ctx context.Context, packageName string) (*Config, error) {
	log := zerolog.Ctx(ctx)

	if c.pkgConfigCache == nil {
		log.Debug().Msg("package cache is nil")
		c.pkgConfigCache = make(map[string]*Config)
	} else if pkgConf, ok := c.pkgConfigCache[packageName]; ok {
		log.Debug().Msgf("package cache is not nil, returning cached result")
		return pkgConf, nil
	}

	pkgConfig := reflect.New(reflect.ValueOf(c).Elem().Type()).Interface()
	copier.Copy(pkgConfig, c)
	pkgConfigTyped := pkgConfig.(*Config)

	configMap, err := c.getPackageConfigMap(ctx, packageName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get map config for package")
	}

	configSection, ok := configMap["config"]
	if !ok {
		log.Debug().Msg("config section not provided for package")
		return pkgConfigTyped, nil
	}

	decoder, err := c.getDecoder(pkgConfigTyped)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get decoder")
	}
	if err := decoder.Decode(configSection); err != nil {
		return nil, err
	}
	c.pkgConfigCache[packageName] = pkgConfigTyped
	return pkgConfigTyped, nil
}

func (c *Config) ShouldGenerateInterface(ctx context.Context, packageName, interfaceName string) (bool, error) {
	pkgConfig, err := c.GetPackageConfig(ctx, packageName)
	if err != nil {
		return false, err
	}

	interfacesSection, err := c.getInterfacesSection(ctx, packageName, interfaceName)
	if err != nil {
		return false, err
	}
	_, interfaceExists := interfacesSection[interfaceName]
	return pkgConfig.All || interfaceExists, nil
}

func (c *Config) getInterfacesSection(ctx context.Context, packageName string, interfaceName string) (map[string]any, error) {
	pkgMap, err := c.getPackageConfigMap(ctx, packageName)
	if err != nil {
		return nil, err
	}
	interfaceSection, exists := pkgMap["interfaces"]
	if !exists {
		return make(map[string]any), nil
	}
	return interfaceSection.(map[string]any), nil
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
		return nil, errors.Wrapf(err, "failed to get config for package when iterating over interface")
	}
	interfacesSection, err := c.getInterfacesSection(ctx, packageName, interfaceName)
	if err != nil {
		return nil, err
	}

	// Copy the package-level config to our interface-level config
	pkgConfigCopy := reflect.New(reflect.ValueOf(pkgConfig).Elem().Type()).Interface()
	if err := copier.Copy(pkgConfigCopy, pkgConfig); err != nil {
		return nil, errors.Wrap(err, "failed to create a copy of package config")
	}
	baseConfigTyped := pkgConfigCopy.(*Config)

	interfaceSection, ok := interfacesSection[interfaceName]
	if !ok {
		log.Debug().Msg("interface not defined in package configuration")
		return []*Config{baseConfigTyped}, nil
	}

	interfaceSectionTyped, ok := interfaceSection.(map[string]any)
	if !ok {
		// check if it's an empty map... sometimes we just want to "enable"
		// the interface but not provide any additional config beyond what
		// is provided at the package level
		if reflect.ValueOf(&interfaceSection).Elem().IsZero() {
			return []*Config{baseConfigTyped}, nil
		}
		msgString := "bad type provided for interface config"
		log.Error().Msgf(msgString)
		return nil, errors.New(msgString)
	}

	configSection, ok := interfaceSectionTyped["config"]
	if ok {
		log.Debug().Msg("config section exists for interface")
		// if `config` is provided, we'll overwrite the values in our
		// baseConfigTyped struct to act as the "new" base config.
		// This will allow us to set the default values for the interface
		// but override them further for each mock defined in the
		// `configs` section.
		decoder, err := c.getDecoder(baseConfigTyped)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to create mapstructure decoder")
		}
		if err := decoder.Decode(configSection); err != nil {
			return nil, errors.Wrapf(err, "unable to decode interface config")
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
			currentInterfaceConfig := reflect.New(reflect.ValueOf(baseConfigTyped).Elem().Type()).Interface()
			if err := copier.Copy(currentInterfaceConfig, baseConfigTyped); err != nil {
				return nil, errors.Wrap(err, "failed to copy package config")
			}

			// decode the new values into the struct
			decoder, err := c.getDecoder(currentInterfaceConfig)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to create mapstructure decoder")
			}
			if err := decoder.Decode(configMap); err != nil {
				return nil, errors.Wrapf(err, "unable to decode interface config")
			}

			configs = append(configs, currentInterfaceConfig.(*Config))
		}
		return configs, nil
	}
	log.Debug().Msg("configs section doesn't exist for interface")

	if len(configs) == 0 {
		configs = append(configs, baseConfigTyped)
	}
	return configs, nil
}
