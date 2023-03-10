package pkg

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/chigopher/pathlib"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/vektra/mockery/v2/pkg/config"
	"github.com/vektra/mockery/v2/pkg/logging"
)

type Cleanup func() error

type OutputStreamProvider interface {
	GetWriter(context.Context, *Interface) (io.Writer, error, Cleanup)
}

type StdoutStreamProvider struct {
}

func (*StdoutStreamProvider) GetWriter(ctx context.Context, iface *Interface) (io.Writer, error, Cleanup) {
	return os.Stdout, nil, func() error { return nil }
}

type FileOutputStreamProvider struct {
	Config                    config.Config
	BaseDir                   string
	InPackage                 bool
	InPackageSuffix           bool
	TestOnly                  bool
	Case                      string
	KeepTree                  bool
	KeepTreeOriginalDirectory string
	FileName                  string
}

func (p *FileOutputStreamProvider) GetWriter(ctx context.Context, iface *Interface) (io.Writer, error, Cleanup) {
	log := zerolog.Ctx(ctx).With().Str(logging.LogKeyInterface, iface.Name).Logger()
	ctx = log.WithContext(ctx)

	var path string

	caseName := iface.Name
	if p.Case == "underscore" || p.Case == "snake" {
		caseName = p.underscoreCaseName(caseName)
	}

	if p.KeepTree {
		absOriginalDir, err := filepath.Abs(p.KeepTreeOriginalDirectory)
		if err != nil {
			return nil, err, func() error { return nil }
		}
		relativePath := strings.TrimPrefix(
			filepath.Join(filepath.Dir(iface.FileName), p.filename(caseName)),
			absOriginalDir)

		// as it's not possible to import from internal path, we have to replace it in mocks when KeepTree is used
		relativePath = strings.Replace(relativePath, "/internal/", "/internal_/", -1)

		path = filepath.Join(p.BaseDir, relativePath)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, err, func() error { return nil }
		}
	} else if p.InPackage {
		path = filepath.Join(filepath.Dir(iface.FileName), p.filename(caseName))
	} else {
		path = filepath.Join(p.BaseDir, p.filename(caseName))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, err, func() error { return nil }
		}
	}

	log = log.With().Str(logging.LogKeyPath, path).Logger()
	ctx = log.WithContext(ctx)

	log.Debug().Msgf("creating writer to file")
	f, err := os.Create(path)
	if err != nil {
		return nil, err, func() error { return nil }
	}

	return f, nil, func() error {
		return f.Close()
	}
}

func (p *FileOutputStreamProvider) filename(name string) string {
	if p.FileName != "" {
		return p.FileName
	}

	if p.InPackage && p.TestOnly {
		if p.InPackageSuffix {
			return name + "_mock_test.go"
		}

		return "mock_" + name + "_test.go"
	} else if p.InPackage && !p.KeepTree {
		if p.InPackageSuffix {
			return name + "_mock.go"
		}

		return "mock_" + name + ".go"
	} else if p.TestOnly {
		return name + "_test.go"
	}

	return name + ".go"
}

// shamelessly taken from http://stackoverflow.com/questions/1175208/elegant-python-function-to-convert-camelcase-to-camel-caseo
func (*FileOutputStreamProvider) underscoreCaseName(caseName string) string {
	rxp1 := regexp.MustCompile("(.)([A-Z][a-z]+)")
	s1 := rxp1.ReplaceAllString(caseName, "${1}_${2}")
	rxp2 := regexp.MustCompile("([a-z0-9])([A-Z])")
	return strings.ToLower(rxp2.ReplaceAllString(s1, "${1}_${2}"))
}

