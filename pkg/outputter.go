package pkg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/chigopher/pathlib"
	"github.com/iancoleman/strcase"
	"github.com/rs/zerolog"

	"github.com/vektra/mockery/v2/pkg/config"
	"github.com/vektra/mockery/v2/pkg/logging"
	"github.com/vektra/mockery/v2/pkg/stackerr"
)

var ErrInfiniteLoop = fmt.Errorf("infinite loop in template variables detected")

// Functions available in the template for manipulating
//
// Since the map and its functions are stateless, it exists as
// a package var rather than being initialized on every call
// in [parseConfigTemplates] and [generator.printTemplate]
var templateFuncMap = template.FuncMap{
	// String inspection and manipulation
	"contains":    strings.Contains,
	"hasPrefix":   strings.HasPrefix,
	"hasSuffix":   strings.HasSuffix,
	"join":        strings.Join,
	"replace":     strings.Replace,
	"replaceAll":  strings.ReplaceAll,
	"split":       strings.Split,
	"splitAfter":  strings.SplitAfter,
	"splitAfterN": strings.SplitAfterN,
	"trim":        strings.Trim,
	"trimLeft":    strings.TrimLeft,
	"trimPrefix":  strings.TrimPrefix,
	"trimRight":   strings.TrimRight,
	"trimSpace":   strings.TrimSpace,
	"trimSuffix":  strings.TrimSuffix,

	// Regular expression matching
	"matchString": regexp.MatchString,
	"quoteMeta":   regexp.QuoteMeta,

	// Filepath manipulation
	"base":  filepath.Base,
	"clean": filepath.Clean,
	"dir":   filepath.Dir,

	// Basic access to reading environment variables
	"expandEnv": os.ExpandEnv,
	"getenv":    os.Getenv,
}

type Cleanup func() error

type OutputStreamProvider interface {
	GetWriter(context.Context, *Interface) (io.Writer, error, Cleanup)
}

