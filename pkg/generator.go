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
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"github.com/rs/zerolog"
	"golang.org/x/tools/imports"

	"github.com/vektra/mockery/v2/pkg/logging"
)

const mockConstructorParamTypeNamePrefix = "mockConstructorTestingT"

var invalidIdentifierChar = regexp.MustCompile("[^[:digit:][:alpha:]_]")

func DetermineOutputPackageName(
	interfaceFileName string,
	interfacePackageName string,
	packageNamePrefix string,
	packageName string,
	keepTree bool,
	inPackage bool,
) string {
	var pkg string

	if keepTree && inPackage {
		pkg = filepath.Dir(interfaceFileName)
	} else if inPackage {
		pkg = filepath.Dir(interfaceFileName)
	} else if (packageName == "" || packageName == "mocks") && packageNamePrefix != "" {
		// go with package name prefix only when package name is empty or default and package name prefix is specified
		pkg = fmt.Sprintf("%s%s", packageNamePrefix, interfacePackageName)
	} else {
		pkg = packageName
	}
	return pkg
}

type GeneratorConfig struct {
	Boilerplate          string
	DisableVersionString bool
	Exported             bool
	InPackage            bool
	KeepTree             bool
	Note                 string
	MockBuildTags        string
	PackageName          string
	PackageNamePrefix    string
	StructName           string
	UnrollVariadic       bool
	WithExpecter         bool
	ReplaceType          []string
}

// Generator is responsible for generating the string containing
// imports and the mock struct that will later be written out as file.
type Generator struct {
	config GeneratorConfig

	buf bytes.Buffer

	iface *Interface
	pkg   string

	localizationCache map[string]string
	packagePathToName map[string]string
	nameToPackagePath map[string]string
	replaceTypeCache  []*replaceTypeItem
}

// NewGenerator builds a Generator.
func NewGenerator(ctx context.Context, c GeneratorConfig, iface *Interface, pkg string) *Generator {
	if pkg == "" {
		pkg = DetermineOutputPackageName(
			iface.FileName,
			iface.Pkg.Name(),
			c.PackageNamePrefix,
			c.PackageName,
			c.KeepTree,
			c.InPackage,
		)
	}
	g := &Generator{
		config:            c,
		iface:             iface,
		pkg:               pkg,
		localizationCache: make(map[string]string),
		packagePathToName: make(map[string]string),
		nameToPackagePath: make(map[string]string),
	}

	g.parseReplaceTypes(ctx)
	g.addPackageImportWithName(ctx, "github.com/stretchr/testify/mock", "mock", nil)

	return g
}

func (g *Generator) GenerateAll(ctx context.Context) error {
	g.GenerateBoilerplate(g.config.Boilerplate)
	g.GeneratePrologueNote(g.config.Note)
	g.GenerateBuildTags(g.config.MockBuildTags)
	g.GeneratePrologue(ctx, g.pkg)
	return g.Generate(ctx)
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
	if o.Pkg() == nil || o.Pkg().Name() == "main" ||
		(!g.config.KeepTree && g.config.InPackage && o.Pkg() == g.iface.Pkg) {
		return o.Name()
	}
	pkg := g.addPackageImport(ctx, o.Pkg(), o)
	name := o.Name()
	g.checkReplaceType(ctx, func(from *replaceType, to *replaceType) bool {
		if o.Pkg().Path() == from.pkg && name == from.typ {
			name = to.typ
			return false
		}
		return true
	})
	return pkg + "." + name
}

func (g *Generator) addPackageImport(ctx context.Context, pkg *types.Package, o *types.TypeName) string {
	return g.addPackageImportWithName(ctx, pkg.Path(), pkg.Name(), o)
}

func (g *Generator) checkReplaceType(ctx context.Context, f func(from *replaceType, to *replaceType) bool) {
	// check most specific first
	for _, hasType := range []bool{true, false} {
		for _, item := range g.replaceTypeCache {
			if (item.from.typ != "") == hasType {
				if !f(item.from, item.to) {
					break
				}
			}
		}
	}
}

