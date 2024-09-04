package cmd

import "embed"

type TemplateAndCondition interface {
	// FS returns the embed.FS that contains the templates.
	// Note, that because of how templates are loaded, only the file name is used to refer the template.
	// therefore, the file names must be unique.
	FS() embed.FS

	// Root returns the template to run to produce the output file
	Root() string
	ShouldGenerate(info GeneratedInfo) bool
}

type BaseGenerate struct {
	FSValue   embed.FS
	RootValue string
}

func (b *BaseGenerate) FS() embed.FS {
	return b.FSValue
}

func (b *BaseGenerate) Root() string {
	return b.RootValue
}

func (b *BaseGenerate) ShouldGenerate(GeneratedInfo) bool {
	return true
}

func (b *BaseGenerate) ForTestOnly() TemplateAndCondition {
	return TestHelperGenerate{BaseGenerate: b}
}

type TestHelperGenerate struct {
	*BaseGenerate
}

func (a TestHelperGenerate) ShouldGenerate(info GeneratedInfo) bool {
	// Consensus algorithms will come with richer manually-created mocks, so helpers are not generated
	// common doesn't have helpers because it's not a capability, but rather a placeholder for common types.
	return info.CapabilityType != "common" && info.CapabilityType != "consensus"
}
