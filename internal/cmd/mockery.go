package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vektra/mockery/v3/config"
	"github.com/vektra/mockery/v3/internal"
	pkg "github.com/vektra/mockery/v3/internal"
	"github.com/vektra/mockery/v3/internal/file"
	"github.com/vektra/mockery/v3/internal/logging"
	"github.com/vektra/mockery/v3/internal/stackerr"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/tools/go/packages"
)

var ErrCfgFileNotFound = errors.New("config file not found")

func NewRootCmd() (*cobra.Command, error) {
	var pFlags *pflag.FlagSet
	cmd := &cobra.Command{
		Use:   "mockery",
		Short: "Generate mock objects for your Go interfaces",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pFlags.Parse(args); err != nil {
				fmt.Printf("failed to parse flags: %s", err.Error())
				os.Exit(1)
			}
			log, err := logging.GetLogger("info")
			if err != nil {
				fmt.Printf("failed to get logger: %s\n", err.Error())
				os.Exit(1)
			}
			ctx := log.WithContext(context.Background())

			r, err := GetRootApp(ctx, pFlags)
			if err != nil {
				logFatalErr(ctx, err)
			}
			if err := r.Run(); err != nil {
				logFatalErr(ctx, err)
			}
		},
	}
	pFlags = cmd.PersistentFlags()
	pFlags.String("config", "", "config file to use")
	pFlags.String("log-level", os.Getenv("MOCKERY_LOG_LEVEL"), "Level of logging")

	cmd.AddCommand(NewShowConfigCmd())
	cmd.AddCommand(NewVersionCmd())
	cmd.AddCommand(NewInitCmd())
	cmd.AddCommand(NewMigrateCmd())
	return cmd, nil
}

func logFatalErr(ctx context.Context, err error) {
	log := zerolog.Ctx(ctx)
	log.Fatal().Err(err).Msg("app failed")
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
	Config config.RootConfig
}

func GetRootApp(ctx context.Context, flags *pflag.FlagSet) (*RootApp, error) {
	r := &RootApp{}
	config, _, err := config.NewRootConfig(ctx, flags)
	if err != nil {
		return nil, fmt.Errorf("getting config: %w", err)
	}
	r.Config = *config
	return r, nil
}

// CollectionConfig contains the common, file-global configuration
// parameters for an InterfaceCollection. Some mockery config parameters
// are attributes of the entire file, rather than specific interfaces.
// For such attributes, we must assert that all interfaces within
// a collection have these values identically set.
type CollectionConfig struct {
	InPackage                   *bool
	OutFilePath                 string
	OutPkgName                  string
	SrcPkgPath                  string
	Template                    string
	TemplateSchema              string
	RequireTemplateSchemaExists bool
}

// InterfaceCollection maintains a list of *pkg.Interface and asserts that all
// the interfaces in the collection belong to the same source package. It also
// asserts that various properties of the interfaces added to the collection are
// uniform.
type InterfaceCollection struct {
	srcPkg     *packages.Package
	interfaces []*internal.Interface
	config     CollectionConfig
}

func NewInterfaceCollection(
	srcPkgPath string,
	outFilePath string,
	srcPkg *packages.Package,
	outPkgName string,
	template string,
	templateSchema string,
	inPackage *bool,
	requireTemplateSchemaExists bool,
) *InterfaceCollection {
	config := CollectionConfig{
		InPackage:                   inPackage,
		SrcPkgPath:                  srcPkgPath,
		OutFilePath:                 outFilePath,
		OutPkgName:                  outPkgName,
		Template:                    template,
		TemplateSchema:              templateSchema,
		RequireTemplateSchemaExists: requireTemplateSchemaExists,
	}
	return &InterfaceCollection{
		srcPkg:     srcPkg,
		interfaces: make([]*internal.Interface, 0),
		config:     config,
	}
}