type StdoutStreamProvider struct{}

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
	// ctx = log.WithContext(ctx)

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
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err, func() error { return nil }
		}
	} else if p.InPackage {
		path = filepath.Join(filepath.Dir(iface.FileName), p.filename(caseName))
	} else {
		path = filepath.Join(p.BaseDir, p.filename(caseName))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err, func() error { return nil }
		}
	}

	log = log.With().Str(logging.LogKeyPath, path).Logger()

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
func parseConfigTemplates(ctx context.Context, c *config.Config, iface *Interface) error {
	log := zerolog.Ctx(ctx)

	isExported := ast.IsExported(iface.Name)
	var mock string
	if isExported {
		mock = "Mock"
	} else {
		mock = "mock"
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	var interfaceDirRelative string
	interfaceDir := pathlib.NewPath(iface.FileName).Parent()
	interfaceDirRelativePath, err := interfaceDir.RelativeToStr(workingDir)
	if errors.Is(err, pathlib.ErrRelativeTo) {
		log.Debug().
			Stringer("interface-dir", interfaceDir).
			Str("working-dir", workingDir).
			Msg("can't make interfaceDir relative to working dir. Setting InterfaceDirRelative to package path.")

		interfaceDirRelative = iface.Pkg.Path()
	} else {
		interfaceDirRelative = interfaceDirRelativePath.String()
	}

	// data is the struct sent to the template parser
	data := struct {
		InterfaceDir            string
		InterfaceDirRelative    string
		InterfaceName           string
		InterfaceNameCamel      string
		InterfaceNameLowerCamel string
		InterfaceNameSnake      string
		InterfaceNameLower      string
		Mock                    string
		MockName                string
		PackageName             string
		PackagePath             string
	}{
		InterfaceDir:            filepath.Dir(iface.FileName),
		InterfaceDirRelative:    interfaceDirRelative,
		InterfaceName:           iface.Name,
		InterfaceNameCamel:      strcase.ToCamel(iface.Name),
		InterfaceNameLowerCamel: strcase.ToLowerCamel(iface.Name),
		InterfaceNameSnake:      strcase.ToSnake(iface.Name),
		InterfaceNameLower:      strings.ToLower(iface.Name),
		Mock:                    mock,
		MockName:                c.MockName,
		PackageName:             iface.Pkg.Name(),
		PackagePath:             iface.Pkg.Path(),
	}
	// These are the config options that we allow
	// to be parsed by the templater. The keys are
	// just labels we're using for logs/errors
	templateMap := map[string]*string{
		"filename": &c.FileName,
		"dir":      &c.Dir,
		"mockname": &c.MockName,
		"outpkg":   &c.Outpkg,
	}

	numIterations := 0
	changesMade := true
	for changesMade {
		if numIterations >= 20 {
			msg := "infinite loop in template variables detected"
			log.Error().Msg(msg)
			for key, val := range templateMap {
				l := log.With().Str("variable-name", key).Str("variable-value", *val).Logger()
				l.Error().Msg("config variable value")
			}
			return ErrInfiniteLoop
		}
		// Templated variables can refer to other templated variables,
		// so we need to continue parsing the templates until it can't
		// be parsed anymore.
		changesMade = false

		for name, attributePointer := range templateMap {
			oldVal := *attributePointer

			attributeTempl, err := template.New("interface-template").Funcs(templateFuncMap).Parse(*attributePointer)
			if err != nil {
				return fmt.Errorf("failed to parse %s template: %w", name, err)
			}
			var parsedBuffer bytes.Buffer

			if err := attributeTempl.Execute(&parsedBuffer, data); err != nil {
				return fmt.Errorf("failed to execute %s template: %w", name, err)
			}
			*attributePointer = parsedBuffer.String()
			if *attributePointer != oldVal {
				changesMade = true
			}
		}
		numIterations += 1
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
		interfaceConfig.LogUnsupportedPackagesConfig(ctx)

		log.Debug().Msg("getting mock generator")

		if err := parseConfigTemplates(ctx, interfaceConfig, iface); err != nil {
			return fmt.Errorf("failed to parse config template: %w", err)
		}

		g := GeneratorConfig{
			Boilerplate:          m.boilerplate,
			DisableVersionString: interfaceConfig.DisableVersionString,
			Exported:             interfaceConfig.Exported,
			InPackage:            interfaceConfig.InPackage,
			KeepTree:             interfaceConfig.KeepTree,
			Note:                 interfaceConfig.Note,
			PackageName:          interfaceConfig.Outpkg,
			PackageNamePrefix:    interfaceConfig.Packageprefix,
			StructName:           interfaceConfig.MockName,
			UnrollVariadic:       interfaceConfig.UnrollVariadic,
			WithExpecter:         interfaceConfig.WithExpecter,
			ReplaceType:          interfaceConfig.ReplaceType,
		}
		generator := NewGenerator(ctx, g, iface, "")

		log.Debug().Msg("generating mock in-memory")
		if err := generator.GenerateAll(ctx); err != nil {
			return err
		}

		outputPath := pathlib.NewPath(interfaceConfig.Dir).Join(interfaceConfig.FileName)
		if err := outputPath.Parent().MkdirAll(); err != nil {
			return stackerr.NewStackErrf(err, "failed to mkdir parents of: %v", outputPath)
		}

		fileLog := log.With().Stringer(logging.LogKeyFile, outputPath).Logger()
		fileLog.Info().Msg("writing to file")
		file, err := outputPath.OpenFile(os.O_RDWR | os.O_CREATE | os.O_TRUNC)
		if err != nil {
			return stackerr.NewStackErrf(err, "failed to open output file for mock: %v", outputPath)
		}
		defer file.Close()
		if err := generator.Write(file); err != nil {
			return stackerr.NewStackErrf(err, "failed to write to file")
		}
	}
	return nil
}
