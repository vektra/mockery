package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/chigopher/pathlib"
	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	pkg "github.com/vektra/mockery/v3/internal"
	"github.com/vektra/mockery/v3/internal/logging"
	"github.com/vektra/mockery/v3/internal/stackerr"
	"golang.org/x/tools/go/packages"
)

var (
	cfgFile            = ""
	ErrCfgFileNotFound = errors.New("config file not found")
)

func NewRootCmd() (*cobra.Command, error) {
	viperCfg, err := getConfig(nil, nil)
	if err != nil && !errors.Is(err, ErrCfgFileNotFound) {
		return nil, err
	}
	cmd := &cobra.Command{
		Use:   "mockery",
		Short: "Generate mock objects for your Golang interfaces",
		Run: func(cmd *cobra.Command, args []string) {
			r, err := GetRootAppFromViper(viperCfg)
			if err != nil {
				printStackTrace(err)
				os.Exit(1)
			}
			if err := r.Run(); err != nil {
				printStackTrace(err)
				os.Exit(1)
			}
		},
	}

	pFlags := cmd.PersistentFlags()
	pFlags.StringVar(&cfgFile, "config", "", "config file to use")
	pFlags.String("tags", "", "space-separated list of additional build tags to load packages")
	pFlags.String("mock-build-tags", "", "set the build tags of the generated mocks. Read more about the format: https://pkg.go.dev/cmd/go#hdr-Build_constraints")
	pFlags.String("log-level", "info", "Level of logging")
	pFlags.String("boilerplate-file", "", "File to read a boilerplate text from. Text should be a go block comment, i.e. /* ... */")
	pFlags.Bool("unroll-variadic", true, "For functions with variadic arguments, do not unroll the arguments into the underlying testify call. Instead, pass variadic slice as-is.")

	if err := viperCfg.BindPFlags(pFlags); err != nil {
		panic(fmt.Sprintf("failed to bind PFlags: %v", err))
	}

	cmd.AddCommand(NewShowConfigCmd())
	cmd.AddCommand(NewVersionCmd())
	return cmd, nil
}

func printStackTrace(e error) {
	fmt.Printf("%v\n", e)

	if stack, ok := stackerr.GetStack(e); ok {
		fmt.Printf("%+s\n", stack)
	}
}

// Execute executes the cobra CLI workflow
func Execute() {
	cmd, err := NewRootCmd()
	if err != nil {
		os.Exit(1)
	}
	if cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func getConfig(
	baseSearchPath *pathlib.Path,
	configPath *pathlib.Path,
) (*viper.Viper, error) {
	viperObj := viper.NewWithOptions(viper.KeyDelimiter("::"))
	if baseSearchPath == nil {
		currentWorkingDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		baseSearchPath = pathlib.NewPath(currentWorkingDir)
	}
	if viperObj == nil {
		viperObj = viper.NewWithOptions(viper.KeyDelimiter("::"))
	}

	viperObj.SetEnvPrefix("MOCKERY")
	viperObj.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viperObj.AutomaticEnv()

	if configPath == nil && cfgFile != "" {
		// Use config file from the flag.
		viperObj.SetConfigFile(cfgFile)
	} else if configPath != nil {
		viperObj.SetConfigFile(configPath.String())
	} else if viperObj.IsSet("config") {
		viperObj.SetConfigFile(viperObj.GetString("config"))
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to find homedir")
		}

		currentDir := baseSearchPath

		for {
			viperObj.AddConfigPath(currentDir.String())
			if len(currentDir.Parts()) <= 1 {
				break
			}
			currentDir = currentDir.Parent()
		}

		viperObj.AddConfigPath(home)
		viperObj.SetConfigName(".mockery")
	}
	if err := viperObj.ReadInConfig(); err != nil {
		log, _ := logging.GetLogger("debug")
		log.Debug().Err(err).Msg("couldn't read any config file")
		return viperObj, ErrCfgFileNotFound
	}

	viperObj.Set("config", viperObj.ConfigFileUsed())
	return viperObj, nil
}

type RootApp struct {
	pkg.Config
}

func GetRootAppFromViper(v *viper.Viper) (*RootApp, error) {
	r := &RootApp{}
	config, err := pkg.NewConfigFromViper(v)
	if err != nil {
		return nil, stackerr.NewStackErrf(err, "failed to get config")
	}
	r.Config = *config
	return r, nil
}

// InterfaceCollection maintains a list of *pkg.Interface and asserts that all
// the interfaces in the collection belong to the same source package. It also
// asserts that various properties of the interfaces added to the collection are
// uniform.
type InterfaceCollection struct {
	// Mockery needs to assert that certain properties of the added interfaces
	// are uniform for all members of the collection. This includes things like
	// 1. Package name of the output mock file
	// 2. Source package path (only one package per output file is allowed)
	srcPkgPath  string
	outFilePath *pathlib.Path
	srcPkg      *packages.Package
	pkgName     string
	interfaces  []*pkg.Interface
}

func NewInterfaceCollection(
	srcPkgPath string,
	outFilePath *pathlib.Path,
	srcPkg *packages.Package,
	pkgName string,
) *InterfaceCollection {
	return &InterfaceCollection{
		srcPkgPath:  srcPkgPath,
		outFilePath: outFilePath,
		srcPkg:      srcPkg,
		interfaces:  make([]*pkg.Interface, 0),
	}
}

