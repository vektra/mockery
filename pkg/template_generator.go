package pkg

import (
	"github.com/vektra/mockery/v2/pkg/registry"
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

func (g *TemplateGenerator) Generate() error {
	templ, err := template.New(g.config.Style)
	if err != nil {
		return err
	}
	data := registry.

	return nil
}
