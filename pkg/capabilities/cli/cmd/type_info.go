package cmd

import "github.com/smartcontractkit/capabilities/libs/capabilities"

type TypeInfo struct {
	CapabilityType   capabilities.CapabilityType
	RootType         string
	SchemaID         string
	SchemaOutputFile string
}
