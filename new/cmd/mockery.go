package cmd

import (
	"context"
	"fmt"
	"go/ast"
	"go/types"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/vektra/mockery/v2/pkg"

	"golang.org/x/tools/go/packages"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vektra/mockery/v2/pkg/config"
	"github.com/vektra/mockery/v2/pkg/logging"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	cfgFile = ""
)

func init() {
	cobra.OnInitialize(initConfig)
}

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mockery",
		Short: "Generate mock objects for your Golang interfaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := getRootAppFromViper(viper.GetViper())
			if err != nil {
				printStackTrace(err)
				return err
			}
			return r.Run()
		},
	}

	pFlags := cmd.PersistentFlags()
	pFlags.StringVar(&cfgFile, "config", "", "config file to use")
	pFlags.Bool("print", false, "print the generated mock to stdout")
	pFlags.String("case", "camel", "name the mocked file using casing convention [camel, snake, underscore]")
	pFlags.String("note", "", "comment to insert into prologue of each generated file")
	pFlags.String("cpuprofile", "", "write cpu profile to file")
	pFlags.Bool("version", false, "prints the installed version of mockery")
	pFlags.Bool("quiet", false, `suppresses logger output (equivalent to --log-level="")`)
	pFlags.String("tags", "", "space-separated list of additional build tags to use")
	pFlags.String("log-level", "info", "Level of logging")
	pFlags.BoolP("dry-run", "d", false, "Do a dry run, don't modify any files")
	pFlags.Bool("disable-version-string", false, "Do not insert the version string into the generated mock file.")
	pFlags.String("boilerplate-file", "", "File to read a boilerplate text from. Text should be a go block comment, i.e. /* ... */")
	pFlags.Bool("unroll-variadic", true, "For functions with variadic arguments, do not unroll the arguments into the underlying testify call. Instead, pass variadic slice as-is.")
	pFlags.Bool("exported", false, "Generates public mocks for private interfaces.")

	_ = viper.BindPFlags(pFlags)

	cmd.AddCommand(NewShowConfigCmd())
	return cmd
}

func printStackTrace(e error) {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	fmt.Printf("%v\n", e)
	if err, ok := e.(stackTracer); ok {
		for _, f := range err.StackTrace() {
			fmt.Printf("%+s:%d\n", f, f)
		}
	}

}

// Execute executes the cobra CLI workflow
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		//printStackTrace(err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.SetEnvPrefix("mockery")
	viper.AutomaticEnv()

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else if viper.IsSet("config") {
		viper.SetConfigFile(viper.GetString("config"))
	}
	// Note we purposely ignore the error. Don't care if we can't find a config file.
	if err := viper.ReadInConfig(); err == nil {
		_, _ = fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}
}

type rootApp struct {
	config.Config
}

func getRootAppFromViper(v *viper.Viper) (*rootApp, error) {
	r := &rootApp{}
	if err := v.UnmarshalExact(&r.Config); err != nil {
		return nil, errors.Wrapf(err, "failed to get config")
	}
	return r, nil
}

func (r *rootApp) Run() error {
	var err error

	if r.Quiet {
		// if "quiet" flag is set, disable logging
		r.Config.LogLevel = ""
	}

	log, err := getLogger(r.Config.LogLevel)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		return err
	}
	log = log.With().Bool(logging.LogKeyDryRun, r.Config.DryRun).Logger()
	log.Info().Msgf("Starting mockery")
	ctx := log.WithContext(context.Background())

	if r.Config.Version {
		fmt.Println(config.GetSemverInfo())
		return nil
	}
	if r.Config.Profile != "" {
		f, err := os.Create(r.Config.Profile)
		if err != nil {
			return errors.Wrapf(err, "failed to create profile file")
		}
		if err = pprof.StartCPUProfile(f); err != nil {
			return errors.Wrapf(err, "failed to start profiling")
		}
		defer pprof.StopCPUProfile()
	}

	var boilerplate string
	if r.Config.BoilerplateFile != "" {
		data, err := ioutil.ReadFile(r.Config.BoilerplateFile)
		if err != nil {
			log.Fatal().Msgf("Failed to read boilerplate file %s: %v", r.Config.BoilerplateFile, err)
		}
		boilerplate = string(data)
	}

	pkgs, err := r.loadPackages()
	if err != nil {
		return err
	}

	type packageDef struct {
		pkg        *packages.Package
		interfaces map[string]string
	}

	packagesByPath := make(map[string]packageDef)
	for _, p := range pkgs {
		if len(p.Errors) > 0 {
			var errs []string
			for _, e := range p.Errors {
				errs = append(errs, e.Error())
			}
			return fmt.Errorf(
				"encountered error(s) during loading package %s: %s",
				p.PkgPath,
				strings.Join(errs, "; "),
			)
		}

		packagesByPath[p.PkgPath] = packageDef{
			pkg:        p,
			interfaces: packageInterfaces(p),
		}
	}

	for _, mockDef := range r.Config.Mocks {
		if pkgDef, ok := packagesByPath[mockDef.Package]; ok {
			if sourceFile, found := pkgDef.interfaces[mockDef.Interface]; found {
				iface := locateInterface(pkgDef.pkg.Types, mockDef.Interface, sourceFile)
				if iface == nil {
					return fmt.Errorf("cannot locate interface %s in package %s", mockDef.Interface, mockDef.Package)
				}

				if err = r.generateMock(ctx, boilerplate, mockDef, iface); err != nil {
					return err
				}

			} else {
				return fmt.Errorf("package %s does not contain interface %s", mockDef.Package, mockDef.Interface)
			}

		} else {
			return fmt.Errorf("package %s was expected to load, but it did not", mockDef.Package)
		}
	}

	return nil
}

