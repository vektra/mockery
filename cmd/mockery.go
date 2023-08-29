package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime/pprof"
	"strings"

	"github.com/chigopher/pathlib"
	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vektra/mockery/v2/pkg"
	"github.com/vektra/mockery/v2/pkg/config"
	"github.com/vektra/mockery/v2/pkg/logging"
	"github.com/vektra/mockery/v2/pkg/stackerr"
	"golang.org/x/tools/go/packages"
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
	pFlags.String("name", "", "name or matching regular expression of interface to generate mock for")
	pFlags.Bool("print", false, "print the generated mock to stdout")
	pFlags.String("output", "", "directory to write mocks to")
	pFlags.String("outpkg", "mocks", "name of generated package")
	pFlags.String("packageprefix", "", "prefix for the generated package name, it is ignored if outpkg is also specified.")
	pFlags.String("dir", "", "directory to search for interfaces")
	pFlags.BoolP("recursive", "r", false, "recurse search into sub-directories")
	pFlags.StringArray("exclude", nil, "prefixes of subdirectories and files to exclude from search")
	pFlags.Bool("all", false, "generates mocks for all found interfaces in all sub-directories")
	pFlags.Bool("inpackage", false, "generate a mock that goes inside the original package")
	pFlags.Bool("inpackage-suffix", false, "use filename '_mock' suffix instead of 'mock_' prefix for InPackage mocks")
	pFlags.Bool("testonly", false, "generate a mock in a _test.go file")
	pFlags.String("case", "", "name the mocked file using casing convention [camel, snake, underscore]")
	pFlags.String("note", "", "comment to insert into prologue of each generated file")
	pFlags.String("cpuprofile", "", "write cpu profile to file")
	pFlags.Bool("version", false, "prints the installed version of mockery")
	pFlags.Bool("quiet", false, `suppresses logger output (equivalent to --log-level="")`)
	pFlags.Bool("keeptree", false, "keep the tree structure of the original interface files into a different repository. Must be used with XX")
	pFlags.String("tags", "", "space-separated list of additional build tags to use")
	pFlags.String("filename", "", "name of generated file (only works with -name and no regex)")
	pFlags.String("structname", "", "name of generated struct (only works with -name and no regex)")
	pFlags.String("log-level", "info", "Level of logging")
	pFlags.String("srcpkg", "", "source pkg to search for interfaces")
	pFlags.BoolP("dry-run", "d", false, "Do a dry run, don't modify any files")
	pFlags.Bool("disable-version-string", false, "Do not insert the version string into the generated mock file.")
	pFlags.String("boilerplate-file", "", "File to read a boilerplate text from. Text should be a go block comment, i.e. /* ... */")
	pFlags.Bool("unroll-variadic", true, "For functions with variadic arguments, do not unroll the arguments into the underlying testify call. Instead, pass variadic slice as-is.")
	pFlags.Bool("exported", false, "Generates public mocks for private interfaces.")
	pFlags.Bool("with-expecter", false, "Generate expecter utility around mock's On, Run and Return methods with explicit types. This option is NOT compatible with -unroll-variadic=false")
	pFlags.StringArray("replace-type", nil, "Replace types")

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
	config.Config
}

func GetRootAppFromViper(v *viper.Viper) (*RootApp, error) {
	r := &RootApp{}
	config, err := config.NewConfigFromViper(v)
	if err != nil {
		return nil, stackerr.NewStackErrf(err, "failed to get config")
	}
	r.Config = *config
	return r, nil
}