func (i *InterfaceCollection) Append(ctx context.Context, iface *pkg.Interface) error {
	log := zerolog.Ctx(ctx).With().
		Str(logging.LogKeyInterface, iface.Name).
		Str(logging.LogKeyPackageName, iface.Pkg.Name).
		Str(logging.LogKeyPackagePath, iface.Pkg.PkgPath).
		Str("expected-package-path", i.srcPkgPath).Logger()
	if iface.Pkg.PkgPath != i.srcPkgPath {
		msg := "cannot mix interfaces from different packages in the same file."
		log.Error().Msg(msg)
		return errors.New(msg)
	}
	if i.outFilePath.String() != pathlib.NewPath(iface.Config.Dir).Join(iface.Config.FileName).String() {
		msg := "all mocks within an InterfaceCollection must have the same output file path"
		log.Error().Msg(msg)
		return errors.New(msg)
	}
	if i.pkgName != iface.Config.PkgName {
		msg := "all mocks in an output file must have the same pkgname"
		log.Error().Str("expected-pkgname", i.pkgName).Str("interface-pkgname", iface.Config.PkgName).Msg(msg)
		return errors.New(msg)
	}
	i.interfaces = append(i.interfaces, iface)
	return nil
}

func (r *RootApp) Run() error {
	log, err := logging.GetLogger(r.Config.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		return err
	}
	log.Info().Msgf("Starting mockery")
	if r.Config.Config == "" {
		log.Error().Msg("Config file not found. Mockery requires a config file to run.")
		return ErrCfgFileNotFound
	}
	log.Info().Msgf("Using config: %s", r.Config.Config)
	ctx := log.WithContext(context.Background())

	if err := r.Config.Initialize(ctx); err != nil {
		return err
	}

	buildTags := strings.Split(r.Config.BuildTags, " ")

	configuredPackages, err := r.Config.GetPackages(ctx)
	if err != nil {
		return fmt.Errorf("failed to get package from config: %w", err)
	}
	if len(configuredPackages) == 0 {
		log.Error().Msg("no packages specified in config")
		return nil
	}
	parser := pkg.NewParser(buildTags)

	log.Info().Msg("Parsing configured packages...")
	interfaces, err := parser.ParsePackages(ctx, configuredPackages)
	if err != nil {
		log.Error().Err(err).Msg("unable to parse packages")
		return err
	}
	log.Info().Msg("Done parsing configured packages.")
	// maps the following:
	// outputFilePath|fullyQualifiedInterfaceName|[]*pkg.Interface
	// The reason why we need an interior map of fully qualified interface name
	// to a slice of *pkg.Interface (which represents all information necessary
	// to create the output mock) is because mockery allows multiple mocks to be
	// created for each input interface.
	mockFileToInterfaces := map[string]*InterfaceCollection{}

	for _, iface := range interfaces {
		ifaceLog := log.
			With().
			Str(logging.LogKeyInterface, iface.Name).
			Str(logging.LogKeyQualifiedName, iface.Pkg.Types.Path()).
			Logger()

		ifaceCtx := ifaceLog.WithContext(ctx)

		shouldGenerate, err := r.Config.ShouldGenerateInterface(ifaceCtx, iface.Pkg.Types.Path(), iface.Name)
		if err != nil {
			return err
		}
		if !shouldGenerate {
			ifaceLog.Debug().Msg("config doesn't specify to generate this interface, skipping.")
			continue
		}
		ifaceLog.Info().Msg("Adding interface")
		ifaceConfigs, err := r.Config.GetInterfaceConfig(ctx, iface.Pkg.PkgPath, iface.Name)
		if err != nil {
			return err
		}
		for _, ifaceConfig := range ifaceConfigs {
			if err := ifaceConfig.ParseTemplates(ctx, iface, iface.Pkg); err != nil {
				log.Err(err).Msg("Can't parse config templates for interface")
				return err
			}
			filePath := ifaceConfig.FilePath(ctx).String()

			_, ok := mockFileToInterfaces[filePath]
			if !ok {
				mockFileToInterfaces[filePath] = NewInterfaceCollection(
					iface.Pkg.PkgPath,
					pathlib.NewPath(ifaceConfig.Dir).Join(ifaceConfig.FileName),
					iface.Pkg,
					iface.Config.PkgName,
				)
			}
			mockFileToInterfaces[filePath].Append(
				ctx,
				pkg.NewInterface(
					iface.Name,
					iface.FileName,
					iface.File,
					iface.Pkg,
					ifaceConfig),
			)
		}
	}

	for outFilePath, interfacesInFile := range mockFileToInterfaces {
		fileLog := log.With().Str("file", outFilePath).Logger()
		fileCtx := fileLog.WithContext(ctx)

		fileLog.Debug().Int("interfaces-in-file-len", len(interfacesInFile.interfaces)).Msgf("%v", interfacesInFile)
		packageConfig, err := r.Config.GetPackageConfig(fileCtx, interfacesInFile.srcPkgPath)
		if err != nil {
			return err
		}
		if err := packageConfig.ParseTemplates(ctx, nil, interfacesInFile.srcPkg); err != nil {
			return err
		}
		generator, err := pkg.NewTemplateGenerator(
			fileCtx,
			interfacesInFile.interfaces[0].Pkg,
			interfacesInFile.outFilePath.Parent(),
			packageConfig.Template,
			pkg.Formatter(r.Config.Formatter),
			interfacesInFile.pkgName,
		)
		if err != nil {
			return err
		}
		fileLog.Info().Msg("Executing template")
		templateBytes, err := generator.Generate(fileCtx, interfacesInFile.interfaces)
		if err != nil {
			return err
		}

		outFile := pathlib.NewPath(outFilePath)
		if err := outFile.Parent().MkdirAll(); err != nil {
			log.Err(err).Msg("failed to mkdir parent directories of mock file")
			return stackerr.NewStackErr(err)
		}
		fileLog.Info().Msg("Writing template to file")
		if err := outFile.WriteFile(templateBytes); err != nil {
			return stackerr.NewStackErr(err)
		}
	}

	return nil
}
