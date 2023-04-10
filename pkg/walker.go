package pkg

import (
	"context"
	"fmt"
	"io"
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
	LimitOne  bool
	BuildTags []string
}

type WalkerVisitor interface {
	VisitWalk(context.Context, *Interface) error
}

func (w *Walker) Walk(ctx context.Context, visitor WalkerVisitor) (generated bool) {
	log := zerolog.Ctx(ctx)
	ctx = log.WithContext(ctx)

	log.Info().Msgf("Walking")
	log.Debug().Str("baseDir", w.BaseDir).Msg("starting walk at base dir")

	parser := NewParser(w.BuildTags)
	w.doWalk(ctx, parser, w.BaseDir)

	err := parser.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking: %v\n", err)
		os.Exit(1)
	}

	for _, iface := range parser.Interfaces() {
		if strings.HasPrefix(iface.Name, mockConstructorParamTypeNamePrefix) {
			continue
		}

		if w.ExcludePath(iface.FileName) {
			continue
		}

		if !w.Filter.MatchString(iface.Name) {
			continue
		}
		err := visitor.VisitWalk(ctx, iface)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking %s: %s\n", iface.Name, err)
			os.Exit(1)
		}
		generated = true
		if w.LimitOne {
			return
		}
	}

	return
}

func (w *Walker) doWalk(ctx context.Context, parser *Parser, dir string) (generated bool) {
	log := zerolog.Ctx(ctx)
	ctx = log.WithContext(ctx)

	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") || strings.HasPrefix(file.Name(), "_") {
			continue
		}

		path := filepath.Join(dir, file.Name())
		if w.ExcludePath(path) {
			continue
		}

		if file.IsDir() {
			if w.Recursive {
				generated = w.doWalk(ctx, parser, path) || generated
				if generated && w.LimitOne {
					return
				}
			}
			continue
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			continue
		}

		err = parser.Parse(ctx, path)
		if err != nil {
			log.Err(err).Msgf("Error parsing file")
			continue
		}
	}

	return
}

type GeneratorVisitorConfig struct {
	Boilerplate          string
	DisableVersionString bool
	Exported             bool
	InPackage            bool
	KeepTree             bool
	Note                 string
	// The name of the output package, if InPackage is false (defaults to "mocks")
	PackageName       string
	PackageNamePrefix string
	StructName        string
	UnrollVariadic    bool
	WithExpecter      bool
	ReplaceType       []string
}

type GeneratorVisitor struct {
	config       GeneratorVisitorConfig
	dryRun       bool
	outputStream OutputStreamProvider
}

func NewGeneratorVisitor(
	config GeneratorVisitorConfig,
	outputStream OutputStreamProvider,
	dryRun bool,
) *GeneratorVisitor {
	return &GeneratorVisitor{
		config:       config,
		dryRun:       dryRun,
		outputStream: outputStream,
	}
}

func (v *GeneratorVisitor) VisitWalk(ctx context.Context, iface *Interface) error {
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

	out, err, closer := v.outputStream.GetWriter(ctx, iface)
	if err != nil {
		log.Err(err).Msgf("Unable to get writer")
		os.Exit(1)
	}
	defer func() {
		if err := closer(); err != nil {
			log.Err(err).Msgf("Failed to close output stream")
		}
	}()

	generatorConfig := GeneratorConfig{
		Boilerplate:          v.config.Boilerplate,
		DisableVersionString: v.config.DisableVersionString,
		Exported:             v.config.Exported,
		InPackage:            v.config.InPackage,
		KeepTree:             v.config.KeepTree,
		Note:                 v.config.Note,
		PackageName:          v.config.PackageName,
		PackageNamePrefix:    v.config.PackageNamePrefix,
		StructName:           v.config.StructName,
		UnrollVariadic:       v.config.UnrollVariadic,
		WithExpecter:         v.config.WithExpecter,
		ReplaceType:          v.config.ReplaceType,
	}

	gen := NewGenerator(ctx, generatorConfig, iface, "")
	log.Info().Msgf("Generating mock")
	if err := gen.GenerateAll(ctx); err != nil {
		log.Err(err).Msg("generation failed")
		return err
	}

	if !v.dryRun {
		log.Info().Msgf("writing mock to file")
		err = gen.Write(out)
		if err != nil {
			return err
		}
	}

	return nil
}
