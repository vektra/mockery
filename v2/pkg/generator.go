package pkg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"github.com/rs/zerolog"
	"github.com/vektra/mockery/v2/pkg/config"
	"github.com/vektra/mockery/v2/pkg/logging"
	"golang.org/x/tools/imports"
)

const mockConstructorParamTypeNamePrefix = "mockConstructorTestingT"

var invalidIdentifierChar = regexp.MustCompile("[^[:digit:][:alpha:]_]")

// Generator is responsible for generating the string containing
// imports and the mock struct that will later be written out as file.
type Generator struct {
	config.Config
	buf bytes.Buffer

	iface *Interface
	pkg   string

	localizationCache map[string]string
	packagePathToName map[string]string
	nameToPackagePath map[string]string
}

// NewGenerator builds a Generator.
func NewGenerator(ctx context.Context, c config.Config, iface *Interface, pkg string) *Generator {
	g := &Generator{
		Config:            c,
		iface:             iface,
		pkg:               pkg,
		localizationCache: make(map[string]string),
		packagePathToName: make(map[string]string),
		nameToPackagePath: make(map[string]string),
	}

	g.addPackageImportWithName(ctx, "github.com/stretchr/testify/mock", "mock")

	return g
}

func (g *Generator) populateImports(ctx context.Context) {
	log := zerolog.Ctx(ctx)

	log.Debug().Msgf("populating imports")

	// imports from generic type constraints
	if tParams := g.iface.NamedType.TypeParams(); tParams != nil && tParams.Len() > 0 {
		for i := 0; i < tParams.Len(); i++ {
			g.renderType(ctx, tParams.At(i).Constraint())
		}
	}

	// imports from type arguments
	if tArgs := g.iface.NamedType.TypeArgs(); tArgs != nil && tArgs.Len() > 0 {
		for i := 0; i < tArgs.Len(); i++ {
			g.renderType(ctx, tArgs.At(i))
		}
	}

	for _, method := range g.iface.Methods() {
		ftype := method.Signature
		g.addImportsFromTuple(ctx, ftype.Params())
		g.addImportsFromTuple(ctx, ftype.Results())
		g.renderType(ctx, g.iface.NamedType)
	}
}

func (g *Generator) addImportsFromTuple(ctx context.Context, list *types.Tuple) {
	for i := 0; i < list.Len(); i++ {
		// We use renderType here because we need to recursively
		// resolve any types to make sure that all named types that
		// will appear in the interface file are known
		g.renderType(ctx, list.At(i).Type())
	}
}

// getPackageScopedType returns the appropriate string representation for the
// object TypeName. The string may either be the unqualified name (in the case
// the mock will live in the same package as the interface being mocked, e.g.
// `Foo`) or the package pathname (in the case the type lives in a package
// external to the mock, e.g. `packagename.Foo`).
func (g *Generator) getPackageScopedType(ctx context.Context, o *types.TypeName) string {
	if o.Pkg() == nil || o.Pkg().Name() == "main" || (!g.KeepTree && g.InPackage && o.Pkg() == g.iface.Pkg) {
		return o.Name()
	}
	return g.addPackageImport(ctx, o.Pkg()) + "." + o.Name()
}

func (g *Generator) addPackageImport(ctx context.Context, pkg *types.Package) string {
	return g.addPackageImportWithName(ctx, pkg.Path(), pkg.Name())
}

func (g *Generator) addPackageImportWithName(ctx context.Context, path, name string) string {
	if existingName, pathExists := g.packagePathToName[path]; pathExists {
		return existingName
	}

	nonConflictingName := g.getNonConflictingName(path, name)
	g.packagePathToName[path] = nonConflictingName
	g.nameToPackagePath[nonConflictingName] = path
	return nonConflictingName
}

