package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/vektra/mockery/mockery"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/tools/go/packages"
)

const regexMetadataChars = "\\.+*?()|[]{}^$"

type Config struct {
	fName          string
	fPrint         bool
	fOutput        string
	fOutpkg        string
	fPackagePrefix string
	fDir           string
	fRecursive     bool
	fAll           bool
	fIP            bool
	fTO            bool
	fCase          string
	fNote          string
	fProfile       string
	fVersion       bool
	fSrcPkg        string
	quiet          bool
	fkeepTree      bool
	buildTags      string
	fFileName      string
	fStructName    string
	fLogLevel      string
}

func main() {
	config := parseConfigFromArgs(os.Args)

	var recursive bool
	var filter *regexp.Regexp
	var err error
	var limitOne bool

	if config.quiet {
		// if "quiet" flag is set, set os.Stdout to /dev/null to suppress all output to Stdout
		os.Stdout = os.NewFile(uintptr(syscall.Stdout), os.DevNull)
		config.fLogLevel = ""
	}

	log, err := getLogger(config.fLogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	log.Info().Msgf("Starting mockery")
	ctx := log.WithContext(context.Background())

	if config.fVersion {
		fmt.Println(mockery.SemVer)
		return
	} else if config.fName != "" && config.fAll {
		log.Fatal().Msgf("Specify -name or -all, but not both")
	} else if (config.fFileName != "" || config.fStructName != "") && config.fAll {
		log.Fatal().Msgf("Cannot specify -filename or -structname with -all")
	} else if config.fDir != "" && config.fDir != "." && config.fSrcPkg != "" {
		log.Fatal().Msgf("Specify -dir or -srcpkg, but not both")
	} else if config.fName != "" {
		recursive = config.fRecursive
		if strings.ContainsAny(config.fName, regexMetadataChars) {
			if filter, err = regexp.Compile(config.fName); err != nil {
				log.Fatal().Err(err).Msgf("Invalid regular expression provided to -name")
			} else if config.fFileName != "" || config.fStructName != "" {
				log.Fatal().Msgf("Cannot specify -filename or -structname with regex in -name")
			}
		} else {
			filter = regexp.MustCompile(fmt.Sprintf("^%s$", config.fName))
			limitOne = true
		}
	} else if config.fAll {
		recursive = true
		filter = regexp.MustCompile(".*")
	} else {
		log.Fatal().Msgf("Use -name to specify the name of the interface or -all for all interfaces found")
	}

	if config.fkeepTree {
		config.fIP = false
	}

	if config.fProfile != "" {
		f, err := os.Create(config.fProfile)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to create profile file")
		}

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	var osp mockery.OutputStreamProvider
	if config.fPrint {
		osp = &mockery.StdoutStreamProvider{}
	} else {
		osp = &mockery.FileOutputStreamProvider{
			BaseDir:                   config.fOutput,
			InPackage:                 config.fIP,
			TestOnly:                  config.fTO,
			Case:                      config.fCase,
			KeepTree:                  config.fkeepTree,
			KeepTreeOriginalDirectory: config.fDir,
			FileName:                  config.fFileName,
		}
	}

	baseDir := config.fDir

	if config.fSrcPkg != "" {
		pkgs, err := packages.Load(&packages.Config{
			Mode: packages.NeedFiles,
		}, config.fSrcPkg)
		if err != nil || len(pkgs) == 0 {
			log.Fatal().Err(err).Msgf("Failed to load package %s", config.fSrcPkg)
		}

		// NOTE: we only pass one package name (config.fSrcPkg) to packages.Load
		// it should return one package at most
		pkg := pkgs[0]

		if pkg.Errors != nil {
			log.Fatal().Err(pkg.Errors[0]).Msgf("Failed to load package %s", config.fSrcPkg)
		}

		if len(pkg.GoFiles) == 0 {
			log.Fatal().Msgf("No go files in package %s", config.fSrcPkg)
		}
		baseDir = filepath.Dir(pkg.GoFiles[0])
	}

	visitor := &mockery.GeneratorVisitor{
		InPackage:         config.fIP,
		Note:              config.fNote,
		Osp:               osp,
		PackageName:       config.fOutpkg,
		PackageNamePrefix: config.fPackagePrefix,
		StructName:        config.fStructName,
	}

	walker := mockery.Walker{
		BaseDir:   baseDir,
		Recursive: recursive,
		Filter:    filter,
		LimitOne:  limitOne,
		BuildTags: strings.Split(config.buildTags, " "),
	}

	generated := walker.Walk(ctx, visitor)

	if config.fName != "" && !generated {
		log.Fatal().Msgf("Unable to find '%s' in any go files under this path", config.fName)
	}
}

func parseConfigFromArgs(args []string) Config {
	config := Config{}

	flagSet := flag.NewFlagSet(args[0], flag.ExitOnError)

	flagSet.StringVar(&config.fName, "name", "", "name or matching regular expression of interface to generate mock for")
	flagSet.BoolVar(&config.fPrint, "print", false, "print the generated mock to stdout")
	flagSet.StringVar(&config.fOutput, "output", "./mocks", "directory to write mocks to")
	flagSet.StringVar(&config.fOutpkg, "outpkg", "mocks", "name of generated package")
	flagSet.StringVar(&config.fPackagePrefix, "packageprefix", "", "prefix for the generated package name, it takes precedence over outpkg")
	flagSet.StringVar(&config.fDir, "dir", ".", "directory to search for interfaces")
	flagSet.BoolVar(&config.fRecursive, "recursive", false, "recurse search into sub-directories")
	flagSet.BoolVar(&config.fAll, "all", false, "generates mocks for all found interfaces in all sub-directories")
	flagSet.BoolVar(&config.fIP, "inpkg", false, "generate a mock that goes inside the original package")
	flagSet.BoolVar(&config.fTO, "testonly", false, "generate a mock in a _test.go file")
	flagSet.StringVar(&config.fCase, "case", "camel", "name the mocked file using casing convention [camel, snake, underscore]")
	flagSet.StringVar(&config.fNote, "note", "", "comment to insert into prologue of each generated file")
	flagSet.StringVar(&config.fProfile, "cpuprofile", "", "write cpu profile to file")
	flagSet.BoolVar(&config.fVersion, "version", false, "prints the installed version of mockery")
	flagSet.BoolVar(&config.quiet, "quiet", false, "suppress output to stdout")
	flagSet.BoolVar(&config.fkeepTree, "keeptree", false, "keep the tree structure of the original interface files into a different repository. Must be used with XX")
	flagSet.StringVar(&config.buildTags, "tags", "", "space-separated list of additional build tags to use")
	flagSet.StringVar(&config.fFileName, "filename", "", "name of generated file (only works with -name and no regex)")
	flagSet.StringVar(&config.fStructName, "structname", "", "name of generated struct (only works with -name and no regex)")
	flagSet.StringVar(&config.fLogLevel, "log-level", "info", "Level of logging")
	flagSet.StringVar(&config.fSrcPkg, "srcpkg", "", "source pkg to search for interfaces")

	flagSet.Parse(args[1:])

	return config
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
