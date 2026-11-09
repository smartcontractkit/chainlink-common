package main

import (
	"os"

	"github.com/smartcontractkit/chainlink-protos/cre/go/installer/pkg"
)

func main() {
	gen := pkg.ProtocGen{Plugins: []pkg.Plugin{pkg.GoPlugin, {Name: "go-grpc"}}}
	gen.AddSourceDirectories(".")
	for i := 1; i < len(os.Args); i++ {
		if err := gen.GenerateFile(os.Args[i], "."); err != nil {
			panic(err)
		}
	}
}