func (r *RootApp) Run() error {
	var recursive bool
	var filter *regexp.Regexp
	var err error
	var limitOne bool

	if r.Quiet {
		// if "quiet" flag is set, disable logging
		r.Config.LogLevel = ""
	}

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

	var osp pkg.OutputStreamProvider
	if r.Config.Print {
		osp = &pkg.StdoutStreamProvider{}
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
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to determine configured packages: %w", err)
	}
	if len(configuredPackages) != 0 {
		r.Config.LogUnsupportedPackagesConfig(ctx)

		configuredPackages, err := r.Config.GetPackages(ctx)
		if err != nil {
			return fmt.Errorf("failed to get package from config: %w", err)
		}
		parser := pkg.NewParser(buildTags)

		if err := parser.ParsePackages(ctx, configuredPackages); err != nil {
			log.Error().Err(err).Msg("unable to parse packages")
			return err
		}
		log.Info().Msg("done parsing, loading")
		if err := parser.Load(); err != nil {
			log.Err(err).Msgf("failed to load parser")
			return nil
		}
		log.Info().Msg("done loading, visiting interface nodes")
		for _, iface := range parser.Interfaces() {
			ifaceLog := log.
				With().
				Str(logging.LogKeyInterface, iface.Name).
				Str(logging.LogKeyQualifiedName, iface.QualifiedName).
				Logger()

			ifaceCtx := ifaceLog.WithContext(ctx)

			shouldGenerate, err := r.Config.ShouldGenerateInterface(ifaceCtx, iface.QualifiedName, iface.Name)
			if err != nil {
				return err
			}
			if !shouldGenerate {
				ifaceLog.Debug().Msg("config doesn't specify to generate this interface, skipping.")
				continue
			}
			ifaceLog.Debug().Msg("config specifies to generate this interface")

			outputter := pkg.NewOutputter(&r.Config, boilerplate, true)
			if err := outputter.Generate(ifaceCtx, iface); err != nil {
				return err
			}
		}

		return nil
	}

	if r.Config.Name != "" && r.Config.All {
		log.Fatal().Msgf("Specify --name or --all, but not both")
	} else if (r.Config.FileName != "" || r.Config.StructName != "") && r.Config.All {
		log.Fatal().Msgf("Cannot specify --filename or --structname with --all")
	} else if r.Config.Dir != "" && r.Config.Dir != "." && r.Config.SrcPkg != "" {
		log.Fatal().Msgf("Specify --dir or --srcpkg, but not both")
	} else if r.Config.Name != "" {
		recursive = r.Config.Recursive
		if strings.ContainsAny(r.Config.Name, regexMetadataChars) {
			if filter, err = regexp.Compile(r.Config.Name); err != nil {
				log.Fatal().Err(err).Msgf("Invalid regular expression provided to -name")
			} else if r.Config.FileName != "" || r.Config.StructName != "" {
				log.Fatal().Msgf("Cannot specify --filename or --structname with regex in --name")
			}
		} else {
			filter = regexp.MustCompile(fmt.Sprintf("^%s$", r.Config.Name))
			limitOne = true
		}
	} else if r.Config.All {
		recursive = true
		filter = regexp.MustCompile(".*")
	} else {
		log.Fatal().Msgf("Use --name to specify the name of the interface or --all for all interfaces found")
	}

	warnDeprecated(
		ctx,
		"use of the packages config will be the only way to generate mocks in v3. Please migrate your config to use the packages feature.",
		map[string]any{
			"url":       logging.DocsURL("/features/#packages-configuration"),
			"migration": logging.DocsURL("/migrating_to_packages/"),
		})

	if r.Config.Profile != "" {
		f, err := os.Create(r.Config.Profile)
		if err != nil {
			return stackerr.NewStackErrf(err, "Failed to create profile file")
		}

		if err := pprof.StartCPUProfile(f); err != nil {
			return fmt.Errorf("failed to start CPU profile: %w", err)
		}
		defer pprof.StopCPUProfile()
	}

	baseDir := r.Config.Dir

	if osp == nil {
		osp = &pkg.FileOutputStreamProvider{
			Config:                    r.Config,
			BaseDir:                   r.Config.Output,
			InPackage:                 r.Config.InPackage,
			InPackageSuffix:           r.Config.InPackageSuffix,
			TestOnly:                  r.Config.TestOnly,
			Case:                      r.Config.Case,
			KeepTree:                  r.Config.KeepTree,
			KeepTreeOriginalDirectory: r.Config.Dir,
			FileName:                  r.Config.FileName,
		}
	}

	if r.Config.SrcPkg != "" {
		pkgs, err := packages.Load(&packages.Config{
			Mode: packages.NeedFiles,
		}, r.Config.SrcPkg)
		if err != nil || len(pkgs) == 0 {
			log.Fatal().Err(err).Msgf("Failed to load package %s", r.Config.SrcPkg)
		}

		// NOTE: we only pass one package name (config.SrcPkg) to packages.Load
		// it should return one package at most
		pkg := pkgs[0]

		if pkg.Errors != nil {
			log.Fatal().Err(pkg.Errors[0]).Msgf("Failed to load package %s", r.Config.SrcPkg)
		}

		if len(pkg.GoFiles) == 0 {
			log.Fatal().Msgf("No go files in package %s", r.Config.SrcPkg)
		}
		baseDir = filepath.Dir(pkg.GoFiles[0])
	}

	walker := pkg.Walker{
		Config:    r.Config,
		BaseDir:   baseDir,
		Recursive: recursive,
		Filter:    filter,
		LimitOne:  limitOne,
		BuildTags: buildTags,
	}

	visitor := pkg.NewGeneratorVisitor(pkg.GeneratorVisitorConfig{
		Boilerplate:          boilerplate,
		DisableVersionString: r.Config.DisableVersionString,
		Exported:             r.Config.Exported,
		InPackage:            r.Config.InPackage,
		KeepTree:             r.Config.KeepTree,
		Note:                 r.Config.Note,
		PackageName:          r.Config.Outpkg,
		PackageNamePrefix:    r.Config.Packageprefix,
		StructName:           r.Config.StructName,
		UnrollVariadic:       r.Config.UnrollVariadic,
		WithExpecter:         r.Config.WithExpecter,
		ReplaceType:          r.Config.ReplaceType,
	}, osp, r.Config.DryRun)

	generated := walker.Walk(ctx, visitor)

	if r.Config.Name != "" && !generated {
		log.Error().Msgf("Unable to find '%s' in any go files under this path", r.Config.Name)
		return fmt.Errorf("unable to find interface")
	}

	return nil
}

func warn(ctx context.Context, prefix string, message string, fields map[string]any) {
	log := zerolog.Ctx(ctx)
	event := log.Warn()
	if fields != nil {
		event = event.Fields(fields)
	}
	event.Msgf("%s: %s", prefix, message)
}

func info(ctx context.Context, prefix string, message string, fields map[string]any) {
	log := zerolog.Ctx(ctx)
	event := log.Info()
	if fields != nil {
		event = event.Fields(fields)
	}
	event.Msgf("%s: %s", prefix, message)
}

func warnDeprecated(ctx context.Context, message string, fields map[string]any) {
	warn(ctx, "DEPRECATION", message, fields)
}
