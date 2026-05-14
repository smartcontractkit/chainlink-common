package main

import (
	"github.com/smartcontractkit/chainlink-protos/cre/go/installer/pkg"
)

const scvalGoPackage = "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/stellar/scval"

func main() {
	gen := &pkg.ProtocGen{Plugins: []pkg.Plugin{pkg.GoPlugin, {Name: "go-grpc"}}}
	gen.AddSourceDirectories("../..", ".")
	gen.LinkPackage(pkg.Packages{
		Proto: "capabilities/blockchain/stellar/v1alpha/scval.proto",
		Go:    scvalGoPackage,
	})
	if err := gen.GenerateFile("stellar.proto", "."); err != nil {
		panic(err)
	}
}
