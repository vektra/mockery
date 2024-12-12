package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/chigopher/pathlib"
	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vektra/mockery/v2/pkg"
	"github.com/vektra/mockery/v2/pkg/logging"
	"github.com/vektra/mockery/v2/pkg/stackerr"
)

var (
	cfgFile  = ""
	viperCfg *viper.Viper
)

func init() {
	cobra.OnInitialize(func() { initConfig(nil, viperCfg, nil) })
}

func NewRootCmd() *cobra.Command {
	viperCfg = viper.NewWithOptions(viper.KeyDelimiter("::"))
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
	pFlags.String("outpkg", "mocks", "name of generated package")
	pFlags.String("packageprefix", "", "prefix for the generated package name, it is ignored if outpkg is also specified.")
	pFlags.String("dir", "", "directory to search for interfaces")
	pFlags.BoolP("recursive", "r", false, "recurse search into sub-directories")
	pFlags.StringArray("exclude", nil, "prefixes of subdirectories and files to exclude from search")
	pFlags.Bool("all", false, "generates mocks for all found interfaces in all sub-directories")
	pFlags.String("note", "", "comment to insert into prologue of each generated file")
	pFlags.String("cpuprofile", "", "write cpu profile to file")
	pFlags.Bool("version", false, "prints the installed version of mockery")
	pFlags.String("tags", "", "space-separated list of additional build tags to load packages")
	pFlags.String("mock-build-tags", "", "set the build tags of the generated mocks. Read more about the format: https://pkg.go.dev/cmd/go#hdr-Build_constraints")
	pFlags.String("filename", "", "name of generated file (only works with -name and no regex)")
	pFlags.String("structname", "", "name of generated struct (only works with -name and no regex)")
	pFlags.String("log-level", "info", "Level of logging")
	pFlags.String("srcpkg", "", "source pkg to search for interfaces")
	pFlags.BoolP("dry-run", "d", false, "Do a dry run, don't modify any files")
	pFlags.Bool("disable-version-string", false, "Do not insert the version string into the generated mock file.")
	pFlags.String("boilerplate-file", "", "File to read a boilerplate text from. Text should be a go block comment, i.e. /* ... */")
	pFlags.Bool("unroll-variadic", true, "For functions with variadic arguments, do not unroll the arguments into the underlying testify call. Instead, pass variadic slice as-is.")

	if err := viperCfg.BindPFlags(pFlags); err != nil {
		panic(fmt.Sprintf("failed to bind PFlags: %v", err))
	}

	cmd.AddCommand(NewShowConfigCmd())
	return cmd
}

func printStackTrace(e error) {
	fmt.Printf("%v\n", e)

	if stack, ok := stackerr.GetStack(e); ok {
		fmt.Printf("%+s\n", stack)
	}
}

// Execute executes the cobra CLI workflow
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func initConfig(
	baseSearchPath *pathlib.Path,
	viperObj *viper.Viper,
	configPath *pathlib.Path,
) *viper.Viper {
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

	if !viperObj.GetBool("disable-config-search") {
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
			log.Info().Msg("couldn't read any config file")
		}
	}

	viperObj.Set("config", viperObj.ConfigFileUsed())
	return viperObj
}

const regexMetadataChars = "\\.+*?()|[]{}^$"

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

func (r *RootApp) Run() error {
	log, err := logging.GetLogger(r.Config.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		return err
	}
	log = log.With().Bool(logging.LogKeyDryRun, r.Config.DryRun).Logger()
	log.Info().Msgf("Starting mockery")
	log.Info().Msgf("Using config: %s", r.Config.Config)
	ctx := log.WithContext(context.Background())

	if err := r.Config.Initialize(ctx); err != nil {
		return err
	}

	if r.Config.Version {
		fmt.Println(logging.GetSemverInfo())
		return nil
	}
	buildTags := strings.Split(r.Config.BuildTags, " ")

	var boilerplate string
	if r.Config.BoilerplateFile != "" {
		data, err := os.ReadFile(r.Config.BoilerplateFile)
		if err != nil {
			log.Fatal().Msgf("Failed to read boilerplate file %s: %v", r.Config.BoilerplateFile, err)
		}
		boilerplate = string(data)
	}

	configuredPackages, err := r.Config.GetPackages(ctx)
	if err != nil {
		return fmt.Errorf("failed to get package from config: %w", err)
	}
	if len(configuredPackages) == 0 {
		log.Error().Msg("no packages specified in config")
		return nil
	}
	parser := pkg.NewParser(buildTags)

	interfaces, err := parser.ParsePackages(ctx, configuredPackages)
	if err != nil {
		log.Error().Err(err).Msg("unable to parse packages")
		return err
	}
	mockFileToInterfaces := map[string][]*pkg.Interface{}

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
		ifaceLog.Debug().Msg("config specifies to generate this interface")
		ifaceConfigs, err := r.Config.GetInterfaceConfig(ctx, iface.Pkg.PkgPath, iface.Name)
		if err != nil {
			return err
		}
		for _, ifaceConfig := range ifaceConfigs {
			if interfaces, ok := mockFileToInterfaces[ifaceConfig.FilePath(ctx).String()]; !ok {
				interfaces = []*pkg.Interface{}
				mockFileToInterfaces[ifaceConfig.FilePath(ctx).String()] = interfaces
			}
			interfaces = append(interfaces, iface)
		}
	}

	return nil
}
