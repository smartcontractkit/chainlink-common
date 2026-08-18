package main

import "github.com/smartcontractkit/chainlink-protos/cre/go/installer/pkg"

func main() {
	gen := &pkg.ProtocGen{Plugins: []pkg.Plugin{pkg.GoPlugin}}
	gen.AddSourceDirectories(".")
	if err := gen.GenerateFile("ocr3_types.proto", "."); err != nil {
		panic(err)
	}
}
