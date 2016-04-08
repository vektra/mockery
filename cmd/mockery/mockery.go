package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/vektra/mockery/mockery"
)

const regexMetadataChars = "\\.+*?()|[]{}^$"

type Config struct {
	fName      string
	fPrint     bool
	fOutput    string
	fDir       string
	fRecursive bool
	fAll       bool
	fIP        bool
	fTO        bool
	fCase      string
	fNote      string
	Verbose    bool
}

func main() {
	config := parseConfigFromArgs(os.Args)

	if !config.Verbose {
		log.SetOutput(ioutil.Discard)
	}

	log.Printf("Parsed Config: %s", config)

	var recursive bool
	var filter *regexp.Regexp
	var err error
	var limitOne bool

	if config.fName != "" && config.fAll {
		fmt.Fprintln(os.Stderr, "Specify -name or -all, but not both")
		os.Exit(1)
	} else if config.fName != "" {
		recursive = config.fRecursive
		if strings.ContainsAny(config.fName, regexMetadataChars) {
			if filter, err = regexp.Compile(config.fName); err != nil {
				fmt.Fprintln(os.Stderr, "Invalid regular expression provided to -name")
				os.Exit(1)
			}
		} else {
			filter = regexp.MustCompile(fmt.Sprintf("^%s$", config.fName))
			limitOne = true
		}
	} else if config.fAll {
		recursive = true
		filter = regexp.MustCompile(".*")
	} else {
		fmt.Fprintln(os.Stderr, "Use -name to specify the name of the interface or -all for all interfaces found")
		os.Exit(1)
	}

	var osp mockery.OutputStreamProvider
	if config.fPrint {
		osp = &mockery.StdoutStreamProvider{}
	} else {
		osp = &mockery.FileOutputStreamProvider{
			BaseDir:   config.fOutput,
			InPackage: config.fIP,
			TestOnly:  config.fTO,
			Case:      config.fCase,
		}
	}
	log.Printf("Set OutputProvider to %s", osp)

	visitor := &mockery.GeneratorVisitor{
		InPackage: config.fIP,
		Note:      config.fNote,
		Osp:       osp,
	}
	log.Printf("Using Visitor: %s", visitor)

	walker := mockery.Walker{
		BaseDir:   config.fDir,
		Recursive: recursive,
		Filter:    filter,
		LimitOne:  limitOne,
	}
	log.Printf("Using Walker: %s", walker)

	generated := walker.Walk(visitor)

	if config.fName != "" && !generated {
		fmt.Printf("Unable to find %s in any go files under this path\n", config.fName)
		os.Exit(1)
	}
	log.Print("Success")
}

func parseConfigFromArgs(args []string) Config {
	config := Config{}

	flagSet := flag.NewFlagSet(args[0], flag.ExitOnError)

	flagSet.StringVar(&config.fName, "name", "", "name or matching regular expression of interface to generate mock for")
	flagSet.BoolVar(&config.fPrint, "print", false, "print the generated mock to stdout")
	flagSet.StringVar(&config.fOutput, "output", "./mocks", "directory to write mocks to")
	flagSet.StringVar(&config.fDir, "dir", ".", "directory to search for interfaces")
	flagSet.BoolVar(&config.fRecursive, "recursive", false, "recurse search into sub-directories")
	flagSet.BoolVar(&config.fAll, "all", false, "generates mocks for all found interfaces in all sub-directories")
	flagSet.BoolVar(&config.fIP, "inpkg", false, "generate a mock that goes inside the original package")
	flagSet.BoolVar(&config.fTO, "testonly", false, "generate a mock in a _test.go file")
	flagSet.StringVar(&config.fCase, "case", "camel", "name the mocked file using casing convention")
	flagSet.StringVar(&config.fNote, "note", "", "comment to insert into prologue of each generated file")
	flagSet.BoolVar(&config.Verbose, "verbose", false, "Enables verbose logging")

	flagSet.Parse(args[1:])

	return config
}
