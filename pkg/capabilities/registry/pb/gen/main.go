package main

import (
	"github.com/smartcontractkit/chainlink-protos/cre/go/installer/pkg"
)

func main() {
	gen := &pkg.ProtocGen{Plugins: []pkg.Plugin{pkg.GoPlugin, {Name: "go-grpc"}}}
	// "../../../" is pkg/, which is what makes the capabilities/pb imports
	// resolve; this file's own directory covers the local proto.
	gen.AddSourceDirectories(".", "../../../")
	if err := gen.GenerateFile("registry_service.proto", ""); err != nil {
		panic(err)
	}
}
