package main

// This custom generator is needed because codec.proto imports "values/v1/values.proto"
// which requires special package mapping and proto path resolution that simple protoc
// commands couldn't handle. ProtocGen does this automatically.
import "github.com/smartcontractkit/chainlink-protos/cre/go/installer/pkg"

func main() {
	gen := &pkg.ProtocGen{Plugins: []pkg.Plugin{pkg.GoPlugin, {Name: "go-grpc"}}}
	gen.AddSourceDirectories("../../../../..", ".")
	if err := gen.GenerateFile("codec.proto", "."); err != nil {
		panic(err)
	}
}
