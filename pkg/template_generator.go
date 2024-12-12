package pkg

import (
	"bytes"
	"context"
	"fmt"
	"go/format"
	"go/token"
	"go/types"

	"github.com/rs/zerolog"
	"github.com/vektra/mockery/v2/pkg/config"
	"github.com/vektra/mockery/v2/pkg/registry"
	"github.com/vektra/mockery/v2/pkg/template"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

type TemplateGenerator struct {
	templateName string
	registry     *registry.Registry
}

func NewTemplateGenerator(srcPkg *packages.Package, outPkgPath string, templateName string) (*TemplateGenerator, error) {
	reg, err := registry.New(srcPkg, outPkgPath)
	if err != nil {
		return nil, fmt.Errorf("creating new registry: %w", err)
	}

	return &TemplateGenerator{
		templateName: templateName,
		registry:     reg,
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

func (g *TemplateGenerator) methodData(ctx context.Context, method *types.Func) template.MethodData {
	methodScope := g.registry.MethodScope()

	signature := method.Type().(*types.Signature)
	params := make([]template.ParamData, signature.Params().Len())
	for j := 0; j < signature.Params().Len(); j++ {
		param := signature.Params().At(j)
		params[j] = template.ParamData{
			Var:      methodScope.AddVar(ctx, param, ""),
			Variadic: signature.Variadic() && j == signature.Params().Len()-1,
		}
	}

	returns := make([]template.ParamData, signature.Results().Len())
	for j := 0; j < signature.Results().Len(); j++ {
		param := signature.Results().At(j)
		returns[j] = template.ParamData{
			Var:      methodScope.AddVar(ctx, param, "Out"),
			Variadic: false,
		}
	}
	return template.MethodData{
		Name:    method.Name(),
		Params:  params,
		Returns: returns,
	}
}

func explicitConstraintType(typeParam *types.Var) (t types.Type) {
	underlying := typeParam.Type().Underlying().(*types.Interface)
	// check if any of the embedded types is either a basic type or a union,
	// because the generic type has to be an alias for one of those types then
	for j := 0; j < underlying.NumEmbeddeds(); j++ {
		t := underlying.EmbeddedType(j)
		switch t := t.(type) {
		case *types.Basic:
			return t
		case *types.Union: // only unions of basic types are allowed, so just take the first one as a valid type constraint
			return t.Term(0).Type()
		}
	}
	return nil
}

func (g *TemplateGenerator) typeParams(ctx context.Context, tparams *types.TypeParamList) []template.TypeParamData {
	var tpd []template.TypeParamData
	if tparams == nil {
		return tpd
	}

	tpd = make([]template.TypeParamData, tparams.Len())

	scope := g.registry.MethodScope()
	for i := 0; i < len(tpd); i++ {
		tp := tparams.At(i)
		typeParam := types.NewParam(token.Pos(i), tp.Obj().Pkg(), tp.Obj().Name(), tp.Constraint())
		tpd[i] = template.TypeParamData{
			ParamData:  template.ParamData{Var: scope.AddVar(ctx, typeParam, "")},
			Constraint: explicitConstraintType(typeParam),
		}
	}

	return tpd
}

func (g *TemplateGenerator) Generate(ctx context.Context, ifaceName string, ifaceConfig *config.Config) ([]byte, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("generating templated mock for interface")

	iface, tparams, err := g.registry.LookupInterface(ifaceName)
	if err != nil {
		return []byte{}, err
	}

	methods := make([]template.MethodData, iface.NumMethods())
	for i := 0; i < iface.NumMethods(); i++ {
		methods[i] = g.methodData(ctx, iface.Method(i))
	}

	// For now, mockery only supports one mock per file, which is why we're creating
	// a single-element list. moq seems to have supported multiple mocks per file.
	mockData := []template.MockData{
		{
			InterfaceName: ifaceName,
			MockName:      ifaceConfig.MockName,
			TypeParams:    g.typeParams(ctx, tparams),
			Methods:       methods,
		},
	}
	data := template.Data{
		PkgName:     ifaceConfig.Outpkg,
		Mocks:       mockData,
		TemplateMap: ifaceConfig.TemplateMap,
		StubImpl:    false,
	}
	if data.MocksSomeMethod() {
		log.Debug().Msg("interface mocks some method, importing sync package")
		g.registry.AddImport(ctx, types.NewPackage("sync", "sync"))
	}

	var inPackage bool
	if ifaceConfig.Dir == g.registry.SrcPkg().Types.Path() {
		log.Debug().Str("iface-dir", ifaceConfig.Dir).Str("pkg-path", g.registry.SrcPkg().Types.Path()).Msg("interface is inpackage")
		inPackage = true
	}

	if !inPackage {
		data.SrcPkgQualifier = g.registry.SrcPkgName() + "."
		skipEnsure, ok := ifaceConfig.TemplateMap["skip-ensure"]
		if !ok || !skipEnsure.(bool) {
			log.Debug().Str("src-pkg", g.registry.SrcPkg().PkgPath).Msg("skip-ensure is false. Adding import for source package.")
			imprt := g.registry.AddImport(ctx, g.registry.SrcPkg().Types)
			log.Debug().Msgf("imprt: %v", imprt)
			data.SrcPkgQualifier = imprt.Qualifier() + "."
		}
	}
	data.Imports = g.registry.Imports()

	templ, err := template.New(g.templateName)
	if err != nil {
		return []byte{}, fmt.Errorf("creating new template: %w", err)
	}

	var buf bytes.Buffer
	log.Debug().Msg("executing template")
	if err := templ.Execute(&buf, data); err != nil {
		return []byte{}, fmt.Errorf("executing template: %w", err)
	}

	log.Debug().Msg("formatting file")
	formatted, err := g.format(buf.Bytes(), ifaceConfig)
	if err != nil {
		return []byte{}, fmt.Errorf("formatting mock file: %w", err)
	}
	return formatted, nil
}

func goimports(src []byte) ([]byte, error) {
	formatted, err := imports.Process("filename", src, &imports.Options{
		TabWidth:  8,
		TabIndent: true,
		Comments:  true,
		Fragment:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("goimports: %s", err)
	}

	return formatted, nil
}

func gofmt(src []byte) ([]byte, error) {
	formatted, err := format.Source(src)
	if err != nil {
		return nil, fmt.Errorf("go/format: %s", err)
	}

	return formatted, nil
}
