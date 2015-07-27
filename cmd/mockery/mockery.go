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

var fName = flag.String("name", "", "name of interface to generate mock for")
var fPrint = flag.Bool("print", false, "print the generated mock to stdout")
var fOutput = flag.String("output", "./mocks", "directory to write mocks to")
var fDir = flag.String("dir", ".", "directory to search for interfaces")
var fAll = flag.Bool("all", false, "generates mocks for all found interfaces")
var fIP = flag.Bool("inpkg", false, "generate a mock that goes inside the original package")
var fCase = flag.String("case", "camel", "name the mocked file using casing convention")
var fNote = flag.String("note", "", "comment to insert into prologue of each generated file")

func checkDir(p *mockery.Parser, dir, name string) bool {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		path := filepath.Join(dir, file.Name())

		if file.IsDir() {
			if *fAll {
				ret := checkDir(p, path, name)
				if ret {
					return true
				}
			} else {
				continue
			}
		}

		if !strings.HasSuffix(path, ".go") {
			continue
		}

		err = p.Parse(path)
		if err != nil {
			continue
		}

		node, err := p.Find(name)
		if err != nil {
			continue
		}

		if node != nil {
			return true
		}
	}

	return false
}

func main() {
	flag.Parse()

	if *fAll {
		mockAll()
	} else {
		mockFor(*fName)
	}
}

func walkDir(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		path := filepath.Join(dir, file.Name())

		if file.IsDir() {
			if *fAll {
				walkDir(path)
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
			genMock(iface)
		}
	}

	return
}

func mockAll() {
	walkDir(*fDir)
}

func mockFor(name string) {
	if name == "" {
		fmt.Fprintf(os.Stderr, "Use -name to specify the name of the interface\n")
		os.Exit(1)
	}

	parser := mockery.NewParser()

	ret := checkDir(parser, *fDir, name)
	if !ret {
		fmt.Printf("Unable to find %s in any go files under this path\n", name)
		os.Exit(1)
	}

	iface, err := parser.Find(name)
	if err != nil {
		fmt.Printf("Error finding %s: %s\n", name, err)
		os.Exit(1)
	}

	genMock(iface)
}

func genMock(iface *mockery.Interface) {
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
	if *fCase == "underscore" {
		caseName = underscoreCaseName(caseName)
	}

	if *fPrint {
		out = os.Stdout
	} else {
		var path string

		if *fIP {
			path = filepath.Join(filepath.Dir(iface.Path), "mock_"+caseName+".go")
		} else {
			path = filepath.Join(*fOutput, caseName+".go")
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

	if *fIP {
		gen.GenerateIPPrologue()
	} else {
		gen.GeneratePrologue(pkg)
	}

	gen.GeneratePrologueNote(*fNote)

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
