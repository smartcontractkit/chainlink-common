package cmd

type TemplateAndCondition interface {
	Template() string
	ShouldGenerate(info GeneratedInfo) bool
}

type BaseGenerate struct {
	TemplateValue string
}

func (a BaseGenerate) Template() string {
	return a.TemplateValue
}

func (a BaseGenerate) ShouldGenerate(GeneratedInfo) bool {
	return true
}

type TestHelperGenerate struct {
	TemplateValue string
}

func (a TestHelperGenerate) Template() string {
	return a.TemplateValue
}

func (a TestHelperGenerate) ShouldGenerate(info GeneratedInfo) bool {
	// Consensus algorithms will come with richer manually-created mocks, so helpers are not generated
	// common doesn't have helpers because it's not a capability, but rather a placeholder for common types.
	return info.CapabilityType != "common" && info.CapabilityType != "consensus"
}