func (g *Generator) getNonConflictingName(path, name string) string {
	if !g.importNameExists(name) && (!g.InPackage || g.iface.Pkg.Name() != name) {
		// do not allow imports with the same name as the package when inPackage
		return name
	}

	// The path will always contain '/' because it is enforced by Go import system
	directories := strings.Split(path, "/")

	cleanedDirectories := make([]string, 0, len(directories))
	for _, directory := range directories {
		cleaned := invalidIdentifierChar.ReplaceAllString(directory, "_")
		cleanedDirectories = append(cleanedDirectories, cleaned)
	}
	numDirectories := len(cleanedDirectories)
	var prospectiveName string
	for i := 1; i <= numDirectories; i++ {
		prospectiveName = strings.Join(cleanedDirectories[numDirectories-i:], "")
		if !g.importNameExists(prospectiveName) && (!g.InPackage || g.iface.Pkg.Name() != prospectiveName) {
			// do not allow imports with the same name as the package when inPackage
			return prospectiveName
		}
	}
	// Try adding numbers to the given name
	i := 2
	for {
		prospectiveName = fmt.Sprintf("%v%d", name, i)
		if !g.importNameExists(prospectiveName) {
			return prospectiveName
		}
		i++
	}
}

func (g *Generator) importNameExists(name string) bool {
	_, nameExists := g.nameToPackagePath[name]
	return nameExists
}

func (g *Generator) maybeMakeNameExported(name string, export bool) string {
	if export && !ast.IsExported(name) {
		return g.makeNameExported(name)
	}

	return name
}

func (g *Generator) makeNameExported(name string) string {
	r, n := utf8.DecodeRuneInString(name)

	if unicode.IsUpper(r) {
		return name
	}

	return string(unicode.ToUpper(r)) + name[n:]
}

func (g *Generator) mockName() string {
	if g.StructName != "" {
		return g.StructName
	}

	if !g.KeepTree && g.InPackage {
		if g.Exported || ast.IsExported(g.iface.Name) {
			return "Mock" + g.iface.Name
		}

		return "mock" + g.makeNameExported(g.iface.Name)
	}

	return g.maybeMakeNameExported(g.iface.Name, g.Exported)
}

// getTypeConstraintString returns type constraint string for a given interface.
//  For instance, a method using this constraint:
//
//    func Foo[T Stringer](s []T) (ret []string) {
//
//    }
//
// The constraint returned will be "[T Stringer]"
//
// https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#type-parameters
func (g *Generator) getTypeConstraintString(ctx context.Context) string {
	tp := g.iface.NamedType.TypeParams()
	if tp == nil || tp.Len() == 0 {
		return ""
	}
	qualifiedParams := make([]string, 0, tp.Len())
	for i := 0; i < tp.Len(); i++ {
		param := tp.At(i)
		qualifiedParams = append(qualifiedParams, fmt.Sprintf("%s %s", param.String(), g.renderType(ctx, param.Constraint())))
	}
	return fmt.Sprintf("[%s]", strings.Join(qualifiedParams, ", "))
}

// getInstantiatedTypeString returns the "instantiated" type names for a given
// constraint list. For instance, if your interface has the constraints
// `[S Stringer, I int, C Comparable]`, this method would return: `[S, I, C]`
func (g *Generator) getInstantiatedTypeString() string {
	tp := g.iface.NamedType.TypeParams()
	if tp == nil || tp.Len() == 0 {
		return ""
	}
	params := make([]string, 0, tp.Len())
	for i := 0; i < tp.Len(); i++ {
		params = append(params, tp.At(i).String())
	}
	return fmt.Sprintf("[%s]", strings.Join(params, ", "))
}

func (g *Generator) expecterName() string {
	return g.mockName() + "_Expecter"
}

func (g *Generator) sortedImportNames() (importNames []string) {
	for name := range g.nameToPackagePath {
		importNames = append(importNames, name)
	}
	sort.Strings(importNames)
	return
}

func (g *Generator) generateImports(ctx context.Context) {
	log := zerolog.Ctx(ctx)

	log.Debug().Msgf("generating imports")
	log.Debug().Msgf("%v", g.nameToPackagePath)

	pkgPath := g.nameToPackagePath[g.iface.Pkg.Name()]
	// Sort by import name so that we get a deterministic order
	for _, name := range g.sortedImportNames() {
		logImport := log.With().Str(logging.LogKeyImport, g.nameToPackagePath[name]).Logger()
		logImport.Debug().Msgf("found import")

		path := g.nameToPackagePath[name]
		if !g.KeepTree && g.InPackage && path == pkgPath {
			logImport.Debug().Msgf("import (%s) equals interface's package path (%s), skipping", path, pkgPath)
			continue
		}
		g.printf("import %s \"%s\"\n", name, path)
	}
}

