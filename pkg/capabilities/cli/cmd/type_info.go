package cmd

import "github.com/smartcontractkit/chainlink-common/pkg/capabilities"

type TypeInfo struct {
	CapabilityType capabilities.CapabilityType
	RootType       string
	SchemaID       string
}
