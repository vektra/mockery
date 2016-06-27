package mockery

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"code.uber.internal/go-common.git/x/log"

	"golang.org/x/tools/imports"
)

// Generator is responsible for generating the string containing
// imports and the mock struct that will later be written out as file.
type Generator struct {
	buf bytes.Buffer

	ip                bool
	iface             *Interface
	pkg               string
	packageToName     map[string]string
	imports           []*ast.ImportSpec
	localPackageName  *string
	localizationCache map[string]string
}

// NewGenerator builds a Generator.
func NewGenerator(iface *Interface, pkg string, inPackage bool) *Generator {
	return &Generator{
		iface:             iface,
		pkg:               pkg,
		ip:                inPackage,
		localizationCache: make(map[string]string),
	}
}

func (g *Generator) getLocalPackageName() string {
	if g.localPackageName == nil {
		localName := g.getLocalPackageNameFromPackageMap(g.getPackageToName())
		g.localPackageName = &localName
	}
	return *g.localPackageName
}

func (g *Generator) getLocalPackageNameFromPackageMap(packageToName map[string]string) string {
	localPackageName := g.iface.Pkg.Name()
	for path, name := range packageToName {
		if localPackageName == name && path != g.getInterfacePackagePath() {
			return "_interfacePackage"
		}
	}
	return localPackageName
}

func (g *Generator) getInterfacePackagePath() string {
	return g.getLocalizedPathFromPackage(g.iface.Pkg)
}

func (g *Generator) getLocalizedPathFromPackage(pkg *types.Package) string {
	return g.getLocalizedPath(pkg.Path())
}

// TODO(@IvanMalison): Is there not a better way to get the actual
// import path of a package?
func (g *Generator) getLocalizedPath(path string) string {
	if strings.HasSuffix(path, ".go") {
		path, _ = filepath.Split(path)
	}
	if localized, ok := g.localizationCache[path]; ok {
		return localized
	}
	directories := strings.Split(path, string(filepath.Separator))
	numDirectories := len(directories)
	vendorIndex := -1
	for i := 1; i <= numDirectories; i++ {
		dir := directories[numDirectories-i]
		if dir == "vendor" {
			vendorIndex = numDirectories - i
			break
		}
	}

	toReturn := path
	if vendorIndex >= 0 {
		toReturn = filepath.Join(directories[vendorIndex+1:]...)
	} else if filepath.IsAbs(path) {
		packageRoot := filepath.Join(os.Getenv("GOPATH"), "src")
		if strings.HasPrefix(path, packageRoot) {
			packagePath, err := filepath.Rel(packageRoot, path)
			if err == nil {
				toReturn = packagePath
			} else {
				log.Warn("Unable to localize path %v, %v", path, err)
			}
		}
	}

	g.localizationCache[path] = toReturn
	return toReturn
}

func (g *Generator) getPackageToName() map[string]string {
	if g.packageToName == nil {
		g.packageToName = make(map[string]string)
		for _, imp := range g.iface.File.Imports {
			importName, err := g.getNameForImport(imp)
			if err == nil {
				g.packageToName[g.unescapedImportPath(imp)] = importName
			} else {
				log.Warn(err)
			}
		}
		g.packageToName[g.getInterfacePackagePath()] =
			g.getLocalPackageNameFromPackageMap(g.packageToName)
	}
	return g.packageToName
}

func (g *Generator) unescapedImportPath(imp *ast.ImportSpec) string {
	return strings.Replace(imp.Path.Value, "\"", "", -1)
}

func (g *Generator) getNameForImport(imp *ast.ImportSpec) (string, error) {
	if imp.Name != nil {
		return imp.Name.Name, nil
	}
	unescapedPath := g.unescapedImportPath(imp)
	for _, p := range g.iface.Pkg.Imports() {
		if g.getLocalizedPathFromPackage(p) == unescapedPath {
			return p.Name(), nil
		}
	}
	return "", fmt.Errorf("unable to find package name for import: %v", unescapedPath)
}

func (g *Generator) mockName() string {
	if g.ip {
		if ast.IsExported(g.iface.Name) {
			return "Mock" + g.iface.Name
		}
		first := true
		return "mock" + strings.Map(func(r rune) rune {
			if first {
				first = false
				return unicode.ToUpper(r)
			}
			return r
		}, g.iface.Name)
	}

	return g.iface.Name
}

func (g *Generator) getImportStringFromSpec(imp *ast.ImportSpec) string {
	if name, ok := g.getPackageToName()[g.unescapedImportPath(imp)]; ok {
		return fmt.Sprintf("import %s %s\n", name, imp.Path.Value)
	}
	return fmt.Sprintf("import %s\n", imp.Path.Value)
}

func (g *Generator) generateImports() {
	if g.iface.File.Imports == nil {
		return
	}

	for _, imp := range g.iface.File.Imports {
		g.printf(g.getImportStringFromSpec(imp))
	}
}

