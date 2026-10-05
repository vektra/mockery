package internal

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektra/mockery/v3/template"
	"golang.org/x/tools/go/packages"
)

type matryerSourceImporter struct {
	pkg *types.Package
}

func (i matryerSourceImporter) Import(path string) (*types.Package, error) {
	if path == i.pkg.Path() {
		return i.pkg, nil
	}
	return nil, fmt.Errorf("unexpected import %q", path)
}

func TestMatryerEnsureImports(t *testing.T) {
	const sourcePath = "example.com/source"
	const source = "package source; type First interface{}; type Second interface{}; type Value struct{}"

	for _, tt := range []struct {
		name          string
		fileSkip      bool
		interfaceSkip []bool
		inPackage     bool
		keepImport    bool
		explicitAlias bool
		wantImports   int
	}{
		{name: "default ensure", interfaceSkip: []bool{false}, wantImports: 1},
		{name: "interface skips ensure", interfaceSkip: []bool{true}},
		{name: "all interfaces skip ensure", interfaceSkip: []bool{true, true}},
		{name: "interface overrides file skip", fileSkip: true, interfaceSkip: []bool{false}, wantImports: 1},
		{name: "mixed interfaces", interfaceSkip: []bool{true, false}, wantImports: 1},
		{name: "mixed interfaces override file skip", fileSkip: true, interfaceSkip: []bool{false, true}, wantImports: 1},
		{name: "file and interfaces skip", fileSkip: true, interfaceSkip: []bool{true, true}},
		{name: "existing type import remains", interfaceSkip: []bool{true}, keepImport: true, wantImports: 1},
		{name: "later interface supplies source alias", interfaceSkip: []bool{false, false}, explicitAlias: true, wantImports: 1},
		{name: "skipped sibling supplies source alias", fileSkip: true, interfaceSkip: []bool{false, true}, explicitAlias: true, wantImports: 1},
		{name: "same package skips", interfaceSkip: []bool{true}, inPackage: true},
		{name: "same package ensures", fileSkip: true, interfaceSkip: []bool{false}, inPackage: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			sourceFile, err := parser.ParseFile(fset, "source.go", source, parser.SkipObjectResolution)
			require.NoError(t, err)
			sourceConfig := types.Config{}
			sourcePkg, err := sourceConfig.Check(sourcePath, fset, []*ast.File{sourceFile}, nil)
			require.NoError(t, err)

			pkgName, pkgPath, qualifier := "mocks", "example.com/mocks", "source."
			if tt.inPackage {
				pkgName, pkgPath, qualifier = "source", sourcePath, ""
			}
			registry, err := template.NewRegistry(&packages.Package{Name: "source", PkgPath: sourcePath, Types: sourcePkg}, pkgPath, tt.inPackage)
			require.NoError(t, err)
			if tt.keepImport {
				registry.AddImport("source", sourcePath)
			}
			interfaces := make(template.Interfaces, len(tt.interfaceSkip))
			for i, skip := range tt.interfaceSkip {
				name := []string{"First", "Second"}[i]
				interfaces[i] = template.Interface{
					Name: name, StructName: name + "Mock", TemplateData: template.TemplateData{"skip-ensure": skip},
				}
			}
			if tt.keepImport {
				interfaces[0].TemplateData["struct-preamble"] = "Value source.Value"
			}
			wantQualifier := qualifier
			if tt.explicitAlias {
				interfaces[1].TemplateData["add-import"] = []any{map[string]any{"name": "src", "pkgPath": sourcePath}}
				wantQualifier = "src."
			}
			data := template.Data{
				PkgName: pkgName, SrcPkgQualifier: qualifier, Registry: registry,
				TemplateData: template.TemplateData{"skip-ensure": tt.fileSkip}, Interfaces: interfaces,
			}
			tmpl, err := template.New(templateMatryer, "matryer")
			require.NoError(t, err)
			var output bytes.Buffer
			require.NoError(t, tmpl.Execute(&output, data))

			generated, err := parser.ParseFile(fset, "mocks.go", output.Bytes(), parser.SkipObjectResolution)
			require.NoError(t, err)
			files := []*ast.File{generated}
			if tt.inPackage {
				files = append(files, sourceFile)
			}
			checker := types.Config{Importer: matryerSourceImporter{pkg: sourcePkg}}
			_, err = checker.Check(pkgPath, fset, files, nil)
			require.NoError(t, err, output.String())
			require.Len(t, registry.Imports(), tt.wantImports)
			for i, skip := range tt.interfaceSkip {
				assertion := "var _ " + wantQualifier + interfaces[i].Name
				require.Equal(t, !skip, strings.Contains(output.String(), assertion), output.String())
			}
		})
	}
}
