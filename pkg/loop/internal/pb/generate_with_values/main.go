package main

import (
	"os"

	"github.com/smartcontractkit/chainlink-common/pkg/values/installer/pkg"
)

func main() {
	if len(os.Args) < 2 {
		panic("Usage: generate_with_values <path to values file>")
	}

	gen := pkg.ProtocGen{Plugins: []pkg.Plugin{{Name: "go-grpc"}}}
	gen.AddSourceDirectories(".", "../../../")
	if err := gen.Generate(os.Args[1], "."); err != nil {
		panic(err)
	}
}
