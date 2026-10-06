package main

import (
	"os"

	"github.com/smartcontractkit/chainlink-protos/cre/go/installer/pkg"
)

func main() {
	gen := pkg.ProtocGen{Plugins: []pkg.Plugin{pkg.GoPlugin, {Name: "go-grpc"}}}
	// "." resolves capabilities_registry.proto itself; the second entry resolves
	// pkg/-rooted imports such as capabilities/pb/registry.proto.
	gen.AddSourceDirectories(".", "../../../../")
	for i := 1; i < len(os.Args); i++ {
		if err := gen.GenerateFile(os.Args[i], "."); err != nil {
			panic(err)
		}
	}
}
