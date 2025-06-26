package main

import "github.com/smartcontractkit/chainlink-common/pkg/values/installer/pkg"

func main() {
	gen := &pkg.ProtocGen{Plugins: []pkg.Plugin{
		{Name: "cre", Path: "../.."},
	}}
	gen.LinkPackage(pkg.Packages{Go: "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/pb", Proto: "tools/generator/v1alpha/cre_metadata.proto"})
	gen.LinkPackage(pkg.Packages{Go: "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb", Proto: "sdk/v1alpha/sdk.proto"})
	if err := gen.Generate("tools/generator/v1alpha/cre_metadata.proto", "."); err != nil {
		panic(err)
	}
}
