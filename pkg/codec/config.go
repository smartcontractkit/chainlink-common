package codec

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type ModifiersConfig []ModifierConfig

func (m *ModifiersConfig) UnmarshalJSON(data []byte) error {
	var rawDeserialized []map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawDeserialized); err != nil {
		return err
	}

	for _, d := range rawDeserialized {
		for k, v := range d {
			delete(d, k)
			d[strings.ToLower(k)] = v
		}
	}

	*m = make([]ModifierConfig, len(rawDeserialized))

	for i, d := range rawDeserialized {
		mType := ModifierType(strings.ToLower(string(d["type"])))
		switch mType {
		case RenameModifier:
			(*m)[i] = &RenameModifierConfig{}
		case DropModifier:
			(*m)[i] = &DropModifierConfig{}
		case HardCodeModifier:
			(*m)[i] = &HardCodeConfig{}
		case ExtractElementModifierType:
			(*m)[i] = &ElementExtractorConfig{}
		default:
			return fmt.Errorf("%w: unknown modifier type: %s", types.ErrInvalidConfig, mType)
		}

		// json.Unmarshal(d, (*m)[i])
	}
	return nil
}

func (m *ModifiersConfig) ToModifier() (Modifier, error) {
	modifier := make(ChainModifier, len(*m))
	for i, c := range *m {
		mod, err := c.ToModifier()
		if err != nil {
			return nil, err
		}
		modifier[i] = mod
	}
	return modifier, nil
}

type ModifierType string

const (
	RenameModifier             ModifierType = "rename"
	DropModifier               ModifierType = "drop"
	HardCodeModifier           ModifierType = "hard code"
	ExtractElementModifierType ModifierType = "extract element"
)

type ModifierConfig interface {
	ToModifier() (Modifier, error)
}

type RenameModifierConfig struct {
	Fields map[string]string
}

func (*RenameModifierConfig) ToModifier() (Modifier, error) {
	return nil, nil
}

type DropModifierConfig struct {
	Fields []string
}

func (*DropModifierConfig) ToModifier() (Modifier, error) {
	return nil, nil
}

type ElementExtractorConfig struct {
	Extractions map[string]ElementExtractorLocation
}

func (*ElementExtractorConfig) ToModifier() (Modifier, error) {
	return nil, nil
}

type HardCodeConfig struct {
	OffChainValues map[string]any
	OnChainValues  map[string]any
}

func (*HardCodeConfig) ToModifier() (Modifier, error) {
	return nil, nil
}
