package main

import "github.com/smartcontractkit/chainlink-protos/cre/go/installer/pkg"

func main() {
	gen := &pkg.ProtocGen{Plugins: []pkg.Plugin{pkg.GoPlugin, {Name: "go-grpc"}}}
	gen.AddSourceDirectories("../..", ".")
	if err := gen.GenerateFile("aptos.proto", "."); err != nil {
		panic(err)
	}
}
