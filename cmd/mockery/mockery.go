package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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
}

func main() {
	config := Config{}

	flag.StringVar(&config.fName, "name", "", "name or matching regular expression of interface to generate mock for")
	flag.BoolVar(&config.fPrint, "print", false, "print the generated mock to stdout")
	flag.StringVar(&config.fOutput, "output", "./mocks", "directory to write mocks to")
	flag.StringVar(&config.fDir, "dir", ".", "directory to search for interfaces")
	flag.BoolVar(&config.fRecursive, "recursive", false, "recurse search into sub-directories")
	flag.BoolVar(&config.fAll, "all", false, "generates mocks for all found interfaces in all sub-directories")
	flag.BoolVar(&config.fIP, "inpkg", false, "generate a mock that goes inside the original package")
	flag.BoolVar(&config.fTO, "testonly", false, "generate a mock in a _test.go file")
	flag.StringVar(&config.fCase, "case", "camel", "name the mocked file using casing convention")
	flag.StringVar(&config.fNote, "note", "", "comment to insert into prologue of each generated file")

	flag.Parse()

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

	generated := walkDir(config, config.fDir, recursive, filter, limitOne)

	if config.fName != "" && !generated {
		fmt.Printf("Unable to find %s in any go files under this path\n", config.fName)
		os.Exit(1)
	}
}

func walkDir(config Config, dir string, recursive bool, filter *regexp.Regexp, limitOne bool) (generated bool) {
	files, err := ioutil.ReadDir(config.fDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		path := filepath.Join(dir, file.Name())

		if file.IsDir() {
			if recursive {
				generated = walkDir(config, path, recursive, filter, limitOne) || generated
				if generated && limitOne {
					return
				}
			}
			continue
		}

		if !strings.HasSuffix(path, ".go") {
			continue
		}

		p := mockery.NewParser()

		err = p.Parse(path)
		if err != nil {
			continue
		}

		for _, iface := range p.Interfaces() {
			if !filter.MatchString(iface.Name) {
				continue
			}
			genMock(iface, config)
			generated = true
			if limitOne {
				return
			}
		}
	}

	return
}

func genMock(iface *mockery.Interface, config Config) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Unable to generated mock for '%s': %s\n", iface.Name, r)
			return
		}
	}()

	var out io.Writer

	pkg := "mocks"
	name := iface.Name
	caseName := iface.Name
	if config.fCase == "underscore" {
		caseName = underscoreCaseName(caseName)
	}

	if config.fPrint {
		out = os.Stdout
	} else {
		var path string

		if config.fIP {
			path = filepath.Join(filepath.Dir(iface.Path), filename(caseName, config))
		} else {
			path = filepath.Join(config.fOutput, filename(caseName, config))
			os.MkdirAll(filepath.Dir(path), 0755)
			pkg = filepath.Base(filepath.Dir(path))
		}

		f, err := os.Create(path)
		if err != nil {
			fmt.Printf("Unable to create output file for generated mock: %s\n", err)
			os.Exit(1)
		}

		defer f.Close()

		out = f

		fmt.Printf("Generating mock for: %s\n", name)
	}

	gen := mockery.NewGenerator(iface)

	if config.fIP {
		gen.GenerateIPPrologue()
	} else {
		gen.GeneratePrologue(pkg)
	}

	gen.GeneratePrologueNote(config.fNote)

	err := gen.Generate()
	if err != nil {
		fmt.Printf("Error with %s: %s\n", name, err)
		os.Exit(1)
	}

	err = gen.Write(out)
	if err != nil {
		fmt.Printf("Error writing %s: %s\n", name, err)
		os.Exit(1)
	}
}

// shamelessly taken from http://stackoverflow.com/questions/1175208/elegant-python-function-to-convert-camelcase-to-camel-caseo
func underscoreCaseName(caseName string) string {
	rxp1 := regexp.MustCompile("(.)([A-Z][a-z]+)")
	s1 := rxp1.ReplaceAllString(caseName, "${1}_${2}")
	rxp2 := regexp.MustCompile("([a-z0-9])([A-Z])")
	return strings.ToLower(rxp2.ReplaceAllString(s1, "${1}_${2}"))
}

func filename(name string, config Config) string {
	if config.fIP && config.fTO {
		return "mock_" + name + "_test.go"
	} else if config.fIP {
		return "mock_" + name + ".go"
	} else if config.fTO {
		return name + "_test.go"
	}
	return name + ".go"
}