func (g *Generator) addPackageImportWithName(ctx context.Context, path, name string, o *types.TypeName) string {
	log := zerolog.Ctx(ctx)
	replaced := false
	g.checkReplaceType(ctx, func(from *replaceType, to *replaceType) bool {
		if o != nil && path == from.pkg && (from.typ == "" || o.Name() == from.typ || o.Name() == from.param) {
			log.Debug().Str("from", path).Str("to", to.pkg).Msg("changing package path")
			replaced = true
			path = to.pkg
			if to.alias != "" {
				log.Debug().Str("from", name).Str("to", to.alias).Msg("changing alias name")
				name = to.alias
			}
			return false
		}
		return true
	})
	if replaced {
		log.Debug().Str("to-path", path).Str("to-name", name).Msg("successfully replaced type")
	}

	if existingName, pathExists := g.packagePathToName[path]; pathExists {
		return existingName
	}

	nonConflictingName := g.getNonConflictingName(path, name)
	g.packagePathToName[path] = nonConflictingName
	g.nameToPackagePath[nonConflictingName] = path
	return nonConflictingName
}

func (g *Generator) parseReplaceTypes(ctx context.Context) {
	for _, replace := range g.config.ReplaceType {
		r := strings.SplitN(replace, "=", 2)
		if len(r) != 2 {
			log := zerolog.Ctx(ctx)
			log.Error().Msgf("invalid replace type value: %s", replace)
			continue
		}

		g.replaceTypeCache = append(g.replaceTypeCache, &replaceTypeItem{
			from: parseReplaceType(r[0]),
			to:   parseReplaceType(r[1]),
		})
	}
}

