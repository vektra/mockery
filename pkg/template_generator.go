package pkg

import (
	"bytes"
	"context"
	"fmt"
	"go/format"
	"go/token"
	"go/types"

	"github.com/rs/zerolog"
	"github.com/vektra/mockery/v2/pkg/registry"
	"github.com/vektra/mockery/v2/pkg/template"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

type Formatter string

const (
	FORMAT_GOIMPORRTS Formatter = "goimports"
	FORMAT_NOOP       Formatter = "noop"
)

type TemplateGenerator struct {
	templateName string
	registry     *registry.Registry
	formatter    Formatter
	pkgConfig    *Config
	inPackage    bool
}

func NewTemplateGenerator(
	srcPkg *packages.Package,
	outPkgPath string,
	templateName string,
	formatter Formatter,
	pkgConfig *Config,
) (*TemplateGenerator, error) {
	reg, err := registry.New(srcPkg, outPkgPath)
	if err != nil {
		return nil, fmt.Errorf("creating new registry: %w", err)
	}

	return &TemplateGenerator{
		templateName: templateName,
		registry:     reg,
		formatter:    formatter,
		pkgConfig:    pkgConfig,
		inPackage:    srcPkg.PkgPath == outPkgPath,
	}, nil
}

func (g *TemplateGenerator) format(src []byte) ([]byte, error) {
	switch g.formatter {
	case FORMAT_GOIMPORRTS:
		return goimports(src)

	case FORMAT_NOOP:
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

func (g *TemplateGenerator) Generate(
	ctx context.Context,
	interfaces []*Interface,
) ([]byte, error) {
	log := zerolog.Ctx(ctx)
	mockData := []template.MockData{}
	for _, ifaceMock := range interfaces {
		iface, tparams, err := g.registry.LookupInterface(ifaceMock.Name)
		if err != nil {
			return []byte{}, err
		}

		methods := make([]template.MethodData, iface.NumMethods())
		for i := 0; i < iface.NumMethods(); i++ {
			methods[i] = g.methodData(ctx, iface.Method(i))
		}

		mockData = append(mockData, template.MockData{
			InterfaceName: ifaceMock.Name,
			MockName:      ifaceMock.Config.MockName,
			TypeParams:    g.typeParams(ctx, tparams),
			Methods:       methods,
			TemplateData:  ifaceMock.Config.TemplateData,
		})
	}

	data := template.Data{
		PkgName:         g.pkgConfig.PkgName,
		SrcPkgQualifier: "",
		Mocks:           mockData,
		TemplateData:    g.pkgConfig.TemplateData,
	}
	if !g.inPackage {
		data.SrcPkgQualifier = g.registry.SrcPkgName() + "."
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
	// TODO: Grabbing ifaceConfigs[0].Formatter doesn't make sense. We should instead
	// grab the formatter as specified in the topmost interface-level config.
	formatted, err := g.format(buf.Bytes())
	if err != nil {
		log.Err(err).Msg("can't format mock file, printing buffer.")
		fmt.Print(buf.String())
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
