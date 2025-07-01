package main

import (
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/values/installer/pkg"
)

type protocHelper struct{}

var _ pkg.ProtocHelper = protocHelper{}

func (p protocHelper) FullGoPackageName(c *pkg.CapabilityConfig) string {
	prefix := "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/test_capabilities/"
	categoryParts := strings.Split(c.Category, "/")
	if len(categoryParts) > 1 {
		prefix += strings.Join(categoryParts[1:], "/") + "/"
	}
	return prefix + c.Pkg
}

func (p protocHelper) SdkPgk() string {
	return "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pb"
}
