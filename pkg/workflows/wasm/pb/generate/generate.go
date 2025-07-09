package main

import "github.com/smartcontractkit/chainlink-common/pkg/values/installer/pkg"

func main() {
	gen := &pkg.ProtocGen{}
	gen.AddSourceDirectories("../../..", ".")
	if err := gen.GenerateFile("wasm.proto", "."); err != nil {
		panic(err)
	}
}
