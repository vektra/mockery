package pkg

import (
	"github.com/vektra/mockery/v2/pkg/config"
	"github.com/vektra/mockery/v2/pkg/template"
)

type TemplateGeneratorConfig struct {
	Style string
}
type TemplateGenerator struct {
	config TemplateGeneratorConfig
}

func NewTemplateGenerator(config TemplateGeneratorConfig) *TemplateGenerator {
	return &TemplateGenerator{
		config: config,
	}
}

func (g *TemplateGenerator) Generate(iface *Interface, ifaceConfig *config.Config) error {
	templ, err := template.New(g.config.Style)
	if err != nil {
		return err
	}
	imports := Imports{}
	for _, method := range iface.Methods() {
		method.populateImports(imports)
	}
	// TODO: Work on getting these imports into the template

	data := template.Data{
		PkgName:         ifaceConfig.Outpkg,
		SrcPkgQualifier: iface.Pkg.Name() + ".",
		Imports: 
	}

	return nil
}
