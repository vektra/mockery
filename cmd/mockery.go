package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vektra/mockery/pkg/mockery"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/tools/go/packages"
)

var (
	cfgFile = ""
	rootCmd = &cobra.Command{
		Use:   "mockery",
		Short: "Generate mock objects for your Golang interfaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := GetRootAppFromViper(viper.GetViper())
			if err != nil {
				printStackTrace(err)
				return nil
			}
			return r.Run()
		},
	}
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func printStackTrace(e error) {
	fmt.Printf("%v\n", e)
	if err, ok := e.(stackTracer); ok {
		for _, f := range err.StackTrace() {
			fmt.Printf("%+s:%d\n", f, f)
		}
	}

}

// Execute executes the cobra CLI workflow
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		//printStackTrace(err)
		os.Exit(1)
	}
}

type Config struct {
	All        bool
	BuildTags  string
	Case       string
	Config     string
	Cpuprofile string
	Dir        string
	FileName   string
	InPackage  bool
	KeepTree   bool
	LogLevel   string `mapstructure:"log-level"`
	Name       string
	Note       string
	Outpkg     string
	Output     string
	Print      bool
	Profile    string
	Quiet      bool
	Recursive  bool
	SrcPkg     string
	StructName string
	Tags       string
	TestOnly   bool
	Version    bool
}

func init() {
	cobra.OnInitialize(initConfig)

	pFlags := rootCmd.PersistentFlags()
	pFlags.StringVar(&cfgFile, "config", "", "config file to use")
	pFlags.String("name", "", "name or matching regular expression of interface to generate mock for")
	pFlags.Bool("print", false, "print the generated mock to stdout")
	pFlags.String("output", "./mocks", "directory to write mocks to")
	pFlags.String("outpkg", "mocks", "name of generated package")
	pFlags.String("dir", ".", "directory to search for interfaces")
	pFlags.BoolP("recursive", "r", false, "recurse search into sub-directories")
	pFlags.Bool("all", false, "generates mocks for all found interfaces in all sub-directories")
	pFlags.Bool("inpackage", false, "generate a mock that goes inside the original package")
	pFlags.Bool("testonly", false, "generate a mock in a _test.go file")
	pFlags.String("case", "camel", "name the mocked file using casing convention [camel, snake, underscore]")
	pFlags.String("note", "", "comment to insert into prologue of each generated file")
	pFlags.String("cpuprofile", "", "write cpu profile to file")
	pFlags.Bool("version", false, "prints the installed version of mockery")
	pFlags.Bool("quiet", false, "suppress output to stdout")
	pFlags.Bool("keeptree", false, "keep the tree structure of the original interface files into a different repository. Must be used with XX")
	pFlags.String("tags", "", "space-separated list of additional build tags to use")
	pFlags.String("filename", "", "name of generated file (only works with -name and no regex)")
	pFlags.String("structname", "", "name of generated struct (only works with -name and no regex)")
	pFlags.String("log-level", "info", "Level of logging")
	pFlags.String("srcpkg", "", "source pkg to search for interfaces")

	viper.BindPFlags(pFlags)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to find homedir")
		}

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigName(".mockery")
	}

	viper.SetEnvPrefix("mockery")
	viper.AutomaticEnv()

	// Note we purposely ignore the error. Don't care if we can't find a config file.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}
}

const regexMetadataChars = "\\.+*?()|[]{}^$"

type RootApp struct {
	Config
}

func GetRootAppFromViper(v *viper.Viper) (*RootApp, error) {
	r := &RootApp{}
	if err := v.UnmarshalExact(&r.Config); err != nil {
		return nil, errors.Wrapf(err, "failed to get config")
	}
	return r, nil
}

func (r *RootApp) Run() error {
	var recursive bool
	var filter *regexp.Regexp
	var err error
	var limitOne bool

	if r.Quiet {
		// if "quiet" flag is set, set os.Stdout to /dev/null to suppress all output to Stdout
		os.Stdout = os.NewFile(uintptr(syscall.Stdout), os.DevNull)
		r.Config.LogLevel = ""
	}

	log, err := getLogger(r.Config.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		return err
	}
	log.Info().Msgf("Starting mockery")
	ctx := log.WithContext(context.Background())

	if r.Config.Version {
		fmt.Println(mockery.SemVer)
		return nil
	} else if r.Config.Name != "" && r.Config.All {
		log.Fatal().Msgf("Specify --name or --all, but not both")
	} else if (r.Config.FileName != "" || r.Config.StructName != "") && r.Config.All {
		log.Fatal().Msgf("Cannot specify --filename or --structname with --all")
	} else if r.Config.Dir != "" && r.Config.Dir != "." && r.Config.SrcPkg != "" {
		log.Fatal().Msgf("Specify -dir or -srcpkg, but not both")
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
	if r.Config.KeepTree && r.Config.InPackage {
		log.Fatal().Msgf("--keeptree and --inpackage are mutually exclusive")
	}

	if r.Config.Profile != "" {
		f, err := os.Create(r.Config.Profile)
		if err != nil {
			return errors.Wrapf(err, "Failed to create profile file")
		}

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	var osp mockery.OutputStreamProvider
	if r.Config.Print {
		osp = &mockery.StdoutStreamProvider{}
	} else {
		osp = &mockery.FileOutputStreamProvider{
			BaseDir:                   r.Config.Output,
			InPackage:                 r.Config.InPackage,
			TestOnly:                  r.Config.TestOnly,
			Case:                      r.Config.Case,
			KeepTree:                  r.Config.KeepTree,
			KeepTreeOriginalDirectory: r.Config.Dir,
			FileName:                  r.Config.FileName,
		}
	}

	baseDir := r.Config.Dir

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

	visitor := &mockery.GeneratorVisitor{
		InPackage:   r.Config.InPackage,
		Note:        r.Config.Note,
		Osp:         osp,
		PackageName: r.Config.Outpkg,
		StructName:  r.Config.StructName,
	}

	walker := mockery.Walker{
		BaseDir:   baseDir,
		Recursive: recursive,
		Filter:    filter,
		LimitOne:  limitOne,
		BuildTags: strings.Split(r.Config.BuildTags, " "),
	}

	generated := walker.Walk(ctx, visitor)

	if r.Config.Name != "" && !generated {
		log.Fatal().Msgf("Unable to find '%s' in any go files under this path", r.Config.Name)
	}
	return nil
}

type timeHook struct{}

func (t timeHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
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
		Str("version", mockery.SemVer).
		Logger()

	return log, nil
}
