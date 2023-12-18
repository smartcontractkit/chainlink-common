package codec

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type ModifiersConfig []ModifierConfig

func (m *ModifiersConfig) UnmarshalJSON(data []byte) error {
	var rawDeserialized []json.RawMessage
	if err := json.Unmarshal(data, &rawDeserialized); err != nil {
		fmt.Printf("!!!!!!!!\nUnmarshalJSON err\n%v\n!!!!!!!!", string(data))
		return fmt.Errorf("%w: %w", types.ErrInvalidConfig, err)
	}

	*m = make([]ModifierConfig, len(rawDeserialized))

	for i, d := range rawDeserialized {
		t := typer{}
		if err := json.Unmarshal(d, &t); err != nil {
			return fmt.Errorf("%w: %w", types.ErrInvalidConfig, err)
		}

		mType := ModifierType(strings.ToLower(t.Type))
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

		if err := json.Unmarshal(d, (*m)[i]); err != nil {
			return fmt.Errorf("%w: %w", types.ErrInvalidConfig, err)
		}
	}
	return nil
}

func (m *ModifiersConfig) ToModifier(onChainHooks ...mapstructure.DecodeHookFunc) (Modifier, error) {
	modifier := make(MultiModifier, len(*m))
	for i, c := range *m {
		mod, err := c.ToModifier(onChainHooks...)
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
	ToModifier(onChainHooks ...mapstructure.DecodeHookFunc) (Modifier, error)
}

type RenameModifierConfig struct {
	Fields map[string]string
}

func (r *RenameModifierConfig) ToModifier(_ ...mapstructure.DecodeHookFunc) (Modifier, error) {
	return NewRenamer(r.Fields), nil
}

type DropModifierConfig struct {
	Fields []string
}

func (d *DropModifierConfig) ToModifier(onChainHooks ...mapstructure.DecodeHookFunc) (Modifier, error) {
	fields := map[string]string{}
	for i, f := range d.Fields {
		// using a private variable will make the field not serialize, essentially dropping the field
		fields[f] = fmt.Sprintf("dropFieldPrivateName%d", i)
	}

	return NewRenamer(fields), nil
}

type ElementExtractorConfig struct {
	Extractions map[string]ElementExtractorLocation
}

func (e *ElementExtractorConfig) ToModifier(onChainHooks ...mapstructure.DecodeHookFunc) (Modifier, error) {
	return NewElementExtractor(e.Extractions), nil
}

type HardCodeConfig struct {
	OnChainValues  map[string]any
	OffChainValues map[string]any
}

func (h *HardCodeConfig) ToModifier(onChainHooks ...mapstructure.DecodeHookFunc) (Modifier, error) {
	return NewHardCoder(h.OnChainValues, h.OffChainValues, onChainHooks...)
}

type typer struct {
	Type string
}
