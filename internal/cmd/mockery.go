package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/chigopher/pathlib"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	pkg "github.com/vektra/mockery/v3/internal"
	"github.com/vektra/mockery/v3/internal/logging"
	"github.com/vektra/mockery/v3/internal/stackerr"
	"github.com/vektra/mockery/v3/template"
	"golang.org/x/tools/go/packages"
)

var ErrCfgFileNotFound = errors.New("config file not found")

func NewRootCmd() (*cobra.Command, error) {
	var pFlags *pflag.FlagSet
	cmd := &cobra.Command{
		Use:   "mockery",
		Short: "Generate mock objects for your Golang interfaces",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pFlags.Parse(args); err != nil {
				fmt.Printf("failed to parse flags: %s", err.Error())
				os.Exit(1)
			}
			level, err := pFlags.GetString("log-level")
			if err != nil {
				fmt.Printf("failed to get log-level from flags: %s\n", err.Error())
				os.Exit(1)
			}
			log, err := logging.GetLogger(level)
			if err != nil {
				fmt.Printf("failed to get logger: %s\n", err.Error())
				os.Exit(1)
			}
			ctx := log.WithContext(context.Background())

			r, err := GetRootApp(ctx, pFlags)
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
	pFlags = cmd.PersistentFlags()
	pFlags.String("config", "", "config file to use")
	pFlags.String("tags", "", "space-separated list of additional build tags to load packages")
	pFlags.String("mock-build-tags", "", "set the build tags of the generated mocks. Read more about the format: https://pkg.go.dev/cmd/go#hdr-Build_constraints")
	pFlags.String("log-level", os.Getenv("MOCKERY_LOG_LEVEL"), "Level of logging")
	pFlags.String("boilerplate-file", "", "File to read a boilerplate text from. Text should be a go block comment, i.e. /* ... */")
	pFlags.Bool("unroll-variadic", true, "For functions with variadic arguments, do not unroll the arguments into the underlying testify call. Instead, pass variadic slice as-is.")

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
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type RootApp struct {
	Config template.RootConfig
}

func GetRootApp(ctx context.Context, flags *pflag.FlagSet) (*RootApp, error) {
	r := &RootApp{}
	config, _, err := template.NewRootConfig(ctx, flags)
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
	outPkgName  string
	interfaces  []*template.Interface
	template    string
}

func NewInterfaceCollection(
	srcPkgPath string,
	outFilePath *pathlib.Path,
	srcPkg *packages.Package,
	outPkgName string,
	templ string,
) *InterfaceCollection {
	return &InterfaceCollection{
		srcPkgPath:  srcPkgPath,
		outFilePath: outFilePath,
		srcPkg:      srcPkg,
		outPkgName:  outPkgName,
		interfaces:  make([]*template.Interface, 0),
		template:    templ,
	}
}

func (i *InterfaceCollection) Append(ctx context.Context, iface *template.Interface) error {
	collectionFilepath := i.outFilePath.String()
	interfaceFilepath := iface.Config.FilePath().String()
	log := zerolog.Ctx(ctx).With().
		Str(logging.LogKeyInterface, iface.Name).
		Str("collection-pkgname", i.outPkgName).
		Str("interface-pkgname", *iface.Config.PkgName).
		Str("collection-pkgpath", i.srcPkgPath).
		Str("interface-pkgpath", iface.Pkg.PkgPath).
		Str("collection-filepath", collectionFilepath).
		Str("interface-filepath", interfaceFilepath).
		Logger()

	if collectionFilepath != interfaceFilepath {
		msg := "all mocks in an InterfaceCollection must have the same output file path"
		log.Error().Msg(msg)
		return errors.New(msg)
	}
	if i.outPkgName != *iface.Config.PkgName {
		msg := "all mocks in an output file must have the same pkgname"
		log.Error().Str("interface-pkgname", *iface.Config.PkgName).Msg(msg)
		return errors.New(msg)
	}
	if i.srcPkgPath != iface.Pkg.PkgPath {
		msg := "all mocks in an output file must come from the same source package"
		log.Error().Msg(msg)
		return errors.New(msg)
	}
	if i.template != *iface.Config.Template {
		msg := "all mocks in an output file must use the same template"
		log.Error().Str("expected-template", i.template).Str("interface-template", *iface.Config.Template).Msg(msg)
		return errors.New(msg)
	}
	i.interfaces = append(i.interfaces, iface)
	return nil
}

func (r *RootApp) Run() error {
	log, err := logging.GetLogger(*r.Config.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		return err
	}
	log.Info().Str("config-file", r.Config.ConfigFileUsed().String()).Msgf("Starting mockery")
	ctx := log.WithContext(context.Background())

	if err := r.Config.Initialize(ctx); err != nil {
		return err
	}

	buildTags := strings.Split(*r.Config.BuildTags, " ")

	configuredPackages, err := r.Config.GetPackages(ctx)
	if err != nil {
		return fmt.Errorf("failed to get package from config: %w", err)
	}
	if len(configuredPackages) == 0 {
		log.Error().Msg("no packages specified in config")
		return fmt.Errorf("no packages specified in config")
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
			Str(logging.LogKeyPackagePath, iface.Pkg.Types.Path()).
			Logger()

		ifaceCtx := ifaceLog.WithContext(ctx)

		pkgConfig, err := r.Config.GetPackageConfig(ctx, iface.Pkg.PkgPath)
		if err != nil {
			return fmt.Errorf("getting package %s: %w", iface.Pkg.PkgPath, err)
		}
		ifaceLog.Debug().Str("root-mock-name", *r.Config.Config.MockName).Str("pkg-mock-name", *pkgConfig.Config.MockName).Msg("mock-name during first GetPackageConfig")

		shouldGenerate, err := pkgConfig.ShouldGenerateInterface(ifaceCtx, iface.Name)
		if err != nil {
			return err
		}
		if !shouldGenerate {
			ifaceLog.Debug().Msg("config doesn't specify to generate this interface, skipping")
			continue
		}
		if pkgConfig.Interfaces == nil {
			ifaceLog.Debug().Msg("interfaces is nil")
		}
		ifaceConfig := pkgConfig.GetInterfaceConfig(ctx, iface.Name)
		for _, ifaceConfig := range ifaceConfig.Configs {
			if err := ifaceConfig.ParseTemplates(ifaceCtx, iface, iface.Pkg); err != nil {
				log.Err(err).Msg("Can't parse config templates for interface")
				return err
			}
			filePath := ifaceConfig.FilePath().Clean()
			ifaceLog.Info().Str("collection", filePath.String()).Msg("adding interface to collection")

			_, ok := mockFileToInterfaces[filePath.String()]
			if !ok {
				mockFileToInterfaces[filePath.String()] = NewInterfaceCollection(
					iface.Pkg.PkgPath,
					filePath,
					iface.Pkg,
					*ifaceConfig.PkgName,
					*ifaceConfig.Template,
				)
			}
			if err := mockFileToInterfaces[filePath.String()].Append(
				ctx,
				template.NewInterface(
					iface.Name,
					iface.FileName,
					iface.File,
					iface.Pkg,
					ifaceConfig),
			); err != nil {
				return err
			}
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
		if err := packageConfig.Config.ParseTemplates(ctx, nil, interfacesInFile.srcPkg); err != nil {
			return err
		}

		generator, err := pkg.NewTemplateGenerator(
			fileCtx,
			interfacesInFile.srcPkg,
			interfacesInFile.outFilePath.Parent(),
			*packageConfig.Config.Template,
			pkg.Formatter(*r.Config.Formatter),
			packageConfig.Config,
			interfacesInFile.outPkgName,
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
		outFileExists, err := outFile.Exists()
		if err != nil {
			fileLog.Err(err).Msg("can't determine if outfile exists")
			return fmt.Errorf("determining if outfile exists: %w", err)
		}
		if outFileExists && !packageConfig.Config.ForceFileWrite {
			fileLog.Error().Bool("force-file-write", packageConfig.Config.ForceFileWrite).Msg("output file exists, can't write mocks")
			return fmt.Errorf("outfile exists")
		}

		if err := outFile.WriteFile(templateBytes); err != nil {
			return stackerr.NewStackErr(err)
		}
	}

	return nil
}
