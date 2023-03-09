package pkg

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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

// outputFilePath determines where a particular mock should reside on-disk. This function is
// specific to the `packages` config option. It respects the configuration provided in the
// `packages` section, but provides sensible defaults.
func outputFilePath(
	ctx context.Context,
	iface *Interface,
	interfaceConfig *config.Config,
	mockName string,
) (*pathlib.Path, error) {
	var filename string
	var outputdir string
	log := zerolog.Ctx(ctx)

	outputFileTemplateString := interfaceConfig.FileName
	outputDirTemplate := interfaceConfig.Dir

	log.Debug().Msgf("output filename template is: %v", outputFileTemplateString)
	log.Debug().Msgf("output dir template is: %v", outputDirTemplate)

	templ := templates.New("output-file-template")

	// The fields available to the template strings
	data := struct {
		InterfaceDir            string
		InterfaceName           string
		InterfaceNameCamel      string
		InterfaceNameLowerCamel string
		InterfaceNameSnake      string
		PackageName             string
		PackagePath             string
		MockName                string
	}{
		InterfaceDir:            filepath.Dir(iface.FileName),
		InterfaceName:           iface.Name,
		InterfaceNameCamel:      strcase.ToCamel(iface.Name),
		InterfaceNameLowerCamel: strcase.ToLowerCamel(iface.Name),
		InterfaceNameSnake:      strcase.ToSnake(iface.Name),
		PackageName:             iface.Pkg.Name(),
		PackagePath:             iface.Pkg.Path(),
		MockName:                mockName,
	}

	// Get the name of the file from a template string
	filenameTempl, err := templ.Parse(outputFileTemplateString)
	if err != nil {
		return nil, err
	}

	var filenameBuffer bytes.Buffer

	if err := filenameTempl.Execute(&filenameBuffer, data); err != nil {
		return nil, err
	}
	filename = filenameBuffer.String()
	log.Debug().Msgf("filename is: %v", filename)

	// Get the name of the output dir
	outputDirTempl, err := templ.Parse(outputDirTemplate)
	if err != nil {
		return nil, err
	}
	var outputDirBuffer bytes.Buffer
	if err := outputDirTempl.Execute(&outputDirBuffer, data); err != nil {
		return nil, err
	}
	outputdir = outputDirBuffer.String()

	return pathlib.NewPath(outputdir).Join(filename), nil
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

		outputPath, err := outputFilePath(
			ctx,
			iface,
			interfaceConfig,
			generator.mockName(),
		)
		if err != nil {
			return errors.Wrap(err, "failed to determine file path")
		}
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
