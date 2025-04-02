package workflows

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/invopop/jsonschema"
	validate "github.com/santhosh-tekuri/jsonschema/v5"
	"sigs.k8s.io/yaml"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

func GenerateJSONSchema() ([]byte, error) {
	schema := jsonschema.Reflect(&WorkflowSpecYaml{})

	return json.MarshalIndent(schema, "", "  ")
}

func ParseWorkflowSpecYaml(data string) (sdk.WorkflowSpec, error) {
	var url = "https://github.com/smartcontractkit/chainlink/"
	w := WorkflowSpecYaml{}
	err := yaml.Unmarshal([]byte(data), &w)
	if err != nil {
		return sdk.WorkflowSpec{}, err
	}
	schemaStr, err := GenerateJSONSchema()
	if err != nil {
		return sdk.WorkflowSpec{}, err
	}

	schema, err := validate.CompileString(url, string(schemaStr))
	if err != nil {
		return sdk.WorkflowSpec{}, err
	}

	var jsonToValidate any
	err = yaml.Unmarshal([]byte(data), &jsonToValidate)
	if err != nil {
		return sdk.WorkflowSpec{}, err
	}

	err = schema.Validate(jsonToValidate)
	if err != nil {
		return sdk.WorkflowSpec{}, err
	}
	w.yaml = data

	return w.ToWorkflowSpec(), nil
}

// WorkflowSpecYaml is the YAML representation of a workflow spec.
//
// It allows for multiple ways of defining a workflow spec, which we later
// convert to a single representation, `WorkflowSpec`.
type WorkflowSpecYaml struct {
	// NOTE: Name and Owner are constrained the onchain representation in [github.com/smartcontractkit/chainlink-common/blob/main/pkg/capabilities/consensus/ocr3/types/Metadata]

	Name string `json:"name,omitempty" jsonschema:"pattern=^[0-9A-Za-z_\\-]+$,maxLength=10"` // plain text string exactly 10 characters long, or  empty name allowed for anonymous workflows
	//Name nameYaml `json:"name"`
	Owner string `json:"owner,omitempty" jsonschema:"pattern=^0x[0-9a-fA-F]{40}$"` // 20 bytes represented as hex string with 0x prefix, or empty owner allowed for anonymous workflows
	// Triggers define a starting condition for the workflow, based on specific events or conditions.
	Triggers []triggerDefinitionYaml `json:"triggers" jsonschema:"required"`
	// Actions represent a discrete operation within the workflow, potentially transforming input data.
	Actions []stepDefinitionYaml `json:"actions,omitempty"`
	// Consensus encapsulates the logic for aggregating and validating the results from various nodes.
	Consensus []stepDefinitionYaml `json:"consensus,omitempty"`
	// Targets represents the final step of the workflow, delivering the processed data to a specified location.
	Targets []stepDefinitionYaml `json:"targets" jsonschema:"required"`

	yaml string // original yaml spec
}

// ToWorkflowSpec converts a WorkflowSpecYaml to a WorkflowSpec.
//
// We support multiple ways of defining a workflow spec yaml,
// but internally we want to work with a single representation.
func (w WorkflowSpecYaml) ToWorkflowSpec() sdk.WorkflowSpec {
	triggers := make([]sdk.StepDefinition, 0, len(w.Triggers))
	for _, t := range w.Triggers {
		sd := t.toStepDefinition()
		sd.CapabilityType = capabilities.CapabilityTypeTrigger
		triggers = append(triggers, sd)
	}

	actions := make([]sdk.StepDefinition, 0, len(w.Actions))
	for _, a := range w.Actions {
		sd := a.toStepDefinition()
		sd.CapabilityType = capabilities.CapabilityTypeAction
		actions = append(actions, sd)
	}

	consensus := make([]sdk.StepDefinition, 0, len(w.Consensus))
	for _, c := range w.Consensus {
		sd := c.toStepDefinition()
		sd.CapabilityType = capabilities.CapabilityTypeConsensus
		consensus = append(consensus, sd)
	}

	targets := make([]sdk.StepDefinition, 0, len(w.Targets))
	for _, t := range w.Targets {
		sd := t.toStepDefinition()
		sd.CapabilityType = capabilities.CapabilityTypeTarget
		targets = append(targets, sd)
	}

	return sdk.WorkflowSpec{
		Triggers:  triggers,
		Actions:   actions,
		Consensus: consensus,
		Targets:   targets,
		Name:      w.Name,
		Owner:     w.Owner,
	}
}