// GeneratePrologue generates the prologue of the mock.
func (g *Generator) GeneratePrologue(pkg string) {
	if g.ip {
		g.printf("package %s\n\n", g.iface.Pkg.Name())
	} else {
		g.printf("package %v\n\n", pkg)
		g.printf(
			"import %s \"%s\"\n", g.getLocalPackageName(), g.getInterfacePackagePath(),
		)
	}
	g.printf("import \"github.com/stretchr/testify/mock\"\n\n")
	g.generateImports()
	g.printf("\n")
}

// GeneratePrologueNote adds a note after the prologue to the output
// string.
func (g *Generator) GeneratePrologueNote(note string) {
	if note != "" {
		g.printf("\n")
		for _, n := range strings.Split(note, "\\n") {
			g.printf("// %s\n", n)
		}
		g.printf("\n")
	}
}

// ErrNotInterface is returned when the given type is not an interface
// type.
var ErrNotInterface = errors.New("expression not an interface")

func (g *Generator) printf(s string, vals ...interface{}) {
	fmt.Fprintf(&g.buf, s, vals...)
}

var builtinTypes = map[string]bool{
	"ComplexType": true,
	"FloatType":   true,
	"IntegerType": true,
	"Type":        true,
	"Type1":       true,
	"bool":        true,
	"byte":        true,
	"complex128":  true,
	"complex64":   true,
	"error":       true,
	"float32":     true,
	"float64":     true,
	"int":         true,
	"int16":       true,
	"int32":       true,
	"int64":       true,
	"int8":        true,
	"rune":        true,
	"string":      true,
	"uint":        true,
	"uint16":      true,
	"uint32":      true,
	"uint64":      true,
	"uint8":       true,
	"uintptr":     true,
}

type namer interface {
	Name() string
}

func (g *Generator) getInFilePackageNameFromPackage(p *types.Package) string {
	path := p.Path()
	path = g.getLocalizedPathFromPackage(p)
	if name, ok := g.getPackageToName()[path]; ok {
		return name
	}
	log.Warnf("Could not find package name for %v", path)
	return p.Name()
}

func (g *Generator) renderType(t types.Type) string {
	switch t := t.(type) {
	case *types.Named:
		o := t.Obj()
		if o.Pkg() == nil || o.Pkg().Name() == "main" || (g.ip && o.Pkg().Name() == g.pkg) {
			return o.Name()
		}
		return g.getInFilePackageNameFromPackage(o.Pkg()) + "." + o.Name()
	case *types.Basic:
		return t.Name()
	case *types.Pointer:
		return "*" + g.renderType(t.Elem())
	case *types.Slice:
		return "[]" + g.renderType(t.Elem())
	case *types.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), g.renderType(t.Elem()))
	case *types.Signature:
		switch t.Results().Len() {
		case 0:
			return fmt.Sprintf(
				"func(%s)",
				g.renderTypeTuple(t.Params()),
			)
		case 1:
			return fmt.Sprintf(
				"func(%s) %s",
				g.renderTypeTuple(t.Params()),
				g.renderType(t.Results().At(0).Type()),
			)
		default:
			return fmt.Sprintf(
				"func(%s)(%s)",
				g.renderTypeTuple(t.Params()),
				g.renderTypeTuple(t.Results()),
			)
		}
	case *types.Map:
		kt := g.renderType(t.Key())
		vt := g.renderType(t.Elem())

		return fmt.Sprintf("map[%s]%s", kt, vt)
	case *types.Chan:
		switch t.Dir() {
		case types.SendRecv:
			return "chan " + g.renderType(t.Elem())
		case types.RecvOnly:
			return "<-chan " + g.renderType(t.Elem())
		default:
			return "chan<- " + g.renderType(t.Elem())
		}
	case *types.Struct:
		var fields []string

		for i := 0; i < t.NumFields(); i++ {
			f := t.Field(i)

			if f.Anonymous() {
				fields = append(fields, g.renderType(f.Type()))
			} else {
				fields = append(fields, fmt.Sprintf("%s %s", f.Name(), g.renderType(f.Type())))
			}
		}

		return fmt.Sprintf("struct{%s}", strings.Join(fields, ";"))
	case *types.Interface:
		if t.NumMethods() != 0 {
			panic("Unable to mock inline interfaces with methods")
		}

		return "interface{}"
	case namer:
		return t.Name()
	default:
		panic(fmt.Sprintf("un-namable type: %#v (%T)", t, t))
	}
}

func (g *Generator) renderTypeTuple(tup *types.Tuple) string {
	var parts []string

	for i := 0; i < tup.Len(); i++ {
		v := tup.At(i)

		parts = append(parts, g.renderType(v.Type()))
	}

	return strings.Join(parts, " , ")
}

