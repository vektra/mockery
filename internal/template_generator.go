package internal

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"go/format"
	"go/token"
	"go/types"
	"os"
	"strings"

	"github.com/chigopher/pathlib"
	"github.com/rs/zerolog"
	"github.com/vektra/mockery/v3/config"
	"github.com/vektra/mockery/v3/internal/stackerr"
	"github.com/vektra/mockery/v3/template"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

type Formatter string

const (
	FormatGofmt     Formatter = "gofmt"
	FormatGoImports Formatter = "goimports"
	FormatNoop      Formatter = "noop"
)

var (
	//go:embed mock_matryer.templ
	templateMatryer string
	//go:embed mock_testify.templ
	templateTestify string
)

var styleTemplates = map[string]string{
	"matryer": templateMatryer,
	"testify": templateTestify,
}

// findPkgPath returns the fully-qualified go import path of a given dir. The
// dir must be relative to a go.mod file. In the case it isn't, an error is returned.
func findPkgPath(dirPath *pathlib.Path) (string, error) {
	if err := dirPath.MkdirAll(); err != nil {
		return "", stackerr.NewStackErr(err)
	}
	dir, err := dirPath.ResolveAll()
	if err != nil {
		return "", stackerr.NewStackErr(err)
	}
	var goModFile *pathlib.Path
	cursor := dir
	for i := 0; ; i++ {
		if i == 1000 {
			return "", stackerr.NewStackErr(errors.New("failed to find go.mod after 1000 iterations"))
		}
		goMod := cursor.Join("go.mod")
		goModExists, err := goMod.Exists()
		if err != nil {
			return "", stackerr.NewStackErr(err)
		}
		if !goModExists {
			parent := cursor.Parent()
			// Hit the root path
			if cursor.String() == parent.String() {
				return "", stackerr.NewStackErrf(
					ErrGoModNotFound, "parsing package path for %s", dir.String())
			}
			cursor = parent
			continue
		}
		goModFile = goMod
		break
	}
	dirRelative, err := dir.RelativeTo(goModFile.Parent())
	if err != nil {
		return "", stackerr.NewStackErr(err)
	}
	fileBytes, err := goModFile.ReadFile()
	if err != nil {
		return "", stackerr.NewStackErr(err)
	}
	scanner := bufio.NewScanner(bytes.NewReader(fileBytes))
	// Iterate over each line
	for scanner.Scan() {
		if !strings.HasPrefix(scanner.Text(), "module") {
			continue
		}
		moduleName := strings.Split(scanner.Text(), "module ")[1]
		return pathlib.NewPath(moduleName, pathlib.PathWithSeperator("/")).
			JoinPath(dirRelative).
			Clean().
			String(), nil
	}
	return "", stackerr.NewStackErr(ErrGoModInvalid)
}

type TemplateGenerator struct {
	templateName string
	registry     *template.Registry
	formatter    Formatter
	inPackage    bool
	pkgConfig    *config.Config
	pkgName      string
}

func NewTemplateGenerator(
	ctx context.Context,
	srcPkg *packages.Package,
	outPkgFSPath *pathlib.Path,
	templateName string,
	formatter Formatter,
	pkgConfig *config.Config,
	pkgName string,
) (*TemplateGenerator, error) {
	srcPkgFSPath := pathlib.NewPath(srcPkg.GoFiles[0]).Parent()
	log := zerolog.Ctx(ctx).With().
		Stringer("srcPkgFSPath", srcPkgFSPath).
		Stringer("outPkgFSPath", outPkgFSPath).
		Str("src-pkg-name", srcPkg.Name).
		Str("out-pkg-name", pkgName).
		Logger()
	if !outPkgFSPath.IsAbsolute() {
		cwd, err := os.Getwd()
		if err != nil {
			log.Err(err).Msg("failed to get current working directory")
			return nil, stackerr.NewStackErr(err)
		}
		outPkgFSPath = pathlib.NewPath(cwd).JoinPath(outPkgFSPath)
	}
	outPkgPath, err := findPkgPath(outPkgFSPath)
	if err != nil {
		log.Err(err).Msg("failed to find output package path")
		return nil, err
	}
	log = log.With().Str("outPkgPath", outPkgPath).Logger()

	var inPackage bool
	// Note: Technically, go allows test files to have a different package name
	// than non-test files. In this case, the test files have to import the source
	// package just as if it were in a different directory.
	if pkgName == srcPkg.Name && srcPkgFSPath.Equals(outPkgFSPath) {
		log.Debug().Msg("output package detected to be in-package of original package")
		inPackage = true
	} else {
		log.Debug().Msg("output package detected to not be in-package of original package")
	}

	reg, err := template.NewRegistry(srcPkg, outPkgPath, inPackage)
	if err != nil {
		return nil, fmt.Errorf("creating new registry: %w", err)
	}

	return &TemplateGenerator{
		templateName: templateName,
		registry:     reg,
		formatter:    formatter,
		inPackage:    inPackage,
		pkgConfig:    pkgConfig,
		pkgName:      pkgName,
	}, nil
}