type Mapping map[string]any

func (m *Mapping) UnmarshalJSON(b []byte) error {
	mp := map[string]any{}

	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()

	err := d.Decode(&mp)
	if err != nil {
		return err
	}

	nm, err := convertNumbers(mp)
	if err != nil {
		return err
	}

	*m = (Mapping)(nm)
	return err
}

// convertNumber detects if a json.Number is an integer or a decimal and converts it to the appropriate type.
//
// Supported type conversions:
// - json.Number -> int64
// - json.Number -> float64
func convertNumber(el any) (any, error) {
	switch elv := el.(type) {
	case json.Number:
		if strings.Contains(elv.String(), ".") {
			return elv.Float64()
		}

		return elv.Int64()
	default:
		return el, nil
	}
}

func convertNumbers(m map[string]any) (map[string]any, error) {
	nm := map[string]any{}
	for k, v := range m {
		switch tv := v.(type) {
		case map[string]any:
			cm, err := convertNumbers(tv)
			if err != nil {
				return nil, err
			}

			nm[k] = cm
		case []any:
			na := make([]any, len(tv))
			for i, v := range tv {
				cv, err := convertNumber(v)
				if err != nil {
					return nil, err
				}

				na[i] = cv
			}

			nm[k] = na
		default:
			cv, err := convertNumber(v)
			if err != nil {
				return nil, err
			}

			nm[k] = cv
		}
	}

	return nm, nil
}

func (m Mapping) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any(m))
}

// triggerDefinitionYaml is the YAML representation of a trigger step in a workflow.
// Like `stepDefinitionYaml`, this will get reduced to a single representation for
// all steps, `sdk.StepDefinition`.
// NOTE: unlike stepDefinitionYaml, this omits the `inputs` field, which isn't used
// for triggers.
type triggerDefinitionYaml struct {
	// A universally unique name for a capability will be defined under the “id” property. The uniqueness will, eventually, be enforced in the Capability Registry.
	//
	// Semver must be used to specify the version of the Capability at the end of the id field. Capability versions must be immutable.
	//
	// There are two ways to specify an id - using a string as a fully qualified ID or a structured table. When using a table, labels are ordered alphanumerically and joined into a string following a
	//  {name}:{label1_key}_{label1_value}:{label2_key}_{label2_value}@{version}
	// pattern.
	//
	// The “id” supports [a-z0-9_-:] characters followed by an @ and [semver regex] at the end.
	//
	// Validation must throw an error if:
	//
	// Unsupported characters are used.
	//
	// Example (string)
	//  id: read_chain:chain_ethereum:network_mainnet@1.0.0
	//
	// Example (table)
	//
	//  id:
	//    name: read_chain
	//    version: 1.0.0
	//    labels:
	//      chain: ethereum
	//      network: mainnet
	//
	// [semver regex]: https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
	ID stepDefinitionID `json:"id" jsonschema:"required"`

	// Actions and Consensus capabilities have a required “ref” property that must be unique within a Workflow file (not universally) This property enables referencing outputs and is required because Actions and Consensus always need to be referenced in the following phases. Triggers can optionally specify  if they need to be referenced.
	//
	// The “ref” supports [a-z0-9_] characters.
	//
	// Validation must throw an error if:
	//  - Unsupported characters are used.
	//  - The same “ref” appears in the workflow multiple times.
	//  - “ref” is used on a Target capability.
	//  - “ref” has a circular reference.
	//
	// NOTE: Should introduce a custom validator to cover trigger case
	Ref string `json:"ref,omitempty" jsonschema:"pattern=^[a-z0-9_-]+$"`

	// The configuration of a Capability will be done using the “config” property. Each capability is responsible for defining an external interface used during setup. This interface may be unique or identical, meaning multiple Capabilities might use the same configuration properties.
	//
	// The interpolation of “inputs”
	//
	// Interpolation of self inputs is allowed from within the “config” property.
	//
	// Example
	//  targets:
	//    - id: write_polygon_mainnet@1.0.0
	//      inputs:
	//        report:
	//          - consensus.evm_median.outputs.report
	//      config:
	//        address: "0xaabbcc"
	//        method: "updateFeedValues(report bytes, role uint8)"
	//        params: [$(inputs.report), 1]
	Config Mapping `json:"config" jsonschema:"required"`
}

