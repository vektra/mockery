package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/vektra/mockery/mockery"
)

var fName = flag.String("name", "", "name of interface to generate mock for")
var fPrint = flag.Bool("print", false, "print the generated mock to stdout")

func checkDir(p *mockery.Parser, dir, name string) bool {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, file := range files {
		path := filepath.Join(dir, file.Name())

		if file.IsDir() {
			ret := checkDir(p, path, name)
			if ret {
				return true
			}
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

	if *fName == "" {
		fmt.Fprintf(os.Stderr, "Use -name to specify the name of the interface")
		os.Exit(1)
	}

	parser := mockery.NewParser()

	ret := checkDir(parser, ".", *fName)
	if !ret {
		fmt.Printf("Unable to find %s in any go files under this path\n")
		os.Exit(1)
	}

	var out io.Writer

	if *fPrint {
		out = os.Stdout
	} else {
		os.Mkdir("mocks", 0755)

		f, err := os.Create(fmt.Sprintf("mocks/%s.go", *fName))
		if err != nil {
			fmt.Printf("Unable to create output file for generated mock: %s\n", err)
			os.Exit(1)
		}

		defer f.Close()

		out = f
	}

	gen := mockery.NewGenerator(parser, out)
	err := gen.Setup(*fName)
	if err != nil {
		fmt.Printf("Error with %s: %s\n", *fName, err)
		os.Exit(1)
	}

	gen.GeneratePrologue()
	gen.Generate()
}
