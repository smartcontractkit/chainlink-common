package main

import "github.com/smartcontractkit/chainlink-common/pkg/values/installer/pkg"

func main() {
	gen := &pkg.ProtocGen{}
	gen.LinkPackage(pkg.Packages{
		Go:    "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb",
		Proto: "sdk/v1alpha/sdk.proto",
	})
	gen.LinkPackage(pkg.Packages{
		Go:    "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb",
		Proto: "tools/generator/v1alpha/cre_metadata.proto",
	})
	if err := gen.GenerateFile("sdk/v1alpha/sdk.proto", "."); err != nil {
		panic(err)
	}
	if err := gen.GenerateFile("tools/generator/v1alpha/cre_metadata.proto", "."); err != nil {
		panic(err)
	}
}
