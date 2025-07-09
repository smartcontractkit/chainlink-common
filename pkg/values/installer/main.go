package main

import "github.com/smartcontractkit/chainlink-common/pkg/values/installer/pkg"

func main() {
	gen := &pkg.ProtocGen{}

	if err := gen.GenerateFile("values/v1/values.proto", "."); err != nil {
		panic(err)
	}
}
