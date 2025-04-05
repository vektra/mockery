package cmd

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/chigopher/pathlib"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/vektra/mockery/v3/config"
	internalConfig "github.com/vektra/mockery/v3/internal/config"
	"github.com/vektra/mockery/v3/internal/logging"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

func NewMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate v2 config to v3.",
		Long:  `This command automatically migrates a v2 config to v3.`,
		Run: func(cmd *cobra.Command, args []string) {
			logLevel, err := cmd.Flags().GetString("log-level")
			if err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}
			if logLevel == "" {
				logLevel = "info"
			}
			log, err := logging.GetLogger(logLevel)
			if err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}

			ctx := log.WithContext(context.Background())
			v2ConfPath, err := cmd.Flags().GetString("config")
			if err != nil {
				log.Err(err).Msg("failed to get parameter")
				os.Exit(1)
			}
			v3ConfigPath, err := cmd.Flags().GetString("outfile")
			if err != nil {
				log.Err(err).Msg("failed to get parameter")
				os.Exit(1)
			}

			if err := run(
				ctx,
				v2ConfPath,
				v3ConfigPath,
			); err != nil {
				log.Err(err).Msg("failed to run")
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}
		},
	}
	flags := cmd.PersistentFlags()
	flags.String("outfile", ".mockery_v3.yml", "Location of the ouptut v3 file.")

	return cmd
}

type tableWriter struct {
	seenMessages     map[string]any
	tbl              table.Writer
	idx              int
	termWidth        int
	messageWrapWidth int
}

func newTableWriter(ctx context.Context) *tableWriter {
	log := zerolog.Ctx(ctx)
	tbl := table.NewWriter()

	width, _, err := term.GetSize(int(os.Stdout.Fd())) //nolint:gosec // integer overflow warnings are inevitable because of argument types
	if err != nil {
		log.Warn().Err(err).Msg("failed to get terminal size")
	} else {
		tbl.SetOutputMirror(os.Stdout)
	}

	tbl.SetTitle("Deprecations")
	tbl.Style().Title.Align = text.AlignCenter
	tbl.Style().Box = table.StyleBoxRounded
	tbl.Style().Size.WidthMax = width
	tbl.Style().Options.SeparateRows = true
	tbl.Style().Options.SeparateColumns = false

	tbl.AppendHeader(
		table.Row{
			"Idx",
			"Deprecation Type",
			"Message",
		},
	)

	return &tableWriter{
		seenMessages:     map[string]any{},
		tbl:              tbl,
		idx:              0,
		termWidth:        width,
		messageWrapWidth: width - 35,
	}
}

func (t *tableWriter) Append(depType string, msg string) {
	if _, seen := t.seenMessages[msg]; seen {
		return
	}
	t.seenMessages[msg] = struct{}{}
	t.tbl.AppendRow(table.Row{
		fmt.Sprintf("%d", t.idx),
		depType,
		text.WrapSoft(msg, t.messageWrapWidth),
	})
	t.idx++
}

func (t *tableWriter) Render() {
	t.tbl.Render()
}

