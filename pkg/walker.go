package pkg

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/vektra/mockery/v2/pkg/config"
	"github.com/vektra/mockery/v2/pkg/logging"

	"github.com/rs/zerolog"
)

type Walker struct {
	config.Config
	BaseDir   string
	Recursive bool
	Filter    *regexp.Regexp
	BuildTags []string

	// Deprecated: This field is no longer used and does not affect the walking process.
	LimitOne bool
}

type WalkerVisitor interface {
	VisitWalk(context.Context, *Interface) error
}

func (this *Walker) Walk(ctx context.Context, visitor WalkerVisitor) bool {
	log := zerolog.Ctx(ctx)
	ctx = log.WithContext(ctx)

	log.Info().Msgf("Walking")

	parser := NewParser(this.BuildTags)
	err := this.doWalk(ctx, parser, this.BaseDir)
	if err != nil {
		log.Err(err).Msgf("Failed to walk target sources %q.", this.BaseDir)
		return false
	}

	if err = parser.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Error walking: %v\n", err)
		os.Exit(1)
	}

	var generated bool
	for _, iface := range parser.Interfaces() {
		if !this.Filter.MatchString(iface.Name) {
			continue
		}
		err = visitor.VisitWalk(ctx, iface)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking %s: %s\n", iface.Name, err)
			os.Exit(1)
		}

		generated = true
	}

	return generated
}

func (this *Walker) doWalk(ctx context.Context, p *Parser, pattern string) error {
	log := zerolog.Ctx(ctx)
	ctx = log.WithContext(ctx)

	err := p.Parse(ctx, pattern)
	if err != nil {
		log.Err(err).Msgf("Error parsing")
		return err
	}

	files, err := ioutil.ReadDir(pattern)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if os.IsNotExist(err) || !this.Recursive {
		// If we are not targeting a physical directory or we are not recursing we can exit early.
		return nil
	}

	if strings.HasPrefix(filepath.Dir(pattern), ".") || strings.HasPrefix(filepath.Dir(pattern), "_") {
		log.Debug().Msgf("Skipping path %q as it is prefixed with either '.' or '_'.", pattern)
		return nil
	}

	for _, fi := range files {
		if !fi.IsDir() {
			continue
		}

		if err = this.doWalk(ctx, p, filepath.Join(pattern, fi.Name())); err != nil {
			return err
		}
	}

	return nil
}

type GeneratorVisitor struct {
	config.Config
	InPackage bool
	Note      string
	Osp       OutputStreamProvider
	// The name of the output package, if InPackage is false (defaults to "mocks")
	PackageName       string
	PackageNamePrefix string
	StructName        string
}

func (this *GeneratorVisitor) VisitWalk(ctx context.Context, iface *Interface) error {
	log := zerolog.Ctx(ctx).With().
		Str(logging.LogKeyInterface, iface.Name).
		Str(logging.LogKeyQualifiedName, iface.QualifiedName).
		Logger()
	ctx = log.WithContext(ctx)

	defer func() {
		if r := recover(); r != nil {
			log.Error().Msgf("Unable to generate mock: %s", r)
			return
		}
	}()

	var out io.Writer
	var pkg string

	if this.InPackage {
		pkg = filepath.Dir(iface.FileName)
	} else if (this.PackageName == "" || this.PackageName == "mocks") && this.PackageNamePrefix != "" {
		// go with package name prefix only when package name is empty or default and package name prefix is specified
		pkg = fmt.Sprintf("%s%s", this.PackageNamePrefix, iface.Pkg.Name())
	} else {
		pkg = this.PackageName
	}

	out, err, closer := this.Osp.GetWriter(ctx, iface)
	if err != nil {
		log.Err(err).Msgf("Unable to get writer")
		os.Exit(1)
	}
	defer closer()

	gen := NewGenerator(ctx, this.Config, iface, pkg)
	gen.GeneratePrologueNote(this.Note)
	gen.GeneratePrologue(ctx, pkg)

	err = gen.Generate(ctx)
	if err != nil {
		return err
	}

	log.Info().Msgf("Generating mock")
	if !this.Config.DryRun {
		err = gen.Write(out)
		if err != nil {
			return err
		}
	}

	return nil
}