// GeneratePrologue generates the prologue of the mock.
func (g *Generator) GeneratePrologue(ctx context.Context, pkg string) {
	g.populateImports(ctx)
	if g.InPackage {
		g.printf("package %s\n\n", g.iface.Pkg.Name())
	} else {
		g.printf("package %v\n\n", pkg)
	}

	g.generateImports(ctx)
	g.printf("\n")
}

// GeneratePrologueNote adds a note after the prologue to the output
// string.
func (g *Generator) GeneratePrologueNote(note string) {
	prologue := "// Code generated by mockery"
	if !g.Config.DisableVersionString {
		prologue += fmt.Sprintf(" %s", config.GetSemverInfo())
	}
	prologue += ". DO NOT EDIT.\n"

	g.printf(prologue)
	if note != "" {
		g.printf("\n")
		for _, n := range strings.Split(note, "\\n") {
			g.printf("// %s\n", n)
		}
	}
	g.printf("\n")
}

// GenerateBoilerplate adds a boilerplate text. It should be called
// before any other generator methods to ensure the text is on top.
func (g *Generator) GenerateBoilerplate(boilerplate string) {
	if boilerplate != "" {
		g.printf("%s\n", boilerplate)
	}
}

// ErrNotInterface is returned when the given type is not an interface
// type.
var ErrNotInterface = errors.New("expression not an interface")

func (g *Generator) printf(s string, vals ...interface{}) {
	fmt.Fprintf(&g.buf, s, vals...)
}

var templates = template.New("base template")

func (g *Generator) printTemplate(data interface{}, templateString string) {
	err := templates.ExecuteTemplate(&g.buf, templateString, data)
	if err != nil {
		tmpl, err := templates.New(templateString).Parse(templateString)
		if err != nil {
			// couldn't compile template
			panic(err)
		}
		if err := tmpl.Execute(&g.buf, data); err != nil {
			panic(err)
		}
	}
}

type namer interface {
	Name() string
}