func run(ctx context.Context, confPathStr string, v3ConfPath string) error {
	var confPath *pathlib.Path
	var err error

	log := zerolog.Ctx(ctx)
	if confPathStr == "" {
		confPath, err = internalConfig.FindConfig()
		if err != nil {
			return fmt.Errorf("finding config: %w", err)
		}
	} else {
		confPath = pathlib.NewPath(confPathStr)
	}
	log.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Stringer("config", confPath)
	})
	log.Info().Msg("using config")

	var v2 V2RootConfig
	f, err := confPath.OpenFile(os.O_RDONLY)
	if err != nil {
		return fmt.Errorf("opening config file: %w", err)
	}
	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	if err := decoder.Decode(&v2); err != nil {
		log.Error().Msg("v2 config could not be decoded. Are you sure this is a v2 config file?")
		return fmt.Errorf("decoding v2 config: %w", err)
	}

	var v3 config.RootConfig
	v3.Config = &config.Config{}
	v3.TemplateData = map[string]any{}
	v3.Template = addr("testify")

	tbl := newTableWriter(ctx)

	migrateConfig(ctx, tbl, &v2.V2Config, &v3.Config)
	for pkgName, pkgConfig := range v2.Packages {
		pkgLog := log.With().Str("pkg-name", pkgName).Logger()
		pkgCtx := pkgLog.WithContext(ctx)

		v3PkgConfig := &config.PackageConfig{}
		if v3.Packages == nil {
			v3.Packages = map[string]*config.PackageConfig{}
		}
		v3.Packages[pkgName] = v3PkgConfig
		migrateConfig(pkgCtx, tbl, pkgConfig.Config, &v3PkgConfig.Config)

		for interfaceName, interfaceConfig := range pkgConfig.Interfaces {
			ifaceLog := pkgLog.With().Str("interface-name", interfaceName).Logger()
			ifaceCtx := ifaceLog.WithContext(pkgCtx)

			v3InterfaceConfig := config.InterfaceConfig{}
			if v3PkgConfig.Interfaces == nil {
				v3PkgConfig.Interfaces = map[string]*config.InterfaceConfig{}
			}
			v3PkgConfig.Interfaces[interfaceName] = &v3InterfaceConfig

			migrateConfig(ifaceCtx, tbl, interfaceConfig.Config, &v3InterfaceConfig.Config)

			for _, v2SubConfig := range interfaceConfig.Configs {
				v3SubConfig := &config.Config{}
				v3InterfaceConfig.Configs = append(v3InterfaceConfig.Configs, v3SubConfig)
				migrateConfig(ifaceCtx, tbl, &v2SubConfig, &v3SubConfig)
			}
		}
	}

	outFile := pathlib.NewPath(v3ConfPath)
	file, err := outFile.OpenFile(os.O_CREATE | os.O_RDWR | os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("opening .mockery_v3.yml: %w", err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()
	encoder.SetIndent(2)

	log.Info().Str("v3-config", outFile.String()).Msg("writing v3 config")
	if err := encoder.Encode(v3); err != nil {
		return fmt.Errorf("encoding .mockery_v3.yml: %w", err)
	}

	if len(tbl.seenMessages) != 0 {
		log.Warn().Msg("breaking changes detected that possibly require manual intervention. See table below.")
		tbl.Render()
	}
	return nil
}

func checkDeprecatedTemplateVariables(
	ctx context.Context,
	conf *V2Config,
	tbl *tableWriter,
) {
	log := zerolog.Ctx(ctx)
	confValue := reflect.ValueOf(conf).Elem()
	for i := range confValue.NumField() {
		fieldValue := confValue.Field(i)

		isPointerToString := fieldValue.Kind() == reflect.Pointer && fieldValue.Elem().Kind() == reflect.String
		isString := fieldValue.Kind() == reflect.String

		if !isPointerToString && !isString {
			log.Debug().Str("field", fieldValue.String()).Bool("pointerToString", isPointerToString).Bool("string", isString).Msg("field is not a pointer")
			continue
		}
		var fieldAsString string
		if isString {
			fieldAsString = fieldValue.Interface().(string)
		} else {
			fieldAsString = fieldValue.Elem().Interface().(string)
		}
		log.Debug().Str("field-as-string", fieldAsString).Str("field-name", confValue.Type().Field(i).Name).Msg("field as string")

		for _, deprecatedVariable := range []struct {
			name    string
			message string
		}{
			{
				name:    "InterfaceNameCamel",
				message: "InterfaceNameCamel template variable has been deleted. Use \"{{ .InterfaceName | camelcase }}\" instead",
			},
			{
				name:    "InterfaceNameLowerCamel",
				message: "InterfaceNameLowerCamel template variable has been deleted. Use \"{{ .InterfaceName | camelcase | firstLower }}\" instead",
			},
			{
				name:    "InterfaceNameSnake",
				message: "InterfaceNameSnake template variable has been deleted. Use \"{{ .InterfaceName | snakecase }}\" instead",
			},
			{
				name:    "InterfaceNameLower",
				message: "InterfaceNameLower template variable has been deleted. Use \"{{ .InterfaceName | lower }}\" instead",
			},
		} {
			if strings.Contains(fieldAsString, deprecatedVariable.name) {
				tbl.Append("template-variable", deprecatedVariable.message)
			}
		}
	}
}

func migrateConfig(
	ctx context.Context,
	tbl *tableWriter,
	v2Config *V2Config,
	v3Config **config.Config,
) {
	if v2Config == nil {
		return
	}
	checkDeprecatedTemplateVariables(ctx, v2Config, tbl)

	// We do this so we can lazily create a new `config` section if necessary.
	// It's kind of gross, but the double pointer is necessary to update the struct
	// that contains the *config.Config pointer.
	if *v3Config == nil {
		*v3Config = &config.Config{}
	}
	v3 := *v3Config
	v3.All = v2Config.All
	v3.Anchors = v2Config.Anchors
	if v2Config.BoilerplateFile != nil {
		if v3.TemplateData == nil {
			v3.TemplateData = map[string]any{}
		}
		v3.TemplateData["boilerplate-file"] = v2Config.BoilerplateFile
	}

	if v2Config.BuildTags != nil {
		tbl.Append("deprecated-parameter", "`tags` is no longer supported, parameter not migrated. Use `template-data.mock-build-tags` instead.")
	}
	if v2Config.Case != nil {
		tbl.Append("deprecated-parameter", "`case` is no longer supported. Use `structname` to specify the name and exported-ness of the output mocks.")
	}
	v3.ConfigFile = v2Config.Config
	if v2Config.Cpuprofile != nil {
		tbl.Append("deprecated-parameter", "`cpuprofile` is not supported in v3, however we welcome PRs to implement the feature: https://github.com/vektra/mockery/issues/956")
	}
	v3.Dir = v2Config.Dir
	if v2Config.DisableConfigSearch != nil {
		tbl.Append("deprecated-parameter", "`disable-config-search` is permanently disabled in v3.")
	}
	// disable-deprecation-warnings: no deprecations in v3
	// disabled-deprecation-warnings: no deprecations in v3
	if v2Config.DisableFuncMocks == nil || !*v2Config.DisableFuncMocks {
		tbl.Append("deprecated-parameter", "`disable-func-mocks` permanently enabled in v3.")
	}
	if v2Config.DisableVersionString == nil || !*v2Config.DisableVersionString {
		tbl.Append("deprecated-parameter", "`disable-version-string` is permanently set to True in v3.")
	}
	if v2Config.DryRun != nil && *v2Config.DryRun {
		tbl.Append("deprecated-parameter", "`dry-run` not supported in v3.")
	}
	v3.ExcludeSubpkgRegex = v2Config.Exclude
	v3.ExcludeRegex = v2Config.ExcludeRegex
	if v2Config.Exported != nil {
		tbl.Append("deprecated-parameter", "`exported` is no longer supported. Use `structname` instead.")
	}
	if v2Config.FailOnMissing == nil || (v2Config.FailOnMissing != nil && *v2Config.FailOnMissing == false) {
		tbl.Append("deprecated-parameter", "`fail-on-missing` is permanently set to True in v3.")
	}
	// inpackage: deleted, should work automatically.
	if v2Config.InPackageSuffix != nil {
		tbl.Append("deprecated-parameter", "`inpackage-suffix` is no longer supported in v3.")
	}
	if v2Config.IncludeAutoGenerated != nil && *v2Config.IncludeAutoGenerated == false {
		tbl.Append("deprecated-parameter", "`include-auto-generated` is not supported in v3, but PRs are welcome: https://github.com/vektra/mockery/issues/954")
	}
	v3.IncludeRegex = v2Config.IncludeRegex
	if v2Config.Issue845Fix == nil || *v2Config.Issue845Fix == false {
		tbl.Append("deprecated-parameter", "`issue-845-fix` is permanently set to True in v3.")
	}
	if v2Config.KeepTree != nil && *v2Config.KeepTree == true {
		tbl.Append("deprecated-parameter", "`keeptree` is not supported in v3. Use `dir` to specify where interfaces are located.")
	}
	v3.LogLevel = v2Config.LogLevel
	if v2Config.MockBuildTags != nil {
		if v3.TemplateData == nil {
			v3.TemplateData = map[string]any{}
		}
		v3.TemplateData["mock-build-tags"] = *v2Config.MockBuildTags
	}
	v3.StructName = v2Config.MockName
	if v2Config.Name != nil {
		tbl.Append("deprecated-parameter", "`name` is no longer supported. Use `structname` instead.")
	}
	if v2Config.Note != nil {
		tbl.Append("deprecated-parameter", "`note` is no longer supported.")
	}
	v3.PkgName = v2Config.Outpkg
	if v2Config.Output != nil {
		tbl.Append("deprecated-parameter", "`output` was replaced by `dir` in v2. This value is ignored.")
	}
	if v2Config.Packageprefix != nil {
		tbl.Append("deprecated-parameter", "`packageprefix` was replaced by `outpkg` in v2. This value is ignored.")
	}
	if v2Config.Print != nil && *v2Config.Print == true {
		tbl.Append("deprecated-parameter", "`print` is not supported in v3.")
	}
	if v2Config.Profile != nil {
		tbl.Append("deprecated-parameter", "`profile` is not supported in v3, but PRs are welcome to implement it: https://github.com/vektra/mockery/issues/955")
	}
	if v2Config.Quiet != nil && *v2Config.Quiet == true {
		tbl.Append("deprecated-parameter", "`quiet` is not supported in v3. Use `log-level` instead.")
	}
	v3.Recursive = v2Config.Recursive
	if len(v2Config.ReplaceType) != 0 {
		tbl.Append("deprecated-parameter", "`replace-type` has moved to a new schema. Cannot automatically migrate. Please visit https://vektra.github.io/mockery/latest-v3/replace-type/ for more information.")
	}
	if v2Config.ResolveTypeAlias != nil && *v2Config.ResolveTypeAlias == true {
		tbl.Append("deprecated-parameter", "`resolve-type-alias` is permanently set to False in v3. Type aliases typically should never be resolved.")
	}
	if v2Config.SrcPkg != nil {
		tbl.Append("deprecated-parameter", "`srcpkg` is not supported in v3. Use the `packages` configuration instead.")
	}
	if v2Config.StructName != nil {
		tbl.Append("deprecated-parameter", "`structname` was replaced by `structname` in v2. This value is ignored.")
	}
	if v2Config.TestOnly != nil {
		tbl.Append("deprecated-parameter", "`testonly` was replaced by `filename` in v2. This value is ignored and not supported in v3.")
	}
	if v2Config.UnrollVariadic != nil {
		if v3.TemplateData == nil {
			v3.TemplateData = map[string]any{}
		}
		v3.TemplateData["unroll-variadic"] = *v2Config.UnrollVariadic
	}
	if v2Config.WithExpecter != nil {
		if v3.TemplateData == nil {
			v3.TemplateData = map[string]any{}
		}
		v3.TemplateData["with-expecter"] = *v2Config.WithExpecter
	}
}

type V2RootConfig struct {
	V2Config `yaml:",inline"`
	Packages map[string]V2PackageConfig `yaml:"packages"`
}

type V2PackageConfig struct {
	Config     *V2Config                    `yaml:"config"`
	Interfaces map[string]V2InterfaceConfig `yaml:"interfaces"`
}

type V2InterfaceConfig struct {
	Config  *V2Config  `yaml:"config"`
	Configs []V2Config `yaml:"configs"`
}

type V2Config struct {
	All                         *bool          `yaml:"all"`
	Anchors                     map[string]any `yaml:"_anchors"`
	BoilerplateFile             *string        `yaml:"boilerplate-file"` // MOVED: moved to `template-data.boilerplate-file`
	BuildTags                   *string        `yaml:"tags"`             // DELETED: use mock-build-tags instead
	Case                        *string        `yaml:"case"`             // DELETED: caseness is specified using template variables/functions in `structname`.
	Config                      *string        `yaml:"config"`
	Cpuprofile                  *string        `yaml:"cpuprofile"` // DELETED: not an option in v3
	Dir                         *string        `yaml:"dir"`
	DisableConfigSearch         *bool          `yaml:"disable-config-search"` // DEPRECATED: permanently set to `false` in v3
	DisableDeprecationWarnings  *bool          `yaml:"disable-deprecation-warnings"`
	DisabledDeprecationWarnings *[]string      `yaml:"disabled-deprecation-warnings"`
	DisableFuncMocks            *bool          `yaml:"disable-func-mocks"`     // DEPRECATED: func-mocks no longer generated.
	DisableVersionString        *bool          `yaml:"disable-version-string"` // DEPRECATED: set to true in v3
	DryRun                      *bool          `yaml:"dry-run"`
	Exclude                     []string       `yaml:"exclude"` // MOVED: moved to `exclude-subpkg-regex` in v3
	ExcludeRegex                *string        `yaml:"exclude-regex"`
	Exported                    *bool          `yaml:"exported"`        // DELETED: Use templated parameters to define upper/lower case-ness of mock names.
	FailOnMissing               *bool          `yaml:"fail-on-missing"` // DEPRECATED: set to true permanently in v3
	FileName                    *string        `yaml:"filename"`
	InPackage                   *bool          `yaml:"inpackage"`              // DELETED: mockery automatically detects the appropriate value for the parameter.
	InPackageSuffix             *bool          `yaml:"inpackage-suffix"`       // DELETED: Use `packages` config.
	IncludeAutoGenerated        *bool          `yaml:"include-auto-generated"` // DELETED: not supported in v3. PRs to port functionality to v3 welcome.
	IncludeRegex                *string        `yaml:"include-regex"`
	Issue845Fix                 *bool          `yaml:"issue-845-fix"` // DEPRECATED: set to true in v3
	KeepTree                    *bool          `yaml:"keeptree"`      // DELETED: mockery uses templated parameters to specify directory and filename locations.
	LogLevel                    *string        `yaml:"log-level"`
	MockBuildTags               *string        `yaml:"mock-build-tags"` // MOVED: moved to `template-data.mock-build-tags`
	MockName                    *string        `yaml:"mockname"`
	Name                        *string        `yaml:"name"`          // DELETED: not supported
	Note                        *string        `yaml:"note"`          // DELETED: not supported
	Outpkg                      *string        `yaml:"outpkg"`        // DEPRECATED: Use `pkgname` instead
	Output                      *string        `yaml:"output"`        // DELETED: Use `packages` config
	Packageprefix               *string        `yaml:"packageprefix"` // DEPRECATED: use `pkgname`
	Print                       *bool          `yaml:"print"`         // DEPRECATED: printing mocks not an option
	Profile                     *string        `yaml:"profile"`       // DELETED: not an option in v3
	Quiet                       *bool          `yaml:"quiet"`         // DEPRECATED: deleted in v3 in favor of log-level
	Recursive                   *bool          `yaml:"recursive"`
	ReplaceType                 []string       `yaml:"replace-type"`       // DEPRECATED: moved to new schema in v3
	ResolveTypeAlias            *bool          `yaml:"resolve-type-alias"` // DEPRECATED: permanently set to false in v3
	SrcPkg                      *string        `yaml:"srcpkg"`             // DELETED: Use `packages` config.
	StructName                  *string        `yaml:"structname"`         // MOVED: moved to `structname` in v3
	TestOnly                    *bool          `yaml:"testonly"`           // DEPRECATED: use `filename` to generate `_test.go` suffix.
	UnrollVariadic              *bool          `yaml:"unroll-variadic"`    // MOVED: moved to `template-data.unroll-variadic`
	Version                     *bool          `yaml:"version"`
	WithExpecter                *bool          `yaml:"with-expecter"` // DEPRECATED: set to true in v3
}
