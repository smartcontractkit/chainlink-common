package capabilities

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
	jsonvalidate "github.com/santhosh-tekuri/jsonschema/v5"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type Validator struct {
	reflector           *jsonschema.Reflector
	requestConfig       any
	requestConfigSchema *values.String
}

var _ Validatable = (*Validator)(nil)

func NewValidator(requestConfig any, reflector *jsonschema.Reflector) *Validator {
	return &Validator{
		requestConfig: requestConfig,
		reflector:     reflector,
	}
}

func (v *Validator) GetRequestConfigJSONSchema() *CapabilityResponse {
	if v.requestConfigSchema != nil {
		return &CapabilityResponse{
			Value: v.requestConfigSchema,
		}
	}
	schema := jsonschema.Reflect(v.requestConfig)
	if v.reflector != nil {
		schema = v.reflector.Reflect(v.requestConfig)
	}
	schemaBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return &CapabilityResponse{
			Err: err,
		}
	}

	wrapped := values.NewString(string(schemaBytes))
	v.requestConfigSchema = wrapped
	return &CapabilityResponse{
		Value: wrapped,
	}
}

func (v *Validator) ValidateConfig(config *values.Map) error {
	var x any
	err := config.UnwrapTo(&x)
	if err != nil {
		return err
	}

	generatedSchema := v.GetRequestConfigJSONSchema().Value.(*values.String).Underlying
	jsonSchema, err := jsonvalidate.CompileString("github.com/smartcontractkit/chainlink", generatedSchema)
	if err != nil {
		return err
	}

	err = jsonSchema.Validate(x)
	if err != nil {
		return err
	}

	return nil
}
