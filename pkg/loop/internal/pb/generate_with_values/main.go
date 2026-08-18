package main

import (
	"os"

	"github.com/smartcontractkit/chainlink-protos/cre/go/installer/pkg"
)

func main() {
	if len(os.Args) < 2 {
		panic("Usage: generate_with_values <path to values file>")
	}

	gen := pkg.ProtocGen{Plugins: []pkg.Plugin{pkg.GoPlugin, {Name: "go-grpc"}}}
	gen.AddSourceDirectories(".", "../../../")
	for i := 1; i < len(os.Args); i++ {
		if err := gen.GenerateFile(os.Args[i], "."); err != nil {
			panic(err)
		}
	}
}
