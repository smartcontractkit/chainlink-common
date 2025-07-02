package main

import "github.com/smartcontractkit/chainlink-common/pkg/values/installer/pkg"

func main() {
	gen := &pkg.ProtocGen{Plugins: []pkg.Plugin{{Name: "go-grpc"}}}
	gen.AddSourceDirectories("../..", ".")
	if err := gen.GenerateFile("evm.proto", "."); err != nil {
		panic(err)
	}
}
