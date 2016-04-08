package mockery

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Walker struct {
	BaseDir   string
	Recursive bool
	Filter    *regexp.Regexp
	LimitOne  bool
}

type WalkerVisitor interface {
	VisitWalk(*Interface) error
}

func (this *Walker) Walk(visitor WalkerVisitor) (generated bool) {
	return this.doWalk(this.BaseDir, visitor)
}

func (this *Walker) doWalk(dir string, visitor WalkerVisitor) (generated bool) {
	log.Printf("Walker::doWalk(%s)", dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Printf("Walker::doWalk: error reading directory %s: %s", dir, err)
		return
	}

	for _, file := range files {
		log.Printf("Walker::doWalk(%s): Checking file: %s", dir, file.Name())
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		path := filepath.Join(dir, file.Name())

		if file.IsDir() {
			log.Printf("Walker::doWalk(%s): %s is a directory", dir, file.Name())
			if this.Recursive {
				generated = this.doWalk(path, visitor) || generated
				if generated && this.LimitOne {
					log.Printf("Walker::doWalk(%s): is directory && generated && LimitOne", dir)
					return
				}
			}
			continue
		}

		if !strings.HasSuffix(path, ".go") {
			continue
		}

		p := NewParser()

		err = p.Parse(path)
		if err != nil {
			log.Printf("Walker::doWalk(%s): parse(%s) error: %s", dir, path, err)
			continue
		}
		for _, iface := range p.Interfaces() {
			log.Printf("Walker::doWalk(%s): checking interface: %s", dir, iface)
			if !this.Filter.MatchString(iface.Name) {
				log.Printf("Walker::doWalk(%s): %s does not match filter %s, skipping", dir, iface.Name, this.Filter)
				continue
			}
			err := visitor.VisitWalk(iface)
			if err != nil {
				fmt.Printf("Error walking %s: %s\n", iface.Name, err)
				os.Exit(1)
			}
			generated = true
			if this.LimitOne {
				log.Printf("Walker::doWalk(%s): generated && LimtOne", dir)
				return
			}
		}
	}

	return
}

type GeneratorVisitor struct {
	InPackage bool
	Note      string
	Osp       OutputStreamProvider
}

func (this *GeneratorVisitor) VisitWalk(iface *Interface) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Unable to generated mock for '%s': %s\n", iface.Name, r)
			return
		}
	}()

	var out io.Writer

	pkg := "mocks"

	out, err, closer := this.Osp.GetWriter(iface, pkg)
	if err != nil {
		fmt.Printf("Unable to get writer for %s: %s", iface.Name, err)
		os.Exit(1)
	}
	defer closer()

	gen := NewGenerator(iface)

	if this.InPackage {
		gen.GenerateIPPrologue()
	} else {
		gen.GeneratePrologue(pkg)
	}

	gen.GeneratePrologueNote(this.Note)

	err = gen.Generate()
	if err != nil {
		return err
	}

	err = gen.Write(out)
	if err != nil {
		return err
	}
	return nil
}