func (r *rootApp) loadPackages() ([]*packages.Package, error) {
	var packagePatterns []string
	for _, mockDef := range r.Config.Mocks {
		packagePatterns = append(packagePatterns, mockDef.Package)
	}

	cfg := packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedTypes | packages.NeedTypesSizes |
			packages.NeedSyntax | packages.NeedTypesInfo,
	}
	buildTags := strings.Split(r.Config.BuildTags, " ")
	if len(buildTags) > 0 {
		cfg.BuildFlags = []string{"-tags", strings.Join(buildTags, ",")}
	}

	return packages.Load(&cfg, packagePatterns...)
}

func packageInterfaces(pkg *packages.Package) map[string]string {
	nv := newNodeVisitor()
	for i, fileSyntax := range pkg.Syntax {
		nv.currentFile = pkg.CompiledGoFiles[i]
		ast.Walk(nv, fileSyntax)
	}
	return nv.declaredInterfaces
}

type nodeVisitor struct {
	currentFile        string
	declaredInterfaces map[string]string
}

func newNodeVisitor() *nodeVisitor {
	return &nodeVisitor{
		declaredInterfaces: make(map[string]string, 0),
	}
}

// Visit implements ast.Visitor
func (nv *nodeVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.TypeSpec:
		switch n.Type.(type) {
		case *ast.InterfaceType, *ast.FuncType:
			nv.declaredInterfaces[n.Name.Name] = nv.currentFile
		}
	}
	return nv
}

func locateInterface(p *types.Package, ifaceName string, sourceFile string) *pkg.Interface {
	scope := p.Scope()

	obj := scope.Lookup(ifaceName)
	if obj == nil {
		return nil
	}

	typ, ok := obj.Type().(*types.Named)
	if !ok {
		return nil
	}

	name := typ.Obj().Name()
	if typ.Obj().Pkg() == nil {
		return nil
	}

	result := &pkg.Interface{
		Name:          name,
		Pkg:           p,
		QualifiedName: p.Path(),
		NamedType:     typ,
		FileName:      sourceFile,
	}

	iface, ok := typ.Underlying().(*types.Interface)
	if ok {
		result.IsFunction = false
		result.ActualInterface = iface
	} else {
		sig, ok := typ.Underlying().(*types.Signature)
		if !ok {
			return nil
		}
		result.IsFunction = true
		result.SingleFunction = &pkg.Method{Name: "Execute", Signature: sig}
	}
	return result
}

func (r *rootApp) generateMock(ctx context.Context, boilerplate string, mockDef config.MockDef, iface *pkg.Interface) error {
	var osp pkg.OutputStreamProvider
	if r.Config.Print {
		osp = &pkg.StdoutStreamProvider{}
	} else {
		output := mockDef.Output
		if mockDef.InPackage {
			output = mockDef.Package
		} else if output == "" {
			output = path.Join(filepath.Dir(iface.FileName), mockDef.GeneratedPackageName())
		}

		osp = &pkg.FileOutputStreamProvider{
			Config:    r.Config,
			BaseDir:   output,
			InPackage: mockDef.InPackage,
			TestOnly:  mockDef.TestOnly,
			Case:      r.Config.Case,
			FileName:  mockDef.FileName,
		}
	}

	packageName := mockDef.Outpkg
	if packageName == "" {
		packageName = "mocks"
	}

	visitor := &pkg.GeneratorVisitor{
		Config:      r.Config,
		InPackage:   mockDef.InPackage,
		Note:        r.Config.Note,
		Boilerplate: boilerplate,
		Osp:         osp,
		PackageName: packageName,
		StructName:  mockDef.StructName,
	}

	return visitor.VisitWalk(ctx, iface)
}

type timeHook struct{}

// Run implements zerolog.Hook.Run.
func (t timeHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	e.Time("time", time.Now())
}

func getLogger(levelStr string) (zerolog.Logger, error) {
	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		return zerolog.Logger{}, errors.Wrapf(err, "Couldn't parse log level")
	}
	out := os.Stderr
	writer := zerolog.ConsoleWriter{
		Out:        out,
		TimeFormat: time.RFC822,
	}
	if !terminal.IsTerminal(int(out.Fd())) {
		writer.NoColor = true
	}
	log := zerolog.New(writer).
		Hook(timeHook{}).
		Level(level).
		With().
		Str("version", config.GetSemverInfo()).
		Logger()

	return log, nil
}