// parseConfigTemplates parses various templated strings
// in the config struct into their fully defined values. This mutates
// the config object passed.
func parseConfigTemplates(c *config.Config, iface *Interface) error {
	// data is the struct sent to the template parser
	data := struct {
		InterfaceDir            string
		InterfaceName           string
		InterfaceNameCamel      string
		InterfaceNameLowerCamel string
		InterfaceNameSnake      string
		PackageName             string
		PackagePath             string
	}{
		InterfaceDir:            filepath.Dir(iface.FileName),
		InterfaceName:           iface.Name,
		InterfaceNameCamel:      strcase.ToCamel(iface.Name),
		InterfaceNameLowerCamel: strcase.ToLowerCamel(iface.Name),
		InterfaceNameSnake:      strcase.ToSnake(iface.Name),
		PackageName:             iface.Pkg.Name(),
		PackagePath:             iface.Pkg.Path(),
	}
	templ := template.New("interface-template")

	// These are the config options that we allow
	// to be parsed by the templater. The keys are
	// just labels we're using for logs/errors
	templateMap := map[string]*string{
		"filename":   &c.FileName,
		"dir":        &c.Dir,
		"structname": &c.StructName,
		"outpkg":     &c.Outpkg,
	}

	for name, attributePointer := range templateMap {
		attributeTempl, err := templ.Parse(*attributePointer)
		if err != nil {
			return fmt.Errorf("failed to parse %s template: %w", name, err)
		}
		var parsedBuffer bytes.Buffer

		if err := attributeTempl.Execute(&parsedBuffer, data); err != nil {
			return fmt.Errorf("failed to execute %s template: %w", name, err)
		}
		*attributePointer = parsedBuffer.String()
	}

	return nil

}

// Outputter wraps the Generator struct. It calls the generator
// to create the mock implementations in-memory, then has additional
// logic to determine where the mock should be written to on disk.
type Outputter struct {
	boilerplate string
	config      *config.Config
	dryRun      bool
}

func NewOutputter(
	config *config.Config,
	boilerplate string,
	dryRun bool,
) *Outputter {
	return &Outputter{
		boilerplate: boilerplate,
		config:      config,
		dryRun:      dryRun,
	}
}

func (m *Outputter) Generate(ctx context.Context, iface *Interface) error {
	log := zerolog.Ctx(ctx).With().
		Str(logging.LogKeyInterface, iface.Name).
		Str(logging.LogKeyQualifiedName, iface.QualifiedName).
		Logger()
	ctx = log.WithContext(ctx)

	shouldGenerate, err := m.config.ShouldGenerateInterface(ctx, iface.QualifiedName, iface.Name)
	if err != nil {
		return err
	}
	if !shouldGenerate {
		log.Debug().Msg("config doesn't specify to generate this interface, skipping.")
		return nil
	}
	log.Debug().Msg("config specifies to generate this interface")
	log.Info().Msg("generating mocks for interface")

	log.Debug().Msg("getting config for interface")
	interfaceConfigs, err := m.config.GetInterfaceConfig(ctx, iface.QualifiedName, iface.Name)
	if err != nil {
		return err
	}

	for _, interfaceConfig := range interfaceConfigs {
		log.Debug().Msg("getting mock generator")

		parseConfigTemplates(interfaceConfig, iface)

		g := GeneratorConfig{
			Boilerplate:          m.boilerplate,
			DisableVersionString: interfaceConfig.DisableVersionString,
			Exported:             interfaceConfig.Exported,
			InPackage:            interfaceConfig.InPackage,
			KeepTree:             interfaceConfig.KeepTree,
			Note:                 interfaceConfig.Note,
			PackageName:          interfaceConfig.Outpkg,
			PackageNamePrefix:    interfaceConfig.Packageprefix,
			StructName:           interfaceConfig.StructName,
			UnrollVariadic:       interfaceConfig.UnrollVariadic,
			WithExpecter:         interfaceConfig.WithExpecter,
		}
		generator := NewGenerator(ctx, g, iface, "")

		log.Debug().Msg("generating mock in-memory")
		if err := generator.GenerateAll(ctx); err != nil {
			return err
		}

		outputPath := pathlib.NewPath(interfaceConfig.Dir).Join(interfaceConfig.FileName)
		if err := outputPath.Parent().MkdirAll(); err != nil {
			return errors.Wrapf(err, "failed to mkdir parents of: %v", outputPath)
		}

		fileLog := log.With().Stringer(logging.LogKeyFile, outputPath).Logger()
		fileLog.Info().Msg("writing to file")
		file, err := outputPath.OpenFile(os.O_RDWR | os.O_CREATE)
		if err != nil {
			return errors.Wrapf(err, "failed to open output file for mock: %v", outputPath)
		}
		defer file.Close()
		if err := generator.Write(file); err != nil {
			return errors.Wrapf(err, "failed to write to file")
		}
	}
	return nil
}
