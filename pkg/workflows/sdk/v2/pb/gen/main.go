package main

import "github.com/smartcontractkit/chainlink-common/pkg/values/installer/pkg"

func main() {
	gen := &pkg.ProtocGen{}
	gen.LinkPackage(pkg.Packages{
		Go:    "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb",
		Proto: "sdk/v1alpha/sdk.proto",
	})
	if err := gen.Generate("sdk/v1alpha/sdk.proto", "."); err != nil {
		panic(err)
	}
}
