package main

import "github.com/smartcontractkit/chainlink-common/pkg/values/installer/pkg"

func main() {
	gen := &pkg.ProtocGen{Plugins: []pkg.Plugin{}}
	gen.LinkPackage(pkg.Packages{Go: "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pb", Proto: "tools/generator/v1alpha/cre_metadata.proto"})
	gen.LinkPackage(pkg.Packages{Go: "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pb", Proto: "sdk/v1alpha/sdk.proto"})
	if err := gen.Generate("tools/generator/v1alpha/cre_metadata.proto", "."); err != nil {
		panic(err)
	}

	if err := gen.Generate("sdk/v1alpha/sdk.proto", "."); err != nil {
		panic(err)
	}
}