// toStepDefinition converts a stepDefinitionYaml to a sdk.StepDefinition.
//
// `sdk.StepDefinition` is the converged representation of a step in a workflow.
func (s triggerDefinitionYaml) toStepDefinition() sdk.StepDefinition {
	return sdk.StepDefinition{
		Ref:    s.Ref,
		ID:     s.ID.String(),
		Config: s.Config,
	}
}

// stepDefinitionYaml is the YAML representation of a step in a workflow.
//
// It allows for multiple ways of defining a step, which we later
// convert to a single representation, `sdk.StepDefinition`.
type stepDefinitionYaml struct {
	// A universally unique name for a capability will be defined under the “id” property. The uniqueness will, eventually, be enforced in the Capability Registry.
	//
	// Semver must be used to specify the version of the Capability at the end of the id field. Capability versions must be immutable.
	//
	// There are two ways to specify an id - using a string as a fully qualified ID or a structured table. When using a table, labels are ordered alphanumerically and joined into a string following a
	//  {name}:{label1_key}_{label1_value}:{label2_key}_{label2_value}@{version}
	// pattern.
	//
	// The “id” supports [a-z0-9_-:] characters followed by an @ and [semver regex] at the end.
	//
	// Validation must throw an error if:
	//
	// Unsupported characters are used.
	//
	// Example (string)
	//  id: read_chain:chain_ethereum:network_mainnet@1.0.0
	//
	// Example (table)
	//
	//  id:
	//    name: read_chain
	//    version: 1.0.0
	//    labels:
	//      chain: ethereum
	//      network: mainnet
	//
	// [semver regex]: https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
	ID stepDefinitionID `json:"id" jsonschema:"required"`

	// Actions and Consensus capabilities have a required “ref” property that must be unique within a Workflow file (not universally) This property enables referencing outputs and is required because Actions and Consensus always need to be referenced in the following phases. Triggers can optionally specify  if they need to be referenced.
	//
	// The “ref” supports [a-z0-9_] characters.
	//
	// Validation must throw an error if:
	//  - Unsupported characters are used.
	//  - The same “ref” appears in the workflow multiple times.
	//  - “ref” is used on a Target capability.
	//  - “ref” has a circular reference.
	//
	// NOTE: Should introduce a custom validator to cover trigger case
	Ref string `json:"ref,omitempty" jsonschema:"pattern=^[a-z0-9_-]+$"`

	// Capabilities can specify an additional optional ”inputs” property. It allows specifying a dependency on the result of one or more other capabilities. These are always runtime values that cannot be provided upfront. It takes a map of the argument name internal to the capability and an explicit reference to the values.
	//
	// References are specified using the [id].[path_to_value] pattern.
	//
	// The interpolation of “inputs” is allowed
	//
	// Validation must throw an error if:
	//  - Input reference cannot be resolved.
	//  - Input is defined on triggers
	// NOTE: Should introduce a custom validator to cover trigger case
	Inputs inputs `json:"inputs" jsonschema:"required"`

	// The configuration of a Capability will be done using the “config” property. Each capability is responsible for defining an external interface used during setup. This interface may be unique or identical, meaning multiple Capabilities might use the same configuration properties.
	//
	// The interpolation of “inputs”
	//
	// Interpolation of self inputs is allowed from within the “config” property.
	//
	// Example
	//  targets:
	//    - id: write_polygon_mainnet@1.0.0
	//      inputs:
	//        report:
	//          - consensus.evm_median.outputs.report
	//      config:
	//        address: "0xaabbcc"
	//        method: "updateFeedValues(report bytes, role uint8)"
	//        params: [$(inputs.report), 1]
	Config Mapping `json:"config" jsonschema:"required"`
}

// toStepDefinition converts a stepDefinitionYaml to a sdk.StepDefinition.
//
// `sdk.StepDefinition` is the converged representation of a step in a workflow.
func (s stepDefinitionYaml) toStepDefinition() sdk.StepDefinition {
	return sdk.StepDefinition{
		Ref: s.Ref,
		ID:  s.ID.String(),
		Inputs: sdk.StepInputs{
			OutputRef: s.Inputs.outputRef,
			Mapping:   s.Inputs.mapping,
		},
		Config: s.Config,
	}
}

// input is the Mapping or interpolation-style representation of the "inputs" field in a sdk.StepDefinition.
// If an interpolation value is provided, it must obey the following rules be a valid interpolation string,
// of the form "$(<REF>.outputs)", where ref is either "trigger" or a ref in the workflow.
type inputs struct {
	mapping   Mapping
	outputRef string
}

