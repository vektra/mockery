package pkg

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go/types"
	"strings"
)

func GenerateVanillaMock(interfacePath string, interfaceName string) (*VanillaOutput, error) {
	parser := NewParser(nil)
	ctx := context.Background()
	err := parser.Parse(ctx, interfacePath)

	if err != nil {
		return nil, err
	}

	err = parser.Load()

	if err != nil {
		return nil, err
	}
	iface, err := parser.Find(interfaceName)

	if err != nil {
		return nil, err
	}

	if iface == nil {
		return nil, errors.New("iface is nil")
	}

	v := VanillaOutput{
		iface: iface,
	}

	for _, m := range iface.Methods() {
		v.parseMethod(m)
	}
	return &v, nil
}

type VanillaOutput struct {
	iface  *Interface
	pkg    string
	fields []string
	impls  []string
}

func (v *VanillaOutput) mockStructName() string {
	return v.iface.Name + "VMock"
}

func (v *VanillaOutput) ifaceFirstLetter() string {
	return strings.ToLower(string(v.iface.Name[0]))
}

func (v *VanillaOutput) addField(fn string) {
	v.fields = append(v.fields, fn)
}

func (v *VanillaOutput) addImpl(impl string) {
	v.impls = append(v.impls, impl)
}

func (v *VanillaOutput) Output() string {
	template := `type %s struct %s
%s
`
	name := v.mockStructName()
	fnsStr := strings.Join(append([]string{"{"}, v.fields...), "\n\t")
	implsStr := strings.Join(append([]string{"}"}, v.impls...), "\n\n")
	return fmt.Sprintf(template, name, fnsStr, implsStr)
}

func (v *VanillaOutput) parseMethod(m *Method) {
	sig := m.Signature.String()
	fName := m.Name + "Fn"
	v.addField(fmt.Sprintf("%s %s", fName, sig))

	params := m.Signature.Params()

	newFnParams := []string{}
	mockFnInputs := []string{}
	for i := 0; i < params.Len(); i++ {
		param := params.At(i)
		pName := param.Name()
		pType := param.Type().String()

		if i == params.Len() - 1 && m.Signature.Variadic() {
			switch t := param.Type().(type) {
			case *types.Slice:
				pType = "..." + t.Elem().String()
			default:
				panic("bad variadic type!")
			}
		}

		if pName == "" {
			pName = fmt.Sprintf("%s%d", strings.ToLower(string(pType[0])), i)
		}
		newParam := fmt.Sprintf("%s %s", pName, pType)
		newFnParams = append(newFnParams, newParam)

		input := pName
		if i == params.Len()-1 && m.Signature.Variadic() {
			input = input + "..."
		}

		mockFnInputs = append(mockFnInputs, input)
	}

	newSig := strings.Join(newFnParams, ", ")
	passArgs := strings.Join(mockFnInputs, ", ")
	rets := m.Signature.Results().String()
	receiver := v.mockStructName()
	iLetter := v.ifaceFirstLetter()
	ret := "return "
	if m.Signature.Results().Len() == 0 {
		ret = ""
	}

	impl := fmt.Sprintf(`func (%s %s) %s(%s) %s {
	%s%s.%s(%s)
}`, iLetter, receiver, m.Name, newSig, rets, ret, iLetter, fName, passArgs)
	v.addImpl(impl)
}