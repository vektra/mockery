package pkg

import (
	"context"
	"errors"
	"go/ast"
	"go/types"
	"strings"

	"github.com/rs/zerolog"
	"golang.org/x/tools/go/packages"
)

type packageLoadEntry struct {
	pkgs []*packages.Package
	err  error
}

type Parser struct {
	parserPackages []*types.Package
	conf           packages.Config
}

func NewParser(buildTags []string) *Parser {
	var conf packages.Config
	conf.Mode = packages.NeedTypes |
		packages.NeedTypesSizes |
		packages.NeedSyntax |
		packages.NeedTypesInfo |
		packages.NeedImports |
		packages.NeedName |
		packages.NeedFiles |
		packages.NeedCompiledGoFiles

	if len(buildTags) > 0 {
		conf.BuildFlags = []string{"-tags", strings.Join(buildTags, ",")}
	}
	p := &Parser{
		parserPackages: make([]*types.Package, 0),
		conf:           conf,
	}
	return p
}

func (p *Parser) ParsePackages(ctx context.Context, packageNames []string) ([]*Interface, error) {
	log := zerolog.Ctx(ctx)
	interfaces := []*Interface{}

	packages, err := packages.Load(&p.conf, packageNames...)
	if err != nil {
		return nil, err
	}
	for _, pkg := range packages {
		pkgLog := log.With().Str("package", pkg.PkgPath).Logger()
		pkgCtx := pkgLog.WithContext(ctx)

		if len(pkg.GoFiles) == 0 {
			continue
		}
		for _, err := range pkg.Errors {
			log.Err(err).Msg("encountered error when loading package")
		}
		if len(pkg.Errors) != 0 {
			return nil, errors.New("error occurred when loading packages")
		}
		for fileIdx, file := range pkg.GoFiles {
			fileLog := pkgLog.With().Str("file", file).Logger()
			fileLog.Debug().Msg("found file")
			fileCtx := fileLog.WithContext(pkgCtx)

			fileSyntax := pkg.Syntax[fileIdx]
			nv := NewNodeVisitor(fileCtx)
			ast.Walk(nv, fileSyntax)

			scope := pkg.Types.Scope()
			for _, declaredInterface := range nv.declaredInterfaces {
				ifaceLog := fileLog.With().Str("interface", declaredInterface).Logger()

				obj := scope.Lookup(declaredInterface)

				typ, ok := obj.Type().(*types.Named)
				if !ok {
					ifaceLog.Debug().Msg("interface is not named, skipping")
					continue
				}

				if !types.IsInterface(obj.Type()) {
					ifaceLog.Debug().Msg("type is not an interface, skipping")
					continue
				}

				name := typ.Obj().Name()

				if typ.Obj().Pkg() == nil {
					continue
				}

				interfaces = append(interfaces, NewInterface(
					name,
					file,
					fileSyntax,
					pkg,
					// Leave the config nil because we don't yet know if
					// the interface should even be generated in the first
					// place.
					nil,
				))
			}
		}
	}
	return interfaces, nil
}
