package template

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/xeipuuv/gojsonschema"
)

var ErrTemplateDataSchemaValidation = errors.New("unable to verify template-data schema")

type TemplateData map[string]any

// VerifyJSONSchema verifies that the contents of the type adhere to the schema
// defined by the schema argument.
//
// This method is not meant to be used directly by templates.
func (t TemplateData) VerifyJSONSchema(ctx context.Context, schema *gojsonschema.Schema) error {
	log := zerolog.Ctx(ctx)

	result, err := schema.Validate(gojsonschema.NewGoLoader(t))
	if err != nil {
		return fmt.Errorf("validating json schema: %w", err)
	}
	if !result.Valid() {
		log.Error().Msg("issue with template-data json schema, see messages below:")
		for _, resultErr := range result.Errors() {
			log.Error().Msg(resultErr.String())
		}
		return ErrTemplateDataSchemaValidation
	}
	log.Debug().Msg("validated json schema successfully")
	return nil
}