func (g *TemplateGenerator) format(src []byte) ([]byte, error) {
	switch g.formatter {
	case FormatGoImports:
		return goimports(src)
	case FormatGofmt:
		return gofmt(src)
	case FormatNoop:
		return src, nil
	}

	return nil, fmt.Errorf("unknown formatter type: %s", g.formatter)
}

func (g *TemplateGenerator) methodData(ctx context.Context, method *types.Func, ifaceConfig *config.Config) (template.MethodData, error) {
	log := zerolog.Ctx(ctx)

	methodScope := g.registry.MethodScope()

	signature := method.Type().(*types.Signature)
	params := make([]template.ParamData, signature.Params().Len())

	for j := 0; j < signature.Params().Len(); j++ {
		param := signature.Params().At(j)
		log.Debug().Str("param-string", param.String()).Msg("found parameter")
		for _, imprt := range g.registry.Imports() {
			log.Debug().Str("import", imprt.Path()).Str("import-qualifier", imprt.Qualifier()).Msg("existing imports")
		}

		var paramPkgPath string
		var paramObjName string
		switch t := param.Type().(type) {
		case *types.Named:
			pkg := t.Obj().Pkg()
			if pkg != nil {
				paramPkgPath = pkg.Path()
			}
			paramObjName = t.Obj().Name()
		case *types.Alias:
			pkg := t.Obj().Pkg()
			if pkg != nil {
				paramPkgPath = pkg.Path()
			}
			paramObjName = t.Obj().Name()
		}
		replacement := ifaceConfig.GetReplacement(paramPkgPath, paramObjName)
		if replacement != nil {
			log.Debug().Str("replace-to-pkg-path", replacement.PkgPath).Str("replace-to-type-name", replacement.TypeName).Msg("found replacement")
		} else {
			log.Debug().Str("param-pkg-path", paramPkgPath).Msg("replacement not found")
		}
		v, err := methodScope.AddVar(ctx, param, "", replacement)
		if err != nil {
			return template.MethodData{}, err
		}
		params[j] = template.ParamData{
			Var:      v,
			Variadic: signature.Variadic() && j == signature.Params().Len()-1,
		}
	}

	returns := make([]template.ParamData, signature.Results().Len())
	for j := 0; j < signature.Results().Len(); j++ {
		param := signature.Results().At(j)

		var paramPkgPath string
		var paramObjName string
		switch t := param.Type().(type) {
		case *types.Named:
			pkg := t.Obj().Pkg()
			if pkg != nil {
				paramPkgPath = pkg.Path()
			}
			paramObjName = t.Obj().Name()
		case *types.Alias:
			pkg := t.Obj().Pkg()
			if pkg != nil {
				paramPkgPath = pkg.Path()
			}
			paramObjName = t.Obj().Name()
		}

		replacement := ifaceConfig.GetReplacement(paramPkgPath, paramObjName)
		v, err := methodScope.AddVar(ctx, param, "", replacement)
		if err != nil {
			return template.MethodData{}, err
		}
		returns[j] = template.ParamData{
			Var:      v,
			Variadic: false,
		}
	}
	return template.MethodData{
		Name:    method.Name(),
		Params:  params,
		Returns: returns,
		Scope:   methodScope,
	}, nil
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

func (g *TemplateGenerator) typeParams(ctx context.Context, tparams *types.TypeParamList) ([]template.TypeParamData, error) {
	var tpd []template.TypeParamData
	if tparams == nil {
		return tpd, nil
	}

	tpd = make([]template.TypeParamData, tparams.Len())

	scope := g.registry.MethodScope()
	for i := 0; i < len(tpd); i++ {
		tp := tparams.At(i)
		typeParam := types.NewParam(token.Pos(i), tp.Obj().Pkg(), tp.Obj().Name(), tp.Constraint())
		v, err := scope.AddVar(ctx, typeParam, "", nil)
		if err != nil {
			return nil, err
		}
		tpd[i] = template.TypeParamData{
			ParamData:  template.ParamData{Var: v},
			Constraint: explicitConstraintType(typeParam),
		}
	}

	return tpd, nil
}

func (g *TemplateGenerator) Generate(
	ctx context.Context,
	interfaces []*config.Interface,
) ([]byte, error) {
	log := zerolog.Ctx(ctx)
	mockData := []template.Interface{}
	for _, ifaceMock := range interfaces {
		ifaceLog := log.With().
			Str("interface-name", ifaceMock.Name).
			Str("package-path", ifaceMock.Pkg.PkgPath).
			Str("mock-name", *ifaceMock.Config.StructName).
			Logger()
		ctx := ifaceLog.WithContext(ctx)

		ifaceLog.Debug().Msg("looking up interface in registry")
		iface, tparams, err := g.registry.LookupInterface(ifaceMock.Name)
		if err != nil {
			log.Err(err).Msg("error looking up interface")
			return []byte{}, err
		}
		ifaceLog.Debug().Msg("found interface")

		methods := make([]template.MethodData, iface.NumMethods())
		for i := 0; i < iface.NumMethods(); i++ {
			methodData, err := g.methodData(ctx, iface.Method(i), ifaceMock.Config)
			if err != nil {
				return nil, err
			}
			methods[i] = methodData
		}
		// Now that all methods have been generated, we need to resolve naming
		// conflicts that arise between variable names and package qualifiers.
		for _, method := range methods {
			method.Scope.ResolveVariableNameCollisions(
				zerolog.
					Ctx(ctx).
					With().
					Str("method-name", method.Name).
					Logger().
					WithContext(ctx))
		}

		ifaceLog.Debug().Str("template-data", fmt.Sprintf("%v", ifaceMock.Config.TemplateData)).Msg("printing template data")
		tParams, err := g.typeParams(ctx, tparams)
		if err != nil {
			return nil, err
		}
		mockData = append(mockData, template.Interface{
			Name:         ifaceMock.Name,
			StructName:   *ifaceMock.Config.StructName,
			TypeParams:   tParams,
			Methods:      methods,
			TemplateData: ifaceMock.Config.TemplateData,
		})
	}

	data := template.Data{
		PkgName:         g.pkgName,
		SrcPkgQualifier: "",
		Interfaces:      mockData,
		TemplateData:    g.pkgConfig.TemplateData,
	}
	if !g.inPackage {
		data.SrcPkgQualifier = g.registry.SrcPkgName() + "."
	}
	data.Imports = g.registry.Imports()

	var templateString string
	if strings.HasPrefix(g.templateName, "file://") {
		templatePath := pathlib.NewPath(strings.SplitAfterN(g.templateName, "file://", 2)[1])
		templateBytes, err := templatePath.ReadFile()
		if err != nil {
			log.Err(err).Str("template-path", g.templateName).Msg("Failed to read template")
			return nil, err
		}
		templateString = string(templateBytes)
	} else {
		var styleExists bool
		templateString, styleExists = styleTemplates[g.templateName]
		if !styleExists {
			return nil, stackerr.NewStackErrf(nil, "template '%s' does not exist", g.templateName)
		}
	}

	templ, err := template.New(templateString, g.templateName)
	if err != nil {
		return []byte{}, fmt.Errorf("creating new template: %w", err)
	}

	var buf bytes.Buffer
	log.Debug().Msg("executing template")
	if err := templ.Execute(&buf, data); err != nil {
		return []byte{}, fmt.Errorf("executing template: %w", err)
	}

	log.Debug().Msg("formatting file in-memory")
	formatted, err := g.format(buf.Bytes())
	if err != nil {
		scanner := bufio.NewScanner(strings.NewReader(buf.String()))
		for i := 1; scanner.Scan(); i++ {
			fmt.Printf("%d:\t%s\n", i, scanner.Text())
		}
		log.Err(err).Msg("can't format mock file in-memory")
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
