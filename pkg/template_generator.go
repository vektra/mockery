package pkg

import (
	"bytes"
	"context"
	"fmt"
	"go/types"

	"github.com/chigopher/pathlib"
	"github.com/rs/zerolog"
	"github.com/vektra/mockery/v2/pkg/config"
	"github.com/vektra/mockery/v2/pkg/registry"
	"github.com/vektra/mockery/v2/pkg/template"
	"golang.org/x/tools/go/packages"
)

type TemplateGeneratorConfig struct {
	Style string
}
type TemplateGenerator struct {
	config   TemplateGeneratorConfig
	registry *registry.Registry
}

func NewTemplateGenerator(srcPkg *packages.Package, config TemplateGeneratorConfig, outPkg string) (*TemplateGenerator, error) {
	reg, err := registry.New(srcPkg, outPkg)
	if err != nil {
		return nil, fmt.Errorf("creating new registry: %w", err)
	}

	return &TemplateGenerator{
		config:   config,
		registry: reg,
	}, nil
}

func (g *TemplateGenerator) format(src []byte, ifaceConfig *config.Config) ([]byte, error) {
	switch ifaceConfig.Formatter {
	case "goimports":
		return goimports(src)

	case "noop":
		return src, nil
	}

	return gofmt(src)
}

func (g *TemplateGenerator) Generate(ctx context.Context, iface *Interface, ifaceConfig *config.Config) error {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("generating templated mock for interface")

	imports := Imports{}
	for _, method := range iface.Methods() {
		method.populateImports(imports)
	}
	methods := make([]template.MethodData, iface.ActualInterface.NumMethods())

	for i := 0; i < iface.ActualInterface.NumMethods(); i++ {
		method := iface.ActualInterface.Method(i)
		methodScope := g.registry.MethodScope()

		signature := method.Type().(*types.Signature)
		params := make([]template.ParamData, signature.Params().Len())
		for j := 0; j < signature.Params().Len(); j++ {
			param := signature.Params().At(j)
			params[j] = template.ParamData{
				Var:      methodScope.AddVar(param, ""),
				Variadic: signature.Variadic() && j == signature.Params().Len()-1,
			}
		}

		returns := make([]template.ParamData, signature.Results().Len())
		for j := 0; j < signature.Results().Len(); j++ {
			param := signature.Results().At(j)
			returns[j] = template.ParamData{
				Var:      methodScope.AddVar(param, "Out"),
				Variadic: false,
			}
		}

		methods[i] = template.MethodData{
			Name:    method.Name(),
			Params:  params,
			Returns: returns,
		}

	}

	// For now, mockery only supports one mock per file, which is why we're creating
	// a single-element list. moq seems to have supported multiple mocks per file.
	mockData := []template.MockData{
		{
			InterfaceName: iface.Name,
			MockName:      ifaceConfig.MockName,
			Methods:       methods,
		},
	}
	data := template.Data{
		PkgName:         ifaceConfig.Outpkg,
		SrcPkgQualifier: iface.Pkg.Name() + ".",
		Mocks:           mockData,
		StubImpl:        false,
		SkipEnsure:      false,
		WithResets:      false,
	}
	if data.MocksSomeMethod() {
		log.Debug().Msg("interface mocks some method, importing sync package")
		g.registry.AddImport(types.NewPackage("sync", "sync"))
	}
	if g.registry.SrcPkgName() != ifaceConfig.Outpkg {
		data.SrcPkgQualifier = g.registry.SrcPkgName() + "."
		if !ifaceConfig.SkipEnsure {
			log.Debug().Str("src-pkg", g.registry.SrcPkg().Path()).Msg("skip-ensure is false. Adding import for source package.")
			imprt := g.registry.AddImport(g.registry.SrcPkg())
			log.Debug().Msgf("imprt: %v", imprt)
			data.SrcPkgQualifier = imprt.Qualifier() + "."
		}
	}
	data.Imports = g.registry.Imports()

	templ, err := template.New(g.config.Style)
	if err != nil {
		return fmt.Errorf("creating new template: %w", err)
	}

	var buf bytes.Buffer
	log.Debug().Msg("executing template")
	if err := templ.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	//log.Debug().Msg("formatting file")
	//formatted, err := g.format(buf.Bytes(), ifaceConfig)
	//if err != nil {
	//	return fmt.Errorf("formatting mock file: %w", err)
	//}
	formatted := buf.Bytes()

	outPath := pathlib.NewPath(ifaceConfig.Dir).Join(ifaceConfig.FileName)
	log.Debug().Stringer("path", outPath).Msg("writing to path")
	if err := outPath.WriteFile(formatted); err != nil {
		log.Error().Err(err).Msg("couldn't write to output file")
		return fmt.Errorf("writing to output file: %w", err)
	}
	return nil
}
