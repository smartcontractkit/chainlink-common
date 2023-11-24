package codec

type ModifiersConfig struct {
	Modifiers []ModifierConfig
}

func (m *ModifiersConfig) UnmarshalJSON(data []byte) error {
	return nil
}

func (m *ModifiersConfig) ToModifier() Modifier {
	modifier := make(ChainModifier, len(m.Modifiers))
	for i, c := range m.Modifiers {
		modifier[i] = c.ToModifier()
	}
	return modifier
}

type ModifierType string

const (
	RenameModifier             ModifierType = "rename"
	RemoveModifier             ModifierType = "remove"
	HardCodeModifier           ModifierType = "hard code"
	ExtractElementModifierType ModifierType = "extract element"
)

type ModifierConfig interface {
	ToModifier() Modifier
}

type ModifierDeserializer struct {
	Type ModifierType
}

type RenameModifierConfig struct {
	Fields map[string]string
}

func (*RenameModifierConfig) isModifcationConfig() {}

type DropModifierConfig struct {
	Fields []string
}

func (*DropModifierConfig) isModifcationConfig() {}

type ElementExtractorConfig struct {
	Extractions map[string]ElementExtractorLocation
}

func (*ElementExtractorConfig) isModifcationConfig() {}

type HardCodeConfig struct {
	Values map[string]any
}

func (*HardCodeConfig) isModifcationConfig() {}