func (g *Generator) renderType(ctx context.Context, typ types.Type) string {
	switch t := typ.(type) {
	case *types.Named:
		name := g.getPackageScopedType(ctx, t.Obj())
		if t.TypeArgs() == nil || t.TypeArgs().Len() == 0 {
			return name
		}
		args := make([]string, 0, t.TypeArgs().Len())
		for i := 0; i < t.TypeArgs().Len(); i++ {
			arg := t.TypeArgs().At(i)
			args = append(args, g.renderType(ctx, arg))
		}
		return fmt.Sprintf("%s[%s]", name, strings.Join(args, ","))
	case *types.TypeParam:
		if t.Constraint() != nil {
			return t.Obj().Name()
		}
		return g.getPackageScopedType(ctx, t.Obj())
	case *types.Basic:
		if t.Kind() == types.UnsafePointer {
			return "unsafe.Pointer"
		}
		return t.Name()
	case *types.Pointer:
		return "*" + g.renderType(ctx, t.Elem())
	case *types.Slice:
		return "[]" + g.renderType(ctx, t.Elem())
	case *types.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), g.renderType(ctx, t.Elem()))
	case *types.Signature:
		switch t.Results().Len() {
		case 0:
			return fmt.Sprintf(
				"func(%s)",
				g.renderTypeTuple(ctx, t.Params()),
			)
		case 1:
			return fmt.Sprintf(
				"func(%s) %s",
				g.renderTypeTuple(ctx, t.Params()),
				g.renderType(ctx, t.Results().At(0).Type()),
			)
		default:
			return fmt.Sprintf(
				"func(%s)(%s)",
				g.renderTypeTuple(ctx, t.Params()),
				g.renderTypeTuple(ctx, t.Results()),
			)
		}
	case *types.Map:
		kt := g.renderType(ctx, t.Key())
		vt := g.renderType(ctx, t.Elem())

		return fmt.Sprintf("map[%s]%s", kt, vt)
	case *types.Chan:
		switch t.Dir() {
		case types.SendRecv:
			return "chan " + g.renderType(ctx, t.Elem())
		case types.RecvOnly:
			return "<-chan " + g.renderType(ctx, t.Elem())
		default:
			return "chan<- " + g.renderType(ctx, t.Elem())
		}
	case *types.Struct:
		var fields []string

		for i := 0; i < t.NumFields(); i++ {
			f := t.Field(i)

			if f.Anonymous() {
				fields = append(fields, g.renderType(ctx, f.Type()))
			} else {
				fields = append(fields, fmt.Sprintf("%s %s", f.Name(), g.renderType(ctx, f.Type())))
			}
		}

		return fmt.Sprintf("struct{%s}", strings.Join(fields, ";"))
	case *types.Interface:
		if t.NumMethods() != 0 {
			panic("Unable to mock inline interfaces with methods")
		}

		rv := []string{"interface{"}
		for i := 0; i < t.NumEmbeddeds(); i++ {
			rv = append(rv, g.renderType(ctx, t.EmbeddedType(i)))
		}
		rv = append(rv, "}")
		sep := ""
		if t.NumEmbeddeds() > 1 {
			sep = "\n"
		}
		return strings.Join(rv, sep)
	case *types.Union:
		rv := make([]string, 0, t.Len())
		for i := 0; i < t.Len(); i++ {
			term := t.Term(i)
			if term.Tilde() {
				rv = append(rv, "~"+g.renderType(ctx, term.Type()))
			} else {
				rv = append(rv, g.renderType(ctx, term.Type()))
			}
		}
		return strings.Join(rv, "|")
	case namer:
		return t.Name()
	default:
		panic(fmt.Sprintf("un-namable type: %#v (%T)", t, t))
	}
}

