package main

import (
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/values/installer/pkg"
)

type ProtocHelper struct{}

func (p ProtocHelper) FullGoPackageName(c *pkg.CapabilityConfig) string {
	base := "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/"
	categoryParts := strings.Split(c.Category, "/")
	if len(categoryParts) > 1 {
		base += strings.Join(categoryParts[1:], "/") + "/"
	}
	return base + c.Pkg
}

var _ pkg.ProtocHelper = ProtocHelper{}