func isNillable(typ types.Type) bool {
	switch t := typ.(type) {
	case *types.Pointer, *types.Array, *types.Map, *types.Interface, *types.Signature, *types.Chan, *types.Slice:
		return true
	case *types.Named:
		return isNillable(t.Underlying())
	}
	return false
}

type paramList struct {
	Names   []string
	Types   []string
	Params  []string
	Nilable []bool
}

func (g *Generator) genList(list *types.Tuple, varadic bool) *paramList {
	var params paramList

	if list == nil {
		return &params
	}

	for i := 0; i < list.Len(); i++ {
		v := list.At(i)

		ts := g.renderType(v.Type())

		if varadic && i == list.Len()-1 {
			t := v.Type()
			switch t := t.(type) {
			case *types.Slice:
				ts = "..." + g.renderType(t.Elem())
			default:
				panic("bad varadic type!")
			}
		}

		pname := v.Name()

		if g.nameCollides(pname) || pname == "" {
			pname = fmt.Sprintf("_a%d", i)
		}

		params.Names = append(params.Names, pname)
		params.Types = append(params.Types, ts)
		params.Params = append(params.Params, fmt.Sprintf("%s %s", pname, ts))
		params.Nilable = append(params.Nilable, isNillable(v.Type()))
	}

	return &params
}

func (g *Generator) nameCollides(pname string) bool {
	if pname == g.pkg {
		return true
	} else if g.iface.Pkg != nil {
		for _, imp := range g.iface.Pkg.Imports() {
			if g.getInFilePackageNameFromPackage(imp) == pname {
				// Argument is same as that of an imported package
				return true
			}
		}
	}
	return false
}

// ErrNotSetup is returned when the generator is not configured.
var ErrNotSetup = errors.New("not setup")

// Generate builds a string that constitutes a valid go source file
// containing the mock of the relevant interface.
func (g *Generator) Generate() error {
	if g.iface == nil {
		return ErrNotSetup
	}

	g.printf(
		"// %s is an autogenerated mock type for the %s type\n", g.mockName(),
		g.iface.Name,
	)
	g.printf(
		"type %s struct {\n\tmock.Mock\n}\n\n", g.mockName(),
	)

	for i := 0; i < g.iface.Type.NumMethods(); i++ {
		fn := g.iface.Type.Method(i)

		ftype := fn.Type().(*types.Signature)
		fname := fn.Name()

		params := g.genList(ftype.Params(), ftype.Variadic())
		returns := g.genList(ftype.Results(), false)

		g.printf(
			"// %s provides a mock function with given fields: %s\n", fname,
			strings.Join(params.Names, ", "),
		)
		g.printf(
			"func (_m *%s) %s(%s) ", g.mockName(), fname,
			strings.Join(params.Params, ", "),
		)

		switch len(returns.Types) {
		case 0:
			g.printf("{\n")
		case 1:
			g.printf("%s {\n", returns.Types[0])
		default:
			g.printf("(%s) {\n", strings.Join(returns.Types, ", "))
		}

		formatParamNames := func() string {
			names := ""
			for i, name := range params.Names {
				if i > 0 {
					names += ", "
				}

				paramType := params.Types[i]
				// for variable args, move the ... to the end.
				if strings.Index(paramType, "...") == 0 {
					name += "..."
				}
				names += name
			}
			return names
		}

		if len(returns.Types) > 0 {
			g.printf("\tret := _m.Called(%s)\n\n", strings.Join(params.Names, ", "))

			var (
				ret []string
			)

			for idx, typ := range returns.Types {
				g.printf("\tvar r%d %s\n", idx, typ)
				g.printf("\tif rf, ok := ret.Get(%d).(func(%s) %s); ok {\n",
					idx, strings.Join(params.Types, ", "), typ)
				g.printf("\t\tr%d = rf(%s)\n", idx, formatParamNames())
				g.printf("\t} else {\n")
				if typ == "error" {
					g.printf("\t\tr%d = ret.Error(%d)\n", idx, idx)
				} else if returns.Nilable[idx] {
					g.printf("\t\tif ret.Get(%d) != nil {\n", idx)
					g.printf("\t\t\tr%d = ret.Get(%d).(%s)\n", idx, idx, typ)
					g.printf("\t\t}\n")
				} else {
					g.printf("\t\tr%d = ret.Get(%d).(%s)\n", idx, idx, typ)
				}
				g.printf("\t}\n\n")

				ret = append(ret, fmt.Sprintf("r%d", idx))
			}

			g.printf("\treturn %s\n", strings.Join(ret, ", "))
		} else {
			g.printf("\t_m.Called(%s)\n", strings.Join(params.Names, ", "))
		}

		g.printf("}\n")
	}

	return nil
}

func (g *Generator) Write(w io.Writer) error {
	opt := &imports.Options{Comments: true}
	theBytes := g.buf.Bytes()

	res, err := imports.Process("mock.go", theBytes, opt)
	if err != nil {
		return err
	}

	w.Write(res)
	return nil
}