func (i *InterfaceCollection) Append(ctx context.Context, iface *internal.Interface) error {
	collectionFilepath := i.config.OutFilePath
	interfaceFilepath := iface.Config.FilePath()
	log := zerolog.Ctx(ctx).With().
		Str(logging.LogKeyInterface, iface.Name).
		Str("collection-pkgname", i.config.OutPkgName).
		Str("interface-pkgname", *iface.Config.PkgName).
		Str("collection-pkgpath", i.config.SrcPkgPath).
		Str("interface-pkgpath", iface.Pkg.PkgPath).
		Str("collection-filepath", collectionFilepath).
		Str("interface-filepath", interfaceFilepath).
		Logger()

	if collectionFilepath != interfaceFilepath {
		msg := "all mocks in an InterfaceCollection must have the same output file path"
		log.Error().Msg(msg)
		return errors.New(msg)
	}
	if i.config.OutPkgName != *iface.Config.PkgName {
		msg := "all mocks in an output file must have the same pkgname"
		log.Error().Str("interface-pkgname", *iface.Config.PkgName).Msg(msg)
		return errors.New(msg)
	}
	if i.config.SrcPkgPath != iface.Pkg.PkgPath {
		msg := "all mocks in an output file must come from the same source package"
		log.Error().Msg(msg)
		return errors.New(msg)
	}
	if i.config.Template != *iface.Config.Template {
		msg := "all mocks in an output file must use the same template"
		log.Error().Str("expected-template", i.config.Template).Str("interface-template", *iface.Config.Template).Msg(msg)
		return errors.New(msg)
	}
	i.interfaces = append(i.interfaces, iface)
	return nil
}