func (g *Generator) renderTypeTuple(ctx context.Context, tup *types.Tuple) string {
	var parts []string

	for i := 0; i < tup.Len(); i++ {
		v := tup.At(i)

		parts = append(parts, g.renderType(ctx, v.Type()))
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
	Names      []string
	Types      []string
	Params     []string
	ParamsIntf []string
	Nilable    []bool
	Variadic   bool
}

func (g *Generator) genList(ctx context.Context, list *types.Tuple, variadic bool) *paramList {
	var params paramList

	if list == nil {
		return &params
	}

	for i := 0; i < list.Len(); i++ {
		v := list.At(i)

		ts := g.renderType(ctx, v.Type())

		if variadic && i == list.Len()-1 {
			t := v.Type()
			switch t := t.(type) {
			case *types.Slice:
				params.Variadic = true
				ts = "..." + g.renderType(ctx, t.Elem())
			default:
				panic("bad variadic type!")
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

		if strings.Contains(ts, "...") {
			params.ParamsIntf = append(params.ParamsIntf, fmt.Sprintf("%s ...interface{}", pname))
		} else {
			params.ParamsIntf = append(params.ParamsIntf, fmt.Sprintf("%s interface{}", pname))
		}
	}

	return &params
}

func (g *Generator) nameCollides(pname string) bool {
	if pname == g.pkg {
		return true
	}
	return g.importNameExists(pname)
}

// ErrNotSetup is returned when the generator is not configured.
var ErrNotSetup = errors.New("not setup")

// Generate builds a string that constitutes a valid go source file
// containing the mock of the relevant interface.
func (g *Generator) Generate(ctx context.Context) error {
	g.populateImports(ctx)
	if g.iface == nil {
		return ErrNotSetup
	}

	g.printf(
		"// %s is an autogenerated mock type for the %s type\n",
		g.mockName(), g.iface.Name,
	)

	g.printf(
		"type %s%s struct {\n\tmock.Mock\n}\n\n", g.mockName(), g.getTypeConstraintString(ctx),
	)

	if g.WithExpecter {
		g.generateExpecterStruct(ctx)
	}

	for _, method := range g.iface.Methods() {

		// It's probably possible, but not worth the trouble for prototype
		if method.Signature.Variadic() && g.WithExpecter && !g.UnrollVariadic {
			return fmt.Errorf("cannot generate a valid expecter for variadic method with unroll-variadic=false")
		}

		ftype := method.Signature
		fname := method.Name

		params := g.genList(ctx, ftype.Params(), ftype.Variadic())
		returns := g.genList(ctx, ftype.Results(), false)

		if len(params.Names) == 0 {
			g.printf("// %s provides a mock function with given fields:\n", fname)
		} else {
			g.printf(
				"// %s provides a mock function with given fields: %s\n", fname,
				strings.Join(params.Names, ", "),
			)
		}
		g.printf(
			"func (_m *%s%s) %s(%s) ", g.mockName(), g.getInstantiatedTypeString(), fname,
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

		formattedParamNames := ""
		setOfParamNames := make(map[string]struct{}, len(params.Names))
		for i, name := range params.Names {
			if i > 0 {
				formattedParamNames += ", "
			}

			paramType := params.Types[i]
			// for variable args, move the ... to the end.
			if strings.Index(paramType, "...") == 0 {
				name += "..."
			}
			formattedParamNames += name

			setOfParamNames[name] = struct{}{}
		}

		called := g.generateCalled(params, formattedParamNames) // _m.Called invocation string

		if len(returns.Types) > 0 {
			retVariable := resolveCollision(setOfParamNames, "ret")
			g.printf("\t%s := %s\n\n", retVariable, called)

			ret := make([]string, len(returns.Types))

			for idx, typ := range returns.Types {
				g.printf("\tvar r%d %s\n", idx, typ)
				g.printf("\tif rf, ok := %s.Get(%d).(func(%s) %s); ok {\n",
					retVariable, idx, strings.Join(params.Types, ", "), typ)
				g.printf("\t\tr%d = rf(%s)\n", idx, formattedParamNames)
				g.printf("\t} else {\n")
				if typ == "error" {
					g.printf("\t\tr%d = %s.Error(%d)\n", idx, retVariable, idx)
				} else if returns.Nilable[idx] {
					g.printf("\t\tif %s.Get(%d) != nil {\n", retVariable, idx)
					g.printf("\t\t\tr%d = %s.Get(%d).(%s)\n", idx, retVariable, idx, typ)
					g.printf("\t\t}\n")
				} else {
					g.printf("\t\tr%d = %s.Get(%d).(%s)\n", idx, retVariable, idx, typ)
				}
				g.printf("\t}\n\n")

				ret[idx] = fmt.Sprintf("r%d", idx)
			}

			g.printf("\treturn %s\n", strings.Join(ret, ", "))
		} else {
			g.printf("\t%s\n", called)
		}

		g.printf("}\n")

		// Construct expecter helper functions
		if g.WithExpecter {
			g.generateExpecterMethodCall(ctx, method, params, returns)
		}
	}

	g.generateConstructor(ctx)

	return nil
}

func (g *Generator) generateExpecterStruct(ctx context.Context) {
	data := struct {
		MockName, ExpecterName string
		InstantiatedTypeString string
		TypeConstraint         string
	}{
		MockName:               g.mockName(),
		ExpecterName:           g.expecterName(),
		InstantiatedTypeString: g.getInstantiatedTypeString(),
		TypeConstraint:         g.getTypeConstraintString(ctx),
	}
	g.printTemplate(data, `
type {{.ExpecterName}}{{ .TypeConstraint }} struct {
	mock *mock.Mock
}

func (_m *{{.MockName}}{{ .InstantiatedTypeString }}) EXPECT() *{{.ExpecterName}}{{ .InstantiatedTypeString }} {
	return &{{.ExpecterName}}{{ .InstantiatedTypeString }}{mock: &_m.Mock}
}
`)
}

func (g *Generator) generateExpecterMethodCall(ctx context.Context, method *Method, params, returns *paramList) {

	data := struct {
		MockName, ExpecterName string
		CallStruct             string
		MethodName             string
		Params, Returns        *paramList
		LastParamName          string
		LastParamType          string
		NbNonVariadic          int
		InstantiatedTypeString string
		TypeConstraint         string
	}{
		MockName:               g.mockName(),
		ExpecterName:           g.expecterName(),
		CallStruct:             fmt.Sprintf("%s_%s_Call", g.mockName(), method.Name),
		MethodName:             method.Name,
		Params:                 params,
		Returns:                returns,
		InstantiatedTypeString: g.getInstantiatedTypeString(),
		TypeConstraint:         g.getTypeConstraintString(ctx),
	}

	// Get some info about parameters for variadic methods, way easier than doing it in golang template directly
	if data.Params.Variadic {
		data.LastParamName = data.Params.Names[len(data.Params.Names)-1]
		data.LastParamType = strings.TrimLeft(data.Params.Types[len(data.Params.Types)-1], "...")
		data.NbNonVariadic = len(data.Params.Types) - 1
	}

	g.printTemplate(data, `
// {{.CallStruct}} is a *mock.Call that shadows Run/Return methods with type explicit version for method '{{.MethodName}}'
type {{.CallStruct}}{{ .TypeConstraint }} struct {
	*mock.Call
}

// {{.MethodName}} is a helper method to define mock.On call
{{- range .Params.Params}}
//  - {{.}}
{{- end}}
func (_e *{{.ExpecterName}}{{ .InstantiatedTypeString }}) {{.MethodName}}({{range .Params.ParamsIntf}}{{.}},{{end}}) *{{.CallStruct}}{{ .InstantiatedTypeString }} {
	return &{{.CallStruct}}{{ .InstantiatedTypeString }}{Call: _e.mock.On("{{.MethodName}}",
			{{- if not .Params.Variadic }}
				{{- range .Params.Names}}{{.}},{{end}}
			{{- else }}
				append([]interface{}{
					{{- range $i, $name := .Params.Names }}
						{{- if (lt $i $.NbNonVariadic)}} {{$name}},
						{{- else}} }, {{$name}}...
						{{- end}}
					{{- end}} )...
			{{- end }} )}
}

func (_c *{{.CallStruct}}{{ .InstantiatedTypeString }}) Run(run func({{range .Params.Params}}{{.}},{{end}})) *{{.CallStruct}}{{ .InstantiatedTypeString }} {
	_c.Call.Run(func(args mock.Arguments) {
	{{- if not .Params.Variadic }}
		run({{range $i, $type := .Params.Types }}args[{{$i}}].({{$type}}),{{end}})
	{{- else}}
		variadicArgs := make([]{{.LastParamType}}, len(args) - {{.NbNonVariadic}})
		for i, a := range args[{{.NbNonVariadic}}:] {
			if a != nil {
				variadicArgs[i] = a.({{.LastParamType}})
			}
		}
		run(
		{{- range $i, $type := .Params.Types }}
			{{- if (lt $i $.NbNonVariadic)}}args[{{$i}}].({{$type}}),
			{{- else}}variadicArgs...)
			{{- end}}
		{{- end}}
	{{- end}}
	})
	return _c
}

func (_c *{{.CallStruct}}{{ .InstantiatedTypeString }}) Return({{range .Returns.Params}}{{.}},{{end}}) *{{.CallStruct}}{{ .InstantiatedTypeString }} {
	_c.Call.Return({{range .Returns.Names}}{{.}},{{end}})
	return _c
}
`)
}

func (g *Generator) generateConstructor(ctx context.Context) {
	const constructorTemplate = `
type {{ .ConstructorTestingInterfaceName }} interface {
	mock.TestingT
	Cleanup(func())
}

// {{ .ConstructorName }} creates a new instance of {{ .MockName }}. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func {{ .ConstructorName }}{{ .TypeConstraint }}(t {{ .ConstructorTestingInterfaceName }}) *{{ .MockName }}{{ .InstantiatedTypeString }} {
	mock := &{{ .MockName }}{{ .InstantiatedTypeString }}{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
`
	mockName := g.mockName()
	constructorName := g.maybeMakeNameExported("new"+g.makeNameExported(mockName), ast.IsExported(mockName))

	data := struct {
		ConstructorName                 string
		ConstructorTestingInterfaceName string
		InstantiatedTypeString          string
		MockName                        string
		TypeConstraint                  string
	}{
		ConstructorName:                 constructorName,
		ConstructorTestingInterfaceName: mockConstructorParamTypeNamePrefix + constructorName,
		InstantiatedTypeString:          g.getInstantiatedTypeString(),
		MockName:                        mockName,
		TypeConstraint:                  g.getTypeConstraintString(ctx),
	}
	g.printTemplate(data, constructorTemplate)
}

// generateCalled returns the Mock.Called invocation string and, if necessary, prints the
// steps to prepare its argument list.
//
// It is separate from Generate to avoid cyclomatic complexity through early return statements.
func (g *Generator) generateCalled(list *paramList, formattedParamNames string) string {
	namesLen := len(list.Names)
	if namesLen == 0 {
		return "_m.Called()"
	}

	if !list.Variadic {
		return "_m.Called(" + formattedParamNames + ")"
	}

	if !g.UnrollVariadic {
		return "_m.Called(" + strings.Join(list.Names, ", ") + ")"
	}

	var variadicArgsName string
	variadicName := list.Names[namesLen-1]

	// list.Types[] will contain a leading '...'. Strip this from the string to
	// do easier comparison.
	strippedIfaceType := strings.Trim(list.Types[namesLen-1], "...")
	variadicIface := strippedIfaceType == "interface{}"

	if variadicIface {
		// Variadic is already of the interface{} type, so we don't need special handling.
		variadicArgsName = variadicName
	} else {
		// Define _va to avoid "cannot use t (type T) as type []interface {} in append" error
		// whenever the variadic type is non-interface{}.
		g.printf("\t_va := make([]interface{}, len(%s))\n", variadicName)
		g.printf("\tfor _i := range %s {\n\t\t_va[_i] = %s[_i]\n\t}\n", variadicName, variadicName)
		variadicArgsName = "_va"
	}

	// _ca will hold all arguments we'll mirror into Called, one argument per distinct value
	// passed to the method.
	//
	// For example, if the second argument is variadic and consists of three values,
	// a total of 4 arguments will be passed to Called. The alternative is to
	// pass a total of 2 arguments where the second is a slice with those 3 values from
	// the variadic argument. But the alternative is less accessible because it requires
	// building a []interface{} before calling Mock methods like On and AssertCalled for
	// the variadic argument, and creates incompatibility issues with the diff algorithm
	// in github.com/stretchr/testify/mock.
	//
	// This mirroring will allow argument lists for methods like On and AssertCalled to
	// always resemble the expected calls they describe and retain compatibility.
	//
	// It's okay for us to use the interface{} type, regardless of the actual types, because
	// Called receives only interface{} anyway.
	g.printf("\tvar _ca []interface{}\n")

	if namesLen > 1 {
		nonVariadicParamNames := formattedParamNames[0:strings.LastIndex(formattedParamNames, ",")]
		g.printf("\t_ca = append(_ca, %s)\n", nonVariadicParamNames)
	}
	g.printf("\t_ca = append(_ca, %s...)\n", variadicArgsName)

	return "_m.Called(_ca...)"
}

func (g *Generator) Write(w io.Writer) error {
	opt := &imports.Options{Comments: true}
	theBytes := g.buf.Bytes()

	res, err := imports.Process("mock.go", theBytes, opt)
	if err != nil {
		line := "--------------------------------------------------------------------------------------------"
		fmt.Fprintf(os.Stderr, "Between the lines is the file (mock.go) mockery generated in-memory but detected as invalid:\n%s\n%s\n%s\n", line, g.buf.String(), line)
		return err
	}

	w.Write(res)
	return nil
}

func resolveCollision(names map[string]struct{}, variable string) string {
	ret := variable

	for i := len(names); true; i++ {
		_, ok := names[ret]
		if !ok {
			break
		}

		ret = fmt.Sprintf("%s_%d", variable, i)
	}

	return ret
}