var outputRefTokenRe = regexp.MustCompile(`^\$\([a-z0-9_-]+\.outputs\)$`)

func (i *inputs) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err == nil {
		if !outputRefTokenRe.MatchString(s) {
			return errors.New("incorrect ref provided as inputs: ref must be of the form `$(<ref>.outputs)`")
		}

		i.outputRef = s
		return nil
	}

	var m Mapping
	err = json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	i.mapping = m
	return nil
}

func (i *inputs) MarshalJSON() ([]byte, error) {
	if i.outputRef != "" {
		return json.Marshal(i.outputRef)
	}

	return json.Marshal(i.mapping)
}

// JSONSchema returns the JSON schema for a inputs
//
// The schema is a oneOf schema that allows either a string or a Mapping.
func (inputs) JSONSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{DoNotReference: true, ExpandedStruct: true}
	var m Mapping
	mappingSchema := reflector.Reflect(&m)
	mappingSchema.ID = ""
	mappingSchema.Version = ""

	outputRefSchema := &jsonschema.Schema{
		Pattern: outputRefTokenRe.String(),
		Type:    "string",
		ID:      "",
		Version: "",
	}

	return &jsonschema.Schema{
		Title: "id",
		OneOf: []*jsonschema.Schema{
			outputRefSchema,
			mappingSchema,
		},
	}
}

// stepDefinitionID represents both the string and table representations of the "id" field in a sdk.StepDefinition.
type stepDefinitionID struct {
	idStr   string
	idTable *stepDefinitionTableID
}

func (s stepDefinitionID) String() string {
	if s.idStr != "" {
		return s.idStr
	}

	return s.idTable.String()
}

func (s *stepDefinitionID) UnmarshalJSON(data []byte) error {
	// Unmarshal the JSON data into a map to determine if it's a string or a table
	var m string
	err := json.Unmarshal(data, &m)
	if err == nil {
		s.idStr = m
		return nil
	}

	// If the JSON data is a table, unmarshal it into a stepDefinitionTableID
	var table stepDefinitionTableID
	err = json.Unmarshal(data, &table)
	if err != nil {
		return err
	}
	s.idTable = &table
	return nil
}

func (s *stepDefinitionID) MarshalJSON() ([]byte, error) {
	if s.idStr != "" {
		return json.Marshal(s.idStr)
	}

	return json.Marshal(s.idTable)
}

// JSONSchema returns the JSON schema for a stepDefinitionID.
//
// The schema is a oneOf schema that allows either a string or a table.
func (stepDefinitionID) JSONSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{DoNotReference: true, ExpandedStruct: true}
	tableSchema := reflector.Reflect(&stepDefinitionTableID{})
	tableSchema.ID = ""
	tableSchema.Version = ""
	// Allow for a-z, 0-9, _, -, and : characters as the capability type, followed by a semver regex enforcing a full version.
	//
	// Prereleases and build metadata are also allowed
	//
	// Ex. read_chain:chain_ethereum:network_mainnet@1.0.0
	// Ex. read_chain:chain_ethereum:network_mainnet@1.0.0-rc1.1+build1
	stringSchema := &jsonschema.Schema{
		Pattern: "^[a-z0-9_\\-:]+@(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$",
		Type:    "string",
		ID:      "",
		Version: "",
	}

	return &jsonschema.Schema{
		Title: "id",
		OneOf: []*jsonschema.Schema{
			stringSchema,
			tableSchema,
		},
	}
}

// stepDefinitionTableID is the structured representation of a stepDefinitionID.
type stepDefinitionTableID struct {
	Name string `json:"name"`
	// This pattern is the same as the one used in stepDefinitionID.JSONSchema()
	Version string            `json:"version" jsonschema:"pattern=(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$"`
	Labels  map[string]string `json:"labels"`
}

// String returns the string representation of a stepDefinitionTableID.
//
// It follows the format:
//
//	{name}:{label1_key}_{label1_value}:{label2_key}_{label2_value}@{version}
//
// where labels are ordered alphanumerically.
func (s stepDefinitionTableID) String() string {
	labels := make([]string, 0, len(s.Labels))
	for k, v := range s.Labels {
		labels = append(labels, fmt.Sprintf("%s_%s", k, v))
	}
	slices.Sort(labels)

	return fmt.Sprintf("%s:%s@%s", s.Name, strings.Join(labels, ":"), s.Version)
}