func (g *Generator) getNonConflictingName(path, name string) string {
	if !g.importNameExists(name) && (!g.config.InPackage || g.iface.Pkg.Name() != name) {
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
		if !g.importNameExists(prospectiveName) && (!g.config.InPackage || g.iface.Pkg.Name() != prospectiveName) {
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
	if g.config.StructName != "" {
		return g.config.StructName
	}

	if !g.config.KeepTree && g.config.InPackage {
		if g.config.Exported || ast.IsExported(g.iface.Name) {
			return "Mock" + g.iface.Name
		}

		return "mock" + g.makeNameExported(g.iface.Name)
	}

	return g.maybeMakeNameExported(g.iface.Name, g.config.Exported)
}

// getTypeConstraintString returns type constraint string for a given interface.
//
//	For instance, a method using this constraint:
//	  func Foo[T Stringer](s []T) (ret []string) {
//	  }
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
param:
	for i := 0; i < tp.Len(); i++ {
		param := tp.At(i)
		str := param.String()
		typ := g.renderType(ctx, param.Constraint())

		for _, t := range g.replaceTypeCache {
			if str == t.from.param {
				// Skip removed generic constraints
				if t.from.rmvParam {
					continue param
				}

				// Import replaced generic constraints
				pkg := g.addPackageImportWithName(ctx, t.to.pkg, t.to.alias, param.Obj())
				typ = pkg + "." + t.to.typ
			}
		}

		qualifiedParams = append(qualifiedParams, fmt.Sprintf("%s %s", str, typ))
	}

	if len(qualifiedParams) == 0 {
		return ""
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
param:
	for i := 0; i < tp.Len(); i++ {
		str := tp.At(i).String()

		// Skip replaced generic types
		for _, t := range g.replaceTypeCache {
			if str == t.from.param && t.from.rmvParam {
				continue param
			}
		}

		params = append(params, str)
	}
	if len(params) == 0 {
		return ""
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

	pkgPath := g.nameToPackagePath[g.iface.Pkg.Name()]
	// Sort by import name so that we get a deterministic order
	for _, name := range g.sortedImportNames() {
		logImport := log.With().Str(logging.LogKeyImport, g.nameToPackagePath[name]).Logger()
		logImport.Debug().Msgf("found import")

		path := g.nameToPackagePath[name]
		if !g.config.KeepTree && g.config.InPackage && path == pkgPath {
			logImport.Debug().Msgf("import (%s) equals interface's package path (%s), skipping", path, pkgPath)
			continue
		}
		g.printf("import %s \"%s\"\n", name, path)
	}
}

// GeneratePrologue generates the prologue of the mock.
func (g *Generator) GeneratePrologue(ctx context.Context, pkg string) {
	g.populateImports(ctx)
	if g.config.InPackage {
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
	if !g.config.DisableVersionString {
		prologue += fmt.Sprintf(" %s", logging.GetSemverInfo())
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

func (g *Generator) GenerateBuildTags(buildTags string) {
	if buildTags != "" {
		g.printf("//go:build %s\n\n", buildTags)
	}
}

// ErrNotInterface is returned when the given type is not an interface
// type.
var ErrNotInterface = errors.New("expression not an interface")

func (g *Generator) printf(s string, vals ...interface{}) {
	fmt.Fprintf(&g.buf, s, vals...)
}

var templates = template.New("base template")

func (g *Generator) printTemplateBytes(data interface{}, templateString string) *bytes.Buffer {
	tmpl, err := templates.New(templateString).Funcs(templateFuncMap).Parse(templateString)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, data)
	if err != nil {
		panic(err)
	}

	return &buf
}

func (g *Generator) printTemplate(data interface{}, templateString string) {
	g.buf.Write(g.printTemplateBytes(data, templateString).Bytes())
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
			name := t.Obj().Name()
			pkg := ""

			g.checkReplaceType(ctx, func(from *replaceType, to *replaceType) bool {
				// Replace with the new type if it is being removed as a constraint
				if t.Obj().Pkg().Path() == from.pkg && name == from.param && from.rmvParam {
					name = to.typ
					if to.pkg != from.pkg {
						pkg = g.addPackageImport(ctx, t.Obj().Pkg(), t.Obj())
					}
					return false
				}
				return true
			})

			if pkg != "" {
				return pkg + "." + name
			}
			return name
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
				g.renderTypeTuple(ctx, t.Params(), t.Variadic()),
			)
		case 1:
			return fmt.Sprintf(
				"func(%s) %s",
				g.renderTypeTuple(ctx, t.Params(), t.Variadic()),
				g.renderType(ctx, t.Results().At(0).Type()),
			)
		default:
			return fmt.Sprintf(
				"func(%s)(%s)",
				g.renderTypeTuple(ctx, t.Params(), t.Variadic()),
				g.renderTypeTuple(ctx, t.Results(), false),
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
				field := fmt.Sprintf("%s %s", f.Name(), g.renderType(ctx, f.Type()))
				tag := t.Tag(i)
				if tag != "" {
					field += " `" + tag + "`"
				}
				fields = append(fields, field)
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

func (g *Generator) renderTypeTuple(ctx context.Context, tup *types.Tuple, variadic bool) string {
	var parts []string

	for i := 0; i < tup.Len(); i++ {
		v := tup.At(i)

		if variadic && i == tup.Len()-1 {
			t := v.Type()
			elem := t.(*types.Slice).Elem()

			parts = append(parts, "..."+g.renderType(ctx, elem))
		} else {
			parts = append(parts, g.renderType(ctx, v.Type()))
		}
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

func (p *paramList) FormattedParamNames() string {
	formattedParamNames := ""
	for i, name := range p.Names {
		if i > 0 {
			formattedParamNames += ", "
		}

		paramType := p.Types[i]
		// for variable args, move the ... to the end.
		if strings.Index(paramType, "...") == 0 {
			name += "..."
		}
		formattedParamNames += name
	}

	return formattedParamNames
}

func (p *paramList) ReturnNames() []string {
	names := make([]string, 0, len(p.Names))
	for i := 0; i < len(p.Names); i++ {
		names = append(names, fmt.Sprintf("r%d", i))
	}
	return names
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
		if ts == pname {
			pname = fmt.Sprintf("%s%d", pname, i)
		}

		if g.nameCollides(pname) || pname == "" {
			pname = fmt.Sprintf("_a%d", i)
		}

		params.Names = append(params.Names, pname)
		params.Types = append(params.Types, ts)

		params.Params = append(params.Params, fmt.Sprintf("%s %s", pname, ts))
		params.Nilable = append(params.Nilable, isNillable(v.Type()))

		if variadic && i == list.Len()-1 {
			params.ParamsIntf = append(params.ParamsIntf, fmt.Sprintf("%s ...interface{}", pname))
		} else {
			params.ParamsIntf = append(params.ParamsIntf, fmt.Sprintf("%s interface{}", pname))
		}
	}

	return &params
}

func (g *Generator) nameCollides(pname string) bool {
	if pname == "_" {
		return true
	}
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

	if g.config.WithExpecter {
		g.generateExpecterStruct(ctx)
	}

	for _, method := range g.iface.Methods() {
		g.generateMethod(ctx, method)
	}

	g.generateConstructor(ctx)

	return nil
}

func (g *Generator) generateMethod(ctx context.Context, method *Method) {
	ftype := method.Signature
	fname := method.Name

	params := g.genList(ctx, ftype.Params(), ftype.Variadic())
	returns := g.genList(ctx, ftype.Results(), false)
	preamble, called := g.generateCalled(params, returns)

	data := struct {
		FunctionName           string
		Params                 *paramList
		Returns                *paramList
		MockName               string
		InstantiatedTypeString string
		RetVariableName        string
		Preamble               string
		Called                 string
	}{
		FunctionName:           fname,
		Params:                 params,
		Returns:                returns,
		MockName:               g.mockName(),
		InstantiatedTypeString: g.getInstantiatedTypeString(),
		RetVariableName:        resolveCollision(params.Names, "ret"),
		Preamble:               preamble,
		Called:                 called,
	}

	g.printTemplate(data, `
// {{.FunctionName}} provides a mock function with given fields: {{join .Params.Names ", "}}
func (_m *{{.MockName}}{{.InstantiatedTypeString}}) {{.FunctionName}}({{join .Params.Params ", "}}) {{if (gt (len .Returns.Types) 1)}}({{end}}{{join .Returns.Types ", "}}{{if (gt (len .Returns.Types) 1)}}){{end}} {
{{- .Preamble -}}
{{- if not .Returns.Types}}
	{{- .Called}}
{{- else}}
	{{- .RetVariableName}} := {{.Called}}

	if len({{.RetVariableName}}) == 0 {
		panic("no return value specified for {{.FunctionName}}")
	}

	{{range $idx, $name := .Returns.ReturnNames}}
	var {{$name}} {{index $.Returns.Types $idx -}}
	{{end}}
	{{if gt (len .Returns.Types) 1 -}}
	if rf, ok := {{.RetVariableName}}.Get(0).(func({{join .Params.Types ", "}}) ({{join .Returns.Types ", "}})); ok {
		return rf({{.Params.FormattedParamNames}})
	}
	{{end}}
	{{- range $idx, $name := .Returns.ReturnNames}}
	{{- if $idx}}

	{{end}}
	{{- $typ := index $.Returns.Types $idx -}}
	if rf, ok := {{$.RetVariableName}}.Get({{$idx}}).(func({{join $.Params.Types ", "}}) {{$typ}}); ok {
		r{{$idx}} = rf({{$.Params.FormattedParamNames}})
	} else {
		{{- if eq "error" $typ -}}
		r{{$idx}} = {{$.RetVariableName}}.Error({{$idx}})
		{{- else if (index $.Returns.Nilable $idx) -}}
		if {{$.RetVariableName}}.Get({{$idx}}) != nil {
			r{{$idx}} = {{$.RetVariableName}}.Get({{$idx}}).({{$typ}})
		}
		{{- else -}}
		r{{$idx}} = {{$.RetVariableName}}.Get({{$idx}}).({{$typ}})
		{{- end -}}
	}
	{{- end}}

	return {{join .Returns.ReturnNames ", "}}
{{- end}}
}
`)

	// Construct expecter helper functions
	if g.config.WithExpecter {
		g.generateExpecterMethodCall(ctx, method, params, returns)
	}
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

func (_c *{{.CallStruct}}{{ .InstantiatedTypeString }}) RunAndReturn(run func({{range .Params.Types}}{{.}},{{end}})({{range .Returns.Types}}{{.}},{{end}})) *{{.CallStruct}}{{ .InstantiatedTypeString }} {
	_c.Call.Return(run)
	return _c
}
`)
}

func (g *Generator) generateConstructor(ctx context.Context) {
	const constructorTemplate = `
// {{ .ConstructorName }} creates a new instance of {{ .MockName }}. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func {{ .ConstructorName }}{{ .TypeConstraint }}(t interface {
	mock.TestingT
	Cleanup(func())
}) *{{ .MockName }}{{ .InstantiatedTypeString }} {
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

// generateCalled returns the Mock.Called invocation string and, if necessary, a preamble with the
// steps to prepare its argument list.
//
// It is separate from Generate to avoid cyclomatic complexity through early return statements.
func (g *Generator) generateCalled(list *paramList, returnList *paramList) (preamble string, called string) {
	namesLen := len(list.Names)
	if namesLen == 0 || !list.Variadic || !g.config.UnrollVariadic {
		if list.Variadic && !g.config.UnrollVariadic && g.config.WithExpecter {
			isFuncReturns := len(returnList.Names) > 0

			var tmpRet, tmpRetWithAssignment string
			if isFuncReturns {
				tmpRet = resolveCollision(list.Names, "tmpRet")
				tmpRetWithAssignment = fmt.Sprintf("%s = ", tmpRet)
			}

			calledBytes := g.printTemplateBytes(
				struct {
					ParamList                 *paramList
					ParamNamesWithoutVariadic []string
					VariadicName              string
					IsFuncReturns             bool
					TmpRet                    string
					TmpRetWithAssignment      string
				}{
					ParamList:                 list,
					ParamNamesWithoutVariadic: list.Names[:len(list.Names)-1],
					VariadicName:              list.Names[namesLen-1],
					IsFuncReturns:             isFuncReturns,
					TmpRet:                    tmpRet,
					TmpRetWithAssignment:      tmpRetWithAssignment,
				},
				`{{ if .IsFuncReturns }}var {{ .TmpRet }} mock.Arguments {{ end }}
	if len({{ .VariadicName }}) > 0 {
		{{ .TmpRetWithAssignment }}_m.Called({{ join .ParamList.Names ", " }})
	} else {
		{{ .TmpRetWithAssignment }}_m.Called({{ join .ParamNamesWithoutVariadic ", " }})
	}
`,
			)

			return calledBytes.String(), tmpRet
		}

		called = "_m.Called(" + strings.Join(list.Names, ", ") + ")"
		return
	}

	var variadicArgsName string
	variadicName := list.Names[namesLen-1]

	// list.Types[] will contain a leading '...'. Strip this from the string to
	// do easier comparison.
	strippedIfaceType := strings.Trim(list.Types[namesLen-1], "...")
	variadicIface := strippedIfaceType == "interface{}" || strippedIfaceType == "any"

	if variadicIface {
		// Variadic is already of the interface{} type, so we don't need special handling.
		variadicArgsName = variadicName
	} else {
		// Define _va to avoid "cannot use t (type T) as type []interface {} in append" error
		// whenever the variadic type is non-interface{}.
		preamble += fmt.Sprintf("\t_va := make([]interface{}, len(%s))\n", variadicName)
		preamble += fmt.Sprintf("\tfor _i := range %s {\n\t\t_va[_i] = %s[_i]\n\t}\n", variadicName, variadicName)
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
	preamble += ("\tvar _ca []interface{}\n")

	if namesLen > 1 {
		formattedParamNames := list.FormattedParamNames()
		nonVariadicParamNames := formattedParamNames[0:strings.LastIndex(formattedParamNames, ",")]
		preamble += fmt.Sprintf("\t_ca = append(_ca, %s)\n", nonVariadicParamNames)
	}
	preamble += fmt.Sprintf("\t_ca = append(_ca, %s...)\n", variadicArgsName)

	called = "_m.Called(_ca...)"
	return
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

	_, err = w.Write(res)
	if err != nil {
		return fmt.Errorf("failed to write generator: %w", err)
	}
	return nil
}

func resolveCollision(names []string, variable string) string {
	ret := variable
	set := make(map[string]struct{})
	for _, n := range names {
		set[n] = struct{}{}
	}

	for i := len(names); true; i++ {
		_, ok := set[ret]
		if !ok {
			break
		}

		ret = fmt.Sprintf("%s_%d", variable, i)
	}

	return ret
}

type replaceType struct {
	alias    string
	pkg      string
	typ      string
	param    string
	rmvParam bool
}

type replaceTypeItem struct {
	from *replaceType
	to   *replaceType
}

func parseReplaceType(t string) *replaceType {
	ret := &replaceType{}
	r := strings.SplitN(t, ":", 2)
	if len(r) > 1 {
		ret.alias = r[0]
		t = r[1]
	}

	// Match type parameter substitution
	match := regexp.MustCompile(`\[(.*?)\]$`).FindStringSubmatch(t)
	if len(match) >= 2 {
		ret.param, ret.rmvParam = strings.CutPrefix(match[1], "-")
		t = strings.ReplaceAll(t, match[0], "")
	}

	lastDot := strings.LastIndex(t, ".")
	lastSlash := strings.LastIndex(t, "/")
	if lastDot == -1 || (lastSlash > -1 && lastDot < lastSlash) {
		ret.pkg = t
	} else {
		ret.pkg = t[:lastDot]
		ret.typ = t[lastDot+1:]
	}
	return ret
}