func (r *RootApp) Run() error {
	remoteTemplateCache := make(map[string]*internal.RemoteTemplate)

	log, err := logging.GetLogger(*r.Config.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		return err
	}
	log.Info().Str("config-file", r.Config.ConfigFileUsed()).Msgf("Starting mockery")
	ctx := log.WithContext(context.Background())

	buildTags := strings.Split(*r.Config.BuildTags, " ")

	configuredPackages, err := r.Config.GetPackages(ctx)
	if err != nil {
		return fmt.Errorf("failed to get package from config: %w", err)
	}
	if len(configuredPackages) == 0 {
		log.Error().Msg("no packages specified in config")
		return fmt.Errorf("no packages specified in config")
	}
	parser := pkg.NewParser(buildTags, r.Config)

	// Let's build a missing map here to keep track of seen interfaces.
	// (pkg -> list of interface names)
	// After seeing an interface it'll be deleted from the map, keeping only
	// missing interfaces or packages in there.
	//
	// NOTE: We do that here without relying on parser, because parses iterates
	// over existing go files and interfaces, while user could've had a typo in
	// interface or pacakge name, making it impossible for parser to find these
	// files/interfaces in the first place.
	log.Debug().Msg("Making seen map...")
	missingMap := make(map[string]map[string]struct{}, len(configuredPackages))
	for _, p := range configuredPackages {
		config, err := r.Config.GetPackageConfig(ctx, p)
		if err != nil {
			return err
		}
		if _, ok := missingMap[p]; !ok {
			missingMap[p] = make(map[string]struct{}, len(config.Interfaces))
		}

		for ifaceName := range config.Interfaces {
			missingMap[p][ifaceName] = struct{}{}
		}
	}

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
	// to create the output mock) is because mockery allows multiple mocks to
	// be created for each input interface.
	mockFileToInterfaces := map[string]*InterfaceCollection{}

	for _, iface := range interfaces {
		ifaceLog := log.
			With().
			Str(logging.LogKeyInterface, iface.Name).
			Str(logging.LogKeyPackagePath, iface.Pkg.Types.Path()).
			Logger()

		if _, exist := missingMap[iface.Pkg.PkgPath]; exist {
			delete(missingMap[iface.Pkg.PkgPath], iface.Name)

			if len(missingMap[iface.Pkg.PkgPath]) == 0 {
				delete(missingMap, iface.Pkg.PkgPath)
			}
		}

		ifaceCtx := ifaceLog.WithContext(ctx)

		pkgConfig, err := r.Config.GetPackageConfig(ctx, iface.Pkg.PkgPath)
		if err != nil {
			return fmt.Errorf("getting package %s: %w", iface.Pkg.PkgPath, err)
		}
		ifaceLog.Debug().Str("root-mock-name", *r.Config.Config.StructName).Str("pkg-mock-name", *pkgConfig.Config.StructName).Msg("mock-name during first GetPackageConfig")

		// Extract configuration from directive comments in the interface documentation
		directiveConfig, err := config.ExtractDirectiveConfig(ifaceCtx, iface.GenDecl)
		if err != nil {
			return fmt.Errorf("extracting directive config for interface %s: %w", iface.Name, err)
		}

		ifaceConfig, err := pkgConfig.GetInterfaceConfig(ctx, iface.Name, directiveConfig)
		if err != nil {
			return fmt.Errorf("getting interface config for %s: %w", iface.Name, err)
		}

		shouldGenerate, err := pkgConfig.ShouldGenerateInterface(ifaceCtx, iface.Name, *ifaceConfig.Config, directiveConfig != nil)
		if err != nil {
			return err
		}
		if !shouldGenerate {
			ifaceLog.Debug().Msg("config doesn't specify to generate this interface, skipping")
			continue
		}

		for _, ifaceConfig := range ifaceConfig.Configs {
			if err := ifaceConfig.ParseTemplates(ifaceCtx, iface.FilePath, iface.Name, iface.Pkg); err != nil {
				log.Err(err).Msg("Can't parse config templates for interface")
				return err
			}
			filePath := ifaceConfig.FilePath()
			ifaceLog.Info().Str("collection", filePath).Msg("adding interface to collection")

			_, ok := mockFileToInterfaces[filePath]
			if !ok {
				mockFileToInterfaces[filePath] = NewInterfaceCollection(
					iface.Pkg.PkgPath,
					filePath,
					iface.Pkg,
					*ifaceConfig.PkgName,
					*ifaceConfig.Template,
					*ifaceConfig.TemplateSchema,
					ifaceConfig.InPackage,
					*ifaceConfig.RequireTemplateSchemaExists,
				)
				ifaceLog.Debug().Str("file-path", filePath).Msg("creating new interface collection")
			}
			ifaceWithTypes := internal.NewInterface(
				iface.Name,
				iface.TypeSpec,
				iface.GenDecl,
				iface.FilePath,
				iface.FileSyntax,
				iface.Pkg,
				ifaceConfig)
			ifaceLog.Debug().Str("ifaceWithTypes", fmt.Sprintf("%+v", ifaceWithTypes)).Msg("created iface with types")
			if err := mockFileToInterfaces[filePath].Append(
				ctx,
				ifaceWithTypes,
			); err != nil {
				return err
			}
		}
	}

	for outFilePath, collection := range mockFileToInterfaces {
		fileLog := log.With().Str("file", outFilePath).Logger()
		fileCtx := fileLog.WithContext(ctx)

		packageConfig, err := r.Config.GetPackageConfig(fileCtx, collection.config.SrcPkgPath)
		if err != nil {
			return err
		}

		interfacesInFileDir := filepath.ToSlash(filepath.Dir(collection.config.OutFilePath))
		generator, err := pkg.NewTemplateGenerator(
			fileCtx,
			collection.srcPkg,
			interfacesInFileDir,
			collection.config.Template,
			collection.config.TemplateSchema,
			collection.config.RequireTemplateSchemaExists,
			remoteTemplateCache,
			pkg.Formatter(*r.Config.Formatter),
			packageConfig.Config,
			collection.config.OutPkgName,
			collection.config.InPackage,
		)
		if err != nil {
			return err
		}
		fileLog.Info().Msg("Executing template")
		templateBytes, err := generator.Generate(fileCtx, collection.interfaces)
		if err != nil {
			return err
		}

		outFileDir := filepath.ToSlash(filepath.Dir(outFilePath))
		if outFileDir != "" {
			if err := os.MkdirAll(outFileDir, 0o755); err != nil {
				log.Err(err).Msg("failed to mkdir parent directories of mock file")
				return stackerr.NewStackErr(err)
			}
		}
		fileLog.Info().Msg("Writing template to file")
		exists, err := file.Exists(outFilePath)
		if err != nil {
			fileLog.Err(err).Msg("can't determine if outfile exists")
			return fmt.Errorf("determining if outfile exists: %w", err)
		}
		if exists && !*packageConfig.Config.ForceFileWrite {
			fileLog.Error().Bool("force-file-write", *packageConfig.Config.ForceFileWrite).Msg("output file exists, can't write mocks")
			return fmt.Errorf("outfile exists")
		}

		file, err := os.Create(outFilePath)
		if err != nil {
			return stackerr.NewStackErr(err)
		}
		defer file.Close()
		if _, err = file.Write(templateBytes); err != nil {
			return stackerr.NewStackErr(err)
		}
	}

	// The loop above could exit early, so sometimes warnings won't be shown
	// until other errors are fixed
	var foundMissing bool
	for packagePath := range missingMap {
		for ifaceName := range missingMap[packagePath] {
			foundMissing = true
			log.Error().
				Str(logging.LogKeyInterface, ifaceName).
				Str(logging.LogKeyPackagePath, packagePath).
				Str(logging.LogKeyDocsURL, logging.DocsURL("/faq/#error-interface-not-found-in-source")).
				Msg("interface not found in source. View the linked documentation for possible reasons why this might occur.")
		}
	}
	if foundMissing {
		os.Exit(1)
	}

	return nil
}
